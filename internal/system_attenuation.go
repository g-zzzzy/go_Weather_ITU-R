package go_Weather_ITUR

import (
	"go_Weather_ITUR/internal/itur"
	"go_Weather_ITUR/internal/utils"
	"log"
	"sync"
	"time"
)

type AttenuationSystem struct {
	BasicSystem
}

func NewAttenuationSystem(interval int64) *AttenuationSystem {
	return &AttenuationSystem{
		BasicSystem{
			name:     "AttenuationSystem",
			interval: interval,
		},
	}
}

func CalculateUpdateSatelliteLink(link *Link, satMovement *SatelliteMovementComponent, stationPos *StationPositionComponent, pre float64) {

	latGS, lonGS := stationPos.Lat, stationPos.Lon
	el := utils.Elevation_angle(satMovement.PosZ, satMovement.PosX, satMovement.PosY, latGS, lonGS)
	f := 22.5 // GHz
	p := 0.1
	hs := 0.1
	R001 := pre // mm/h
	tau := 45.0 // dB
	var Ls float64
	link.Ar = itur.RainAttenuation(latGS, lonGS, f, el, hs, p, R001, tau, Ls)
}

func CalculateSatelliteLink(satMovement *SatelliteMovementComponent, stationPos *StationPositionComponent, pre float64) float64 {

	latGS, lonGS := stationPos.Lat, stationPos.Lon
	el := utils.Elevation_angle(satMovement.PosZ, satMovement.PosX, satMovement.PosY, latGS, lonGS)
	f := 22.5 // GHz
	p := 0.1
	hs := 0.1
	R001 := pre // mm/h
	tau := 45.0 // dB
	var Ls float64
	return itur.RainAttenuation(latGS, lonGS, f, el, hs, p, R001, tau, Ls)
}

func (s *AttenuationSystem) Update(dt int64, cm *ComponentManager, w *World, t time.Time) {
	log.Printf("AttenuationSystem update...")
	startTime := time.Now()

	// flag := w.parallelFlag // 0 = 块间并行，1 = 块内并行
	// log.Printf("AttenuationSystem parallel flag: %d", flag)
	// numWGs := w.numWGs
	linkBlocks := cm.Links

	var cnt int
	var mu sync.Mutex

	// 块间并行
	var wg sync.WaitGroup
	for _, block := range linkBlocks {
		wg.Add(1)
		go func(block []Link) {
			defer wg.Done()
			localCount := 0
			for i := range block {
				updateLink(&block[i], cm, t)
				localCount++
			}
			mu.Lock()
			cnt += localCount
			mu.Unlock()
		}(block)
	}
	wg.Wait()

	log.Printf("Attenuation computed count: %d", cnt)
	log.Printf("AttenuationSystem update time: %v", time.Since(startTime))

}

func parallelUpdateBlock(block []Link, cm *ComponentManager, cnt *int, t time.Time, numWGs int, mu *sync.Mutex) {
	var wg sync.WaitGroup
	n := len(block)
	if n == 0 {
		return
	}

	chunkSize := (n + numWGs - 1) / numWGs
	for i := 0; i < n; i += chunkSize {
		end := i + chunkSize
		if end > n {
			end = n
		}
		wg.Add(1)
		go func(links []Link) {
			defer wg.Done()
			localCount := 0
			for i := range links {
				updateLink(&links[i], cm, t)
				localCount++
			}
			mu.Lock()
			*cnt += localCount
			mu.Unlock()
		}(block[i:end])
	}
	wg.Wait()
}

func updateLink(link *Link, cm *ComponentManager, t time.Time) {
	satID := link.SourceID
	staID := link.TargetID
	sta := &cm.StationPositionComponents[staID]
	pre := cm.StationWeather[staID].Precipitation
	sat := &cm.GlobalSatPositions[satID]
	// CalculateUpdateSatelliteLink(link, sat, sta, pre)

	latGS, lonGS := sta.Lat, sta.Lon
	el := utils.Elevation_angle(sat.PosZ, sat.PosX, sat.PosY, latGS, lonGS)
	f := 22.5 // GHz
	p := 0.1
	hs := 0.1
	R001 := pre // mm/h
	tau := 45.0 // dB
	var Ls float64
	link.Ar = itur.RainAttenuation(latGS, lonGS, f, el, hs, p, R001, tau, Ls)
}

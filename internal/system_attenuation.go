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
	cnt := 0
	// linkIdx := 0

	// for i := range cm.Links {
	// 	satID := cm.Links[i].SourceID
	// 	staID := cm.Links[i].TargetID
	// 	sta := &cm.StationPositionComponents[staID]
	// 	pre := cm.StationWeather[staID].Precipitation
	// 	sat := &cm.SatelliteMovementComponents[satID]
	// 	cm.Links[i].Ar = CalculateSatelliteLink(sat, sta, pre)
	// 	cnt++
	// }

	numWorkers := 4
	linkCount := len(cm.Links)
	batchSize := (linkCount + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for worker := 0; worker < numWorkers; worker++ {
		start := worker * batchSize
		end := (worker + 1) * batchSize
		if end > linkCount {
			end = linkCount
		}

		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				satID := cm.Links[i].SourceID
				staID := cm.Links[i].TargetID
				sta := &cm.StationPositionComponents[staID]
				pre := cm.StationWeather[staID].Precipitation
				sat := &cm.SatelliteMovementComponents[satID]
				cm.Links[i].Ar = CalculateSatelliteLink(sat, sta, pre)
				cnt++
			}
		}(start, end)
	}

	wg.Wait()

	log.Printf("Attenuation computed count: %d", cnt)
	log.Printf("AttenuationSystem update time: %v", time.Since(startTime))

}

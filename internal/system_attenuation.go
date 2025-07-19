package go_Weather_ITUR

import (
	"go_Weather_ITUR/internal/itur"
	"go_Weather_ITUR/internal/utils"
	"log"
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

func CalculateSatelliteLink(link *LinkComponent, satMovement *SatelliteMovementComponent, stationPos *StationPositionComponent) {
	pre := link.EnvironmentIdx.Precipitation
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

func (s *AttenuationSystem) Update(dt int64, cm *ComponentManager, w *World, t time.Time) {
	log.Printf("AttenuationSystem update...")
	startTime := time.Now()
	cnt := 0
	linkIdx := 0
	for satIdx := range cm.SatelliteMovementComponents {
		sat := &cm.SatelliteMovementComponents[satIdx]

		// 批处理该卫星的所有 station links
		for linkIdx < len(cm.LinkComponents) && cm.LinkComponents[linkIdx].SourceID == satIdx {
			link := &cm.LinkComponents[linkIdx]
			sta := &cm.StationPositionComponents[link.TargetID]

			// 示例：计算衰减
			CalculateSatelliteLink(link, sat, sta)

			linkIdx++
			cnt++
		}
	}

	log.Printf("Attenuation computed count: %d", cnt)
	log.Printf("AttenuationSystem update time: %v", time.Since(startTime))

}

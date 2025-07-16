package go_Weather_ITUR

import (
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

func (s *AttenuationSystem) Update(dt int64, cm *ComponentManager, w *World) {
	log.Printf("AttenuationSystem update...")
	startTime := time.Now()
	cnt := 0
	for LinkKey, linkComp := range cm.LinkComponents {
		if !linkComp.Connected {
			continue
		}
		cnt++
		sourceID := LinkKey.SourceID
		targetID := LinkKey.TargetID
		satIdx, ok1 := cm.MovementEntityToIndex[sourceID]
		posIdx, ok2 := cm.StationEntityToIndex[targetID]
		// weatherIdx, ok3 := cm.WeatherEntityToIndex[targetID]
		_ = cm.SatelliteMovementComponents[satIdx]
		// _ = cm.WeatherComponents[weatherIdx]
		_ = cm.StationPositionComponents[posIdx]

		if !ok1 || !ok2 {
			continue
		}

		// pre := stationWeather.Precipitation
		// latSat, lonSat, hSat := utils.XYZToLatLonAlt(
		// 	satMovement.PosX,
		// 	satMovement.PosY,
		// 	satMovement.PosZ,
		// )

		// latGS, lonGS := stationPos.Lat, stationPos.Lon
		// el := utils.Elevation_angle(hSat, latSat, lonSat, latGS, lonGS)

		// f := 22.5 // GHz
		// p := 0.1
		// hs := 0.1 // km
		// R001 := pre
		// tau := 45.0
		// var Ls float64
		// Ar := itur.RainAttenuation(latGS, lonGS, f, el, hs, p, R001, tau, Ls)

		linkComp.Ar = 0

		// log.Printf("Link Sat %d - Sta %d: Ar=%.2f", sourceID, targetID, Ar)

	}
	log.Printf("Attenuation computed count: %d", cnt)
	log.Printf("AttenuationSystem update time: %v", time.Since(startTime))

}

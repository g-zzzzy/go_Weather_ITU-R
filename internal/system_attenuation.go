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

func (s *AttenuationSystem) Update(dt int64, cm *ComponentManager, w *World) {
	log.Printf("AttenuationSystem update...")
	startTime := time.Now()
	cnt := 0
	for linkKey, linkComp := range cm.LinkComponents {
		if !linkComp.Connected {
			continue
		}
		cnt++
		sourceID := linkKey.SourceID
		targetID := linkKey.TargetID

		satIdx, satOk := cm.MovementEntityToIndex[sourceID]
		weatherIdx, weatherOk := cm.WeatherEntityToIndex[targetID]
		stationIdx, posOk := cm.StationEntityToIndex[targetID]

		if !satOk || !weatherOk || !posOk {
			continue
		}

		satMovement := &cm.SatelliteMovementComponents[satIdx]
		stationWeather := &cm.WeatherIndexComponents[weatherIdx]
		stationPos := &cm.StationPositionComponents[stationIdx]

		pre := stationWeather.precipitation
		latSat, lonSat, hSat := utils.XYZToLatLonAlt(
			satMovement.Position.X,
			satMovement.Position.Y,
			satMovement.Position.Z,
		)

		latGS, lonGS := stationPos.Lat, stationPos.Lon
		el := utils.Elevation_angle(hSat, latSat, lonSat, latGS, lonGS)

		f := 22.5
		p := 0.1
		hs := 0.1
		R001 := pre
		tau := 45.0
		var Ls float64
		Ar := itur.RainAttenuation(latGS, lonGS, f, el, hs, p, R001, tau, Ls)

		cm.AttenuationComponents[linkKey] = AttenuationComponent{Attenuation: Ar}
	}
	log.Printf("Attenuation computed count: %d", cnt)
	endTime := time.Now()
	log.Printf("AttenuationSystem update time: %v", endTime.Sub(startTime))

}

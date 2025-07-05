package go_Weather_ITUR

import (
	"go_Weather_ITUR/internal/itur"
	"go_Weather_ITUR/internal/utils"
	"log"
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
	cnt := 0
	for LinkKey, linkComp := range cm.LinkComponents {
		if !linkComp.Connected {
			continue
		}
		cnt++
		sourceID := LinkKey.SourceID
		targetID := LinkKey.TargetID
		satMovement, satOk := cm.SatelliteMovementComponents[sourceID]
		stationWeather, stationOk := cm.WeatherIndexComponents[targetID]
		stationPos, posOk := cm.StationPositionComponents[targetID]

		if !satOk || !stationOk || !posOk {
			continue
		}

		pre := stationWeather.precipitation
		latSat, lonSat, hSat := utils.XYZToLatLonAlt(
			satMovement.Position.X,
			satMovement.Position.Y,
			satMovement.Position.Z,
		)

		latGS, lonGS := stationPos.Lat, stationPos.Lon
		el := utils.Elevation_angle(hSat, latSat, lonSat, latGS, lonGS)

		f := 22.5 // GHz
		p := 0.1
		hs := 0.1 // km
		R001 := pre
		tau := 45.0
		var Ls float64
		Ar := itur.RainAttenuation(latGS, lonGS, f, el, hs, p, R001, tau, Ls)

		cm.AttenuationComponents[LinkKey] = AttenuationComponent{
			Attenuation: Ar,
		}

		// log.Printf("Link Sat %d - Sta %d: Ar=%.2f", sourceID, targetID, Ar)

	}
	log.Printf("Attenuation computed count: %d", cnt)

}

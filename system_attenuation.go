package go_Weather_ITUR

import (
	"fmt"
	"go_Weather_ITUR/itur"
	"go_Weather_ITUR/utils"
	"log"
)

type AttenuationSystem struct {
	BasicSystem
	attenuations map[uint64]*AttenuationEntity
}

func (s *AttenuationSystem) Add(attenuation *AttenuationEntity, w *World) {
	if s.attenuations == nil {
		s.attenuations = make(map[uint64]*AttenuationEntity)
	}
	s.attenuations[attenuation.GetBasicEntity().id] = attenuation
	w.componentManager.AddComponent(EntityID(attenuation.id), AttenuationComponent{})
}

func (s *AttenuationSystem) Update(dt int64, cm *ComponentManager, w *World) {
	s.AddElapsed(dt)
	if s.ShouldUpdate(s.elapsed) {
		println("attenuation compute")
		satelliteCount := len(cm.satelliteMovementComponents)
		stationCount := len(cm.weatherIndexComponents)

		cm.attenuationComponents = make([][]AttenuationComponent, satelliteCount)
		for i := range cm.attenuationComponents {
			cm.attenuationComponents[i] = make([]AttenuationComponent, stationCount)
		}

		satelliteIDs, err1 := w.GetSystemEntityIDs("SatelliteSystem")
		if err1 != nil {
			fmt.Println("Error:", err1)
		}
		stationIDs, err2 := w.GetSystemEntityIDs("StationSystem")
		if err2 != nil {
			fmt.Println("Error:", err2)
		} else {
			fmt.Println("stationID:", stationIDs)
		}

		for _, satelliteID := range satelliteIDs {
			for _, stationID := range stationIDs {
				satelliteMovement := cm.satelliteMovementComponents[satelliteID]
				weatherIndex := cm.weatherIndexComponents[stationID]
				pre := weatherIndex.precipitation
				// T := weatherIndex.T
				lat_sat, lon_sat, h_sat := utils.XYZToLatLonAlt(satelliteMovement.position.X, satelliteMovement.position.Y, satelliteMovement.position.Z)
				log.Printf("卫星纬度: %.2f, 卫星经度： %.2f, 卫星高度(m)： %.2f", lat_sat, lon_sat, h_sat)

				lat := cm.stationPositionComponents[stationID].lat
				lon := cm.stationPositionComponents[stationID].lon
				el := utils.Elevation_angle(h_sat, lat_sat, lon_sat, lat, lon)

				f := 22.5 //GHz
				// D := 1.2	//m
				p := 0.1 //Unavailability (Vals exceeded 0.1% of time)

				//RainAttenuation(lat, lon, f, el, hs, p, R001, tau, Ls float64)
				hs := 0.1 //km
				R001 := pre
				tau := 45.0
				var Ls float64
				Ar := itur.RainAttenuation(lat, lon, f, el, hs, p, R001, tau, Ls)

				cm.attenuationComponents[satelliteID][stationID] = AttenuationComponent{
					attenuation: Ar,
				}
			}
		}

		for _, satelliteID := range satelliteIDs {
			for _, stationID := range stationIDs {
				log.Printf("卫星%d 与 地面站%d 间通信的雨衰为%.2f dB", satelliteID, stationID, cm.attenuationComponents[satelliteID][stationID])
			}
		}
	}
}

func (s *AttenuationSystem) GetEntityIDs() []uint64 {
	var ids []uint64
	for id := range s.attenuations {
		ids = append(ids, id)
	}
	return ids
}

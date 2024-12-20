package go_Weather_ITUR

import (
	"fmt"
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
				attenuation := satelliteMovement.position.X + satelliteMovement.position.Y + satelliteMovement.position.Z + weatherIndex.precipitation + weatherIndex.T
				cm.attenuationComponents[satelliteID][stationID] = AttenuationComponent{
					attenuation: attenuation,
				}
			}
		}

		for _, satelliteID := range satelliteIDs {
			for _, stationID := range stationIDs {
				log.Printf("卫星%d 与 地面站%d 通信的损失为%.2f", satelliteID, stationID, cm.attenuationComponents[satelliteID][stationID])
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

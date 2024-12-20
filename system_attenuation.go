package go_Weather_ITUR

import "log"

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

		for satelliteID := 0; satelliteID < satelliteCount; satelliteID++ {
			for stationID := 0; stationID < stationCount; stationID++ {
				
				satelliteMovement := cm.satelliteMovementComponents[satelliteID]
				weatherIndex := cm.weatherIndexComponents[stationID]

				attenuation := satelliteMovement.position.X + satelliteMovement.position.Y + satelliteMovement.position.Z + weatherIndex.precipitation + weatherIndex.T
				cm.attenuationComponents[satelliteID][stationID] = AttenuationComponent{
					attenuation: attenuation,
				}
			}
		}

		for satelliteID := 0; satelliteID < satelliteCount; satelliteID++ {
			for stationID := 0; stationID < stationCount; stationID++ {
				log.Printf("卫星%d 与 地面站%d 通信的损失为%.2f", satelliteID, stationID, cm.attenuationComponents[satelliteID][stationID])
			}
		}

	}
}

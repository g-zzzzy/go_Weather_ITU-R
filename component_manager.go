package go_Weather_ITUR

type ComponentManager struct {
	// satellitePositionComponents []SatellitePositionComponent
	// satelliteVelocityComponents []SatelliteVelocityComponent
	satelliteMovementComponents []SatelliteMovementComponent
	stationPositionComponents   []StationPositionComponent
	weatherIndexComponents      []WeatherIndexComponent
}

func NewComponentManager() *ComponentManager {
	return &ComponentManager{
		// satellitePositionComponents: make([]SatellitePositionComponent, 0),
		// satelliteVelocityComponents: make([]SatelliteVelocityComponent, 0),
		satelliteMovementComponents: make([]SatelliteMovementComponent, 0),
		stationPositionComponents:   make([]StationPositionComponent, 0),
		weatherIndexComponents:      make([]WeatherIndexComponent, 0),
	}
}

func (cm *ComponentManager) AddComponent(entityID EntityID, component interface{}) {
	switch comp := component.(type) {
	// case SatellitePositionComponent:
	// 	for len(cm.satellitePositionComponents) <= int(entityID) {
	// 		cm.satellitePositionComponents = append(cm.satellitePositionComponents, SatellitePositionComponent{satellite.Vector3{X: 0.0, Y: 0.0, Z: 0.0}})
	// 	}
	// 	cm.satellitePositionComponents[entityID] = comp
	// case SatelliteVelocityComponent:
	// 	for len(cm.satelliteVelocityComponents) <= int(entityID) {
	// 		cm.satelliteVelocityComponents = append(cm.satelliteVelocityComponents, SatelliteVelocityComponent{satellite.Vector3{X: 0.0, Y: 0.0, Z: 0.0}})
	// 	}
	// 	cm.satelliteVelocityComponents[entityID] = comp
	case SatelliteMovementComponent:
		for len(cm.satelliteMovementComponents) <= int(entityID) {
			cm.satelliteMovementComponents = append(cm.satelliteMovementComponents, SatelliteMovementComponent{})
		}
		cm.satelliteMovementComponents[entityID] = comp
	case StationPositionComponent:
		for len(cm.stationPositionComponents) <= int(entityID) {
			cm.stationPositionComponents = append(cm.stationPositionComponents, StationPositionComponent{})
		}
		cm.stationPositionComponents[entityID] = comp
	case WeatherIndexComponent:
		for len(cm.weatherIndexComponents) <= int(entityID) {
			cm.weatherIndexComponents = append(cm.weatherIndexComponents, WeatherIndexComponent{})
		}
		cm.weatherIndexComponents[entityID] = comp
	}

}

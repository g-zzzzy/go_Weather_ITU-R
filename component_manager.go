package go_Weather_ITUR

type ComponentManager struct {
	satelliteMovementComponents []SatelliteMovementComponent
	stationPositionComponents   []StationPositionComponent
	weatherIndexComponents      []WeatherIndexComponent
	attenuationComponents       [][]AttenuationComponent
}

func NewComponentManager() *ComponentManager {
	return &ComponentManager{
		satelliteMovementComponents: make([]SatelliteMovementComponent, 0),
		stationPositionComponents:   make([]StationPositionComponent, 0),
		weatherIndexComponents:      make([]WeatherIndexComponent, 0),
		attenuationComponents:       make([][]AttenuationComponent, 0),
	}
}

func (cm *ComponentManager) AddComponent(entityID EntityID, component interface{}) {
	switch comp := component.(type) {
	case SatelliteMovementComponent:
		for len(cm.satelliteMovementComponents) <= int(entityID) {
			cm.satelliteMovementComponents = append(cm.satelliteMovementComponents, SatelliteMovementComponent{})
		}
		cm.satelliteMovementComponents[entityID] = comp
	case WeatherIndexComponent:
		for len(cm.weatherIndexComponents) <= int(entityID) {
			cm.weatherIndexComponents = append(cm.weatherIndexComponents, WeatherIndexComponent{})
		}
		cm.weatherIndexComponents[entityID] = comp
	case StationPositionComponent:
		for len(cm.stationPositionComponents) <= int(entityID) {
			cm.stationPositionComponents = append(cm.stationPositionComponents, StationPositionComponent{})
		}
		cm.stationPositionComponents[entityID] = comp

	}

}

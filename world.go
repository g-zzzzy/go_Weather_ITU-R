package go_Weather_ITUR

import "fmt"

type World struct {
	systems          []System
	componentManager *ComponentManager
}

func NewWorld() *World {
	return &World{
		componentManager: NewComponentManager(),
		systems:          make([]System, 0),
	}
}

func (w *World) AddSystem(system System) {
	w.systems = append(w.systems, system)
}

func (w *World) GetSystem() []System {
	return w.systems
}

func (w *World) Update(dt int64) {
	for _, system := range w.GetSystem() {
		system.Update(dt, w.componentManager, w)
	}
}

func (w *World) GetSystemEntityIDs(targetSystemType string) ([]uint64, error) {
	for _, system := range w.systems {
		switch targetSystemType {
		case "StationSystem":
			if stationSystem, ok := system.(*StationSystem); ok {
				return stationSystem.GetEntityIDs(), nil
			}
		case "SatelliteSystem":
			if satelliteSystem, ok := system.(*SatelliteSystem); ok {
				return satelliteSystem.GetEntityIDs(), nil
			}
		}
	}
	return nil, fmt.Errorf("system of type not found: %s", targetSystemType)
}

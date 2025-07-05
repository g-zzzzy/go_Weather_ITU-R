package go_Weather_ITUR

import (
	"errors"
)

type EntityID int

type World struct {
	Systems      []System
	Components   *ComponentManager
	nextEntityID EntityID
}

func NewWorld() *World {
	return &World{
		Components: &ComponentManager{
			TLEComponents:               make(map[EntityID]TLEComponent),
			SatelliteSGP4Components:     make(map[EntityID]SatelliteSGP4Component),
			SatelliteMovementComponents: make(map[EntityID]SatelliteMovementComponent),
			StationPositionComponents:   make(map[EntityID]StationPositionComponent),
			WeatherIndexComponents:      make(map[EntityID]WeatherIndexComponent),
			AttenuationComponents:       make(map[LinkKey]AttenuationComponent),
			LinkComponents:              make(map[LinkKey]LinkComponent),
		},
	}
}

func (w *World) GetSystemEntityIDs(name string) ([]EntityID, error) {
	for _, sys := range w.Systems {
		if sys.Name() == name {
			return sys.GetEntityIDs(), nil
		}
	}
	return nil, errors.New("system not found: " + name)
}

func (w *World) NewEntity() EntityID {
	id := w.nextEntityID
	w.nextEntityID++
	return id
}

func (w *World) AddSystem(s System) {
	w.Systems = append(w.Systems, s)
}

func (w *World) Update(dt int64) {
	// startTime := time.Now()
	for _, system := range w.Systems {
		system.Update(dt, w.Components, w)

		// system.AddElapsed(dt)
		// if system.ShouldUpdate(system.GetElapsed()) {
		// 	system.Update(dt, w.Components, w)
		// }
	}
	// endTime := time.Now()
	// log.Printf("Update time: %v", endTime.Sub(startTime))
}

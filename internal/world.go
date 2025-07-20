package go_Weather_ITUR

import (
	"errors"
	"time"
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
			SatelliteMovementComponents: make([]SatelliteMovementComponent, 0),
			MovementEntityToIndex:       make(map[EntityID]int),
			StationPositionComponents:   make([]StationPositionComponent, 0),
			StationEntityToIndex:        make(map[EntityID]int),
			Links:                       make([]Link, 0),
			StationWeather:              make([]EnvironmentIndex, 0),
		},
		nextEntityID: 0,
		Systems:      make([]System, 0),
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
		time := time.Now()
		system.Update(dt, w.Components, w, time)

		// system.AddElapsed(dt)
		// if system.ShouldUpdate(system.GetElapsed()) {
		// 	system.Update(dt, w.Components, w)
		// }
	}
	// endTime := time.Now()
	// log.Printf("Update time: %v", endTime.Sub(startTime))
}

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
			SatelliteMovementComponents: make([]SatelliteMovementComponent, 0),
			MovementEntityToIndex:       make(map[EntityID]int),
			StationPositionComponents:   make([]StationPositionComponent, 0),
			StationEntityToIndex:        make(map[EntityID]int),
			WeatherComponents:           make([]WeatherComponent, 0),
			WeatherEntityToIndex:        make(map[EntityID]int),
			AttenuationInputComponents:  make([]AttenuationInputComponent, 0),
			// AttenuationInputEntityToIndex:  make(map[EntityID]int),
			// AttenuationOutputComponents:    make([]AttenuationOutputComponent, 0),
			// AttenuationOutputEntityToIndex: make(map[EntityID]int),
			LinkComponents: make(map[LinkKey]LinkComponent),
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
		system.Update(dt, w.Components, w)

		// system.AddElapsed(dt)
		// if system.ShouldUpdate(system.GetElapsed()) {
		// 	system.Update(dt, w.Components, w)
		// }
	}
	// endTime := time.Now()
	// log.Printf("Update time: %v", endTime.Sub(startTime))
}

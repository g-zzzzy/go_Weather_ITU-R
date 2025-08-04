package go_Weather_ITUR

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joshuaferrara/go-satellite"
)

type EntityID int

type World struct {
	Systems      map[SystemType]System
	Components   *ComponentManager
	nextEntityID EntityID
}

func NewWorld() *World {
	return &World{
		Components: &ComponentManager{
			// TLEComponents:               make(map[EntityID]TLEComponent),
			SatelliteSGP4Components:     make([]SatelliteSGP4Component, 0),
			SatelliteMovementComponents: make([]SatelliteMovementComponent, 0),
			// MovementEntityToIndex:       make(map[EntityID]int),
			StationPositionComponents: make([]StationPositionComponent, 0),
			// StationEntityToIndex:        make(map[EntityID]int),
			Links:          make([]Link, 0),
			StationWeather: make([]EnvironmentIndex, 0),
		},
		nextEntityID: 0,
		Systems:      make(map[SystemType]System),
	}
}

func (world *World) InitSatellite(satelliteCount int, satelliteSystem *SatelliteSystem) {
	filename_tle := "data/satellite_4000.txt"
	file, err := os.Open(filename_tle)
	if err != nil {
		fmt.Println("Error Loading TLE:", err)
	} else {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		readSat := 0
		for scanner.Scan() && readSat < satelliteCount {
			l1 := scanner.Text()

			if !scanner.Scan() {
				break
			}
			l2 := scanner.Text()
			entityID := world.NewEntity()
			tleComponent := TLEComponent{
				Line1:     l1,
				Line2:     l2,
				GravConst: "wgs72",
			}
			// world.Components.TLEComponents[entityID] = tleComponent

			// log.Printf("TLEToSat")
			satelliteSGP4Component := SatelliteSGP4Component{
				// Satrec: satellite.TLEToSat(tleComponent.Line1, tleComponent.Line2, tleComponent.GravConst),
				Satrec: satellite.ParseTLE(tleComponent.Line1, tleComponent.Line2, tleComponent.GravConst),
			}
			for len(world.Components.SatelliteSGP4Components) <= int(entityID) {
				world.Components.SatelliteSGP4Components = append(world.Components.SatelliteSGP4Components, SatelliteSGP4Component{})
			}
			world.Components.SatelliteSGP4Components[entityID] = satelliteSGP4Component

			movementComp := SatelliteMovementComponent{EntityID: entityID}
			for len(world.Components.SatelliteMovementComponents) <= int(entityID) {
				world.Components.SatelliteMovementComponents = append(world.Components.SatelliteMovementComponents, SatelliteMovementComponent{})
			}
			world.Components.SatelliteMovementComponents = append(world.Components.SatelliteMovementComponents, movementComp)
			// world.Components.MovementEntityToIndex[entityID] = len(world.Components.SatelliteMovementComponents) - 1

			satelliteSystem.AddEntityID(entityID)
			readSat++

		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Scanner Error: ", err)
		}
	}
}

func (world *World) InitStation(stationCount int, stationSystem *StationSystem) {

	filename_station := "data/terminal.txt"
	file, err := os.Open(filename_station)
	if err != nil {
		fmt.Println("Error Loading Station:", err)
	} else {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		readSta := 0
		for scanner.Scan() && readSta < stationCount {
			line := scanner.Text()
			parts := strings.Fields(line)
			if len(parts) != 2 {
				continue
			}
			lat, err1 := strconv.ParseFloat(parts[0], 64)
			lon, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 != nil || err2 != nil {
				fmt.Println("Error Lat: ", err1)
				fmt.Println("Error Lon: ", err2)
				continue
			} else {
				entityID := world.NewEntity()
				for len(world.Components.StationPositionComponents) <= int(entityID) {
					world.Components.StationPositionComponents = append(world.Components.StationPositionComponents, StationPositionComponent{})
				}
				world.Components.StationPositionComponents = append(world.Components.StationPositionComponents, StationPositionComponent{
					EntityID: entityID,
					Lat:      lat,
					Lon:      lon,
				})
				for len(world.Components.StationWeather) <= int(entityID) {
					world.Components.StationWeather = append(world.Components.StationWeather, EnvironmentIndex{})
				}
				world.Components.StationWeather = append(world.Components.StationWeather, EnvironmentIndex{})

				stationSystem.AddEntityID(entityID)
				readSta++
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Scanner Error: ", err)
		}
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

func (w *World) AddSystem(systemType SystemType, s System) {
	if w.Systems == nil {
		w.Systems = make(map[SystemType]System)
	}
	w.Systems[systemType] = s
}

func (w *World) Update(dt int64) {
	var wg sync.WaitGroup
	now := time.Now()

	// 并行更新 Satellite 和 Station
	wg.Add(2)
	go func() {
		defer wg.Done()
		if sys, ok := w.Systems[SatelliteSystemType]; ok {
			sys.Update(dt, w.Components, w, now)
		}
	}()
	go func() {
		defer wg.Done()
		if sys, ok := w.Systems[StationSystemType]; ok {
			sys.Update(dt, w.Components, w, now)
		}
	}()
	wg.Wait()

	// 顺序更新 Topo 和 Attenuation
	if sys, ok := w.Systems[TopoSystemType]; ok {
		sys.Update(dt, w.Components, w, now)
	}
	if sys, ok := w.Systems[AttenuationSystemType]; ok {
		sys.Update(dt, w.Components, w, now)
	}

	// startTime := time.Now()
	// for _, system := range w.Systems {
	// 	time := time.Now()
	// 	system.Update(dt, w.Components, w, time)

	// 	// system.AddElapsed(dt)
	// 	// if system.ShouldUpdate(system.GetElapsed()) {
	// 	// 	system.Update(dt, w.Components, w)
	// 	// }
	// }
	// endTime := time.Now()
	// log.Printf("Update time: %v", endTime.Sub(startTime))
}

package main

import (
	"bufio"
	"fmt"
	internal "go_Weather_ITUR/internal"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joshuaferrara/go-satellite"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <station_num> <satellite_nums>", os.Args[0])
	}
	// 解析参数
	stationCount, err1 := strconv.Atoi(os.Args[1])
	satelliteCount, err2 := strconv.Atoi(os.Args[2])
	if err1 != nil || err2 != nil {
		log.Fatalf("Invalid arguments: %v, %v", err1, err2)
	}
	fmt.Printf("ECS running with %d stations and %d satellites\n", stationCount, satelliteCount)

	startTime := time.Now()

	world := internal.NewWorld()

	satelliteSystem := internal.NewSatelliteSystem(5)
	stationSystem := internal.NewStationSystem(10)
	// topoSystem := internal.NewTopoSystem(10)
	// attenuationSystem := internal.NewAttenuationSystem(10)

	world.AddSystem(satelliteSystem)
	world.AddSystem(stationSystem)
	// world.AddSystem(topoSystem)
	// world.AddSystem(attenuationSystem)

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
			tleComponent := internal.TLEComponent{
				Line1:     l1,
				Line2:     l2,
				GravConst: "wgs72",
			}
			world.Components.TLEComponents[entityID] = tleComponent

			// log.Printf("TLEToSat")
			satelliteSGP4Component := internal.SatelliteSGP4Component{

				// Satrec: satellite.TLEToSat(tleComponent.Line1, tleComponent.Line2, tleComponent.GravConst),
				Satrec: satellite.ParseTLE(tleComponent.Line1, tleComponent.Line2, tleComponent.GravConst),
			}
			world.Components.SatelliteSGP4Components[entityID] = satelliteSGP4Component

			movementComp := internal.SatelliteMovementComponent{EntityID: entityID}
			world.Components.SatelliteMovementComponents = append(world.Components.SatelliteMovementComponents, movementComp)
			world.Components.MovementEntityToIndex[entityID] = len(world.Components.SatelliteMovementComponents) - 1

			satelliteSystem.AddEntityID(entityID)
			readSat++

		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Scanner Error: ", err)
		}
	}

	filename_station := "data/terminal.txt"
	file, err = os.Open(filename_station)
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
				posComponent := internal.StationPositionComponent{Lat: lat, Lon: lon}
				world.Components.StationPositionComponents = append(world.Components.StationPositionComponents, posComponent)
				world.Components.StationEntityToIndex[entityID] = len(world.Components.StationPositionComponents) - 1

				weatherComp := internal.WeatherIndexComponent{EntityID: entityID}
				world.Components.WeatherIndexComponents = append(world.Components.WeatherIndexComponents, weatherComp)
				world.Components.WeatherEntityToIndex[entityID] = len(world.Components.WeatherIndexComponents) - 1

				stationSystem.AddEntityID(entityID)
				readSta++

			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Scanner Error: ", err)
		}
	}
	// for i := 0; i < 100; i++ {

	for i := 0; i < 100; i++ {
		world.Update(1)
		log.Printf("Run %d complete\n", i+1)
	}

	endTime := time.Now()
	log.Println("Total update time: ", endTime.Sub(startTime))
}

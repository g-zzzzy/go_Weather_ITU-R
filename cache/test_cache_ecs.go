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

	_ "net/http/pprof"

	"github.com/joshuaferrara/go-satellite"
)

func main() {
	// go func() {
	// for i := 0; i < 20; i++ {

	startTime := time.Now()

	world := internal.NewWorld()

	satelliteSystem := internal.NewSatelliteSystem(5)
	stationSystem := internal.NewStationSystem(10)
	topoSystem := internal.NewTopoSystem(10)
	attenuationSystem := internal.NewAttenuationSystem(10)

	world.AddSystem(satelliteSystem)
	world.AddSystem(stationSystem)
	world.AddSystem(topoSystem)
	world.AddSystem(attenuationSystem)

	filename_tle := "data/satellite_tle_data2000.txt"
	file, err := os.Open(filename_tle)
	if err != nil {
		fmt.Println("Error Loading TLE:", err)
	} else {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
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

			satelliteSGP4Component := internal.SatelliteSGP4Component{
				Satrec: satellite.TLEToSat(tleComponent.Line1, tleComponent.Line2, tleComponent.GravConst),
			}
			world.Components.SatelliteSGP4Components[entityID] = satelliteSGP4Component
			world.Components.SatelliteMovementComponents[entityID] = internal.SatelliteMovementComponent{}

			satelliteSystem.AddEntityID(entityID)
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Scanner Error: ", err)
		}
	}

	filename_station := "data/station_data500.txt"
	file, err = os.Open(filename_station)
	if err != nil {
		fmt.Println("Error Loading Station:", err)
	} else {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
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
				posComponent := internal.StationPositionComponent{
					Lat: lat,
					Lon: lon,
				}
				world.Components.StationPositionComponents[entityID] = posComponent
				world.Components.WeatherIndexComponents[entityID] = internal.WeatherIndexComponent{}
				stationSystem.AddEntityID(entityID)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Scanner Error: ", err)
		}
	}
	// for i := 0; i < 100; i++ {

	world.Update(1)
	endTime := time.Now()
	log.Println("Update time: ", endTime.Sub(startTime))
	// }
	// }
	// world.Update(1)
	// }()

	// addr := "0.0.0.0:6060" // pprof 服务监听端口
	// log.Println("pprof server listening on", addr)
	// if err := http.ListenAndServe(addr, nil); err != nil {
	// 	log.Fatal(err)
	// }

}

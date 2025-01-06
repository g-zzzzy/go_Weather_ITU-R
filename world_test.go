package go_Weather_ITUR

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestWorldUpdate(t *testing.T) {
	world := NewWorld()

	satelliteSystem := &SatelliteSystem{
		BasicSystem: BasicSystem{
			interval: 5,
			elapsed:  0,
		},
	}
	stationSystem := &StationSystem{
		BasicSystem: BasicSystem{
			interval: 10,
			elapsed:  0,
		},
	}
	attenuationSystem := &AttenuationSystem{
		BasicSystem: BasicSystem{
			interval: 10,
			elapsed:  0,
		},
	}
	world.AddSystem(satelliteSystem)
	world.AddSystem(stationSystem)
	world.AddSystem(attenuationSystem)

	filename_tle := "data/satellite_tle_data.txt"
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
			satellite := &SatelliteEntity{
				BasicEntity: NewBasic(),
				TLE: TLEComponent{
					line1:     l1,
					line2:     l2,
					gravConst: "wgs72",
				},
			}
			satelliteSystem.Add(satellite, world)
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Scanner Error: ", err)
		}
	}

	filename_station := "data/station_data.txt"
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
			} else {
				station := &StationEntity{
					BasicEntity: NewBasic(),
					position: StationPositionComponent{
						lat: lat,
						lon: lon,
					},
				}
				stationSystem.Add(station, world)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Scanner Error: ", err)
		}
	}

	for i := 1; i <= 20; i++ {
		println("Tick:", i)
		world.Update(1)
		time.Sleep(1000 * time.Millisecond)
	}
}

package go_Weather_ITUR

import (
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
	// stationSystem := &StationSystem{
	// 	BasicSystem: BasicSystem{
	// 		interval: 0,
	// 		elapsed:  0,
	// 	},
	// }
	// weatherSystem := &WeatherSystem{
	// 	BasicSystem: BasicSystem{
	// 		interval: 10,
	// 		elapsed:  0,
	// 	},
	// }
	world.AddSystem(satelliteSystem)
	// world.AddSystem(stationSystem)
	// world.AddSystem(weatherSystem)

	satellite1 := &SatelliteEntity{
		BasicEntity: NewBasic(),
		TLE: TLEComponent{
			line1:     "1 06251U 62025E   06176.82412014  .00008885  00000-0  12808-3 0  3985",
			line2:     "2 06251  58.0579  54.0425 0030035 139.1568 221.1854 15.56387291  6774",
			gravConst: "wgs72",
		},
	}
	satellite2 := &SatelliteEntity{
		BasicEntity: NewBasic(),
		TLE: TLEComponent{
			line1:     "1 23599U 95029B   06171.76535463  .00085586  12891-6  12956-2 0  2905",
			line2:     "2 23599   6.9327   0.2849 5782022 274.4436  25.2425  4.47796565123555",
			gravConst: "wgs72",
		},
	}
	satelliteSystem.Add(satellite1, world)
	satelliteSystem.Add(satellite2, world)

	for i := 1; i <= 20; i++ {
		println("Tick:", i)
		world.Update(1)
		time.Sleep(1000 * time.Millisecond)
	}
}

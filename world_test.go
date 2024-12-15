package go_Weather_ITUR

import (
	"fmt"
	"testing"
	"time"
)

func TestWorldUpdate(t *testing.T) {
	world := &World{}

	satelliteSystem := &SatelliteSystem{
		BasicSystem: BasicSystem{
			interval: 5,
			elapsed:  0,
		},
	}
	stationSystem := &StationSystem{
		BasicSystem: BasicSystem{
			interval: 0,
			elapsed:  0,
		},
	}
	weatherSystem := &WeatherSystem{
		BasicSystem: BasicSystem{
			interval: 10,
			elapsed:  0,
		},
	}
	world.AddSystem(satelliteSystem)
	world.AddSystem(stationSystem)
	world.AddSystem(weatherSystem)

	satellite1 := &SatelliteEntity{
		BasicEntity: NewBasic(),
		TLE: TLEComponent{
			line1:     "1 23599U 95029B   06171.76535463  .00085586  12891-6  12956-2 0  2905",
			line2:     "2 23599   6.9327   0.2849 5782022 274.4436  25.2425  4.47796565123555",
			gravConst: "wgs72",
		},
	}
	satellite2 := &SatelliteEntity{
		BasicEntity: NewBasic(),
		TLE: TLEComponent{
			line1:     "1 06251U 62025E   06176.82412014  .00008885  00000-0  12808-3 0  3985",
			line2:     "2 06251  58.0579  54.0425 0030035 139.1568 221.1854 15.56387291  6774",
			gravConst: "wgs72",
		},
	}
	satelliteSystem.Add(satellite1)
	satelliteSystem.Add(satellite2)

	station1 := &StationEntity{
		BasicEntity: NewBasic(),
		position: StationPositionComponent{
			lat: 10.0,
			lon: 20.0,
		},
	}
	stationSystem.Add(station1)

	weather1 := &WeatherIndexEntity{
		BasicEntity: NewBasic(),
		indexs: WeatherIndexComponent{
			T:             20.1,
			P:             20.2,
			V_t:           20.4,
			rho:           20.5,
			precipitation: 20.1,
			hr:            25.5,
		},
	}
	weatherSystem.Add(weather1)

	for _, system := range world.GetSystem() {
		for _, entity := range system.GetEntity() {
			if satellite, ok := entity.(*SatelliteEntity); ok {
				fmt.Printf("卫星实体ID: %d，位置信息：纬度 %.2f，经度 %.2f，高度 %.2f\n", satellite.GetBasicEntity().id, satellite.position.position.X, satellite.position.position.Y, satellite.position.position.Z)
			} else if station, ok := entity.(*StationEntity); ok {
				fmt.Printf("地面站ID：%d, 位置信息：纬度 %.2f，经度 %.2f\n", station.GetBasicEntity().id, station.position.lat, station.position.lon)
			} else if weather, ok := entity.(*WeatherIndexEntity); ok {
				fmt.Printf("天气ID：%d, 温度信息： %.2f，降雨： %.2f", weather.GetBasicEntity().id, weather.indexs.T, weather.indexs.precipitation)
			}
		}
	}

	for i := 1; i <= 10; i++ {
		println("Tick:", i)
		world.Update(1)
		time.Sleep(1000 * time.Millisecond)
	}
}

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
		position: SatellitePositionComponent{
			lat: 15.0,
			lon: 25.0,
			h:   35.0,
		},
	}
	satellite2 := &SatelliteEntity{
		BasicEntity: NewBasic(),
		position: SatellitePositionComponent{
			lat: 10.0,
			lon: 20.0,
			h:   30.0,
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
				fmt.Printf("卫星实体ID: %d，位置信息：纬度 %.2f，经度 %.2f，高度 %.2f\n", satellite.GetBasicEntity().id, satellite.position.lat, satellite.position.lon, satellite.position.h)
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

package go_Weather_ITUR

import (
	"context"
	"go_Weather_ITUR/weather"
	"log"

	"google.golang.org/grpc"
)

type StationSystem struct {
	BasicSystem
	stations map[uint64]*StationEntity
}

func (s *StationSystem) Add(station *StationEntity, w *World) {
	if s.stations == nil {
		s.stations = make(map[uint64]*StationEntity)
	}
	log.Printf("station %d added", station.GetBasicEntity().id)
	s.stations[station.GetBasicEntity().id] = station
	w.componentManager.AddComponent(EntityID(station.id), WeatherIndexComponent{})
}

func (s *StationSystem) Update(dt int64, cm *ComponentManager, w *World) {
	s.AddElapsed(dt)
	if s.ShouldUpdate(s.elapsed) {
		println("weather index update")
		conn, err := grpc.Dial("10.0.0.52:50051", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()

		client := weather.NewWeatherServiceClient(conn)

		// 具体的update
		for i := 0; i < len(cm.weatherIndexComponents); i++ {
			weatherIndexComponent := &cm.weatherIndexComponents[i]

			if _, exists := s.stations[uint64(i)]; exists {
				lat := float32(s.stations[uint64(i)].position.lat)
				lon := float32(s.stations[uint64(i)].position.lon)
				request := &weather.WeatherRequest{
					StartDate:    "2023-08-07",
					EndDate:      "2023-08-16",
					SpecificDate: "2023-08-07",
					Time:         "12",
					Lat:          lat,
					Lon:          lon,
				}
				response, err := client.GetWeather(context.Background(), request)
				if err != nil {
					log.Fatalf("could not get weather: %v", err)
					continue
				}
				weatherIndexComponent.T = float64(response.Temp)
				weatherIndexComponent.precipitation = float64(response.Precipitation)
				log.Printf("降水：%.2f, 温度：%.2f", cm.weatherIndexComponents[i].precipitation, cm.weatherIndexComponents[i].T)
			}
		}
	}
}

func (s *StationSystem) GetEntityIDs() []uint64 {
	var ids []uint64
	for id := range s.stations {
		ids = append(ids, id)
	}
	return ids
}

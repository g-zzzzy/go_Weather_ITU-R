package go_Weather_ITUR

import (
	"context"
	"go_Weather_ITUR/weather"
	"log"

	"google.golang.org/grpc"
)

type WeatherSystem struct {
	BasicSystem
	weatherIndexs map[uint64]*WeatherIndexEntity
}

func (s *WeatherSystem) Add(weatherIndex *WeatherIndexEntity, w *World) {
	if s.weatherIndexs == nil {
		s.weatherIndexs = make(map[uint64]*WeatherIndexEntity)
	}
	s.weatherIndexs[weatherIndex.GetBasicEntity().id] = weatherIndex
	w.componentManager.AddComponent(EntityID(weatherIndex.id), WeatherIndexComponent{})
}

func (s *WeatherSystem) Update(dt int64, cm *ComponentManager) {
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

			if _, exists := s.weatherIndexs[uint64(i)]; exists {
				request := &weather.WeatherRequest{
					StartDate: "2023-08-07",
					EndDate:   "2023-08-16",
					Time:      "12",
					Lat:       31.0,
					Lon:       121.5,
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

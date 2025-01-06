package go_Weather_ITUR

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type StationSystem struct {
	BasicSystem
	stations map[uint64]*StationEntity
}

type ForecastResponse struct {
	Hourly struct {
		Temperature2m   []float64 `json:"temperature_2m"`
		Precipitation   []float64 `json:"precipitation"`
		SurfacePressure []float64 `json:"surface_pressure"`
	} `json:"hourly"`
}

func (s *StationSystem) Add(station *StationEntity, w *World) {
	if s.stations == nil {
		s.stations = make(map[uint64]*StationEntity)
	}
	log.Printf("station %d added", station.GetBasicEntity().id)
	s.stations[station.GetBasicEntity().id] = station
	w.componentManager.AddComponent(EntityID(station.id), station.position)
	w.componentManager.AddComponent(EntityID(station.id), WeatherIndexComponent{})
}

func (s *StationSystem) Update(dt int64, cm *ComponentManager, w *World) {
	s.AddElapsed(dt)
	if s.ShouldUpdate(s.elapsed) {
		println("weather index update")

		for i := 0; i < len(cm.weatherIndexComponents); i++ {
			weatherIndexComponent := &cm.weatherIndexComponents[i]

			if _, exists := s.stations[uint64(i)]; exists {
				lat := float32(s.stations[uint64(i)].position.lat)
				lon := float32(s.stations[uint64(i)].position.lon)

				url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.2f&longitude=%.2f&hourly=temperature_2m,precipitation,surface_pressure", lat, lon)
				// 发送 GET 请求
				resp, err := http.Get(url)
				if err != nil {
					log.Fatalf("Error making request: %v", err)
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatalf("Error reading response body: %v", err)
				}
				// 解析 JSON 数据
				var forecast ForecastResponse
				err = json.Unmarshal(body, &forecast)
				if err != nil {
					log.Fatalf("Error unmarshalling JSON: %v", err)
				}

				weatherIndexComponent.T = float64(forecast.Hourly.Temperature2m[0])
				weatherIndexComponent.precipitation = float64(forecast.Hourly.Precipitation[0])
				weatherIndexComponent.P = float64(forecast.Hourly.SurfacePressure[0])
				log.Printf("降水：%.2f, 温度：%.2f, 气压：%.2f", cm.weatherIndexComponents[i].precipitation, cm.weatherIndexComponents[i].T, cm.weatherIndexComponents[i].P)
			}

			// conn, err := grpc.Dial("10.0.0.52:50051", grpc.WithInsecure())
			// if err != nil {
			// 	log.Fatalf("did not connect: %v", err)
			// }
			// defer conn.Close()

			// client := weather.NewWeatherServiceClient(conn)

			// // 具体的update
			// for i := 0; i < len(cm.weatherIndexComponents); i++ {
			// 	weatherIndexComponent := &cm.weatherIndexComponents[i]

			// 	if _, exists := s.stations[uint64(i)]; exists {
			// 		lat := float32(s.stations[uint64(i)].position.lat)
			// 		lon := float32(s.stations[uint64(i)].position.lon)
			// 		request := &weather.WeatherRequest{
			// 			StartDate:    "2023-08-07",
			// 			EndDate:      "2023-08-16",
			// 			SpecificDate: "2023-08-07",
			// 			Time:         "12",
			// 			Lat:          lat,
			// 			Lon:          lon,
			// 		}
			// 		response, err := client.GetWeather(context.Background(), request)
			// 		if err != nil {
			// 			log.Fatalf("could not get weather: %v", err)
			// 			continue
			// 		}
			// 		weatherIndexComponent.T = float64(response.Temp)
			// 		weatherIndexComponent.precipitation = float64(response.Precipitation)
			// 		log.Printf("降水：%.2f, 温度：%.2f", cm.weatherIndexComponents[i].precipitation, cm.weatherIndexComponents[i].T)
			// 	}
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

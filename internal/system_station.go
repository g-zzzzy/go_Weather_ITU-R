package go_Weather_ITUR

import (
	"log"
	"time"
)

type StationSystem struct {
	BasicSystem
}

// type HourlyData struct {
// 	Precipitation   []float64 json:"precipitation"
// 	SurfacePressure []float64 json:"surface_pressure"
// 	Temperature2m   []float64 json:"temperature_2m"
// }

// type ForecastResponse struct {
// 	Hourly struct {
// 		Temperature2m   []float64 json:"temperature_2m"
// 		Precipitation   []float64 json:"precipitation"
// 		SurfacePressure []float64 json:"surface_pressure"
// 	} json:"hourly"
// }

// type StationWeather struct {
// 	Lat    float64
// 	Lon    float64
// 	Hourly HourlyData json:"hourly"
// }

func NewStationSystem(interval int64) *StationSystem {
	return &StationSystem{
		BasicSystem: BasicSystem{
			interval: interval,
			name:     "StationSystem",
		},
	}
}

// func fetchWeatherFromAPI(lat, lon float64) (t, precip, pressure float64) {
// 	/**
// 	temperature_2m: C
// 	precipitation: mm/h
// 	surface_pressure: hPa
// 	*/
// 	log.Printf("Fetching weather data...")
// 	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.2f&longitude=%.2f&hourly=temperature_2m,precipitation,surface_pressure", lat, lon)
// 	// 发送 GET 请求
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		log.Fatalf("Error making request: %v", err)
// 	}
// 	defer resp.Body.Close()
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalf("Error reading response body: %v", err)
// 	}
// 	// 解析 JSON 数据
// 	var forecast ForecastResponse
// 	err = json.Unmarshal(body, &forecast)

// 	// fmt.Printf("parsed forecast: %+v\n", forecast)
// 	if err != nil {
// 		log.Fatalf("Error unmarshalling JSON: %v", err)
// 	}
// 	currentTime := time.Now()
// 	hours, _, _ := currentTime.Clock()
// 	t = forecast.Hourly.Temperature2m[hours]
// 	precip = forecast.Hourly.Precipitation[hours]
// 	pressure = forecast.Hourly.SurfacePressure[hours]
// 	return t, precip, pressure
// }

// func loadWeatherData(filePath string) ([]StationWeather, error) {
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()
// 	var data []StationWeather
// 	decoder := json.NewDecoder(file)
// 	if err := decoder.Decode(&data); err != nil {
// 		return nil, err
// 	}
// 	return data, nil
// }

// func getWeatherFromFile(lat, lon float64) (t, precip, pressure float64) {
// 	weatherData, err := loadWeatherData("data/weather_data.json")
// 	if err != nil {
// 		log.Fatalf("Error loading weather data: %v", err)
// 	}
// 	hours := time.Now().Hour()

// 	for _, wd := range weatherData {
// 		if wd.Lat == lat && wd.Lon == lon {
// 			return wd.Hourly.Temperature2m[hours], wd.Hourly.Precipitation[hours], wd.Hourly.SurfacePressure[hours]
// 		}
// 	}
// 	return 0, 0, 0
// }

func getWeatherBasedOnTerminal(stationPos *StationPositionComponent) EnvironmentIndex {
	return EnvironmentIndex{Temperature2m: 10.0, Precipitation: 0.0, Pressure: 1010.0} // 返回一个环境指数，实际应用中应根据站点位置获取真实数据
}
func (s *StationSystem) Update(dt int64, cm *ComponentManager, w *World, t time.Time) {
	log.Printf("StationSystem update...")
	startTime := time.Now()
	count := 0
	for i, link := range cm.LinkComponents {
		idx := link.TargetID

		stationPos := &cm.StationPositionComponents[idx]
		EnvironmentIdx := getWeatherBasedOnTerminal(stationPos)
		cm.LinkComponents[i].EnvironmentIdx = EnvironmentIdx
		count++
	}

	log.Printf("Station count: %d", count)
	log.Printf("StationSystem update time: %v", time.Since(startTime))

}

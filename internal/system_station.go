package go_Weather_ITUR

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type StationSystem struct {
	BasicSystem
}

type HourlyData struct {
	Precipitation   []float64 `json:"precipitation"`
	SurfacePressure []float64 `json:"surface_pressure"`
	Temperature2m   []float64 `json:"temperature_2m"`
}

type ForecastResponse struct {
	Hourly struct {
		Temperature2m   []float64 `json:"temperature_2m"`
		Precipitation   []float64 `json:"precipitation"`
		SurfacePressure []float64 `json:"surface_pressure"`
	} `json:"hourly"`
}

type StationWeather struct {
	Lat    float64
	Lon    float64
	Hourly HourlyData `json:"hourly"`
}

func NewStationSystem(interval int64) *StationSystem {
	return &StationSystem{
		BasicSystem: BasicSystem{
			interval: interval,
			name:     "StationSystem",
		},
	}
}

func fetchWeatherFromAPI(lat, lon float64) (t, precip, pressure float64) {
	/**
	temperature_2m: C
	precipitation: mm/h
	surface_pressure: hPa
	*/
	log.Printf("Fetching weather data...")
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.2f&longitude=%.2f&hourly=temperature_2m,precipitation,surface_pressure", lat, lon)
	// 发送 GET 请求
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	// 解析 JSON 数据
	var forecast ForecastResponse
	err = json.Unmarshal(body, &forecast)

	// fmt.Printf("parsed forecast: %+v\n", forecast)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}
	currentTime := time.Now()
	hours, _, _ := currentTime.Clock()
	t = forecast.Hourly.Temperature2m[hours]
	precip = forecast.Hourly.Precipitation[hours]
	pressure = forecast.Hourly.SurfacePressure[hours]
	return t, precip, pressure
}

func loadWeatherData(filePath string) ([]StationWeather, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var data []StationWeather
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func getWeatherFromFile(lat, lon float64) (t, precip, pressure float64) {
	weatherData, err := loadWeatherData("data/weather_data.json")
	if err != nil {
		log.Fatalf("Error loading weather data: %v", err)
	}
	hours := time.Now().Hour()

	for _, wd := range weatherData {
		if wd.Lat == lat && wd.Lon == lon {
			return wd.Hourly.Temperature2m[hours], wd.Hourly.Precipitation[hours], wd.Hourly.SurfacePressure[hours]
		}
	}
	return 0, 0, 0
}

func (s *StationSystem) Update(dt int64, cm *ComponentManager, w *World) {
	log.Printf("StationSystem update...")
	for entityID, weatherIndexComponent := range cm.WeatherIndexComponents {
		station, exists := cm.StationPositionComponents[entityID]
		if !exists {
			continue
		}

		/**
		先用读取文件的方式获取天气数据而不是API调用
		*/
		lat := float64(station.Lat)
		lon := float64(station.Lon)

		// fetchWeatherFromAPI(lat, lon)
		// t, precip, pressure := getWeatherFromFile(lat, lon)

		t := 30.0
		precip := 10.0
		pressure := 1013.25 + lat/lon

		weatherIndexComponent.T = float64(t)
		weatherIndexComponent.precipitation = float64(precip)
		weatherIndexComponent.P = float64(pressure)
		// log.Printf("GMT时刻：%d, 降水：%.2f mm/h, 温度：%.2f C, 气压：%.2f hPa", hours, weatherIndexComponent.precipitation, weatherIndexComponent.T, weatherIndexComponent.P)

		// weather.RainRate += 0.05 * float64(dt)
		// cm.WeatherIndexComponents[entityID] = weather
		// fmt.Printf("[StationSystem] Entity %d: rain=%.2f\n", entityID, weather.RainRate)
	}

}

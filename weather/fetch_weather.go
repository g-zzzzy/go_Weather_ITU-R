package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type HourlyData struct {
	Temperature2m   []float64 `json:"temperature_2m"`
	Precipitation   []float64 `json:"precipitation"`
	SurfacePressure []float64 `json:"surface_pressure"`
}

type StationWeather struct {
	Lat    float64
	Lon    float64
	Hourly HourlyData `json:"hourly"`
}

type ForecastResponse struct {
	Hourly HourlyData `json:"hourly"`
}

func main() {
	filename := "data/station_data5.txt"
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening station file: %v", err)
	}
	defer file.Close()

	var weatherData []StationWeather

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
			log.Printf("Error parsing coordinates: %v, %v", err1, err2)
			continue
		}
		weather := fetchWeather(lat, lon)
		weatherData = append(weatherData, StationWeather{
			Lat:    lat,
			Lon:    lon,
			Hourly: weather.Hourly,
		})
	}

	outFile, err := os.Create("data/weather_data.json")
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(weatherData); err != nil {
		log.Fatalf("Error encoding JSON: %v", err)
	}

	log.Printf("Fetched weather for %d stations and saved to data/weather_data.json", len(weatherData))

}

func fetchWeather(lat, lon float64) ForecastResponse {
	if lon > 180 {
		lon = lon - 360
	}
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.2f&longitude=%.2f&hourly=temperature_2m,precipitation,surface_pressure",
		lat, lon,
	)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	var forecast ForecastResponse
	if err := json.Unmarshal(body, &forecast); err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	log.Printf("Fetched weather for lat=%.2f lon=%.2f", lat, lon)
	return forecast
}

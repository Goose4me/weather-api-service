package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type WeatherService struct {
}

func NewWeatherService() *WeatherService {
	return &WeatherService{}
}

type WeatherResponse struct {
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
}

type WeatherData struct {
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
	Description string  `json:"description"`
}

const api_addres = "https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric"

func (ws *WeatherService) callWeatherAPI(city, apiKey string) (*WeatherResponse, error) {
	var result WeatherResponse

	url := fmt.Sprintf(api_addres, city, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("failed to fetch weather: %s\n", err.Error())

		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("API error: %s\n", string(body))

		return nil, fmt.Errorf("API error %s", string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("invalid response JSON: %s\n", err.Error())

		return nil, err
	}

	return &result, nil
}

func (ws *WeatherService) GetWeather(city string) (*WeatherData, error) {

	weatherResponse, err := ws.callWeatherAPI(city, os.Getenv("WEATHER_API"))
	if err != nil {

		return nil, err
	}

	var weatherData WeatherData

	weatherData.Temperature = weatherResponse.Main.Temp
	weatherData.Humidity = weatherResponse.Main.Humidity
	weatherData.Description = weatherResponse.Weather[0].Description

	return &weatherData, err
}

package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type WeatherService struct {
	client       HTTPClient
	apiKey       string
	weatherCache WeatherCacheInterface
}

type WeatherServiceInterface interface {
	GetWeather(city string) (*WeatherData, error)
}

type WeatherCacheInterface interface {
	Get(city string) (*WeatherData, bool)
	Set(city string, data *WeatherData)
}

func NewWeatherService(client HTTPClient, apiKey string, weatherCache WeatherCacheInterface) *WeatherService {
	if client == nil {
		client = &http.Client{}
	}

	return &WeatherService{
		client:       client,
		apiKey:       apiKey,
		weatherCache: weatherCache,
	}
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

var ErrCityNotFound = errors.New("city not found")

func (ws *WeatherService) callWeatherAPI(city string) (*WeatherResponse, error) {
	var result WeatherResponse

	var api_addres = os.Getenv("WEATHER_API_ADDRESS")

	url := fmt.Sprintf(api_addres, city, ws.apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ws.client.Do(req)
	if err != nil {
		log.Printf("failed to fetch weather: %s\n", err.Error())

		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("API error: %s\n", string(body))

		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrCityNotFound
		} else {
			return nil, fmt.Errorf("API error %s", string(body))
		}
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("invalid response JSON: %s\n", err.Error())

		return nil, err
	}

	return &result, nil
}

func (ws *WeatherService) GetWeather(city string) (*WeatherData, error) {
	// Check cache first
	if data, found := ws.weatherCache.Get(city); found {
		log.Printf("Cache hit for city: %s\n", city)
		return data, nil
	}

	// Fallback to external API
	weatherResponse, err := ws.callWeatherAPI(city)
	if err != nil {
		return nil, err
	}

	var weatherData WeatherData

	weatherData.Temperature = weatherResponse.Main.Temp
	weatherData.Humidity = weatherResponse.Main.Humidity
	weatherData.Description = weatherResponse.Weather[0].Description

	ws.weatherCache.Set(city, &weatherData)

	return &weatherData, err
}

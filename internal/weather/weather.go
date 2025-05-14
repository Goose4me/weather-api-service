package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

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

func GetWeather(city, apiKey string) (WeatherResponse, error) {
	var result WeatherResponse

	url := fmt.Sprintf(api_addres, city, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("failed to fetch weather: %s\n", err.Error())

		return result, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("API error: %s\n", string(body))

		return result, fmt.Errorf("API error %s", string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("invalid response JSON: %s\n", err.Error())

		return result, err
	}

	fmt.Printf("Temperature in %s: %.1f°C\n", city, result.Main.Temp)
	fmt.Printf("Humidity in %s: %d%%\n", city, result.Main.Humidity)
	fmt.Printf("Description in %s: %s\n", city, result.Weather[0].Description)

	return result, nil
}

func WeatherHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Unsupported method", http.StatusBadRequest)
		return
	}

	query := req.URL.Query()

	city := query.Get("city")
	if city == "" {
		http.Error(w, "City is empty", http.StatusNotFound)
		return
	}

	var weatherData WeatherData

	weatherResponse, err := GetWeather(city, os.Getenv("WEATHER_API"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	weatherData.Temperature = weatherResponse.Main.Temp
	weatherData.Humidity = weatherResponse.Main.Humidity
	weatherData.Description = weatherResponse.Weather[0].Description

	if err = json.NewEncoder(w).Encode(weatherData); err != nil {
		log.Printf("Encoding error %s", err.Error())

		http.Error(w, "Encoding error", http.StatusInternalServerError)
	}
}

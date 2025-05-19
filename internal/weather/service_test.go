package weather_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"weather-app/internal/weather"
)

// helper to temporarily override the environment variable
func withEnv(key, value string, fn func()) {
	original := os.Getenv(key)
	os.Setenv(key, value)
	defer os.Setenv(key, original)
	fn()
}

func TestGetWeather_Success(t *testing.T) {
	// Create test server that mimics OpenWeatherMap API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{
			"main": {"temp": 22.5, "humidity": 60},
			"weather": [{"description": "clear sky"}]
		}`)
	}))
	defer server.Close()

	original := os.Getenv("WEATHER_API_ADDRESS")
	api_addres := server.URL + "?q=%s&appid=%s"
	os.Setenv("WEATHER_API_ADDRESS", api_addres)
	defer os.Setenv("WEATHER_API_ADDRESS", original)

	withEnv("WEATHER_API", "dummy", func() {
		ws := weather.NewWeatherService(nil, "test_api")

		data, err := ws.GetWeather("Kyiv")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if data.Temperature != 22.5 || data.Humidity != 60 || data.Description != "clear sky" {
			t.Errorf("unexpected result: %+v", data)
		}
	})
}

func TestGetWeather_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"city not found"}`, http.StatusNotFound)
	}))
	defer server.Close()

	original := os.Getenv("WEATHER_API_ADDRESS")
	api_addres := server.URL + "?q=%s&appid=%s"
	os.Setenv("WEATHER_API_ADDRESS", api_addres)
	defer os.Setenv("WEATHER_API_ADDRESS", original)

	withEnv("WEATHER_API", "dummy", func() {
		ws := weather.NewWeatherService(nil, "test_api")

		_, err := ws.GetWeather("InvalidCity")
		if !errors.Is(err, weather.ErrCityNotFound) {
			t.Fatalf("expected ErrCityNotFound, got %v", err)
		}
	})
}

func TestGetWeather_BadJSON(t *testing.T) {
	// Simulate malformed JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `invalid-json`)
	}))
	defer server.Close()

	original := os.Getenv("WEATHER_API_ADDRESS")
	api_addres := server.URL + "?q=%s&appid=%s"
	os.Setenv("WEATHER_API_ADDRESS", api_addres)
	defer os.Setenv("WEATHER_API_ADDRESS", original)

	withEnv("WEATHER_API", "dummy", func() {
		ws := weather.NewWeatherService(nil, "test_api")
		_, err := ws.GetWeather("Kyiv")
		if err == nil {
			t.Fatal("expected error due to bad JSON, got nil")
		}
	})
}

func TestGetWeather_GenericIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"city not found"}`, http.StatusBadRequest)
	}))
	defer server.Close()

	original := os.Getenv("WEATHER_API_ADDRESS")
	api_addres := server.URL + "?q=%s&appid=%s"
	os.Setenv("WEATHER_API_ADDRESS", api_addres)
	defer os.Setenv("WEATHER_API_ADDRESS", original)

	withEnv("WEATHER_API", "dummy", func() {
		ws := weather.NewWeatherService(nil, "test_api")

		_, err := ws.GetWeather("InvalidCity")
		if err == nil {
			t.Fatal("expected error due to some generic issue, got nil")
		}
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestGetWeather_RequestIssue(t *testing.T) {
	fakeClient := &http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("simulated network failure")
		}),
	}

	ws := weather.NewWeatherService(fakeClient, "test_api")

	_, err := ws.GetWeather("Lviv")
	if err == nil {
		t.Fatal("expected error due to some generic issue, got nil")
	}
}

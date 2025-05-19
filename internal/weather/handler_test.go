package weather_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"weather-app/internal/weather"
)

type MockWeatherService struct {
	GetWeatherFunc func(city string) (*weather.WeatherData, error)
}

func (m *MockWeatherService) GetWeather(city string) (*weather.WeatherData, error) {
	return m.GetWeatherFunc(city)
}

func TestWeatherHandler_Success(t *testing.T) {
	mockSvc := &MockWeatherService{
		GetWeatherFunc: func(city string) (*weather.WeatherData, error) {
			return &weather.WeatherData{
				Temperature: 25.0,
				Humidity:    70,
				Description: "clear sky",
			}, nil
		},
	}

	handler := weather.NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/weather?city=Kyiv", nil)
	rec := httptest.NewRecorder()

	handler.Handler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}

	var data weather.WeatherData
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if data.Temperature != 25.0 || data.Humidity != 70 || data.Description != "clear sky" {
		t.Errorf("unexpected response data: %+v", data)
	}
}

func TestWeatherHandler_CityNotFound(t *testing.T) {
	mockSvc := &MockWeatherService{
		GetWeatherFunc: func(city string) (*weather.WeatherData, error) {
			return nil, weather.ErrCityNotFound
		},
	}

	handler := weather.NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/weather?city=Atlantis", nil)
	rec := httptest.NewRecorder()

	handler.Handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "city not found") {
		t.Errorf("expected 'city not found' error, got %s", rec.Body.String())
	}
}

func TestWeatherHandler_InternalError(t *testing.T) {
	mockSvc := &MockWeatherService{
		GetWeatherFunc: func(city string) (*weather.WeatherData, error) {
			return nil, errors.New("DB timeout")
		},
	}

	handler := weather.NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/weather?city=Kyiv", nil)
	rec := httptest.NewRecorder()

	handler.Handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), weather.GenericErrorMsg) {
		t.Errorf("expected generic error message, got %s", rec.Body.String())
	}
}

func TestWeatherHandler_MissingCity(t *testing.T) {
	mockSvc := &MockWeatherService{}

	handler := weather.NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/weather", nil)
	rec := httptest.NewRecorder()

	handler.Handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "City parameter is empty") {
		t.Errorf("expected missing city error, got %s", rec.Body.String())
	}
}

func TestWeatherHandler_UnsupportedMethod(t *testing.T) {
	mockSvc := &MockWeatherService{}

	handler := weather.NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/weather?city=Kyiv", nil)
	rec := httptest.NewRecorder()

	handler.Handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "Unsupported method") {
		t.Errorf("expected method error, got %s", rec.Body.String())
	}
}

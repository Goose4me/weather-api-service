package cache_test

import (
	"testing"
	"time"
	"weather-app/internal/weather"
	"weather-app/internal/weather/cache"
)

func TestWeatherCache_SetAndGet(t *testing.T) {
	weatherCache := cache.NewWeatherCache(1 * time.Minute)

	expected := &weather.WeatherData{
		Temperature: 20.5,
		Humidity:    80,
		Description: "Cloudy",
	}

	weatherCache.Set("Kyiv", expected)

	result, ok := weatherCache.Get("Kyiv")
	if !ok {
		t.Fatal("expected weatherCache hit but got miss")
	}

	if result.Temperature != expected.Temperature ||
		result.Humidity != expected.Humidity ||
		result.Description != expected.Description {
		t.Errorf("got %+v, want %+v", result, expected)
	}
}

func TestWeatherCache_ExpiredEntry(t *testing.T) {
	weatherCache := cache.NewWeatherCache(10 * time.Millisecond)

	weatherCache.Set("Lviv", &weather.WeatherData{Temperature: 18})
	time.Sleep(20 * time.Millisecond)

	_, ok := weatherCache.Get("Lviv")
	if ok {
		t.Error("expected weatherCache miss due to expiration, got hit")
	}
}

func TestWeatherCache_MissingEntry(t *testing.T) {
	weatherCache := cache.NewWeatherCache(1 * time.Minute)

	_, ok := weatherCache.Get("Odesa")
	if ok {
		t.Error("expected weatherCache miss for unset key")
	}
}

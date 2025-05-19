package cache

import (
	"sync"
	"time"
	"weather-app/internal/weather"
)

type CacheItem struct {
	Data      *weather.WeatherData
	ExpiresAt time.Time
}

type WeatherCache struct {
	mu    sync.RWMutex
	store map[string]*CacheItem
	ttl   time.Duration
}

func NewWeatherCache(ttl time.Duration) *WeatherCache {
	return &WeatherCache{
		store: make(map[string]*CacheItem),
		ttl:   ttl,
	}
}

func (c *WeatherCache) Get(city string) (*weather.WeatherData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.store[city]
	if !exists || time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	return item.Data, true
}

func (c *WeatherCache) Set(city string, data *weather.WeatherData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[city] = &CacheItem{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

package main

import (
	"log"
	"net/http"
	"weather-app/internal/subscription"
	"weather-app/internal/weather"
)

func main() {
	log.Println("Starting Weather Server...")

	http.HandleFunc("/api/weather", weather.WeatherHandler)
	http.HandleFunc("/api/subscribe", subscription.SubscriptionHandler)

	addr := ":8080"
	log.Printf("Server running at %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

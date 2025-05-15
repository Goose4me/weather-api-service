package main

import (
	"log"
	"net/http"
	"weather-app/internal/database"
	"weather-app/internal/subscription"
	"weather-app/internal/weather"
)

func main() {
	log.Println("Starting Weather Server...")

	db, err := database.InitDB()

	if err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}

	subService := subscription.NewSubscriptionService(db)
	subHandler := subscription.NewHandler(subService)

	http.HandleFunc("/api/weather", weather.WeatherHandler)
	http.HandleFunc("/api/subscribe", subHandler.Handler)

	addr := ":8080"
	log.Printf("Server running at %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

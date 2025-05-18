package main

import (
	"log"
	"net/http"
	"weather-app/internal/database"
	"weather-app/internal/subscription"
	"weather-app/internal/weather"
)

// Use for cases like "/api/confirm" instead "/api/confirm/"
func wrongQueryHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "404 page not found", http.StatusNotFound)
}

func main() {
	log.Println("Starting Weather Server...")

	db, err := database.InitDB()

	if err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}

	subService := subscription.NewSubscriptionService(db)
	subHandler := subscription.NewHandler(subService)

	weatherService := weather.NewWeatherService()
	weatherHandler := weather.NewHandler(weatherService)

	// Weather service
	http.HandleFunc("/api/weather", weatherHandler.Handler)

	// Subscription service
	http.HandleFunc("/api/subscribe", subHandler.SubscribeHandler)
	http.HandleFunc("/api/confirm/", subHandler.ConfirmHandler)
	http.HandleFunc("/api/confirm", wrongQueryHandler)
	http.HandleFunc("/api/unsubscribe/", subHandler.UnsubscribeHandler)
	http.HandleFunc("/api/unsubscribe", wrongQueryHandler)

	addr := ":8080"
	log.Printf("Server running at %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

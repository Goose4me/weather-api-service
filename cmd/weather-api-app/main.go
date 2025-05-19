package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/cors"

	"weather-app/internal/database"
	"weather-app/internal/database/repository"
	"weather-app/internal/mail"
	"weather-app/internal/subscription"
	"weather-app/internal/weather"
	"weather-app/internal/weather/cache"
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

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	APIKey := os.Getenv("MAILSENDER_API_KEY")
	msw := mail.NewMailSenderWrapper(APIKey)
	mailService := mail.NewMailService(userRepo, msw)

	subService := subscription.NewSubscriptionService(userRepo, tokenRepo, mailService)
	subHandler := subscription.NewHandler(subService)

	weatherCache := cache.NewWeatherCache(time.Minute * 30)
	weatherService := weather.NewWeatherService(nil, os.Getenv("WEATHER_API"), weatherCache)
	weatherHandler := weather.NewHandler(weatherService)

	// Weather service
	http.HandleFunc("/api/weather", weatherHandler.Handler)

	// Subscription service
	http.HandleFunc("/api/subscribe", subHandler.SubscribeHandler)
	http.HandleFunc("/api/confirm/", subHandler.ConfirmHandler)
	http.HandleFunc("/api/confirm", wrongQueryHandler)
	http.HandleFunc("/api/unsubscribe/", subHandler.UnsubscribeHandler)
	http.HandleFunc("/api/unsubscribe", wrongQueryHandler)

	// fix CORS problem
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})
	handlerWithCORS := c.Handler(http.DefaultServeMux)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handlerWithCORS, // or your custom router
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server started on %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-sigChan
	log.Println("Shutdown signal received")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown gracefully
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly")
}

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"weather-app/internal/database"
	"weather-app/internal/database/repository"
	"weather-app/internal/mail"
	"weather-app/internal/scheduler"
)

const dailyUpdateHour = 12

func main() {
	log.Println("Starting Mail Sending Service...")

	db, err := database.InitDB()

	if err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}

	userRepo := repository.NewUserRepository(db)

	APIKey := os.Getenv("MAILSENDER_API_KEY")
	msw := mail.NewMailSenderWrapper(APIKey)
	mailService := mail.NewMailService(userRepo, msw)

	ctx, cancel := context.WithCancel(context.Background())

	done := scheduler.Start(ctx, time.Hour, func(currentTime time.Time) {
		regularUpdate := make(chan struct{})

		go func() {
			defer close(regularUpdate)
			log.Println("Regular update started")

			err := mailService.SendWeatherUpdate(mail.Hourly)

			if err != nil {
				log.Printf("Regular update error: %s\n", err.Error())
			}

		}()

		if currentTime.Hour() == dailyUpdateHour && currentTime.Minute() == 0 {
			log.Println("Daily update started")

			err := mailService.SendWeatherUpdate(mail.Daily)

			if err != nil {
				log.Printf("Daily update error: %s\n", err.Error())
			}
		}

		<-regularUpdate
	})

	// Handle SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Scheduler started. Press Ctrl+C to stop...")

	<-sigChan
	log.Println("Stopping scheduler...")
	cancel() // Cancel the scheduler
	<-done   // Wait for the scheduler to be done

	log.Println("Scheduler stopped cleanly.")
}

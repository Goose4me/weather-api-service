package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"weather-app/internal/database"
	"weather-app/internal/mail"
	"weather-app/internal/scheduler"
)

func main() {
	log.Println("Starting Mail Sending Service...")

	ctx, cancel := context.WithCancel(context.Background())

	db, err := database.InitDB()

	if err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}

	mailService := mail.NewMailService(db)

	done := scheduler.Start(ctx, time.Second*30, func() {
		log.Println("I executed")
		err := mailService.SendWeatherUpdate(mail.Hourly)

		if err != nil {
			log.Println(err.Error())
		}
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

package scheduler

import (
	"context"
	"log"
	"time"
)

func Start(ctx context.Context, interval time.Duration, task func()) (done chan struct{}) {
	done = make(chan struct{})

	go func() {
		defer close(done)

		for {
			now := time.Now()
			next := now.Truncate(interval).Add(interval)
			sleepDuration := time.Until(next)

			log.Printf("Waiting until %s (every %v)", next.Format("15:04:05"), interval)

			timer := time.NewTimer(sleepDuration)
			select {
			case <-ctx.Done():
				timer.Stop()
				log.Println("Scheduler cancelled before next execution.")
				return

			case <-timer.C:
				log.Println("Running scheduled task")
				task()
			}
		}
	}()

	return done
}

package scheduler_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
	"weather-app/internal/scheduler"
)

func TestScheduler_ExecutesTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var callCount atomic.Int32

	// task increments counter
	task := func(t time.Time) {
		callCount.Add(1)
		cancel() // stop after first call
	}

	done := scheduler.Start(ctx, 10*time.Millisecond, task)

	select {
	case <-done:
		// check if task was called
		if callCount.Load() < 1 {
			t.Errorf("expected task to be called at least once")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout: scheduler did not finish")
	}
}

func TestScheduler_CancelBeforeRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var called atomic.Bool
	task := func(t time.Time) {
		called.Store(true)
	}

	done := scheduler.Start(ctx, 100*time.Millisecond, task)

	cancel() // cancel immediately

	select {
	case <-done:
		if called.Load() {
			t.Errorf("expected task not to be called")
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("timeout: scheduler did not exit after cancel")
	}
}

package worker

import (
	"context"
	"time"
)

type Sleep func(context.Context, time.Duration) error

// Retry runs a bounded operation with exponential backoff. Callers persist the
// failed job/event separately, so no provider payload is held in process logs.
func Retry(ctx context.Context, attempts int, base time.Duration, sleep Sleep, operation func(context.Context) error) error {
	if attempts < 1 {
		attempts = 1
	}
	var err error
	for attempt := 0; attempt < attempts; attempt++ {
		if err = operation(ctx); err == nil {
			return nil
		}
		if attempt+1 == attempts {
			return err
		}
		delay := base << attempt
		if err = sleep(ctx, delay); err != nil {
			return err
		}
	}
	return err
}
func ContextSleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

package worker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryBacksOffWithoutSleepingInTest(t *testing.T) {
	calls := 0
	var delays []time.Duration
	err := Retry(context.Background(), 3, time.Second, func(_ context.Context, d time.Duration) error { delays = append(delays, d); return nil }, func(context.Context) error {
		calls++
		if calls < 3 {
			return errors.New("temporary")
		}
		return nil
	})
	if err != nil || calls != 3 || len(delays) != 2 || delays[0] != time.Second || delays[1] != 2*time.Second {
		t.Fatalf("calls=%d delays=%v err=%v", calls, delays, err)
	}
}

package indexer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func noopSleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func TestWithRetry(t *testing.T) {
	app := &App{sleep: noopSleep}

	t.Run("succeeds on first attempt, no retries", func(t *testing.T) {
		callCount := 0
		fn := func() error {
			callCount++
			return nil
		}

		err := app.withRetry(context.Background(), fn)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if callCount != 1 {
			t.Errorf("callCount = %d, want 1", callCount)
		}
	})

	t.Run("succeeds after two failures", func(t *testing.T) {
		callCount := 0
		fn := func() error {
			callCount++
			if callCount < 3 {
				return errors.New("simulated failure")
			}
			return nil
		}

		err := app.withRetry(context.Background(), fn)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if callCount != 3 {
			t.Errorf("callCount = %d, want 3", callCount)
		}
	})

	t.Run("fails after exhausting retries, returns last error", func(t *testing.T) {
		callCount := 0
		fn := func() error {
			callCount++
			return fmt.Errorf("failure #%d", callCount)
		}

		err := app.withRetry(context.Background(), fn)

		wantAttempts := int(MaxJobRetries) + 1
		if callCount != wantAttempts {
			t.Errorf("callCount = %d, want %d", callCount, wantAttempts)
		}
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		wantMsg := fmt.Sprintf("failure #%d", wantAttempts)
		if !strings.Contains(err.Error(), wantMsg) {
			t.Errorf("error = %v, want it to contain %q", err, wantMsg)
		}
	})

	t.Run("returns immediately when context already cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		callCount := 0
		fn := func() error {
			callCount++
			return errors.New("simulated failure")
		}

		err := app.withRetry(ctx, fn)

		if err == nil {
			t.Fatal("expected error due to cancelled context, got nil")
		}
		if callCount != 1 {
			t.Errorf("callCount = %d, want 1 (should not retry after cancellation)", callCount)
		}
	})
}

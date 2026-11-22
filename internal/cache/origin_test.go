package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestOriginFetch(t *testing.T) {
	origin := NewDemoOrigin()

	started := time.Now()

	value, err := origin.Fetch(
		context.Background(),
		"/video/a",
	)

	if err != nil {
		t.Fatal(err)
	}

	if value != "Video content A" {
		t.Errorf("Unexpected content: %s", value)
	}

	t.Logf("Origin fetch took %v", time.Since(started))
}

func TestOriginNotFound(t *testing.T) {
	origin := NewDemoOrigin()

	_, err := origin.Fetch(
		context.Background(),
		"/video/unknown",
	)

	if !errors.Is(err, ErrOriginNotFound) {
		t.Errorf("Expected ErrOriginNotFound, got %v", err)
	}
}

func TestOriginCancellation(t *testing.T) {
	origin := NewDemoOrigin()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	cancel()

	_, err := origin.Fetch(ctx, "/video/a")

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected cancellation, got %v", err)
	}
}

func TestOriginDeadline(t *testing.T) {
	origin := NewDemoOrigin()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Millisecond,
	)
	defer cancel()

	_, err := origin.Fetch(ctx, "/video/a")

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected deadline exceeded, got %v", err)
	}
}

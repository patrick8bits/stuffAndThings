package cache

import (
	"context"
	"errors"
	"testing"
)

func TestServiceMissThenHit(t *testing.T) {
	service, err := NewDemoService(2)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	first, err := service.Fetch(ctx, "/video/a")
	if err != nil {
		t.Fatal(err)
	}

	second, err := service.Fetch(ctx, "/video/a")
	if err != nil {
		t.Fatal(err)
	}

	if first.Source != "origin" {
		t.Errorf("Expected origin, got %s", first.Source)
	}

	if second.Source != "cache" {
		t.Errorf("Expected cache, got %s", second.Source)
	}

	if first.Content != second.Content {
		t.Error("Cached content does not match origin")
	}

	stats := service.Snapshot()

	if stats.Hits != 1 || stats.Misses != 1 {
		t.Errorf("Unexpected statistics: %+v", stats)
	}

	t.Logf("Origin: %d ms", first.LatencyMS)
	t.Logf("Cache: %d ms", second.LatencyMS)
	t.Logf("Statistics: %+v", stats)
}

func TestServiceEviction(t *testing.T) {
	service, err := NewDemoService(2)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Populate both cache slots.
	for _, key := range []string{"/video/a", "/video/b"} {
		if _, err := service.Fetch(ctx, key); err != nil {
			t.Fatal(err)
		}
	}

	// Make A the most recently used.
	if _, err := service.Fetch(ctx, "/video/a"); err != nil {
		t.Fatal(err)
	}

	// C should cause B to be evicted.
	if _, err := service.Fetch(ctx, "/video/c"); err != nil {
		t.Fatal(err)
	}

	stats := service.Snapshot()

	if stats.Evictions != 1 {
		t.Errorf("Expected 1 eviction, got %d",
			stats.Evictions)
	}

	// B must now come from the origin again.
	result, err := service.Fetch(ctx, "/video/b")
	if err != nil {
		t.Fatal(err)
	}

	if result.Source != "origin" {
		t.Errorf("Expected origin fetch for evicted B")
	}

	t.Logf("Statistics after re-fetch: %+v",
		service.Snapshot())
}

func TestServiceErrors(t *testing.T) {
	if _, err := NewDemoService(0); err == nil {
		t.Error("Expected invalid capacity error")
	}

	service, err := NewDemoService(2)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Fetch(
		context.Background(),
		"/video/unknown",
	)

	if !errors.Is(err, ErrOriginNotFound) {
		t.Errorf("Expected origin not found, got %v", err)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err = service.Fetch(ctx, "/video/a")

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected cancellation, got %v", err)
	}
}

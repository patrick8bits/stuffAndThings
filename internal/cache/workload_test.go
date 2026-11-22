package cache

import (
	"context"
	"errors"
	"math"
	"testing"
)

func TestDemoWorkload(t *testing.T) {
	report, err := RunDemoWorkload(
		context.Background(),
		3,
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(report.Requests) != 12 {
		t.Errorf("Expected 12 requests, got %d",
			len(report.Requests))
	}

	if report.Stats.Hits != 7 {
		t.Errorf("Expected 7 hits, got %d",
			report.Stats.Hits)
	}

	if report.Stats.Misses != 5 {
		t.Errorf("Expected 5 misses, got %d",
			report.Stats.Misses)
	}

	if report.Stats.Evictions != 2 {
		t.Errorf("Expected 2 evictions, got %d",
			report.Stats.Evictions)
	}

	if report.Stats.Size != 3 {
		t.Errorf("Expected 3 cached objects, got %d",
			report.Stats.Size)
	}

	expectedRate := 7.0 / 12.0 * 100

	if math.Abs(report.HitRatePct-expectedRate) > 0.001 {
		t.Errorf("Unexpected hit rate: %.2f%%",
			report.HitRatePct)
	}

	// The first A request must fetch from origin.
	if report.Requests[0].Source != "origin" {
		t.Error("Expected first A request to miss")
	}

	// A should have been evicted by request D.
	if report.Requests[10].Source != "origin" {
		t.Error("Expected A to be fetched again after eviction")
	}

	// D should still be cached for the final request.
	if report.Requests[11].Source != "cache" {
		t.Error("Expected final D request to hit")
	}

	t.Logf("Requests: %d", len(report.Requests))
	t.Logf("Hits: %d", report.Stats.Hits)
	t.Logf("Misses: %d", report.Stats.Misses)
	t.Logf("Evictions: %d", report.Stats.Evictions)
	t.Logf("Hit ratio: %.1f%%", report.HitRatePct)
	t.Logf("Total execution: %d ms", report.TotalMS)
}

func TestDemoWorkloadErrors(t *testing.T) {
	_, err := RunDemoWorkload(
		context.Background(),
		0,
	)

	if err == nil {
		t.Error("Expected invalid capacity error")
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err = RunDemoWorkload(ctx, 3)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected cancellation, got %v", err)
	}
}

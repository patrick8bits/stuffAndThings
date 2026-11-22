package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrick8bits/stuffAndThings/internal/cache"
)

func TestCacheWorkloadHandler(t *testing.T) {
	handler := newCacheWorkloadHandler()

	tests := []struct {
		name          string
		method        string
		url           string
		expectedHTTP  int
		expectedCap   int
		expectedHits  int
		expectedMiss  int
		expectedEvict int
	}{
		{
			"Default capacity",
			"GET",
			"/api/cache/workload",
			200, 3, 7, 5, 2,
		},
		{
			"Larger cache",
			"GET",
			"/api/cache/workload?capacity=5",
			200, 5, 8, 4, 0,
		},
		{
			"Zero capacity",
			"GET",
			"/api/cache/workload?capacity=0",
			400, 0, 0, 0, 0,
		},
		{
			"Capacity too large",
			"GET",
			"/api/cache/workload?capacity=6",
			400, 0, 0, 0, 0,
		},
		{
			"Invalid capacity",
			"GET",
			"/api/cache/workload?capacity=abc",
			400, 0, 0, 0, 0,
		},
		{
			"Unsupported method",
			"POST",
			"/api/cache/workload",
			405, 0, 0, 0, 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(
				tt.method, tt.url, nil,
			)

			recorder := httptest.NewRecorder()
			handler(recorder, request)

			if recorder.Code != tt.expectedHTTP {
				t.Fatalf("Expected HTTP %d, got %d",
					tt.expectedHTTP, recorder.Code)
			}

			if tt.expectedHTTP != http.StatusOK {
				return
			}

			var report cache.WorkloadReport

			if err := json.NewDecoder(recorder.Body).
				Decode(&report); err != nil {
				t.Fatal(err)
			}

			if len(report.Requests) != 12 {
				t.Errorf("Expected 12 requests, got %d",
					len(report.Requests))
			}

			stats := report.Stats

			if stats.Capacity != tt.expectedCap ||
				stats.Hits != tt.expectedHits ||
				stats.Misses != tt.expectedMiss ||
				stats.Evictions != tt.expectedEvict {

				t.Errorf("Unexpected cache statistics: %+v",
					stats)
			}

			t.Logf("Capacity: %d, hits: %d, misses: %d",
				stats.Capacity, stats.Hits, stats.Misses)
			t.Logf("Hit ratio: %.1f%%", report.HitRatePct)
		})
	}
}

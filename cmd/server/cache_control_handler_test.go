package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrick8bits/stuffAndThings/internal/cache"
)

func TestCacheControlHandlers(t *testing.T) {
	service, err := cache.NewDemoService(3)
	if err != nil {
		t.Fatal(err)
	}

	statsHandler := newCacheStatsHandler(service)
	resetHandler := newCacheResetHandler(service)

	call := func(handler http.HandlerFunc, method, path string) *httptest.ResponseRecorder {
		t.Helper()

		request := httptest.NewRequest(method, path, nil)
		recorder := httptest.NewRecorder()

		handler(recorder, request)

		return recorder
	}

	decodeStats := func(recorder *httptest.ResponseRecorder) cache.Stats {
		t.Helper()

		var response struct {
			Stats cache.Stats `json:"stats"`
		}

		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}

		return response.Stats
	}

	// Verify the initial cache is empty.
	initial := call(statsHandler, http.MethodGet, "/api/cache/stats")

	if initial.Code != http.StatusOK {
		t.Fatalf("Initial stats returned HTTP %d", initial.Code)
	}

	stats := decodeStats(initial)

	if stats.Size != 0 || stats.Capacity != 3 {
		t.Errorf("Unexpected initial stats: %+v", stats)
	}

	// Populate the cache by requesting an object.
	result, err := service.Fetch(context.Background(), "/video/a")
	if err != nil {
		t.Fatal(err)
	}

	if result.Source != "origin" {
		t.Fatalf("Expected initial origin fetch")
	}

	// Confirm statistics changed.
	populated := call(statsHandler, http.MethodGet, "/api/cache/stats")
	stats = decodeStats(populated)

	if stats.Size != 1 || stats.Misses != 1 {
		t.Errorf("Unexpected populated stats: %+v", stats)
	}

	// Reset using POST.
	reset := call(resetHandler, http.MethodPost, "/api/cache/reset")

	if reset.Code != http.StatusOK {
		t.Fatalf("Reset returned HTTP %d", reset.Code)
	}

	stats = decodeStats(reset)

	if stats.Capacity != 3 ||
		stats.Size != 0 ||
		stats.Hits != 0 ||
		stats.Misses != 0 ||
		stats.Evictions != 0 {
		t.Errorf("Reset did not clear cache: %+v", stats)
	}

	// Previously cached objects must now miss.
	result, err = service.Fetch(context.Background(), "/video/a")
	if err != nil {
		t.Fatal(err)
	}

	if result.Source != "origin" {
		t.Errorf("Expected origin fetch after reset")
	}

	// Verify unsupported methods are rejected.
	invalidStats := call(statsHandler, http.MethodPost, "/api/cache/stats")
	invalidReset := call(resetHandler, http.MethodGet, "/api/cache/reset")

	if invalidStats.Code != http.StatusMethodNotAllowed {
		t.Error("Stats endpoint should reject POST")
	}

	if invalidReset.Code != http.StatusMethodNotAllowed {
		t.Error("Reset endpoint should reject GET")
	}

	t.Log("Cache statistics and reset verified")
}

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrick8bits/stuffAndThings/internal/cache"
)

func TestCacheHandler(t *testing.T) {
	service, err := cache.NewDemoService(3)
	if err != nil {
		t.Fatal(err)
	}

	handler := newCacheHandler(service)

	// First request should miss; second should hit.
	for _, expected := range []string{"origin", "cache"} {
		request := httptest.NewRequest(
			http.MethodGet,
			"/api/cache?key=/video/a",
			nil,
		)

		recorder := httptest.NewRecorder()
		handler(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d",
				recorder.Code)
		}

		var response struct {
			Result cache.FetchResult `json:"result"`
			Stats  cache.Stats       `json:"stats"`
		}

		if err := json.NewDecoder(recorder.Body).
			Decode(&response); err != nil {
			t.Fatal(err)
		}

		if response.Result.Source != expected {
			t.Errorf("Expected %s, got %s",
				expected, response.Result.Source)
		}

		t.Logf("Source: %s, latency: %d ms",
			response.Result.Source,
			response.Result.LatencyMS)
	}

	stats := service.Snapshot()

	if stats.Hits != 1 || stats.Misses != 1 {
		t.Errorf("Unexpected statistics: %+v", stats)
	}

	// Invalid content must be rejected.
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/cache?key=/invalid",
		nil,
	)

	recorder := httptest.NewRecorder()
	handler(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d",
			recorder.Code)
	}

	// Only GET requests are supported.
	request = httptest.NewRequest(
		http.MethodPost,
		"/api/cache?key=/video/a",
		nil,
	)

	recorder = httptest.NewRecorder()
	handler(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d",
			recorder.Code)
	}
}

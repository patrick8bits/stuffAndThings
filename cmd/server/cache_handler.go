package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/patrick8bits/stuffAndThings/internal/cache"
)

func newCacheHandler(service *cache.Service) http.HandlerFunc {
	// Limit the number of simultaneous API requests.
	slots := make(chan struct{}, 8)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method not allowed",
				http.StatusMethodNotAllowed)
			return
		}

		key := r.URL.Query().Get("key")

		// Only permit our five predefined demo objects.
		switch key {
		case "/video/a", "/video/b", "/video/c",
			"/video/d", "/video/e":
			// Valid key.
		default:
			http.Error(w, "Invalid content key",
				http.StatusBadRequest)
			return
		}

		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
		default:
			http.Error(w, "Cache demo busy",
				http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(
			r.Context(),
			2*time.Second,
		)
		defer cancel()

		result, err := service.Fetch(ctx, key)

		if err != nil {
			status := http.StatusBadGateway

			if errors.Is(err, context.DeadlineExceeded) {
				status = http.StatusGatewayTimeout
			}

			http.Error(w, "Cache fetch failed", status)
			return
		}

		response := struct {
			Result cache.FetchResult `json:"result"`
			Stats  cache.Stats       `json:"stats"`
		}{
			Result: result,
			Stats:  service.Snapshot(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Cache response encoding failed: %v", err)
		}
	}
}

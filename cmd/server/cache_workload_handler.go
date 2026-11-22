package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/patrick8bits/stuffAndThings/internal/cache"
)

func newCacheWorkloadHandler() http.HandlerFunc {
	// Allow at most two simultaneous workload runs.
	slots := make(chan struct{}, 2)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method not allowed",
				http.StatusMethodNotAllowed)
			return
		}

		// Default to a cache capacity of three objects.
		capacity := 3

		if raw := r.URL.Query().Get("capacity"); raw != "" {
			parsed, err := strconv.Atoi(raw)

			if err != nil || parsed < 1 || parsed > 5 {
				http.Error(w, "Capacity must be between 1 and 5",
					http.StatusBadRequest)
				return
			}

			capacity = parsed
		}

		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
		default:
			http.Error(w, "Workload service busy",
				http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(
			r.Context(),
			5*time.Second,
		)
		defer cancel()

		report, err := cache.RunDemoWorkload(ctx, capacity)

		if err != nil {
			status := http.StatusInternalServerError

			if errors.Is(err, context.DeadlineExceeded) {
				status = http.StatusGatewayTimeout
			}

			log.Printf("Cache workload failed: %v", err)
			http.Error(w, "Cache workload failed", status)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")

		if err := json.NewEncoder(w).Encode(report); err != nil {
			log.Printf("Workload response encoding failed: %v", err)
		}
	}
}

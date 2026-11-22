package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/patrick8bits/stuffAndThings/internal/health"
)

func newHealthHandler(
	client *http.Client,
	targets []health.Target,
) http.HandlerFunc {

	slots := make(chan struct{}, 2)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method not allowed",
				http.StatusMethodNotAllowed)
			return
		}

		mode := r.URL.Query().Get("mode")

		if mode != "sequential" && mode != "concurrent" {
			http.Error(w, "Invalid monitoring mode",
				http.StatusBadRequest)
			return
		}

		// Limit simultaneous monitoring requests.
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
		default:
			http.Error(w, "Health monitor busy",
				http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(
			r.Context(),
			5*time.Second,
		)

		defer cancel()

		var report health.Report

		switch mode {
		case "sequential":
			report = health.CheckSequential(
				ctx, client, targets,
			)

		case "concurrent":
			report = health.CheckConcurrent(
				ctx, client, targets,
			)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")

		if err := json.NewEncoder(w).Encode(report); err != nil {
			log.Printf("Failed to encode health report: %v", err)
		}
	}
}

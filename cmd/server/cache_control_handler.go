package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/patrick8bits/stuffAndThings/internal/cache"
)

// Encode a consistent statistics response.
func writeCacheStats(w http.ResponseWriter, stats cache.Stats) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	response := struct {
		Stats cache.Stats `json:"stats"`
	}{
		Stats: stats,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Cache stats encoding failed: %v", err)
	}
}

// GET /api/cache/stats
func newCacheStatsHandler(service *cache.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method not allowed",
				http.StatusMethodNotAllowed)
			return
		}

		writeCacheStats(w, service.Snapshot())
	}
}

// POST /api/cache/reset
func newCacheResetHandler(service *cache.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method not allowed",
				http.StatusMethodNotAllowed)
			return
		}

		if err := service.Reset(); err != nil {
			if errors.Is(err, cache.ErrResetBusy) {
				http.Error(w, "Cache is busy",
					http.StatusConflict)
				return
			}

			log.Printf("Cache reset failed: %v", err)
			http.Error(w, "Cache reset failed",
				http.StatusInternalServerError)
			return
		}

		writeCacheStats(w, service.Snapshot())
	}
}

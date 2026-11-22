package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/patrick8bits/stuffAndThings/internal/cache"
)

func newCacheOrderHandler(service *cache.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(
				w,
				"Method not allowed",
				http.StatusMethodNotAllowed,
			)
			return
		}

		response := struct {
			Order []string `json:"order"`
		}{
			Order: service.KeysLRU(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Cache order encoding failed: %v", err)
		}
	}
}

package main

import (
	"net/http"

	"github.com/patrick8bits/stuffAndThings/internal/cache"
)

func cacheSourceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")

		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	component := r.URL.Query().Get("component")

	source, found := cache.Source(component)

	if !found {
		http.Error(
			w,
			"Unknown cache component",
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	_, _ = w.Write([]byte(source))
}

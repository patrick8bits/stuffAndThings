package main

import (
	"net/http"

	"github.com/patrick8bits/stuffAndThings/internal/sorting"
)

func sortSourceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")

		http.Error(w, "Method not allowed",
			http.StatusMethodNotAllowed)
		return
	}

	algorithm := r.URL.Query().Get("algorithm")

	source, found := sorting.Source(algorithm)

	if !found {
		http.Error(w, "Unknown sorting algorithm",
			http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	w.Write([]byte(source))
}

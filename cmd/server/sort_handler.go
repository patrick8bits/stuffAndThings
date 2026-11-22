package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/patrick8bits/stuffAndThings/internal/sorting"
)

type SortRequest struct {
	Numbers []int `json:"numbers"`
}

type SortResponse struct {
	Bubble    sorting.Trace `json:"bubble"`
	Insertion sorting.Trace `json:"insertion"`
	Quick     sorting.Trace `json:"quick"`
}

func sortHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed",
			http.StatusMethodNotAllowed)
		return
	}

	// Limit the incoming request size.
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	var request SortRequest

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid JSON",
			http.StatusBadRequest)
		return
	}

	if decoder.Decode(&struct{}{}) != io.EOF {
		http.Error(w, "Unexpected request data",
			http.StatusBadRequest)
		return
	}

	// Validate the number of elements.
	if len(request.Numbers) < 2 ||
		len(request.Numbers) > 64 {
		http.Error(w, "Provide between 2 and 64 numbers",
			http.StatusBadRequest)
		return
	}

	// Limit values to our visualization range.
	for _, number := range request.Numbers {
		if number < 1 || number > 100 {
			http.Error(w, "Numbers must be between 1 and 100",
				http.StatusBadRequest)
			return
		}
	}

	response := SortResponse{
		Bubble:    sorting.TraceBubbleSort(request.Numbers),
		Insertion: sorting.TraceInsertionSort(request.Numbers),
		Quick:     sorting.TraceQuickSort(request.Numbers),
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

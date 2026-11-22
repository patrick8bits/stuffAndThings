package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestSortHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
	}{
		{
			"Valid request",
			http.MethodPost,
			`{"numbers":[42,17,95,8,63]}`,
			http.StatusOK,
		},
		{
			"Invalid JSON",
			http.MethodPost,
			`{"numbers":`,
			http.StatusBadRequest,
		},
		{
			"Empty slice",
			http.MethodPost,
			`{"numbers":[]}`,
			http.StatusBadRequest,
		},
		{
			"Too few numbers",
			http.MethodPost,
			`{"numbers":[42]}`,
			http.StatusBadRequest,
		},
		{
			"Value too high",
			http.MethodPost,
			`{"numbers":[42,101]}`,
			http.StatusBadRequest,
		},
		{
			"Value too low",
			http.MethodPost,
			`{"numbers":[42,0]}`,
			http.StatusBadRequest,
		},
		{
			"Unsupported HTTP method",
			http.MethodGet,
			"",
			http.StatusMethodNotAllowed,
		},
		{
			"Unexpected request data",
			http.MethodPost,
			`{"numbers":[42,17]} {}`,
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(
				tt.method,
				"/api/sort",
				strings.NewReader(tt.body),
			)

			request.Header.Set(
				"Content-Type", "application/json",
			)

			recorder := httptest.NewRecorder()

			sortHandler(recorder, request)

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d",
					tt.expectedStatus, recorder.Code)
			}
		})
	}
}

func TestSortHandlerResponse(t *testing.T) {
	body := `{"numbers":[42,17,95,8,63]}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/sort",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	sortHandler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("Unexpected status: %d", recorder.Code)
	}

	var response SortResponse

	err := json.NewDecoder(recorder.Body).Decode(&response)

	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expected := []int{8, 17, 42, 63, 95}

	for _, trace := range []struct {
		name  string
		steps [][]int
	}{
		{
			"Bubble Sort",
			[][]int{
				response.Bubble.Steps[0].Values,
				response.Bubble.Steps[len(response.Bubble.Steps)-1].Values,
			},
		},
		{
			"Insertion Sort",
			[][]int{
				response.Insertion.Steps[0].Values,
				response.Insertion.Steps[len(response.Insertion.Steps)-1].Values,
			},
		},
		{
			"Quick Sort",
			[][]int{
				response.Quick.Steps[0].Values,
				response.Quick.Steps[len(response.Quick.Steps)-1].Values,
			},
		},
	} {
		t.Run(trace.name, func(t *testing.T) {
			if !slices.Equal(trace.steps[0],
				[]int{42, 17, 95, 8, 63}) {
				t.Error("Incorrect initial frame")
			}

			if !slices.Equal(trace.steps[1], expected) {
				t.Errorf("Incorrect final frame: %v",
					trace.steps[1])
			}
		})
	}
}

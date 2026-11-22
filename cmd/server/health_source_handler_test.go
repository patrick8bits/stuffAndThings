package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthSourceHandler(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		component string
		expected  int
	}{
		{"Monitor", "GET", "monitor", 200},
		{"Simulator", "GET", "simulator", 200},
		{"Sequential", "GET", "sequential", 200},
		{"Concurrent", "GET", "concurrent", 200},
		{"Unknown component", "GET", "invalid", 404},
		{"Missing component", "GET", "", 404},
		{"Unsupported method", "POST", "monitor", 405},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(
				tt.method,
				"/api/health/source?component="+tt.component,
				nil,
			)

			recorder := httptest.NewRecorder()

			healthSourceHandler(recorder, request)

			if recorder.Code != tt.expected {
				t.Fatalf("Expected HTTP %d, got %d",
					tt.expected, recorder.Code)
			}

			if tt.expected == http.StatusOK {
				if !strings.Contains(
					recorder.Body.String(),
					"package health",
				) {
					t.Error("Expected Go source code")
				}

				if !strings.HasPrefix(
					recorder.Header().Get("Content-Type"),
					"text/plain",
				) {
					t.Error("Incorrect content type")
				}
			}
		})
	}
}

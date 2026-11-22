package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name           string
		httpStatus     int
		expectedStatus string
	}{
		{
			"Healthy server",
			http.StatusOK,
			"healthy",
		},
		{
			"Unhealthy server",
			http.StatusServiceUnavailable,
			"unhealthy",
		},
		{
			"Internal server error",
			http.StatusInternalServerError,
			"unhealthy",
		},
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			server := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(tt.httpStatus)
					},
				),
			)

			defer server.Close()

			target := Target{
				Name: "test-server",
				URL:  server.URL,
			}

			result := Check(
				context.Background(),
				client,
				target,
			)

			if result.Status != tt.expectedStatus {
				t.Errorf("Expected %s, got %s",
					tt.expectedStatus, result.Status)
			}

			if result.HTTPStatus != tt.httpStatus {
				t.Errorf("Expected HTTP %d, got %d",
					tt.httpStatus, result.HTTPStatus)
			}

			t.Logf("Status: %s, HTTP: %d, Latency: %d ms",
				result.Status,
				result.HTTPStatus,
				result.LatencyMS)
		})
	}
}

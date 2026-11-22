package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/patrick8bits/stuffAndThings/internal/health"
)

func TestHealthHandler(t *testing.T) {
	simulator := httptest.NewServer(
		http.HandlerFunc(health.DemoHandler),
	)

	defer simulator.Close()

	client := simulator.Client()
	client.Timeout = 2 * time.Second

	targets := health.DemoTargets(simulator.URL)

	handler := newHealthHandler(client, targets)

	tests := []struct {
		name           string
		method         string
		mode           string
		expectedStatus int
	}{
		{
			"Sequential monitoring",
			http.MethodGet,
			"sequential",
			http.StatusOK,
		},
		{
			"Concurrent monitoring",
			http.MethodGet,
			"concurrent",
			http.StatusOK,
		},
		{
			"Invalid mode",
			http.MethodGet,
			"invalid",
			http.StatusBadRequest,
		},
		{
			"Missing mode",
			http.MethodGet,
			"",
			http.StatusBadRequest,
		},
		{
			"Unsupported method",
			http.MethodPost,
			"concurrent",
			http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(
				tt.method,
				"/api/health?mode="+tt.mode,
				nil,
			)

			recorder := httptest.NewRecorder()

			handler(recorder, request)

			if recorder.Code != tt.expectedStatus {
				t.Fatalf("Expected HTTP %d, got %d",
					tt.expectedStatus, recorder.Code)
			}

			// Only successful requests contain reports.
			if tt.expectedStatus != http.StatusOK {
				return
			}

			var report health.Report

			err := json.NewDecoder(
				recorder.Body,
			).Decode(&report)

			if err != nil {
				t.Fatalf("Failed to decode report: %v", err)
			}

			if report.Mode != tt.mode {
				t.Errorf("Unexpected mode: %s", report.Mode)
			}

			if report.Total != 6 {
				t.Errorf("Expected 6 servers, got %d",
					report.Total)
			}

			if report.Healthy != 5 {
				t.Errorf("Expected 5 healthy, got %d",
					report.Healthy)
			}

			if report.Unhealthy != 1 {
				t.Errorf("Expected 1 unhealthy, got %d",
					report.Unhealthy)
			}

			if len(report.Results) != len(targets) {
				t.Fatalf("Incorrect number of results")
			}

			for i, result := range report.Results {
				if result.Name != targets[i].Name {
					t.Errorf("Incorrect server order at %d", i)
				}

				expectedStatus := http.StatusOK
				expectedHealth := "healthy"

				if result.Name == "cdn-lax-01" {
					expectedStatus = http.StatusServiceUnavailable
					expectedHealth = "unhealthy"
				}

				if result.HTTPStatus != expectedStatus {
					t.Errorf("Incorrect HTTP status for %s",
						result.Name)
				}

				if result.Status != expectedHealth {
					t.Errorf("Incorrect health status for %s",
						result.Name)
				}
			}

			t.Logf("%s completed in %d ms",
				report.Mode, report.TotalMS)
		})
	}
}

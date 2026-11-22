package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDemoFleet(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(DemoHandler),
	)

	defer server.Close()

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	targets := DemoTargets(server.URL)

	if len(targets) != 6 {
		t.Fatalf("Expected 6 targets, got %d", len(targets))
	}

	for _, target := range targets {
		t.Run(target.Name, func(t *testing.T) {
			result := Check(
				context.Background(),
				client,
				target,
			)

			expected := http.StatusOK

			if target.Name == "cdn-lax-01" {
				expected = http.StatusServiceUnavailable
			}

			if result.HTTPStatus != expected {
				t.Errorf("Expected HTTP %d, got %d",
					expected, result.HTTPStatus)
			}

			t.Logf("%s: HTTP %d, %s, %d ms",
				result.Name,
				result.HTTPStatus,
				result.Status,
				result.LatencyMS)
		})
	}
}

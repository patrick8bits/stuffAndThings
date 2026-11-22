package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckSequential(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(DemoHandler),
	)

	defer server.Close()

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	targets := DemoTargets(server.URL)

	report := CheckSequential(
		context.Background(),
		client,
		targets,
	)

	if report.Mode != "sequential" {
		t.Errorf("Unexpected mode: %s", report.Mode)
	}

	if report.Total != 6 {
		t.Errorf("Expected 6 servers, got %d", report.Total)
	}

	if report.Healthy != 5 {
		t.Errorf("Expected 5 healthy servers, got %d",
			report.Healthy)
	}

	if report.Unhealthy != 1 {
		t.Errorf("Expected 1 unhealthy server, got %d",
			report.Unhealthy)
	}

	if len(report.Results) != len(targets) {
		t.Fatalf("Incorrect number of results")
	}

	for i, result := range report.Results {
		if result.Name != targets[i].Name {
			t.Errorf("Unexpected result order at index %d", i)
		}
	}

	t.Logf("Mode: %s", report.Mode)
	t.Logf("Total servers: %d", report.Total)
	t.Logf("Healthy: %d", report.Healthy)
	t.Logf("Unhealthy: %d", report.Unhealthy)
	t.Logf("Total execution time: %d ms", report.TotalMS)
}

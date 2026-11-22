package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckConcurrent(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(DemoHandler),
	)

	defer server.Close()

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	targets := DemoTargets(server.URL)

	sequential := CheckSequential(
		context.Background(),
		client,
		targets,
	)

	concurrent := CheckConcurrent(
		context.Background(),
		client,
		targets,
	)

	if concurrent.Mode != "concurrent" {
		t.Errorf("Unexpected mode: %s", concurrent.Mode)
	}

	if concurrent.Total != 6 {
		t.Errorf("Expected 6 servers, got %d",
			concurrent.Total)
	}

	if concurrent.Healthy != 5 {
		t.Errorf("Expected 5 healthy servers, got %d",
			concurrent.Healthy)
	}

	if concurrent.Unhealthy != 1 {
		t.Errorf("Expected 1 unhealthy server, got %d",
			concurrent.Unhealthy)
	}

	if len(concurrent.Results) != len(targets) {
		t.Fatalf("Incorrect number of results")
	}

	for i, result := range concurrent.Results {
		expected := sequential.Results[i]

		if result.Name != expected.Name {
			t.Errorf("Incorrect result order at index %d", i)
		}

		if result.Status != expected.Status {
			t.Errorf("Status mismatch for %s", result.Name)
		}

		if result.HTTPStatus != expected.HTTPStatus {
			t.Errorf("HTTP status mismatch for %s", result.Name)
		}
	}

	t.Logf("Sequential: %d ms", sequential.TotalMS)
	t.Logf("Concurrent: %d ms", concurrent.TotalMS)

	if concurrent.TotalMS > 0 {
		speedup := float64(sequential.TotalMS) /
			float64(concurrent.TotalMS)

		t.Logf("Performance improvement: %.2fx", speedup)
	}
}

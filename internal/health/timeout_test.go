package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckTimeout(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				select {
				case <-r.Context().Done():
					return

				case <-time.After(2 * time.Second):
					w.WriteHeader(http.StatusOK)
				}
			},
		),
	)

	defer server.Close()

	target := Target{
		Name: "slow-server",
		URL:  server.URL,
	}

	t.Run("Client timeout", func(t *testing.T) {
		client := &http.Client{
			Timeout: 100 * time.Millisecond,
		}

		result := Check(
			context.Background(),
			client,
			target,
		)

		if result.Status != "timeout" {
			t.Errorf("Expected timeout, got %s",
				result.Status)
		}

		t.Logf("Status: %s, latency: %d ms",
			result.Status, result.LatencyMS)
	})

	t.Run("Context deadline", func(t *testing.T) {
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		ctx, cancel := context.WithTimeout(
			context.Background(),
			100*time.Millisecond,
		)

		defer cancel()

		result := Check(ctx, client, target)

		if result.Status != "timeout" {
			t.Errorf("Expected timeout, got %s",
				result.Status)
		}
	})

	t.Run("Cancelled context", func(t *testing.T) {
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		ctx, cancel := context.WithCancel(
			context.Background(),
		)

		cancel()

		result := Check(ctx, client, target)

		if result.Status != "cancelled" {
			t.Errorf("Expected cancelled, got %s",
				result.Status)
		}
	})
}

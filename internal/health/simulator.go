package health

import (
	"net/http"
	"strings"
	"time"
)

type simulatedServer struct {
	name   string
	delay  time.Duration
	status int
}

var demoFleet = []simulatedServer{
	{"cdn-den-01", 120 * time.Millisecond, http.StatusOK},
	{"cdn-den-02", 180 * time.Millisecond, http.StatusOK},
	{"cdn-chi-01", 250 * time.Millisecond, http.StatusOK},
	{"cdn-nyc-01", 340 * time.Millisecond, http.StatusOK},
	{"cdn-lax-01", 500 * time.Millisecond, http.StatusServiceUnavailable},
	{"cdn-sea-01", 210 * time.Millisecond, http.StatusOK},
}

// Generate health check targets for our simulated fleet.
func DemoTargets(baseURL string) []Target {
	targets := make([]Target, 0, len(demoFleet))

	baseURL = strings.TrimRight(baseURL, "/")

	for _, server := range demoFleet {
		targets = append(targets, Target{
			Name: server.name,
			URL:  baseURL + "/sim/" + server.name,
		})
	}

	return targets
}

// Simulate an HTTP response from a CDN server.
func DemoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed",
			http.StatusMethodNotAllowed)
		return
	}

	for _, server := range demoFleet {
		if r.URL.Path != "/sim/"+server.name {
			continue
		}

		timer := time.NewTimer(server.delay)
		defer timer.Stop()

		select {
		case <-timer.C:
			w.WriteHeader(server.status)

		case <-r.Context().Done():
			return
		}

		return
	}

	http.NotFound(w, r)
}

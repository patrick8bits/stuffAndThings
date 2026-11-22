package health

import (
	"context"
	"net/http"
	"time"
)

type Report struct {
	Mode      string   `json:"mode"`
	Results   []Result `json:"results"`
	TotalMS   int64    `json:"total_ms"`
	Total     int      `json:"total"`
	Healthy   int      `json:"healthy"`
	Unhealthy int      `json:"unhealthy"`
}

func CheckSequential(
	ctx context.Context,
	client *http.Client,
	targets []Target,
) Report {

	started := time.Now()

	results := make([]Result, 0, len(targets))

	for _, target := range targets {
		result := Check(ctx, client, target)

		results = append(results, result)
	}

	elapsed := time.Since(started)

	return buildReport("sequential", results, elapsed)
}

// Build a summary from the individual health check results.
func buildReport(
	mode string,
	results []Result,
	elapsed time.Duration,
) Report {

	report := Report{
		Mode:    mode,
		Results: results,
		TotalMS: elapsed.Milliseconds(),
		Total:   len(results),
	}

	for _, result := range results {
		if result.Status == "healthy" {
			report.Healthy++
		} else {
			report.Unhealthy++
		}
	}

	return report
}

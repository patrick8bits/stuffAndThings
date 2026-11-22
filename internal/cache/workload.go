package cache

import (
	"context"
	"fmt"
	"time"
)

// A repeatable workload used to demonstrate cache behavior.
var demoRequests = []string{
	"/video/a",
	"/video/b",
	"/video/c",
	"/video/a",
	"/video/b",
	"/video/c",
	"/video/a",
	"/video/b",
	"/video/c",
	"/video/d",
	"/video/a",
	"/video/d",
}

type WorkloadReport struct {
	Requests   []FetchResult `json:"requests"`
	Stats      Stats         `json:"stats"`
	HitRatePct float64       `json:"hit_rate_pct"`
	TotalMS    int64         `json:"total_ms"`
}

// RunDemoWorkload executes a fixed sequence against a fresh cache.
func RunDemoWorkload(
	ctx context.Context,
	capacity int,
) (WorkloadReport, error) {

	service, err := NewDemoService(capacity)
	if err != nil {
		return WorkloadReport{}, err
	}

	results := make([]FetchResult, 0, len(demoRequests))

	started := time.Now()

	for _, key := range demoRequests {
		result, err := service.Fetch(ctx, key)

		if err != nil {
			return WorkloadReport{}, fmt.Errorf(
				"fetch %s: %w",
				key,
				err,
			)
		}

		results = append(results, result)
	}

	stats := service.Snapshot()

	totalRequests := stats.Hits + stats.Misses

	hitRate := 0.0

	if totalRequests > 0 {
		hitRate = float64(stats.Hits) /
			float64(totalRequests) * 100
	}

	return WorkloadReport{
		Requests:   results,
		Stats:      stats,
		HitRatePct: hitRate,
		TotalMS:    time.Since(started).Milliseconds(),
	}, nil
}

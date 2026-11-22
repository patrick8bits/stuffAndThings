package health

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

type Target struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Result struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	HTTPStatus int    `json:"http_status"`
	LatencyMS  int64  `json:"latency_ms"`
}

func Check(
	ctx context.Context,
	client *http.Client,
	target Target,
) Result {

	result := Result{
		Name:   target.Name,
		Status: "unreachable",
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		target.URL,
		nil,
	)

	if err != nil {
		return result
	}

	started := time.Now()

	response, err := client.Do(request)

	result.LatencyMS = time.Since(started).Milliseconds()

	if err != nil {
		var netErr net.Error

		switch {
		case errors.Is(err, context.Canceled):
			result.Status = "cancelled"

		case errors.Is(err, context.DeadlineExceeded):
			result.Status = "timeout"

		case errors.As(err, &netErr) && netErr.Timeout():
			result.Status = "timeout"
		}

		return result
	}

	defer response.Body.Close()

	result.HTTPStatus = response.StatusCode

	if response.StatusCode == http.StatusOK {
		result.Status = "healthy"
	} else {
		result.Status = "unhealthy"
	}

	return result
}

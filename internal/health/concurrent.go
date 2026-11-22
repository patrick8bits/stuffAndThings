package health

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type indexedResult struct {
	index  int
	result Result
}

func CheckConcurrent(
	ctx context.Context,
	client *http.Client,
	targets []Target,
) Report {

	started := time.Now()

	var wg sync.WaitGroup

	resultsChan := make(chan indexedResult, len(targets))

	for i, target := range targets {
		wg.Add(1)

		go func(index int, target Target) {
			defer wg.Done()

			result := Check(ctx, client, target)

			resultsChan <- indexedResult{
				index:  index,
				result: result,
			}
		}(i, target)
	}

	wg.Wait()

	close(resultsChan)

	results := make([]Result, len(targets))

	for item := range resultsChan {
		results[item.index] = item.result
	}

	elapsed := time.Since(started)

	return buildReport("concurrent", results, elapsed)
}

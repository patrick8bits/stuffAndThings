package health

import (
	"sync"
	"testing"
	"time"
)

func TestConcurrencyBasics(t *testing.T) {
	started := time.Now()

	var wg sync.WaitGroup

	results := make(chan string, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			time.Sleep(500 * time.Millisecond)

			results <- "Task completed"
		}(i)
	}

	wg.Wait()

	close(results)

	completed := 0

	for result := range results {
		t.Log(result)
		completed++
	}

	if completed != 3 {
		t.Errorf("Expected 3 results, got %d", completed)
	}

	t.Logf("Total execution time: %v", time.Since(started))
}

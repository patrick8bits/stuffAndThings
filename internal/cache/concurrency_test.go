package cache

import (
	"context"
	"sync"
	"testing"
)

// Verify that concurrent Get and Put calls are safe.
func TestConcurrentLRUAccess(t *testing.T) {
	c, err := New(3)
	if err != nil {
		t.Fatal(err)
	}

	keys := []string{
		"/video/a",
		"/video/b",
		"/video/c",
		"/video/d",
		"/video/e",
	}

	const workers = 20
	const iterations = 50

	var wg sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for i := 0; i < iterations; i++ {
				key := keys[(id+i)%len(keys)]

				c.Put(key, "content")
				c.Get(key)
			}
		}(worker)
	}

	wg.Wait()

	stats := c.Snapshot()

	// Each worker called Get exactly 50 times.
	expectedGets := workers * iterations

	actualGets := stats.Hits + stats.Misses

	if actualGets != expectedGets {
		t.Errorf(
			"Expected %d lookups, got %d",
			expectedGets,
			actualGets,
		)
	}

	if stats.Size > stats.Capacity {
		t.Errorf(
			"Cache exceeded capacity: %+v",
			stats,
		)
	}

	t.Logf("Concurrent cache statistics: %+v", stats)
}

// Demonstrate simultaneous misses for the same object.
func TestConcurrentSameObject(t *testing.T) {
	service, err := NewDemoService(3)
	if err != nil {
		t.Fatal(err)
	}

	const requests = 10

	type outcome struct {
		result FetchResult
		err    error
	}

	outcomes := make(chan outcome, requests)
	start := make(chan struct{})

	var wg sync.WaitGroup

	for i := 0; i < requests; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			// Wait until all workers may begin together.
			<-start

			result, err := service.Fetch(
				context.Background(),
				"/video/a",
			)

			outcomes <- outcome{
				result: result,
				err:    err,
			}
		}()
	}

	// Release all workers at approximately the same time.
	close(start)

	wg.Wait()
	close(outcomes)

	originFetches := 0

	for item := range outcomes {
		if item.err != nil {
			t.Errorf("Fetch failed: %v", item.err)
			continue
		}

		if item.result.Content != "Video content A" {
			t.Errorf("Unexpected content: %s",
				item.result.Content)
		}

		if item.result.Source == "origin" {
			originFetches++
		}
	}

	stats := service.Snapshot()

	if stats.Hits+stats.Misses != requests {
		t.Errorf("Expected %d lookups, got %+v",
			requests, stats)
	}

	if stats.Size != 1 {
		t.Errorf("Expected one cached object, got %d",
			stats.Size)
	}

	t.Logf("Simultaneous client requests: %d", requests)
	t.Logf("Origin fetches: %d", originFetches)
	t.Logf("Cache statistics: %+v", stats)

	if originFetches != 1 {
		t.Errorf(
			"Expected exactly 1 origin fetch, got %d",
			originFetches,
		)
	}
}

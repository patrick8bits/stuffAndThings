package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrResetBusy = errors.New(
	"cannot reset cache while origin fetches are in progress",
)

// Tracks one origin fetch that other requests can share.
type inFlight struct {
	done  chan struct{}
	value string
	err   error
}

type Service struct {
	mu       sync.Mutex
	cache    *Cache
	origin   *Origin
	inFlight map[string]*inFlight
}

type FetchResult struct {
	Key       string `json:"key"`
	Content   string `json:"content"`
	Source    string `json:"source"`
	LatencyMS int64  `json:"latency_ms"`
}

// NewDemoService creates a CDN cache with a simulated origin.
func NewDemoService(capacity int) (*Service, error) {
	cache, err := New(capacity)
	if err != nil {
		return nil, err
	}

	return &Service{
		cache:    cache,
		origin:   NewDemoOrigin(),
		inFlight: make(map[string]*inFlight),
	}, nil
}

// Fetch retrieves content from the cache or origin.
// Overlapping misses for the same key share one origin fetch.
func (s *Service) Fetch(
	ctx context.Context,
	key string,
) (FetchResult, error) {

	if err := ctx.Err(); err != nil {
		return FetchResult{}, err
	}

	started := time.Now()

	s.mu.Lock()

	// First, check the local cache.
	value, found := s.cache.Get(key)

	if found {
		s.mu.Unlock()

		return FetchResult{
			Key:       key,
			Content:   value,
			Source:    "cache",
			LatencyMS: time.Since(started).Milliseconds(),
		}, nil
	}

	// Is someone already fetching this object?
	if pending, exists := s.inFlight[key]; exists {
		s.mu.Unlock()

		select {
		case <-ctx.Done():
			return FetchResult{}, ctx.Err()

		case <-pending.done:
			if err := ctx.Err(); err != nil {
				return FetchResult{}, err
			}

			if pending.err != nil {
				return FetchResult{}, pending.err
			}

			return FetchResult{
				Key:       key,
				Content:   pending.value,
				Source:    "shared",
				LatencyMS: time.Since(started).Milliseconds(),
			}, nil
		}
	}

	// We're the first requester: register this origin fetch.
	pending := &inFlight{
		done: make(chan struct{}),
	}

	s.inFlight[key] = pending
	s.mu.Unlock()

	// Network work happens outside the mutex.
	value, err := s.origin.Fetch(ctx, key)

	if err == nil {
		err = ctx.Err()
	}

	s.mu.Lock()

	if err == nil {
		s.cache.Put(key, value)
	}

	// Publish the result to all waiting requests.
	pending.value = value
	pending.err = err

	delete(s.inFlight, key)
	close(pending.done)

	s.mu.Unlock()

	if err != nil {
		return FetchResult{}, err
	}

	return FetchResult{
		Key:       key,
		Content:   value,
		Source:    "origin",
		LatencyMS: time.Since(started).Milliseconds(),
	}, nil
}

// Snapshot returns the current cache statistics.
func (s *Service) Snapshot() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.cache.Snapshot()
}

// Reset clears cached objects and resets statistics.
func (s *Service) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reject resets while origin fetches are running.
	if len(s.inFlight) > 0 {
		return ErrResetBusy
	}

	capacity := s.cache.Snapshot().Capacity

	freshCache, err := New(capacity)
	if err != nil {
		return err
	}

	s.cache = freshCache

	return nil
}

// KeysLRU returns the current cache ordering safely.
func (s *Service) KeysLRU() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.cache.KeysLRU()
}

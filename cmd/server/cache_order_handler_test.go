package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/patrick8bits/stuffAndThings/internal/cache"
)

func TestCacheOrderHandler(t *testing.T) {
	service, err := cache.NewDemoService(3)
	if err != nil {
		t.Fatal(err)
	}

	handler := newCacheOrderHandler(service)

	getOrder := func() []string {
		t.Helper()

		request := httptest.NewRequest(
			http.MethodGet,
			"/api/cache/order",
			nil,
		)

		recorder := httptest.NewRecorder()
		handler(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("Unexpected HTTP status: %d", recorder.Code)
		}

		var response struct {
			Order []string `json:"order"`
		}

		if err := json.NewDecoder(recorder.Body).
			Decode(&response); err != nil {
			t.Fatal(err)
		}

		return response.Order
	}

	// An empty cache should have no ordering.
	if len(getOrder()) != 0 {
		t.Fatal("Expected empty initial cache")
	}

	ctx := context.Background()

	// Populate cache: A is oldest, C is newest.
	for _, key := range []string{
		"/video/a", "/video/b", "/video/c",
	} {
		if _, err := service.Fetch(ctx, key); err != nil {
			t.Fatal(err)
		}
	}

	// Access A again, moving it to the newest position.
	if _, err := service.Fetch(ctx, "/video/a"); err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"/video/b",
		"/video/c",
		"/video/a",
	}

	if actual := getOrder(); !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}

	// Adding D should evict B.
	if _, err := service.Fetch(ctx, "/video/d"); err != nil {
		t.Fatal(err)
	}

	expected = []string{
		"/video/c",
		"/video/a",
		"/video/d",
	}

	if actual := getOrder(); !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}

	// Reset must clear the ordering.
	if err := service.Reset(); err != nil {
		t.Fatal(err)
	}

	if len(getOrder()) != 0 {
		t.Error("Expected empty order after reset")
	}

	// Only GET is supported.
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/cache/order",
		nil,
	)

	recorder := httptest.NewRecorder()
	handler(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Error("Expected HTTP 405 for POST")
	}
}

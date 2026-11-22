package cache

import "testing"

func TestLRUEviction(t *testing.T) {
	c, err := New(2)

	if err != nil {
		t.Fatal(err)
	}

	c.Put("A", "Object A")
	c.Put("B", "Object B")

	// Access A so B becomes least recently used.
	value, found := c.Get("A")

	if !found || value != "Object A" {
		t.Fatal("Expected cache hit for A")
	}

	// Inserting C should evict B, not A.
	c.Put("C", "Object C")

	if _, found := c.Get("B"); found {
		t.Error("Expected B to be evicted")
	}

	if _, found := c.Get("A"); !found {
		t.Error("Expected A to remain cached")
	}

	if _, found := c.Get("C"); !found {
		t.Error("Expected C to remain cached")
	}

	stats := c.Snapshot()

	if stats.Evictions != 1 {
		t.Errorf("Expected 1 eviction, got %d",
			stats.Evictions)
	}

	if stats.Hits != 3 || stats.Misses != 1 {
		t.Errorf("Unexpected statistics: %+v", stats)
	}

	if stats.Size != 2 {
		t.Errorf("Expected 2 cached objects, got %d",
			stats.Size)
	}

	t.Logf("Cache statistics: %+v", stats)
}

func TestInvalidCapacity(t *testing.T) {
	_, err := New(0)

	if err == nil {
		t.Error("Expected error for zero capacity")
	}
}

func TestUpdateExistingObject(t *testing.T) {
	c, err := New(2)

	if err != nil {
		t.Fatal(err)
	}

	c.Put("A", "Original")
	c.Put("A", "Updated")

	value, found := c.Get("A")

	if !found || value != "Updated" {
		t.Error("Expected updated cached value")
	}

	if c.Snapshot().Size != 1 {
		t.Error("Updating an object should not increase cache size")
	}
}

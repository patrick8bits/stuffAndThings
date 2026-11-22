package cache

import (
	"container/list"
	"errors"
	"sync"
)

type entry struct {
	key   string
	value string
}

type Stats struct {
	Capacity  int `json:"capacity"`
	Size      int `json:"size"`
	Hits      int `json:"hits"`
	Misses    int `json:"misses"`
	Evictions int `json:"evictions"`
}

type Cache struct {
	mu       sync.Mutex
	capacity int

	items map[string]*list.Element
	order *list.List

	hits      int
	misses    int
	evictions int
}

// New creates an empty cache with a fixed capacity.
func New(capacity int) (*Cache, error) {
	if capacity <= 0 {
		return nil, errors.New("cache capacity must be positive")
	}

	return &Cache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}, nil
}

// Get retrieves an object and marks it as recently used.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, found := c.items[key]

	if !found {
		c.misses++
		return "", false
	}

	c.hits++
	c.order.MoveToFront(element)

	item := element.Value.(*entry)

	return item.value, true
}

// Put inserts or updates an object in the cache.
func (c *Cache) Put(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update an existing object.
	if element, found := c.items[key]; found {
		item := element.Value.(*entry)

		item.value = value
		c.order.MoveToFront(element)
		return
	}

	// Insert a new object at the front.
	item := &entry{
		key:   key,
		value: value,
	}

	element := c.order.PushFront(item)
	c.items[key] = element

	// Evict the least recently used object if full.
	if c.order.Len() > c.capacity {
		oldest := c.order.Back()
		oldestItem := oldest.Value.(*entry)

		delete(c.items, oldestItem.key)
		c.order.Remove(oldest)

		c.evictions++
	}
}

// Snapshot returns a safe copy of cache statistics.
func (c *Cache) Snapshot() Stats {
	c.mu.Lock()
	defer c.mu.Unlock()

	return Stats{
		Capacity:  c.capacity,
		Size:      c.order.Len(),
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
	}
}

// KeysLRU returns cached keys from least to most recently used.
func (c *Cache) KeysLRU() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make([]string, 0, c.order.Len())

	// The back of the list is least recently used.
	for element := c.order.Back(); element != nil; element = element.Prev() {

		item := element.Value.(*entry)
		keys = append(keys, item.key)
	}

	return keys
}

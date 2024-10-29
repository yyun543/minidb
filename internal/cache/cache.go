package cache

import (
	"sync"
	"time"
)

type CacheEntry struct {
	Value      interface{}
	Expiration time.Time
}

type Cache struct {
	data map[string]CacheEntry
	mu   sync.RWMutex
	ttl  time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
		data: make(map[string]CacheEntry),
		ttl:  ttl,
	}
	go cache.cleanup()
	return cache
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = CacheEntry{
		Value:      value,
		Expiration: time.Now().Add(c.ttl),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.Expiration) {
		delete(c.data, key)
		return nil, false
	}

	return entry.Value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.After(entry.Expiration) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

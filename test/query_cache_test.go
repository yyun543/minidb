package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/cache"
	"github.com/yyun543/minidb/internal/executor"
)

func TestQueryCache_BasicOperations(t *testing.T) {
	config := &cache.CacheConfig{
		MaxSizeMB:        10, // 10MB for testing
		TTL:              1 * time.Minute,
		MinComputeTimeMs: 5,
	}

	qc := cache.NewQueryCache(config)
	require.NotNil(t, qc)

	// Create test result set
	result := &executor.ResultSet{
		Headers: []string{"id", "name"},
	}

	// Test cache miss
	key := cache.CacheKey{
		SQL:          "SELECT * FROM users",
		Database:     "testdb",
		TableVersion: 1,
	}

	cached, found := qc.Get(key)
	assert.False(t, found, "Should be cache miss initially")
	assert.Nil(t, cached)

	// Test cache put
	computeTime := 100 * time.Millisecond
	qc.Put(key, result, computeTime)

	// Give Ristretto time to process the write
	time.Sleep(10 * time.Millisecond)

	// Test cache hit
	cached, found = qc.Get(key)
	assert.True(t, found, "Should be cache hit after Put")
	assert.NotNil(t, cached)
	assert.Equal(t, result, cached)

	// Verify metrics
	metrics := qc.GetMetrics()
	assert.Greater(t, metrics.Hits, int64(0), "Should have at least 1 hit")
	assert.Greater(t, metrics.Misses, int64(0), "Should have at least 1 miss")
}

func TestQueryCache_TTLExpiration(t *testing.T) {
	config := &cache.CacheConfig{
		MaxSizeMB:        10,
		TTL:              100 * time.Millisecond, // Short TTL for testing
		MinComputeTimeMs: 5,
	}

	qc := cache.NewQueryCache(config)

	key := cache.CacheKey{
		SQL:          "SELECT COUNT(*) FROM orders",
		Database:     "testdb",
		TableVersion: 1,
	}

	result := &executor.ResultSet{
		Headers: []string{"count"},
	}

	// Put entry
	qc.Put(key, result, 50*time.Millisecond)
	time.Sleep(10 * time.Millisecond) // Wait for Ristretto to process

	// Should be available immediately
	cached, found := qc.Get(key)
	assert.True(t, found, "Should find entry before TTL expiration")
	assert.NotNil(t, cached)

	// Wait for TTL to expire
	time.Sleep(200 * time.Millisecond)

	// Should be expired
	cached, found = qc.Get(key)
	assert.False(t, found, "Should not find entry after TTL expiration")
	assert.Nil(t, cached)
}

func TestQueryCache_MinComputeTime(t *testing.T) {
	config := &cache.CacheConfig{
		MaxSizeMB:        10,
		TTL:              1 * time.Minute,
		MinComputeTimeMs: 50, // Only cache queries > 50ms
	}

	qc := cache.NewQueryCache(config)

	key := cache.CacheKey{
		SQL:          "SELECT * FROM fast_table",
		Database:     "testdb",
		TableVersion: 1,
	}

	result := &executor.ResultSet{
		Headers: []string{"id"},
	}

	// Try to cache fast query (should not be cached)
	qc.Put(key, result, 10*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	// Should not be in cache
	_, found := qc.Get(key)
	assert.False(t, found, "Fast query should not be cached")

	// Cache slow query (should be cached)
	qc.Put(key, result, 100*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	// Should be in cache
	_, found = qc.Get(key)
	assert.True(t, found, "Slow query should be cached")
}

func TestQueryCache_Clear(t *testing.T) {
	qc := cache.NewQueryCache(cache.DefaultCacheConfig())

	// Add some entries
	for i := 0; i < 5; i++ {
		key := cache.CacheKey{
			SQL:          "SELECT * FROM users WHERE id = " + string(rune(i)),
			Database:     "testdb",
			TableVersion: 1,
		}
		result := &executor.ResultSet{Headers: []string{"id"}}
		qc.Put(key, result, 50*time.Millisecond)
	}

	time.Sleep(20 * time.Millisecond) // Wait for Ristretto to process

	// Verify entries exist
	metrics := qc.GetMetrics()
	assert.Greater(t, metrics.Entries, int64(0), "Should have entries before clear")

	// Clear cache
	qc.Clear()

	// Verify cache is empty
	metrics = qc.GetMetrics()
	assert.Equal(t, int64(0), metrics.Entries, "Should have no entries after clear")
}

func TestQueryCache_CacheKeyString(t *testing.T) {
	key1 := cache.CacheKey{
		SQL:          "SELECT * FROM users",
		Database:     "db1",
		TableVersion: 100,
	}

	key2 := cache.CacheKey{
		SQL:          "SELECT * FROM users",
		Database:     "db1",
		TableVersion: 101, // Different version
	}

	// Different versions should produce different keys
	assert.NotEqual(t, key1.String(), key2.String())

	key3 := cache.CacheKey{
		SQL:          "SELECT * FROM users",
		Database:     "db1",
		TableVersion: 100,
	}

	// Same parameters should produce same key
	assert.Equal(t, key1.String(), key3.String())
}

func TestQueryCache_Metrics(t *testing.T) {
	qc := cache.NewQueryCache(cache.DefaultCacheConfig())

	key := cache.CacheKey{
		SQL:          "SELECT * FROM products",
		Database:     "testdb",
		TableVersion: 1,
	}

	result := &executor.ResultSet{
		Headers: []string{"id", "name", "price"},
	}

	// Initial cache miss
	_, found := qc.Get(key)
	assert.False(t, found)

	// Put and get
	qc.Put(key, result, 50*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	_, found = qc.Get(key)
	assert.True(t, found)

	// Check metrics
	metrics := qc.GetMetrics()
	assert.Greater(t, metrics.Hits, int64(0), "Should have hits")
	assert.Greater(t, metrics.Misses, int64(0), "Should have misses")
	assert.GreaterOrEqual(t, metrics.HitRate, 0.0, "Hit rate should be >= 0")
	assert.LessOrEqual(t, metrics.HitRate, 1.0, "Hit rate should be <= 1")
}

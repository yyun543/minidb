package cache

import (
	"sync"
	"time"
)

// CacheEntry 表示缓存中的一个条目
type Entry struct {
	Value      interface{} // 缓存的值
	Expiration time.Time   // 过期时间
}

// Cache 实现了一个简单的内存缓存
type Cache struct {
	data map[string]Entry // 存储缓存数据
	mu   sync.RWMutex     // 读写锁保护并发访问
	ttl  time.Duration    // 缓存条目的生存时间
}

// New 创建一个新的缓存实例
// ttl 参数指定缓存条目的默认生存时间
func New(ttl time.Duration) *Cache {
	cache := &Cache{
		data: make(map[string]Entry),
		ttl:  ttl,
	}
	// 启动后台清理过期条目的goroutine
	go cache.startCleanup()
	return cache
}

// Set 添加或更新缓存条目
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = Entry{
		Value:      value,
		Expiration: time.Now().Add(c.ttl),
	}
}

// Get 获取缓存条目
// 返回值: (缓存的值, 是否存在)
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(entry.Expiration) {
		// 如果过期，删除该条目
		go c.Delete(key) // 异步删除以避免在读锁时删除
		return nil, false
	}

	return entry.Value, true
}

// Delete 删除缓存条目
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Clear 清空所有缓存条目
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]Entry)
}

// startCleanup 启动定期清理过期条目的goroutine
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		c.cleanup()
	}
}

// cleanup 清理过期的缓存条目
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.Expiration) {
			delete(c.data, key)
		}
	}
}

// Size 返回当前缓存中的条目数量
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

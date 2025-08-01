package cache

import (
	"document-server/internal/config"
	storage "document-server/internal/storage/document"
	"sync"
	"time"
)

// cacheItem - это обертка для хранения значения и времени его истечения
type cacheItem struct {
	value     *storage.Document
	expiresAt time.Time
}

type InMemoryCache struct {
	store sync.Map
	ttl   time.Duration
}

func NewInMemoryCache(cfg config.CacheConfig) *InMemoryCache {
	return &InMemoryCache{
		ttl: time.Duration(cfg.TTL * int(time.Minute)),
	}
}

func (c *InMemoryCache) Get(key string) (*storage.Document, bool) {
	itemInterface, ok := c.store.Load(key)
	if !ok {
		return nil, false
	}

	item, ok := itemInterface.(cacheItem)
	if !ok {
		return nil, false
	}

	if time.Now().After(item.expiresAt) {
		c.store.Delete(key)
		return nil, false
	}

	return item.value, true
}

func (c *InMemoryCache) Set(key string, doc *storage.Document) {
	expiresAt := time.Now().Add(c.ttl)
	c.store.Store(key, cacheItem{
		value:     doc,
		expiresAt: expiresAt,
	})
}

func (c *InMemoryCache) Delete(key string) {
	c.store.Delete(key)
}

// Package cache
package cache

import (
	"hash/fnv"
	"orderkeeper/internal/models"
	"sync"
	"time"
)

const (
	shardCount              = 256
	defaultEvictionInterval = 5 * time.Minute
	defaultTTL              = 10 * time.Minute
)

type cacheItem struct {
	order     models.Order
	expiresAt time.Time
}

type cacheShard struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

type OrderCache struct {
	shards []*cacheShard
	stop   chan struct{}
}

func NewOrderCache() *OrderCache {
	c := &OrderCache{
		shards: make([]*cacheShard, shardCount),
		stop:   make(chan struct{}),
	}
	for i := 0; i < shardCount; i++ {
		c.shards[i] = &cacheShard{
			items: make(map[string]cacheItem),
		}
	}

	go c.runEvictionLoop(defaultEvictionInterval)

	return c
}

func (c *OrderCache) getShard(key string) *cacheShard {
	hasher := fnv.New32a()
	hasher.Write([]byte(key))
	return c.shards[hasher.Sum32()%shardCount]
}

func (c *OrderCache) Set(order models.Order) {
	shard := c.getShard(order.OrderUID)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	shard.items[order.OrderUID] = cacheItem{
		order:     order,
		expiresAt: time.Now().Add(defaultTTL),
	}
}

func (c *OrderCache) Get(uid string) (models.Order, bool) {
	shard := c.getShard(uid)
	shard.mu.RLock()
	defer shard.mu.RUnlock()

	item, exists := shard.items[uid]
	if !exists || time.Now().After(item.expiresAt) {
		return models.Order{}, false
	}

	return item.order, true
}

func (c *OrderCache) LoadFromDB(orders []models.Order) {
	for _, order := range orders {
		c.Set(order)
	}
}

func (c *OrderCache) Count() int {
	count := 0
	now := time.Now()
	for i := 0; i < shardCount; i++ {
		shard := c.shards[i]
		shard.mu.RLock()
		for _, item := range shard.items {
			if now.Before(item.expiresAt) {
				count++
			}
		}
		shard.mu.RUnlock()
	}
	return count
}

func (c *OrderCache) runEvictionLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.evictExpired()
		case <-c.stop:
			return
		}
	}
}

func (c *OrderCache) evictExpired() {
	now := time.Now()
	for i := 0; i < shardCount; i++ {
		shard := c.shards[i]
		shard.mu.Lock()
		for uid, item := range shard.items {
			if now.After(item.expiresAt) {
				delete(shard.items, uid)
			}
		}
		shard.mu.Unlock()
	}
}

func (c *OrderCache) StopEviction() {
	close(c.stop)
}

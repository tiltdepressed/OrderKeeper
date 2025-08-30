// Package cache реализует потокобезопасный, шардированный LRU-кеш.
package cache

import (
	"container/list"
	"hash/fnv"
	"orderkeeper/internal/models"
	"sync"
)

const (
	defaultMaxCacheSize = 1000
	shardCount          = 256
)

type cacheEntry struct {
	key   string
	order models.Order
}

type cacheShard struct {
	mu       sync.Mutex
	capacity int
	ll       *list.List
	items    map[string]*list.Element
}

type OrderCache struct {
	shards []*cacheShard
}

func NewOrderCache() *OrderCache {
	c := &OrderCache{
		shards: make([]*cacheShard, shardCount),
	}
	shardCapacity := defaultMaxCacheSize / shardCount
	if shardCapacity < 1 {
		shardCapacity = 1
	}

	for i := 0; i < shardCount; i++ {
		c.shards[i] = &cacheShard{
			capacity: shardCapacity,
			ll:       list.New(),
			items:    make(map[string]*list.Element),
		}
	}
	return c
}

func (c *OrderCache) getShard(key string) *cacheShard {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(key))
	return c.shards[hasher.Sum32()%shardCount]
}

func (c *OrderCache) Set(order models.Order) {
	shard := c.getShard(order.OrderUID)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if elem, ok := shard.items[order.OrderUID]; ok {
		shard.ll.MoveToFront(elem)
		elem.Value.(*cacheEntry).order = order
		return
	}

	if shard.ll.Len() >= shard.capacity {
		oldest := shard.ll.Back()
		if oldest != nil {
			removedEntry := shard.ll.Remove(oldest).(*cacheEntry)
			delete(shard.items, removedEntry.key)
		}
	}

	newEntry := &cacheEntry{key: order.OrderUID, order: order}
	elem := shard.ll.PushFront(newEntry)
	shard.items[order.OrderUID] = elem
}

func (c *OrderCache) Get(uid string) (models.Order, bool) {
	shard := c.getShard(uid)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if elem, ok := shard.items[uid]; ok {
		shard.ll.MoveToFront(elem)
		return elem.Value.(*cacheEntry).order, true
	}

	return models.Order{}, false
}

func (c *OrderCache) LoadFromDB(orders []models.Order) {
	for _, order := range orders {
		c.Set(order)
	}
}

func (c *OrderCache) Count() int {
	count := 0
	for i := 0; i < shardCount; i++ {
		shard := c.shards[i]
		shard.mu.Lock()
		count += shard.ll.Len()
		shard.mu.Unlock()
	}
	return count
}

// Package cache
package cache

import (
	"orderkeeper/internal/models"
	"sync"
)

type OrderCache struct {
	mu    sync.RWMutex
	items map[string]models.Order
}

func NewOrderCache() *OrderCache {
	return &OrderCache{
		items: make(map[string]models.Order),
	}
}

func (c *OrderCache) Set(order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[order.OrderUID] = order
}

func (c *OrderCache) Get(uid string) (models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, exists := c.items[uid]
	return order, exists
}

func (c *OrderCache) LoadFromDB(orders []models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, order := range orders {
		c.items[order.OrderUID] = order
	}
}

func (c *OrderCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

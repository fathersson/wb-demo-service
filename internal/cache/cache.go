package cache

import (
	"sync"

	"github.com/fathersson/wb-demo-service/internal/models"
)

type Cache struct {
	sync.RWMutex
	Orders map[string]models.Order
}

func NewCache() *Cache {
	return &Cache{
		Orders: make(map[string]models.Order),
	}
}

func (c *Cache) SetCache(orderUID string, order models.Order) {
	c.Lock()
	c.Orders[orderUID] = order
	c.Unlock()
}

func (c *Cache) GetCache(orderUID string) (models.Order, bool) {
	c.RLock()
	order, ok := c.Orders[orderUID]
	c.Unlock()

	return order, ok
}

package cache

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/fathersson/wb-demo-service/internal/models"
	"github.com/fathersson/wb-demo-service/internal/repository"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=CacheInterface --output=./cachemocks --with-expecter
type CacheInterface interface {
	SetCache(orderUID string, order models.Order)
	GetCache(orderUID string) (models.Order, bool)
}

type Cache struct {
	mu     sync.RWMutex
	Orders map[string]models.Order
	keys   []string // порядок добавления записей в кэш
	maxLen int      // лимит для кэша
}

func NewCache() *Cache {
	return &Cache{
		Orders: make(map[string]models.Order),
		maxLen: 1000,
	}
}

// Если orderUID нет в кэше, добавляем в keys
// Таким образом, когда наш кэш достигнет лимита
// Мы удалим самый старый элемент в keys и сдвинем слайс на 1 позицию
func (c *Cache) SetCache(orderUID string, order models.Order) {
	c.mu.Lock()
	if _, ok := c.Orders[orderUID]; !ok {
		c.keys = append(c.keys, orderUID)
	}
	if len(c.keys) > c.maxLen {
		oldKey := c.keys[0]
		delete(c.Orders, oldKey)
		c.keys = c.keys[1:]
	}
	c.Orders[orderUID] = order
	c.mu.Unlock()
}

func (c *Cache) GetCache(orderUID string) (models.Order, bool) {
	c.mu.RLock()
	order, ok := c.Orders[orderUID]
	c.mu.RUnlock()

	return order, ok
}

// Загружает кэш из базы при старте
func NewCacheFromDB(db repository.OrderRepository) (*Cache, error) {
	cache := NewCache()

	rows, err := db.Query("SELECT order_uid, track_number, entry, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders")
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к базе: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		// mu := &sync.Mutex{}
		var order models.Order

		// заполняем основные поля заказа
		if err := rows.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID,
			&order.DateCreated, &order.OofShard); err != nil {
			return nil, fmt.Errorf("ошибка сканирования заказа: %w", err)
		}

		// заполняем поля delivery
		err = db.QueryRow(`SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid = $1`, order.OrderUID).
			Scan(&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
				&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
				&order.Delivery.Email)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("ошибка сканирования delivery: %w", err)
		}

		// заполняем поля payment
		err = db.QueryRow(`SELECT transaction, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payment WHERE transaction = $1`, order.OrderUID).
			Scan(&order.Payment.Transaction, &order.Payment.Currency, &order.Payment.Provider,
				&order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank,
				&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("ошибка сканирования payment: %w", err)
		}

		// заполняем поля items
		itemRows, err := db.Query(`SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1`, order.OrderUID)
		if err != nil {
			return nil, fmt.Errorf("ошибка запроса к базе: %w", err)
		}

		for itemRows.Next() {
			var item models.Item

			if err := itemRows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
				&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand,
				&item.Status); err != nil {
				itemRows.Close()
				return nil, fmt.Errorf("ошибка сканирования items: %w", err)
			}
			order.Items = append(order.Items, item)
		}
		itemRows.Close()

		// сохраняем в кэш
		cache.SetCache(order.OrderUID, order)

		log.Printf("Заказ %s загружен в кэш", order.OrderUID)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к базе: %w", err)
	}

	log.Printf("Инициализация кэша завершена. Загружено %d заказов", len(cache.Orders))
	return cache, nil
}

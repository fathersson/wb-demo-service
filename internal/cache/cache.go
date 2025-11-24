package cache

import (
	"database/sql"
	"log"
	"sync"

	"github.com/fathersson/wb-demo-service/internal/models"
)

type Cache struct {
	mu     sync.RWMutex
	Orders map[string]models.Order
}

func NewCache() *Cache {
	return &Cache{
		Orders: make(map[string]models.Order),
	}
}

func (c *Cache) SetCache(orderUID string, order models.Order) {
	c.mu.Lock()
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
func NewCacheFromDB(db *sql.DB) *Cache {
	cache := NewCache()

	rows, err := db.Query("SELECT order_uid, track_number, entry, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		// mu := &sync.Mutex{}
		var order models.Order

		// заполняем основные поля заказа
		if err := rows.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID,
			&order.DateCreated, &order.OofShard); err != nil {
			log.Fatal(err)
			return nil
		}

		// заполняем поля delivery
		err = db.QueryRow(`SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid = $1`, order.OrderUID).
			Scan(&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
				&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
				&order.Delivery.Email)
		if err != nil && err != sql.ErrNoRows {
			log.Fatal(err)
			return nil
		}

		// заполняем поля payment
		err = db.QueryRow(`SELECT transaction, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payment WHERE transaction = $1`, order.OrderUID).
			Scan(&order.Payment.Transaction, &order.Payment.Currency, &order.Payment.Provider,
				&order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank,
				&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
		if err != nil && err != sql.ErrNoRows {
			log.Fatal(err)
			return nil
		}

		// заполняем поля items
		itemRows, err := db.Query(`SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1`, order.OrderUID)
		if err != nil {
			log.Fatal(err)
			return nil
		}

		for itemRows.Next() {
			var item models.Item

			if err := itemRows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
				&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand,
				&item.Status); err != nil {
				itemRows.Close()
				log.Fatal(err)
				return nil
			}
			order.Items = append(order.Items, item)
		}
		itemRows.Close()

		// сохраняем в кэш
		cache.mu.Lock()
		cache.Orders[order.OrderUID] = order
		cache.mu.Unlock()

		log.Printf("Заказ %s загружен в кэш", order.OrderUID)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
		return nil
	}

	log.Printf("Инициализация кэша завершена. Загружено %d заказов", len(cache.Orders))
	return cache
}

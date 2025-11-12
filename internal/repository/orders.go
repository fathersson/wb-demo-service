package repository

import (
	"context"
	"database/sql"
	"log"

	"github.com/fathersson/wb-demo-service/internal/models"
)

// SaveOrder сохраняет заказ и все связанные данные в базе в одной транзакции
func SaveOrder(ctx context.Context, db *sql.DB, order models.Order) error {
	// Начало транзакции
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Таблица orders
	var orderUID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO orders 
		(order_uid, track_number, entry, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING order_uid`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.CustomerID,
		order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard,
	).Scan(&orderUID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Таблица delivery
	_, err = tx.ExecContext(ctx,
		`INSERT INTO delivery
		(order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		orderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Таблица payment
	_, err = tx.ExecContext(ctx,
		`INSERT INTO payment
		(transaction, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		order.Payment.Transaction, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Таблица items
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO items
			(order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			orderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Order %s успешно сохранён", orderUID)
	return nil
}

// // Загружает кэш из базы при старте
// func NewCacheFromDB(db *sql.DB, cache *cache.Cache) *cache.Cache {

// 	rows, err := db.Query("SELECT order_uid, track_number, entry, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders")
// 	if err != nil {
// 		log.Fatal(err)
// 		return nil
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		mu := &sync.Mutex{}
// 		var order models.Order

// 		// заполняем основные поля заказа
// 		if err := rows.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
// 			&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID,
// 			&order.DateCreated, &order.OofShard); err != nil {
// 			log.Fatal(err)
// 			return nil
// 		}

// 		// заполняем поля delivery
// 		err = db.QueryRow(`SELECT name, phone, zip, city, address, region, email FROM delivery WHERE order_uid = $1`, order.OrderUID).
// 			Scan(&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
// 				&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
// 				&order.Delivery.Email)
// 		if err != nil && err != sql.ErrNoRows {
// 			log.Fatal(err)
// 			return nil
// 		}

// 		// заполняем поля payment
// 		err = db.QueryRow(`SELECT transaction, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payment WHERE order_uid = $1`, order.OrderUID).
// 			Scan(&order.Payment.Transaction, &order.Payment.Currency, &order.Payment.Provider,
// 				&order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank,
// 				&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
// 		if err != nil && err != sql.ErrNoRows {
// 			log.Fatal(err)
// 			return nil
// 		}

// 		// заполняем поля items
// 		itemRows, err := db.Query(`SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1`, order.OrderUID)
// 		if err != nil {
// 			log.Fatal(err)
// 			return nil
// 		}

// 		for itemRows.Next() {
// 			var item models.Item

// 			if err := itemRows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
// 				&item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand,
// 				&item.Status); err != nil {
// 				itemRows.Close()
// 				log.Fatal(err)
// 				return nil
// 			}
// 			order.Items = append(order.Items, item)
// 		}
// 		itemRows.Close()

// 		// сохраняем в кэш
// 		mu.Lock()
// 		cache.Orders[order.OrderUID] = order
// 		mu.Unlock()

// 		log.Printf("Заказ %s загружен в кэш", order.OrderUID)
// 	}

// 	err = rows.Err()
// 	if err != nil {
// 		log.Fatal(err)
// 		return nil
// 	}

// 	log.Printf("Инициализация кэша завершена. Загружено %d заказов", len(cache.Orders))
// 	return cache
// }

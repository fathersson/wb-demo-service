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

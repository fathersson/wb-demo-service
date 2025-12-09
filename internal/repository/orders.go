package repository

import (
	"context"
	"database/sql"

	"github.com/fathersson/wb-demo-service/internal/models"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=OrderRepository --output=./repomocks --with-expecter
type OrderRepository interface {
	SaveOrder(ctx context.Context, order models.Order) error
	GetOrderById(ctx context.Context, orderUID string) (models.Order, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) Query(query string, args ...any) (*sql.Rows, error) {
	return r.db.Query(query, args...)
}

func (r *PostgresRepo) QueryRow(query string, args ...any) *sql.Row {
	return r.db.QueryRow(query, args...)
}

// SaveOrder сохраняет заказ и все связанные данные в базе в одной транзакции
func (r *PostgresRepo) SaveOrder(ctx context.Context, order models.Order) error {
	// Начало транзакции
	tx, err := r.db.BeginTx(ctx, nil)
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
		(order_uid, transaction, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		orderUID, order.Payment.Transaction, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
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

	// log.Printf("Заказ %s успешно сохранён", orderUID)
	return nil
}

// Берем заказ по order_uid из бд
func (r *PostgresRepo) GetOrderById(ctx context.Context, orderUID string) (models.Order, error) {
	var order models.Order

	// Запрос в бд, данные таблицы orders
	err := r.db.QueryRowContext(ctx, "SELECT * FROM orders WHERE order_uid = $1", orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID,
		&order.DateCreated, &order.OofShard,
	)
	if err != nil {
		return models.Order{}, err
	}

	// Запрос в бд, данные таблицы delivery
	err = r.db.QueryRowContext(ctx, "SELECT * FROM delivery WHERE order_uid = $1", orderUID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
		&order.Delivery.Email,
	)
	if err != nil {
		return models.Order{}, err
	}

	// Запрос в бд, данные таблицы payment
	err = r.db.QueryRowContext(ctx, "SELECT * FROM payment WHERE transaction = $1", orderUID).Scan(
		&order.Payment.Transaction, &order.Payment.Currency, &order.Payment.Provider,
		&order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank,
		&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
	)
	if err != nil {
		return models.Order{}, err
	}

	// Запрос в бд, данные таблицы items
	rows, err := r.db.QueryContext(ctx, "SELECT * FROM items WHERE order_uid = $1", orderUID)
	if err != nil {
		return models.Order{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		if err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale,
			&item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		); err != nil {
			return models.Order{}, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

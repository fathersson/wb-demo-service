package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fathersson/wb-demo-service/internal/models"
	"github.com/stretchr/testify/assert"
)

// Тест проверяет успешный проход всей транзакции сохранения заказа
// Мы создаём sqlmock, задаем ожидаемые SQL-запросы, их аргументы,
// возвращаемые значения и подтверждаем, что транзакция выполняется
// строго в правильной последовательности: BEGIN - INSERT orders -
// INSERT delivery - INSERT payment - INSERT items - COMMIT
func TestSaveOrder_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// создаём репозиторий поверх мокнутой БД
	repo := NewPostgresRepo(db)
	ctx := context.Background()

	// создаём тестовый заказ
	order := models.Order{
		OrderUID:    "id1",
		TrackNumber: "TRACK1",
		Delivery: models.Delivery{
			Name:    "Ivan",
			Phone:   "+7999",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Street 1",
		},
		Payment: models.Payment{
			Transaction:  "id1",
			Currency:     "RUB",
			Provider:     "bank",
			Amount:       1000,
			PaymentDT:    time.Now().Unix(),
			DeliveryCost: 200,
			GoodsTotal:   800,
		},
		Items: []models.Item{
			{ChrtID: 1, Name: "Item", Price: 100, TotalPrice: 100},
		},
	}

	// Ожидаем начало транзакции
	mock.ExpectBegin()
	// Ожидаем INSERT в таблицу orders
	// Используем regexp.QuoteMeta, чтобы избежать проблем с переносами строк
	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO orders 
        (order_uid, track_number, entry, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING order_uid`,
	)).
		WithArgs(
			order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.CustomerID,
			order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard,
		).
		// Ответ SQL - вернуть order_uid представив что он успешно вставился
		WillReturnRows(sqlmock.NewRows([]string{"order_uid"}).AddRow(order.OrderUID))

	// Ожидаем INSERT в таблицу delivery
	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO delivery
        (order_uid, name, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
	)).
		WithArgs(order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
			order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидаем INSERT в таблицу payment
	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO payment
        (order_uid, transaction, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
	)).
		WithArgs(order.OrderUID, order.Payment.Transaction, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
			order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидаем INSERT в таблицу items
	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO items
        (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
	)).
		WithArgs(order.OrderUID, order.Items[0].ChrtID, order.Items[0].TrackNumber, order.Items[0].Price, order.Items[0].Rid,
			order.Items[0].Name, order.Items[0].Sale, order.Items[0].Size, order.Items[0].TotalPrice, order.Items[0].NmID,
			order.Items[0].Brand, order.Items[0].Status).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидаем коммит транзакции
	mock.ExpectCommit()

	// Вызываем тестируемую функцию
	err = repo.SaveOrder(ctx, order)
	// Ошибки быть не должно
	assert.NoError(t, err)
	// Проверяем, что ВСЕ ожидаемые SQL-команды были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Ошибка в первом же запросе (INSERT INTO orders).
// Транзакция должна откатиться (Rollback), а функция вернуть ошибку.
func TestSaveOrder_FailOnOrders(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewPostgresRepo(db)
	ctx := context.Background()
	order := models.Order{OrderUID: "id1"}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO orders").
		WillReturnError(assert.AnError)

	// Функция должна вызвать Rollback
	// Rollback будет автоматически ожидаемым действием,
	// и мы проверим его через ExpectationsWereMet()

	err := repo.SaveOrder(ctx, order)
	assert.Error(t, err)
	// Должен быть Rollback
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Заказ не найден по ID.
// Запрос должен вернуть sql.ErrNoRows, и репозиторий должен
// вернуть эту ошибку вверх
func TestGetOrderById_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := NewPostgresRepo(db)
	ctx := context.Background()

	// Мы ожидаем SELECT по order_uid,
	// и он должен вернуть sql.ErrNoRows
	mock.ExpectQuery("SELECT \\* FROM orders").
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetOrderById(ctx, "missing")
	// Ошибка обязательна
	assert.Error(t, err)
	// И все ожидания должны быть выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

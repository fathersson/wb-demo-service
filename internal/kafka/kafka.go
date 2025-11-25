// Незаконченный задачи, нужно выполнить после того как подниму сервер

// 1. Парсинг сообщений
// Используй json.Unmarshal([]byte, &order)
// Оберни в обработку ошибок:
// если JSON некорректный — логируй и пропускай,
// если order_uid пустой — игнорируй сообщение.

// 2. Проверка и логирование
// Для теста:
// выведи order_uid в консоль при каждом сообщении,
// убедись, что consumer не падает на некорректных данных.

package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/fathersson/wb-demo-service/internal/cache"
	"github.com/fathersson/wb-demo-service/internal/config"
	"github.com/fathersson/wb-demo-service/internal/models"
	"github.com/fathersson/wb-demo-service/internal/repository"
	"github.com/segmentio/kafka-go"
)

// NewReader создает и возвращает настроенный kafka.Reader
func NewReader(cfg config.KafkaConfig) *kafka.Reader {
	defer log.Println("Kafka reader запущен")
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cfg.Broker},
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		StartOffset:    kafka.FirstOffset, // читаем с начала при первом запуске
		CommitInterval: 0,
	})
}

// ConsumeMessages читает сообщения из Kafka
func ConsumeMessages(reader *kafka.Reader, db *sql.DB, cache *cache.Cache) {
	log.Println("Kafka consumer запущен")
	var order models.Order
	ctx := context.Background()

	// Читаем сообщения
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			log.Println("Ошибка чтения заказа:", err)
			continue
		}

		// Парсим JSON
		err = json.Unmarshal(msg.Value, &order)
		if err != nil {
			log.Println("Ошибка парсинга JSON:", err)
			// Пропускаем некорректное сообщение, не коммитим
			continue
		}

		// Проверка корректности order_uid
		if order.OrderUID == "" {
			log.Println("Пустой order_uid, пропускаем")
			continue
		}
		// Сообщение корректное
		log.Printf("Получили заказ %s из %s", order.OrderUID, order.Delivery.City)

		// проводим транзакцию в бд
		err = repository.SaveOrder(ctx, db, order)
		if err != nil {
			log.Println("Ошибка сохранения заказа:", err)
			continue
		}
		log.Printf("Заказ %s сохранен в базе данных", order.OrderUID)

		// Добавляем сообщение в кэш
		cache.SetCache(order.OrderUID, order)
		log.Printf("Заказ %s добавлен в кэш", order.OrderUID)

		// Посылаем сигнал в Kafka, что мы обработали его сообщение
		err = reader.CommitMessages(ctx, msg)
		if err != nil {
			panic(err)
		}
		// Принтуем в консоль
		log.Printf("Консьюмер кафки обработал заказ %s", order.OrderUID)

		// Тело заказа
		log.Printf("Тело заказа: %+v", order)
	}
}

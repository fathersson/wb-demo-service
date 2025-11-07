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
	"encoding/json"
	"log"

	"github.com/fathersson/wb-demo-service/internal/config"
	"github.com/fathersson/wb-demo-service/internal/models"
	"github.com/segmentio/kafka-go"
)

// NewReader создает и возвращает настроенный kafka.Reader
func NewReader(cfg config.KafkaConfig) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cfg.Broker},
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		StartOffset:    kafka.FirstOffset, // читаем с начала при первом запуске
		CommitInterval: 0,
	})
}

// ConsumeMessages читает сообщения из Kafka
func ConsumeMessages(reader *kafka.Reader) {
	ctx := context.Background()

	// Читаем сообщения
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			panic(err)
		}

		// Парсим JSON
		var order models.Order
		err = json.Unmarshal(msg.Value, &order)
		if err != nil {
			log.Println("Ошибка парсинга JSON:", err)
			continue // пропускаем это сообщение, чтобы не коммитить некорректные данные
		}

		// Проверка корректности order_uid
		if order.OrderUID == "" {
			log.Println("Пустой order_uid, пропускаем")
			continue
		}
		// Сообщение корректное
		log.Printf("Получили заказ %s из %s", order.OrderUID, order.Delivery.City)

		// Печатаем сообщение
		// fmt.Printf("message at offset %d: %s = %s\n", msg.Offset, string(msg.Key), string(msg.Value))

		// Посылаем сигнал в Kafka, что мы обработали его сообщение
		err = reader.CommitMessages(ctx, msg)
		if err != nil {
			panic(err)
		}
	}
}

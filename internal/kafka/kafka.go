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
	"log"

	"github.com/fathersson/wb-demo-service/internal/config"
	"github.com/fathersson/wb-demo-service/internal/models"
	"github.com/fathersson/wb-demo-service/internal/repository"
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
func ConsumeMessages(reader *kafka.Reader, db *sql.DB) {
	var order models.Order
	ctx := context.Background()

	// Читаем сообщения
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			log.Println("Ошибка чтения заказа:", err)
			continue
		}

		// Проверка корректности order_uid
		if order.OrderUID == "" {
			log.Println("Пустой order_uid, пропускаем")
			continue
		}
		// Сообщение корректное
		log.Printf("Получили заказ %s из %s", order.OrderUID, order.Delivery.City)

		// // Начинаем транзакцию бд
		// tx, err := db.BeginTx(ctx, nil)
		// if err != nil {
		// 	log.Println("Не удалось начать транзакцию:", err)
		// 	continue
		// }

		// проводим транзакцию в бд
		err = repository.SaveOrder(ctx, db, order)
		if err != nil {
			log.Println("Ошибка сохранения заказа:", err)
			continue
		}

		// Посылаем сигнал в Kafka, что мы обработали его сообщение
		err = reader.CommitMessages(ctx, msg)
		if err != nil {
			panic(err)
		}
	}
}

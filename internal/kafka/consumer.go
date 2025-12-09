package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/fathersson/wb-demo-service/internal/cache"
	"github.com/fathersson/wb-demo-service/internal/config"
	"github.com/fathersson/wb-demo-service/internal/models"
	"github.com/fathersson/wb-demo-service/internal/repository"
	"github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=MessageReader --output=./kafkamocks --with-expecter
type MessageReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

// NewReader создает и возвращает настроенный kafka.Reader
func NewReader(cfg config.KafkaConfig) *kafka.Reader {
	log.Println("Создаем Kafka reader")
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cfg.Broker},
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		StartOffset:    kafka.FirstOffset, // читаем с начала при первом запуске
		CommitInterval: 0,
	})
}

// ConsumeMessages читает сообщения из Kafka
func ConsumeMessages(reader MessageReader, db repository.OrderRepository, cache cache.CacheInterface, ctx context.Context) {
	log.Println("Kafka consumer запущен")
	var order models.Order
	var validate = validator.New()

	// Читаем сообщения
	for {
		select {
		case <-ctx.Done():
			log.Println("Kafka consumer завершен")
			return
		default:
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

			err = validate.Struct(order)
			if err != nil {
				log.Printf("Сообщение некорректно, ошибка:%s", err)
				continue
			}
			if len(order.Items) == 0 {
				log.Println("Сообщение некорректно, нет товаров в заказе")
				continue
			}

			// Сообщение корректное
			log.Printf("Получили заказ %s из %s", order.OrderUID, order.Delivery.City)

			// проводим транзакцию в бд
			err = db.SaveOrder(ctx, order)
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
				log.Println("Ошибка коммита сообщения:", err)
				continue
			}
			// Принтуем в консоль
			log.Printf("Консьюмер кафки обработал заказ %s", order.OrderUID)

			// Тело заказа
			log.Printf("Тело заказа: %+v", order)
		}
	}
}

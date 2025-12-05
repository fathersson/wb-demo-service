package kafka

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/segmentio/kafka-go"

	"github.com/fathersson/wb-demo-service/internal/config"
	"github.com/fathersson/wb-demo-service/internal/models"
)

// NewWriter — создаёт Kafka producer
func NewWriter(cfg config.KafkaConfig) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(cfg.Broker),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
	}
}

// Generator - каждые 10 секунд отправляет заказ
func Generator(writer *kafka.Writer, ctx context.Context) {
	log.Println("Kafka producer запущен")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	sendValid := true // флаг: true - валидное, false - битое

	for {
		select {
		case <-ctx.Done():
			log.Println("Writer остановлен")
			return
		case <-ticker.C:
			var data []byte
			var err error

			if sendValid {
				// генерируем валидный заказ
				order := generateOrder()
				data, err = json.Marshal(order)
				if err != nil {
					log.Println("Ошибка сериализации:", err)
					continue
				}
				log.Printf("Отправлен ВАЛИДНЫЙ заказ %s", order.OrderUID)
			} else {
				// генерируем битое сообщение
				data = BrokenOrder()
				log.Printf("Отправлено БИТОЕ сообщение: %s", string(data))
			}

			err = writer.WriteMessages(ctx, kafka.Message{Value: data})
			if err != nil {
				log.Println("Ошибка отправки:", err)
			}

			// меняем флаг: следующее сообщение будет другого типа
			sendValid = !sendValid
		}
	}

}

// generateOrder создаёт случайный Order с помощью faker
func generateOrder() models.Order {
	now := time.Now()
	UUID := faker.UUIDDigit()

	return models.Order{
		OrderUID:    UUID,
		TrackNumber: faker.UUIDHyphenated(),
		Entry:       "WBIL",

		Delivery: models.Delivery{
			Name:    faker.Name(),
			Phone:   faker.Phonenumber(),
			Zip:     "123456",
			City:    "Moscow",
			Address: "Павла Хохотушкина 25",
			Region:  "Region",
			Email:   faker.Email(),
		},

		Payment: models.Payment{
			Transaction:  UUID,
			Currency:     "RUB",
			Provider:     "bank",
			Amount:       rand.Intn(10000),
			PaymentDT:    now.Unix(),
			Bank:         "Sber",
			DeliveryCost: rand.Intn(500),
			GoodsTotal:   rand.Intn(10000),
			CustomFee:    0,
		},

		Items: []models.Item{
			{
				ChrtID:      rand.Intn(999999),
				TrackNumber: faker.UUIDHyphenated(),
				Price:       rand.Intn(5000),
				Rid:         faker.UUIDDigit(),
				Name:        faker.Word(),
				TotalPrice:  rand.Intn(5000),
				NmID:        rand.Intn(500000),
				Brand:       "Brand",
				Status:      202,
			},
		},

		Locale:      "ru",
		CustomerID:  faker.Username(),
		ShardKey:    "2",
		SmID:        rand.Intn(100),
		DateCreated: now,
		OofShard:    "1",
	}
}

func BrokenOrder() []byte {
	switch rand.Intn(10) {

	case 0:
		// Сломанный JSON
		return []byte("{bad json")

	case 1:
		// Неверные типы верхнего уровня
		return []byte(`{"order_uid": 123, "track_number": 777}`)

	case 2:
		// Отсутствуют обязательные поля (delivery, payment, items)
		return []byte(`{"order_uid":"AAA111","track_number":"BBB222"}`)

	case 3:
		// Не валидная структура delivery
		return []byte(`{"order_uid":"A1","delivery":{"phone":123}}`)

	case 4:
		// Пустой items — невалидно по твоему валидатору (min=1)
		return []byte(`{"order_uid":"A1","items":[]}`)

	case 5:
		// Неверные типы внутри items
		return []byte(`{
			"order_uid": "XYZ",
			"items": [{"chrt_id": "not_number"}]
		}`)

	case 6:
		// Отсутствует поле payment в целом
		return []byte(`{"order_uid":"A1","delivery":{"name":"test"}}`)

	case 7:
		// payment со сломанным типом
		return []byte(`{
			"order_uid": "A1",
			"payment": {"amount": "not_number"}
		}`)

	case 8:
		// Некорректный email во вложенных данных
		return []byte(`{
			"order_uid": "A1",
			"delivery": {"name":"test","phone":"123","zip":"123456","city":"MSK","address":"Street","email":"not_email"}
		}`)

	case 9:
		// Вообще не JSON
		return []byte("hello world")

	default:
		return []byte("???")
	}
}

package kafka

import (
	"context"
	"testing"
	"time"

	cachemocks "github.com/fathersson/wb-demo-service/internal/cache/cachemocks"
	kafkamocks "github.com/fathersson/wb-demo-service/internal/kafka/kafkamocks"
	repomocks "github.com/fathersson/wb-demo-service/internal/repository/repomocks"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

func TestConsumeMessages_Success(t *testing.T) {
	reader := kafkamocks.NewMessageReader(t)
	repo := repomocks.NewOrderRepository(t)
	cache := cachemocks.NewCacheInterface(t)

	validJSON := `{
		"order_uid": "test123",
		"track_number": "TRACK001",
		"delivery": {"name":"Ivan","phone":"+79990000000","zip":"123456","city":"Moscow","address":"Street 1"},
		"payment": {"transaction":"test123","currency":"RUB","provider":"bank","amount":1000,"payment_dt":1234567890,"delivery_cost":200,"goods_total":800},
		"items": [{"chrt_id":1,"name":"Item","price":100,"total_price":100}]
	}`
	msg := kafka.Message{Value: []byte(validJSON)}

	// Основной контекст + отмена после первой успешной итерации
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reader.EXPECT().
		FetchMessage(mock.Anything).
		Return(msg, nil).
		Once().
		Run(func(args mock.Arguments) {
			// Как только успешно считали и дальше всё обработаем — гасим контекст,
			// чтобы цикл вышел по ctx.Done()
			cancel()
		})

	repo.EXPECT().
		SaveOrder(mock.Anything, mock.AnythingOfType("models.Order")).
		Return(nil).
		Once()

	cache.EXPECT().
		SetCache("test123", mock.AnythingOfType("models.Order")).
		Return().
		Once()

	reader.EXPECT().
		CommitMessages(mock.Anything, msg).
		Return(nil).
		Once()

	// Завершающий FetchMessage нам не обязателен: контекст уже отменён,
	// но если консьюмер успеет ещё раз дернуть FetchMessage, разрешаем:
	reader.EXPECT().
		FetchMessage(mock.Anything).
		Return(kafka.Message{}, context.Canceled).
		Maybe()

	done := make(chan struct{})
	go func() {
		ConsumeMessages(reader, repo, cache, ctx)
		close(done)
	}()
	<-done

	reader.AssertExpectations(t)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

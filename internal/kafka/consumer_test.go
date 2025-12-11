package kafka

import (
	"context"
	"errors"
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

func TestConsumeMessages_Bad(t *testing.T) {
	reader := kafkamocks.NewMessageReader(t)
	repo := repomocks.NewOrderRepository(t)
	cache := cachemocks.NewCacheInterface(t)

	badJSON := `{}`
	msg := kafka.Message{Value: []byte(badJSON)}

	// Основной контекст + отмена после первой успешной итерации
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// Первый fetch - невалидное сообщение
	reader.EXPECT().
		FetchMessage(mock.Anything).
		Return(msg, nil).
		Once().
		Run(func(args mock.Arguments) {
			// Как только успешно считали и дальше всё обработаем — гасим контекст,
			// чтобы цикл вышел по ctx.Done()
			cancel()
		})

	// Для выхода из цикла разрешаем/ожидаем ещё один Fetch с context.Canceled
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

	// Эти вызовы НЕ должны быть
	repo.AssertNotCalled(t, "SaveOrder", mock.Anything, mock.Anything)
	cache.AssertNotCalled(t, "SetCache", mock.Anything, mock.Anything)
	reader.AssertNotCalled(t, "CommitMessages", mock.Anything, mock.Anything)

	// Проверяем, что ожидания по FetchMessage выполнены
	reader.AssertExpectations(t)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestConsumeMessages_BadFetch(t *testing.T) {
	reader := kafkamocks.NewMessageReader(t)
	repo := repomocks.NewOrderRepository(t)
	cache := cachemocks.NewCacheInterface(t)

	// Основной контекст + отмена после первой успешной итерации
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// Первый вызов FetchMessage вернёт ошибку
	reader.EXPECT().
		FetchMessage(mock.Anything).
		Return(kafka.Message{}, errors.New("fetch error")).
		Once().
		Run(func(args mock.Arguments) {
			// Как только успешно считали и дальше всё обработаем — гасим контекст,
			// чтобы цикл вышел по ctx.Done()
			cancel()
		})

	// Для выхода из цикла разрешаем ещё один вызов, который вернёт context.Canceled
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

	// Проверяем, что побочные методы не вызывались
	repo.AssertNotCalled(t, "SaveOrder", mock.Anything, mock.Anything)
	cache.AssertNotCalled(t, "SetCache", mock.Anything, mock.Anything)
	reader.AssertNotCalled(t, "CommitMessages", mock.Anything, mock.Anything)

	reader.AssertExpectations(t)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestConsumeMessages_BadItmes(t *testing.T) {
	reader := kafkamocks.NewMessageReader(t)
	repo := repomocks.NewOrderRepository(t)
	cache := cachemocks.NewCacheInterface(t)

	validJSON := `{
		"order_uid": "test123",
		"track_number": "TRACK001",
		"delivery": {"name":"Ivan","phone":"+79990000000","zip":"123456","city":"Moscow","address":"Street 1"},
		"payment": {"transaction":"test123","currency":"RUB","provider":"bank","amount":1000,"payment_dt":1234567890,"delivery_cost":200,"goods_total":800},
		"items": []
	}`
	msg := kafka.Message{Value: []byte(validJSON)}

	// Основной контекст + отмена после первой успешной итерации
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reader.EXPECT().
		FetchMessage(mock.Anything).
		Return(msg, nil).
		Once().
		Run(func(args mock.Arguments) { cancel() }) // чтобы выйти по ctx.Done()

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

	// Проверяем, что побочные методы не вызывались
	repo.AssertNotCalled(t, "SaveOrder", mock.Anything, mock.Anything)
	cache.AssertNotCalled(t, "SetCache", mock.Anything, mock.Anything)
	reader.AssertNotCalled(t, "CommitMessages", mock.Anything, mock.Anything)

	reader.AssertExpectations(t)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestConsumeMessages_BadSaveOrder(t *testing.T) {
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
		Run(func(args mock.Arguments) { cancel() })

	repo.EXPECT().
		SaveOrder(mock.Anything, mock.AnythingOfType("models.Order")).
		Return(errors.New("error SaveOrder")).
		Once()

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

	cache.AssertNotCalled(t, "SetCache", mock.Anything, mock.Anything)
	reader.AssertNotCalled(t, "CommitMessages", mock.Anything, mock.Anything)

	reader.AssertExpectations(t)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestConsumeMessages_BadCommit(t *testing.T) {
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
		Run(func(args mock.Arguments) { cancel() })

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
		Return(errors.New("error CommitMessages")).
		Once()

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

func TestConsumeMessages_BadContext(t *testing.T) {
	reader := kafkamocks.NewMessageReader(t)
	repo := repomocks.NewOrderRepository(t)
	cache := cachemocks.NewCacheInterface(t)

	// Основной контекст + отмена после первой успешной итерации
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	cancel()

	// Разрешаем (но не требуем) FetchMessage -> context.Canceled
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

	reader.AssertNotCalled(t, "FetchMessage", mock.Anything)
	repo.AssertNotCalled(t, "SaveOrder", mock.Anything, mock.Anything)
	cache.AssertNotCalled(t, "SetCache", mock.Anything, mock.Anything)
	reader.AssertNotCalled(t, "CommitMessages", mock.Anything, mock.Anything)

	reader.AssertExpectations(t)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

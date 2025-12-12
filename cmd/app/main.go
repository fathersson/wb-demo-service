package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fathersson/wb-demo-service/internal/cache"
	"github.com/fathersson/wb-demo-service/internal/config"
	"github.com/fathersson/wb-demo-service/internal/db"
	"github.com/fathersson/wb-demo-service/internal/kafka"
	"github.com/fathersson/wb-demo-service/internal/repository"
	"github.com/fathersson/wb-demo-service/internal/server"
)

func main() {
	// Контекст и wait group для graceful shutdown всех фоновых горутин
	var wg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Загрузка конфигурации (.env/окружение), без неё приложение не стартует
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	// Подключение к PostgreSQL: открываем соединение, проверяем Ping, закрываем при выходе
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		log.Fatal("Не удалось подключиться к базе:", err)
	}
	defer database.Close()
	log.Println("Соединение с базой данных установлено")

	// Репозиторий поверх *sql.DB
	postgres := repository.NewPostgresRepo(database)

	// Инициализация (загрузка) in-memory кэша из БД при старте
	orderCache, err := cache.NewCacheFromDB(postgres)
	if err != nil {
		log.Fatal("Ошибка загрузки кэша:", err)
	}

	// Kafka producer: в отдельной горутине генерирует валидные/битые сообщения до остановки контекста
	writer := kafka.NewWriter(cfg.Kafka)
	wg.Add(1)
	go func() {
		defer wg.Done()
		kafka.Generator(writer, ctx)
	}()

	// Kafka consumer: читает, валидирует, сохраняет в БД и кэш, работает пока не остановится контекст
	reader := kafka.NewReader(cfg.Kafka)
	defer reader.Close()

	wg.Add(1)
	go func() {
		defer wg.Done()
		kafka.ConsumeMessages(reader, postgres, orderCache, ctx)
	}()

	// HTTP сервер, хендлеры используют кэш и репозиторий
	srv := server.NewServer(cfg.HttpServer, orderCache, postgres)

	// Запуск HTTP сервера в горутине, фатал при ошибке кроме штатного закрытия
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Ошибка запуска сервера:", err)
		}
	}()

	// Ожидание сигнала, затем мягкая остановка HTTP и завершение горутин с таймаутом
	<-ctx.Done()
	ctxShutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Println("Ошибка завершения сервера с shutdown")
	}

	wg.Wait()

}

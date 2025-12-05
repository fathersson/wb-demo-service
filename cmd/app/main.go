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
	"github.com/fathersson/wb-demo-service/internal/server"
)

func main() {
	// 0. Контекст для корректного завершения
	var wg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 1. Загружаем .env и конфиг
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	// 2. Соединение с бд
	database, err := db.Connect(&cfg.Database)
	if err != nil {
		log.Fatal("Не удалось подключиться к базе:", err)
	}
	defer database.Close()
	log.Println("Соединение с базой данных установлено")

	// Подгружаем кэш из бд
	orderCache, err := cache.NewCacheFromDB(database)
	if err != nil {
		log.Fatal("Ошибка загрузки кэша:", err)
	}

	writer := kafka.NewWriter(cfg.Kafka)
	wg.Add(1)
	go func() {
		defer wg.Done()
		kafka.Generator(writer, ctx)
	}()

	// 3. Читаем сообщения не блокируя основной поток
	reader := kafka.NewReader(cfg.Kafka)
	defer reader.Close()

	wg.Add(1)
	go func() {
		defer wg.Done()
		kafka.ConsumeMessages(reader, database, orderCache, ctx)
	}()

	// 5. Создаем обьект http.Server
	srv := server.NewServer(cfg.HttpServer, orderCache, database)

	// 6. Запускаем сервер
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Ошибка запуска сервера:", err)
		}
	}()

	<-ctx.Done()
	ctxShutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Println("Ошибка завершения сервера с shutdown")
	}

	wg.Wait()

}

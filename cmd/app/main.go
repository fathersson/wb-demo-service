package main

import (
	"log"

	"github.com/fathersson/wb-demo-service/internal/cache"
	"github.com/fathersson/wb-demo-service/internal/config"
	"github.com/fathersson/wb-demo-service/internal/db"
	"github.com/fathersson/wb-demo-service/internal/kafka"
	"github.com/fathersson/wb-demo-service/internal/server"
)

func main() {
	// 1. Загружаем конфиг
	cfg := config.Load()
	// Создаем кэш
	// cache := cache.NewCache()

	// 2. Соединение с бд
	db := db.Connect(&cfg.Database)
	defer db.Close()

	// Подгружаем кэш из бд
	cache := cache.NewCacheFromDB(db)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// 3. Подключение к Kafka
	reader := kafka.NewReader(cfg.Kafka)
	defer reader.Close()

	// 4. Читаем сообщения не блокируя основной поток
	go kafka.ConsumeMessages(reader, db, cache)

	// 5. Создаем обьект http.Server
	srv := server.NewServer(cfg.HttpServer)
	log.Println("Сервер будет запущен на", cfg.HttpServer.Port)

	// 6. Запускаем сервер
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}

}

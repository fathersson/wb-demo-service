package app

import (
	"github.com/fathersson/wb-demo-service/internal/config"
	"github.com/fathersson/wb-demo-service/internal/db"
)

func main() {
	// 1. Загружаем конфиг
	cfg := config.Load()

	// 2. Соединение с бд
	db := db.Connect(&cfg.Database)
	db.Close()

	// 3. Подключение к Kafka

}

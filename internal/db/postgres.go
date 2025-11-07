package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/fathersson/wb-demo-service/internal/config"
	_ "github.com/lib/pq"
)

func Connect(cfg *config.DatabaseConfig) *sql.DB {
	// Форматируем строку подключения, подставляя реальные значения из cfg
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBname)

	// Открываем соединение
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// Проверяем соединение
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Соединение с базой данных установлено")
	return db
}

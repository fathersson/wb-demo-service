package db

import (
	"database/sql"
	"fmt"

	"github.com/fathersson/wb-demo-service/internal/config"
	_ "github.com/lib/pq"
)

func Connect(cfg *config.DatabaseConfig) (*sql.DB, error) {
	// Форматируем строку подключения, подставляя реальные значения из cfg
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	// Открываем соединение
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия базы: %w", err)
	}

	// Проверяем соединение
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе: %w", err)
	}

	return db, nil
}

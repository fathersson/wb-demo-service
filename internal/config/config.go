package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Config - агрегирует все настройки приложения: HTTP, БД и Kafka
type Config struct {
	HttpServer HttpServer     `yaml:"http_server"`
	Database   DatabaseConfig `yaml:"database"`
	Kafka      KafkaConfig    `yaml:"kafka"`
}

// HttpServer - конфигурация HTTP-сервера (порт берётся из env/конфига)
type HttpServer struct {
	Port int `yaml:"port" env:"HTTP_PORT"`
}

// DatabaseConfig - настройки подключения к PostgreSQL
// Поля читаются из переменных окружения (env)
type DatabaseConfig struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST"`
	Port     string `yaml:"port" env:"POSTGRES_PORT"`
	User     string `yaml:"user" env:"POSTGRES_USER"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
	DBName   string `yaml:"dbname" env:"POSTGRES_DB"`
}

// KafkaConfig - настройки брокера Kafka (адрес, топик, group ID)
// Значения приходят из env/конфига через cleanenv
type KafkaConfig struct {
	Broker    string `yaml:"broker" env:"KAFKA_BROKER"`
	Zookeeper string `yaml:"zookeeper" env:"KAFKA_ZOOKEEPER"`
	Topic     string `yaml:"topic" env:"KAFKA_TOPIC"`
	GroupID   string `yaml:"groupID" env:"KAFKA_GROUP_ID"`
	// Commit    bool   `yaml:"commit"`
}

// Load - грузит .env и переменные окружения в структуру Config
// При ошибке возвращает error, вызывающий должен обработать/остановить приложение
func Load() (*Config, error) {
	// Читаем переменные окружения из env файла
	if err := godotenv.Load(".env"); err != nil {
		return nil, fmt.Errorf("ошбика загрузки .env файла: %w", err)
	}

	// Содаем обьект конфига
	var cfg Config

	// Читаем конфиг
	// if err := cleanenv.ReadConfig("config.yaml", &cfg); err != nil {
	// 	return nil, fmt.Errorf("ошибка чтения config.yaml: %w", err)
	// }

	// Подмешиваем переменные окружения в конфиг
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("ошибка чтения переменных окружения: %w", err)
	}

	return &cfg, nil
}

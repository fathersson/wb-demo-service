package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Config — конфигурация приложения
type Config struct {
	HttpServer HttpServer     `yaml:"http_server"`
	Database   DatabaseConfig `yaml:"database"`
	Kafka      KafkaConfig    `yaml:"kafka"`
}

// HttpServer — конфигурация сервера
type HttpServer struct {
	Port string `yaml:"port"`
}

// DatabaseConfig — конфигурация базы данных
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `env:"POSTGRES_PASSWORD"`
	DBname   string `yaml:"name"`
}

// KafkaConfig — конфигурация Kafka
type KafkaConfig struct {
	Broker    string `yaml:"broker"`
	Zookeeper string `yaml:"zookeeper"`
	Topic     string `yaml:"topic"`
	GroupID   string `yaml:"groupID"`
	// Commit    bool   `yaml:"commit"`
}

// Load — читаем Yaml и ENV, возвращаем конфиг
func Load() *Config {
	// Читаем переменные окружения из env файла
	_ = godotenv.Load(".env")
	// Содаем обьект конфига
	var cfg Config

	// Читаем конфиг
	if err := cleanenv.ReadConfig("config.yaml", &cfg); err != nil {
		log.Fatal(err)
	}

	// Подмешиваем переменные окружения в конфиг
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatal(err)
	}

	return &cfg
}

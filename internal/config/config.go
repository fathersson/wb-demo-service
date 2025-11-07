package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config — конфигурация приложения
type Config struct {
	HttpServer HttpServer     `yaml:"http_server"`
	Database   DatabaseConfig `yaml:"database"`
	Kafka      KafkaConfig    `yaml:"kafka"`
}

// HttpServer — конфигурация сервера
type HttpServer struct {
	Address string `yaml:"address"`
	// Host string `yaml:"host"`
	// Port string `yaml:"port"`
}

// DatabaseConfig — конфигурация базы данных
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `env:"DB_PASSWORD" yaml:"password"`
	DBname   string `yaml:"name"`
}

// KafkaConfig — конфигурация Kafka
type KafkaConfig struct {
	Broker    string `yaml:"broker"`
	Zookeeper string `yaml:"zookeeper"`
}

// Load — читаем Yaml и ENV, возвращаем конфиг
func Load() *Config {
	var cfg Config

	if err := cleanenv.ReadConfig("config.yaml", &cfg); err != nil {
		log.Fatal(err)
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatal(err)
	}

	return &cfg
}

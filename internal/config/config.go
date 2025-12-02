package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env" env-default:"prod"`
	Postgres   Postgres   `yaml:"postgres"`
	HTTPServer HTTPServer `yaml:"http_server"`
}

type Postgres struct {
	Dsn string `yaml:"dsn" env-required:"true"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"30s"`
}

func MastLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/local.yaml"
	}

	// Проверка на существование файла конфигурации
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("файл конфигурации не существует: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("не удалось прочитать конфигурацию: %v", err)
	}

	return &cfg
}

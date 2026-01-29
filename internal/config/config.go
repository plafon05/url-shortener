package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env      string `yaml:"env"`
	Postgres struct {
		Dsn string `yaml:"dsn"`
	} `yaml:"postgres"`
	HTTPServer struct {
		Address     string        `yaml:"address"`
		Timeout     time.Duration `yaml:"timeout"`
		IdleTimeout time.Duration `yaml:"idle_timeout"`
		User        string        `yaml:"user"`
		Password    string        `yaml:"password"`
	} `yaml:"http_server"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("cannot read config: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("cannot unmarshal config: %v", err)
	}

	cfg.Postgres.Dsn = os.ExpandEnv(cfg.Postgres.Dsn)
	cfg.HTTPServer.User = os.ExpandEnv(cfg.HTTPServer.User)
	cfg.HTTPServer.Password = os.ExpandEnv(cfg.HTTPServer.Password)

	return &cfg
}

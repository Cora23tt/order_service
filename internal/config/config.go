package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
}

func Load() (*Config, error) {
	_ = godotenv.Load() // не ругаемся, если .env нет — работаем с os.Getenv

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Port:        os.Getenv("PORT"),
	}

	// Простейшая валидация
	if cfg.DatabaseURL == "" || cfg.JWTSecret == "" {
		return nil, fmt.Errorf("missing required env variables")
	}

	// если порт не указан — по умолчанию :8080
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}

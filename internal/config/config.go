package config

import (
	"fmt"
	"os"

	"github.com/MaksimCpp/OzonUrlShortener/internal/utils"
	"github.com/joho/godotenv"
)

type Config struct {
	Storage string
	DBURL   string
	AppPort string
	BaseURL string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	baseURL := os.Getenv("BASE_URL")
	err := utils.CheckURL(baseURL)

	if err != nil {
		return nil, fmt.Errorf("BASE_URL: %w", err)
	}
	storage := os.Getenv("STORAGE")
	if storage == "" {
		storage = "postgres"
	}

	return &Config{
		Storage: storage,
		DBURL:   os.Getenv("DATABASE_URL"),
		AppPort: os.Getenv("APP_PORT"),
		BaseURL: baseURL,
	}, nil
}

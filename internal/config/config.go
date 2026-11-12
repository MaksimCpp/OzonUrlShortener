package config

import (
	"os"

	"github.com/MaksimCpp/OzonUrlShortener/internal/utils"
	"github.com/joho/godotenv"
)

type Config struct {
	DBURL   string
	AppHost string
	AppPort string
	BaseURL string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	baseURL := os.Getenv("BASE_URL")
	err := utils.CheckURL(baseURL)

	if err != nil {
		return nil, err
	}
	
	return &Config{
		DBURL: os.Getenv("DATABASE_URL"),
		AppHost: os.Getenv("APP_HOST"),
		AppPort: os.Getenv("APP_PORT"),
		BaseURL: baseURL,
	}, nil
}
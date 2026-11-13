package main

import (
	"context"
	"log"
	"net/http"

	"github.com/MaksimCpp/OzonUrlShortener/internal/config"
	httpdelivery "github.com/MaksimCpp/OzonUrlShortener/internal/delivery/http"
	"github.com/MaksimCpp/OzonUrlShortener/internal/repository/postgresql"
	"github.com/MaksimCpp/OzonUrlShortener/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatal(err.Error())
	}

	pool, err := pgxpool.New(context.Background(), cfg.DBURL)

	if err != nil {
		log.Fatal(err.Error())
	}

	defer pool.Close()

	urlRepo := postgresql.NewPostgreSQLURLRepository(pool)
	urlService := service.NewURLService(urlRepo, cfg.BaseURL)
	urlHanlder := httpdelivery.NewURLHandler(urlService)

	mux := http.NewServeMux()
	urlHanlder.Route(mux)

	err = http.ListenAndServe(":"+cfg.AppPort, mux)

	if err != nil {
		log.Fatal(err.Error())
	}
}

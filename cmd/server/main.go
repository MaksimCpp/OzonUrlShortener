package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/MaksimCpp/OzonUrlShortener/internal/config"
	httpdelivery "github.com/MaksimCpp/OzonUrlShortener/internal/delivery/http"
	"github.com/MaksimCpp/OzonUrlShortener/internal/domain"
	"github.com/MaksimCpp/OzonUrlShortener/internal/repository/inmemory"
	"github.com/MaksimCpp/OzonUrlShortener/internal/repository/postgresql"
	"github.com/MaksimCpp/OzonUrlShortener/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatal(err.Error())
	}

	storage := flag.String("storage", cfg.Storage, "storage backend: postgres or inmemory (overrides STORAGE)")
	flag.Parse()

	var urlRepo domain.URLRepository

	switch *storage {
	case "inmemory":
		urlRepo = inmemory.NewURLRepository()

	case "postgres":
		if cfg.DBURL == "" {
			log.Fatal("DATABASE_URL is required for postgres storage")
		}

		pool, err := pgxpool.New(context.Background(), cfg.DBURL)

		if err != nil {
			log.Fatal(err)
		}

		defer pool.Close()

		urlRepo = postgresql.NewPostgreSQLURLRepository(pool)

	default:
		log.Fatalf("unknown storage %q: expected postgres or inmemory", *storage)
	}

	log.Printf("storage: %s", *storage)
	
	urlService := service.NewURLService(urlRepo, cfg.BaseURL)
	urlHandler := httpdelivery.NewURLHandler(urlService)

	mux := http.NewServeMux()
	urlHandler.Route(mux)

	err = http.ListenAndServe(":" + cfg.AppPort, mux)

	if err != nil {
		log.Fatal(err.Error())
	}
}

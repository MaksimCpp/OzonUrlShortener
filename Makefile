COMPOSE := docker compose
POSTGRES_COMPOSE := $(COMPOSE) -f docker-compose.yaml
INMEMORY_COMPOSE := $(COMPOSE) -f docker-compose.inmemory.yaml

.PHONY: up-postgres up-inmemory down test test-service test-handler

up-postgres:
	$(MAKE) down
	$(POSTGRES_COMPOSE) up --build -d

up-inmemory:
	$(MAKE) down
	$(INMEMORY_COMPOSE) up --build -d

down:
	$(POSTGRES_COMPOSE) down --remove-orphans

test:
	go test -race ./...

test-service:
	go test -v ./internal/service

test-handler:
	go test -v ./internal/delivery/http

COMPOSE := docker compose
POSTGRES_COMPOSE := $(COMPOSE) -f docker-compose.yaml
INMEMORY_COMPOSE := $(COMPOSE) -f docker-compose.inmemory.yaml

up-postgres:
	$(MAKE) down
	$(POSTGRES_COMPOSE) up --build -d

up-inmemory:
	$(MAKE) down
	$(INMEMORY_COMPOSE) up --build -d

down:
	$(POSTGRES_COMPOSE) down --remove-orphans

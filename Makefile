ENV_FILE ?= .env
SEED_FILE ?= scripts/demo-seed.sql
POSTGRES_HOST_OVERRIDE ?=
POSTGRES_PORT_OVERRIDE ?=

.PHONY: seed seed-local

seed:
	go run ./cmd/seed --env $(ENV_FILE) --file $(SEED_FILE)

seed-local:
	$(if $(POSTGRES_PORT_OVERRIDE),POSTGRES_DATABASE_URL= POSTGRES_HOST=$(if $(POSTGRES_HOST_OVERRIDE),$(POSTGRES_HOST_OVERRIDE),localhost) POSTGRES_PORT=$(POSTGRES_PORT_OVERRIDE) go run ./cmd/seed --env $(ENV_FILE) --file $(SEED_FILE),POSTGRES_DATABASE_URL= POSTGRES_HOST=$(if $(POSTGRES_HOST_OVERRIDE),$(POSTGRES_HOST_OVERRIDE),localhost) go run ./cmd/seed --env $(ENV_FILE) --file $(SEED_FILE))

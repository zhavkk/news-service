include ../../Makefile

CONFIG_PATH=./config/config.yaml
COMPOSE_PATH=./config/docker-compose.yml
MIGRATIONS_DIR=./migrations/postgres
DB_URL=postgres://user:password@localhost:5432/news_service?sslmode=disable

.PHONY: build
build: gen-pb
	@go build -o bin/auth_service cmd/auth_service/main.go

.PHONY: run
run: build
	bin/auth_service --path $(CONFIG_PATH)

.PHONY: compose-up
compose-up:
	@docker compose -f $(COMPOSE_PATH) up -d
	@docker compose -f $(COMPOSE_PATH) exec -T postgres bash -c \
	'until pg_isready -U testuser -d auth_db; do sleep 1; done'


.PHONY: compose-down
compose-down:
	@docker compose -f $(COMPOSE_PATH) down 

.PHONY: migrate-up
migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

.PHONY: migrate-down
migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

.PHONY: migrate-create
migrate-create:
	@read -p "Enter migration name: " name; \
    goose -dir $(MIGRATIONS_DIR) create $${name} sql

.PHONY: test-deps-up
test-deps-up: compose-up migrate-up

CONFIG_PATH=./src/news/config/config.yaml
COMPOSE_PATH=./src/news/config/docker-compose.yml
MIGRATIONS_DIR=./src/news/migrations/postgres
DB_URL?=postgres://user:password@localhost:5432/news_service?sslmode=disable
DB_URL_TEST?=postgres://user:password@localhost:5432/news_service_test?sslmode=disable



.PHONY: compose-up
compose-up:
	@docker compose -f $(COMPOSE_PATH) --project-directory ./config up -d

.PHONY: compose-down
compose-down:
	@docker compose -f $(COMPOSE_PATH) --project-directory ./config down 

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

.PHONY: test-migrate-up
test-migrate-up:
	@echo "==> Applying migrations to TEST database..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL_TEST)" up

.PHONY: test-migrate-down
test-migrate-down:
	@echo "==> Rolling back migrations from TEST database..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL_TEST)" down


.PHONY: test
test: test-migrate-up
	@echo "==> Running tests..."
	@go test -v -race ./...

.PHONY: deps-up
deps-up: compose-up migrate-up


.PHONY: go-deps
go-deps:
	@echo "==> Installing Go dependencies..."
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: go-lint
go-lint:
	@echo "==> Running linter..."
	@golangci-lint run ./... -v

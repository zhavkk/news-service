SERVICE := news

.PHONY: go-lint
go-lint:
	@echo "Running Go linter..."
	@if [ -z "$$(find . -type f -name '*.go')" ]; then \
		echo "No Go files found."; \
	else \
		golangci-lint run; \
	fi


.PHONY: go-deps
go-deps:
	go mod tidy
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/golang/mock/mockgen@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: go-fmt
go-fmt:
	gofumpt -w .


.PHONY: lint
lint: go-lint 

.PHONY: deps
deps: go-deps

.PHONY: fmt
fmt: go-fmt

.PHONY: test
test:
	@echo "Running tests..."
	@if [ -z "$$(find . -type f -name '*.go')" ]; then \
		echo "No Go files found."; \
	else \
		go test ./... -v; \
	fi

.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t news-service:latest .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 news-service:latest
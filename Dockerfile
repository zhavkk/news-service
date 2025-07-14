FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o news-service ./src/news/cmd/news

FROM alpine:3.18

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/news-service .
COPY --from=builder /go/bin/goose /usr/local/bin/

COPY entrypoint.sh .

RUN mkdir -p ./src/news/config
COPY ./src/news/config/config.yml ./src/news/config/config.yml
COPY ./src/news/migrations ./src/news/migrations

ENV TZ=Europe/Moscow

ENTRYPOINT ["./entrypoint.sh"]

EXPOSE 8080

CMD ["./news-service"]
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o news-service .

FROM alpine:3.18

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/news-service .

ENV TZ=Europe/Moscow

EXPOSE 8080

CMD ["./news-service"]

# Этап сборки
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o news-service .

# Финальный этап
FROM alpine:3.18

WORKDIR /app

# Устанавливаем зависимости
RUN apk --no-cache add ca-certificates tzdata

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/news-service .

# Устанавливаем переменную окружения для временной зоны
ENV TZ=Europe/Moscow

# Определяем порт, на котором будет работать приложение
EXPOSE 8080

# Команда запуска приложения
CMD ["./news-service"]

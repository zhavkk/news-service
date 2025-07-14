# Сервис новостей (News Service)

REST API сервис для управления новостями, реализованный на языке Go. Приложение позволяет создавать, читать, обновлять, удалять и просматривать списки новостей с фильтрацией и пагинацией.

##  Основные возможности

-   **CRUD операции:** Полный набор операций для управления новостями.
-   **Поиск и фильтрация:** Получение списка новостей с поиском по заголовку и фильтрацией по категории.
-   **Пагинация:** Поддержка постраничной выдачи списка новостей.
-   **Временные рамки:** Возможность отображения новости только в заданном временном интервале (`start_time` / `end_time`). Реализовал так, что при GET запросах, параметр check_visibility изначально true. Поэтому дефолтно будут отображаться только свежие новости. При желании можно выставить в false и будут отображаться все новости. Новость доступна через API только если текущая дата и время находятся внутри указанного диапазона.
-   **Кеширование:** Использование Redis для кеширования запросов на получение новостей по ID.
-   **Логирование:** Структурированное логирование с использованием `slog`.
-   **API Спецификация:** Документация API сгенерирована с помощью Swagger.
-   **Тестирование:** Покрытие интеграционными тестами для слоя репозитория.
-   **Чистая архитектура:** Проект разделен на логические слои (handler, service, repository).

##  Технологический стек

-   **Язык:** Go 1.24
-   **База данных:** PostgreSQL
-   **Кеш:** Redis
-   **HTTP фреймворк:** Fiber
-   **Миграции:** goose
-   **Контейнеризация:** Docker, Docker Compose
-   **Документация API:** Swagger

## Инструкция по запуску

### Предварительные требования

-   [Docker](https://www.docker.com/) и [Docker Compose](https://docs.docker.com/compose/)
-   [Go](https://go.dev/) (для локального запуска без Docker)
-   [Make](https://www.gnu.org/software/make/)
-   [goose](https://github.com/pressly/goose) (для ручного управления миграциями)

### 1. Запуск через Docker Compose

1.  **Клонируйте репозиторий:**
    ```bash
    git clone <URL_вашего_репозитория>
    cd news-service
    ```

2.  **Запустите сборку и запуск контейнеров:**
    ```bash
    docker-compose up --build
    ```
    Приложение будет доступно по адресу `http://localhost:8080`.

### 2. Локальный запуск (без Docker для приложения)
Нужно будет изменить src/news/config/config.yml с имен docke-compose сервисов на localhost: в dburl и redis host. 

Приложение будет доступно по адресу `http://localhost:8080`.

## API Документация (Swagger)

После запуска приложения спецификация API будет доступна в браузере по адресу:

**[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

## Схема Базы Данных

База данных состоит из двух таблиц: `news` для хранения основной информации о новостях и `content_blocks` для хранения контента.

```sql
-- Таблица для новостей
CREATE TABLE news (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    category TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL
);

-- Таблица для блоков контента
CREATE TABLE content_blocks (
    id BIGSERIAL PRIMARY KEY,
    news_id BIGINT NOT NULL REFERENCES news(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('text','link')),
    content TEXT NOT NULL,
    position INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (news_id, position)
);

-- Индексы для ускорения запросов
CREATE INDEX idx_content_news_id ON content_blocks(news_id);
CREATE INDEX idx_news_visibility ON news(start_time, end_time);
```

Миграции находятся в директории `/src/news/migrations/postgres` и управляются с помощью `goose`.

## Тестирование

Для запуска тестов необходимо запустить базы данных и применить миграции к тестовой БД. Тестировал только слой репозитория, при наличии большего времени обязательно нужно протестировать сервисный слой, сгенерировав моки redis и postgre с помощью mockgen

1.  **Поднимите окружение для тестов:**
    ```bash
    # Запускает Docker-контейнеры и создает тестовую БД
    docker-compose up -d 
    ```
    *Если контейнеры уже запущены, убедитесь, что тестовая БД создана. При необходимости пересоздайте контейнер postgres с полной очисткой тома: `docker-compose down --volumes && docker-compose up -d`*

2.  **Запустите тесты:**
    Команда `make test` автоматически применит миграции к тестовой базе данных и запустит все тесты.
    ```bash
    make test
    ```

## Структура запросов к API

Все запросы должны отправляться на базовый URL `http://localhost:8080/api/v1`.

### 1. Создание новости

-   **Метод:** `POST`
-   **Путь:** `/news`
-   **Тело запроса:** JSON-объект с данными о новости.

```bash
curl -X POST http://localhost:8080/api/v1/news \
-H "Content-Type: application/json" \
-d '{
  "title": "Заголовок новой новости",
  "category": "Технологии",
  "start_time": "2025-07-14T00:00:00Z",
  "end_time": "2025-08-15T00:00:00Z",
  "content": [
    {
      "type": "text",
      "content": "Это текстовый блок для нашей новости.",
      "position": 1
    },
    {
      "type": "link",
      "content": "https://example.com/image.jpg",
      "position": 2
    }
  ]
}'
```

### 2. Получение списка новостей

-   **Метод:** `GET`
-   **Путь:** `/news`
-   **Параметры запроса (Query Params):**
    -   `page` (int, default: 1): Номер страницы.
    -   `limit` (int, default: 10): Количество элементов на странице.
    -   `search` (string): Поиск по заголовку.
    -   `category` (string): Фильтр по категории.
    -   `sort_by` (string, default: `created_at`): Поле для сортировки.
    -   `sort_dir` (string, default: `desc`): Направление сортировки (`asc` или `desc`).
    -   `check_visibility` (bool, default: `true`): Проверять ли временные рамки.

```bash
# Пример: получить первую страницу с 5 новостями из категории "Спорт"
curl "http://localhost:8080/api/v1/news?page=1&limit=5&category=Спорт"
```

### 3. Получение новости по ID

-   **Метод:** `GET`
-   **Путь:** `/news/{id}`

```bash
curl http://localhost:8080/api/v1/news/1
```
При этом если current_time < start_time || current_time > end_time, то выводиться ничего не будет. При демонстрации работы апишки нужно быть аккуратным чтобы убедиться что все работает. 
### 4. Обновление новости

-   **Метод:** `PUT`
-   **Путь:** `/news/{id}`
-   **Тело запроса:** JSON-объект с полями для обновления. Можно обновлять не все поля.

```bash
curl -X PUT http://localhost:8080/api/v1/news/1 \
-H "Content-Type: application/json" \
-d '{
  "title": "Это обновленный заголовок",
  "category": "Обновления"
}'
```

### 5. Удаление новости

-   **Метод:** `DELETE`
-   **Путь:** `/news/{id}`

```bash
curl -X DELETE http://localhost:8080/api/v1/news/1
```
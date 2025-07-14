#!/bin/sh
set -e

echo "Running database migrations..."
goose -dir ./src/news/migrations/postgres postgres "$DB_URL" up

echo "Starting application..."
exec "$@"
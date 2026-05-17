#!/bin/bash
set -e

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASS=${DB_PASS:-postgres}
DB_NAME=${DB_NAME:-mdm}

DSN="postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

echo "Running database migrations..."
goose -dir migrations postgres "$DSN" up

echo "Starting server..."
exec ./server

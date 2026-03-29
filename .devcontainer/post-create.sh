#!/usr/bin/env bash
set -euo pipefail

echo "==> Installing Go dependencies..."
go mod download

echo "==> Running database migrations..."
migrate -path ./db/migrations \
  -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}" \
  up || true

echo "==> Generating Swagger docs..."
swag init -g cmd/api/main.go -o docs/swagger --parseDependency --parseInternal 2>/dev/null || true

echo "==> Dev container ready!"

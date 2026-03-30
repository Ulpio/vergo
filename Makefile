.PHONY: help run dev build test test-integration test-fuzz lint fmt vet generate swagger swagger-check sqlc sqlc-check clean \
       docker-up docker-down docker-reset \
       migrate-sync migrate-up migrate-down migrate-create deps

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ── Development ──────────────────────────────────────────────────────

run: ## Run server locally
	go run cmd/api/main.go

dev: ## Run server with hot-reload (air)
	@command -v air >/dev/null 2>&1 || { echo "Installing air..."; go install github.com/air-verse/air@latest; }
	air

build: ## Compile binary to ./bin/vergo
	go build -o bin/vergo cmd/api/main.go

# ── Testing ──────────────────────────────────────────────────────────

test: ## Run unit tests
	go test ./... -count=1

test-v: ## Run unit tests (verbose)
	go test -v ./... -count=1

test-cover: ## Run tests with coverage report
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out
	@rm -f coverage.out

test-integration: ## Run integration tests (requires Docker)
	go test ./tests/integration/... -v -count=1 -tags=integration

test-fuzz: ## Run fuzz tests for 30s
	go test ./internal/http/middleware/... -fuzz=. -fuzztime=30s

# ── Code Quality ─────────────────────────────────────────────────────

lint: ## Run golangci-lint
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...
	@command -v goimports >/dev/null 2>&1 && goimports -w . || true

vet: ## Run go vet
	go vet ./...

generate: sqlc ## Run all code generation
	go generate ./...

swagger: ## Regenerate Swagger docs
	@command -v swag >/dev/null 2>&1 || { echo "Installing swag..."; go install github.com/swaggo/swag/cmd/swag@latest; }
	swag init -g cmd/api/main.go -o docs/swagger --parseDependency --parseInternal

swagger-check: swagger ## Verify Swagger docs are up-to-date (CI)
	@git diff --exit-code docs/swagger/ || (echo "Swagger docs are out of date. Run 'make swagger' and commit." && exit 1)

sqlc: ## Regenerate sqlc code
	@command -v sqlc >/dev/null 2>&1 || { echo "Installing sqlc..."; go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest; }
	sqlc generate

sqlc-check: sqlc ## Verify sqlc code is up-to-date (CI)
	@git diff --exit-code internal/repo/ || (echo "sqlc code is out of date. Run 'make sqlc' and commit." && exit 1)

check: vet lint test ## Run vet + lint + test (CI equivalent)

# ── Docker ───────────────────────────────────────────────────────────

docker-up: ## Start containers (PostgreSQL)
	docker-compose up -d

docker-down: ## Stop containers
	docker-compose down

docker-reset: ## Reset database (removes volumes)
	docker-compose down -v
	docker-compose up -d

# ── Migrations ───────────────────────────────────────────────────────

migrate-sync: ## Sync migrations between db/ and internal/pkg/db/
	./scripts/sync-migrations.sh

migrate-up: ## Run pending migrations
	migrate -path ./db/migrations \
		-database "postgres://app:app@localhost:5432/vergo?sslmode=disable" up

migrate-down: ## Revert last migration
	migrate -path ./db/migrations \
		-database "postgres://app:app@localhost:5432/vergo?sslmode=disable" down 1

migrate-create: ## Create new migration (usage: make migrate-create NAME=my_migration)
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=my_migration"; exit 1; fi
	@NEXT_NUM=$$(ls -1 db/migrations/*.up.sql 2>/dev/null | wc -l | xargs); \
	NEXT_NUM=$$(printf "%04d" $$(($$NEXT_NUM + 1))); \
	FILE="db/migrations/$${NEXT_NUM}_$(NAME).up.sql"; \
	echo "-- Migration: $(NAME)" > $$FILE; \
	echo "-- Created: $$(date)" >> $$FILE; \
	echo "" >> $$FILE; \
	cp $$FILE internal/pkg/db/migrations/$${NEXT_NUM}_$(NAME).up.sql; \
	echo "Created: $$FILE (edit and run make migrate-up)"

# ── Dependencies ─────────────────────────────────────────────────────

deps: ## Download and tidy dependencies
	go mod download
	go mod tidy

clean: ## Remove build artifacts and temp files
	rm -rf bin/ tmp/ coverage.out

.DEFAULT_GOAL := help

.PHONY: help run build test clean docker-up docker-down migrate-sync migrate-up migrate-down migrate-create

help: ## Mostra esta ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Executa o servidor localmente
	@echo "🚀 Iniciando servidor..."
	go run cmd/api/main.go

dev: ## Executa o servidor com hot-reload (air)
	@echo "🔥 Iniciando servidor com hot-reload..."
	air

build: ## Compila o binário
	@echo "🔨 Compilando..."
	go build -o bin/vergo cmd/api/main.go
	@echo "✅ Binário criado em ./bin/vergo"

test: ## Executa os testes
	@echo "🧪 Executando testes..."
	go test -v ./...

clean: ## Limpa arquivos temporários e binários
	@echo "🧹 Limpando..."
	rm -rf bin/ tmp/
	@echo "✅ Limpeza concluída"

docker-up: ## Sobe os containers (PostgreSQL + migrations)
	@echo "🐳 Subindo containers..."
	docker-compose up -d
	@echo "✅ Containers rodando"

docker-down: ## Derruba os containers
	@echo "🐳 Parando containers..."
	docker-compose down
	@echo "✅ Containers parados"

docker-reset: ## Reseta completamente o banco (remove volumes)
	@echo "⚠️  Resetando banco de dados..."
	docker-compose down -v
	docker-compose up -d
	@echo "✅ Banco resetado e migrations aplicadas"

migrate-sync: ## Sincroniza migrations entre diretórios
	@echo "🔄 Sincronizando migrations..."
	./scripts/sync-migrations.sh

migrate-up: ## Executa migrations manualmente
	@echo "⬆️  Executando migrations..."
	migrate -path ./db/migrations \
		-database "postgres://app:app@localhost:5432/vergo?sslmode=disable" up
	@echo "✅ Migrations aplicadas"

migrate-down: ## Reverte última migration
	@echo "⬇️  Revertendo última migration..."
	migrate -path ./db/migrations \
		-database "postgres://app:app@localhost:5432/vergo?sslmode=disable" down 1
	@echo "✅ Migration revertida"

migrate-create: ## Cria uma nova migration (uso: make migrate-create NAME=nome_da_migration)
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Erro: Especifique NAME=nome_da_migration"; \
		exit 1; \
	fi
	@NEXT_NUM=$$(ls -1 db/migrations/*.up.sql 2>/dev/null | wc -l | xargs); \
	NEXT_NUM=$$(printf "%04d" $$(($$NEXT_NUM + 1))); \
	FILE="db/migrations/$${NEXT_NUM}_$(NAME).up.sql"; \
	echo "-- Migration: $(NAME)" > $$FILE; \
	echo "-- Created: $$(date)" >> $$FILE; \
	echo "" >> $$FILE; \
	echo "CREATE TABLE IF NOT EXISTS example (" >> $$FILE; \
	echo "    id TEXT PRIMARY KEY," >> $$FILE; \
	echo "    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()" >> $$FILE; \
	echo ");" >> $$FILE; \
	cp $$FILE internal/pkg/db/migrations/$${NEXT_NUM}_$(NAME).up.sql; \
	echo "✅ Migration criada: $$FILE"; \
	echo "📝 Edite o arquivo e sincronize com: make migrate-sync"

deps: ## Instala dependências
	@echo "📦 Instalando dependências..."
	go mod download
	go mod tidy
	@echo "✅ Dependências instaladas"

fmt: ## Formata o código
	@echo "🎨 Formatando código..."
	go fmt ./...
	@echo "✅ Código formatado"

lint: ## Executa o linter
	@echo "🔍 Executando linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint não instalado. Instalando..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run ./...; \
	fi
	@echo "✅ Linter executado"

.DEFAULT_GOAL := help


FROM mcr.microsoft.com/devcontainers/go:1-1.25-bookworm

# Go tools used by the project
RUN go install github.com/air-verse/air@latest \
 && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
 && go install github.com/swaggo/swag/cmd/swag@latest \
 && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Python + sqlfluff (for SQL lint)
RUN apt-get update \
 && apt-get install -y --no-install-recommends python3-pip \
 && pip3 install --break-system-packages sqlfluff \
 && apt-get clean && rm -rf /var/lib/apt/lists/*

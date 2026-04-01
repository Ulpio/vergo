
# Vergo — SaaS Starter (Go)

[![CI](https://github.com/Ulpio/vergo/actions/workflows/ci.yml/badge.svg)](https://github.com/Ulpio/vergo/actions/workflows/ci.yml)
[![CD](https://github.com/Ulpio/vergo/actions/workflows/cd.yml/badge.svg)](https://github.com/Ulpio/vergo/actions/workflows/cd.yml)
[![CodeQL](https://github.com/Ulpio/vergo/actions/workflows/codeql.yml/badge.svg)](https://github.com/Ulpio/vergo/actions/workflows/codeql.yml)

Production-ready **multi-tenant SaaS boilerplate** in Go. Ships with JWT auth, RBAC, Stripe billing, webhooks, API keys, S3 storage, OpenTelemetry observability, and type-safe SQL — ready to build on.

---

## Features

| Category | What's included |
|----------|----------------|
| **Auth** | Signup, login, refresh token rotation, forgot/reset password, logout, logout-all |
| **Multi-tenant** | Organizations, memberships (owner/admin/member), tenant middleware via `X-Org-ID` |
| **RBAC** | Role-based access control per organization |
| **API Keys** | Programmatic access with `sk_...` tokens (SHA-256 hashed, optional expiry) |
| **Billing** | Stripe Checkout, subscriptions, webhook handler, plan gating middleware |
| **Webhooks** | CRUD endpoints, HMAC-SHA256 signing, dispatcher with exponential backoff retries |
| **Storage** | S3-compatible presigned uploads/downloads, file metadata tracking |
| **Audit** | Immutable audit log with actor, action, entity, metadata, and optional filters |
| **Data Layer** | PostgreSQL + sqlc type-safe generated queries |
| **Observability** | OpenTelemetry (traces + metrics), structured logging (slog), Jaeger, Prometheus |
| **Security** | Fuzz testing (JWT + RBAC), rate limiting, graceful shutdown, SQL lint (sqlfluff) |
| **DX** | Dev Container, Swagger UI, Makefile, hot-reload (air), Dependabot |
| **CI/CD** | GitHub Actions: build, test, vet, golangci-lint, sqlc check, swagger check, Docker push to GHCR |

---

## Quickstart

```bash
git clone git@github.com:Ulpio/vergo.git
cd vergo
cp .env.example .env

# Start Postgres + infra
docker compose up -d

# Run API
make run
# or with hot-reload:
make dev
```

API: http://localhost:8080/healthz
Swagger UI: http://localhost:8080/swagger/index.html

### Dev Container (VS Code / Cursor)

1. Install [Dev Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension
2. **Ctrl+Shift+P** > *"Dev Containers: Reopen in Container"*

Includes: Go 1.25, PostgreSQL 16, migrate, golangci-lint, swag, air, sqlfluff.

### Docker (Production)

```bash
docker build -t vergo .
docker run -p 8080:8080 --env-file .env vergo
```

Image is also published to GHCR on every push to main:
```bash
docker pull ghcr.io/ulpio/vergo:main
```

---

## API Endpoints

### Public (no auth)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/auth/signup` | Register |
| POST | `/v1/auth/login` | Login |
| POST | `/v1/auth/refresh` | Rotate tokens |
| POST | `/v1/auth/logout` | Revoke refresh token |
| POST | `/v1/auth/forgot-password` | Request password reset |
| POST | `/v1/auth/reset-password` | Reset password with token |
| POST | `/v1/billing/webhook` | Stripe webhook (signature verified) |

### Authenticated (Bearer JWT)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/me` | Current user |
| POST | `/v1/auth/logout-all` | Revoke all sessions |
| GET/POST | `/v1/context` | Get/set active org |
| POST | `/v1/orgs` | Create organization |
| GET | `/v1/orgs/:id` | Get organization |

### Authenticated + Tenant (requires `X-Org-ID` or active context)

| Method | Path | Roles | Description |
|--------|------|-------|-------------|
| POST | `/v1/orgs/:id/members` | admin | Add member |
| PATCH | `/v1/orgs/:id/members/:userId` | admin | Update role |
| DELETE | `/v1/orgs/:id/members/:userId` | admin | Remove member |
| DELETE | `/v1/orgs/:id` | owner | Delete org |
| GET/POST | `/v1/projects` | member+ | List/create projects |
| GET/PATCH/DELETE | `/v1/projects/:id` | member+ | Manage project |
| GET | `/v1/audit` | admin | Audit log (with filters) |
| POST/GET/DELETE | `/v1/api-keys` | member+ | Manage API keys |
| POST/GET/PATCH | `/v1/webhooks/endpoints` | member+ | Manage webhooks |
| POST | `/v1/webhooks/test` | member+ | Test webhook delivery |
| POST | `/v1/billing/checkout-session` | member+ | Create Stripe checkout |
| GET | `/v1/billing/subscription` | member+ | Get subscription |
| GET | `/v1/billing/usage` | member+ | Usage vs plan limits |
| POST/GET | `/v1/storage/presign*` | member+ | Upload/download URLs |
| POST/GET/DELETE | `/v1/storage/files*` | member+ | File metadata |

API keys (`Bearer sk_...`) can be used as an alternative to JWT for programmatic access.

---

## Project Structure

```
cmd/api/main.go                    # Entrypoint
internal/
  auth/                            # JWT, refresh store, password reset
  domain/
    user/                          # User service (signup, login, password)
    org/                           # Org + membership service
    project/                       # Project CRUD
    audit/                         # Audit log
    apikey/                        # API key management
    webhook/                       # Webhook endpoints + dispatcher
    billing/                       # Stripe billing + plan limits
    file/                          # File metadata
    userctx/                       # Active org context
  http/
    handlers/                      # HTTP handlers
    middleware/                    # Auth, tenant, RBAC, rate limit, plan
    router/                        # Route registration + wiring
  repo/                            # sqlc generated code
  pkg/
    config/                        # Env-based configuration
    db/                            # Database connection + migrations
    telemetry/                     # OpenTelemetry setup
    logging/                       # Structured logger (slog)
    ratelimit/                     # Token bucket rate limiter
  storage/s3/                      # S3-compatible storage client
db/
  migrations/                      # SQL migrations (0001-0011)
  queries/                         # sqlc query files
.github/workflows/                 # CI + CD + CodeQL + Dependabot
```

---

## Configuration

All config via environment variables. See [`.env.example`](.env.example) for defaults.

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_PORT` | 8080 | API port |
| `APP_ENV` | dev | Environment (dev/production) |
| `DB_HOST`/`DB_PORT`/`DB_USER`/`DB_PASSWORD`/`DB_NAME` | localhost/5432/app/app/vergo | PostgreSQL |
| `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` | dev-* | JWT signing keys |
| `S3_BUCKET` / `S3_ENDPOINT` | - | S3-compatible storage |
| `STRIPE_SECRET_KEY` | - | Stripe API key |
| `STRIPE_WEBHOOK_SECRET` | - | Stripe webhook signing secret |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | - | OTLP gRPC endpoint |
| `METRICS_PORT` | 0 (disabled) | Prometheus scrape port |

---

## Make Targets

```bash
make help          # Show all targets
make run           # Run server
make dev           # Run with hot-reload (air)
make test          # Unit tests
make lint          # golangci-lint
make swagger       # Regenerate Swagger docs
make sqlc          # Regenerate sqlc code
make check         # vet + lint + test (CI equivalent)
make docker-up     # Start Postgres + infra
make docker-reset  # Reset database
make migrate-up    # Run migrations
```

---

## License

[MIT](./LICENSE)

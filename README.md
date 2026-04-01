
# Vergo

**Production-ready multi-tenant SaaS backend in Go.**

[![CI](https://github.com/Ulpio/vergo/actions/workflows/ci.yml/badge.svg)](https://github.com/Ulpio/vergo/actions/workflows/ci.yml)
[![CD](https://github.com/Ulpio/vergo/actions/workflows/cd.yml/badge.svg)](https://github.com/Ulpio/vergo/actions/workflows/cd.yml)
[![CodeQL](https://github.com/Ulpio/vergo/actions/workflows/codeql.yml/badge.svg)](https://github.com/Ulpio/vergo/actions/workflows/codeql.yml)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

Vergo is a complete SaaS backend that handles the infrastructure every B2B application needs — authentication, organizations, billing, webhooks, storage, and observability — so you can focus on your product's domain logic.

---

## What can you build with Vergo?

Vergo provides the backend foundation for any multi-tenant B2B SaaS. Fork it, add your domain logic, ship:

- **Project management tool** (like Linear, Asana) — orgs, members, RBAC, and audit are ready
- **Developer platform** (like Vercel, Railway) — API keys, webhooks, and billing are built-in
- **Internal tooling platform** — tenant isolation, role-based access, and file storage out of the box
- **Collaboration SaaS** (like Notion, Figma) — multi-org, invites, and plan gating from day one
- **API-first product** — full Swagger docs, API key auth, webhook delivery with retries
- **Marketplace / Platform** — Stripe subscriptions, usage tracking, feature gating per plan

Every feature is production-grade: hashed tokens, HMAC-signed webhooks, structured logs, distributed tracing, fuzz-tested auth. Not a tutorial project.

---

## Architecture

```
                                    +-------------------+
                                    |   Stripe / S3     |
                                    +--------+----------+
                                             |
Client ──> Gin Router ──> Middleware Chain ──> Handlers ──> Domain Services ──> PostgreSQL
           (rate limit)   (auth, tenant,      (HTTP)       (business logic)    (sqlc queries)
                           RBAC, plan gate)
                                                              |
                                                     Webhook Dispatcher ──> External URLs
                                                     (background, retries)
```

### Design decisions

| Decision | Rationale |
|----------|-----------|
| **sqlc** over GORM/ent | Type-safe SQL without magic. Queries are plain `.sql` files, generated Go code is reviewed in CI |
| **Domain services** over repository pattern | Each domain owns its queries and business rules. No shared generic repository |
| **Gin** as HTTP framework | Mature, fast, great middleware ecosystem. Easy to swap if needed |
| **`database/sql`** over pgx pool | Simpler interface, works with sqlc out of the box, compatible with all Postgres tooling |
| **Hash-based tokens** (API keys, reset) | Plaintext never stored. SHA-256 hashed at rest, compared by hash lookup |
| **Background dispatcher** for webhooks | Decoupled delivery with exponential backoff. No external queue dependency for MVP |

---

## Features

| Category | What's included |
|----------|----------------|
| **Auth** | Signup, login, refresh token rotation, forgot/reset password, logout, logout-all |
| **Multi-tenant** | Organizations, memberships (owner/admin/member), tenant middleware via `X-Org-ID` |
| **RBAC** | Role-based access control per organization with `RequireRole` middleware |
| **API Keys** | Programmatic access with `sk_...` tokens (SHA-256 hashed, optional expiry) |
| **Billing** | Stripe Checkout, subscriptions, webhook handler, plan gating (`free`/`pro`/`enterprise`) |
| **Webhooks** | CRUD endpoints, HMAC-SHA256 signing, dispatcher with exponential backoff (5 retries) |
| **Storage** | S3-compatible presigned uploads/downloads with file metadata tracking |
| **Audit** | Immutable audit log with actor, action, entity, metadata, and filterable queries |
| **Data Layer** | PostgreSQL + sqlc type-safe generated queries across 11 migrations |
| **Observability** | OpenTelemetry (traces + metrics), structured logging (slog), Jaeger, Prometheus |
| **Security** | Fuzz testing (JWT + RBAC), rate limiting, graceful shutdown, SQL lint (sqlfluff) |
| **DX** | Dev Container, Swagger UI, Makefile (20+ targets), hot-reload (air), Dependabot |
| **CI/CD** | Build, test, vet, golangci-lint, sqlc + swagger freshness checks, Docker multi-arch push to GHCR |

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
```

```bash
# Register a user
curl -s localhost:8080/v1/auth/signup \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"secret123"}' | jq .

# Create an org
TOKEN=$(curl -s localhost:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"secret123"}' | jq -r .access_token)

curl -s localhost:8080/v1/orgs \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Acme Corp"}' | jq .

# Create a project (with org context)
ORG_ID=$(curl -s localhost:8080/v1/context \
  -H "Authorization: Bearer $TOKEN" | jq -r .org_id)

curl -s localhost:8080/v1/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Org-ID: $ORG_ID" \
  -H 'Content-Type: application/json' \
  -d '{"name":"My First Project","description":"Built with Vergo"}' | jq .
```

**Swagger UI**: http://localhost:8080/swagger/index.html

### Dev Container

Open in VS Code / Cursor with everything pre-installed:

1. Install [Dev Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension
2. **Ctrl+Shift+P** > *"Dev Containers: Reopen in Container"*

Includes: Go 1.25, PostgreSQL 16, migrate, golangci-lint, swag, air, sqlfluff.

### Docker (Production)

```bash
docker build -t vergo .
docker run -p 8080:8080 --env-file .env vergo
```

Pre-built images on every push to main:
```bash
docker pull ghcr.io/ulpio/vergo:main
```

---

## API Reference

Full interactive docs available at `/swagger/index.html` when running locally.

### Public

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/auth/signup` | Register |
| POST | `/v1/auth/login` | Login (returns JWT pair) |
| POST | `/v1/auth/refresh` | Rotate token pair |
| POST | `/v1/auth/logout` | Revoke refresh token |
| POST | `/v1/auth/forgot-password` | Request password reset |
| POST | `/v1/auth/reset-password` | Reset password with token |
| POST | `/v1/billing/webhook` | Stripe webhook (signature verified) |
| GET | `/healthz` | Health check |

### Authenticated (Bearer JWT or API Key)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/me` | Current user profile |
| POST | `/v1/auth/logout-all` | Revoke all sessions |
| GET/POST | `/v1/context` | Get/set active org |
| POST | `/v1/orgs` | Create organization |
| GET | `/v1/orgs/:id` | Get organization |

### Tenant-scoped (requires org context)

| Method | Path | Minimum Role | Description |
|--------|------|-------------|-------------|
| POST/PATCH/DELETE | `/v1/orgs/:id/members*` | admin | Manage members |
| DELETE | `/v1/orgs/:id` | owner | Delete organization |
| CRUD | `/v1/projects*` | member | Project management |
| GET | `/v1/audit` | admin | Filterable audit log |
| CRUD | `/v1/api-keys*` | member | API key management |
| CRUD | `/v1/webhooks/endpoints*` | member | Webhook configuration |
| POST | `/v1/webhooks/test` | member | Test webhook delivery |
| POST | `/v1/billing/checkout-session` | member | Start Stripe checkout |
| GET | `/v1/billing/subscription` | member | Current subscription |
| GET | `/v1/billing/usage` | member | Usage vs plan limits |
| CRUD | `/v1/storage/*` | member | File uploads/downloads |

---

## Project Structure

```
cmd/api/                           # Application entrypoint
internal/
  auth/                            # JWT, refresh tokens, password reset
  domain/
    user/                          # Registration, login, password management
    org/                           # Organizations + memberships
    project/                       # Project CRUD
    audit/                         # Immutable audit trail
    apikey/                        # API key lifecycle
    webhook/                       # Endpoints + background dispatcher
    billing/                       # Stripe integration + plan limits
    file/                          # File metadata
    userctx/                       # Active org context
  http/
    handlers/                      # Request/response handling
    middleware/                    # Auth, tenant, RBAC, rate limit, plan gate
    router/                        # Route registration + dependency wiring
  repo/                            # sqlc generated type-safe queries
  pkg/                             # Shared infrastructure (config, db, telemetry, logging)
  storage/s3/                      # S3-compatible storage client
db/
  migrations/                      # 11 SQL migrations
  queries/                         # sqlc query definitions
```

---

## Configuration

All configuration via environment variables. Copy `.env.example` to `.env` and adjust:

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_PORT` | `8080` | HTTP port |
| `APP_ENV` | `dev` | `dev` enables Swagger UI, `production` sets Gin to release mode |
| `DB_*` | localhost | PostgreSQL connection |
| `JWT_ACCESS_SECRET` | `dev-access` | Access token signing key |
| `JWT_REFRESH_SECRET` | `dev-refresh` | Refresh token signing key |
| `S3_BUCKET` / `S3_ENDPOINT` | - | S3-compatible storage (MinIO locally) |
| `STRIPE_SECRET_KEY` | - | Stripe API key for billing |
| `STRIPE_WEBHOOK_SECRET` | - | Stripe webhook signature verification |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | - | OTLP gRPC endpoint for traces/metrics |
| `METRICS_PORT` | `0` | Prometheus scrape port (0 = disabled) |
| `RATE_LIMIT_RPS` / `RATE_LIMIT_BURST` | `20` / `40` | Rate limiter config |

---

## Development

```bash
make help              # Show all targets
make dev               # Hot-reload with air
make test              # Unit tests
make test-fuzz         # Fuzz tests (JWT + RBAC)
make lint              # golangci-lint
make check             # vet + lint + test (CI equivalent)
make swagger           # Regenerate Swagger docs
make sqlc              # Regenerate sqlc queries
make docker-up         # Start Postgres + infra
make docker-reset      # Reset database (drop volumes)
make migrate-up        # Run pending migrations
make migrate-create NAME=xyz  # Create new migration
```

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25 |
| HTTP | Gin |
| Database | PostgreSQL 16 |
| Query generation | sqlc |
| Auth | JWT (golang-jwt/v5) + bcrypt |
| Billing | Stripe Go SDK v82 |
| Storage | AWS SDK v2 (S3-compatible) |
| Observability | OpenTelemetry, slog, Jaeger, Prometheus |
| CI/CD | GitHub Actions, GHCR, Docker multi-arch |
| Code quality | golangci-lint, sqlfluff, CodeQL, Dependabot |

---

## License

[MIT](./LICENSE)

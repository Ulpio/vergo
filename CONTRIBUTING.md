# Contributing to Vergo

Thanks for your interest in contributing!

## Setup

1. Fork and create a branch from `main`
2. `cp .env.example .env`
3. `docker compose up -d` (PostgreSQL + infra)
4. `make dev` (hot-reload) or `make run`

Or open in **Dev Container** for a zero-config environment.

## Workflow

1. Make your changes
2. Run `make check` (vet + lint + test)
3. If you changed SQL queries: `make sqlc`
4. If you changed handler annotations: `make swagger`
5. Commit following [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` new feature
   - `fix:` bug fix
   - `docs:` documentation
   - `refactor:` code restructure
   - `ci:` CI/CD changes
6. Open a pull request

## Code Guidelines

- Keep handlers thin — business logic belongs in domain services
- SQL queries live in `db/queries/`, generated with sqlc
- New tables need a migration in `db/migrations/`
- Follow standard Go conventions (`gofmt`, `go vet`)

## Tests

```bash
make test              # Unit tests
make test-fuzz         # Fuzz tests (JWT + RBAC)
make test-integration  # Integration tests (requires Docker)
```

## Code of Conduct

Be respectful and collaborative. We're all here to build something great.
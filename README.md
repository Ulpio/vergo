
# Vergo — SaaS Starter (Go)

[![CI](https://github.com/Ulpio/vergo/actions/workflows/ci.yml/badge.svg)](https://github.com/Ulpio/vergo/actions/workflows/ci.yml)
[![CodeQL](https://github.com/Ulpio/vergo/actions/workflows/codeql.yml/badge.svg)](https://github.com/Ulpio/vergo/actions/workflows/codeql.yml)

Boilerplate de **SaaS multi-tenant** escrito em **Golang**, projetado para servir como base sólida em aplicações modernas.  
Inclui autenticação JWT, RBAC, auditoria, webhooks, billing (Stripe), integração com S3 e observabilidade (OpenTelemetry).

---

## ✨ Features (MVP)
- Estrutura em camadas: **handlers → services → repo**
- **Autenticação/JWT** com refresh tokens
- **RBAC** por organização (multi-tenant)
- **Postgres + sqlc** para queries tipadas
- **Docker Compose** para ambiente local
- **CI/CD** com GitHub Actions (build + test + CodeQL + Dependabot)
- **Observabilidade** preparada (tracing/metrics/logging)

---

## 🚀 Quickstart
Clone e rode:

```bash
git clone git@github.com:Ulpio/vergo.git
cd vergo

go mod tidy
go run ./cmd/api

# API disponível em http://localhost:8080/healthz
```

Para subir Postgres + PgAdmin:
```bash
docker compose up -d db pgadmin
```

---

## 🧱 Estrutura do projeto
```
cmd/api/main.go                # entrypoint da API
internal/
  http/router/router.go        # rotas
  http/middleware/             # middlewares (auth, rbac, tenant)
  domain/{user,org,project}/   # serviços de domínio
  pkg/config/                  # configuração (env)
  repo/                        # código gerado pelo sqlc
db/
  migrations/                  # migrations SQL
  queries/                     # queries do sqlc
scripts/seed/                  # seed de dados
.github/workflows/             # CI/CD
```

---

## 🧭 Roadmap
- [ ] Implementar **Auth** (signup/login/refresh)
- [ ] Adicionar **Tenant + RBAC**
- [ ] CRUD de **Projects** + Audit log
- [ ] Integração com **Stripe** (billing)
- [ ] Upload com **S3 presign**
- [ ] Observabilidade com OpenTelemetry

---

## 📜 Licença
[MIT](./LICENSE)

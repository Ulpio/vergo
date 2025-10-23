# 🚀 Sistema de Migrations Automáticas

Este projeto possui um sistema de migrations totalmente **automático** que garante que seu banco de dados esteja sempre atualizado.

## ✨ Execução Automática

As migrations são executadas **automaticamente** nas seguintes situações:

### 🖥️ 1. Ao Iniciar o Servidor da Aplicação

Quando você executa o servidor, as migrations são aplicadas antes do servidor iniciar:

```bash
# Forma direta
go run cmd/api/main.go

# Com hot-reload (recomendado para desenvolvimento)
air

# Ou usando o Makefile
make run
make dev
```

Você verá mensagens como:
```
🔌 Conectando ao banco de dados...
✅ Conexão com o banco estabelecida
🚀 Executando migrations...
✅ migrations aplicadas com sucesso
```

### 🐳 2. Ao Subir o PostgreSQL via Docker

Quando você usa o Docker Compose, as migrations são executadas automaticamente:

```bash
# Subir os containers
docker-compose up

# Ou com o Makefile
make docker-up
```

O serviço `migrate` aguarda o PostgreSQL ficar saudável e então aplica todas as migrations pendentes.

## 📝 Como Criar uma Nova Migration

### Opção 1: Usando o Makefile (Recomendado)

```bash
make migrate-create NAME=adiciona_campo_status
```

Isso irá:
1. Criar o arquivo `db/migrations/XXXX_adiciona_campo_status.up.sql`
2. Criar uma cópia em `internal/pkg/db/migrations/` (para o embed)
3. Incluir um template básico para você começar

### Opção 2: Manualmente

1. Crie um arquivo em `db/migrations/` com o próximo número (formato: `XXXX_nome.up.sql`):
   ```
   db/migrations/0007_sua_migration.up.sql
   ```

2. Escreva seu SQL:
   ```sql
   -- Migration: adiciona campo status aos projetos
   
   ALTER TABLE projects ADD COLUMN status TEXT NOT NULL DEFAULT 'active';
   CREATE INDEX idx_projects_status ON projects(status);
   ```

3. Sincronize as migrations:
   ```bash
   make migrate-sync
   # ou
   ./scripts/sync-migrations.sh
   ```

## 🔄 Sincronização de Migrations

As migrations existem em **dois locais**:

1. **`db/migrations/`** - Arquivos fonte (use este para criar/editar)
2. **`internal/pkg/db/migrations/`** - Cópia embedada no binário Go

Sempre que modificar algo em `db/migrations/`, execute:

```bash
make migrate-sync
```

Isso garante que as alterações sejam embedadas no binário.

## 🛠️ Comandos Úteis

```bash
# Ver todos os comandos disponíveis
make help

# Subir ambiente de desenvolvimento
make docker-up

# Resetar banco completamente (remove dados!)
make docker-reset

# Executar migrations manualmente (requer migrate CLI)
make migrate-up

# Reverter última migration
make migrate-down

# Criar nova migration
make migrate-create NAME=nome_descritivo
```

## 🎯 Fluxo de Trabalho Recomendado

### Desenvolvimento Local

1. Suba o PostgreSQL:
   ```bash
   make docker-up
   ```

2. Execute o servidor (que aplicará migrations):
   ```bash
   make dev
   ```

3. Ao adicionar nova migration:
   ```bash
   make migrate-create NAME=minha_feature
   # Edite o arquivo criado
   make dev  # Reinicie o servidor
   ```

### Produção

As migrations estão **embedadas no binário**, então basta:

1. Compilar o binário:
   ```bash
   make build
   ```

2. Executar (migrations serão aplicadas automaticamente):
   ```bash
   ./bin/vergo
   ```

## 🔍 Verificação Manual

Se quiser verificar o status das migrations manualmente:

```bash
# Instalar migrate CLI (se não tiver)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Ver status
migrate -path ./db/migrations \
  -database "postgres://app:app@localhost:5432/vergo?sslmode=disable" \
  version
```

## 📂 Estrutura Atual das Migrations

```
0001_users_orgs.up.sql       - Usuários, organizações, memberships
0002_projects.up.sql         - Projetos
0003_refresh_tokens.up.sql   - Tokens de refresh
0004_audit_logs.up.sql       - Logs de auditoria
0005_user_contexts.up.sql    - Contextos de usuário
0006_files.up.sql            - Arquivos/storage
```

> **Nota:** As migrations seguem o formato `{número}_{descrição}.up.sql` conforme exigido pelo golang-migrate.

## 💡 Dicas

- ✅ **Sempre** use `CREATE TABLE IF NOT EXISTS` para segurança
- ✅ **Sempre** adicione índices necessários na mesma migration
- ✅ **Teste** suas migrations em desenvolvimento antes de enviar para produção
- ✅ **Documente** migrations complexas com comentários SQL
- ✅ **Sincronize** após criar/editar migrations (`make migrate-sync`)
- ⚠️ **Cuidado** ao reverter migrations em produção

## 🐛 Troubleshooting

### "Erro ao carregar migrations do embed"

Execute `make migrate-sync` para sincronizar os diretórios.

### "no such table"

As migrations não foram executadas. Reinicie o servidor ou execute `make docker-reset`.

### "dirty database version"

O banco está em estado inconsistente. Execute:

```bash
make docker-reset  # Remove tudo e recria
```

### Quero rodar migrations sem subir o servidor

```bash
make migrate-up
```

Requer a ferramenta `migrate` CLI instalada.

## 📚 Mais Informações

- [golang-migrate](https://github.com/golang-migrate/migrate) - Biblioteca utilizada
- [Makefile](./Makefile) - Comandos disponíveis (`make help`)


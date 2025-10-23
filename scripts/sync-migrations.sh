#!/bin/bash
# Script para sincronizar migrations entre os diretórios

set -e

SOURCE_DIR="./db/migrations"
TARGET_DIR="./internal/pkg/db/migrations"

echo "🔄 Sincronizando migrations..."
echo "   Origem: $SOURCE_DIR"
echo "   Destino: $TARGET_DIR"

# Criar diretório de destino se não existir
mkdir -p "$TARGET_DIR"

# Copiar apenas arquivos .up.sql (migrations)
cp -v "$SOURCE_DIR"/*.up.sql "$TARGET_DIR/" 2>/dev/null || true

echo "✅ Migrations sincronizadas com sucesso!"
echo ""
echo "📝 Arquivos copiados:"
ls -1 "$TARGET_DIR"/*.up.sql | xargs -n 1 basename


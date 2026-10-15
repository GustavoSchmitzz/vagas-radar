#!/usr/bin/env bash
set -e

DB_NAME="${POSTGRES_DB:-vagas_radar}"
DB_USER="${POSTGRES_USER:-postgres}"
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"

echo "==> Verificando/Criando banco de dados '${DB_NAME}' no PostgreSQL..."

if psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -lqt | cut -d \| -f 1 | grep -qw "${DB_NAME}"; then
    echo "✅ Banco de dados '${DB_NAME}' já existe!"
else
    echo "==> Criando banco '${DB_NAME}'..."
    createdb -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" "${DB_NAME}" || psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -c "CREATE DATABASE ${DB_NAME};"
    echo "✅ Banco '${DB_NAME}' criado com sucesso!"
fi

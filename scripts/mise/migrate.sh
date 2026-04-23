#!/usr/bin/env bash

set -euo pipefail

if ! command -v migrate >/dev/null 2>&1; then
    echo "comando obrigatorio nao encontrado: migrate" >&2
    exit 1
fi

host="${LOCAL_POSTGRES_HOST:-${POSTGRES_HOST:-localhost}}"
port="${LOCAL_POSTGRES_PORT:-${POSTGRES_PORT:-5432}}"
user="${POSTGRES_USER:-}"
password="${POSTGRES_PASSWORD:-}"
db="${POSTGRES_DB:-}"
sslmode="${POSTGRES_SSLMODE:-disable}"

if [[ -z "$user" || -z "$password" || -z "$db" ]]; then
    echo "variaveis insuficientes para montar a URL local do Postgres (POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB)" >&2
    exit 1
fi

db_url="postgres://${user}:${password}@${host}:${port}/${db}?sslmode=${sslmode}"
max_retries="${MIGRATION_MAX_RETRIES:-30}"
sleep_seconds="${MIGRATION_RETRY_DELAY_SECONDS:-2}"
attempt=1

until migrate -path "./migrations" -database "$db_url" up; do
    if [[ "$attempt" -ge "$max_retries" ]]; then
        echo "erro ao executar as migrations apos ${attempt} tentativas" >&2
        exit 1
    fi

    echo "falha ao executar migrations. tentativa ${attempt}/${max_retries}; aguardando ${sleep_seconds}s..."
    attempt=$((attempt + 1))
    sleep "$sleep_seconds"
done

echo "migrations aplicadas com sucesso"

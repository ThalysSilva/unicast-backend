#!/usr/bin/env bash

set -euo pipefail

CONTAINER_NAME="${POSTGRES_CONTAINER_NAME:-postgres-unicast}"
SEED_FILE="${SEED_FILE:-scripts/demo-seed.sql}"
POSTGRES_USER="${POSTGRES_USER:-}"
POSTGRES_DB="${POSTGRES_DB:-}"

if [[ -z "$POSTGRES_USER" || -z "$POSTGRES_DB" ]]; then
    echo "variaveis insuficientes para executar seed (POSTGRES_USER, POSTGRES_DB)" >&2
    exit 1
fi

if [[ ! -f "$SEED_FILE" ]]; then
    echo "arquivo de seed nao encontrado: $SEED_FILE" >&2
    exit 1
fi

docker exec -i "$CONTAINER_NAME" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" < "$SEED_FILE"

echo "seed aplicada com sucesso em ${CONTAINER_NAME}/${POSTGRES_DB}"

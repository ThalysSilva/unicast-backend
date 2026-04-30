#!/usr/bin/env bash

set -euo pipefail

CONTAINER_NAME="${API_CONTAINER_NAME:-unicast-api}"
MIGRATIONS_PATH="${MIGRATIONS_PATH:-/root/migrations}"
ACTION="${MIGRATE_ACTION:-up}"

docker exec -i "$CONTAINER_NAME" sh -c "migrate -path '$MIGRATIONS_PATH' -database \"\$POSTGRES_DATABASE_URL\" '$ACTION'"

echo "migrations executadas com sucesso em ${CONTAINER_NAME} (${ACTION})"

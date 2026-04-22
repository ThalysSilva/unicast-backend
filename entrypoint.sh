#!/bin/sh

set -eu

MAX_RETRIES="${MIGRATION_MAX_RETRIES:-30}"
SLEEP_SECONDS="${MIGRATION_RETRY_DELAY_SECONDS:-2}"
ATTEMPT=1

until migrate -path /root/migrations -database "$POSTGRES_DATABASE_URL" up; do
    if [ "$ATTEMPT" -ge "$MAX_RETRIES" ]; then
        echo "Erro ao executar as migrações após ${ATTEMPT} tentativas"
        exit 1
    fi

    echo "Falha ao executar as migrações. Tentativa ${ATTEMPT}/${MAX_RETRIES}; aguardando ${SLEEP_SECONDS}s..."
    ATTEMPT=$((ATTEMPT + 1))
    sleep "$SLEEP_SECONDS"
done

exec ./unicast-api

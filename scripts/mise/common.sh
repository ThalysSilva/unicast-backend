#!/usr/bin/env bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

profile="${1:-}"

if [[ -z "$profile" ]]; then
    echo "uso: $0 <dev|prod>" >&2
    exit 1
fi

case "$profile" in
    dev)
        ENV_FILE="$PROJECT_ROOT/.env.development"
        COMPOSE_FILE="$PROJECT_ROOT/docker-compose-dev.yaml"
        PROFILE_NAME="development"
        ;;
    prod)
        ENV_FILE="$PROJECT_ROOT/.env"
        COMPOSE_FILE="$PROJECT_ROOT/docker-compose.yaml"
        PROFILE_NAME="production"
        ;;
    *)
        echo "perfil invalido: $profile (esperado: dev ou prod)" >&2
        exit 1
        ;;
esac

require_command() {
    local cmd="${1:?missing command name}"
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "comando obrigatorio nao encontrado: $cmd" >&2
        exit 1
    fi
}

load_env() {
    if [[ ! -f "$ENV_FILE" ]]; then
        echo "arquivo de ambiente nao encontrado: $ENV_FILE" >&2
        exit 1
    fi

    set -a
    # shellcheck disable=SC1090
    source "$ENV_FILE"
    set +a
}

compose_cmd() {
    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

postgres_local_url() {
    local host port user password db sslmode

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

    printf 'postgres://%s:%s@%s:%s/%s?sslmode=%s' \
        "$user" \
        "$password" \
        "$host" \
        "$port" \
        "$db" \
        "$sslmode"
}

#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./common.sh
source "$SCRIPT_DIR/common.sh" "${PROFILE:-prod}"

MODE="${1:-direct}"
ENV_FILE_OVERRIDE="${ENV_FILE_OVERRIDE:-${ENV_FILE:-$PROJECT_ROOT/.env}}"
SEED_FILE="${SEED_FILE:-scripts/demo-seed.sql}"
POSTGRES_HOST_OVERRIDE="${POSTGRES_HOST_OVERRIDE:-}"
POSTGRES_PORT_OVERRIDE="${POSTGRES_PORT_OVERRIDE:-}"

require_command go

cd "$PROJECT_ROOT"

case "$MODE" in
    direct)
        exec go run ./cmd/seed --env "$ENV_FILE_OVERRIDE" --file "$SEED_FILE"
        ;;
    local)
        unset POSTGRES_DATABASE_URL
        export POSTGRES_HOST="${POSTGRES_HOST_OVERRIDE:-localhost}"
        if [[ -n "$POSTGRES_PORT_OVERRIDE" ]]; then
            export POSTGRES_PORT="$POSTGRES_PORT_OVERRIDE"
        fi
        exec go run ./cmd/seed --env "$ENV_FILE_OVERRIDE" --file "$SEED_FILE"
        ;;
    *)
        echo "modo invalido: $MODE (esperado: direct ou local)" >&2
        exit 1
        ;;
esac

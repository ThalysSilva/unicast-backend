#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./common.sh
source "$SCRIPT_DIR/common.sh" "${1:-}"
load_env
require_command ss

PORT_TO_FREE="${2:-${POSTGRES_PORT:-5432}}"

if command -v fuser >/dev/null 2>&1; then
    if fuser "${PORT_TO_FREE}/tcp" >/dev/null 2>&1; then
        echo "encerrando processos na porta ${PORT_TO_FREE} via fuser"
        fuser -k "${PORT_TO_FREE}/tcp" >/dev/null 2>&1 || true
        sleep 1
    fi
fi

SS_OUTPUT="$(ss -ltnp 2>/dev/null || true)"

if [[ -z "$SS_OUTPUT" ]]; then
    echo "nao foi possivel inspecionar sockets com ss. Se necessario, rode manualmente com permissao elevada: sudo ss -ltnp | grep ':${PORT_TO_FREE}'" >&2
    exit 1
fi

LINES="$(printf '%s\n' "$SS_OUTPUT" | grep ":${PORT_TO_FREE}\\b" || true)"

if [[ -z "$LINES" ]]; then
    echo "nenhum processo escutando na porta ${PORT_TO_FREE}"
    exit 0
fi

PIDS="$(printf '%s\n' "$LINES" | grep -o 'pid=[0-9]\+' | cut -d= -f2 | sort -u)"

if [[ -z "$PIDS" ]]; then
    echo "a porta ${PORT_TO_FREE} continua ocupada, mas ss nao revelou PIDs utilizaveis. Isso costuma acontecer com processos root/docker no WSL." >&2
    echo "inspecione manualmente com: sudo ss -ltnp | grep ':${PORT_TO_FREE}'" >&2
    echo "mate manualmente os PIDs encontrados ou derrube os containers/processos correspondentes." >&2
    printf '%s\n' "$LINES"
    exit 1
fi

echo "encerrando PIDs na porta ${PORT_TO_FREE}: ${PIDS//$'\n'/ }"
kill $PIDS
sleep 1

REMAINING="$(ss -ltnp 2>/dev/null | grep ":${PORT_TO_FREE}\\b" || true)"
if [[ -n "$REMAINING" ]]; then
    echo "alguns processos continuam na porta ${PORT_TO_FREE}; aplicando SIGKILL"
    kill -9 $PIDS 2>/dev/null || true
fi

FINAL_STATE="$(ss -ltnp 2>/dev/null | grep ":${PORT_TO_FREE}\\b" || true)"
if [[ -n "$FINAL_STATE" ]]; then
    echo "a porta ${PORT_TO_FREE} continua ocupada apos as tentativas automaticas." >&2
    echo "inspecione manualmente com: sudo ss -ltnp | grep ':${PORT_TO_FREE}'" >&2
    printf '%s\n' "$FINAL_STATE"
    exit 1
fi

echo "porta ${PORT_TO_FREE} liberada"

#!/bin/sh

migrate -path /root/migrations -database "$POSTGRES_DATABASE_URL" up

if [ $? -ne 0 ]; then
    echo "Erro ao executar as migrações"
    exit 1
fi

exec ./unicast-api
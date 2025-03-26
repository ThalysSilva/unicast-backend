#!/bin/bash
set -a
source .env
set +a
swag init -g cmd/main.go
go run cmd/main.go
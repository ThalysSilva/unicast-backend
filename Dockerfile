FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o unicast-api ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz -C /usr/local/bin

WORKDIR /root/
COPY --from=builder /app/unicast-api .
COPY migrations /root/migrations
COPY entrypoint.sh /root/entrypoint.sh
RUN chmod +x /root/entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/root/entrypoint.sh"]
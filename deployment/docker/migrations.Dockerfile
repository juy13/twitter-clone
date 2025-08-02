FROM golang:1.24-alpine AS migration_runner

WORKDIR /app

RUN go install github.com/pressly/goose/v3/cmd/goose@latest
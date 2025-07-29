FROM golang:1.24 AS builder
WORKDIR /app

COPY ../../ .
RUN go mod download
RUN go mod tidy

WORKDIR /app
RUN mkdir /app/build

WORKDIR /app/cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o ../../build/api

FROM alpine:latest AS twitter-clone-api
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/build/api .

FROM golang:1.24 AS builder
WORKDIR /app

COPY . .
RUN go mod download
RUN go mod tidy

WORKDIR /app
RUN mkdir /app/build

RUN cd /app/cmd/api && CGO_ENABLED=0 GOOS=linux go build -o ../../build/api && \
    cd /app/cmd/worker && CGO_ENABLED=0 GOOS=linux go build -o ../../build/worker && \
    cd /app/cmd/ws_server && CGO_ENABLED=0 GOOS=linux go build -o ../../build/ws_server 

FROM alpine:latest AS twitter-clone-api
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/build/api .

FROM alpine:latest AS twitter-clone-worker
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/build/worker .

FROM alpine:latest AS twitter-clone-ws_socket
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/build/ws_server .

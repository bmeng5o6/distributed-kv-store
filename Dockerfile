FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o kv-store ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/kv-store .
EXPOSE 8080
ENTRYPOINT ["./kv-store"]
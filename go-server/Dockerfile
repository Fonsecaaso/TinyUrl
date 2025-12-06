# Etapa 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server .

# Etapa 2: Imagem final mínima
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/server .
COPY .env .

EXPOSE 8080

# O app vai ler variáveis de ambiente para conectar ao RDS
CMD ["/app/server"]

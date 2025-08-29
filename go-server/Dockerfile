# Usar a imagem oficial do Go
FROM golang:1.21

# Instalar PostgreSQL client para migrations
RUN apt-get update && apt-get install -y postgresql-client && rm -rf /var/lib/apt/lists/*

# Definir o diretório de trabalho no container
WORKDIR /app

# Copiar o código fonte para o diretório de trabalho
COPY . .

# Baixar as dependências do Go
RUN go mod tidy

# Expor a porta 8080
EXPOSE 8080

# Comando para rodar o app diretamente com go run
CMD ["go", "run", "."]

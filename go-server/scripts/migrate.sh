#!/bin/bash

# Migration script para validar e executar migrations
set -e

# Aguardar PostgreSQL estar disponível
echo "Aguardando PostgreSQL estar disponível..."
until pg_isready -h postgres -p 5432 -U postgres; do
  echo "PostgreSQL não está pronto ainda. Aguardando..."
  sleep 2
done

echo "PostgreSQL está disponível!"

# Verificar se a tabela urls existe
echo "Verificando se as migrations foram executadas..."

TABLE_EXISTS=$(PGPASSWORD=password psql -h postgres -U postgres -d postgres -t -c \
  "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'urls');")

if [[ "$TABLE_EXISTS" =~ "t" ]]; then
  echo "Migration já executada. Tabela 'urls' existe."
else
  echo "Executando migrations..."
  PGPASSWORD=password psql -h postgres -U postgres -d postgres -f /app/migrations/create_urls.sql
  echo "Migration executada com sucesso!"
fi

echo "Iniciando aplicação..."
exec go run .
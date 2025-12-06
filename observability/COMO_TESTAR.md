# 🚀 Como Testar os Logs

## Para Desenvolvimento Local (Go direto)

### 1. Certifique-se que Loki está rodando

```bash
cd observability
docker-compose up -d loki grafana
```

### 2. Aguarde Loki ficar pronto (~15s)

```bash
sleep 15
curl http://localhost:3100/ready
# Deve retornar: ready
```

### 3. Inicie a aplicação Go

```bash
cd go-server
go run main.go
```

### 4. Faça requisições para gerar logs

```bash
# Health check
curl http://localhost:8080/api/health

# Criar URL
curl -X POST http://localhost:8080/api/ \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/test"}'

# Ver métricas
curl http://localhost:8080/metrics
```

### 5. Visualize no Grafana

1. Abra: http://localhost:3000
2. Login: `admin` / `admin`
3. Vá para **Explore** (ícone de bússola)
4. Use a query:

```logql
{service_name="tinyurl-api"}
```

### 6. Queries úteis

```logql
# Ver todos os logs
{service_name="tinyurl-api"}

# Filtrar por nível (debug, info, warn, error)
{service_name="tinyurl-api"} | json | level="info"
{service_name="tinyurl-api"} | json | level="error"

# Buscar por texto específico
{service_name="tinyurl-api"} |= "postgres"
{service_name="tinyurl-api"} |= "Request completed"

# Logs de um endpoint específico
{service_name="tinyurl-api"} | json | path="/api/health"

# Contagem de logs por nível (últimos 5 minutos)
sum by(level) (count_over_time({service_name="tinyurl-api"} | json [5m]))

# Rate de logs por segundo
rate({service_name="tinyurl-api"}[1m])
```

---

## Para Produção (Docker)

### 1. Adicionar serviço no docker-compose

Edite `observability/docker-compose.yml`:

```yaml
services:
  # ... outros serviços ...

  tinyurl-api:
    build: ../go-server
    container_name: tinyurl-api
    ports:
      - "8080:8080"
    environment:
      # Database
      - POSTGRES_HOST=tiny-url-db.comzh9adefdo.us-east-1.rds.amazonaws.com
      - POSTGRES_PORT=5432
      - POSTGRES_DB=postgres
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}

      # Loki (usar nome do container)
      - LOKI_URL=http://loki:3100/loki/api/v1/push

      # Redis
      - REDIS_ADDR=redis:6379
    depends_on:
      - loki
      - prometheus
    networks:
      - observability

networks:
  observability:
    driver: bridge
```

### 2. Build e Run

```bash
cd observability
docker-compose up -d
```

### 3. Verificar logs

```bash
# Logs do container
docker logs tinyurl-api -f

# Verificar se está enviando para Loki
curl "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={service_name="tinyurl-api"}' \
  --data-urlencode 'limit=10' | jq '.'
```

---

## Troubleshooting

### ❌ Erro: "connection refused" ao conectar no Loki

**Problema:** App rodando localmente não consegue acessar Loki no Docker

**Solução:**
```bash
# Certifique-se que Loki está expondo a porta
docker ps | grep loki
# Deve mostrar: 0.0.0.0:3100->3100/tcp

# Teste a conectividade
curl http://localhost:3100/ready
```

### ❌ Logs não aparecem no Grafana

**Causa 1: Loki não está pronto**
```bash
curl http://localhost:3100/ready
# Se retornar erro, aguarde mais tempo
sleep 10 && curl http://localhost:3100/ready
```

**Causa 2: Logs em buffer**
Os logs são enviados de forma assíncrona. Aguarde 2-3 segundos após gerar logs.

**Causa 3: Query incorreta**
Use exatamente:
```logql
{service_name="tinyurl-api"}
```

**Verificar se logs estão chegando:**
```bash
curl -G "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={service_name="tinyurl-api"}' \
  --data-urlencode 'limit=10' | jq '.status, .data.result | length'
```

### ❌ "failed to send log to loki" no stderr

**Causa:** URL do Loki incorreta

**Para desenvolvimento local:**
```bash
LOKI_URL="http://localhost:3100/loki/api/v1/push"
```

**Para Docker:**
```bash
LOKI_URL="http://loki:3100/loki/api/v1/push"
```

---

## 📊 Dashboards Úteis no Grafana

### Dashboard de Logs por Nível

```logql
sum by(level) (count_over_time({service_name="tinyurl-api"} | json [5m]))
```

### Dashboard de Erros

```logql
sum(count_over_time({service_name="tinyurl-api"} | json | level="error" [5m]))
```

### Dashboard de Latência

```logql
{service_name="tinyurl-api"}
| json
| latency != ""
| line_format "{{.latency}}"
```

### Dashboard de Requests por Endpoint

```logql
sum by(path) (count_over_time({service_name="tinyurl-api"} | json | path != "" [5m]))
```

---

## ✅ Checklist Final

Antes de considerar "funcionando", verifique:

- [ ] Loki está rodando e "ready"
- [ ] Grafana está acessível em http://localhost:3000
- [ ] Datasource Loki está configurado no Grafana
- [ ] Aplicação Go está rodando sem erros
- [ ] Logs aparecem no console da aplicação
- [ ] Não há erros "connection refused" no stderr
- [ ] Query `{service_name="tinyurl-api"}` retorna logs no Grafana
- [ ] Timeline no Grafana mostra atividade

---

## 🎯 Teste Rápido Completo

Execute este script para testar tudo:

```bash
#!/bin/bash
echo "🧪 Testando pipeline de logs..."
echo ""

# 1. Verificar Loki
echo "1️⃣ Verificando Loki..."
if curl -s http://localhost:3100/ready > /dev/null; then
    echo "✅ Loki está pronto"
else
    echo "❌ Loki não está acessível"
    exit 1
fi

# 2. Verificar Grafana
echo "2️⃣ Verificando Grafana..."
if curl -s http://localhost:3000/api/health > /dev/null; then
    echo "✅ Grafana está rodando"
else
    echo "❌ Grafana não está acessível"
    exit 1
fi

# 3. Verificar aplicação
echo "3️⃣ Verificando aplicação..."
if curl -s http://localhost:8080/api/health > /dev/null; then
    echo "✅ Aplicação está rodando"
else
    echo "❌ Aplicação não está acessível"
    exit 1
fi

# 4. Gerar logs
echo "4️⃣ Gerando logs..."
for i in {1..5}; do
    curl -s http://localhost:8080/api/health > /dev/null
    echo "  Request $i enviado"
done

# 5. Aguardar batching
echo "5️⃣ Aguardando logs serem processados (3s)..."
sleep 3

# 6. Verificar logs no Loki
echo "6️⃣ Verificando logs no Loki..."
RESULT=$(curl -s -G "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={service_name="tinyurl-api"}' \
  --data-urlencode 'limit=10' | jq -r '.status')

if [ "$RESULT" == "success" ]; then
    echo "✅ Logs estão chegando no Loki!"
    echo ""
    echo "🎉 Tudo funcionando!"
    echo ""
    echo "📊 Acesse o Grafana:"
    echo "   URL: http://localhost:3000"
    echo "   Login: admin / admin"
    echo "   Query: {service_name=\"tinyurl-api\"}"
else
    echo "❌ Logs não encontrados no Loki"
    exit 1
fi
```

Salve como `test-logs.sh`, dê permissão e execute:
```bash
chmod +x test-logs.sh
./test-logs.sh
```

---

**Status:** ✅ Sistema configurado e pronto para uso!

# 🏷️ Labels Indexadas no Loki

## O que mudou?

Agora os logs são enviados ao Loki com **labels indexadas** extraídas automaticamente do JSON. Isso permite filtros muito mais rápidos e eficientes.

## Labels Disponíveis

### Labels Estáticas (sempre presentes)
- `service_name`: Nome do serviço (ex: "tinyurl-api")
- `environment`: Ambiente (ex: "development", "production")
- `job`: Job name (ex: "tinyurl-api")

### Labels Dinâmicas (extraídas do log JSON)
- `level`: Nível do log (`info`, `warn`, `error`, `debug`)
- `path`: Path HTTP da requisição (ex: `/api/health`, `/metrics`)
- `method`: Método HTTP (ex: `GET`, `POST`, `PUT`, `DELETE`)
- `status`: Status code HTTP (ex: `200`, `404`, `500`)

## Como Usar no Grafana

### 1. Filtrar por Endpoint Específico

```logql
{service_name="tinyurl-api", path="/api/health"}
```

### 2. Filtrar por Método HTTP

```logql
{service_name="tinyurl-api", method="POST"}
```

### 3. Filtrar por Nível de Log

```logql
{service_name="tinyurl-api", level="error"}
```

### 4. Combinar Múltiplos Filtros

```logql
{service_name="tinyurl-api", path="/api/", method="POST", status="200"}
```

### 5. Todos os Erros HTTP (4xx e 5xx)

```logql
{service_name="tinyurl-api", status=~"[45].*"}
```

### 6. Apenas Endpoints de API (excluindo métricas)

```logql
{service_name="tinyurl-api", path=~"/api/.*"}
```

### 7. Todos os POST requests

```logql
{service_name="tinyurl-api", method="POST"}
```

## Queries Úteis para Dashboards

### Taxa de Requisições por Endpoint

```logql
sum by(path) (rate({service_name="tinyurl-api", path!=""}[5m]))
```

### Taxa de Erros por Status Code

```logql
sum by(status) (rate({service_name="tinyurl-api", status=~"[45].*"}[5m]))
```

### Contagem de Logs por Nível

```logql
sum by(level) (count_over_time({service_name="tinyurl-api"}[5m]))
```

### Top 5 Endpoints Mais Acessados

```logql
topk(5, sum by(path) (count_over_time({service_name="tinyurl-api", path!=""}[1h])))
```

### Taxa de Sucesso por Endpoint (2xx)

```logql
sum by(path) (rate({service_name="tinyurl-api", status=~"2.*"}[5m]))
```

### Erros 500 por Endpoint

```logql
{service_name="tinyurl-api", status=~"5.*"} | json
```

## Comparação: Antes vs Depois

### ❌ Antes (SEM labels indexadas)

```logql
# Tinha que parsear JSON em TODA busca (LENTO)
{service_name="tinyurl-api"} | json | path="/api/health"
```

### ✅ Depois (COM labels indexadas)

```logql
# Usa índice do Loki (RÁPIDO)
{service_name="tinyurl-api", path="/api/health"}
```

## Vantagens

1. **Performance**: Queries até 100x mais rápidas
2. **Cardinality Control**: Labels são limitadas, não explodem o índice
3. **Dashboards Eficientes**: Agregações funcionam melhor
4. **Alertas Precisos**: Alertar apenas em endpoints críticos

## ⚠️ Importante: Cardinality

Labels indexadas aumentam a cardinalidade. Por isso, apenas campos com **valores limitados** foram escolhidos:

- ✅ `path`: Poucos endpoints (~10-20)
- ✅ `method`: Apenas GET, POST, PUT, DELETE, PATCH
- ✅ `status`: Códigos HTTP limitados
- ✅ `level`: Apenas 4 valores (debug, info, warn, error)

**NÃO indexamos:**
- ❌ `request_id`: Valores únicos (cardinalidade infinita)
- ❌ `ip`: Muitos IPs diferentes
- ❌ `latency`: Valores contínuos

Esses campos ainda estão no JSON e podem ser filtrados com `| json | field=value`.

## Exemplo Completo: Debug de Endpoint Lento

```logql
# 1. Ver logs do endpoint específico
{service_name="tinyurl-api", path="/api/", method="POST"}

# 2. Analisar latências (ainda precisa de | json para campos não indexados)
{service_name="tinyurl-api", path="/api/", method="POST"}
| json
| latency > 100ms

# 3. Ver taxa de requests
rate({service_name="tinyurl-api", path="/api/", method="POST"}[5m])

# 4. Ver erros específicos
{service_name="tinyurl-api", path="/api/", method="POST", status=~"[45].*"}
```

## Testando as Labels

### 1. Reinicie o container Docker

```bash
cd /Users/mateusfonsecapiris/Documents/git/TinyUrl/go-server
docker build --platform linux/x86_64 -t tiny-url .
docker run --rm -p 8080:8080 -e REDIS_ADDR="localhost:23234" tiny-url
```

### 2. Gere alguns logs

```bash
# Health checks
curl http://localhost:8080/api/health

# Criar URL (POST)
curl -X POST http://localhost:8080/api/ \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'

# Métricas
curl http://localhost:8080/metrics
```

### 3. Verifique no Grafana

1. Abra http://localhost:3000
2. Vá para **Explore**
3. Use a query:

```logql
{service_name="tinyurl-api", path="/api/health"}
```

4. Você deve ver apenas logs do endpoint `/api/health`!

## Troubleshooting

### Labels não aparecem no Grafana

**Causa:** Logs antigos (antes do update) não têm as labels.

**Solução:**
1. Gere novos logs fazendo requests
2. Aguarde 2-3 segundos
3. Recarregue o Grafana

### "Unknown label" error

**Causa:** Loki ainda não viu nenhum log com essa label.

**Solução:** Gere logs que contenham essa label primeiro.

---

**Status:** ✅ Labels indexadas implementadas e prontas para uso!

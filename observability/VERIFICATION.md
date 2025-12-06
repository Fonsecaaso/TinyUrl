# ✅ Verificação da Configuração de Logs

Este documento mostra como verificar se toda a stack de logging está configurada corretamente.

## 🔍 Checklist de Verificação

### 1. Verificar Serviços Docker

```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(grafana|loki|otel|tempo|prometheus)"
```

**Esperado:**
- ✅ grafana - Up - porta 3000
- ✅ loki - Up - porta 3100
- ✅ otel-collector - Up - portas 4317, 4318
- ✅ tempo - Up - porta 3200
- ✅ prometheus - Up - porta 9090

### 2. Verificar Loki

```bash
# Status do Loki
curl http://localhost:3100/ready

# Métricas do Loki
curl -s http://localhost:3100/metrics | grep "loki_ingester_chunks_created_total"
```

**Esperado:**
- Resposta: `ready`
- Métricas devem aparecer

### 3. Verificar OTel Collector

```bash
# Verificar se está rodando
docker logs otel-collector --tail 20

# Verificar endpoint OTLP HTTP
curl -i http://localhost:4318
```

**Esperado:**
- Logs sem erros críticos
- Endpoint responde (mesmo que com erro 405 - método não permitido é ok)

### 4. Verificar Datasources no Grafana

```bash
# Listar datasources
curl -s -u admin:admin http://localhost:3000/api/datasources | jq '.[] | {name: .name, type: .type, isDefault: .isDefault}'
```

**Esperado:**
```json
{
  "name": "Loki",
  "type": "loki",
  "isDefault": true
}
{
  "name": "Prometheus",
  "type": "prometheus",
  "isDefault": false
}
{
  "name": "Tempo",
  "type": "tempo",
  "isDefault": false
}
```

### 5. Testar Conectividade Loki ← OTel Collector

```bash
# Verificar logs do OTel Collector
docker logs otel-collector 2>&1 | grep -i "loki" | tail -10

# Forçar envio de logs (se a aplicação Go estiver rodando)
docker logs otel-collector -f
```

**Esperado:**
- Sem erros de conexão com Loki
- Logs sendo exportados com sucesso

### 6. Verificar Aplicação Go

```bash
# Verificar se está rodando
curl http://localhost:8080/api/health

# Verificar variável de ambiente
cd go-server
grep OTEL_EXPORTER_OTLP_ENDPOINT .env
```

**Esperado:**
- API respondendo
- Variável configurada: `OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4318"`

## 🧪 Teste End-to-End

### Passo 1: Gerar Logs na Aplicação

```bash
# Fazer algumas requisições
curl -X POST http://localhost:8080/api/ \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/test"}'

curl http://localhost:8080/api/health
```

### Passo 2: Aguardar Processamento

```bash
# Aguardar batching do OTel Collector (até 10 segundos)
sleep 12
```

### Passo 3: Verificar Logs no Loki

```bash
# Query direto no Loki
curl -G -s "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={service_name="tinyurl-api"}' \
  --data-urlencode 'limit=10' | jq '.status, .data.result | length'
```

**Esperado:**
```
"success"
1
```

### Passo 4: Verificar no Grafana

1. Abra: http://localhost:3000
2. Login: `admin` / `admin`
3. Vá para **Explore** (ícone de bússola no menu lateral)
4. Verifique se **Loki** está selecionado no topo
5. Use a query:

```logql
{service_name="tinyurl-api"}
```

6. Clique em **Run Query**

**Esperado:**
- Logs devem aparecer
- Timeline com atividade deve estar visível

## 🔧 Troubleshooting

### Problema: Loki não está pronto

```bash
# Verificar logs do Loki
docker logs loki --tail 50

# Reiniciar Loki
cd observability
docker-compose restart loki

# Aguardar 15 segundos
sleep 15
curl http://localhost:3100/ready
```

### Problema: Datasources não aparecem no Grafana

```bash
# Verificar configuração
cat observability/grafana/provisioning/datasources/datasources.yml

# Verificar permissões
ls -la observability/grafana/provisioning/datasources/

# Reiniciar Grafana
cd observability
docker-compose restart grafana

# Verificar logs do Grafana
docker logs grafana --tail 50 | grep -i datasource
```

### Problema: Logs não chegam no Loki

```bash
# 1. Verificar se Go está enviando para OTel
cd go-server
# Procure por logs indicando inicialização do logger

# 2. Verificar se OTel está recebendo
docker logs otel-collector --tail 50 | grep -i "log"

# 3. Verificar se OTel está enviando para Loki
docker logs otel-collector 2>&1 | grep -i "loki\|error"

# 4. Verificar configuração do OTel
cat observability/otel-collector/otel.yaml | grep -A 20 "logs:"

# 5. Testar conectividade OTel → Loki
docker exec otel-collector wget -O- http://loki:3100/ready
```

### Problema: Erro "connection refused" na aplicação Go

```bash
# Verificar se OTel Collector está acessível
curl http://localhost:4318 -v

# Verificar se está rodando
docker ps | grep otel-collector

# Verificar variável de ambiente
echo $OTEL_EXPORTER_OTLP_ENDPOINT
# ou
grep OTEL_EXPORTER_OTLP_ENDPOINT go-server/.env

# Deve ser: localhost:4318 (sem http://)
```

## 📊 Queries de Teste no Grafana

Após confirmar que os logs estão chegando, teste estas queries:

### Query Básica
```logql
{service_name="tinyurl-api"}
```

### Filtrar por Nível
```logql
{service_name="tinyurl-api"} | json | level="info"
{service_name="tinyurl-api"} | json | level="error"
```

### Buscar Texto
```logql
{service_name="tinyurl-api"} |= "postgres"
{service_name="tinyurl-api"} |= "redis"
{service_name="tinyurl-api"} |~ "error|failed"
```

### Métricas
```logql
# Contagem de logs por nível
sum by(level) (count_over_time({service_name="tinyurl-api"} | json [5m]))

# Rate de logs
rate({service_name="tinyurl-api"}[1m])
```

## ✅ Checklist Final

Após executar todos os testes, você deve ter:

- [ ] Todos os containers Docker rodando
- [ ] Loki respondendo "ready"
- [ ] OTel Collector sem erros nos logs
- [ ] Grafana com 3 datasources configurados (Loki, Prometheus, Tempo)
- [ ] Loki como datasource padrão no Grafana
- [ ] Aplicação Go rodando e enviando logs
- [ ] Logs visíveis no Loki via query direta
- [ ] Logs visíveis no Grafana Explore
- [ ] Logs estruturados com campos JSON parseáveis

## 🎯 Resultado Esperado

Se tudo estiver funcionando corretamente:

1. **Console da Aplicação Go**: Logs aparecem no stdout
2. **OTel Collector**: Recebe logs via OTLP e envia para Loki (visível nos logs debug)
3. **Loki**: Armazena logs e responde a queries
4. **Grafana**: Exibe logs de forma visual e permite queries LogQL

## 📝 Script Automatizado

Para automatizar toda essa verificação, use:

```bash
./observability/test-logs.sh
```

Este script verifica automaticamente todos os pontos acima e gera um relatório.

---

**Última atualização:** 2025-12-05
**Versão:** 1.0

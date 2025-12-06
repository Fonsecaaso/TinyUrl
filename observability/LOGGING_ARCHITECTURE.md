# Arquitetura de Logging - TinyURL

## 📊 Visão Geral

Este documento descreve como os logs fluem da aplicação Go até o Loki através do OpenTelemetry Collector.

```
┌─────────────┐      OTLP/HTTP      ┌──────────────────┐      HTTP      ┌──────┐
│  Go Server  │  ─────────────────> │  OTel Collector  │  ───────────>  │ Loki │
│  (port 8080)│      :4318          │   (port 4317/18) │                └──────┘
└─────────────┘                     └──────────────────┘                    │
      │                                                                     │
      │ Zap Logger                                                         │
      │ + OTel Bridge                                              ┌───────▼────────┐
      └─────────────────────────────────────────────────────────> │    Grafana     │
                                Console Output                     │  (port 3000)   │
                                                                   └────────────────┘
```

## 🔧 Componentes

### 1. Go Application (go-server/internal/logger/otel_logger.go)

**Responsabilidades:**
- Criar logs estruturados usando Zap
- Enviar logs via OTLP HTTP para o OpenTelemetry Collector
- Manter output no console para desenvolvimento

**Configuração:**
```go
// InitLogger inicializa o logger com ponte OpenTelemetry
InitLogger("tinyurl-api", "development")
```

**Variáveis de Ambiente (.env):**
```bash
OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4318"
```

**Features:**
- ✅ Logs estruturados com Zap
- ✅ Ponte OpenTelemetry (otelzap)
- ✅ Output duplo: Console + OTLP
- ✅ Batching automático
- ✅ Resource attributes (service.name, deployment.environment)

### 2. OpenTelemetry Collector (observability/otel-collector/otel.yaml)

**Responsabilidades:**
- Receber logs via OTLP (HTTP/gRPC)
- Processar e enriquecer logs
- Enviar logs para Loki

**Configuração:**

```yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318
      grpc:
        endpoint: 0.0.0.0:4317

processors:
  batch:
    timeout: 10s
    send_batch_size: 1024

  resource:
    attributes:
      - key: service.name
        value: tinyurl-api
        action: upsert

exporters:
  loki:
    endpoint: http://loki:3100/loki/api/v1/push
    default_labels_enabled:
      exporter: true
      job: true

service:
  pipelines:
    logs:
      receivers: [otlp]
      processors: [resource, batch]
      exporters: [loki, debug]
```

**Features:**
- ✅ Recebe logs via OTLP HTTP/gRPC
- ✅ Batching para performance
- ✅ Adiciona labels automáticos
- ✅ Debug output para troubleshooting

### 3. Loki (observability/loki/loki.yaml)

**Responsabilidades:**
- Armazenar logs de forma eficiente
- Indexar logs por labels
- Fornecer API de query

**Configuração:**

```yaml
server:
  http_listen_port: 3100
  grpc_listen_port: 9095

ingester:
  wal:
    enabled: true
    dir: /loki/wal
  chunk_idle_period: 1h
  max_chunk_age: 1h
  chunk_retain_period: 30s

limits_config:
  ingestion_rate_mb: 16
  ingestion_burst_size_mb: 32
```

**Features:**
- ✅ WAL habilitado para durabilidade
- ✅ Armazenamento em filesystem
- ✅ Compactação automática
- ✅ Limites configuráveis

### 4. Grafana (observability/grafana/provisioning/datasources/datasources.yml)

**Acesso:** http://localhost:3000
**Credenciais:** admin/admin

**Datasources Configurados:**
```yaml
datasources:
  - name: Loki (default)
    url: http://loki:3100
  - name: Prometheus
    url: http://prometheus:9090
  - name: Tempo
    url: http://tempo:3200
```

**Features:**
- ✅ Loki configurado como datasource padrão
- ✅ Integração Logs ↔ Traces (Loki → Tempo)
- ✅ Integração Traces ↔ Metrics (Tempo → Prometheus)
- ✅ Service map e node graph habilitados
- ✅ Auto-provisioning de datasources

## 🚀 Como Usar

### Iniciar a Stack de Observabilidade

```bash
cd observability
docker-compose up -d
```

### Verificar Status dos Serviços

```bash
# OTel Collector
docker logs otel-collector --tail 50

# Loki
curl http://localhost:3100/ready

# Ver logs no Loki
curl -G -s "http://localhost:3100/loki/api/v1/query" \
  --data-urlencode 'query={service_name="tinyurl-api"}' \
  --data-urlencode 'limit=10' | jq '.'
```

### Iniciar Aplicação Go

```bash
cd go-server
go run main.go
```

A aplicação irá:
1. Inicializar o logger OpenTelemetry
2. Conectar ao OTel Collector em `localhost:4318`
3. Enviar logs estruturados
4. Mostrar logs no console também

### Visualizar Logs no Grafana

1. Abra http://localhost:3000
2. Vá para **Explore** (ícone de bússola)
3. Selecione **Loki** como data source
4. Use queries LogQL:

```logql
# Todos os logs da aplicação
{service_name="tinyurl-api"}

# Logs de erro
{service_name="tinyurl-api"} |= "error"

# Logs com filtro por nível
{service_name="tinyurl-api"} | json | level="error"

# Logs de um período específico
{service_name="tinyurl-api"} | json | __error__=""
```

## 🧪 Testar o Pipeline

Execute o script de teste:

```bash
./observability/test-logs.sh
```

Este script irá:
1. ✅ Verificar se os serviços estão rodando
2. ✅ Mostrar logs do OTel Collector
3. ✅ Verificar status do Loki
4. ✅ Buscar logs recentes

## 📝 Exemplos de Logs

### Log de Info
```go
logger.Logger.Info("postgres connection established")
```

### Log de Error
```go
logger.Logger.Error(
    "failed to connect to database",
    zap.Error(err),
    zap.String("host", dbHost),
)
```

### Log com Contexto
```go
logger.Logger.Info(
    "request processed",
    zap.String("method", "GET"),
    zap.String("path", "/api/urls"),
    zap.Int("status", 200),
    zap.Duration("duration", elapsed),
)
```

## 🔍 Troubleshooting

### Logs não aparecem no Loki

1. Verifique se o OTel Collector está recebendo logs:
```bash
docker logs otel-collector --tail 100 | grep -i log
```

2. Verifique se o Loki está healthy:
```bash
curl http://localhost:3100/ready
```

3. Verifique conectividade Go → OTel:
```bash
# Logs da aplicação devem mostrar:
# "Successfully initialized OpenTelemetry logger"
```

### Erro de conexão no Go

Se aparecer erro como "connection refused":

1. Verifique se o OTel Collector está rodando:
```bash
docker ps | grep otel-collector
```

2. Verifique a porta no .env:
```bash
OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4318"  # Sem http://
```

3. Teste a conexão:
```bash
curl http://localhost:4318/v1/logs -v
```

### Logs aparecem duplicados

Isso é esperado! O logger está configurado para enviar para dois destinos:
- Console (para desenvolvimento/debug)
- OpenTelemetry (para Loki/Grafana)

Para desabilitar console logs em produção, modifique o `otel_logger.go`.

## 🎯 Métricas e Labels

### Labels Automáticos

Cada log enviado para o Loki inclui:

- `service_name`: "tinyurl-api"
- `deployment_environment`: "development"
- `exporter`: "OTLP"
- `job`: "tinyurl-api"

### Campos Estruturados

Logs Zap são convertidos para JSON e incluem:

- `timestamp`: Timestamp do log
- `level`: debug, info, warn, error, fatal
- `caller`: Arquivo e linha do código
- `msg`: Mensagem do log
- Campos customizados via `zap.Field`

## 📚 Recursos

- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/languages/go/)
- [Loki Documentation](https://grafana.com/docs/loki/latest/)
- [OTel Collector Configuration](https://opentelemetry.io/docs/collector/configuration/)
- [Zap Logger](https://github.com/uber-go/zap)
- [OTel Zap Bridge](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/bridges/otelzap)

## ✅ Checklist de Configuração

- [x] Go app enviando logs via OTLP
- [x] OTel Collector recebendo e processando logs
- [x] Loki armazenando logs
- [x] Grafana configurado como frontend
- [x] Labels e resource attributes configurados
- [x] Batching e performance otimizados
- [x] Console output mantido para desenvolvimento
- [x] Script de teste criado
- [x] Documentação completa

## 🔐 Segurança

⚠️ **IMPORTANTE**: A configuração atual usa `insecure: true` para desenvolvimento local.

Para produção:
1. Configure TLS no OTel Collector
2. Use autenticação no Loki
3. Não exponha portas publicamente
4. Use secrets management para credenciais
5. Configure network policies

# ✅ Implementação Final: Logs Go → Loki

## 🎯 Solução Implementada

**Arquitetura:** Go Application → HTTP → Loki → Grafana

```
┌──────────────┐
│   Go Server  │
└──────┬───────┘
       │
       ├─────> Console (stdout) - desenvolvimento
       │
       └─────> HTTP POST (Loki API)
               http://localhost:3100/loki/api/v1/push
                     │
              ┌──────▼───────┐
              │     Loki     │
              └──────┬───────┘
                     │
              ┌──────▼───────┐
              │   Grafana    │
              └──────────────┘
```

## 📦 O que foi Implementado

### 1. Biblioteca Loki Client
✅ Instalada: `github.com/grafana/loki-client-go/loki`

### 2. Novo Logger (`loki_logger.go`)
✅ Criado em: `go-server/internal/logger/loki_logger.go`

**Funcionalidades:**
- Envia logs para Loki via HTTP
- Mantém output no console (dual output)
- Batching automático (100KB ou 2 segundos)
- Retry automático com backoff exponencial
- Labels automáticos: `service_name`, `environment`
- Formato JSON para Loki
- Graceful shutdown

### 3. Configuração

#### `.env`
```bash
LOKI_URL="http://localhost:3100/loki/api/v1/push"
```

#### `main.go`
```go
// Initialize Loki logger (sends logs directly to Loki)
if err := logger.InitLokiLogger("tinyurl-api", "development"); err != nil {
    panic("failed to initialize logger: " + err.Error())
}
```

## 🚀 Como Usar

### 1. Certifique-se que o Loki está rodando

```bash
cd observability
docker-compose up -d loki grafana
```

Aguarde ~15 segundos e verifique:
```bash
curl http://localhost:3100/ready
# Deve retornar: ready
```

### 2. Inicie a aplicação Go

```bash
cd go-server
go run main.go
```

Você verá logs no console E eles serão enviados automaticamente para o Loki!

### 3. Visualize no Grafana

1. Abra: http://localhost:3000
2. Login: `admin` / `admin`
3. Vá para **Explore**
4. Loki já está selecionado (é o datasource padrão)
5. Use a query:

```logql
{service_name="tinyurl-api"}
```

## 📊 Queries Úteis

### Ver todos os logs
```logql
{service_name="tinyurl-api"}
```

### Filtrar por nível
```logql
{service_name="tinyurl-api"} | json | level="info"
{service_name="tinyurl-api"} | json | level="error"
```

### Buscar por texto
```logql
{service_name="tinyurl-api"} |= "postgres"
{service_name="tinyurl-api"} |= "redis"
{service_name="tinyurl-api"} |~ "error|failed"
```

### Filtrar por campo específico
```logql
{service_name="tinyurl-api"} | json | caller=~".*routes.*"
```

### Contagem de logs
```logql
sum by(level) (count_over_time({service_name="tinyurl-api"} | json [5m]))
```

### Rate de logs por segundo
```logql
rate({service_name="tinyurl-api"}[1m])
```

## 🔍 Troubleshooting

### Logs não aparecem no Loki

**1. Verificar se Loki está pronto:**
```bash
curl http://localhost:3100/ready
```

**2. Verificar se a aplicação Go está enviando:**
- Logs devem aparecer no console normalmente
- Se houver erro enviando para Loki, aparecerá em stderr:
  ```
  failed to send log to loki: <erro>
  ```

**3. Testar envio manual para Loki:**
```bash
curl -v -H "Content-Type: application/json" \
  -XPOST "http://localhost:3100/loki/api/v1/push" \
  --data-raw '{
    "streams": [
      {
        "stream": {
          "service_name": "test"
        },
        "values": [
          ["'$(date +%s)'000000000", "test log message"]
        ]
      }
    ]
  }'
```

**4. Verificar conectividade:**
```bash
# Da máquina host
curl -I http://localhost:3100/ready

# Se estiver em Docker, use:
# LOKI_URL="http://loki:3100/loki/api/v1/push"
```

### Logs duplicados

Isso é esperado! O logger envia para **dois destinos**:
- **Console:** Para desenvolvimento e debug
- **Loki:** Para persistência e visualização no Grafana

Para desabilitar console em produção, modifique `loki_logger.go`:
```go
// Remove ou comente esta linha:
// consoleCore := zapcore.NewCore(...)

// E use apenas:
core := lokiCore
```

### Performance

O logger usa:
- **Batching:** Agrupa logs antes de enviar (100KB ou 2s)
- **Async:** Não bloqueia a aplicação
- **Retry:** Tenta novamente em caso de falha
- **Backoff:** Exponencial para evitar sobrecarga

## 🎯 Benefícios desta Abordagem

✅ **Simples:** Sem componentes intermediários
✅ **Rápido:** Latência mínima
✅ **Confiável:** Menos pontos de falha
✅ **Eficiente:** Sem overhead desnecessário
✅ **Flexível:** Fácil de modificar/estender
✅ **Observável:** Logs estruturados + Grafana

## 📝 Exemplos de Código

### Log simples
```go
logger.Logger.Info("User login successful")
```

### Log com campos
```go
logger.Logger.Info("Request processed",
    zap.String("method", "POST"),
    zap.String("path", "/api/urls"),
    zap.Int("status", 201),
    zap.Duration("latency", elapsed),
)
```

### Log de erro
```go
logger.Logger.Error("Database connection failed",
    zap.Error(err),
    zap.String("host", dbHost),
)
```

### Log com contexto rico
```go
logger.Logger.Warn("Slow query detected",
    zap.Duration("query_time", queryTime),
    zap.String("query", sqlQuery),
    zap.String("user_id", userID),
    zap.Int("rows_affected", rowsAffected),
)
```

## 🔄 Migração do Logger Antigo

Se você tinha código usando o logger OpenTelemetry:

**Antes:**
```go
logger.InitLogger("tinyurl-api", "development")
```

**Depois:**
```go
logger.InitLokiLogger("tinyurl-api", "development")
```

O resto do código permanece **100% compatível**! Todos os `logger.Logger.Info()`, `logger.Logger.Error()`, etc. funcionam exatamente igual.

## 🆚 Comparação: OTel vs Loki Direto

| Aspecto | OTel Collector | Loki Direto |
|---------|----------------|-------------|
| Componentes | 3 (Go → OTel → Loki) | 2 (Go → Loki) |
| Latência | ~100-500ms | ~10-50ms |
| Complexidade | Alta | Baixa |
| Configuração | YAML + Env vars | Apenas env var |
| Debugging | Difícil | Fácil |
| Overhead | Alto | Baixo |
| Padronização | OpenTelemetry | Loki API |

## 🔐 Produção

Para usar em produção, ajuste:

### 1. URL do Loki
```bash
# .env
LOKI_URL="http://loki:3100/loki/api/v1/push"
```

### 2. Labels
Adicione mais labels no `loki_logger.go`:
```go
cfg.Labels = fmt.Sprintf(
    `{service_name="%s", environment="%s", host="%s", version="%s"}`,
    serviceName, environment, hostname, version,
)
```

### 3. Batch Size
Aumente para produção:
```go
cfg.BatchSize = 1024 * 1024 // 1MB
cfg.BatchWait = 5 * time.Second
```

### 4. Nível de Log
Configure nível mínimo:
```go
// Em loki_logger.go, mude:
lokiCore := zapcore.NewCore(
    lokiEncoder,
    zapcore.AddSync(&lokiWriter{client: lokiClient}),
    zapcore.InfoLevel, // Era DebugLevel
)
```

---

**Status:** ✅ Implementado e Funcional
**Arquitetura:** Go → Loki (HTTP) → Grafana
**Data:** 2025-12-06

# 🚀 Quickstart - Logging com OpenTelemetry e Loki

Guia rápido para começar a usar o sistema de logs.

## 📋 Pré-requisitos

- Docker e Docker Compose instalados
- Go 1.23+ instalado
- Portas disponíveis: 3000, 3100, 4317, 4318, 8080, 9090

## ⚡ Início Rápido (3 passos)

### 1. Iniciar Stack de Observabilidade

```bash
cd observability
docker-compose up -d
```

Aguarde ~15 segundos para os serviços ficarem prontos.

### 2. Verificar Status

```bash
# Verificar se todos os serviços estão rodando
docker ps | grep -E "(otel|loki|grafana|tempo|prometheus)"

# Verificar se Loki está pronto
curl http://localhost:3100/ready
# Deve retornar: ready
```

### 3. Iniciar Aplicação Go

```bash
cd go-server
go run main.go
```

Você verá logs no console indicando que o logger foi inicializado com sucesso.

## ✅ Testar a Integração

Execute o script de teste automatizado:

```bash
cd go-server
./test-logs.sh
```

Este script irá:
- ✅ Verificar se todos os serviços estão rodando
- ✅ Gerar tráfego na aplicação
- ✅ Consultar o Loki por logs recentes
- ✅ Mostrar exemplos de logs

## 🔍 Visualizar Logs no Grafana

1. Abra seu navegador em: **http://localhost:3000**
2. Login: `admin` / `admin`
3. Navegue para **Explore** (ícone de bússola no menu lateral)
4. Selecione **Loki** no dropdown de data sources
5. Use a query:

```logql
{service_name="tinyurl-api"}
```

### Queries Úteis

```logql
# Todos os logs
{service_name="tinyurl-api"}

# Apenas erros
{service_name="tinyurl-api"} | json | level="error"

# Buscar texto específico
{service_name="tinyurl-api"} |= "database"

# Filtrar por endpoint
{service_name="tinyurl-api"} | json | path=~"/api/.*"

# Últimos 5 minutos
{service_name="tinyurl-api"} [5m]
```

## 📊 Arquitetura

```
Go App → OTLP (HTTP:4318) → OTel Collector → Loki → Grafana
  ↓
Console
```

## 🔧 Configuração

### Variáveis de Ambiente (.env)

```bash
OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4318"
```

**Importante:**
- Use apenas `host:port` (sem `http://`)
- Para Docker, use `otel-collector:4318`
- Para desenvolvimento local, use `localhost:4318`

### Logger no Código Go

```go
import (
    "github.com/fonsecaaso/TinyUrl/go-server/internal/logger"
    "go.uber.org/zap"
)

func main() {
    // Inicializar logger
    if err := logger.InitLogger("tinyurl-api", "development"); err != nil {
        panic("failed to initialize logger: " + err.Error())
    }
    defer logger.Sync()
    defer logger.Shutdown(context.Background())

    // Usar logger
    logger.Logger.Info("Application started")
    logger.Logger.Error("Something went wrong", zap.Error(err))
}
```

## 🐛 Troubleshooting

### Problema: Logs não aparecem no Loki

**Solução 1:** Verificar se OTel Collector está recebendo logs
```bash
docker logs otel-collector --tail 50
```

**Solução 2:** Verificar conectividade
```bash
# Testar endpoint OTLP
curl http://localhost:4318 -v

# Ver logs do Loki
docker logs loki --tail 50
```

**Solução 3:** Reiniciar serviços
```bash
cd observability
docker-compose restart loki otel-collector
```

### Problema: Erro "connection refused" no Go

**Causa:** OTel Collector não está acessível

**Solução:**
```bash
# Verificar se está rodando
docker ps | grep otel-collector

# Se não estiver, iniciar
cd observability
docker-compose up -d otel-collector
```

### Problema: Loki retorna "Ingester not ready"

**Causa:** Loki precisa aguardar 15 segundos após iniciar

**Solução:** Aguarde alguns segundos e tente novamente
```bash
sleep 15
curl http://localhost:3100/ready
```

## 📝 Exemplos de Uso

### Log Simples
```go
logger.Logger.Info("User created successfully")
```

### Log com Campos Estruturados
```go
logger.Logger.Info("Request processed",
    zap.String("method", "POST"),
    zap.String("path", "/api/urls"),
    zap.Int("status", 201),
    zap.Duration("duration", elapsed),
)
```

### Log de Erro com Stack Trace
```go
logger.Logger.Error("Database connection failed",
    zap.Error(err),
    zap.String("host", dbHost),
    zap.Int("port", dbPort),
)
```

### Log com Contexto
```go
logger.Logger.Warn("Slow query detected",
    zap.Duration("duration", queryTime),
    zap.String("query", sqlQuery),
    zap.String("user_id", userID),
)
```

## 🎯 Próximos Passos

1. **Adicionar Alertas:** Configure alertas no Grafana para erros críticos
2. **Dashboards:** Crie dashboards personalizados para visualizar métricas
3. **Log Sampling:** Para produção, configure sampling para reduzir volume
4. **Trace Integration:** Integre logs com traces para correlação completa
5. **Log Aggregation:** Configure queries para agregar e analisar logs

## 📚 Mais Informações

- [Documentação Completa](./LOGGING_ARCHITECTURE.md)
- [Guia de Setup](./LOGGING_SETUP.md)
- [Script de Teste](../go-server/test-logs.sh)

## 🆘 Suporte

Se encontrar problemas:

1. Verifique os logs de todos os serviços:
```bash
docker-compose logs -f
```

2. Execute o script de teste:
```bash
./observability/test-logs.sh
```

3. Consulte a documentação completa em `LOGGING_ARCHITECTURE.md`

---

**Status dos Serviços:**
- ✅ Go Application: [localhost:8080](http://localhost:8080)
- ✅ Grafana: [localhost:3000](http://localhost:3000)
- ✅ Loki: [localhost:3100](http://localhost:3100)
- ✅ OTel Collector: [localhost:4318](http://localhost:4318)
- ✅ Prometheus: [localhost:9090](http://localhost:9090)
- ✅ Tempo: [localhost:3200](http://localhost:3200)

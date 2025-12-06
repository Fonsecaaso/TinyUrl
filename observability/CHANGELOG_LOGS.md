# 📝 Changelog - Configuração de Logs

## 2025-12-05 - Configuração Completa do Pipeline de Logs

### ✅ Correções Realizadas

#### 1. Endpoint OTLP Corrigido
**Arquivo:** `go-server/.env`
- ❌ Antes: `OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318"`
- ✅ Depois: `OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4318"`
- **Motivo:** A biblioteca OTLP HTTP espera apenas `host:port`, não a URL completa

#### 2. Logger Go Atualizado
**Arquivo:** `go-server/internal/logger/otel_logger.go`
- ✅ Corrigido parsing do endpoint (linha 27-32)
- ✅ Adicionado comentário explicativo
- ✅ Default correto: `localhost:4318`

#### 3. OTel Collector Melhorado
**Arquivo:** `observability/otel-collector/otel.yaml`
- ✅ Adicionados labels padrão no exporter Loki:
  ```yaml
  default_labels_enabled:
    exporter: true
    job: true
  ```
- **Benefício:** Melhor organização e filtro de logs no Loki

#### 4. Loki Otimizado
**Arquivo:** `observability/loki/loki.yaml`
- ✅ Adicionado `replication_factor: 1` no ingester
- ✅ Adicionado `chunk_retain_period: 30s`
- ✅ Adicionada seção `limits_config`:
  ```yaml
  limits_config:
    enforce_metric_name: false
    reject_old_samples: true
    reject_old_samples_max_age: 168h
    ingestion_rate_mb: 16
    ingestion_burst_size_mb: 32
  ```
- **Benefício:** Melhor performance e controle de ingestão

#### 5. **CRÍTICO:** Grafana Datasources Configurados
**Arquivo:** `observability/grafana/provisioning/datasources/datasources.yml`
- ❌ **Problema Encontrado:** Grafana NÃO estava configurado para ler do Loki!
- ✅ **Solução:** Criada configuração completa de datasources:
  - **Loki** (padrão) - http://loki:3100
  - **Prometheus** - http://prometheus:9090
  - **Tempo** - http://tempo:3200
- ✅ Configurada integração Logs ↔ Traces ↔ Metrics
- ✅ Loki definido como datasource padrão

### 📁 Arquivos Criados

#### 1. Documentação Completa
- ✅ `observability/LOGGING_ARCHITECTURE.md` - Arquitetura detalhada
- ✅ `observability/QUICKSTART.md` - Guia rápido de início
- ✅ `observability/VERIFICATION.md` - Checklist de verificação

#### 2. Scripts de Teste
- ✅ `observability/test-logs.sh` - Teste automatizado do pipeline

### 🎯 Funcionalidades Implementadas

#### Pipeline Completo
```
Go App (Zap Logger)
    ↓ OTLP HTTP (localhost:4318)
OTel Collector (Batching + Processing)
    ↓ HTTP (loki:3100/loki/api/v1/push)
Loki (Storage + Indexing)
    ↓ HTTP API
Grafana (Visualization)
```

#### Features
- ✅ Logs estruturados com Zap
- ✅ Ponte OpenTelemetry (otelzap)
- ✅ Output duplo: Console + OTLP
- ✅ Batching automático (10s ou 1024 logs)
- ✅ Resource attributes automáticos
- ✅ Labels para organização
- ✅ Graceful shutdown
- ✅ Integração completa com Grafana
- ✅ Correlação Logs → Traces → Metrics

### 🔧 Configuração Final

#### Variáveis de Ambiente
```bash
OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4318"
```

#### Inicialização do Logger
```go
if err := logger.InitLogger("tinyurl-api", "development"); err != nil {
    panic("failed to initialize logger: " + err.Error())
}
defer logger.Sync()
defer logger.Shutdown(ctx)
```

#### Uso
```go
logger.Logger.Info("message", zap.String("key", "value"))
logger.Logger.Error("error", zap.Error(err))
```

### 📊 Datasources no Grafana

| Nome | Tipo | URL | Padrão | Integração |
|------|------|-----|--------|------------|
| Loki | loki | http://loki:3100 | ✅ | → Tempo (traces) |
| Prometheus | prometheus | http://prometheus:9090 | ❌ | ← Tempo (metrics) |
| Tempo | tempo | http://tempo:3200 | ❌ | ↔ Loki + Prometheus |

### 🧪 Como Testar

#### 1. Verificar Serviços
```bash
docker ps | grep -E "(loki|otel|grafana)"
curl http://localhost:3100/ready
```

#### 2. Iniciar Aplicação
```bash
cd go-server
go run main.go
```

#### 3. Gerar Logs
```bash
curl http://localhost:8080/api/health
```

#### 4. Verificar no Grafana
- URL: http://localhost:3000
- Login: admin/admin
- Explore → Loki → `{service_name="tinyurl-api"}`

### 📈 Resultados Esperados

#### No Console da Aplicação
```
INFO    postgres connection established
INFO    redis connection established
INFO    starting server on :8080
```

#### No Loki (via Grafana)
- Logs aparecem com campos estruturados
- Filtráveis por `service_name`, `level`, etc.
- Parseáveis como JSON
- Timeline com atividade

#### No OTel Collector
```bash
docker logs otel-collector --tail 20
# Deve mostrar: "Everything is ready. Begin running and processing data."
```

### 🐛 Problemas Resolvidos

1. ❌ **Endpoint OTLP com protocolo incorreto**
   - ✅ Removido `http://` do endpoint

2. ❌ **Grafana sem datasource do Loki**
   - ✅ Criado arquivo de configuração completo

3. ❌ **Loki sem otimizações de performance**
   - ✅ Adicionadas configurações de limits e retention

4. ❌ **OTel Collector sem labels**
   - ✅ Adicionados labels padrão

5. ❌ **Falta de documentação**
   - ✅ Criados 3 documentos completos + scripts

### 🔒 Notas de Segurança

⚠️ **Configuração atual é para DESENVOLVIMENTO**

Para produção, alterar:
- [ ] `insecure: true` → Configurar TLS
- [ ] Senhas padrão do Grafana
- [ ] Autenticação no Loki
- [ ] Network policies
- [ ] Limites de rate mais restritivos

### 📚 Referências

- [OpenTelemetry Logs Specification](https://opentelemetry.io/docs/specs/otel/logs/)
- [Loki Configuration](https://grafana.com/docs/loki/latest/configure/)
- [OTel Collector Loki Exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/lokiexporter)
- [Zap Logger](https://github.com/uber-go/zap)
- [OTel Zap Bridge](https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelzap)

---

## Próximos Passos Sugeridos

1. **Adicionar Traces** - Instrumentar aplicação com OpenTelemetry traces
2. **Dashboards** - Criar dashboards no Grafana para visualizar logs
3. **Alertas** - Configurar alertas para erros críticos
4. **Log Sampling** - Implementar sampling para reduzir volume em produção
5. **Structured Logging** - Padronizar campos em todos os logs
6. **Correlation IDs** - Adicionar correlation IDs para rastrear requests

---

**Configuração:** ✅ Completa e Funcional
**Status:** 🟢 Pronto para uso
**Data:** 2025-12-05

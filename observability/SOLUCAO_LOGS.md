# 🔧 Solução: Logs do Go para Loki

## ❌ Problema Encontrado

O OpenTelemetry Collector **não tem o exporter do Loki** disponível nativamente na imagem `otel/opentelemetry-collector-contrib:latest`.

Erro:
```
'exporters' unknown type: "loki" for id: "loki"
```

## 🎯 Soluções Possíveis

### Opção 1: Go → OTel Collector (file) → Promtail → Loki
**Complexidade:** Alta
**Vantagem:** Padronizado
**Desvantagem:** Muitos componentes

### Opção 2: Go → Loki (HTTP direto) ⭐ RECOMENDADA
**Complexidade:** Baixa
**Vantagem:** Simples, direto, sem overhead
**Desvantagem:** Não passa pelo OTel Collector

### Opção 3: Atualizar Loki para 3.0+ (suporta OTLP)
**Complexidade:** Alta
**Vantagem:** Suporte nativo a OTLP
**Desvantagem:** Configuração completamente diferente

## ✅ Implementação Recomendada: Opção 2

Vou implementar um logger que envia logs diretamente do Go para o Loki via HTTP.

### Arquitetura Final

```
┌──────────────┐
│   Go Server  │
└──────┬───────┘
       │
       ├─────> Console (stdout)
       │
       └─────> Loki HTTP API
               (http://loki:3100/loki/api/v1/push)
                     │
              ┌──────▼───────┐
              │     Loki     │
              └──────┬───────┘
                     │
              ┌──────▼───────┐
              │   Grafana    │
              └──────────────┘
```

### Benefícios

✅ **Simples:** Sem componentes intermediários
✅ **Rápido:** Latência mínima
✅ **Confiável:** Menos pontos de falha
✅ **Eficiente:** Sem overhead do OTel Collector para logs
✅ **Mantém OTel:** Ainda podemos usar OTel para traces e métricas

### Implementação

#### 1. Instalar biblioteca Loki para Go

```bash
cd go-server
go get github.com/grafana/loki-client-go/loki
```

#### 2. Criar logger híbrido

O logger vai:
- Enviar logs para Loki via HTTP
- Manter output no console para desenvolvimento
- Usar Zap para estruturação

#### 3. Configuração

```go
// Logger configuration
type LoggerConfig struct {
    LokiURL     string // http://localhost:3100
    ServiceName string
    Environment string
}
```

### Alternativa Mais Simples (Atual)

Se você preferir manter a abordagem atual com OpenTelemetry, pode:

1. **Manter OTel apenas para Traces**
2. **Usar log padrão do Zap direto para console**
3. **Adicionar Promtail para scrape dos logs do container Docker**

```yaml
# docker-compose.yml
promtail:
  image: grafana/promtail:2.9.4
  volumes:
    - /var/lib/docker/containers:/var/lib/docker/containers:ro
    - ./promtail.yaml:/etc/promtail/promtail.yaml
```

## 🚀 Próximos Passos

Qual abordagem você prefere?

### A) Simples e Direta (Recomendada)
- ✅ Go envia logs direto para Loki via HTTP
- ✅ Sem OTel Collector para logs
- ✅ OTel Collector apenas para traces/metrics

### B) Completa com Promtail
- ✅ Go escreve logs no console/arquivo
- ✅ Promtail coleta e envia para Loki
- ✅ Mais padronizado

### C) Apenas Console por enquanto
- ✅ Manter logs no console
- ✅ Configurar Loki/Grafana depois
- ✅ Focar em funcionalidades primeiro

---

**Recomendação:** Opção A para desenvolvimento, Opção B para produção.

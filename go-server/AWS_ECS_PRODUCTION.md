# Configuração de Produção na AWS ECS Fargate

Este documento descreve como configurar o backend Go para produção na AWS ECS Fargate com observabilidade completa.

## Variáveis de Ambiente no ECS Task Definition

```json
{
  "containerDefinitions": [
    {
      "name": "go-backend",
      "image": "your-ecr-repo/go-backend:latest",
      "environment": [
        {
          "name": "ENV",
          "value": "production"
        },
        {
          "name": "SERVICE_NAME",
          "value": "go-backend"
        },
        {
          "name": "OTEL_EXPORTER_OTLP_ENDPOINT",
          "value": "http://observability:4318"
        },
        {
          "name": "OTEL_EXPORTER_OTLP_PROTOCOL",
          "value": "http/protobuf"
        },
        {
          "name": "LOKI_ENDPOINT",
          "value": "http://observability/observability/loki/api/v1/push"
        },
        {
          "name": "AWS_ECS_CLUSTER_NAME",
          "value": "tinyurl-prod"
        }
      ],
      "secrets": [
        {
          "name": "POSTGRES_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:..."
        }
      ]
    }
  ]
}
```

## Service Discovery

O backend Go usa **AWS Cloud Map (Service Discovery)** para descobrir o serviço de observabilidade.

### Configuração do Service Discovery

1. Crie um namespace privado no Cloud Map:
```bash
aws servicediscovery create-private-dns-namespace \
  --name internal.tinyurl.local \
  --vpc vpc-xxxxx
```

2. Registre o serviço de observabilidade:
```bash
aws servicediscovery create-service \
  --name observability \
  --dns-config 'NamespaceId="ns-xxxxx",DnsRecords=[{Type="A",TTL="60"}]' \
  --health-check-custom-config FailureThreshold=1
```

3. Na Task Definition do ECS, configure o Service Discovery:
```json
{
  "serviceRegistries": [
    {
      "registryArn": "arn:aws:servicediscovery:...",
      "containerName": "observability-container"
    }
  ]
}
```

## Arquitetura de Rede

```
┌─────────────────────────────────────────────────────────┐
│                     AWS VPC                              │
│                                                          │
│  ┌──────────────┐         ┌──────────────────────────┐ │
│  │   ALB        │────────▶│   ECS Service            │ │
│  │ Public       │         │   (go-backend)           │ │
│  │ Subnets      │         │   Private Subnets        │ │
│  └──────────────┘         └──────────────────────────┘ │
│                                     │                    │
│                                     │ Service Discovery  │
│                                     │ (observability)    │
│                                     ▼                    │
│                           ┌──────────────────────────┐  │
│                           │   ECS Service            │  │
│                           │   (observability)        │  │
│                           │   - OTEL Collector       │  │
│                           │   - Loki                 │  │
│                           │   Private Subnets        │  │
│                           └──────────────────────────┘  │
│                                     │                    │
│                                     ▼                    │
│                           ┌──────────────────────────┐  │
│                           │   Backend Services       │  │
│                           │   - Tempo (EFS)          │  │
│                           │   - Prometheus (EFS)     │  │
│                           └──────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Fluxo de Observabilidade

### Traces
```
go-backend → OTEL Collector (observability:4318) → Tempo → Grafana
```

### Logs
```
go-backend → Loki (observability/observability/loki/api/v1/push) → Grafana
```

### Métricas
```
go-backend → OTEL Collector (observability:4318) → Prometheus → Grafana
```

## Health Checks

### Application Load Balancer (ALB)

Configure o health check do ALB para usar o endpoint `/healthz`:

```json
{
  "HealthCheckProtocol": "HTTP",
  "HealthCheckPath": "/healthz",
  "HealthCheckIntervalSeconds": 30,
  "HealthCheckTimeoutSeconds": 5,
  "HealthyThresholdCount": 2,
  "UnhealthyThresholdCount": 3,
  "Matcher": {
    "HttpCode": "200"
  }
}
```

### Resposta esperada do /healthz

```json
{
  "status": "ok",
  "otel_tracing": true,
  "otel_tracing_works": true,
  "otel_metrics": true,
  "otel_metrics_works": true,
  "loki_logging": true,
  "service_name": "go-backend",
  "environment": "production",
  "timestamp": 1234567890
}
```

## Shutdown Gracioso

O backend Go implementa shutdown gracioso para garantir que:

1. HTTP server para de aceitar novas conexões
2. Aguarda requisições em andamento (até 30 segundos)
3. Flush de spans pendentes (OTEL)
4. Flush de métricas pendentes (OTEL)
5. Flush de logs pendentes (Loki)

### Configuração no ECS

Configure o `stopTimeout` na Task Definition:

```json
{
  "stopTimeout": 60
}
```

Isso dá tempo suficiente para o shutdown gracioso.

## Batch Processing

### Traces
- Batch size: 512 spans
- Queue size: 2048 spans
- Timeout: 5 segundos

### Métricas
- Export interval: 30 segundos

### Logs
- Retry: 3 tentativas com backoff exponencial
- Base delay: 100ms

## Monitoramento

### CloudWatch Container Insights

Ative o Container Insights para métricas de infraestrutura:

```json
{
  "containerInsights": "enabled"
}
```

### Alarmes recomendados

1. **CPU > 80%** por 5 minutos
2. **Memory > 90%** por 5 minutos
3. **Unhealthy target count > 0** por 2 minutos
4. **5xx errors > 10** por 5 minutos

## CORS

Em produção, apenas `https://fonsecaaso.com` é permitido:

```go
AllowOrigins: []string{"https://fonsecaaso.com"}
```

## Endpoint /metrics

O endpoint `/metrics` do Prometheus **NÃO é exposto** em produção.

Apenas em `ENV=local` ele fica disponível.

## Logs Estruturados

Todos os logs incluem automaticamente:

- `trace_id`: ID do trace OpenTelemetry
- `span_id`: ID do span atual
- `request_id`: ID único da requisição
- `method`: Método HTTP
- `path`: Path da requisição
- `status`: Status code da resposta
- `latency_ms`: Latência em milissegundos

Exemplo:
```json
{
  "level": "info",
  "ts": "2024-01-15T10:30:45.123Z",
  "msg": "Request completed",
  "trace_id": "a1b2c3d4e5f6g7h8",
  "span_id": "1234567890abcdef",
  "request_id": "req-xyz123",
  "method": "POST",
  "path": "/api/",
  "status": 201,
  "latency_ms": 45.2,
  "service_name": "go-backend",
  "environment": "production"
}
```

## Troubleshooting

### Backend não consegue enviar traces

1. Verifique se o service discovery está funcionando:
```bash
aws servicediscovery discover-instances \
  --namespace-name internal.tinyurl.local \
  --service-name observability
```

2. Verifique os logs do container:
```bash
aws logs tail /ecs/go-backend --follow
```

### Logs não aparecem no Loki

1. Verifique se o Loki está acessível:
```bash
curl http://observability:3100/ready
```

2. Verifique se o backend está enviando logs:
```bash
# Olhe nos logs do backend por erros "failed to send log to loki"
```

### Métricas não aparecem no Prometheus

1. Verifique se o OTEL Collector está recebendo métricas:
```bash
curl http://observability:8888/metrics | grep otelcol_receiver
```

## Variáveis de Ambiente Completas

```bash
# Application
ENV=production
SERVICE_NAME=go-backend
GO_ENV=production

# Database
POSTGRES_HOST=<rds-endpoint>
POSTGRES_PORT=5432
POSTGRES_DB=postgres
POSTGRES_USER=postgres
POSTGRES_PASSWORD=<from-secrets-manager>

# Redis
REDIS_ADDR=<elasticache-endpoint>:6379

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=http://observability:4318
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
LOKI_ENDPOINT=http://observability/observability/loki/api/v1/push

# AWS
AWS_ECS_CLUSTER_NAME=tinyurl-prod
# TASK_ARN is automatically set by ECS
```

## Checklist de Deploy

- [ ] Variáveis de ambiente configuradas
- [ ] Service discovery configurado
- [ ] Health check do ALB apontando para /healthz
- [ ] stopTimeout configurado para 60 segundos
- [ ] Container Insights ativado
- [ ] Alarmes do CloudWatch configurados
- [ ] CORS configurado para domínio de produção
- [ ] Secrets Manager para senhas sensíveis
- [ ] Tags para recursos (Environment, Service, etc.)

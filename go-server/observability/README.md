# Local Observability Stack

Este diretório contém a configuração completa para rodar a stack de observabilidade localmente usando Docker Compose.

## Componentes

- **OTEL Collector**: Coleta traces e métricas do backend Go
- **Grafana Tempo**: Armazena e consulta traces distribuídos
- **Grafana Loki**: Armazena e consulta logs estruturados
- **Prometheus**: Armazena e consulta métricas
- **Grafana**: Visualização unificada de traces, logs e métricas

## Como usar

### 1. Iniciar a stack de observabilidade

```bash
docker-compose -f docker-compose.observability.yaml up -d
```

### 2. Configurar o backend Go para ambiente local

Edite o arquivo `.env`:

```bash
ENV=local
SERVICE_NAME=go-backend

# OTEL Collector local
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf

# Loki local
LOKI_ENDPOINT=http://localhost:3100/loki/api/v1/push
```

### 3. Rodar o backend Go

```bash
go run main.go
```

### 4. Acessar os dashboards

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Tempo**: http://localhost:3200
- **Loki**: http://localhost:3100

## Verificar a integração

### Traces

1. Acesse o Grafana
2. Vá para **Explore** → selecione **Tempo**
3. Busque por `service.name="go-backend"`
4. Você verá todos os traces das requisições HTTP

### Logs

1. No Grafana, vá para **Explore** → selecione **Loki**
2. Use a query: `{service_name="go-backend"}`
3. Você verá todos os logs estruturados
4. Clique em um log com `trace_id` e veja o botão para ir ao trace correspondente

### Métricas

1. No Grafana, vá para **Explore** → selecione **Prometheus**
2. Use a query: `tinyurl_http_requests_total`
3. Você verá as métricas HTTP coletadas

## Correlação Trace → Log

Quando você visualiza um trace no Tempo, pode clicar em "Logs for this span" e o Grafana automaticamente busca os logs com o mesmo `trace_id`.

## Parar a stack

```bash
docker-compose -f docker-compose.observability.yaml down
```

## Limpar dados

```bash
docker-compose -f docker-compose.observability.yaml down -v
```

## Configuração para Produção (ECS)

Para produção na AWS ECS Fargate, use estas variáveis de ambiente:

```bash
ENV=production
SERVICE_NAME=go-backend

# OTEL Collector via service discovery
OTEL_EXPORTER_OTLP_ENDPOINT=http://observability:4318
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf

# Loki via service discovery
LOKI_ENDPOINT=http://observability/observability/loki/api/v1/push

# AWS ECS attributes
AWS_ECS_CLUSTER_NAME=tinyurl-prod
TASK_ARN=<automatically-set-by-ecs>
```

## Troubleshooting

### Logs não aparecem no Loki

Verifique se o backend consegue enviar para Loki:

```bash
curl -X POST http://localhost:3100/ready
```

### Traces não aparecem no Tempo

Verifique se o OTEL Collector está recebendo:

```bash
curl http://localhost:8888/metrics | grep otelcol_receiver
```

### Métricas não aparecem no Prometheus

Verifique os targets do Prometheus:

```bash
curl http://localhost:9090/api/v1/targets
```

# 📊 Observability - TinyURL

Complete observability system with metrics, logs, and tracing.

## 🚀 Quick Start

### 1. Start Observability Stack

```bash
cd observability
docker-compose up -d
```

### 2. Start Application

```bash
cd go-server
go run main.go
```

### 3. Access Services

| Service | URL | Credentials |
|---------|-----|-------------|
| **Grafana** | http://localhost:3000 | admin / admin |
| **Prometheus** | http://localhost:9090 | - |
| **Loki** | http://localhost:3100 | - |
| **Tempo** | http://localhost:3200 | - |
| **App Metrics** | http://localhost:8080/metrics | - |
| **Health Check** | http://localhost:8080/api/health | - |

## 📈 Available Metrics

### Performance
- Request rate (req/s) total and per endpoint
- Latency (P50, P95, P99) in ms
- Throughput (MB/s) input and output
- Error rate (%) for 4xx and 5xx errors
- Top 10 slowest endpoints

### System
- Memory usage (heap, stack, system) in MB
- Active goroutines
- CPU usage

### Application
- URLs created and accessed
- Cache hit rate
- Requests in progress

## 📝 Logs

### Architecture

```
Go Server → Loki (direct HTTP) → Grafana
     ↓
  Console (development)
```

The application sends logs **directly** to Loki via HTTP (without OpenTelemetry Collector), maintaining simplicity and low latency.

### Useful Queries (LogQL)

```logql
# All logs
{service_name="tinyurl-api"}

# Filter by endpoint
{service_name="tinyurl-api", path="/api/health"}

# Filter by HTTP method
{service_name="tinyurl-api", method="POST"}

# Only errors
{service_name="tinyurl-api", level="error"}

# HTTP errors (4xx and 5xx)
{service_name="tinyurl-api", status=~"[45].*"}

# Search for specific text
{service_name="tinyurl-api"} |= "database"

# Top 5 most accessed endpoints
topk(5, sum by(path) (count_over_time({service_name="tinyurl-api"}[1h])))
```

### Indexed Labels

For fast queries, these labels are indexed:
- `service_name`: "tinyurl-api"
- `environment`: "development" or "production"
- `level`: debug, info, warn, error
- `path`: HTTP endpoint
- `method`: GET, POST, PUT, DELETE
- `status`: HTTP status code

## 🧪 Testing

Generate traffic to see metrics and logs:

```bash
# Create URLs
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/ \
    -H "Content-Type: application/json" \
    -d "{\"url\":\"https://example.com/$i\"}"
done

# Health checks
for i in {1..50}; do
  curl http://localhost:8080/api/health
  sleep 0.1
done
```

## 🔍 Troubleshooting

### Grafana shows no data

```bash
# 1. Check if app is exposing metrics
curl http://localhost:8080/metrics

# 2. Check if Prometheus is collecting
open http://localhost:9090/targets
# Status should be "UP"

# 3. Restart Grafana
cd observability && docker-compose restart grafana
```

### Logs don't appear in Loki

```bash
# 1. Check if Loki is ready
curl http://localhost:3100/ready
# Should return: ready

# 2. Wait a few seconds (batching)
sleep 3

# 3. View Loki logs
docker logs loki --tail 50
```

### Prometheus doesn't collect

```bash
# Check config
cat observability/prometheus/prometheus.yml

# Restart Prometheus
cd observability && docker-compose restart prometheus
```

## 🐳 Docker

### For local development
```bash
LOKI_URL="http://localhost:3100/loki/api/v1/push"
```

### For Docker (same network)
```bash
# Connect to observability network
docker run -p 8080:8080 \
  --network observability_default \
  -e LOKI_URL="http://loki:3100/loki/api/v1/push" \
  tiny-url
```

### Using host.docker.internal
```bash
docker run -p 8080:8080 \
  -e LOKI_URL="http://host.docker.internal:3100/loki/api/v1/push" \
  tiny-url
```

## 🛠️ Useful Commands

```bash
# View raw metrics
curl http://localhost:8080/metrics

# View Prometheus targets
open http://localhost:9090/targets

# Grafana logs
docker logs grafana -f

# Restart everything
cd observability && docker-compose restart

# Stop everything
docker-compose down

# Clean volumes
docker-compose down -v
```

## 📁 Structure

```
TinyUrl/
├── go-server/
│   └── internal/
│       ├── metrics/          # Metrics definitions
│       ├── middleware/       # Collection middleware
│       └── logger/           # Logger with Loki integration
│
└── observability/
    ├── docker-compose.yml    # Observability stack
    ├── prometheus/
    │   └── prometheus.yml    # Prometheus config
    ├── loki/
    │   └── loki.yaml         # Loki config
    └── grafana/
        └── provisioning/
            ├── datasources/  # Auto-configured datasources
            └── dashboards/   # Pre-configured dashboard
```

## 🎯 Useful Prometheus Queries

```promql
# Request rate
sum(rate(http_requests_total[1m]))

# P95 latency
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Memory in MB
memory_usage_bytes{type="heap_alloc"} / 1024 / 1024

# Error rate
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100

# Top 10 endpoints by latency
topk(10, histogram_quantile(0.95, sum by (method, path, le) (rate(http_request_duration_seconds_bucket[5m]))))
```

## 💡 Architecture Decision

### Why direct logs to Loki?

**Advantages:**
- ✅ Lower latency (~10-50ms vs ~100-500ms with OTel)
- ✅ Fewer components (Go → Loki vs Go → OTel → Loki)
- ✅ Simpler to debug
- ✅ Less CPU/memory overhead
- ✅ Fewer failure points

**OpenTelemetry Collector is only used for:**
- Traces (OpenTelemetry standard)
- Metrics that need processing

## 🔐 Production

For production, adjust:

1. **Loki URL** in `.env`:
```bash
LOKI_URL="http://loki:3100/loki/api/v1/push"
```

2. **Grafana credentials**: Change default password

3. **Log batch size**: Increase to 1MB

4. **Log level**: Change from DEBUG to INFO

5. **TLS**: Configure certificates

6. **Authentication**: Add authentication to Loki

7. **Network policies**: Isolate services

## ✅ Verification Checklist

- [ ] All Docker containers running
- [ ] Loki responds "ready" at http://localhost:3100/ready
- [ ] Prometheus targets "UP" at http://localhost:9090/targets
- [ ] Grafana accessible at http://localhost:3000
- [ ] Dashboard "TinyURL API Metrics" visible in Grafana
- [ ] Loki datasource configured and default
- [ ] Go application running without errors
- [ ] Logs appear in console and Grafana
- [ ] Metrics visible in dashboard

---

**Complete documentation:** See individual markdown files in `observability/` for specific details
**Dashboard:** Access Grafana → Dashboards → TinyURL API Metrics
**Status:** ✅ System ready to use

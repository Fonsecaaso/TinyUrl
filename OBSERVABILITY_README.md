# 📊 Observability - TinyURL

Complete metrics and monitoring system for the TinyURL application.

## 🎯 Quick Access

| Service | URL | Credentials |
|---------|-----|-------------|
| **Grafana** (Dashboards) | http://localhost:3000 | admin / admin |
| **Prometheus** (Metrics) | http://localhost:9090 | - |
| **App Metrics** | http://localhost:8080/metrics | - |
| **Health Check** | http://localhost:8080/api/health | - |

## 🚀 How to Use

### 1. Start Observability Services

```bash
cd observability
docker-compose up -d
```

### 2. Start the Application

```bash
cd go-server
go run main.go
```

### 3. Access Grafana

1. Open: http://localhost:3000
2. Login: `admin` / `admin`
3. Go to **Dashboards** → **TinyURL API Metrics**

**Done!** You'll see all metrics in real-time.

## 📈 Available Metrics

### Performance
- ✅ **Requests/second** - Total req/s rate and per endpoint
- ✅ **Latency** - P50, P95, P99 in milliseconds
- ✅ **Throughput** - MB/s input and output
- ✅ **Error Rate** - % of 4xx and 5xx errors

### System
- ✅ **Memory Usage** - Heap, Stack, System (in MB)
- ✅ **Goroutines** - Number of active goroutines
- ✅ **CPU** - CPU usage

### Application
- ✅ **URLs Created** - Total shortened URLs
- ✅ **URLs Accessed** - Total redirects
- ✅ **Cache Hit Rate** - Cache efficiency
- ✅ **Requests In Progress** - Concurrent requests

## 📚 Documentation

- **[QUICKSTART_GRAFANA.md](observability/QUICKSTART_GRAFANA.md)** - 2-minute guide
- **[GRAFANA_SETUP.md](observability/GRAFANA_SETUP.md)** - Complete Grafana guide
- **[METRICS.md](go-server/METRICS.md)** - Detailed documentation of all metrics
- **[QUICKSTART_METRICS.md](go-server/QUICKSTART_METRICS.md)** - Quick metrics guide

## 🧪 Test

Generate traffic to see metrics in action:

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

After a few seconds, the Grafana dashboard will show the data!

## 🏗️ Architecture

```
┌─────────────────┐
│   Go Server     │
│   :8080         │ ← Exposes /metrics
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│   Prometheus    │ ← Collects metrics every 5s
│   :9090         │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│    Grafana      │ ← Visualizes dashboards
│    :3000        │
└─────────────────┘
```

## 📁 File Structure

```
TinyUrl/
├── go-server/
│   ├── internal/
│   │   ├── metrics/
│   │   │   └── metrics.go          # Metric definitions
│   │   └── middleware/
│   │       └── metrics.go          # Collection middleware
│   ├── METRICS.md                  # Metrics docs
│   └── QUICKSTART_METRICS.md       # Quick guide
│
└── observability/
    ├── docker-compose.yml          # Observability stack
    ├── prometheus/
    │   └── prometheus.yml          # Prometheus config
    ├── grafana/
    │   └── provisioning/
    │       ├── datasources/
    │       │   └── prometheus.yml  # Automatic datasource
    │       └── dashboards/
    │           ├── dashboard.yml   # Provisioning config
    │           └── tinyurl-metrics.json  # Dashboard
    ├── GRAFANA_SETUP.md           # Complete setup
    └── QUICKSTART_GRAFANA.md      # Quick start
```

## 🛠️ Useful Commands

```bash
# View raw metrics
curl http://localhost:8080/metrics

# View Prometheus targets
open http://localhost:9090/targets

# Grafana logs
docker logs grafana -f

# Restart Grafana
cd observability && docker-compose restart grafana

# Stop everything
docker-compose down

# Stop and clean volumes
docker-compose down -v
```

## 🎨 Grafana Dashboard

The dashboard includes:
- **2 performance charts** (req/s, latency)
- **2 throughput/error charts**
- **2 resource charts** (memory, goroutines)
- **4 gauges/stats** (in-flight, cache, URLs)

Automatic refresh every **5 seconds**.

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

### Dashboard doesn't appear
```bash
# Check if files were mounted
docker exec grafana ls /etc/grafana/provisioning/dashboards

# Should show: dashboard.yml and tinyurl-metrics.json
```

### Prometheus doesn't collect
```bash
# Check config
cat observability/prometheus/prometheus.yml

# Restart Prometheus
cd observability && docker-compose restart prometheus
```

## 💡 Tips

1. **Bookmark the dashboard** - Click the ⭐ for quick access
2. **Adjust time range** - Use the time selector in the top corner
3. **Zoom in charts** - Click and drag to zoom
4. **Explore queries** - Click "Edit" on any panel
5. **Create alerts** - Use Prometheus Alertmanager

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
```

## 🚀 Next Steps

1. ✅ Basic metrics implemented
2. ✅ Grafana dashboard configured
3. ✅ Automatic provisioning
4. 📋 Configure alerts in Prometheus
5. 📋 Add business metrics (conversion, etc)
6. 📋 Integrate with Loki for logs
7. 📋 Add tracing with Tempo

---

**Complete documentation**: See the markdown files in the `observability/` and `go-server/` folders

# 📊 How to Visualize Metrics in Grafana

## 🚀 Quick Start

### 1. Start the Observability Stack

```bash
cd observability
docker-compose up -d
```

This will start:
- **Grafana** on port `3000`
- **Prometheus** on port `9090`
- **Loki** on port `3100`
- **Tempo** on port `3200`

### 2. Start the Go Application

```bash
cd ../go-server
go run main.go
```

The application will be exposing metrics at: `http://localhost:8080/metrics`

### 3. Access Grafana

Open your browser and go to:

```
http://localhost:3000
```

**Default credentials:**
- **Username**: `admin`
- **Password**: `admin`

### 4. View the Dashboard

The **"TinyURL API Metrics"** dashboard is pre-configured and should appear automatically!

To access it:
1. Click **"Dashboards"** in the sidebar (icon with 4 squares)
2. Search for **"TinyURL API Metrics"**
3. Click to open

## 📈 What You'll See on the Dashboard

### Performance Panels
1. **Request Rate (req/s)** - Request rate per second
2. **Request Latency (P50, P95, P99)** - Request latency
3. **Throughput (MB/s)** - Data transfer volume
4. **Error Rate (%)** - 4xx and 5xx error rate

### System Resource Panels
5. **Memory Usage (MB)** - Memory usage (heap, system)
6. **Goroutines** - Number of active goroutines

### Status Panels
7. **Requests In Flight** - Requests being processed now
8. **Cache Hit Rate (%)** - Cache hit rate
9. **Total URLs Created** - Total URLs created
10. **Total URL Accesses** - Total URL accesses

## 🧪 Test the Metrics

To generate traffic and see the charts in action:

```bash
# Create some URLs
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/ \
    -H "Content-Type: application/json" \
    -d "{\"url\":\"https://example.com/$i\"}"
done

# Make some requests
for i in {1..50}; do
  curl http://localhost:8080/api/health
done
```

## 🔧 Automatic Configuration

Grafana is configured with **automatic provisioning**:

- **Datasource**: Prometheus already configured and connected
- **Dashboard**: TinyURL API Metrics already imported
- **Refresh**: Automatic refresh every 5 seconds

Configuration files:
```
observability/grafana/provisioning/
├── datasources/
│   └── prometheus.yml          # Prometheus connection
└── dashboards/
    ├── dashboard.yml           # Provisioning configuration
    └── tinyurl-metrics.json    # Pre-configured dashboard
```

## 🎯 Useful Queries in Grafana

If you want to create your own panels, here are some useful queries:

### Performance
```promql
# Request rate
sum(rate(http_requests_total[1m]))

# P95 latency per endpoint
histogram_quantile(0.95, sum by (path, le) (rate(http_request_duration_seconds_bucket[5m])))

# Throughput
sum(rate(http_response_size_bytes_sum[1m])) / 1024 / 1024
```

### Resources
```promql
# Heap memory in MB
memory_usage_bytes{type="heap_alloc"} / 1024 / 1024

# Goroutines
go_goroutines_count

# CPU (standard Go metrics)
rate(process_cpu_seconds_total[1m]) * 100
```

### Application
```promql
# Error rate
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100

# Cache hit rate
sum(rate(cache_hits_total[5m])) / (sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m]))) * 100

# URLs created per minute
rate(url_creation_total[1m]) * 60
```

## 🔍 Check if It's Working

### 1. Check if Prometheus is collecting metrics

```bash
# View targets in Prometheus
open http://localhost:9090/targets

# Should show "go-backend" with UP status
```

### 2. Check if metrics are being exposed

```bash
# View raw metrics
curl http://localhost:8080/metrics | grep http_requests_total
```

### 3. Check Grafana logs

```bash
docker logs grafana
```

## 🛠️ Troubleshooting

### Dashboard doesn't appear
```bash
# Restart Grafana
docker-compose restart grafana

# Check logs
docker logs grafana -f
```

### Datasource doesn't connect to Prometheus
1. Check if Prometheus is running: `http://localhost:9090`
2. In Grafana, go to **Configuration > Data Sources > Prometheus**
3. Click **"Test"** to verify the connection

### Charts without data
1. Make sure the Go application is running
2. Make some requests to generate metrics
3. Check Prometheus: `http://localhost:9090/graph`
4. Execute a query: `http_requests_total`

### Stop all services
```bash
cd observability
docker-compose down
```

### Clean everything and start from scratch
```bash
docker-compose down -v  # Also removes volumes
docker-compose up -d
```

## 📚 Additional Resources

- **Prometheus UI**: http://localhost:9090
- **Grafana**: http://localhost:3000
- **App Metrics**: http://localhost:8080/metrics
- **Health Check**: http://localhost:8080/api/health

## 💡 Tips

1. **Bookmark the Dashboard**: Click the star ⭐ for easy access
2. **Adjust Time Range**: Use the selector in the top right corner
3. **Auto refresh**: Already configured for 5 seconds
4. **Explore queries**: Click on any panel and then "Edit" to see the query
5. **Duplicate panels**: Useful for experimenting without losing the original

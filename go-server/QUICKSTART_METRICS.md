# Quick Start: TinyURL Metrics

## 🚀 How to Use

### 1. Start the Application

```bash
cd go-server
go run main.go
```

### 2. Access Metrics

The `/metrics` endpoint is available at:
```
http://localhost:8080/metrics
```

### 3. Test with Requests

Make some requests to generate metrics:

```bash
# Create a URL
curl -X POST http://localhost:8080/api/ \
  -H "Content-Type: application/json" \
  -d '{"url":"https://google.com"}'

# Access the created URL
curl http://localhost:8080/api/abc123

# Health check
curl http://localhost:8080/api/health
```

### 4. Visualize Metrics in Prometheus

1. Access: `http://localhost:9090`
2. Go to **Graph**
3. Try these queries:

```promql
# Request rate per second
rate(http_requests_total[1m])

# P95 latency
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Memory usage (MB)
memory_usage_bytes{type="heap_alloc"} / 1024 / 1024

# Number of goroutines
go_goroutines_count
```

## 📊 Key Metrics

### Performance
- **req/sec**: `rate(http_requests_total[1m])`
- **Latency**: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
- **Throughput**: `rate(http_response_size_bytes_sum[1m])`

### Resources
- **CPU**: `cpu_usage_percent`
- **Memory**: `memory_usage_bytes{type="heap_alloc"}`
- **Goroutines**: `go_goroutines_count`

### Application
- **Error Rate**: `rate(http_requests_total{status=~"5.."}[5m])`
- **URLs Created**: `url_creation_total`
- **Cache Hit Rate**: `cache_hits_total / (cache_hits_total + cache_misses_total)`

## 🎯 Real-time Metrics

To view metrics live, use the `watch` command:

```bash
# View metrics every 2 seconds
watch -n 2 'curl -s http://localhost:8080/metrics | grep -E "(http_requests_total|http_request_duration|memory_usage)"'
```

## 📈 Grafana Dashboard

A pre-configured Grafana dashboard is available at:
```
observability/grafana/dashboards/tinyurl-metrics.json
```

Import this dashboard into Grafana for complete visualization.

## 🔍 Debug

If metrics are not appearing:

1. Check if the application is running: `curl http://localhost:8080/metrics`
2. Check if Prometheus is collecting: `http://localhost:9090/targets`
3. Check application logs
4. Make sure `prometheus.yml` is configured correctly

## 💡 Tips

- System metrics (CPU, memory) are collected every 15 seconds
- Use longer intervals (`[5m]`) for more stable queries
- P95 and P99 are better latency indicators than the average
- Monitor `http_requests_in_flight` to detect bottlenecks

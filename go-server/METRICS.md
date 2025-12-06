# TinyURL Application Metrics

This document describes all metrics exposed by the application at the `/metrics` endpoint.

## Metrics Endpoint

```
GET http://localhost:8080/metrics
```

## Available Metrics

### 1. HTTP Metrics

#### `http_requests_total`
- **Type**: Counter
- **Description**: Total number of HTTP requests
- **Labels**: `method`, `path`, `status`
- **Example Queries**:
  ```promql
  # Request rate per second
  rate(http_requests_total[1m])

  # Requests per endpoint
  sum by (path) (rate(http_requests_total[5m]))

  # 5xx error rate
  rate(http_requests_total{status=~"5.."}[5m])
  ```

#### `http_request_duration_seconds`
- **Type**: Histogram
- **Description**: HTTP request latency in seconds
- **Labels**: `method`, `path`, `status`
- **Buckets**: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
- **Example Queries**:
  ```promql
  # P95 latency PER ENDPOINT (shows each endpoint separately)
  histogram_quantile(0.95, sum by (path, le) (rate(http_request_duration_seconds_bucket[5m])))

  # P95 latency overall (aggregated)
  histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))

  # P99 latency per endpoint
  histogram_quantile(0.99, sum by (path, le) (rate(http_request_duration_seconds_bucket[5m])))

  # Average latency per endpoint
  rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])

  # Top 10 slowest endpoints (P95)
  topk(10, histogram_quantile(0.95, sum by (path, le) (rate(http_request_duration_seconds_bucket[5m]))))

  # Latency per endpoint AND HTTP method
  histogram_quantile(0.95, sum by (method, path, le) (rate(http_request_duration_seconds_bucket[5m])))

  # Latency per endpoint with status code
  histogram_quantile(0.95, sum by (path, status, le) (rate(http_request_duration_seconds_bucket[5m])))
  ```

#### `http_request_size_bytes`
- **Type**: Histogram
- **Description**: HTTP request size in bytes
- **Labels**: `method`, `path`
- **Example Queries**:
  ```promql
  # Average request size
  rate(http_request_size_bytes_sum[5m]) / rate(http_request_size_bytes_count[5m])
  ```

#### `http_response_size_bytes`
- **Type**: Histogram
- **Description**: HTTP response size in bytes
- **Labels**: `method`, `path`, `status`
- **Example Queries**:
  ```promql
  # Average response size
  rate(http_response_size_bytes_sum[5m]) / rate(http_response_size_bytes_count[5m])

  # Throughput (bytes/second)
  rate(http_response_size_bytes_sum[1m])
  ```

#### `http_requests_in_flight`
- **Type**: Gauge
- **Description**: Number of requests currently being processed
- **Example Queries**:
  ```promql
  # Concurrent requests
  http_requests_in_flight
  ```

### 2. Application Metrics

#### `url_creation_total`
- **Type**: Counter
- **Description**: Total URLs created
- **Labels**: `status` (success/error)
- **Where it's generated**: [internal/service/url_service.go:74](../internal/service/url_service.go#L74) (success) and [:68](../internal/service/url_service.go#L68) (error)
- **Example Queries**:
  ```promql
  # Total URLs created successfully
  sum(url_creation_total{status="success"})

  # Creation rate per second
  rate(url_creation_total{status="success"}[1m])
  ```

#### `url_access_total`
- **Type**: Counter
- **Description**: Total URL accesses/redirects
- **Labels**: `status` (success/error/not_found)
- **Where it's generated**: [internal/service/url_service.go:102](../internal/service/url_service.go#L102) (success), [:93](../internal/service/url_service.go#L93) (not_found), [:85](../internal/service/url_service.go#L85) and [:97](../internal/service/url_service.go#L97) (error)
- **Example Queries**:
  ```promql
  # Total successful accesses
  sum(url_access_total{status="success"})

  # Access rate per second
  rate(url_access_total[1m])

  # URL not found rate
  rate(url_access_total{status="not_found"}[5m])
  ```

#### `cache_hits_total`
- **Type**: Counter
- **Description**: Total cache hits
- **Labels**: `cache_type` (redis)
- **Where it's generated**: [internal/repository/url_repository.go:124](../internal/repository/url_repository.go#L124)
- **Example Queries**:
  ```promql
  # Total cache hits
  sum(cache_hits_total)

  # Cache hit rate per second
  rate(cache_hits_total[1m])
  ```

#### `cache_misses_total`
- **Type**: Counter
- **Description**: Total cache misses
- **Labels**: `cache_type` (redis)
- **Where it's generated**: [internal/repository/url_repository.go:129](../internal/repository/url_repository.go#L129)
- **Example Queries**:
  ```promql
  # Cache hit rate percentage
  sum(rate(cache_hits_total[5m])) / (sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m]))) * 100
  ```

### 3. Database Metrics

#### `db_connections_in_use`
- **Type**: Gauge
- **Description**: Number of database connections in use

#### `db_connections_idle`
- **Type**: Gauge
- **Description**: Number of idle connections

#### `db_query_duration_seconds`
- **Type**: Histogram
- **Description**: Query duration in seconds
- **Labels**: `query_type`

### 4. System Metrics

#### `go_goroutines_count`
- **Type**: Gauge
- **Description**: Number of running goroutines
- **Example Queries**:
  ```promql
  # Number of goroutines
  go_goroutines_count
  ```

#### `memory_usage_bytes`
- **Type**: Gauge
- **Description**: Memory usage in bytes
- **Labels**: `type` (alloc, total_alloc, sys, heap_alloc, heap_sys, heap_idle, heap_in_use, stack_in_use)
- **Example Queries**:
  ```promql
  # Heap allocated memory in MB
  memory_usage_bytes{type="heap_alloc"} / 1024 / 1024

  # Total system memory in MB
  memory_usage_bytes{type="sys"} / 1024 / 1024
  ```

#### `cpu_usage_percent`
- **Type**: Gauge
- **Description**: CPU usage percentage

## Standard Prometheus Metrics

In addition to custom metrics, the Prometheus client also exposes standard Go metrics:

- `go_gc_duration_seconds` - Garbage collector duration
- `go_memstats_*` - Detailed memory statistics
- `go_threads` - Number of OS threads
- `process_cpu_seconds_total` - Total CPU time used by the process
- `process_resident_memory_bytes` - Resident memory in bytes
- `process_virtual_memory_bytes` - Virtual memory in bytes

## Useful Prometheus Queries

### Request Rate (Req/Second)
```promql
sum(rate(http_requests_total[1m]))
```

### Request Rate per Endpoint
```promql
sum by (path) (rate(http_requests_total[1m]))
```

### Throughput (MB/second)
```promql
sum(rate(http_response_size_bytes_sum[1m])) / 1024 / 1024
```

### Error Rate
```promql
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100
```

### Latency P50, P95, P99
```promql
# P50 (aggregated)
histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))

# P95 (aggregated)
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# P99 (aggregated)
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))
```

### 🔥 Latency per Endpoint (IMPORTANT!)
```promql
# P95 latency of EACH endpoint (separated by path)
histogram_quantile(0.95, sum by (path, le) (rate(http_request_duration_seconds_bucket[5m])))

# Top 10 slowest endpoints (P95)
topk(10, histogram_quantile(0.95, sum by (path, le) (rate(http_request_duration_seconds_bucket[5m]))))

# Average latency per endpoint
sum by (path) (rate(http_request_duration_seconds_sum[5m])) / sum by (path) (rate(http_request_duration_seconds_count[5m]))

# Slowest endpoints (average)
topk(10, rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m]))

# Latency per endpoint + HTTP method
histogram_quantile(0.95, sum by (method, path, le) (rate(http_request_duration_seconds_bucket[5m])))

# Compare success vs error latency for the same endpoint
histogram_quantile(0.95, sum by (path, status, le) (rate(http_request_duration_seconds_bucket{path="/api/"}[5m])))
```

### Memory Usage (MB)
```promql
memory_usage_bytes{type="heap_alloc"} / 1024 / 1024
```

### Cache Hit Rate
```promql
sum(rate(cache_hits_total[5m])) / (sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m]))) * 100
```

## Visualization in Grafana

To visualize these metrics in Grafana, you can create dashboards with the following panels:

1. **Request Rate**: Graph with `sum(rate(http_requests_total[1m]))`
2. **Latency**: Graph with P50, P95, P99
3. **Error Rate**: Graph with 4xx and 5xx error rates
4. **Throughput**: Graph with bytes/second
5. **Memory**: Graph with heap usage
6. **Goroutines**: Graph with number of goroutines
7. **Requests in Flight**: Gauge with `http_requests_in_flight`

## Testing Metrics

1. Start the application:
   ```bash
   cd go-server
   go run main.go
   ```

2. Make some requests:
   ```bash
   curl -X POST http://localhost:8080/api/ -d '{"url":"https://google.com"}'
   ```

3. Check the metrics:
   ```bash
   curl http://localhost:8080/metrics
   ```

4. Access Prometheus at `http://localhost:9090` and use the queries above.

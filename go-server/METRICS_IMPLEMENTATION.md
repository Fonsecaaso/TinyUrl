# 📍 Where Metrics Are Generated

This document maps exactly where each custom metric is incremented in the code.

## 🎯 URL Metrics

### `url_creation_total`
Incremented when a URL is created (or fails to be created).

**Location**: [internal/service/url_service.go](internal/service/url_service.go)

```go
// Line 74 - Success creating new URL
metrics.URLCreationTotal.WithLabelValues("success").Inc()

// Line 68 - Error creating URL
metrics.URLCreationTotal.WithLabelValues("error").Inc()
```

**When it's triggered**:
- ✅ **success**: When `repo.CreateOrGet()` returns `isNew=true` (URL created successfully)
- ❌ **error**: When `repo.CreateOrGet()` returns error

---

### `url_access_total`
Incremented when a URL is accessed/fetched.

**Location**: [internal/service/url_service.go](internal/service/url_service.go)

```go
// Line 102 - Success fetching URL
metrics.URLAccessTotal.WithLabelValues("success").Inc()

// Line 93 - URL not found
metrics.URLAccessTotal.WithLabelValues("not_found").Inc()

// Line 85 - Invalid short code
metrics.URLAccessTotal.WithLabelValues("error").Inc()

// Line 97 - Database error
metrics.URLAccessTotal.WithLabelValues("error").Inc()
```

**When it's triggered**:
- ✅ **success**: URL found and returned successfully
- 🔍 **not_found**: Valid short code but URL doesn't exist in database
- ❌ **error**: Invalid short code OR database error

---

## 💾 Cache Metrics

### `cache_hits_total`
Incremented when a URL is found in Redis (cache hit).

**Location**: [internal/repository/url_repository.go](internal/repository/url_repository.go)

```go
// Line 124 - Cache hit in Redis
metrics.CacheHitsTotal.WithLabelValues("redis").Inc()
```

**When it's triggered**:
- ✅ URL found in Redis before querying the database
- Avoids PostgreSQL query
- Returns result immediately

---

### `cache_misses_total`
Incremented when a URL is NOT found in Redis (cache miss).

**Location**: [internal/repository/url_repository.go](internal/repository/url_repository.go)

```go
// Line 129 - Cache miss in Redis
metrics.CacheMissesTotal.WithLabelValues("redis").Inc()
```

**When it's triggered**:
- ❌ URL not found in Redis
- Will need to query PostgreSQL
- The URL will be cached after being fetched from database (lines 144-146)

---

## 🌐 HTTP Metrics (Automatic)

These metrics are automatically collected by the middleware.

**Location**: [internal/middleware/metrics.go](internal/middleware/metrics.go)

### Metrics Middleware
```go
// Lines 48-55 - Collects all HTTP metrics
metrics.RecordHTTPMetrics(
    c.Request.Method,
    path,
    status,
    duration,
    requestSize,
    responseSize,
)
```

**Metrics collected automatically**:
- `http_requests_total` - Request counter
- `http_request_duration_seconds` - Latency histogram
- `http_request_size_bytes` - Request size
- `http_response_size_bytes` - Response size
- `http_requests_in_flight` - Concurrent requests gauge

**When they're triggered**:
- For ALL HTTP requests that go through Gin
- Middleware applied at [internal/routes/routes.go:53](internal/routes/routes.go#L53)

---

## 📊 System Metrics (Automatic)

**Location**: [internal/metrics/metrics.go](internal/metrics/metrics.go)

### Periodic Collection
```go
// Lines 145-166 - collectSystemMetrics() function
// Executed every 15 seconds
```

**Metrics collected**:
- `go_goroutines_count` - Number of goroutines
- `memory_usage_bytes{type="alloc"}` - Allocated memory
- `memory_usage_bytes{type="heap_alloc"}` - Heap allocated
- `memory_usage_bytes{type="heap_sys"}` - Heap system
- `memory_usage_bytes{type="sys"}` - Total system
- And other memory metrics

**When they're triggered**:
- Automatically every 15 seconds
- Started at [internal/routes/routes.go:24](internal/routes/routes.go#L24)
- Runs in a separate goroutine

---

## 🧪 How to Test

### Test `url_creation_total`
```bash
# Create a URL (should increment success)
curl -X POST http://localhost:8080/api/ \
  -H "Content-Type: application/json" \
  -d '{"url":"https://google.com"}'

# View metric
curl -s http://localhost:8080/metrics | grep url_creation_total
```

### Test `url_access_total`
```bash
# Access an existing URL (should increment success)
curl http://localhost:8080/api/abc123

# Access non-existent URL (should increment not_found)
curl http://localhost:8080/api/invalid

# View metric
curl -s http://localhost:8080/metrics | grep url_access_total
```

### Test Cache
```bash
# First access (cache miss)
curl http://localhost:8080/api/abc123

# Second access (cache hit, if Redis is running)
curl http://localhost:8080/api/abc123

# View metrics
curl -s http://localhost:8080/metrics | grep cache
```

---

## 📈 Visualization in Prometheus

### URLs Created
```promql
# Cumulative total
sum(url_creation_total{status="success"})

# Rate per second
rate(url_creation_total{status="success"}[1m])
```

### URLs Accessed
```promql
# Total accesses (all statuses)
sum(rate(url_access_total[1m]))

# Success vs error rate
sum by (status) (rate(url_access_total[5m]))
```

### Cache Performance
```promql
# Hit rate percentage
sum(rate(cache_hits_total[5m])) /
(sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m]))) * 100

# Hits per second
rate(cache_hits_total[1m])

# Misses per second
rate(cache_misses_total[1m])
```

---

## 🔍 Debug

If a metric is not appearing:

1. **Check if the feature is being used**
   - `url_creation_total` only appears after creating a URL
   - `cache_hits_total` only appears if Redis is running

2. **Check the code flow**
   - Add logs to confirm code is executing
   - Example: add `fmt.Println()` before incrementing metric

3. **Check the `/metrics` endpoint**
   ```bash
   curl http://localhost:8080/metrics | grep metric_name
   ```

4. **Check in Prometheus**
   - Go to http://localhost:9090
   - Execute the query: `metric_name`

---

## 📝 Quick Summary

| Metric | File | Line(s) | When |
|--------|------|---------|------|
| `url_creation_total` | `service/url_service.go` | 74, 68 | Create URL |
| `url_access_total` | `service/url_service.go` | 102, 93, 85, 97 | Access URL |
| `cache_hits_total` | `repository/url_repository.go` | 124 | Cache hit |
| `cache_misses_total` | `repository/url_repository.go` | 129 | Cache miss |
| HTTP metrics | `middleware/metrics.go` | 48-55 | Every request |
| System metrics | `metrics/metrics.go` | 145-166 | Every 15s |

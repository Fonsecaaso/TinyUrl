# 📊 Latency Metrics by Endpoint

## ✅ What Was Implemented

The application **was already collecting** latency per endpoint from the start! The `http_request_duration_seconds` metric has the following labels:
- `method` - HTTP method (GET, POST, PUT, DELETE, etc.)
- `path` - Endpoint path (e.g., `/api/`, `/api/health`)
- `status` - Response status code (200, 404, 500, etc.)

## 🆕 New Panels in Grafana

**4 new panels** were added to the Grafana dashboard:

### 1. **P95 Latency by Endpoint** (Line Chart)
- **Position**: Full width, right after the stats panels
- **Query**: `histogram_quantile(0.95, sum by (method, path, le) (rate(http_request_duration_seconds_bucket[5m])))`
- **Shows**: Each line represents a specific endpoint with its HTTP method
- **Example**: `GET /api/health`, `POST /api/`, `GET /api/:id`

### 2. **Top 10 Slowest Endpoints (P95)** (Table)
- **Position**: Left side
- **Query**: `topk(10, histogram_quantile(0.95, sum by (method, path, le) (rate(http_request_duration_seconds_bucket[5m]))))`
- **Columns**:
  - **Method**: GET, POST, etc.
  - **Endpoint**: URL path
  - **P95 Latency**: P95 latency in seconds
- **Sorting**: By latency (highest first)
- **Colors**: Green < 50ms, Yellow < 100ms, Orange < 500ms, Red > 500ms

### 3. **Top 10 Slowest Endpoints (Average)** (Table)
- **Position**: Right side
- **Query**: `topk(10, sum by (method, path) (rate(http_request_duration_seconds_sum[5m])) / sum by (method, path) (rate(http_request_duration_seconds_count[5m])))`
- **Columns**:
  - **Method**: GET, POST, etc.
  - **Endpoint**: URL path
  - **Avg Latency**: Average latency in seconds
- **Sorting**: By average latency (highest first)
- **Colors**: Green < 20ms, Yellow < 100ms, Red > 100ms

### 4. **Average Latency by Endpoint (detailed)** (Line Chart)
- **Position**: Full width, bottom section
- **Query**: `rate(http_request_duration_seconds_sum[1m]) / rate(http_request_duration_seconds_count[1m])`
- **Shows**: Average latency of ALL endpoints over time
- **Legend**: Shows `METHOD PATH (STATUS)` - example: `POST /api/ (201)`

## 🔍 How to Use

### In Grafana

1. **Access**: http://localhost:3000
2. **Login**: admin / admin
3. **Dashboard**: TinyURL API Metrics
4. **Scroll to**: Latency panels (after the stats panels)

### Identify Slow Endpoints

**Method 1: Top 10 Tables**
- Look at the two tables side by side
- Compare P95 vs Average to identify variability
- If P95 >> Average, it means there are latency spikes

**Method 2: Line Chart**
- See which lines are highest
- Identify patterns over time
- Detect if latency increases at specific times

### Difference Between P95 and Average

- **Average**: Sum of all latencies / number of requests
  - Can hide latency spikes
  - Sensitive to outliers

- **P95**: 95% of requests are below this value
  - Better for SLA (Service Level Agreement)
  - Shows the experience of most users
  - Example: P95 = 100ms means 95% of requests are < 100ms

## 📊 Useful Queries in Prometheus

### Latency by Endpoint + Method
```promql
# P95 by method and path
histogram_quantile(0.95, sum by (method, path, le) (rate(http_request_duration_seconds_bucket[5m])))

# P99 by method and path
histogram_quantile(0.99, sum by (method, path, le) (rate(http_request_duration_seconds_bucket[5m])))

# Average by method and path
sum by (method, path) (rate(http_request_duration_seconds_sum[5m])) / sum by (method, path) (rate(http_request_duration_seconds_count[5m]))
```

### Top N Slowest Endpoints
```promql
# Top 10 endpoints by P95
topk(10, histogram_quantile(0.95, sum by (method, path, le) (rate(http_request_duration_seconds_bucket[5m]))))

# Top 5 endpoints by average
topk(5, sum by (method, path) (rate(http_request_duration_seconds_sum[5m])) / sum by (method, path) (rate(http_request_duration_seconds_count[5m])))
```

### Compare Methods on Same Endpoint
```promql
# Compare GET vs POST on /api/
histogram_quantile(0.95, sum by (method, le) (rate(http_request_duration_seconds_bucket{path="/api/"}[5m])))
```

### Latency by Status Code
```promql
# Compare success vs error latency
histogram_quantile(0.95, sum by (path, status, le) (rate(http_request_duration_seconds_bucket[5m])))

# Only error requests (5xx)
histogram_quantile(0.95, sum by (path, le) (rate(http_request_duration_seconds_bucket{status=~"5.."}[5m])))
```

## 🎯 Analysis Examples

### Example 1: POST slower than GET on same endpoint
```promql
# See all methods on /api/
histogram_quantile(0.95, sum by (method, le) (rate(http_request_duration_seconds_bucket{path="/api/"}[5m])))
```

**Expected result**:
- POST /api/ → ~50ms (URL creation involves DB write)
- GET /api/:id → ~5ms (read can come from cache)

### Example 2: Identify problematic endpoint
If the "Top 10 Slowest Endpoints (P95)" panel shows:

| Method | Endpoint | P95 Latency |
|--------|----------|-------------|
| POST   | /api/    | 0.850s      |
| GET    | /api/:id | 0.012s      |

This indicates that POST /api/ is taking 850ms at P95, which is high! Possible causes:
- Slow SQL queries
- Missing database indexes
- ID generation taking multiple attempts
- Network latency with Redis/Postgres

### Example 3: Detect performance degradation
Compare the "P95 Latency by Endpoint" chart across different time periods:
- If an endpoint's line is constantly rising → investigate
- If there are spikes at specific times → could be load

## 🔧 Troubleshooting

### Empty panel or no data?
```bash
# 1. Check if the app is running
curl http://localhost:8080/metrics | grep http_request_duration

# 2. Check if Prometheus is collecting
open http://localhost:9090/targets
# Status should be UP

# 3. Make some requests
for i in {1..20}; do curl http://localhost:8080/api/health; done
```

### Too many endpoints on chart?
- Use the "Top 10" table to focus on the most important ones
- On the line chart, click the legend to hide/show specific endpoints

### Latency seems high?
Compare with benchmarks:
- **< 10ms**: Excellent (cache hit)
- **10-50ms**: Good (fast DB query)
- **50-100ms**: Acceptable
- **100-500ms**: Needs attention
- **> 500ms**: Problematic

## 📝 Summary

✅ **Metric already existed**: `http_request_duration_seconds` with labels `method`, `path`, `status`

✅ **4 new panels**:
- P95 chart by endpoint
- Top 10 table (P95)
- Top 10 table (Average)
- Detailed average chart

✅ **HTTP method separation**: Each METHOD + PATH combination is treated separately

✅ **Documentation updated**: [METRICS.md](../go-server/METRICS.md) with specific queries

---

**Suggested next steps**:
1. Configure alerts in Prometheus when P95 > threshold
2. Add distributed tracing with Tempo for latency debugging
3. Create dashboards by region/user if needed

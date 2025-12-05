# 🎯 Grafana - Quick Start (2 minutes)

## Step 1: Access Grafana

Open your browser:
```
http://localhost:3000
```

**Login:**
- Username: `admin`
- Password: `admin`

## Step 2: Access the Dashboard

1. Click **"Dashboards"** in the left sidebar (icon with 4 squares)
2. You'll see: **"TinyURL API Metrics"**
3. Click on it

## Step 3: Generate Traffic (optional)

If the charts are empty, generate some requests:

```bash
# Create URLs
curl -X POST http://localhost:8080/api/ \
  -H "Content-Type: application/json" \
  -d '{"url":"https://google.com"}'

# Health checks
for i in {1..20}; do curl http://localhost:8080/api/health; done
```

## 🎉 Done!

You should see:
- ✅ Request rate per second
- ✅ Latency (P50, P95, P99)
- ✅ Throughput
- ✅ Error rate
- ✅ Memory usage
- ✅ Number of goroutines
- ✅ Requests being processed
- ✅ Cache hit rate
- ✅ URL statistics

## 🔧 Problems?

### Dashboard doesn't appear?
```bash
# Restart Grafana
cd observability
docker-compose restart grafana

# Wait 10 seconds and reload the page
```

### Charts without data?
1. Make sure the Go application is running:
   ```bash
   curl http://localhost:8080/metrics
   ```

2. Check if Prometheus is collecting:
   ```
   http://localhost:9090/targets
   ```
   The "go-backend" status should be **UP** (green)

### Still not working?
See the complete guide: [GRAFANA_SETUP.md](./GRAFANA_SETUP.md)

## 💡 Tip

The dashboard automatically updates every **5 seconds**.

To see more history, change the time range in the top right corner:
- Click "Last 15 minutes"
- Choose "Last 1 hour" or "Last 6 hours"

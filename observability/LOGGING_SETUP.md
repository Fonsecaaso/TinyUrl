# 📝 Logging Setup with OpenTelemetry

This document explains how logging is configured in the TinyURL application using OpenTelemetry and Loki.

## Architecture

```
┌─────────────────┐
│   Go Server     │
│   (Zap Logger)  │
│                 │
│   ↓ OTLP/HTTP  │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│ otel-collector  │ ← Receives logs via OTLP
│   :4318 (HTTP)  │ ← Processes and batches
│   :4317 (gRPC)  │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│      Loki       │ ← Stores logs
│     :3100       │
└────────┬────────┘
         │
         ↓
┌─────────────────┐
│    Grafana      │ ← Visualizes logs
│     :3000       │
└─────────────────┘
```

## Components

### 1. Go Application (Logger)

The Go application uses **Zap** as the logging library with an **OpenTelemetry bridge** that automatically sends logs to the otel-collector.

**Configuration**: [internal/logger/otel_logger.go](../go-server/internal/logger/otel_logger.go)

Key features:
- Dual output: console (for development) + OpenTelemetry (for collection)
- Automatic log level mapping
- Context propagation
- Structured logging with fields

### 2. OpenTelemetry Collector

The collector receives logs via OTLP protocol and forwards them to Loki.

**Configuration**: [otel-collector/otel.yaml](./otel-collector/otel.yaml)

Pipeline:
```yaml
logs:
  receivers: [otlp]        # Receive logs from Go app
  processors: [resource, batch]  # Add metadata and batch
  exporters: [loki, debug]  # Send to Loki and console
```

### 3. Loki

Loki stores and indexes logs for querying.

**Configuration**: [loki/loki.yaml](./loki/loki.yaml)

Access: http://localhost:3100

### 4. Grafana

Grafana provides a UI to query and visualize logs from Loki.

Access: http://localhost:3000 (admin/admin)

## Configuration

### Environment Variables

Set in [.env](../go-server/.env):

```bash
# OpenTelemetry configuration
OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4318"
```

### Service Name

The service name is configured in the logger initialization:

```go
logger.InitLogger("tinyurl-api", "development")
```

This name is used as a label in Loki to filter logs.

## Usage

### Starting the Stack

1. Start observability services:
```bash
cd observability
docker-compose up -d
```

2. Start the Go application:
```bash
cd go-server
go run main.go
```

### Viewing Logs in Grafana

1. Open http://localhost:3000
2. Login: `admin` / `admin`
3. Go to **Explore** (compass icon)
4. Select **Loki** as data source
5. Use LogQL to query logs:

```logql
# All logs from TinyURL service
{service_name="tinyurl-api"}

# Filter by log level
{service_name="tinyurl-api"} |= "level=error"

# Filter by message content
{service_name="tinyurl-api"} |= "postgres"

# Count errors in last 5 minutes
sum(count_over_time({service_name="tinyurl-api"} |= "level=error"[5m]))
```

### Testing Logs

Run the test script:

```bash
cd go-server
./test-logs.sh
```

This will:
1. Check if all services are running
2. Generate traffic to create logs
3. Query Loki to verify logs are being stored
4. Display sample logs

## Log Levels

The application uses the following log levels:

- **DEBUG**: Detailed information for debugging
- **INFO**: General informational messages
- **WARN**: Warning messages (non-critical issues)
- **ERROR**: Error messages (requires attention)
- **FATAL**: Critical errors (application will terminate)

Example in code:

```go
logger.Logger.Info("server started",
    zap.String("port", "8080"),
    zap.String("environment", "development"),
)

logger.Logger.Error("failed to connect to database",
    zap.Error(err),
    zap.String("host", "localhost"),
)
```

## Structured Logging

All logs are structured with fields:

```go
logger.Logger.Info("URL created",
    zap.String("short_code", "abc123"),
    zap.String("original_url", "https://example.com"),
    zap.String("user_id", "user123"),
    zap.Duration("duration", elapsed),
)
```

In Loki, these fields are stored as labels and can be used for filtering:

```logql
{service_name="tinyurl-api"} | json | short_code="abc123"
```

## Labels

Logs are automatically tagged with these labels:

- `service_name`: Service identifier (e.g., "tinyurl-api")
- `deployment_environment`: Environment (e.g., "development", "production")
- `level`: Log level (debug, info, warn, error, fatal)

Additional labels can be added in the resource processor in [otel.yaml](./otel-collector/otel.yaml).

## Performance

### Batching

Logs are batched before sending to Loki to improve performance:

```yaml
processors:
  batch:
    timeout: 10s
    send_batch_size: 1024
```

### Buffering

The Go logger buffers logs in memory and flushes periodically. To manually flush:

```go
logger.Sync()
```

## Troubleshooting

### Logs not appearing in Loki

1. **Check if otel-collector is running:**
```bash
docker ps | grep otel-collector
```

2. **Check otel-collector logs:**
```bash
docker logs otel-collector -f
```

3. **Verify OTLP endpoint:**
```bash
curl http://localhost:4318/v1/logs
```

4. **Check Loki status:**
```bash
curl http://localhost:3100/ready
```

### High log volume

If logs are too verbose, adjust the log level in production:

```go
// In production, use higher log level
if os.Getenv("GO_ENV") == "production" {
    core := zapcore.NewCore(encoder, sink, zapcore.InfoLevel)
}
```

### Logs are delayed

Logs are batched for performance. To see logs faster, reduce batch timeout:

```yaml
processors:
  batch:
    timeout: 1s  # Reduce from 10s
```

## Migration from Promtail

Previously, logs were collected using Promtail which read log files from disk. Now:

✅ **Benefits of OTLP + OpenTelemetry:**
- No file I/O overhead
- Structured logs from the start
- Single telemetry pipeline (logs + traces + metrics)
- Better correlation between signals
- More control over log processing

❌ **Removed:**
- Promtail service
- File-based log collection
- promtail.yaml configuration

## Advanced Configuration

### Custom Log Processors

Add custom processors in [otel.yaml](./otel-collector/otel.yaml):

```yaml
processors:
  # Filter out debug logs in production
  filter/drop_debug:
    logs:
      exclude:
        match_type: strict
        record_attributes:
          - key: level
            value: debug

service:
  pipelines:
    logs:
      receivers: [otlp]
      processors: [resource, filter/drop_debug, batch]
      exporters: [loki]
```

### Multiple Exporters

Send logs to multiple destinations:

```yaml
exporters:
  loki:
    endpoint: http://loki:3100/loki/api/v1/push

  elasticsearch:
    endpoints: [http://elasticsearch:9200]

service:
  pipelines:
    logs:
      receivers: [otlp]
      processors: [resource, batch]
      exporters: [loki, elasticsearch]
```

## Best Practices

1. **Use structured logging**: Always use fields instead of string concatenation
2. **Set appropriate log levels**: Don't log sensitive data, use DEBUG sparingly
3. **Add context**: Include relevant IDs (request_id, user_id, etc.)
4. **Avoid high cardinality labels**: Don't use unique values as labels
5. **Monitor log volume**: Set up alerts for unusual log patterns

## Resources

- [OpenTelemetry Logs](https://opentelemetry.io/docs/concepts/signals/logs/)
- [Loki LogQL](https://grafana.com/docs/loki/latest/query/)
- [Zap Logger](https://github.com/uber-go/zap)
- [OTLP Specification](https://opentelemetry.io/docs/specs/otlp/)

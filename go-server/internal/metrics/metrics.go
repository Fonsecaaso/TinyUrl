package metrics

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	once      sync.Once
	meter     metric.Meter
	initError error

	// HTTP Metrics
	httpRequestsTotal    metric.Int64Counter
	httpRequestDuration  metric.Float64Histogram
	httpRequestSize      metric.Int64Histogram
	httpResponseSize     metric.Int64Histogram
	httpRequestsInFlight metric.Int64UpDownCounter

	// Application Metrics
	urlCreationTotal metric.Int64Counter
	urlAccessTotal   metric.Int64Counter
	cacheHitsTotal   metric.Int64Counter
	cacheMissesTotal metric.Int64Counter

	// Database Metrics
	dbQueryDuration metric.Float64Histogram

	// metricsReady tracks if metrics are initialized
	metricsReady bool
)

// InitMetrics initializes all OTEL metrics instruments
func InitMetrics() error {
	var err error
	once.Do(func() {
		meter = otel.Meter("github.com/fonsecaaso/TinyUrl/go-server")

		// HTTP Metrics
		httpRequestsTotal, err = meter.Int64Counter(
			"http.requests.total",
			metric.WithDescription("Total number of HTTP requests"),
		)
		if err != nil {
			initError = err
			return
		}

		httpRequestDuration, err = meter.Float64Histogram(
			"http.request.duration",
			metric.WithDescription("HTTP request duration in seconds"),
			metric.WithUnit("s"),
			metric.WithExplicitBucketBoundaries(
				0.001, // 1ms
				0.005, // 5ms
				0.010, // 10ms
				0.025, // 25ms
				0.050, // 50ms
				0.100, // 100ms
				0.250, // 250ms
				0.500, // 500ms
				1.0,   // 1s
				2.5,   // 2.5s
				5.0,   // 5s
				10.0,  // 10s
			),
		)
		if err != nil {
			initError = err
			return
		}

		httpRequestSize, err = meter.Int64Histogram(
			"http.request.size",
			metric.WithDescription("HTTP request size in bytes"),
			metric.WithUnit("By"),
			metric.WithExplicitBucketBoundaries(
				100,     // 100 bytes
				1024,    // 1 KB
				5120,    // 5 KB
				10240,   // 10 KB
				51200,   // 50 KB
				102400,  // 100 KB
				512000,  // 500 KB
				1048576, // 1 MB
				5242880, // 5 MB
			),
		)
		if err != nil {
			initError = err
			return
		}

		httpResponseSize, err = meter.Int64Histogram(
			"http.response.size",
			metric.WithDescription("HTTP response size in bytes"),
			metric.WithUnit("By"),
			metric.WithExplicitBucketBoundaries(
				100,     // 100 bytes
				1024,    // 1 KB
				5120,    // 5 KB
				10240,   // 10 KB
				51200,   // 50 KB
				102400,  // 100 KB
				512000,  // 500 KB
				1048576, // 1 MB
				5242880, // 5 MB
			),
		)
		if err != nil {
			initError = err
			return
		}

		httpRequestsInFlight, err = meter.Int64UpDownCounter(
			"http.requests.in_flight",
			metric.WithDescription("Current number of HTTP requests being processed"),
		)
		if err != nil {
			initError = err
			return
		}

		// Application Metrics
		urlCreationTotal, err = meter.Int64Counter(
			"url.creation.total",
			metric.WithDescription("Total number of URLs created"),
		)
		if err != nil {
			initError = err
			return
		}

		urlAccessTotal, err = meter.Int64Counter(
			"url.access.total",
			metric.WithDescription("Total number of URL accesses/redirects"),
		)
		if err != nil {
			initError = err
			return
		}

		cacheHitsTotal, err = meter.Int64Counter(
			"cache.hits.total",
			metric.WithDescription("Total number of cache hits"),
		)
		if err != nil {
			initError = err
			return
		}

		cacheMissesTotal, err = meter.Int64Counter(
			"cache.misses.total",
			metric.WithDescription("Total number of cache misses"),
		)
		if err != nil {
			initError = err
			return
		}

		// Database Metrics
		dbQueryDuration, err = meter.Float64Histogram(
			"db.query.duration",
			metric.WithDescription("Database query duration in seconds"),
			metric.WithUnit("s"),
			metric.WithExplicitBucketBoundaries(
				0.001, // 1ms
				0.005, // 5ms
				0.010, // 10ms
				0.025, // 25ms
				0.050, // 50ms
				0.100, // 100ms
				0.250, // 250ms
				0.500, // 500ms
				1.0,   // 1s
				2.5,   // 2.5s
				5.0,   // 5s
			),
		)
		if err != nil {
			initError = err
			return
		}

		metricsReady = true
	})

	return initError
}

// StartSystemMetricsCollection starts collecting system metrics periodically
func StartSystemMetricsCollection() error {
	if !metricsReady {
		if err := InitMetrics(); err != nil {
			return err
		}
	}

	// Register observable gauges for system metrics
	_, err := meter.Int64ObservableGauge(
		"go.goroutines",
		metric.WithDescription("Number of goroutines"),
		metric.WithInt64Callback(func(ctx context.Context, observer metric.Int64Observer) error {
			observer.Observe(int64(runtime.NumGoroutine()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	// Memory metrics - single gauge with type label
	_, err = meter.Int64ObservableGauge(
		"memory.usage",
		metric.WithDescription("Memory usage in bytes by type"),
		metric.WithUnit("By"),
		metric.WithInt64Callback(func(ctx context.Context, observer metric.Int64Observer) error {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			// heap_alloc: bytes of allocated heap objects
			observer.Observe(int64(m.Alloc), metric.WithAttributes(
				attribute.String("type", "heap_alloc"),
			))

			// heap_in_use: bytes in in-use spans
			observer.Observe(int64(m.HeapInuse), metric.WithAttributes(
				attribute.String("type", "heap_in_use"),
			))

			// sys: total bytes of memory obtained from the OS
			observer.Observe(int64(m.Sys), metric.WithAttributes(
				attribute.String("type", "sys"),
			))

			return nil
		}),
	)
	if err != nil {
		return err
	}

	return nil
}

// StartDatabaseMetricsCollection starts collecting database connection pool metrics
func StartDatabaseMetricsCollection(pool *pgxpool.Pool) error {
	if !metricsReady {
		if err := InitMetrics(); err != nil {
			return err
		}
	}

	if pool == nil {
		return nil // No pool provided, skip DB metrics
	}

	// Database connections in use
	_, err := meter.Int64ObservableGauge(
		"db.connections.in_use",
		metric.WithDescription("Number of database connections currently in use"),
		metric.WithInt64Callback(func(ctx context.Context, observer metric.Int64Observer) error {
			stats := pool.Stat()
			observer.Observe(int64(stats.AcquiredConns()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	// Database connections idle
	_, err = meter.Int64ObservableGauge(
		"db.connections.idle",
		metric.WithDescription("Number of idle database connections in the pool"),
		metric.WithInt64Callback(func(ctx context.Context, observer metric.Int64Observer) error {
			stats := pool.Stat()
			observer.Observe(int64(stats.IdleConns()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	return nil
}

// RecordHTTPMetrics records metrics for an HTTP request
func RecordHTTPMetrics(ctx context.Context, method, path, status string, duration time.Duration, requestSize, responseSize int64) {
	if !metricsReady {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("path", path),
		attribute.String("status", status),
	)

	httpRequestsTotal.Add(ctx, 1, attrs)
	httpRequestDuration.Record(ctx, duration.Seconds(), attrs)
	httpRequestSize.Record(ctx, requestSize, metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("path", path),
	))
	httpResponseSize.Record(ctx, responseSize, attrs)
}

// RecordURLCreation records URL creation metrics
func RecordURLCreation(ctx context.Context, status string) {
	if !metricsReady {
		return
	}
	urlCreationTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("status", status)))
}

// RecordURLAccess records URL access metrics
func RecordURLAccess(ctx context.Context, status string) {
	if !metricsReady {
		return
	}
	urlAccessTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("status", status)))
}

// RecordCacheHit records cache hit metrics
func RecordCacheHit(ctx context.Context, cacheType string) {
	if !metricsReady {
		return
	}
	cacheHitsTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("cache_type", cacheType)))
}

// RecordCacheMiss records cache miss metrics
func RecordCacheMiss(ctx context.Context, cacheType string) {
	if !metricsReady {
		return
	}
	cacheMissesTotal.Add(ctx, 1, metric.WithAttributes(attribute.String("cache_type", cacheType)))
}

// RecordDBQueryDuration records database query duration
func RecordDBQueryDuration(ctx context.Context, queryType string, duration time.Duration) {
	if !metricsReady {
		return
	}
	dbQueryDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attribute.String("query_type", queryType)))
}

// IncrementRequestsInFlight increments the in-flight requests counter
func IncrementRequestsInFlight(ctx context.Context) {
	if !metricsReady {
		return
	}
	httpRequestsInFlight.Add(ctx, 1)
}

// DecrementRequestsInFlight decrements the in-flight requests counter
func DecrementRequestsInFlight(ctx context.Context) {
	if !metricsReady {
		return
	}
	httpRequestsInFlight.Add(ctx, -1)
}

// IsInitialized returns whether metrics are initialized and ready
func IsInitialized() bool {
	return metricsReady
}

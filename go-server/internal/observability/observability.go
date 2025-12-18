package observability

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/logger"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/tracing"
)

// Observability holds all observability components
type Observability struct {
	tracerShutdown    func(ctx context.Context) error
	meterShutdown     func(ctx context.Context) error
	loggerShutdown    func(ctx context.Context) error
	Logger            *zap.Logger
	PrometheusHandler http.Handler
	initialized       ObservabilityStatus
}

// ObservabilityStatus tracks which components are initialized
type ObservabilityStatus struct {
	TracingEnabled bool
	MetricsEnabled bool
	LoggingEnabled bool
}

// Shutdown gracefully shuts down all observability components
func (o *Observability) Shutdown(ctx context.Context) error {
	var errs []error

	if o.tracerShutdown != nil {
		if err := o.tracerShutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer shutdown: %w", err))
		}
	}

	if o.meterShutdown != nil {
		if err := o.meterShutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter shutdown: %w", err))
		}
	}

	if o.loggerShutdown != nil {
		if err := o.loggerShutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("logger shutdown: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("observability shutdown errors: %v", errs)
	}

	return nil
}

// GetStatus returns the current observability status
func (o *Observability) GetStatus() ObservabilityStatus {
	return o.initialized
}

// SetupObservability initializes all observability components
func SetupObservability(ctx context.Context) (*Observability, error) {
	obs := &Observability{}

	serviceName := getEnv("SERVICE_NAME", "go-backend")
	environment := getEnv("ENV", "development")

	// Initialize Resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.DeploymentEnvironment(environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Add AWS ECS attributes if available
	awsAttrs := []resource.Option{}
	if clusterName := os.Getenv("AWS_ECS_CLUSTER_NAME"); clusterName != "" {
		awsAttrs = append(awsAttrs, resource.WithAttributes(
			semconv.AWSECSClusterARN(clusterName),
		))
	}
	if taskARN := os.Getenv("TASK_ARN"); taskARN != "" {
		awsAttrs = append(awsAttrs, resource.WithAttributes(
			semconv.AWSECSTaskARN(taskARN),
		))
	}

	if len(awsAttrs) > 0 {
		awsRes, err := resource.New(ctx, awsAttrs...)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS resource: %w", err)
		}
		res, err = resource.Merge(res, awsRes)
		if err != nil {
			return nil, fmt.Errorf("failed to merge resource with AWS attributes: %w", err)
		}
	}

	// Initialize Tracing
	tracerShutdown, err := initTracing(ctx, res)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}
	obs.tracerShutdown = tracerShutdown
	obs.initialized.TracingEnabled = true

	// Initialize Metrics (only if OTEL endpoint is set)
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		meterShutdown, promHandler, err := initMetrics(ctx, res)
		if err != nil {
			// Metrics are optional, log warning but continue
			fmt.Printf("Warning: failed to initialize metrics: %v\n", err)
		} else {
			obs.meterShutdown = meterShutdown
			obs.PrometheusHandler = promHandler
			obs.initialized.MetricsEnabled = true
		}
	}

	// Initialize Logging
	loggerShutdown, err := initLogging(serviceName, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logging: %w", err)
	}
	obs.loggerShutdown = loggerShutdown
	obs.Logger = logger.Logger
	obs.initialized.LoggingEnabled = true

	// Set global logger
	zap.ReplaceGlobals(logger.Logger)

	return obs, nil
}

// initTracing initializes OpenTelemetry tracing
func initTracing(ctx context.Context, res *resource.Resource) (func(context.Context) error, error) {
	endpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://observability:4318")
	protocol := getEnv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")

	// Clean protocol value: remove quotes and whitespace
	protocol = strings.TrimSpace(strings.Trim(protocol, `"`))

	if strings.ToLower(protocol) != "http/protobuf" {
		return nil, fmt.Errorf("only http/protobuf protocol is supported, got: %q (cleaned from env)", protocol)
	}

	// Create a temporary logger for initialization (logger.Logger might be nil at this point)
	tempLogger, _ := zap.NewProduction()
	if tempLogger == nil {
		tempLogger = zap.NewNop()
	}

	// Create HTTP client with logging transport
	httpClient := &http.Client{
		Transport: tracing.NewLoggingTransport(tempLogger),
	}

	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(stripProtocol(endpoint)),
		otlptracehttp.WithInsecure(), // Using internal network
		otlptracehttp.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Log initialization (use global logger if available, otherwise temp logger)
	loggerToUse := logger.Logger
	if loggerToUse == nil {
		loggerToUse = tempLogger
	}
	loggerToUse.Info("🚀 OTLP Trace Exporter initialized with logging",
		zap.String("endpoint", endpoint),
		zap.String("protocol", protocol),
	)

	// Configure batch span processor with optimized settings
	bsp := sdktrace.NewBatchSpanProcessor(
		exporter,
		sdktrace.WithMaxExportBatchSize(512),
		sdktrace.WithMaxQueueSize(2048),
		sdktrace.WithBatchTimeout(5000000000), // 5 seconds in nanoseconds
	)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return tracerProvider.Shutdown, nil
}

// initMetrics initializes OpenTelemetry metrics with both OTLP push and Prometheus pull exporters
func initMetrics(ctx context.Context, res *resource.Resource) (func(context.Context) error, http.Handler, error) {
	endpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://observability:4318")

	// Create a temporary logger for initialization
	tempLogger, _ := zap.NewProduction()
	if tempLogger == nil {
		tempLogger = zap.NewNop()
	}

	// Create HTTP client with logging transport for metrics
	httpClient := &http.Client{
		Transport: tracing.NewLoggingTransport(tempLogger),
	}

	// Create OTLP exporter for push model (to OTEL Collector)
	otlpExporter, err := otlpmetrichttp.New(
		ctx,
		otlpmetrichttp.WithEndpoint(stripProtocol(endpoint)),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	// Create a custom Prometheus registry
	registry := prometheus.NewRegistry()

	// Create Prometheus exporter for pull model (direct scraping) with the custom registry
	prometheusExporter, err := promexporter.New(
		promexporter.WithRegisterer(registry),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Log initialization (use global logger if available, otherwise temp logger)
	loggerToUse := logger.Logger
	if loggerToUse == nil {
		loggerToUse = tempLogger
	}
	loggerToUse.Info("📊 Metric Exporters initialized",
		zap.String("otlp_endpoint", endpoint),
		zap.String("prometheus_endpoint", "/api/metrics"),
	)

	// Configure periodic reader for OTLP push with optimized settings
	otlpReader := metric.NewPeriodicReader(
		otlpExporter,
		metric.WithInterval(30000000000), // 30 seconds in nanoseconds
	)

	// Create MeterProvider with both readers
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(otlpReader),        // Push to OTEL Collector
		metric.WithReader(prometheusExporter), // Pull from /metrics endpoint
	)

	otel.SetMeterProvider(meterProvider)

	// Create HTTP handler from the Prometheus registry
	prometheusHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	return meterProvider.Shutdown, prometheusHandler, nil
}

// initLogging initializes structured logging with Loki
func initLogging(serviceName, environment string) (func(context.Context) error, error) {
	if err := logger.InitLokiLogger(serviceName, environment); err != nil {
		return nil, fmt.Errorf("failed to initialize Loki logger: %w", err)
	}

	return logger.Shutdown, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// stripProtocol removes http:// or https:// from the beginning of an endpoint
// and extracts only the host:port part, discarding any path.
// OTLP exporters expect only "host:port" format, as they append the standard paths like /v1/traces
func stripProtocol(endpoint string) string {
	// Clean endpoint: remove quotes and whitespace
	endpoint = strings.TrimSpace(strings.Trim(endpoint, `"`))

	// If endpoint doesn't have a protocol, assume it's already in host:port format
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		// Still need to check for path and extract only host:port
		if idx := strings.Index(endpoint, "/"); idx != -1 {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: OTLP endpoint contains path '%s' - extracting only host:port '%s'\n",
				endpoint, endpoint[:idx])
			return endpoint[:idx]
		}
		return endpoint
	}

	// Parse the URL properly
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		// Fallback to simple string stripping if parsing fails
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Failed to parse OTLP endpoint URL '%s': %v - using fallback\n", endpoint, err)
		if strings.HasPrefix(endpoint, "https://") {
			return endpoint[8:]
		}
		if strings.HasPrefix(endpoint, "http://") {
			return endpoint[7:]
		}
		return endpoint
	}

	// Extract host:port (port is included in parsedURL.Host if present)
	hostPort := parsedURL.Host

	// Log warning if path was present and is being discarded
	if parsedURL.Path != "" && parsedURL.Path != "/" {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: OTLP endpoint '%s' contains path '%s' which will be ignored. OTLP library will append standard paths like /v1/traces\n",
			endpoint, parsedURL.Path)
	}

	return hostPort
}

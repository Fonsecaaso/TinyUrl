package observability

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/logger"
)

// Observability holds all observability components
type Observability struct {
	tracerShutdown func(ctx context.Context) error
	meterShutdown  func(ctx context.Context) error
	loggerShutdown func(ctx context.Context) error
	Logger         *zap.Logger
	initialized    ObservabilityStatus
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
		meterShutdown, err := initMetrics(ctx, res)
		if err != nil {
			// Metrics are optional, log warning but continue
			fmt.Printf("Warning: failed to initialize metrics: %v\n", err)
		} else {
			obs.meterShutdown = meterShutdown
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

	if protocol != "http/protobuf" {
		return nil, fmt.Errorf("only http/protobuf protocol is supported, got: %s", protocol)
	}

	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(stripProtocol(endpoint)),
		otlptracehttp.WithInsecure(), // Using internal network
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

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

// initMetrics initializes OpenTelemetry metrics
func initMetrics(ctx context.Context, res *resource.Resource) (func(context.Context) error, error) {
	endpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://observability:4318")

	exporter, err := otlpmetrichttp.New(
		ctx,
		otlpmetrichttp.WithEndpoint(stripProtocol(endpoint)),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	// Configure periodic reader with optimized settings
	reader := metric.NewPeriodicReader(
		exporter,
		metric.WithInterval(30000000000), // 30 seconds in nanoseconds
	)

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(reader),
	)

	otel.SetMeterProvider(meterProvider)

	return meterProvider.Shutdown, nil
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
func stripProtocol(endpoint string) string {
	if len(endpoint) > 7 && endpoint[:7] == "http://" {
		return endpoint[7:]
	}
	if len(endpoint) > 8 && endpoint[:8] == "https://" {
		return endpoint[8:]
	}
	return endpoint
}

package tracing

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitTracer initializes OpenTelemetry tracing with OTLP HTTP exporter
// Deprecated: Use internal/observability.SetupObservability() instead
func InitTracer() func(context.Context) error {
	ctx := context.Background()

	endpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4318")
	serviceName := getEnv("SERVICE_NAME", "go-backend")
	environment := getEnv("ENV", "development")

	// Strip protocol from endpoint (otlptracehttp expects host:port)
	endpoint = stripProtocol(endpoint)

	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create OTLP exporter: %v", err))
	}

	// Create resource with service attributes
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.DeploymentEnvironment(environment),
		),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create resource: %v", err))
	}

	// Add AWS ECS cluster name if available
	if clusterName := os.Getenv("AWS_ECS_CLUSTER_NAME"); clusterName != "" {
		res, err = resource.Merge(
			res,
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.AWSECSClusterARN(clusterName),
			),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to merge resource: %v", err))
		}
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return tracerProvider.Shutdown
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

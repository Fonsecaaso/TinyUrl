package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger   *zap.Logger
	lokiURL  string
	lokiHTTP *http.Client
	// logBuffer []lokiLog
)

// type lokiLog struct {
// 	Timestamp string
// 	Line      string
// }

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

type lokiPushRequest struct {
	Streams []lokiStream `json:"streams"`
}

// InitLokiLogger initializes a logger that sends logs to Loki and console
func InitLokiLogger(serviceName, environment string) error {
	// Get Loki endpoint from environment or use default
	lokiURL = os.Getenv("LOKI_ENDPOINT")
	if lokiURL == "" {
		lokiURL = os.Getenv("LOKI_URL")
		if lokiURL == "" {
			lokiURL = "http://localhost:3100/loki/api/v1/push"
		}
	}

	// Create HTTP client for Loki with retry support
	lokiHTTP = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
			DisableKeepAlives:   false,
		},
	}

	// Create console encoder for development
	consoleConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoder := zapcore.NewConsoleEncoder(consoleConfig)

	// Create Loki encoder (JSON)
	lokiConfig := zap.NewProductionEncoderConfig()
	lokiConfig.TimeKey = "ts"
	lokiConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	lokiEncoder := zapcore.NewJSONEncoder(lokiConfig)

	// Create Loki core that writes to Loki
	lokiCore := zapcore.NewCore(
		lokiEncoder,
		zapcore.AddSync(&lokiWriter{
			serviceName: serviceName,
			environment: environment,
		}),
		zapcore.DebugLevel,
	)

	// Create console core
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	// Combine both cores (tee)
	core := zapcore.NewTee(consoleCore, lokiCore)

	// Create the logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// lokiWriter implements zapcore.WriteSyncer for Loki
type lokiWriter struct {
	serviceName string
	environment string
}

// Write implements io.Writer
func (w *lokiWriter) Write(p []byte) (n int, err error) {
	// Create a copy of the byte slice to prevent race condition
	// The original buffer may be reused by the logger before the goroutine completes
	logData := make([]byte, len(p))
	copy(logData, p)

	// Send log entry to Loki asynchronously
	go func() {
		timestamp := time.Now().UnixNano()
		timestampStr := fmt.Sprintf("%d", timestamp)

		// Extract structured fields from JSON log to use as labels
		labels := map[string]string{
			"service_name": w.serviceName,
			"environment":  w.environment,
			"job":          "tinyurl-api",
		}

		// Parse JSON to extract important fields as labels and add trace_id/span_id
		var logEntry map[string]interface{}
		if err := json.Unmarshal(logData, &logEntry); err == nil {
			// Extract level (info, warn, error, debug)
			if level, ok := logEntry["level"].(string); ok && level != "" {
				labels["level"] = level
			}

			// Extract HTTP path for filtering by endpoint
			if path, ok := logEntry["path"].(string); ok && path != "" {
				labels["path"] = path
			}

			// Extract HTTP method
			if method, ok := logEntry["method"].(string); ok && method != "" {
				labels["method"] = method
			}

			// Extract status code
			if status, ok := logEntry["status"].(float64); ok {
				labels["status"] = fmt.Sprintf("%d", int(status))
			}

			// Extract trace_id if present
			if traceID, ok := logEntry["trace_id"].(string); ok && traceID != "" {
				labels["trace_id"] = traceID
			}

			// Extract span_id if present
			if spanID, ok := logEntry["span_id"].(string); ok && spanID != "" {
				labels["span_id"] = spanID
			}

			// Extract request_id if present
			if requestID, ok := logEntry["request_id"].(string); ok && requestID != "" {
				labels["request_id"] = requestID
			}
		}

		// Create Loki push request
		pushReq := lokiPushRequest{
			Streams: []lokiStream{
				{
					Stream: labels,
					Values: [][]string{
						{timestampStr, string(logData)},
					},
				},
			},
		}

		jsonData, err := json.Marshal(pushReq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal loki request: %v\n", err)
			return
		}

		// Retry logic with exponential backoff
		maxRetries := 3
		baseDelay := 100 * time.Millisecond

		for attempt := 0; attempt < maxRetries; attempt++ {
			req, err := http.NewRequest("POST", lokiURL, bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create loki request: %v\n", err)
				return
			}

			req.Header.Set("Content-Type", "application/json")

			resp, err := lokiHTTP.Do(req)
			if err != nil {
				// Retry with exponential backoff
				if attempt < maxRetries-1 {
					delay := baseDelay * time.Duration(1<<uint(attempt))
					time.Sleep(delay)
					continue
				}
				fmt.Fprintf(os.Stderr, "failed to send log to loki after %d attempts: %v\n", maxRetries, err)
				return
			}

			if resp.StatusCode >= 400 {
				_ = resp.Body.Close()
				if attempt < maxRetries-1 {
					delay := baseDelay * time.Duration(1<<uint(attempt))
					time.Sleep(delay)
					continue
				}
				fmt.Fprintf(os.Stderr, "loki returned error %d after %d attempts\n", resp.StatusCode, maxRetries)
				return
			}

			_ = resp.Body.Close()
			return // Success
		}
	}()

	return len(p), nil
}

// Sync implements zapcore.WriteSyncer
func (w *lokiWriter) Sync() error {
	// Wait a bit for async writes to complete
	time.Sleep(100 * time.Millisecond)
	return nil
}

// Sync flushes any buffered log entries
func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

// Shutdown gracefully shuts down the logger
func Shutdown(ctx context.Context) error {
	// Wait for async writes
	time.Sleep(500 * time.Millisecond)
	return nil
}

// WithTrace extracts trace and span IDs from context and adds them to the logger
func WithTrace(ctx context.Context, logger *zap.Logger) *zap.Logger {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return logger
	}

	spanContext := span.SpanContext()
	return logger.With(
		zap.String("trace_id", spanContext.TraceID().String()),
		zap.String("span_id", spanContext.SpanID().String()),
	)
}

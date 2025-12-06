package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

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
	lokiURL = os.Getenv("LOKI_URL")
	if lokiURL == "" {
		lokiURL = "http://localhost:3100/loki/api/v1/push"
	}

	// Create HTTP client for Loki
	lokiHTTP = &http.Client{
		Timeout: 10 * time.Second,
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

		// Parse JSON to extract important fields as labels
		var logEntry map[string]interface{}
		if err := json.Unmarshal(p, &logEntry); err == nil {
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
		}

		// Create Loki push request
		pushReq := lokiPushRequest{
			Streams: []lokiStream{
				{
					Stream: labels,
					Values: [][]string{
						{timestampStr, string(p)},
					},
				},
			},
		}

		jsonData, err := json.Marshal(pushReq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal loki request: %v\n", err)
			return
		}

		req, err := http.NewRequest("POST", lokiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create loki request: %v\n", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := lokiHTTP.Do(req)
		if err != nil {
			// Don't fail the write if Loki is unavailable
			fmt.Fprintf(os.Stderr, "failed to send log to loki: %v\n", err)
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode >= 400 {
			fmt.Fprintf(os.Stderr, "loki returned error: %d\n", resp.StatusCode)
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

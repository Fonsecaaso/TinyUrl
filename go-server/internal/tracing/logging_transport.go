package tracing

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// loggingTransport is an HTTP transport that logs all requests and responses
type loggingTransport struct {
	base   http.RoundTripper
	logger *zap.Logger
}

// NewLoggingTransport creates a new HTTP transport with logging capabilities
func NewLoggingTransport(logger *zap.Logger) http.RoundTripper {
	return &loggingTransport{
		base:   http.DefaultTransport,
		logger: logger,
	}
}

// RoundTrip implements http.RoundTripper interface with logging
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	startTime := time.Now()

	// Log the outgoing request
	t.logRequest(req)

	// Execute the actual request
	resp, err := t.base.RoundTrip(req)

	duration := time.Since(startTime)

	// Log the response
	if err != nil {
		t.logger.Error("OTLP request failed",
			zap.String("method", req.Method),
			zap.String("url", req.URL.String()),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return nil, err
	}

	t.logResponse(resp, duration)

	return resp, nil
}

// logRequest logs details about the outgoing request
func (t *loggingTransport) logRequest(req *http.Request) {
	// Read and log request body
	var bodyPreview string
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			// Restore the body for actual sending
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Try to decompress if it's gzipped
			decompressed := bodyBytes
			if req.Header.Get("Content-Encoding") == "gzip" {
				if reader, err := gzip.NewReader(bytes.NewReader(bodyBytes)); err == nil {
					if uncompressed, err := io.ReadAll(reader); err == nil {
						decompressed = uncompressed
					}
					reader.Close()
				}
			}

			bodyPreview = formatBodyPreview(decompressed)
		}
	}

	t.logger.Info("📤 Sending OTLP request to otel-collector",
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
		zap.String("host", req.Host),
		zap.String("path", req.URL.Path),
		zap.Int64("content_length", req.ContentLength),
		zap.String("content_type", req.Header.Get("Content-Type")),
		zap.String("content_encoding", req.Header.Get("Content-Encoding")),
		zap.String("user_agent", req.Header.Get("User-Agent")),
		zap.String("body_preview", bodyPreview),
		zap.Any("headers", formatHeaders(req.Header)),
	)
}

// logResponse logs details about the response received
func (t *loggingTransport) logResponse(resp *http.Response, duration time.Duration) {
	// Read response body
	var bodyPreview string
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			// Restore the body for the caller
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			bodyPreview = formatBodyPreview(bodyBytes)
		}
	}

	logFunc := t.logger.Info
	emoji := "✅"
	message := "Received OTLP response from otel-collector"

	if resp.StatusCode >= 400 {
		logFunc = t.logger.Error
		emoji = "❌"
		message = "OTLP request failed with error status"
	} else if resp.StatusCode >= 300 {
		logFunc = t.logger.Warn
		emoji = "⚠️"
		message = "OTLP request received redirect"
	}

	logFunc(fmt.Sprintf("%s %s", emoji, message),
		zap.Int("status_code", resp.StatusCode),
		zap.String("status", resp.Status),
		zap.Duration("duration_ms", duration),
		zap.Int64("content_length", resp.ContentLength),
		zap.String("content_type", resp.Header.Get("Content-Type")),
		zap.String("body_preview", bodyPreview),
		zap.Any("headers", formatHeaders(resp.Header)),
	)
}

// formatBodyPreview formats the body bytes into a readable preview
func formatBodyPreview(bodyBytes []byte) string {
	if len(bodyBytes) == 0 {
		return "(empty)"
	}

	// Limit preview to first 500 characters
	maxPreview := 500
	preview := string(bodyBytes)
	if len(preview) > maxPreview {
		preview = preview[:maxPreview] + "... (truncated)"
	}

	// Check if it's binary data (protobuf)
	if !isPrintable(preview) {
		return fmt.Sprintf("(binary protobuf data, %d bytes)", len(bodyBytes))
	}

	// Try to make it more readable
	preview = strings.ReplaceAll(preview, "\n", " ")
	preview = strings.ReplaceAll(preview, "\t", " ")

	return preview
}

// isPrintable checks if a string contains mostly printable characters
func isPrintable(s string) bool {
	printableCount := 0
	for _, r := range s {
		if r >= 32 && r <= 126 || r == '\n' || r == '\t' {
			printableCount++
		}
	}
	return len(s) > 0 && float64(printableCount)/float64(len(s)) > 0.7
}

// formatHeaders formats HTTP headers for logging, hiding sensitive data
func formatHeaders(headers http.Header) map[string]string {
	formatted := make(map[string]string)
	for key, values := range headers {
		// Hide sensitive headers
		lowerKey := strings.ToLower(key)
		if strings.Contains(lowerKey, "authorization") ||
		   strings.Contains(lowerKey, "token") ||
		   strings.Contains(lowerKey, "secret") {
			formatted[key] = "***REDACTED***"
		} else {
			formatted[key] = strings.Join(values, ", ")
		}
	}
	return formatted
}

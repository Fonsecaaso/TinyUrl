package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// lokiLoggingTransport is an HTTP transport that logs Loki requests
// Note: This logs to stderr directly to avoid infinite recursion
type lokiLoggingTransport struct {
	base http.RoundTripper
}

// NewLokiLoggingTransport creates a new HTTP transport with logging for Loki
func NewLokiLoggingTransport() http.RoundTripper {
	return &lokiLoggingTransport{
		base: http.DefaultTransport,
	}
}

// RoundTrip implements http.RoundTripper interface with logging
func (t *lokiLoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	startTime := time.Now()

	// Log the outgoing request to stderr (to avoid recursion)
	t.logRequest(req)

	// Execute the actual request
	resp, err := t.base.RoundTrip(req)

	duration := time.Since(startTime)

	// Log the response
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ [LOKI] Request failed: method=%s url=%s duration=%v error=%v\n",
			req.Method, req.URL.String(), duration, err)
		return nil, err
	}

	t.logResponse(resp, duration)

	return resp, nil
}

// logRequest logs details about the outgoing request
func (t *lokiLoggingTransport) logRequest(req *http.Request) {
	// Read and log request body
	var bodyPreview string
	var bodySize int
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			bodySize = len(bodyBytes)
			// Restore the body for actual sending
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			bodyPreview = formatBodyPreviewSimple(bodyBytes, 300)
		}
	}

	fmt.Fprintf(os.Stderr, "📤 [LOKI] Sending logs: method=%s url=%s size=%d content_type=%s\n",
		req.Method,
		req.URL.String(),
		bodySize,
		req.Header.Get("Content-Type"),
	)

	if bodyPreview != "" {
		fmt.Fprintf(os.Stderr, "   [LOKI] Body preview: %s\n", bodyPreview)
	}
}

// logResponse logs details about the response received
func (t *lokiLoggingTransport) logResponse(resp *http.Response, duration time.Duration) {
	// Read response body
	var bodyPreview string
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			// Restore the body for the caller
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			bodyPreview = formatBodyPreviewSimple(bodyBytes, 200)
		}
	}

	emoji := "✅"
	if resp.StatusCode >= 400 {
		emoji = "❌"
	} else if resp.StatusCode >= 300 {
		emoji = "⚠️"
	}

	fmt.Fprintf(os.Stderr, "%s [LOKI] Response: status=%d (%s) duration=%v\n",
		emoji,
		resp.StatusCode,
		resp.Status,
		duration,
	)

	if bodyPreview != "" && resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "   [LOKI] Error body: %s\n", bodyPreview)
	}
}

// formatBodyPreviewSimple formats the body bytes into a simple preview
func formatBodyPreviewSimple(bodyBytes []byte, maxLen int) string {
	if len(bodyBytes) == 0 {
		return "(empty)"
	}

	preview := string(bodyBytes)

	// Remove newlines and extra spaces for compact output
	preview = strings.ReplaceAll(preview, "\n", " ")
	preview = strings.ReplaceAll(preview, "\t", " ")
	preview = strings.Join(strings.Fields(preview), " ")

	if len(preview) > maxLen {
		preview = preview[:maxLen] + "..."
	}

	return preview
}

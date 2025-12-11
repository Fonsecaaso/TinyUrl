package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/metrics"
	"github.com/gin-gonic/gin"
)

// MetricsMiddleware collects HTTP metrics for each request
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Increment in-flight requests
		metrics.IncrementRequestsInFlight(ctx)
		defer metrics.DecrementRequestsInFlight(ctx)

		// Record start time
		start := time.Now()

		// Get request size
		requestSize := computeApproximateRequestSize(c.Request)

		// Create custom response writer to capture response size
		blw := &bodyLogWriter{body: make([]byte, 0), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get path (use route pattern instead of actual path for better grouping)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Get status
		status := strconv.Itoa(c.Writer.Status())

		// Get response size
		responseSize := int64(len(blw.body))

		// Record metrics with context
		metrics.RecordHTTPMetrics(
			ctx,
			c.Request.Method,
			path,
			status,
			duration,
			requestSize,
			responseSize,
		)
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}


// computeApproximateRequestSize calculates approximate request size
func computeApproximateRequestSize(r *http.Request) int64 {
	s := int64(0)

	// Add content length if available
	if r.ContentLength > 0 {
		s += r.ContentLength
	}

	// Add approximate size of URL, method, and headers
	s += int64(len(r.Method))
	s += int64(len(r.URL.String()))
	s += int64(len(r.Header) * 50) // Rough estimate: 50 bytes per header

	return s
}

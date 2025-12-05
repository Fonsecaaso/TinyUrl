package middleware

import (
	"strconv"
	"time"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/metrics"
	"github.com/gin-gonic/gin"
)

// MetricsMiddleware collects HTTP metrics for each request
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Increment in-flight requests
		metrics.HTTPRequestsInFlight.Inc()
		defer metrics.HTTPRequestsInFlight.Dec()

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

		// Record metrics
		metrics.RecordHTTPMetrics(
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
func computeApproximateRequestSize(r any) int64 {
	// Type assertion to get the actual request
	req, ok := r.(interface {
		ContentLength() int64
		URL() interface{ String() string }
		Method() string
		Header() interface{ Len() int }
	})

	if !ok {
		// Fallback for standard http.Request
		return 0
	}

	s := int64(0)

	// Add content length if available
	if req.ContentLength() > 0 {
		s += req.ContentLength()
	}

	// Add approximate size of URL, method, and headers
	s += int64(len(req.Method()))
	s += int64(len(req.URL().String()))
	s += int64(req.Header().Len() * 50) // Rough estimate: 50 bytes per header

	return s
}

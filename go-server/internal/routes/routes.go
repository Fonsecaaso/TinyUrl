package route

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/handler"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/metrics"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/middleware"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/repository"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/service"
)

func SetupRouter(redisClient *redis.Client, pgClient *pgxpool.Pool, prometheusHandler http.Handler) *gin.Engine {
	r := gin.New()

	// Start system metrics collection
	if err := metrics.StartSystemMetricsCollection(); err != nil {
		// Log error but don't fail - metrics are optional
		zap.L().Warn("Failed to start system metrics collection", zap.Error(err))
	}

	// Start database metrics collection
	if err := metrics.StartDatabaseMetricsCollection(pgClient); err != nil {
		// Log error but don't fail - metrics are optional
		zap.L().Warn("Failed to start database metrics collection", zap.Error(err))
	}

	urlRepo := repository.NewPostgresURLRepository(pgClient, redisClient)
	userRepo := repository.NewUserRepository(pgClient)
	urlService := service.NewURLService(urlRepo)
	authService := service.NewAuthService(userRepo)
	urlHandler := handler.NewURLHandler(urlService)
	authHandler := handler.NewAuthHandler(authService)

	// Permite /api e /api/ funcionarem igual
	r.RedirectTrailingSlash = true
	r.RemoveExtraSlash = true

	rateLimiter := middleware.NewRateLimiter(100, time.Minute)

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "go-backend"
	}

	r.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/health"))
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware(serviceName))

	// CORS configuration
	allowedOrigins := []string{"https://fonsecaaso.com"}
	// Allow all origins in local development
	if os.Getenv("ENV") == "local" {
		allowedOrigins = []string{"*"}
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", requestIDHeader, "Origin", "Accept"},
		ExposeHeaders:    []string{"Content-Length", requestIDHeader, "Cache-Hit"},
		AllowCredentials: os.Getenv("ENV") == "local", // Only allow credentials in local dev
		MaxAge:           12 * time.Hour,
	}))

	r.Use(requestIDMiddleware())
	r.Use(middleware.MetricsMiddleware())
	r.Use(loggingMiddleware())
	r.Use(rateLimiter.Middleware())

	// Healthz endpoint for observability status
	r.GET("/healthz", healthzCheck())

	// API
	api := r.Group("/api")

	api.GET("/health", healthCheck(redisClient, pgClient))

	// Metrics endpoint - use OTEL Prometheus exporter if available, otherwise use default handler
	if prometheusHandler != nil {
		api.GET("/metrics", gin.WrapH(prometheusHandler))
	} else {
		// Fallback: expose basic Go metrics
		api.GET("/metrics", func(c *gin.Context) {
			c.JSON(200, gin.H{"error": "metrics not configured"})
		})
	}

	api.POST("/", urlHandler.CreateTinyURL)
	api.POST("", urlHandler.CreateTinyURL)
	api.GET("/:id", urlHandler.GetURL)
	api.POST("/signup", authHandler.Register)
	api.POST("/login", authHandler.Login)

	// Rotas protegidas (requerem autenticação)
	protected := api.Group("/user")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/urls", urlHandler.GetUserURLs)
	}

	return r
}

const requestIDHeader = "X-Request-ID"

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header(requestIDHeader, requestID)
		c.Set("requestID", requestID)
		c.Next()
	}
}

func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetString("requestID")

		// Extract trace and span IDs from OpenTelemetry context
		span := trace.SpanFromContext(c.Request.Context())
		spanContext := span.SpanContext()
		traceID := spanContext.TraceID().String()
		spanID := spanContext.SpanID().String()

		// Create logger with trace context
		logger := zap.L().With(
			zap.String("request_id", requestID),
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)

		c.Set("logger", logger)
		c.Next()

		latency := time.Since(start)
		latencyMs := float64(latency.Milliseconds())
		status := c.Writer.Status()

		logger.Info("Request completed",
			zap.Int("status", status),
			zap.Float64("latency_ms", latencyMs),
			zap.Duration("latency", latency),
		)
	}
}

func healthCheck(redisClient *redis.Client, pgClient *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var redisOK bool
		if redisClient != nil {
			redisOK = redisClient.Ping(ctx).Err() == nil
		} else {
			redisOK = false
		}

		pgOK := pgClient.Ping(ctx) == nil

		status := "healthy"
		code := 200

		if !pgOK {
			status = "unhealthy"
			code = 503
		}

		c.JSON(code, gin.H{
			"status":    status,
			"redis":     redisOK,
			"postgres":  pgOK,
			"timestamp": time.Now().Unix(),
		})
	}
}

// func getAllowedOrigins() []string {
// 	// Default origins for development
// 	defaultOrigins := []string{"http://localhost:4200", "http://localhost:8080"}

// 	// Get production origins from environment variable
// 	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
// 	if corsOrigins == "" {
// 		return defaultOrigins
// 	}

// 	// Parse comma-separated origins
// 	origins := strings.Split(corsOrigins, ",")
// 	for i, origin := range origins {
// 		origins[i] = strings.TrimSpace(origin)
// 	}

// 	// Append development origins if in development mode
// 	if os.Getenv("GO_ENV") != "production" {
// 		origins = append(origins, defaultOrigins...)
// 	}

// 	return origins
// }

func healthzCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Check if observability components are configured and working
		otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		lokiEndpoint := os.Getenv("LOKI_ENDPOINT")
		if lokiEndpoint == "" {
			lokiEndpoint = os.Getenv("LOKI_URL")
		}

		otelTracingEnabled := otelEndpoint != ""
		otelMetricsEnabled := otelEndpoint != ""
		lokiLoggingEnabled := lokiEndpoint != ""

		// Try to emit a test trace span to verify tracing works
		tracingWorks := otelTracingEnabled
		if otelTracingEnabled {
			span := trace.SpanFromContext(ctx)
			tracingWorks = span.SpanContext().IsValid()
		}

		// Check if metrics are initialized
		metricsWorks := otelMetricsEnabled && metrics.IsInitialized()

		// Overall health status
		status := "ok"
		code := 200

		// If critical components are down, mark as degraded
		if otelTracingEnabled && !tracingWorks {
			status = "degraded"
			code = 200 // Still return 200 for ALB, but indicate degraded state
		}

		c.JSON(code, gin.H{
			"status":             status,
			"otel_tracing":       otelTracingEnabled,
			"otel_tracing_works": tracingWorks,
			"otel_metrics":       otelMetricsEnabled,
			"otel_metrics_works": metricsWorks,
			"loki_logging":       lokiLoggingEnabled,
			"service_name":       os.Getenv("SERVICE_NAME"),
			"environment":        os.Getenv("ENV"),
			"timestamp":          time.Now().Unix(),
		})
	}
}

func generateRequestID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 12
	id := make([]byte, length)

	for i := 0; i < length; i++ {
		id[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}

	return string(id)
}

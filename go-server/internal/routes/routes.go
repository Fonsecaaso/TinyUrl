package route

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/handler"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/metrics"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/middleware"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/repository"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/service"
)

func SetupRouter(redisClient *redis.Client, pgClient *pgxpool.Pool) *gin.Engine {
	r := gin.New()

	// Start system metrics collection
	metrics.StartSystemMetricsCollection()

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

	r.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/health"))
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("go-backend"))

	// CORS must be first to ensure all responses have proper headers
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", requestIDHeader, "Origin", "Accept"},
		ExposeHeaders:    []string{"Content-Length", requestIDHeader, "Cache-Hit"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(requestIDMiddleware())
	r.Use(middleware.MetricsMiddleware())
	r.Use(loggingMiddleware())
	r.Use(rateLimiter.Middleware())
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API
	api := r.Group("/api")

	api.GET("/health", healthCheck(redisClient, pgClient))
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

		logger := zap.L().With(
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)

		c.Set("logger", logger)
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("Request completed",
			zap.Int("status", status),
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

func generateRequestID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 12
	id := make([]byte, length)

	for i := 0; i < length; i++ {
		id[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}

	return string(id)
}

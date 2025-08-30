package route

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/middleware"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/service"
)

func SetupRouter(redisClient *redis.Client, pgClient *pgxpool.Pool) *gin.Engine {
	r := gin.New()

	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	
	r.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/health"))
	r.Use(gin.Recovery())
	r.Use(requestIDMiddleware())
	r.Use(loggingMiddleware())
	r.Use(rateLimiter.Middleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", requestIDHeader},
		ExposeHeaders:    []string{"Content-Length", requestIDHeader, "Cache-Hit"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", healthCheck(redisClient, pgClient))
	r.POST("/", func(c *gin.Context) { service.CreateTinyUrl(c, redisClient, pgClient) })
	r.GET("/:id", func(c *gin.Context) { service.GetUrl(c, redisClient, pgClient) })

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
		
		redisOK := redisClient.Ping(ctx).Err() == nil
		pgOK := pgClient.Ping(ctx) == nil
		
		status := "healthy"
		code := 200
		
		if !redisOK || !pgOK {
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

func generateRequestID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 12
	id := make([]byte, length)
	
	for i := 0; i < length; i++ {
		id[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	
	return string(id)
}

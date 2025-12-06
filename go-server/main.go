package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fonsecaaso/TinyUrl/go-server/config"
	db "github.com/fonsecaaso/TinyUrl/go-server/internal/database"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/logger"
	route "github.com/fonsecaaso/TinyUrl/go-server/internal/routes"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	// Load .env file FIRST (before initializing logger)
	// This ensures LOKI_URL and other env vars are available
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize Loki logger (sends logs directly to Loki)
	// Now it can read LOKI_URL from the environment
	if err := logger.InitLokiLogger("tinyurl-api", "development"); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer func() {
		_ = logger.Sync()
		_ = logger.Shutdown(ctx)
	}()

	// Replace global zap logger
	zap.ReplaceGlobals(logger.Logger)

	secrets, err := config.LoadConfig()
	if err != nil {
		logger.Logger.Fatal(
			"error loading configuration",
			zap.Error(err),
		)
	}

	pgClient, err := db.NewPostgresClient(secrets)
	if err != nil {
		logger.Logger.Fatal("postgres failed to initialize",
			zap.Error(err),
		)
	}
	logger.Logger.Info("postgres connection established")

	redisClient, err := db.NewRedisClient(secrets)
	if err != nil {
		logger.Logger.Warn("redis failed to initialize, continuing without cache",
			zap.Error(err),
		)
		redisClient = nil
	} else {
		logger.Logger.Info("redis connection established")
	}

	r := route.SetupRouter(redisClient, pgClient)
	logger.Logger.Info("starting server on :8080")

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := r.Run(":8080"); err != nil {
			logger.Logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	<-quit
	logger.Logger.Info("shutting down server...")
}

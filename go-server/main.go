package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fonsecaaso/TinyUrl/go-server/config"
	db "github.com/fonsecaaso/TinyUrl/go-server/internal/database"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/observability"
	route "github.com/fonsecaaso/TinyUrl/go-server/internal/routes"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize observability (tracing, metrics, logging)
	obs, err := observability.SetupObservability(ctx)
	if err != nil {
		log.Fatalf("failed to initialize observability: %v", err)
	}
	defer func() {
		if err := obs.Shutdown(ctx); err != nil {
			log.Printf("error shutting down observability: %v", err)
		}
	}()

	// Log observability initialization status
	status := obs.GetStatus()
	obs.Logger.Info("Observability initialized",
		zap.Bool("tracing_enabled", status.TracingEnabled),
		zap.Bool("metrics_enabled", status.MetricsEnabled),
		zap.Bool("logging_enabled", status.LoggingEnabled),
	)

	secrets, err := config.LoadConfig()
	if err != nil {
		obs.Logger.Fatal(
			"error loading configuration",
			zap.Error(err),
		)
	}

	pgClient, err := db.NewPostgresClient(secrets)
	if err != nil {
		obs.Logger.Fatal("postgres failed to initialize",
			zap.Error(err),
		)
	}
	obs.Logger.Info("postgres connection established")

	redisClient, err := db.NewRedisClient(secrets)
	if err != nil {
		obs.Logger.Warn("redis failed to initialize, continuing without cache",
			zap.Error(err),
		)
		redisClient = nil
	} else {
		obs.Logger.Info("redis connection established")
	}

	r := route.SetupRouter(redisClient, pgClient)
	obs.Logger.Info("starting server on :8080")

	// Create HTTP server with explicit configuration
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Channel to listen for shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			obs.Logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-quit
	obs.Logger.Info("shutting down server gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server gracefully
	if err := srv.Shutdown(shutdownCtx); err != nil {
		obs.Logger.Error("server forced to shutdown", zap.Error(err))
	}

	obs.Logger.Info("server stopped, flushing observability components...")
}

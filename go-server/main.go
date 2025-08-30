package main

import (
	"github.com/fonsecaaso/TinyUrl/go-server/config"
	db "github.com/fonsecaaso/TinyUrl/go-server/internal/database"
	route "github.com/fonsecaaso/TinyUrl/go-server/internal/routes"
	"go.uber.org/zap"
)

func main() {
	logger := zap.Must(zap.NewProduction())
	defer logger.Sync()
	
	zap.ReplaceGlobals(logger)

	secrets, err := config.LoadConfig()
	if err != nil {
		logger.Fatal(
			"error loading configuration",
			zap.Error(err),
		)
	}

	redisClient, err := db.NewRedisClient(secrets)
	if err != nil {
		logger.Fatal("redis failed to initialize",
			zap.Error(err),
		)
	}
	logger.Info("redis connection established")

	pgClient, err := db.NewPostgresClient(secrets)
	if err != nil {
		logger.Fatal("postgres failed to initialize",
			zap.Error(err),
		)
	}
	logger.Info("postgres connection established")

	r := route.SetupRouter(redisClient, pgClient)
	logger.Info("starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("server failed to start", zap.Error(err))
	}
}

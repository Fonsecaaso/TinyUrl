package main

import (
	db "github.com/fonsecaaso/TinyUrl/go-server/internal/database"
	route "github.com/fonsecaaso/TinyUrl/go-server/internal/routes"
	"go.uber.org/zap"
)

func main() {
	logger := zap.Must(zap.NewProduction())
	defer logger.Sync()

	redisClient, err := db.NewRedisClient()
	if err != nil {
		logger.Panic("redis failed to initialize",
			zap.Error(err),
		)
	} else {
		logger.Info("redis is connected")
	}

	pgClient, err := db.NewPostgresClient()
	if err != nil {
		logger.Panic("postgres failed to initialize",
			zap.Error(err),
		)
	} else {
		logger.Info("postgres is connected")
	}

	r := route.SetupRouter(redisClient, pgClient)
	r.Run(":8080")
}

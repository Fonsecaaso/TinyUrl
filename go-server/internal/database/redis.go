package database

import (
	"context"
	"fmt"

	config "github.com/fonsecaaso/TinyUrl/go-server/config"
	"github.com/go-redis/redis/v8"
)

func NewRedisClient(secrets *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     secrets.RedisAddr, // Endereço do Redis (alterar se necessário)
		Password: "",                // Sem senha (default)
		DB:       0,                 // Seleciona o banco de dados 0
	})

	// Health check: Testa a conexão com o Redis
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return rdb, nil
}

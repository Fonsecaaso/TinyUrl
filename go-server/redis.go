package main

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// Função para criar um cliente Redis
func newRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis-server:6379", // Endereço do Redis (alterar se necessário)
		Password: "",                  // Sem senha (default)
		DB:       0,                   // Seleciona o banco de dados 0
	})
	return rdb
}

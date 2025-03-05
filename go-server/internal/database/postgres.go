package database

import (
	"context"
	"fmt"

	config "github.com/fonsecaaso/TinyUrl/go-server/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresClient() (*pgxpool.Pool, error) {
	secrets, _ := config.LoadConfig()

	config, err := pgxpool.ParseConfig(secrets.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	ctx := context.Background()
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

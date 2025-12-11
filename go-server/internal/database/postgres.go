package database

import (
	"context"
	"fmt"
	"time"

	"github.com/exaring/otelpgx"

	config "github.com/fonsecaaso/TinyUrl/go-server/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresClient(secrets *config.Config) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(secrets.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute
	config.ConnConfig.ConnectTimeout = 10 * time.Second
	config.ConnConfig.Tracer = otelpgx.NewTracer()

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

package database

import (
	"context"
	"fmt"
	"time"

	config "github.com/fonsecaaso/TinyUrl/go-server/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresClient(secrets *config.Config) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(secrets.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Optimize connection pool parameters for performance
	// MaxConns: Maximum number of connections in the pool
	config.MaxConns = 25

	// MinConns: Minimum number of connections to maintain (reduces connection overhead)
	config.MinConns = 5

	// MaxConnLifetime: Maximum lifetime of a connection (prevent stale connections)
	config.MaxConnLifetime = 1 * time.Hour

	// MaxConnIdleTime: Close idle connections after this duration (save resources)
	config.MaxConnIdleTime = 30 * time.Minute

	// HealthCheckPeriod: How often to check connection health
	config.HealthCheckPeriod = 1 * time.Minute

	// ConnectTimeout: Maximum time to wait for a new connection
	config.ConnConfig.ConnectTimeout = 10 * time.Second

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

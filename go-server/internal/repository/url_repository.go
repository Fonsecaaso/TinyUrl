package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	ErrURLNotFound   = errors.New("URL not found")
	ErrDatabaseError = errors.New("database error")
	ErrCacheError    = errors.New("cache error")
)

const (
	cacheTimeout = 24 * time.Hour
	dbTimeout    = 5 * time.Second
)

// URLRepository defines the interface for URL data operations
type URLRepository interface {
	Create(ctx context.Context, id, url string) error
	CreateOrGet(ctx context.Context, id, url string) (shortCode string, isNew bool, err error)
	FindByID(ctx context.Context, id string) (string, error)
	FindByURL(ctx context.Context, url string) (string, error)
	IDExists(ctx context.Context, id string) (bool, error)
}

// PostgresURLRepository implements URLRepository using PostgreSQL
type PostgresURLRepository struct {
	db          *pgxpool.Pool
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewPostgresURLRepository creates a new PostgresURLRepository
func NewPostgresURLRepository(db *pgxpool.Pool, redisClient *redis.Client) *PostgresURLRepository {
	return &PostgresURLRepository{
		db:          db,
		redisClient: redisClient,
		logger:      zap.L().With(zap.String("component", "PostgresURLRepository")),
	}
}

// Create inserts a new URL mapping into the database
func (r *PostgresURLRepository) Create(ctx context.Context, id, url string) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	_, err := r.db.Exec(ctx, "INSERT INTO urls (id, url) VALUES ($1, $2)", id, url)
	if err != nil {
		r.logger.Error("Failed to insert URL", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return nil
}

// CreateOrGet atomically checks if URL exists and returns existing short code, or creates new one
func (r *PostgresURLRepository) CreateOrGet(ctx context.Context, id, url string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	// Start a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.Error("Failed to start transaction", zap.Error(err))
		return "", false, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	defer tx.Rollback(ctx)

	// First, try to find existing URL
	var existingID string
	err = tx.QueryRow(ctx, "SELECT id FROM urls WHERE url = $1 LIMIT 1", url).Scan(&existingID)

	if err == nil {
		// URL already exists, return existing short code
		if err := tx.Commit(ctx); err != nil {
			r.logger.Error("Failed to commit transaction", zap.Error(err))
			return "", false, fmt.Errorf("%w: %v", ErrDatabaseError, err)
		}
		r.logger.Info("URL already exists, returning existing short code",
			zap.String("id", existingID),
			zap.String("url", url))
		return existingID, false, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		// Real error occurred
		r.logger.Error("Database query error", zap.Error(err), zap.String("url", url))
		return "", false, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// URL doesn't exist, insert new record
	_, err = tx.Exec(ctx, "INSERT INTO urls (id, url) VALUES ($1, $2)", id, url)
	if err != nil {
		r.logger.Error("Failed to insert URL", zap.Error(err), zap.String("id", id))
		return "", false, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return "", false, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	r.logger.Info("New URL created", zap.String("id", id), zap.String("url", url))
	return id, true, nil
}

// FindByID retrieves a URL by its short code, checking cache first
func (r *PostgresURLRepository) FindByID(ctx context.Context, id string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	// Try cache first if Redis is available
	if r.redisClient != nil {
		val, err := r.redisClient.Get(ctx, id).Result()
		if err == nil {
			r.logger.Debug("URL found in cache", zap.String("id", id))
			return val, nil
		}

		if err != redis.Nil {
			r.logger.Warn("Cache error", zap.Error(err), zap.String("id", id))
		}
	}

	// Query database
	var url string
	err := r.db.QueryRow(ctx, "SELECT url FROM urls WHERE id = $1", id).Scan(&url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("URL not found", zap.String("id", id))
			return "", ErrURLNotFound
		}
		r.logger.Error("Database query error", zap.Error(err), zap.String("id", id))
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Cache the result if Redis is available
	if r.redisClient != nil {
		if err := r.redisClient.Set(ctx, id, url, cacheTimeout).Err(); err != nil {
			r.logger.Warn("Failed to cache URL", zap.Error(err), zap.String("id", id))
		}
	}

	r.logger.Debug("URL found in database", zap.String("id", id))
	return url, nil
}

// FindByURL retrieves the short code for a given URL
func (r *PostgresURLRepository) FindByURL(ctx context.Context, url string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var id string
	err := r.db.QueryRow(ctx, "SELECT id FROM urls WHERE url = $1 LIMIT 1", url).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("URL not found", zap.String("url", url))
			return "", ErrURLNotFound
		}
		r.logger.Error("Database query error", zap.Error(err), zap.String("url", url))
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	r.logger.Debug("Short code found for URL", zap.String("id", id), zap.String("url", url))
	return id, nil
}

// IDExists checks if a given ID already exists in the database
func (r *PostgresURLRepository) IDExists(ctx context.Context, id string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM urls WHERE id = $1", id).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to check ID existence", zap.Error(err), zap.String("id", id))
		return false, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return count > 0, nil
}

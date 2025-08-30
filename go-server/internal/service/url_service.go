package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Body struct {
	URL string `json:"url" binding:"required"`
}

type URLResponse struct {
	Message   string `json:"message"`
	ShortCode string `json:"short_code,omitempty"`
	URL       string `json:"url,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

var (
	ErrInvalidURL      = errors.New("invalid URL format")
	ErrURLNotFound     = errors.New("URL not found")
	ErrDatabaseError   = errors.New("database error")
	ErrCacheError      = errors.New("cache error")
	ErrIDGenerationMax = errors.New("failed to generate unique ID after max attempts")
)

const (
	maxIDGenerationAttempts = 10
	cacheTimeout           = 24 * time.Hour
	dbTimeout              = 5 * time.Second
)

func CreateTinyUrl(c *gin.Context, redisClient *redis.Client, pgClient *pgxpool.Pool) {
	logger := zap.L().With(zap.String("handler", "CreateTinyUrl"))
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	body := Body{}
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request format",
			Code:  "INVALID_JSON",
		})
		return
	}

	if !isValidURL(body.URL) {
		logger.Warn("Invalid URL provided", zap.String("url", body.URL))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid URL format",
			Code:  "INVALID_URL",
		})
		return
	}

	key, err := generateUniqueID(ctx, pgClient)
	if err != nil {
		logger.Error("Failed to generate unique ID", zap.Error(err))
		if errors.Is(err, ErrIDGenerationMax) {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Service temporarily unavailable",
				Code:  "ID_GENERATION_FAILED",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Database error",
				Code:  "DB_ERROR",
			})
		}
		return
	}

	_, err = pgClient.Exec(ctx, "INSERT INTO urls (id, url) VALUES ($1, $2)", key, body.URL)
	if err != nil {
		logger.Error("Failed to store URL", zap.Error(err), zap.String("id", key))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to store URL",
			Code:  "STORAGE_ERROR",
		})
		return
	}

	logger.Info("URL shortened successfully", zap.String("id", key), zap.String("url", body.URL))
	c.JSON(http.StatusCreated, URLResponse{
		Message:   "URL shortened successfully",
		ShortCode: key,
	})
}

func GetUrl(c *gin.Context, redisClient *redis.Client, pgClient *pgxpool.Pool) {
	logger := zap.L().With(zap.String("handler", "GetUrl"))
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "ID parameter is required",
			Code:  "MISSING_ID",
		})
		return
	}

	if !isValidID(id) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid ID format",
			Code:  "INVALID_ID",
		})
		return
	}

	val, err := redisClient.Get(ctx, id).Result()
	if err == nil {
		logger.Info("URL found in cache", zap.String("id", id))
		c.Header("Cache-Hit", "true")
		c.JSON(http.StatusOK, URLResponse{
			Message: "URL retrieved successfully",
			URL:     val,
		})
		return
	}

	if err != redis.Nil {
		logger.Warn("Cache error", zap.Error(err), zap.String("id", id))
	}

	var url string
	err = pgClient.QueryRow(ctx, "SELECT url FROM urls WHERE id = $1", id).Scan(&url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Info("URL not found", zap.String("id", id))
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Short URL not found",
				Code:  "URL_NOT_FOUND",
			})
			return
		}
		logger.Error("Database query error", zap.Error(err), zap.String("id", id))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Database error",
			Code:  "DB_ERROR",
		})
		return
	}

	if err := redisClient.Set(ctx, id, url, cacheTimeout).Err(); err != nil {
		logger.Warn("Failed to cache URL", zap.Error(err), zap.String("id", id))
	}

	logger.Info("URL found in database", zap.String("id", id))
	c.Header("Cache-Hit", "false")
	c.JSON(http.StatusOK, URLResponse{
		Message: "URL retrieved successfully",
		URL:     url,
	})
}

func generateUniqueID(ctx context.Context, pgClient *pgxpool.Pool) (string, error) {
	for attempt := 0; attempt < maxIDGenerationAttempts; attempt++ {
		id := createID()
		var count int
		err := pgClient.QueryRow(ctx, "SELECT COUNT(*) FROM urls WHERE id = $1", id).Scan(&count)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
		}
		if count == 0 {
			return id, nil
		}
	}
	return "", ErrIDGenerationMax
}

func createID() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const idLength = 6
	id := make([]byte, idLength)

	for i := 0; i < idLength; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			panic(fmt.Sprintf("failed to generate random number: %v", err))
		}
		id[i] = chars[randomIndex.Int64()]
	}

	return string(id)
}

func isValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}

	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	return parsed.Scheme != "" && parsed.Host != ""
}

func isValidID(id string) bool {
	if len(id) != 6 {
		return false
	}
	for _, char := range id {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	return true
}

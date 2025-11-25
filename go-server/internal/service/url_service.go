package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/repository"

	"go.uber.org/zap"
)

var (
	ErrInvalidURL      = errors.New("invalid URL format")
	ErrIDGenerationMax = errors.New("failed to generate unique ID after max attempts")
)

const (
	maxIDGenerationAttempts = 10
	idLength                = 6
)

// URLService handles business logic for URL operations
type URLService struct {
	repo   repository.URLRepository
	logger *zap.Logger
}

// NewURLService creates a new URLService
func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{
		repo:   repo,
		logger: zap.L().With(zap.String("component", "URLService")),
	}
}

// ShortenURL creates a short URL for the given URL or returns existing one
// Returns (shortCode, isNew, error) where isNew indicates if a new URL was created
func (s *URLService) ShortenURL(ctx context.Context, rawURL string) (string, bool, error) {
	// Validate URL
	if !s.isValidURL(rawURL) {
		s.logger.Warn("Invalid URL provided", zap.String("url", rawURL))
		return "", false, ErrInvalidURL
	}

	// Normalize URL (add https:// if missing)
	normalizedURL := rawURL
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		normalizedURL = "https://" + rawURL
	}

	// Generate unique ID
	shortCode, err := s.generateUniqueID(ctx)
	if err != nil {
		s.logger.Error("Failed to generate unique ID", zap.Error(err))
		return "", false, err
	}

	// Try to create or get existing URL
	resultCode, isNew, err := s.repo.CreateOrGet(ctx, shortCode, normalizedURL)
	if err != nil {
		s.logger.Error("Failed to store URL", zap.Error(err), zap.String("id", shortCode))
		return "", false, err
	}

	if isNew {
		s.logger.Info("URL shortened successfully", zap.String("id", resultCode), zap.String("url", normalizedURL))
	} else {
		s.logger.Info("URL already exists, returning existing short code", zap.String("id", resultCode), zap.String("url", normalizedURL))
	}

	return resultCode, isNew, nil
}

// GetOriginalURL retrieves the original URL for a given short code
func (s *URLService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	// Validate short code format
	if !s.isValidID(shortCode) {
		s.logger.Warn("Invalid short code format", zap.String("shortCode", shortCode))
		return "", errors.New("invalid short code format")
	}

	// Retrieve URL from repository
	url, err := s.repo.FindByID(ctx, shortCode)
	if err != nil {
		if errors.Is(err, repository.ErrURLNotFound) {
			s.logger.Info("URL not found", zap.String("shortCode", shortCode))
			return "", repository.ErrURLNotFound
		}
		s.logger.Error("Failed to retrieve URL", zap.Error(err), zap.String("shortCode", shortCode))
		return "", err
	}

	s.logger.Info("URL retrieved successfully", zap.String("shortCode", shortCode))
	return url, nil
}

// generateUniqueID generates a unique short code
func (s *URLService) generateUniqueID(ctx context.Context) (string, error) {
	for attempt := 0; attempt < maxIDGenerationAttempts; attempt++ {
		id := s.createID()
		exists, err := s.repo.IDExists(ctx, id)
		if err != nil {
			return "", err
		}
		if !exists {
			return id, nil
		}
	}
	return "", ErrIDGenerationMax
}

// createID generates a random alphanumeric ID
func (s *URLService) createID() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
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

// isValidURL validates URL format
func (s *URLService) isValidURL(rawURL string) bool {
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

// isValidID validates short code format
func (s *URLService) isValidID(id string) bool {
	if len(id) != idLength {
		return false
	}
	for _, char := range id {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') {
			return false
		}
	}
	return true
}

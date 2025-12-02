package service

import (
	"context"
	"errors"
	"testing"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/model"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockURLRepository is a mock implementation of URLRepository
type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) Create(ctx context.Context, url *model.URL) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *MockURLRepository) CreateOrGet(ctx context.Context, url *model.URL) (string, bool, error) {
	args := m.Called(ctx, url)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockURLRepository) FindByID(ctx context.Context, id string) (*model.URL, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URL), args.Error(1)
}

func (m *MockURLRepository) FindByURL(ctx context.Context, url string) (string, error) {
	args := m.Called(ctx, url)
	return args.String(0), args.Error(1)
}

func (m *MockURLRepository) IDExists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockURLRepository) GetUserURLs(ctx context.Context, userId uuid.UUID) ([]model.URL, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.URL), args.Error(1)
}

func setupService(t *testing.T) (*URLService, *MockURLRepository) {
	// Initialize logger for tests
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	mockRepo := new(MockURLRepository)
	service := NewURLService(mockRepo)

	return service, mockRepo
}

func TestNewURLService(t *testing.T) {
	mockRepo := new(MockURLRepository)
	service := NewURLService(mockRepo)

	assert.NotNil(t, service)
	assert.NotNil(t, service.repo)
	assert.NotNil(t, service.logger)
}

func TestShortenURL_Success_NewURL(t *testing.T) {
	service, mockRepo := setupService(t)
	ctx := context.Background()

	testURL := "https://example.com"
	expectedShortCode := "abc123"

	// Mock ID generation: first call says ID doesn't exist
	mockRepo.On("IDExists", ctx, mock.AnythingOfType("string")).Return(false, nil)

	// Mock CreateOrGet: returns new URL
	mockRepo.On("CreateOrGet", ctx, mock.AnythingOfType("*model.URL")).
		Return(expectedShortCode, true, nil)

	shortCode, isNew, err := service.ShortenURL(ctx, testURL, nil)

	assert.NoError(t, err)
	assert.Equal(t, expectedShortCode, shortCode)
	assert.True(t, isNew)
	mockRepo.AssertExpectations(t)
}

func TestShortenURL_Success_ExistingURL(t *testing.T) {
	service, mockRepo := setupService(t)
	ctx := context.Background()

	testURL := "https://example.com"
	existingShortCode := "xyz789"

	// Mock ID generation
	mockRepo.On("IDExists", ctx, mock.AnythingOfType("string")).Return(false, nil)

	// Mock CreateOrGet: returns existing URL
	mockRepo.On("CreateOrGet", ctx, mock.AnythingOfType("*model.URL")).
		Return(existingShortCode, false, nil)

	shortCode, isNew, err := service.ShortenURL(ctx, testURL, nil)

	assert.NoError(t, err)
	assert.Equal(t, existingShortCode, shortCode)
	assert.False(t, isNew)
	mockRepo.AssertExpectations(t)
}

func TestShortenURL_InvalidURL(t *testing.T) {
	service, _ := setupService(t)
	ctx := context.Background()

	testCases := []struct {
		name string
		url  string
	}{
		{"empty URL", ""},
		{"invalid format", "not a valid url"},
		{"missing host", "http://"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := service.ShortenURL(ctx, tc.url, nil)
			assert.ErrorIs(t, err, ErrInvalidURL)
		})
	}
}

func TestShortenURL_URLNormalization(t *testing.T) {
	service, mockRepo := setupService(t)
	ctx := context.Background()

	testURL := "example.com"
	expectedShortCode := "abc123"

	mockRepo.On("IDExists", ctx, mock.AnythingOfType("string")).Return(false, nil)
	mockRepo.On("CreateOrGet", ctx, mock.AnythingOfType("*model.URL")).
		Return(expectedShortCode, true, nil)

	shortCode, isNew, err := service.ShortenURL(ctx, testURL, nil)

	assert.NoError(t, err)
	assert.Equal(t, expectedShortCode, shortCode)
	assert.True(t, isNew)
	mockRepo.AssertExpectations(t)
}

func TestShortenURL_IDGenerationFailure(t *testing.T) {
	service, mockRepo := setupService(t)
	ctx := context.Background()

	testURL := "https://example.com"

	// Mock ID generation: all IDs already exist
	mockRepo.On("IDExists", ctx, mock.AnythingOfType("string")).
		Return(true, nil).Times(maxIDGenerationAttempts)

	_, _, err := service.ShortenURL(ctx, testURL, nil)

	assert.ErrorIs(t, err, ErrIDGenerationMax)
	mockRepo.AssertExpectations(t)
}

func TestShortenURL_RepositoryError(t *testing.T) {
	service, mockRepo := setupService(t)
	ctx := context.Background()

	testURL := "https://example.com"
	dbError := errors.New("database connection failed")

	mockRepo.On("IDExists", ctx, mock.AnythingOfType("string")).Return(false, nil)
	mockRepo.On("CreateOrGet", ctx, mock.AnythingOfType("*model.URL")).
		Return("", false, dbError)

	_, _, err := service.ShortenURL(ctx, testURL, nil)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)
	mockRepo.AssertExpectations(t)
}

func TestGetOriginalURL_Success(t *testing.T) {
	service, mockRepo := setupService(t)
	ctx := context.Background()

	shortCode := "abc123"
	expectedURL := "https://example.com"

	urlModel := &model.URL{
		ID:          shortCode,
		OriginalURL: expectedURL,
	}
	mockRepo.On("FindByID", ctx, shortCode).Return(urlModel, nil)

	url, err := service.GetOriginalURL(ctx, shortCode)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, url)
	mockRepo.AssertExpectations(t)
}

func TestGetOriginalURL_InvalidShortCode(t *testing.T) {
	service, _ := setupService(t)
	ctx := context.Background()

	testCases := []struct {
		name      string
		shortCode string
	}{
		{"too short", "abc"},
		{"too long", "abcdefg"},
		{"invalid characters", "abc!@#"},
		{"with spaces", "abc 123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.GetOriginalURL(ctx, tc.shortCode)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid short code format")
		})
	}
}

func TestGetOriginalURL_NotFound(t *testing.T) {
	service, mockRepo := setupService(t)
	ctx := context.Background()

	shortCode := "abc123"

	mockRepo.On("FindByID", ctx, shortCode).Return(nil, repository.ErrURLNotFound)

	_, err := service.GetOriginalURL(ctx, shortCode)

	assert.ErrorIs(t, err, repository.ErrURLNotFound)
	mockRepo.AssertExpectations(t)
}

func TestGetOriginalURL_RepositoryError(t *testing.T) {
	service, mockRepo := setupService(t)
	ctx := context.Background()

	shortCode := "abc123"
	dbError := errors.New("database connection failed")

	mockRepo.On("FindByID", ctx, shortCode).Return(nil, dbError)

	_, err := service.GetOriginalURL(ctx, shortCode)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)
	mockRepo.AssertExpectations(t)
}

func TestIsValidURL(t *testing.T) {
	service, _ := setupService(t)

	testCases := []struct {
		name     string
		url      string
		expected bool
	}{
		{"valid HTTPS URL", "https://example.com", true},
		{"valid HTTP URL", "http://example.com", true},
		{"valid URL without protocol", "example.com", true},
		{"valid URL with path", "https://example.com/path", true},
		{"valid URL with query", "https://example.com?query=param", true},
		{"valid subdomain", "https://sub.example.com", true},
		{"empty URL", "", false},
		{"invalid format", "not a url", false},
		{"only protocol", "https://", false},
		{"only slashes", "//", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.isValidURL(tc.url)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsValidID(t *testing.T) {
	service, _ := setupService(t)

	testCases := []struct {
		name     string
		id       string
		expected bool
	}{
		{"valid lowercase", "abcdef", true},
		{"valid uppercase", "ABCDEF", true},
		{"valid mixed case", "AbCdEf", true},
		{"valid with numbers", "abc123", true},
		{"valid all numbers", "123456", true},
		{"too short", "abc", false},
		{"too long", "abcdefg", false},
		{"with special chars", "abc!@#", false},
		{"with spaces", "abc 123", false},
		{"with dash", "abc-123", false},
		{"with underscore", "abc_123", false},
		{"empty string", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.isValidID(tc.id)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCreateID(t *testing.T) {
	service, _ := setupService(t)

	// Test that createID generates IDs of correct length
	for i := 0; i < 10; i++ {
		id := service.createID()
		assert.Len(t, id, idLength)

		// Verify all characters are alphanumeric
		for _, char := range id {
			assert.True(t,
				(char >= 'a' && char <= 'z') ||
					(char >= 'A' && char <= 'Z') ||
					(char >= '0' && char <= '9'),
				"ID contains invalid character: %c", char,
			)
		}
	}

	// Test that createID generates different IDs (with high probability)
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := service.createID()
		ids[id] = true
	}

	// We should have generated many unique IDs
	assert.Greater(t, len(ids), 90, "createID should generate mostly unique IDs")
}

func TestGenerateUniqueID_Success(t *testing.T) {
	service, mockRepo := setupService(t)
	ctx := context.Background()

	// First ID check returns false (ID doesn't exist)
	mockRepo.On("IDExists", ctx, mock.AnythingOfType("string")).Return(false, nil).Once()

	id, err := service.generateUniqueID(ctx)

	assert.NoError(t, err)
	assert.Len(t, id, idLength)
	mockRepo.AssertExpectations(t)
}

func TestGenerateUniqueID_RetryAndSuccess(t *testing.T) {
	service, mockRepo := setupService(t)
	ctx := context.Background()

	// First 3 attempts return true (ID exists), 4th returns false
	mockRepo.On("IDExists", ctx, mock.AnythingOfType("string")).Return(true, nil).Times(3)
	mockRepo.On("IDExists", ctx, mock.AnythingOfType("string")).Return(false, nil).Once()

	id, err := service.generateUniqueID(ctx)

	assert.NoError(t, err)
	assert.Len(t, id, idLength)
	mockRepo.AssertExpectations(t)
}

func TestGenerateUniqueID_MaxAttemptsReached(t *testing.T) {
	service, mockRepo := setupService(t)
	ctx := context.Background()

	// All attempts return true (ID exists)
	mockRepo.On("IDExists", ctx, mock.AnythingOfType("string")).
		Return(true, nil).Times(maxIDGenerationAttempts)

	_, err := service.generateUniqueID(ctx)

	assert.ErrorIs(t, err, ErrIDGenerationMax)
	mockRepo.AssertExpectations(t)
}

func TestGenerateUniqueID_RepositoryError(t *testing.T) {
	service, mockRepo := setupService(t)
	ctx := context.Background()

	dbError := errors.New("database connection failed")
	mockRepo.On("IDExists", ctx, mock.AnythingOfType("string")).Return(false, dbError)

	_, err := service.generateUniqueID(ctx)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)
	mockRepo.AssertExpectations(t)
}

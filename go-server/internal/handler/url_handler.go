package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/middleware"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/repository"
	"github.com/fonsecaaso/TinyUrl/go-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CreateURLRequest struct {
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

type URLHandler struct {
	service *service.URLService
	logger  *zap.Logger
}

func NewURLHandler(service *service.URLService) *URLHandler {
	return &URLHandler{
		service: service,
		logger:  zap.L().With(zap.String("component", "URLHandler")),
	}
}

func (h *URLHandler) CreateTinyURL(c *gin.Context) {
	var req CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request format",
			Code:  "INVALID_JSON",
		})
		return
	}

	shortCode, isNew, err := h.service.ShortenURL(c.Request.Context(), req.URL)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if isNew {
		c.JSON(http.StatusCreated, URLResponse{
			Message:   "URL shortened successfully",
			ShortCode: shortCode,
		})
	} else {
		c.JSON(http.StatusOK, URLResponse{
			Message:   "URL already exists, returning existing short code",
			ShortCode: shortCode,
		})
	}
}

func (h *URLHandler) GetURL(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "ID parameter is required",
			Code:  "MISSING_ID",
		})
		return
	}

	url, err := h.service.GetOriginalURL(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, URLResponse{
		Message: "URL retrieved successfully",
		URL:     url,
	})
}

func (h *URLHandler) GetUserURLs(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		h.logger.Warn("Failed to extract user ID from context", zap.Error(err))
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Unauthorized access",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	urls, err := h.service.GetUserURLs(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User URLs retrieved successfully",
		"urls":    urls,
	})
}

func (h *URLHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidURL):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid URL format",
			Code:  "INVALID_URL",
		})
	case errors.Is(err, repository.ErrURLNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Short URL not found",
			Code:  "URL_NOT_FOUND",
		})
	case errors.Is(err, service.ErrIDGenerationMax):
		h.logger.Error("ID generation max attempts reached", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Service temporarily unavailable",
			Code:  "ID_GENERATION_FAILED",
		})
	case errors.Is(err, repository.ErrDatabaseError):
		h.logger.Error("Database error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Database error",
			Code:  "DB_ERROR",
		})
	default:
		h.logger.Error("Unexpected error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Internal server error",
			Code:  "INTERNAL_ERROR",
		})
	}
}

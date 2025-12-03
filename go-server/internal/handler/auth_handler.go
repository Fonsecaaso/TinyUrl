package handler

import (
	"errors"
	"net/http"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	svc    service.AuthService
	logger *zap.Logger
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{
		svc:    svc,
		logger: zap.L().Named("AuthHandler"),
	}
}

// DTOs
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid JSON in Register", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request payload",
			Code:  "INVALID_PAYLOAD",
		})
		return
	}

	user, err := h.svc.Register(c, req.Username, req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid JSON in Login", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request payload",
			Code:  "INVALID_PAYLOAD",
		})
		return
	}

	token, err := h.svc.Login(c, req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	switch {
	// case errors.Is(err, service.ErrInvalidCredentials):
	// 	c.JSON(http.StatusUnauthorized, ErrorResponse{
	// 		Error: "Invalid email or password",
	// 		Code:  "INVALID_CREDENTIALS",
	// 	})
	// case errors.Is(err, service.ErrEmailAlreadyExists):
	// 	c.JSON(http.StatusConflict, ErrorResponse{
	// 		Error: "Email already registered",
	// 		Code:  "EMAIL_EXISTS",
	// 	})
	case errors.Is(err, service.ErrInvalidURL):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid URL format",
			Code:  "INVALID_URL",
		})
	default:
		h.logger.Error("Unexpected service error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Internal server error",
			Code:  "INTERNAL_ERROR",
		})
	}
}

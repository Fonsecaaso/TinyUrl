package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/fonsecaaso/TinyUrl/go-server/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	userIDKey = "user_id"
	claimsKey = "claims"
)

var (
	ErrMissingToken     = errors.New("missing authorization header")
	ErrInvalidToken     = errors.New("invalid or expired token")
	ErrMissingUserID    = errors.New("user_id not found in context")
	ErrInvalidUserIDKey = errors.New("user_id in context is not a valid UUID")
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": ErrMissingToken.Error(),
				"code":  "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		claims, err := token.ValidateToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": ErrInvalidToken.Error(),
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		c.Set(claimsKey, claims)
		c.Set(userIDKey, claims.UserID)

		c.Next()
	}
}

func GetUserIDFromContext(c *gin.Context) *uuid.UUID {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil
	}

	tokenStr := strings.TrimPrefix(auth, "Bearer ")

	claims, err := token.ValidateToken(tokenStr)
	if err != nil {
		return nil
	}

	return claims.UserID
}

func GetClaimsFromContext(c *gin.Context) (*token.CustomClaims, error) {
	value, exists := c.Get(claimsKey)
	if !exists {
		return nil, errors.New("claims not found in context")
	}

	claims, ok := value.(*token.CustomClaims)
	if !ok {
		return nil, errors.New("invalid claims type in context")
	}

	return claims, nil
}

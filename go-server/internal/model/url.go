package model

import (
	"time"

	"github.com/google/uuid"
)

// URL represents a shortened URL entry in the system
type URL struct {
	ID          string     `json:"id" db:"id"`
	OriginalURL string     `json:"url" db:"url" validate:"required,url"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UserID      *uuid.UUID `json:"user_id,omitempty" db:"user_id,omitempty"`
}

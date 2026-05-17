package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents an admin user.
type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // never expose
	CreatedAt    time.Time `json:"created_at"`
}

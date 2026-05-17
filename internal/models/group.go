package models

import (
	"time"

	"github.com/google/uuid"
)

// Group represents a logical collection of devices.
type Group struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

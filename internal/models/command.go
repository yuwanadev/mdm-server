package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CommandStatus represents the lifecycle state of a command.
type CommandStatus string

const (
	CommandStatusPending CommandStatus = "PENDING"
	CommandStatusSent    CommandStatus = "SENT"
	CommandStatusSuccess CommandStatus = "SUCCESS"
	CommandStatusFailed  CommandStatus = "FAILED"
)

// CommandLog records a command sent to a device.
type CommandLog struct {
	ID          uuid.UUID       `json:"id"`
	DeviceID    uuid.UUID       `json:"device_id"`
	CommandType string          `json:"command_type"`
	Payload     json.RawMessage `json:"payload,omitempty"`
	Status      CommandStatus   `json:"status"`
	Message     *string         `json:"message,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

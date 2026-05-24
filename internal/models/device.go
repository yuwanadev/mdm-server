package models

import (
	"time"

	"github.com/google/uuid"
)

// Device represents a registered Android device.
type Device struct {
	ID             uuid.UUID  `json:"id"`
	DeviceName     string     `json:"device_name"`
	Label          string     `json:"label"`
	Notes          string     `json:"notes"`
	GroupID        *uuid.UUID `json:"group_id,omitempty"`
	DeviceModel    *string    `json:"device_model,omitempty"`
	AndroidVersion *string    `json:"android_version,omitempty"`
	AgentVersion   *string    `json:"agent_version,omitempty"`
	TokenHash      string     `json:"-"` // never expose
	IsOnline       bool       `json:"is_online"`
	LastSeen       *time.Time `json:"last_seen,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// DeviceWithStatus is a joined view for list endpoints.
type DeviceWithStatus struct {
	Device
	Status *DeviceStatus `json:"status,omitempty"`
}

// DeviceAccount represents an account registered on the device.
type DeviceAccount struct {
	ID          uuid.UUID `json:"id"`
	DeviceID    uuid.UUID `json:"device_id"`
	AccountName string    `json:"account_name"`
	AccountType string    `json:"account_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DeviceAccountWithDevice is a joined view for the global list of accounts.
type DeviceAccountWithDevice struct {
	DeviceAccount
	DeviceName  string  `json:"device_name"`
	DeviceModel *string `json:"device_model,omitempty"`
}

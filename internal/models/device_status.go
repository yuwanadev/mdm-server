package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DeviceStatus holds the latest status snapshot for a device.
type DeviceStatus struct {
	DeviceID      uuid.UUID       `json:"device_id"`
	Battery           *int            `json:"battery,omitempty"`
	Temperature       *float32        `json:"temperature,omitempty"`
	BatteryHealth     *string         `json:"battery_health,omitempty"`
	BatteryStatus     *string         `json:"battery_status,omitempty"`
	BatteryTechnology *string         `json:"battery_technology,omitempty"`
	BatteryVoltage    *int            `json:"battery_voltage,omitempty"`
	RAMUsage          *int            `json:"ram_usage,omitempty"`     // MB
	StorageTotal      *int            `json:"storage_total,omitempty"` // MB
	StorageUsed       *int            `json:"storage_used,omitempty"`  // MB
	AppVersion        *string         `json:"app_version,omitempty"`
	NetworkInfo       json.RawMessage `json:"network_info,omitempty"`
	ForegroundApp     *string         `json:"foreground_app,omitempty"`
	NetworkStrength   *int            `json:"network_strength,omitempty"`
	Location          json.RawMessage `json:"location,omitempty"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

package websocket

import (
	"encoding/json"
	"time"
)

// WSMessage is the standard message envelope for all WebSocket communication.
type WSMessage struct {
	Type      string          `json:"type"`
	DeviceID  string          `json:"device_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
}

// NewMessage creates a new WSMessage with the current timestamp.
func NewMessage(msgType string, payload interface{}) (*WSMessage, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &WSMessage{
		Type:      msgType,
		Payload:   data,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}

// NewMessageRaw creates a new WSMessage with raw JSON payload.
func NewMessageRaw(msgType string, payload json.RawMessage) *WSMessage {
	return &WSMessage{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

// Message types: Device → Server
const (
	MsgHeartbeat     = "STATUS_UPDATE"
	MsgCommandResult = "COMMAND_RESULT"
	MsgDeviceInfo    = "DEVICE_INFO"
	MsgMirrorFrame   = "MIRROR_FRAME"
)

// Message types: Server → Device
const (
	MsgPing         = "PING"
	MsgCommand      = "COMMAND"
	MsgAck          = "ACK"
	MsgStartMirror  = "START_MIRROR"
	MsgStopMirror   = "STOP_MIRROR"
	MsgTouchEvent   = "TOUCH_EVENT"
	MsgWebRTCSignal = "WEBRTC_SIGNAL"
)

// Message types: Server → Dashboard
const (
	MsgDeviceOnline  = "DEVICE_ONLINE"
	MsgDeviceOffline = "DEVICE_OFFLINE"
	MsgStatusUpdate  = "STATUS_UPDATE"
)

// HeartbeatPayload is sent by the agent every 30s.
type HeartbeatPayload struct {
	Battery           *int            `json:"battery,omitempty"`
	Temperature       *float32        `json:"temperature,omitempty"`
	BatteryHealth     *string         `json:"battery_health,omitempty"`
	BatteryStatus     *string         `json:"battery_status,omitempty"`
	BatteryTechnology *string         `json:"battery_technology,omitempty"`
	BatteryVoltage    *int            `json:"battery_voltage,omitempty"`
	RAMUsage          *int            `json:"ram_usage,omitempty"`
	StorageTotal      *int            `json:"storage_total,omitempty"`
	StorageUsed       *int            `json:"storage_used,omitempty"`
	AppVersion        *string         `json:"app_version,omitempty"`
	NetworkInfo       json.RawMessage `json:"network_info,omitempty"`
	ForegroundApp     *string         `json:"foreground_app,omitempty"`
	NetworkStrength   *int            `json:"network_strength,omitempty"`
	Location          json.RawMessage `json:"location,omitempty"`
}

// DeviceInfoPayload is sent by the agent on connect and on request.
type DeviceInfoPayload struct {
	Model          string `json:"model"`
	Manufacturer   string `json:"manufacturer"`
	AndroidVersion string `json:"android_version"`
	AppVersion     string `json:"app_version"`
}

// CommandPayload is sent from the server to a device.
type CommandPayload struct {
	CommandID string          `json:"command_id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// CommandResultPayload is sent from the device back to the server.
type CommandResultPayload struct {
	CommandID string          `json:"command_id"`
	Success   bool            `json:"success"`
	Message   string          `json:"message,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// AckPayload is sent to the device on successful connection.
type AckPayload struct {
	DeviceID string `json:"device_id"`
}

// DeviceOnlinePayload is sent to dashboard clients.
type DeviceOnlinePayload struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

// TouchEventPayload is sent from dashboard → device for remote touch.
type TouchEventPayload struct {
	Action   string  `json:"action"`   // "tap", "swipe", "long_press"
	X        float64 `json:"x"`        // normalized 0.0-1.0
	Y        float64 `json:"y"`        // normalized 0.0-1.0
	EndX     float64 `json:"end_x,omitempty"` // for swipe
	EndY     float64 `json:"end_y,omitempty"` // for swipe
	Duration int     `json:"duration,omitempty"` // ms, for swipe/long_press
}

// MirrorFramePayload metadata sent alongside binary frame data.
type MirrorFramePayload struct {
	DeviceID string `json:"device_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

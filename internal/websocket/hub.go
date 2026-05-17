package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	ws "github.com/gofiber/contrib/websocket"
)

// Hub manages all active WebSocket connections (devices + dashboards).
type Hub struct {
	mu               sync.RWMutex
	deviceClients    map[uuid.UUID]*Client
	dashboardClients map[*ws.Conn]bool
	mirrorSessions   map[uuid.UUID]*ws.Conn // deviceID -> dashboard conn that is mirroring it
	onDeviceMessage  func(deviceID uuid.UUID, msg *WSMessage)
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		deviceClients:    make(map[uuid.UUID]*Client),
		dashboardClients: make(map[*ws.Conn]bool),
		mirrorSessions:   make(map[uuid.UUID]*ws.Conn),
	}
}

// SetMessageHandler sets the callback for incoming device messages.
func (h *Hub) SetMessageHandler(handler func(deviceID uuid.UUID, msg *WSMessage)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onDeviceMessage = handler
}

// RegisterDevice adds a device client to the hub.
func (h *Hub) RegisterDevice(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Disconnect existing connection for same device (if reconnecting)
	if existing, ok := h.deviceClients[client.DeviceID]; ok {
		log.Printf("[Hub] Replacing existing connection for device %s", client.DeviceID)
		existing.Close()
	}

	h.deviceClients[client.DeviceID] = client
	log.Printf("[Hub] Device registered: %s (%s) — total: %d",
		client.DeviceID, client.DeviceName, len(h.deviceClients))
}

// UnregisterDevice removes a device client from the hub.
func (h *Hub) UnregisterDevice(deviceID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client, ok := h.deviceClients[deviceID]; ok {
		client.Close()
		delete(h.deviceClients, deviceID)
		log.Printf("[Hub] Device unregistered: %s — total: %d",
			deviceID, len(h.deviceClients))
	}
}

// RegisterDashboard adds a dashboard WebSocket connection.
func (h *Hub) RegisterDashboard(conn *ws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.dashboardClients[conn] = true
	log.Printf("[Hub] Dashboard connected — total: %d", len(h.dashboardClients))
}

// UnregisterDashboard removes a dashboard connection.
func (h *Hub) UnregisterDashboard(conn *ws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.dashboardClients, conn)
	log.Printf("[Hub] Dashboard disconnected — total: %d", len(h.dashboardClients))
}

// SendToDevice sends a message to a specific device.
func (h *Hub) SendToDevice(deviceID uuid.UUID, msg *WSMessage) error {
	h.mu.RLock()
	client, ok := h.deviceClients[deviceID]
	h.mu.RUnlock()

	if !ok {
		return ErrDeviceNotConnected
	}

	return client.Send(msg)
}

// BroadcastToDashboards sends a message to all connected dashboard clients.
func (h *Hub) BroadcastToDashboards(msg *WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[Hub] Failed to marshal dashboard broadcast: %v", err)
		return
	}

	for conn := range h.dashboardClients {
		if err := conn.WriteMessage(1, data); err != nil {
			log.Printf("[Hub] Failed to send to dashboard: %v", err)
		}
	}
}

// IsDeviceConnected checks if a device has an active WebSocket connection.
func (h *Hub) IsDeviceConnected(deviceID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.deviceClients[deviceID]
	return ok
}

// HandleDeviceMessage is called by clients when they receive a message.
func (h *Hub) HandleDeviceMessage(deviceID uuid.UUID, msg *WSMessage) {
	h.mu.RLock()
	handler := h.onDeviceMessage
	h.mu.RUnlock()

	if handler != nil {
		handler(deviceID, msg)
	}
}

// StartPingLoop sends periodic pings to all connected devices.
func (h *Hub) StartPingLoop(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			h.mu.RLock()
			clients := make([]*Client, 0, len(h.deviceClients))
			for _, c := range h.deviceClients {
				clients = append(clients, c)
			}
			h.mu.RUnlock()

			pingMsg, _ := NewMessage(MsgPing, nil)
			for _, c := range clients {
				_ = c.Send(pingMsg)
			}
		}
	}()
}

// ConnectedDeviceCount returns the number of connected devices.
func (h *Hub) ConnectedDeviceCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.deviceClients)
}

// Custom errors
var ErrDeviceNotConnected = &HubError{Message: "device not connected"}

type HubError struct {
	Message string
}

func (e *HubError) Error() string {
	return e.Message
}

// StartMirrorSession maps a dashboard to a device for mirror streaming.
func (h *Hub) StartMirrorSession(deviceID uuid.UUID, dashConn *ws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.mirrorSessions[deviceID] = dashConn
	log.Printf("[Hub] Mirror session started for device %s", deviceID)
}

// StopMirrorSession removes the mirror session for a device.
func (h *Hub) StopMirrorSession(deviceID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.mirrorSessions, deviceID)
	log.Printf("[Hub] Mirror session stopped for device %s", deviceID)
}

// StopMirrorSessionsForDashboard removes all mirror sessions for a disconnecting dashboard.
func (h *Hub) StopMirrorSessionsForDashboard(dashConn *ws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for deviceID, conn := range h.mirrorSessions {
		if conn == dashConn {
			delete(h.mirrorSessions, deviceID)
			log.Printf("[Hub] Mirror session cleaned up for device %s (dashboard disconnected)", deviceID)
			// Tell device to stop streaming
			if client, ok := h.deviceClients[deviceID]; ok {
				stopMsg, _ := NewMessage(MsgStopMirror, nil)
				_ = client.Send(stopMsg)
			}
		}
	}
}

// SendMirrorFrame sends a raw binary frame to the dashboard mirroring this device.
func (h *Hub) SendMirrorFrame(deviceID uuid.UUID, frameData []byte) {
	h.mu.RLock()
	dashConn, ok := h.mirrorSessions[deviceID]
	h.mu.RUnlock()

	if !ok {
		return // No active mirror session for this device
	}

	// Send as binary WebSocket message (raw JPEG bytes)
	if err := dashConn.WriteMessage(ws.BinaryMessage, frameData); err != nil {
		log.Printf("[Hub] Failed to send mirror frame to dashboard: %v", err)
		h.StopMirrorSession(deviceID)
	}
}

package websocket

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yuwanadev/mdm-backend/internal/service"
)

// Handler manages WebSocket upgrade and connection lifecycle.
type Handler struct {
	hub           *Hub
	deviceService *service.DeviceService
	authService   *service.AuthService
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *Hub, deviceService *service.DeviceService, authService *service.AuthService) *Handler {
	return &Handler{
		hub:           hub,
		deviceService: deviceService,
		authService:   authService,
	}
}

// DeviceUpgrade handles the WebSocket upgrade for device connections.
// Route: GET /ws/device?token=<raw_device_token>
//
// The token is validated during the upgrade check middleware. The device ID
// is stored in Locals and read here.
func (h *Handler) DeviceUpgrade() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// ... existing logic ...
		deviceID, ok := c.Locals("deviceID").(uuid.UUID)
		if !ok {
			log.Printf("[WS] Device connection rejected: no device ID in context (IP: %s)", c.RemoteAddr())
			c.Close()
			return
		}
		deviceName, _ := c.Locals("deviceName").(string)

		log.Printf("[WS] Device connected successfully: %s (%s) from %s", deviceID, deviceName, c.RemoteAddr())

		ctx := context.Background()

		// Mark device online
		_ = h.deviceService.SetDeviceOnline(ctx, deviceID, true)

		// Create client and register
		client := NewClient(c, deviceID, deviceName, h.hub)
		h.hub.RegisterDevice(client)

		// Send ACK to device
		ackMsg, _ := NewMessage(MsgAck, AckPayload{DeviceID: deviceID.String()})
		_ = client.Send(ackMsg)

		// Notify dashboards
		onlineMsg, _ := NewMessage(MsgDeviceOnline, DeviceOnlinePayload{
			DeviceID:   deviceID.String(),
			DeviceName: deviceName,
		})
		h.hub.BroadcastToDashboards(onlineMsg)

		// Start read/write pumps (blocks until disconnect)
		client.Start()

		// Device disconnected — mark offline
		_ = h.deviceService.SetDeviceOnline(ctx, deviceID, false)

		offlineMsg, _ := NewMessage(MsgDeviceOffline, DeviceOnlinePayload{
			DeviceID:   deviceID.String(),
			DeviceName: deviceName,
		})
		h.hub.BroadcastToDashboards(offlineMsg)

		log.Printf("[WS] Device disconnected: %s (%s)", deviceID, deviceName)
	}, websocket.Config{
		ReadBufferSize:  1024 * 1024 * 4, // 4MB
		WriteBufferSize: 1024 * 1024 * 4, // 4MB
	})
}

// DeviceUpgradeCheck validates the device token before upgrading.
func (h *Handler) DeviceUpgradeCheck() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		token := c.Query("token")
		if token == "" {
			log.Printf("[WS] Upgrade rejected: missing token from %s", c.IP())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "token is required",
			})
		}

		// Authenticate device
		device, err := h.deviceService.AuthenticateDevice(c.Context(), token)
		if err != nil {
			log.Printf("[WS] Upgrade rejected: invalid token '%s' from %s", token, c.IP())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid device token",
			})
		}

		// Store device info for the WebSocket handler
		c.Locals("deviceID", device.ID)
		c.Locals("deviceName", device.DeviceName)

		return c.Next()
	}
}

// DashboardUpgrade handles the WebSocket upgrade for dashboard connections.
// Route: GET /ws/dashboard?token=<jwt>
func (h *Handler) DashboardUpgrade() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		log.Println("[WS] Dashboard connected")
		h.hub.RegisterDashboard(c)

		defer func() {
			// Clean up any mirror sessions for this dashboard
			h.hub.StopMirrorSessionsForDashboard(c)
			h.hub.UnregisterDashboard(c)
			log.Println("[WS] Dashboard disconnected")
		}()

		// Read loop — dashboard can send commands via WS
		for {
			_, data, err := c.ReadMessage()
			if err != nil {
				break
			}

			var msg WSMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}

			log.Printf("[WS] Dashboard message: type=%s", msg.Type)

			switch msg.Type {
			case MsgStartMirror:
				h.handleDashboardStartMirror(c, &msg)
			case MsgStopMirror:
				h.handleDashboardStopMirror(&msg)
			case MsgTouchEvent:
				h.handleDashboardTouchEvent(&msg)
			case MsgWebRTCSignal:
				h.handleDashboardWebRTCSignal(&msg)
			}
		}
	}, websocket.Config{
		ReadBufferSize:  1024 * 1024 * 4, // 4MB
		WriteBufferSize: 1024 * 1024 * 4, // 4MB
	})
}

// DashboardUpgradeCheck validates the JWT before upgrading.
func (h *Handler) DashboardUpgradeCheck() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		token := c.Query("token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "token is required",
			})
		}

		if _, err := h.authService.ValidateToken(token); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		return c.Next()
	}
}

// ── Mirror/Touch helpers ──────────────────────────

type mirrorPayload struct {
	DeviceID string `json:"device_id"`
}

func (h *Handler) handleDashboardStartMirror(dashConn *websocket.Conn, msg *WSMessage) {
	log.Printf("[MIRROR] ← Dashboard sent START_MIRROR, raw payload: %s", string(msg.Payload))

	var payload mirrorPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Printf("[MIRROR] ✗ START_MIRROR: invalid payload: %v", err)
		return
	}

	deviceID, err := uuid.Parse(payload.DeviceID)
	if err != nil {
		log.Printf("[MIRROR] ✗ START_MIRROR: invalid device_id '%s': %v", payload.DeviceID, err)
		return
	}

	// Register mirror session
	h.hub.StartMirrorSession(deviceID, dashConn)
	log.Printf("[MIRROR] ✓ Session registered for device %s", deviceID)

	// Forward START_MIRROR to the device
	startMsg, _ := NewMessage(MsgStartMirror, nil)
	if err := h.hub.SendToDevice(deviceID, startMsg); err != nil {
		log.Printf("[MIRROR] ✗ Device %s not connected — cannot forward START_MIRROR", deviceID)
	} else {
		log.Printf("[MIRROR] → Forwarded START_MIRROR to device %s", deviceID)
	}
}

func (h *Handler) handleDashboardStopMirror(msg *WSMessage) {
	log.Printf("[MIRROR] ← Dashboard sent STOP_MIRROR, raw payload: %s", string(msg.Payload))

	var payload mirrorPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Printf("[MIRROR] ✗ STOP_MIRROR: invalid payload: %v", err)
		return
	}

	deviceID, err := uuid.Parse(payload.DeviceID)
	if err != nil {
		log.Printf("[MIRROR] ✗ STOP_MIRROR: invalid device_id: %v", err)
		return
	}

	h.hub.StopMirrorSession(deviceID)

	// Forward STOP_MIRROR to the device
	stopMsg, _ := NewMessage(MsgStopMirror, nil)
	if err := h.hub.SendToDevice(deviceID, stopMsg); err != nil {
		log.Printf("[MIRROR] ✗ Device %s not connected — cannot forward STOP_MIRROR", deviceID)
	} else {
		log.Printf("[MIRROR] → Forwarded STOP_MIRROR to device %s", deviceID)
	}
}

func (h *Handler) handleDashboardTouchEvent(msg *WSMessage) {
	log.Printf("[TOUCH] ← Dashboard sent TOUCH_EVENT, raw payload: %s", string(msg.Payload))

	// Payload contains device_id + touch coords
	var rawPayload struct {
		DeviceID string          `json:"device_id"`
		Touch    json.RawMessage `json:"touch"`
	}
	if err := json.Unmarshal(msg.Payload, &rawPayload); err != nil {
		log.Printf("[TOUCH] ✗ Invalid payload: %v", err)
		return
	}

	deviceID, err := uuid.Parse(rawPayload.DeviceID)
	if err != nil {
		log.Printf("[TOUCH] ✗ Invalid device_id '%s': %v", rawPayload.DeviceID, err)
		return
	}

	log.Printf("[TOUCH] Parsed: device=%s touch=%s", deviceID, string(rawPayload.Touch))

	// Forward touch event to device
	touchMsg := NewMessageRaw(MsgTouchEvent, rawPayload.Touch)
	if err := h.hub.SendToDevice(deviceID, touchMsg); err != nil {
		log.Printf("[TOUCH] ✗ Device %s not connected — cannot forward TOUCH_EVENT", deviceID)
	} else {
		log.Printf("[TOUCH] → Forwarded TOUCH_EVENT to device %s", deviceID)
	}
}

func (h *Handler) handleDashboardWebRTCSignal(msg *WSMessage) {
	var rawPayload struct {
		DeviceID string          `json:"device_id"`
		Signal   json.RawMessage `json:"signal"`
	}
	if err := json.Unmarshal(msg.Payload, &rawPayload); err != nil {
		log.Printf("[WEBRTC] ✗ Invalid payload: %v", err)
		return
	}

	deviceID, err := uuid.Parse(rawPayload.DeviceID)
	if err != nil {
		log.Printf("[WEBRTC] ✗ Invalid device_id: %v", err)
		return
	}

	// Forward signal to device
	signalMsg := NewMessageRaw(MsgWebRTCSignal, rawPayload.Signal)
	if err := h.hub.SendToDevice(deviceID, signalMsg); err != nil {
		log.Printf("[WEBRTC] ✗ Device %s not connected — cannot forward signal", deviceID)
	} else {
		log.Printf("[WEBRTC] → Forwarded WEBRTC_SIGNAL to device %s", deviceID)
	}
}

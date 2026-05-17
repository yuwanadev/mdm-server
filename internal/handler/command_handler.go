package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yuwanadev/mdm-backend/internal/models"
	"github.com/yuwanadev/mdm-backend/internal/service"
	ws "github.com/yuwanadev/mdm-backend/internal/websocket"
	"github.com/yuwanadev/mdm-backend/pkg/response"
)

type CommandHandler struct {
	commandService *service.CommandService
	hub            *ws.Hub
}

func NewCommandHandler(commandService *service.CommandService, hub *ws.Hub) *CommandHandler {
	return &CommandHandler{
		commandService: commandService,
		hub:            hub,
	}
}

// Send creates and dispatches a command to a device.
// POST /api/devices/:id/commands
func (h *CommandHandler) Send(c *fiber.Ctx) error {
	deviceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid device id")
	}

	var body struct {
		CommandType string          `json:"command_type"`
		Payload     json.RawMessage `json:"payload,omitempty"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if body.CommandType == "" {
		return response.BadRequest(c, "command_type is required")
	}

	// Create command log
	cmdLog, err := h.commandService.SendCommand(c.Context(), deviceID, body.CommandType, body.Payload)
	if err != nil {
		switch err {
		case service.ErrDeviceNotFound:
			return response.NotFound(c, "device not found")
		case service.ErrDeviceOffline:
			return response.BadRequest(c, "device is offline")
		case service.ErrInvalidCommand:
			return response.BadRequest(c, "invalid command type")
		default:
			return response.InternalError(c, "failed to create command")
		}
	}

	// Dispatch via WebSocket
	cmdPayload := ws.CommandPayload{
		CommandID: cmdLog.ID.String(),
		Type:      cmdLog.CommandType,
		Payload:   cmdLog.Payload,
	}
	wsMsg, _ := ws.NewMessage(ws.MsgCommand, cmdPayload)
	if err := h.hub.SendToDevice(deviceID, wsMsg); err != nil {
		// Device disconnected between check and send
		return response.BadRequest(c, "device is no longer connected")
	}

	// Mark as sent
	_ = h.commandService.MarkSent(c.Context(), cmdLog.ID)

	return response.OK(c, cmdLog)
}

// BulkSend creates and dispatches a command to all devices in a group.
// POST /api/groups/:id/commands
func (h *CommandHandler) BulkSend(c *fiber.Ctx) error {
	groupID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}

	var body struct {
		CommandType string          `json:"command_type"`
		Payload     json.RawMessage `json:"payload,omitempty"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if body.CommandType == "" {
		return response.BadRequest(c, "command_type is required")
	}

	// Create command logs
	cmdLogs, err := h.commandService.BulkSendCommand(c.Context(), groupID, body.CommandType, body.Payload)
	if err != nil {
		return response.InternalError(c, "failed to create bulk commands")
	}

	// Dispatch to each device
	sentCount := 0
	for _, log := range cmdLogs {
		cmdPayload := ws.CommandPayload{
			CommandID: log.ID.String(),
			Type:      log.CommandType,
			Payload:   log.Payload,
		}
		wsMsg, _ := ws.NewMessage(ws.MsgCommand, cmdPayload)
		if err := h.hub.SendToDevice(log.DeviceID, wsMsg); err == nil {
			_ = h.commandService.MarkSent(c.Context(), log.ID)
			sentCount++
		}
	}

	return response.OK(c, fiber.Map{
		"total": devicesCount(cmdLogs),
		"sent":  sentCount,
	})
}

func devicesCount(logs []*models.CommandLog) int {
	return len(logs)
}

// History returns command history for a device.
// GET /api/devices/:id/commands
func (h *CommandHandler) History(c *fiber.Ctx) error {
	deviceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid device id")
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	logs, err := h.commandService.GetHistory(c.Context(), deviceID, limit)
	if err != nil {
		return response.InternalError(c, "failed to fetch command history")
	}

	if logs == nil {
		logs = []models.CommandLog{}
	}

	return response.OK(c, logs)
}

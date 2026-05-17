package handler

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yuwanadev/mdm-backend/internal/models"
	"github.com/yuwanadev/mdm-backend/internal/service"
	"github.com/yuwanadev/mdm-backend/pkg/response"
)

type DeviceHandler struct {
	deviceService     *service.DeviceService
	screenshotService *service.ScreenshotService
}

func NewDeviceHandler(deviceService *service.DeviceService, screenshotService *service.ScreenshotService) *DeviceHandler {
	return &DeviceHandler{
		deviceService:     deviceService,
		screenshotService: screenshotService,
	}
}

// List returns all registered devices.
// GET /api/devices
func (h *DeviceHandler) List(c *fiber.Ctx) error {
	devices, err := h.deviceService.GetAllDevices(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to fetch devices")
	}

	if devices == nil {
		devices = []models.Device{} // return [] not null
	}

	return response.OK(c, devices)
}

// Get returns a single device with its status.
// GET /api/devices/:id
func (h *DeviceHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid device id")
	}

	device, err := h.deviceService.GetDeviceWithStatus(c.Context(), id)
	if err != nil {
		return response.NotFound(c, "device not found")
	}

	return response.OK(c, device)
}

// Create registers a new device and returns the raw token (shown only once).
// POST /api/devices
func (h *DeviceHandler) Create(c *fiber.Ctx) error {
	var body struct {
		DeviceName string `json:"device_name"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if body.DeviceName == "" {
		return response.BadRequest(c, "device_name is required")
	}

	result, err := h.deviceService.CreateDevice(c.Context(), body.DeviceName)
	if err != nil {
		return response.InternalError(c, "failed to create device")
	}

	return response.Created(c, result)
}

// Delete removes a device.
// DELETE /api/devices/:id
func (h *DeviceHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid device id")
	}

	if err := h.deviceService.DeleteDevice(c.Context(), id); err != nil {
		return response.InternalError(c, "failed to delete device")
	}

	return response.NoContent(c)
}

// Update updates device details.
// PUT /api/devices/:id
func (h *DeviceHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid device id")
	}

	var body struct {
		DeviceName string     `json:"device_name"`
		Label      string     `json:"label"`
		Notes      string     `json:"notes"`
		GroupID    *uuid.UUID `json:"group_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if body.DeviceName == "" {
		return response.BadRequest(c, "device_name is required")
	}

	if err := h.deviceService.UpdateDevice(c.Context(), id, body.DeviceName, body.Label, body.Notes, body.GroupID); err != nil {
		return response.InternalError(c, "failed to update device")
	}

	return response.NoContent(c)
}

// Screenshot returns the latest screenshot for a device.
// GET /api/devices/:id/screenshot
func (h *DeviceHandler) Screenshot(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid device id")
	}

	path := h.screenshotService.GetScreenshotPath(id)
	fmt.Printf("[DEBUG] Attempting to serve screenshot from: %s\n", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("[ERROR] Screenshot file not found: %s\n", path)
		return response.NotFound(c, "no screenshot available")
	}

	c.Set("Content-Type", "image/jpeg")
	return c.SendFile(path)
}

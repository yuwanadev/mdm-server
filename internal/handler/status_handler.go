package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yuwanadev/mdm-backend/internal/repository"
	"github.com/yuwanadev/mdm-backend/pkg/response"
)

type StatusHandler struct {
	statusRepo *repository.StatusRepo
}

func NewStatusHandler(statusRepo *repository.StatusRepo) *StatusHandler {
	return &StatusHandler{statusRepo: statusRepo}
}

// Get returns the latest status for a device.
// GET /api/devices/:id/status
func (h *StatusHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid device id")
	}

	status, err := h.statusRepo.GetByDeviceID(c.Context(), id)
	if err != nil {
		return response.NotFound(c, "no status data for this device")
	}

	return response.OK(c, status)
}

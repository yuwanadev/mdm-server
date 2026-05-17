package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yuwanadev/mdm-backend/internal/service"
	"github.com/yuwanadev/mdm-backend/pkg/response"
)

type GroupHandler struct {
	groupService *service.GroupService
}

func NewGroupHandler(groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

func (h *GroupHandler) List(c *fiber.Ctx) error {
	groups, err := h.groupService.GetAllGroups(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to fetch groups")
	}
	return response.OK(c, groups)
}

func (h *GroupHandler) Create(c *fiber.Ctx) error {
	var body struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if body.Name == "" {
		return response.BadRequest(c, "name is required")
	}

	group, err := h.groupService.CreateGroup(c.Context(), body.Name)
	if err != nil {
		return response.InternalError(c, "failed to create group")
	}
	return response.Created(c, group)
}

func (h *GroupHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "invalid group id")
	}

	if err := h.groupService.DeleteGroup(c.Context(), id); err != nil {
		return response.InternalError(c, "failed to delete group")
	}
	return response.NoContent(c)
}

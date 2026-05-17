package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yuwanadev/mdm-backend/internal/service"
	"github.com/yuwanadev/mdm-backend/pkg/response"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login authenticates a user and returns JWT tokens.
// POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if body.Username == "" || body.Password == "" {
		return response.BadRequest(c, "username and password are required")
	}

	tokens, err := h.authService.Login(c.Context(), body.Username, body.Password)
	if err != nil {
		return response.Unauthorized(c, "invalid credentials")
	}

	return response.OK(c, tokens)
}

// Refresh generates new tokens from a refresh token.
// POST /api/auth/refresh
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.BadRequest(c, "invalid request body")
	}

	if body.RefreshToken == "" {
		return response.BadRequest(c, "refresh_token is required")
	}

	tokens, err := h.authService.RefreshAccessToken(body.RefreshToken)
	if err != nil {
		return response.Unauthorized(c, "invalid or expired refresh token")
	}

	return response.OK(c, tokens)
}

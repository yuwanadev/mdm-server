package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yuwanadev/mdm-backend/internal/models"
	"github.com/yuwanadev/mdm-backend/internal/repository"
	"github.com/yuwanadev/mdm-backend/pkg/response"
)

type APKHandler struct {
	repo *repository.APKRepo
}

func NewAPKHandler(repo *repository.APKRepo) *APKHandler {
	return &APKHandler{repo: repo}
}

func (h *APKHandler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "no file uploaded")
	}

	packageName := c.FormValue("package_name")
	versionName := c.FormValue("version_name")
	versionCodeStr := c.FormValue("version_code")

	versionCode, _ := strconv.Atoi(versionCodeStr)

	if packageName == "" || versionName == "" || versionCode == 0 {
		return response.BadRequest(c, "missing metadata")
	}

	// Save file
	filename := fmt.Sprintf("%s_%s_%d_%s", packageName, versionName, versionCode, uuid.New().String()[:8])
	savePath := filepath.Join("storage", "apks", filename+".apk")
	
	if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
		return response.InternalError(c, "failed to create storage directory")
	}

	if err := c.SaveFile(file, savePath); err != nil {
		return response.InternalError(c, "failed to save file")
	}

	apk := &models.APK{
		PackageName: packageName,
		VersionName: versionName,
		VersionCode: versionCode,
		FilePath:    savePath,
		FileSize:    file.Size,
	}

	if err := h.repo.Create(c.Context(), apk); err != nil {
		return response.InternalError(c, "failed to save APK metadata")
	}

	return response.OK(c, apk)
}

func (h *APKHandler) List(c *fiber.Ctx) error {
	apks, err := h.repo.List(c.Context())
	if err != nil {
		return response.InternalError(c, "failed to list APKs")
	}
	if apks == nil {
		apks = []models.APK{}
	}
	return response.OK(c, apks)
}

func (h *APKHandler) Download(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.BadRequest(c, "invalid ID")
	}

	apk, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		return response.NotFound(c, "APK not found")
	}

	return c.SendFile(apk.FilePath)
}

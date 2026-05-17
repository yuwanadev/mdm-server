package service

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type ScreenshotService struct {
	basePath string
}

func NewScreenshotService(basePath string) *ScreenshotService {
	// Ensure directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		fmt.Printf("Warning: failed to create screenshot directory: %v\n", err)
	}
	return &ScreenshotService{basePath: basePath}
}

// SaveScreenshot saves a base64 encoded image to storage/screenshots/{device_id}.jpg
func (s *ScreenshotService) SaveScreenshot(deviceID uuid.UUID, base64Data string) (string, error) {
	// Strip prefix if present (e.g. data:image/jpeg;base64,)
	if i := strings.Index(base64Data, ","); i != -1 {
		base64Data = base64Data[i+1:]
	}

	// Use RawStdEncoding to be lenient with padding
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		// Try raw if standard fails
		data, err = base64.RawStdEncoding.DecodeString(base64Data)
		if err != nil {
			return "", fmt.Errorf("failed to decode base64: %w", err)
		}
	}

	fileName := fmt.Sprintf("%s.jpg", deviceID.String())
	filePath := filepath.Join(s.basePath, fileName)

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

func (s *ScreenshotService) GetScreenshotPath(deviceID uuid.UUID) string {
	return filepath.Join(s.basePath, fmt.Sprintf("%s.jpg", deviceID.String()))
}

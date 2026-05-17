package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yuwanadev/mdm-backend/internal/models"
	"github.com/yuwanadev/mdm-backend/internal/repository"
)

var (
	ErrCommandNotFound = errors.New("command not found")
	ErrDeviceOffline   = errors.New("device is offline")
	ErrInvalidCommand  = errors.New("invalid command type")
)

// AllowedCommands is the whitelist of permitted command types.
var AllowedCommands = map[string]bool{
	"PING":            true,
	"GET_DEVICE_INFO": true,
	"OPEN_APP":        true,
	"RESTART_APP":     true,
	"GET_BATTERY":     true,
	"GET_STORAGE":     true,
	"TAKE_SCREENSHOT": true,
	"SHOW_ALERT":     true,
	"LOCK_DEVICE":    true,
	"FACTORY_RESET":  true,
	"INSTALL_APK":    true,
	"SET_DEV_MODE":   true,
	"SET_USB_DEBUGGING": true,
}

type CommandService struct {
	commandRepo       *repository.CommandRepo
	deviceRepo        *repository.DeviceRepo
	screenshotService *ScreenshotService
}

func NewCommandService(commandRepo *repository.CommandRepo, deviceRepo *repository.DeviceRepo, screenshotService *ScreenshotService) *CommandService {
	return &CommandService{
		commandRepo:       commandRepo,
		deviceRepo:        deviceRepo,
		screenshotService: screenshotService,
	}
}

// SendCommand creates a command log and returns it for dispatch via WebSocket.
func (s *CommandService) SendCommand(ctx context.Context, deviceID uuid.UUID, cmdType string, payload json.RawMessage) (*models.CommandLog, error) {
	if !AllowedCommands[cmdType] {
		return nil, ErrInvalidCommand
	}

	// Check device exists
	device, err := s.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return nil, ErrDeviceNotFound
	}

	if !device.IsOnline {
		return nil, ErrDeviceOffline
	}

	return s.commandRepo.Create(ctx, deviceID, cmdType, payload)
}

// BulkSendCommand sends a command to all online devices in a group.
func (s *CommandService) BulkSendCommand(ctx context.Context, groupID uuid.UUID, cmdType string, payload json.RawMessage) ([]*models.CommandLog, error) {
	if !AllowedCommands[cmdType] {
		return nil, ErrInvalidCommand
	}

	devices, err := s.deviceRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var logs []*models.CommandLog
	for _, d := range devices {
		if d.IsOnline {
			log, err := s.commandRepo.Create(ctx, d.ID, cmdType, payload)
			if err == nil {
				logs = append(logs, log)
			}
		}
	}

	return logs, nil
}

// MarkSent marks a command as sent to the device.
func (s *CommandService) MarkSent(ctx context.Context, id uuid.UUID) error {
	return s.commandRepo.SetSent(ctx, id)
}

// HandleResult processes a command result from the device.
func (s *CommandService) HandleResult(ctx context.Context, cmdID uuid.UUID, success bool, message string, result json.RawMessage) error {
	status := models.CommandStatusSuccess
	if !success {
		status = models.CommandStatusFailed
	}

	// Fetch the command to check its type
	cmd, err := s.commandRepo.GetByID(ctx, cmdID)
	if err != nil {
		return err
	}

	// Special handling for screenshots
	if success && cmd.CommandType == "TAKE_SCREENSHOT" {
		fmt.Printf("[DEBUG] Processing screenshot for device %s\n", cmd.DeviceID)
		var data string
		if err := json.Unmarshal(result, &data); err == nil && data != "" {
			fmt.Printf("[DEBUG] Screenshot data received (length: %d)\n", len(data))
			if _, err := s.screenshotService.SaveScreenshot(cmd.DeviceID, data); err != nil {
				fmt.Printf("[ERROR] Failed to save screenshot: %v\n", err)
			} else {
				fmt.Printf("[SUCCESS] Screenshot saved for device %s\n", cmd.DeviceID)
				// Replace the bulky base64 data with a placeholder in the DB
				result = json.RawMessage(`"SCREENSHOT_SAVED"`)
			}
		} else {
			fmt.Printf("[ERROR] Failed to unmarshal screenshot result: %v\n", err)
		}
	}

	return s.commandRepo.UpdateStatus(ctx, cmdID, status, message, result)
}

// GetHistory returns command history for a device.
func (s *CommandService) GetHistory(ctx context.Context, deviceID uuid.UUID, limit int) ([]models.CommandLog, error) {
	return s.commandRepo.GetByDeviceID(ctx, deviceID, limit)
}

// MarkTimeouts marks old commands as timed out.
func (s *CommandService) MarkTimeouts(ctx context.Context, threshold time.Duration) (int64, error) {
	return s.commandRepo.MarkTimeouts(ctx, threshold)
}

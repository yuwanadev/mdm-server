package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/yuwanadev/mdm-backend/internal/models"
	"github.com/yuwanadev/mdm-backend/internal/repository"
)

var (
	ErrDeviceNotFound = errors.New("device not found")
)

type DeviceService struct {
	deviceRepo *repository.DeviceRepo
	statusRepo *repository.StatusRepo
}

func NewDeviceService(deviceRepo *repository.DeviceRepo, statusRepo *repository.StatusRepo) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
		statusRepo: statusRepo,
	}
}

// CreateDeviceResult contains the new device and the raw token (shown only once).
type CreateDeviceResult struct {
	Device   *models.Device `json:"device"`
	RawToken string         `json:"token"` // shown once, then discarded
}

// CreateDevice registers a new device and generates a unique token.
func (s *DeviceService) CreateDevice(ctx context.Context, name string) (*CreateDeviceResult, error) {
	rawToken, err := generateToken(32)
	if err != nil {
		return nil, err
	}

	tokenHash := hashToken(rawToken)

	device, err := s.deviceRepo.Create(ctx, name, tokenHash)
	if err != nil {
		return nil, err
	}

	return &CreateDeviceResult{
		Device:   device,
		RawToken: rawToken,
	}, nil
}

// GetAllDevices returns all registered devices.
func (s *DeviceService) GetAllDevices(ctx context.Context) ([]models.Device, error) {
	return s.deviceRepo.GetAll(ctx)
}

// GetDevice returns a single device by ID.
func (s *DeviceService) GetDevice(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	return s.deviceRepo.GetByID(ctx, id)
}

// GetDeviceWithStatus returns a device with its latest status.
func (s *DeviceService) GetDeviceWithStatus(ctx context.Context, id uuid.UUID) (*models.DeviceWithStatus, error) {
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	status, _ := s.statusRepo.GetByDeviceID(ctx, id) // may not exist yet

	return &models.DeviceWithStatus{
		Device: *device,
		Status: status,
	}, nil
}

// DeleteDevice removes a device.
func (s *DeviceService) DeleteDevice(ctx context.Context, id uuid.UUID) error {
	return s.deviceRepo.Delete(ctx, id)
}

// AuthenticateDevice validates a raw token and returns the matching device.
func (s *DeviceService) AuthenticateDevice(ctx context.Context, rawToken string) (*models.Device, error) {
	tokenHash := hashToken(rawToken)
	device, err := s.deviceRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		// Log the first 8 chars of the hash for debugging without leaking full token
		log.Printf("[Auth] Authentication failed for token hash starting with %s", tokenHash[:8])
		return nil, ErrDeviceNotFound
	}
	return device, nil
}

// SetDeviceOnline updates the online status.
func (s *DeviceService) SetDeviceOnline(ctx context.Context, id uuid.UUID, online bool) error {
	return s.deviceRepo.SetOnline(ctx, id, online)
}

// UpdateDeviceInfo updates model and Android version.
func (s *DeviceService) UpdateDeviceInfo(ctx context.Context, id uuid.UUID, model, androidVersion, agentVersion string) error {
	return s.deviceRepo.UpdateDeviceInfo(ctx, id, model, androidVersion, agentVersion)
}

// UpdateDevice updates editable device details.
func (s *DeviceService) UpdateDevice(ctx context.Context, id uuid.UUID, name, label, notes string, groupID *uuid.UUID) error {
	return s.deviceRepo.Update(ctx, id, name, label, notes, groupID)
}

// MarkOfflineInactive marks devices as offline if they haven't been seen for a while.
func (s *DeviceService) MarkOfflineInactive(ctx context.Context, timeout time.Duration) (int64, error) {
	return s.deviceRepo.MarkOfflineInactive(ctx, timeout)
}

// generateToken creates a cryptographically random hex token.
func generateToken(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken produces a SHA-256 hex digest of a raw token.
// We use SHA-256 (not bcrypt) because tokens are high-entropy random strings,
// not human passwords, so preimage resistance is sufficient.
func hashToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yuwanadev/mdm-backend/internal/models"
)

type DeviceRepo struct {
	pool *pgxpool.Pool
}

func NewDeviceRepo(pool *pgxpool.Pool) *DeviceRepo {
	return &DeviceRepo{pool: pool}
}

// Create registers a new device with a hashed token.
func (r *DeviceRepo) Create(ctx context.Context, name, tokenHash string) (*models.Device, error) {
	var d models.Device
	err := r.pool.QueryRow(ctx,
		`INSERT INTO devices (device_name, label, notes, group_id, token_hash)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, device_name, label, notes, group_id, device_model, android_version, agent_version, token_hash,
		           is_online, last_seen, created_at`,
		name, "", "", nil, tokenHash,
	).Scan(&d.ID, &d.DeviceName, &d.Label, &d.Notes, &d.GroupID, &d.DeviceModel, &d.AndroidVersion, &d.AgentVersion,
		&d.TokenHash, &d.IsOnline, &d.LastSeen, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// GetAll returns all devices.
func (r *DeviceRepo) GetAll(ctx context.Context) ([]models.Device, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, device_name, label, notes, group_id, device_model, android_version, agent_version,
		        is_online, last_seen, created_at
		 FROM devices ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.DeviceName, &d.Label, &d.Notes, &d.GroupID, &d.DeviceModel,
			&d.AndroidVersion, &d.AgentVersion, &d.IsOnline, &d.LastSeen, &d.CreatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// GetByGroupID returns all devices in a group.
func (r *DeviceRepo) GetByGroupID(ctx context.Context, groupID uuid.UUID) ([]models.Device, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, device_name, label, notes, group_id, device_model, android_version, agent_version,
		        is_online, last_seen, created_at
		 FROM devices WHERE group_id = $1 ORDER BY device_name ASC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.DeviceName, &d.Label, &d.Notes, &d.GroupID, &d.DeviceModel,
			&d.AndroidVersion, &d.AgentVersion, &d.IsOnline, &d.LastSeen, &d.CreatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// GetByID returns a single device.
func (r *DeviceRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	var d models.Device
	err := r.pool.QueryRow(ctx,
		`SELECT id, device_name, label, notes, group_id, device_model, android_version, agent_version,
		        is_online, last_seen, created_at
		 FROM devices WHERE id = $1`, id,
	).Scan(&d.ID, &d.DeviceName, &d.Label, &d.Notes, &d.GroupID, &d.DeviceModel, &d.AndroidVersion, &d.AgentVersion,
		&d.IsOnline, &d.LastSeen, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// GetByTokenHash finds a device by its hashed token.
func (r *DeviceRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Device, error) {
	var d models.Device
	err := r.pool.QueryRow(ctx,
		`SELECT id, device_name, label, notes, group_id, device_model, android_version, agent_version,
		        is_online, last_seen, created_at
		 FROM devices WHERE token_hash = $1`, tokenHash,
	).Scan(&d.ID, &d.DeviceName, &d.Label, &d.Notes, &d.GroupID, &d.DeviceModel, &d.AndroidVersion, &d.AgentVersion,
		&d.IsOnline, &d.LastSeen, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// Delete removes a device by ID.
func (r *DeviceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM devices WHERE id = $1`, id)
	return err
}

// SetOnline updates the online status and last_seen timestamp.
func (r *DeviceRepo) SetOnline(ctx context.Context, id uuid.UUID, online bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE devices SET is_online = $1, last_seen = NOW() WHERE id = $2`,
		online, id)
	return err
}

// UpdateDeviceInfo updates model and Android version from agent heartbeat.
func (r *DeviceRepo) UpdateDeviceInfo(ctx context.Context, id uuid.UUID, model, androidVersion, agentVersion string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE devices SET device_model = $1, android_version = $2, agent_version = $3, last_seen = NOW()
		 WHERE id = $4`,
		model, androidVersion, agentVersion, id)
	return err
}

// GetAllTokenHashes returns all token hashes for device auth lookup.
func (r *DeviceRepo) GetAllTokenHashes(ctx context.Context) (map[string]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, token_hash FROM devices`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]uuid.UUID)
	for rows.Next() {
		var id uuid.UUID
		var hash string
		if err := rows.Scan(&id, &hash); err != nil {
			return nil, err
		}
		result[hash] = id
	}
	return result, rows.Err()
}

// Update updates device details.
func (r *DeviceRepo) Update(ctx context.Context, id uuid.UUID, name, label, notes string, groupID *uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE devices SET device_name = $1, label = $2, notes = $3, group_id = $4 WHERE id = $5`,
		name, label, notes, groupID, id)
	return err
}

// MarkOfflineInactive marks devices as offline if they haven't been seen for a while.
func (r *DeviceRepo) MarkOfflineInactive(ctx context.Context, timeout time.Duration) (int64, error) {
	res, err := r.pool.Exec(ctx,
		`UPDATE devices SET is_online = FALSE 
		 WHERE is_online = TRUE AND (last_seen < $1 OR last_seen IS NULL)`,
		time.Now().Add(-timeout))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

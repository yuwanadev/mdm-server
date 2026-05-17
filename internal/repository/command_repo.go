package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yuwanadev/mdm-backend/internal/models"
)

type CommandRepo struct {
	pool *pgxpool.Pool
}

func NewCommandRepo(pool *pgxpool.Pool) *CommandRepo {
	return &CommandRepo{pool: pool}
}

// Create inserts a new command log.
func (r *CommandRepo) Create(ctx context.Context, deviceID uuid.UUID, cmdType string, payload json.RawMessage) (*models.CommandLog, error) {
	var c models.CommandLog
	err := r.pool.QueryRow(ctx,
		`INSERT INTO command_logs (device_id, command_type, payload)
		 VALUES ($1, $2, $3)
		 RETURNING id, device_id, command_type, payload, status, message, result, created_at, completed_at`,
		deviceID, cmdType, payload,
	).Scan(&c.ID, &c.DeviceID, &c.CommandType, &c.Payload,
		&c.Status, &c.Message, &c.Result, &c.CreatedAt, &c.CompletedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateStatus updates the status and result of a command.
func (r *CommandRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status models.CommandStatus, message string, result json.RawMessage) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE command_logs
		 SET status = $1, message = $2, result = $3, completed_at = NOW()
		 WHERE id = $4`,
		status, message, result, id)
	return err
}

// SetSent marks a command as sent.
func (r *CommandRepo) SetSent(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE command_logs SET status = 'SENT' WHERE id = $1`, id)
	return err
}

// GetByDeviceID returns command history for a device.
func (r *CommandRepo) GetByDeviceID(ctx context.Context, deviceID uuid.UUID, limit int) ([]models.CommandLog, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, device_id, command_type, payload, status, message, result, created_at, completed_at
		 FROM command_logs
		 WHERE device_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`, deviceID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.CommandLog
	for rows.Next() {
		var c models.CommandLog
		if err := rows.Scan(&c.ID, &c.DeviceID, &c.CommandType, &c.Payload,
			&c.Status, &c.Message, &c.Result, &c.CreatedAt, &c.CompletedAt); err != nil {
			return nil, err
		}
		logs = append(logs, c)
	}
	return logs, rows.Err()
}

// GetByID returns a single command log.
func (r *CommandRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.CommandLog, error) {
	var c models.CommandLog
	err := r.pool.QueryRow(ctx,
		`SELECT id, device_id, command_type, payload, status, message, result, created_at, completed_at
		 FROM command_logs WHERE id = $1`, id,
	).Scan(&c.ID, &c.DeviceID, &c.CommandType, &c.Payload,
		&c.Status, &c.Message, &c.Result, &c.CreatedAt, &c.CompletedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// MarkTimeouts marks commands that have been SENT but not completed within the threshold as TIMEOUT.
func (r *CommandRepo) MarkTimeouts(ctx context.Context, threshold time.Duration) (int64, error) {
	res, err := r.pool.Exec(ctx,
		`UPDATE command_logs SET status = 'TIMEOUT', completed_at = NOW()
		 WHERE status = 'SENT' AND created_at < $1`,
		time.Now().Add(-threshold))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yuwanadev/mdm-backend/internal/models"
)

type StatusRepo struct {
	pool *pgxpool.Pool
}

func NewStatusRepo(pool *pgxpool.Pool) *StatusRepo {
	return &StatusRepo{pool: pool}
}

// Upsert inserts or updates the device status (one row per device).
func (r *StatusRepo) Upsert(ctx context.Context, s *models.DeviceStatus) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO device_status (device_id, battery, temperature, battery_health, battery_status, 
		    battery_technology, battery_voltage, ram_usage, storage_total, storage_used, 
		    app_version, network_info, foreground_app, network_strength, location, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW())
		 ON CONFLICT (device_id) DO UPDATE SET
		    battery = COALESCE(EXCLUDED.battery, device_status.battery),
		    temperature = COALESCE(EXCLUDED.temperature, device_status.temperature),
		    battery_health = COALESCE(EXCLUDED.battery_health, device_status.battery_health),
		    battery_status = COALESCE(EXCLUDED.battery_status, device_status.battery_status),
		    battery_technology = COALESCE(EXCLUDED.battery_technology, device_status.battery_technology),
		    battery_voltage = COALESCE(EXCLUDED.battery_voltage, device_status.battery_voltage),
		    ram_usage = COALESCE(EXCLUDED.ram_usage, device_status.ram_usage),
		    storage_total = COALESCE(EXCLUDED.storage_total, device_status.storage_total),
		    storage_used = COALESCE(EXCLUDED.storage_used, device_status.storage_used),
		    app_version = COALESCE(EXCLUDED.app_version, device_status.app_version),
		    network_info = COALESCE(EXCLUDED.network_info, device_status.network_info),
		    foreground_app = COALESCE(EXCLUDED.foreground_app, device_status.foreground_app),
		    network_strength = COALESCE(EXCLUDED.network_strength, device_status.network_strength),
		    location = COALESCE(EXCLUDED.location, device_status.location),
		    updated_at = NOW()`,
		s.DeviceID, s.Battery, s.Temperature, s.BatteryHealth, s.BatteryStatus,
		s.BatteryTechnology, s.BatteryVoltage, s.RAMUsage, s.StorageTotal, s.StorageUsed,
		s.AppVersion, s.NetworkInfo, s.ForegroundApp, s.NetworkStrength, s.Location,
	)
	return err
}

// GetByDeviceID returns the latest status for a device.
func (r *StatusRepo) GetByDeviceID(ctx context.Context, deviceID uuid.UUID) (*models.DeviceStatus, error) {
	var s models.DeviceStatus
	err := r.pool.QueryRow(ctx,
		`SELECT device_id, battery, temperature, battery_health, battery_status,
		        battery_technology, battery_voltage, ram_usage, storage_total, storage_used,
		        app_version, network_info, foreground_app, network_strength, location, updated_at
		 FROM device_status WHERE device_id = $1`, deviceID,
	).Scan(&s.DeviceID, &s.Battery, &s.Temperature, &s.BatteryHealth, &s.BatteryStatus,
		&s.BatteryTechnology, &s.BatteryVoltage, &s.RAMUsage, &s.StorageTotal, &s.StorageUsed,
		&s.AppVersion, &s.NetworkInfo, &s.ForegroundApp, &s.NetworkStrength, &s.Location, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

-- +goose Up
ALTER TABLE device_status ADD COLUMN battery_health TEXT;
ALTER TABLE device_status ADD COLUMN battery_status TEXT;
ALTER TABLE device_status ADD COLUMN battery_technology TEXT;
ALTER TABLE device_status ADD COLUMN battery_voltage INTEGER;
ALTER TABLE device_status ADD COLUMN network_strength INTEGER;

-- +goose Down
ALTER TABLE device_status DROP COLUMN battery_health;
ALTER TABLE device_status DROP COLUMN battery_status;
ALTER TABLE device_status DROP COLUMN battery_technology;
ALTER TABLE device_status DROP COLUMN battery_voltage;
ALTER TABLE device_status DROP COLUMN network_strength;

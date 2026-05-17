-- +goose Up
ALTER TABLE device_status ADD COLUMN network_info JSONB;
ALTER TABLE device_status ADD COLUMN foreground_app TEXT;

-- +goose Down
ALTER TABLE device_status DROP COLUMN IF EXISTS network_info;
ALTER TABLE device_status DROP COLUMN IF EXISTS foreground_app;

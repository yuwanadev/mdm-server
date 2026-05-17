-- +goose Up
ALTER TABLE device_status ADD COLUMN IF NOT EXISTS location JSONB;

-- +goose Down
ALTER TABLE device_status DROP COLUMN IF EXISTS location;

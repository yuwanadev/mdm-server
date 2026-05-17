-- +goose Up
ALTER TABLE command_logs ADD COLUMN IF NOT EXISTS message TEXT;

-- +goose Down
ALTER TABLE command_logs DROP COLUMN IF EXISTS message;

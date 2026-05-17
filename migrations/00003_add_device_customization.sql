-- +goose Up
ALTER TABLE devices ADD COLUMN label TEXT NOT NULL DEFAULT '';
ALTER TABLE devices ADD COLUMN notes TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE devices DROP COLUMN label;
ALTER TABLE devices DROP COLUMN notes;

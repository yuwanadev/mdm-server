-- +goose Up
-- +goose StatementBegin
ALTER TABLE devices ADD COLUMN agent_version TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE devices DROP COLUMN IF EXISTS agent_version;
-- +goose StatementEnd

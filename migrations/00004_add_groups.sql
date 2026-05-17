-- +goose Up
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

ALTER TABLE devices ADD COLUMN group_id UUID REFERENCES groups(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE devices DROP COLUMN group_id;
DROP TABLE groups;

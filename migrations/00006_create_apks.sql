-- +goose Up
CREATE TABLE IF NOT EXISTS apks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_name TEXT NOT NULL,
    version_name TEXT NOT NULL,
    version_code INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS apks;

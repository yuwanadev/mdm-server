-- +goose Up

CREATE TABLE device_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID REFERENCES devices(id) ON DELETE CASCADE,
    account_name TEXT NOT NULL,
    account_type TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(device_id, account_name, account_type)
);

CREATE INDEX idx_device_accounts_device_id ON device_accounts(device_id);

-- +goose Down
DROP TABLE IF EXISTS device_accounts;

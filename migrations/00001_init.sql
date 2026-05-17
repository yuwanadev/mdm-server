-- +goose Up

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users (single admin)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Devices
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_name TEXT NOT NULL,
    device_model TEXT,
    android_version TEXT,
    token_hash TEXT UNIQUE NOT NULL,
    is_online BOOLEAN DEFAULT FALSE,
    last_seen TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Device Status (latest snapshot per device)
CREATE TABLE device_status (
    device_id UUID PRIMARY KEY REFERENCES devices(id) ON DELETE CASCADE,
    battery INTEGER,
    temperature REAL,
    ram_usage INTEGER,
    storage_total INTEGER,
    storage_used INTEGER,
    app_version TEXT,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Command Logs
CREATE TABLE command_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID REFERENCES devices(id) ON DELETE CASCADE,
    command_type TEXT NOT NULL,
    payload JSONB,
    status TEXT DEFAULT 'PENDING',
    result JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- Indexes
CREATE INDEX idx_devices_is_online ON devices(is_online);
CREATE INDEX idx_command_logs_device_id ON command_logs(device_id);
CREATE INDEX idx_command_logs_status ON command_logs(status);

-- +goose Down
DROP TABLE IF EXISTS command_logs;
DROP TABLE IF EXISTS device_status;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS users;

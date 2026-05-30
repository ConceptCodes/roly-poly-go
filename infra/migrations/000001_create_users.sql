-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE user_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    api_key VARCHAR(67) UNIQUE NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL
);

CREATE INDEX idx_user_models_deleted_at ON user_models(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS user_models CASCADE;

-- +goose Up
CREATE TABLE option_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    label VARCHAR(255) NOT NULL,
    poll_id UUID NOT NULL REFERENCES poll_models(id),
    votes INTEGER NOT NULL DEFAULT 0,
    UNIQUE(poll_id, label)
);

CREATE INDEX idx_option_models_deleted_at ON option_models(deleted_at);
CREATE INDEX idx_option_models_poll_id ON option_models(poll_id);

-- +goose Down
DROP TABLE IF EXISTS option_models CASCADE;

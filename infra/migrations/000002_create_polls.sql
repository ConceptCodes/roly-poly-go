-- +goose Up
CREATE TABLE poll_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    title VARCHAR(255) NOT NULL,
    description VARCHAR(100),
    closed TIMESTAMPTZ,
    public BOOLEAN NOT NULL DEFAULT true,
    allow_multiple_votes BOOLEAN NOT NULL DEFAULT false,
    user_id UUID NOT NULL REFERENCES user_models(id)
);

CREATE INDEX idx_poll_models_deleted_at ON poll_models(deleted_at);
CREATE INDEX idx_poll_models_user_id ON poll_models(user_id);

-- +goose Down
DROP TABLE IF EXISTS poll_models CASCADE;

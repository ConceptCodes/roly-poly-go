-- +goose Up
CREATE TABLE vote_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id UUID NOT NULL REFERENCES user_models(id),
    option_id UUID NOT NULL REFERENCES option_models(id),
    poll_id UUID NOT NULL REFERENCES poll_models(id),
    UNIQUE(user_id, option_id)
);

CREATE INDEX idx_vote_models_deleted_at ON vote_models(deleted_at);
CREATE INDEX idx_vote_models_user_id ON vote_models(user_id);
CREATE INDEX idx_vote_models_poll_id ON vote_models(poll_id);

-- +goose Down
DROP TABLE IF EXISTS vote_models CASCADE;

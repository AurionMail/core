-- +goose Up
CREATE TABLE identity_private_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identity_id UUID NOT NULL REFERENCES identities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    encrypted_private_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (identity_id, user_id)
);

CREATE INDEX idx_private_keys_user_id ON identity_private_keys (user_id);

-- +goose Down
DROP TABLE IF EXISTS identity_private_keys;
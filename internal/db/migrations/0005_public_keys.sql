-- +goose Up
CREATE TABLE identity_public_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identity_id UUID NOT NULL REFERENCES identities(id) ON DELETE CASCADE,
    wkd_hash TEXT NOT NULL,
    armored_key TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS identity_public_keys;

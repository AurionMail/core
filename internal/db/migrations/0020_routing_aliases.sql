-- +goose Up
CREATE TABLE routing_aliases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_email TEXT UNIQUE NOT NULL,
    target_identity_id UUID NOT NULL REFERENCES identities(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS routing_aliases;

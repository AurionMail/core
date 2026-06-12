-- +goose Up
CREATE TABLE routing_catchall (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    domain TEXT NOT NULL UNIQUE,
    target_identity_id UUID NOT NULL REFERENCES identities(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS routing_catchall;

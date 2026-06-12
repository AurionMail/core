-- +goose Up
CREATE TABLE identity_members (
    identity_id UUID NOT NULL REFERENCES identities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (identity_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS identity_members;

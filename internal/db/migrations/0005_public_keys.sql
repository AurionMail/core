-- +goose Up
CREATE TABLE public_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    armored_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_primary BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_public_keys_email ON public_keys (email);
CREATE INDEX idx_public_keys_user_id ON public_keys (user_id);

-- +goose Down
DROP TABLE IF EXISTS public_keys;

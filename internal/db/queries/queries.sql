-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: InsertPublicKey :one
INSERT INTO public_keys (user_id, email, wkd_hash, armored_key, is_primary)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPublicKeyByWKDHash :one
SELECT *
FROM public_keys
WHERE wkd_hash = $1
LIMIT 1;


-- name: GetPrimaryPublicKey :one
SELECT * FROM public_keys
WHERE email = $1 AND is_primary = TRUE
LIMIT 1;

-- name: InsertPrivateKey :one
INSERT INTO private_keys (user_id, armored_encrypted_key)
VALUES ($1, $2)
RETURNING *;

-- name: GetLatestPrivateKey :one
SELECT * FROM private_keys
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions
WHERE token = $1;

-- name: GetPrimaryPublicKeyByEmail :one
SELECT * FROM public_keys
WHERE email = $1 AND is_primary = TRUE
LIMIT 1;

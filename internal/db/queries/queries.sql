-- name: CreateUser :one
INSERT INTO users (email, password_hash, salt_server, salt_client)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;


-- name: CreateSession :one
INSERT INTO sessions (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions
WHERE token = $1;

-- name: CreateIdentity :one
INSERT INTO identities (email, type)
VALUES ($1, $2)
RETURNING *;

-- name: GetIdentityByEmail :one
SELECT *
FROM identities
WHERE email = $1
LIMIT 1;

-- name: GetIdentityByID :one
SELECT *
FROM identities
WHERE id = $1;

-- name: AddIdentityMember :exec
INSERT INTO identity_members (identity_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveIdentityMember :exec
DELETE FROM identity_members
WHERE identity_id = $1 AND user_id = $2;

-- name: ListIdentityMembers :many
SELECT users.*
FROM identity_members
JOIN users ON users.id = identity_members.user_id
WHERE identity_members.identity_id = $1;

-- name: InsertIdentityPublicKey :one
INSERT INTO identity_public_keys (identity_id, armored_key, wkd_hash, is_active)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetIdentityPublicKeyByWKDHash :one
SELECT *
FROM identity_public_keys
WHERE wkd_hash = $1
LIMIT 1;

-- name: GetActiveIdentityPublicKeys :many
SELECT *
FROM identity_public_keys
WHERE identity_id = $1 AND is_active = TRUE
ORDER BY created_at DESC;

-- name: DeactivateAllIdentityPublicKeys :exec
UPDATE identity_public_keys
SET is_active = FALSE
WHERE identity_id = $1;

-- name: InsertIdentityPrivateKey :one
INSERT INTO identity_private_keys (identity_id, user_id, encrypted_private_key)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetIdentityPrivateKey :one
SELECT *
FROM identity_private_keys
WHERE identity_id = $1 AND user_id = $2
LIMIT 1;

-- name: DeleteIdentityPrivateKey :exec
DELETE FROM identity_private_keys
WHERE identity_id = $1 AND user_id = $2;



-- name: CreateRoutingCatchall :one
INSERT INTO routing_catchall (domain, target_identity_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetRoutingCatchall :one
SELECT *
FROM routing_catchall
WHERE domain = $1
LIMIT 1;

-- name: DeleteRoutingCatchall :exec
DELETE FROM routing_catchall
WHERE domain = $1;

-- name: ResolveIdentityByEmail :one
SELECT i.*
FROM identities i
WHERE i.email = $1
LIMIT 1;

-- name: ListIdentitiesForUser :many
SELECT i.*
FROM identity_members m
JOIN identities i ON i.id = m.identity_id
WHERE m.user_id = $1;

-- name: ListMembersForIdentity :many
SELECT u.*
FROM identity_members m
JOIN users u ON u.id = m.user_id
WHERE m.identity_id = $1;

-- name: GetCatchallIdentity :one
SELECT i.*
FROM routing_catchall c
JOIN identities i ON i.id = c.identity_id
WHERE c.domain = $1;

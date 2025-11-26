-- name: CreateUserWithPassword :one
INSERT INTO users (id, created_at, updated_at, name, password_hash)
VALUES($1,$2,$3,$4,$5)
RETURNING *;


-- name: GetUserByName :one
SELECT * FROM users WHERE name = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $1 , updated_at = $2
WHERE id = $3;

-- name: CreateApiKey :one
INSERT INTO api_keys (id,created_at, updated_at,user_id, key_hash,name,expires_at)
VALUES($1,$2,$3,$4,$5,$6,$7)
RETURNING *;

-- name: GetApiKeyFromHash :one
SELECT * FROM api_keys WHERE key_hash = $1;

-- name: UpdateApiKeyLastUsed :exec
UPDATE api_keys SET last_used_at = $1 WHERE id = $2;

-- name: DeleteApiKey :exec
DELETE FROM api_keys WHERE id = $1 AND user_id = $2;

-- name: GetApiKeysForUsers :many
SELECT * FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC;



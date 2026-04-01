-- name: CreateAPIKey :one
INSERT INTO api_keys (org_id, name, key_prefix, key_hash, created_by, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, org_id, name, key_prefix, created_by, created_at, expires_at;

-- name: ListAPIKeysByOrg :many
SELECT id, org_id, name, key_prefix, created_by, created_at, expires_at, last_used_at
FROM api_keys
WHERE org_id = $1 AND revoked_at IS NULL
ORDER BY created_at DESC;

-- name: RevokeAPIKey :exec
UPDATE api_keys SET revoked_at = now()
WHERE id = $1 AND org_id = $2 AND revoked_at IS NULL;

-- name: GetAPIKeyByHash :one
SELECT id, org_id, name, key_prefix, key_hash, created_by, created_at, expires_at, last_used_at, revoked_at
FROM api_keys
WHERE key_hash = $1 AND revoked_at IS NULL;

-- name: TouchAPIKeyLastUsed :exec
UPDATE api_keys SET last_used_at = now() WHERE id = $1;

-- name: InsertRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, rotated_from)
VALUES ($1, $2, $3, $4, $5);

-- name: GetRefreshToken :one
SELECT user_id, token_hash, expires_at, revoked_at
FROM refresh_tokens
WHERE id = $1;

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;

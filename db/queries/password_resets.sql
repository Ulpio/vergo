-- name: CreatePasswordResetToken :exec
INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3);

-- name: GetPasswordResetByHash :one
SELECT id, user_id, token_hash, expires_at, used_at
FROM password_reset_tokens
WHERE token_hash = $1 AND used_at IS NULL;

-- name: MarkPasswordResetUsed :exec
UPDATE password_reset_tokens SET used_at = now() WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2 WHERE id = $1;

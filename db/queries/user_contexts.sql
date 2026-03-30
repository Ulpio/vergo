-- name: GetActiveOrg :one
SELECT org_id
FROM user_contexts
WHERE user_id = $1;

-- name: UpsertActiveOrg :exec
INSERT INTO user_contexts (user_id, org_id, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (user_id) DO UPDATE
SET org_id = $2, updated_at = NOW();

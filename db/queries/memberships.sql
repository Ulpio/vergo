-- name: UpsertMember :exec
INSERT INTO memberships (org_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (org_id, user_id) DO UPDATE SET role = EXCLUDED.role;

-- name: UpdateMemberRole :execresult
UPDATE memberships
SET role = $3
WHERE org_id = $1 AND user_id = $2;

-- name: DeleteMember :exec
DELETE FROM memberships
WHERE org_id = $1 AND user_id = $2;

-- name: GetMemberRole :one
SELECT role
FROM memberships
WHERE org_id = $1 AND user_id = $2;

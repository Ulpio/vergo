-- name: InsertOrg :exec
INSERT INTO organizations (id, name, owner_user_id, created_at)
VALUES ($1, $2, $3, $4);

-- name: GetOrg :one
SELECT id, name, owner_user_id, created_at
FROM organizations
WHERE id = $1;

-- name: DeleteOrg :exec
DELETE FROM organizations
WHERE id = $1;

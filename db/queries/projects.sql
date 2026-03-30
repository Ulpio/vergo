-- name: ListProjects :many
SELECT id, org_id, name, description, created_by, created_at, updated_at
FROM projects
WHERE org_id = $1
ORDER BY created_at ASC;

-- name: InsertProject :one
INSERT INTO projects (id, org_id, name, description, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id, org_id, name, description, created_by, created_at, updated_at;

-- name: GetProject :one
SELECT id, org_id, name, description, created_by, created_at, updated_at
FROM projects
WHERE id = $1 AND org_id = $2;

-- name: UpdateProject :one
UPDATE projects
SET name = COALESCE(NULLIF($3, ''), name),
    description = $4,
    updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING id, org_id, name, description, created_by, created_at, updated_at;

-- name: DeleteProject :execresult
DELETE FROM projects
WHERE id = $1 AND org_id = $2;

-- name: ListFiles :many
SELECT id, org_id, uploaded_by, bucket, object_key, size_bytes, content_type, created_at, metadata
FROM files
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: InsertFile :one
INSERT INTO files (id, org_id, uploaded_by, bucket, object_key, size_bytes, content_type, created_at, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, org_id, uploaded_by, bucket, object_key, size_bytes, content_type, created_at, metadata;

-- name: GetFile :one
SELECT id, org_id, uploaded_by, bucket, object_key, size_bytes, content_type, created_at, metadata
FROM files
WHERE id = $1 AND org_id = $2;

-- name: DeleteFile :execresult
DELETE FROM files
WHERE id = $1 AND org_id = $2;

-- name: InsertAuditLog :exec
INSERT INTO audit_logs (org_id, actor_id, action, entity, entity_id, metadata, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListAuditLogs :many
SELECT org_id, actor_id, action, entity, entity_id,
       COALESCE(metadata, '{}') AS metadata,
       created_at
FROM audit_logs
WHERE org_id = $1
  AND ($2::text IS NULL OR actor_id = $2)
  AND ($3::text IS NULL OR action = $3)
  AND ($4::text IS NULL OR entity = $4)
  AND ($5::timestamptz IS NULL OR created_at >= $5)
  AND ($6::timestamptz IS NULL OR created_at <= $6)
ORDER BY created_at DESC
LIMIT $7 OFFSET $8;

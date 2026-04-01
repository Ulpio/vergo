-- name: InsertAuditLog :exec
INSERT INTO audit_logs (org_id, actor_id, action, entity, entity_id, metadata, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListAuditLogs :many
SELECT org_id, actor_id, action, entity, entity_id,
       COALESCE(metadata, '{}') AS metadata,
       created_at
FROM audit_logs
WHERE org_id = @org_id
  AND (CAST(sqlc.narg('filter_actor_id') AS text) IS NULL OR actor_id = CAST(sqlc.narg('filter_actor_id') AS text))
  AND (CAST(sqlc.narg('filter_action') AS text) IS NULL OR action = CAST(sqlc.narg('filter_action') AS text))
  AND (CAST(sqlc.narg('filter_entity') AS text) IS NULL OR entity = CAST(sqlc.narg('filter_entity') AS text))
  AND (CAST(sqlc.narg('filter_since') AS timestamptz) IS NULL OR created_at >= CAST(sqlc.narg('filter_since') AS timestamptz))
  AND (CAST(sqlc.narg('filter_until') AS timestamptz) IS NULL OR created_at <= CAST(sqlc.narg('filter_until') AS timestamptz))
ORDER BY created_at DESC
LIMIT @query_limit OFFSET @query_offset;

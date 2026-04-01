-- name: CreateWebhookEndpoint :one
INSERT INTO webhook_endpoints (org_id, url, secret, events)
VALUES ($1, $2, $3, $4)
RETURNING id, org_id, url, events, active, created_at, updated_at;

-- name: ListWebhookEndpoints :many
SELECT id, org_id, url, events, active, created_at, updated_at
FROM webhook_endpoints
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: UpdateWebhookEndpoint :exec
UPDATE webhook_endpoints
SET url = $3, events = $4, active = $5, updated_at = now()
WHERE id = $1 AND org_id = $2;

-- name: GetWebhookEndpoint :one
SELECT id, org_id, url, secret, events, active, created_at, updated_at
FROM webhook_endpoints
WHERE id = $1 AND org_id = $2;

-- name: GetActiveEndpointsForEvent :many
SELECT id, org_id, url, secret, events
FROM webhook_endpoints
WHERE org_id = $1 AND active AND $2 = ANY(events);

-- name: InsertDelivery :one
INSERT INTO webhook_deliveries (endpoint_id, event, payload, next_retry)
VALUES ($1, $2, $3, now())
RETURNING id;

-- name: GetPendingDeliveries :many
SELECT d.id, d.endpoint_id, d.event, d.payload, d.attempts,
       e.url, e.secret
FROM webhook_deliveries d
JOIN webhook_endpoints e ON e.id = d.endpoint_id
WHERE NOT d.delivered AND d.attempts < 5 AND d.next_retry <= now()
ORDER BY d.next_retry
LIMIT $1;

-- name: MarkDelivered :exec
UPDATE webhook_deliveries
SET delivered = true, status_code = $2, response = $3, attempts = attempts + 1
WHERE id = $1;

-- name: MarkRetry :exec
UPDATE webhook_deliveries
SET attempts = attempts + 1, status_code = $2, response = $3, next_retry = $4
WHERE id = $1;

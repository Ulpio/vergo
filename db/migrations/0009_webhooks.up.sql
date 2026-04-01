CREATE TABLE webhook_endpoints (
    id         TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    org_id     TEXT NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    url        TEXT NOT NULL,
    secret     TEXT NOT NULL,
    events     TEXT[] NOT NULL DEFAULT '{}',
    active     BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_webhook_endpoints_org ON webhook_endpoints(org_id) WHERE active;

CREATE TABLE webhook_deliveries (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    endpoint_id TEXT NOT NULL REFERENCES webhook_endpoints(id) ON DELETE CASCADE,
    event       TEXT NOT NULL,
    payload     JSONB NOT NULL,
    status_code INT,
    response    TEXT,
    attempts    INT NOT NULL DEFAULT 0,
    next_retry  TIMESTAMPTZ,
    delivered   BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_webhook_deliveries_pending ON webhook_deliveries(next_retry)
    WHERE NOT delivered AND attempts < 5;

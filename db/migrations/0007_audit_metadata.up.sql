-- Add metadata column for before/after diff tracking
ALTER TABLE audit_logs
ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';

-- Index for querying metadata fields
CREATE INDEX IF NOT EXISTS idx_audit_metadata
ON audit_logs USING gin (metadata);

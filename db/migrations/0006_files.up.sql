CREATE TABLE IF NOT EXISTS files (
  id           TEXT PRIMARY KEY,                       
  org_id       TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  uploaded_by  TEXT NOT NULL REFERENCES users(id) ON DELETE SET NULL,
  bucket       TEXT NOT NULL,
  object_key   TEXT NOT NULL,
  size_bytes   BIGINT,
  content_type TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  metadata     JSONB
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_files_bucket_key ON files(bucket, object_key);
CREATE INDEX IF NOT EXISTS idx_files_org ON files(org_id);
CREATE INDEX IF NOT EXISTS idx_files_uploader ON files(uploaded_by);

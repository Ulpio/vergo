package file

import "time"

type File struct {
	ID          string    `json:"id"`
	OrgID       string    `json:"org_id"`
	UploadedBy  string    `json:"uploaded_by"`
	Bucket      string    `json:"bucket"`
	ObjectKey   string    `json:"object_key"`
	SizeBytes   *int64    `json:"size_bytes,omitempty"`
	ContentType string    `json:"content_type,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Metadata    any       `json:"metadata,omitempty"` // map[string]any serializado em JSONB
}

package handlers

// ErrorResponse is the standard error envelope.
// @Description Standard error response
type ErrorResponse struct {
	Error string `json:"error" example:"not_found"`
}

// ErrorDetailResponse is an error with optional detail.
// @Description Error response with detail
type ErrorDetailResponse struct {
	Error  string `json:"error" example:"create_fail"`
	Detail string `json:"detail,omitempty" example:"duplicate key"`
}

// AuthResponse is returned on signup and login.
// @Description Authentication response with tokens
type AuthResponse struct {
	User         AuthUser `json:"user"`
	AccessToken  string   `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string   `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

// AuthUser is the user object inside auth responses.
type AuthUser struct {
	ID    string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email string `json:"email" example:"user@example.com"`
}

// TokenResponse is returned on token refresh.
// @Description Refreshed token pair
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

// MeResponse wraps the /me endpoint response.
// @Description Current user info
type MeResponse struct {
	User MeUser `json:"user"`
}

// MeUser is the user object inside /me response.
type MeUser struct {
	ID    string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email string `json:"email" example:"user@example.com"`
	OrgID string `json:"org_id,omitempty" example:"org-uuid"`
	Role  string `json:"role,omitempty" example:"admin"`
}

// ContextResponse is the active org context.
// @Description Active organization context
type ContextResponse struct {
	OrgID string `json:"org_id" example:"org-uuid"`
	Role  string `json:"role" example:"admin"`
}

// PresignPutResponse is the presigned PUT URL response.
// @Description Presigned upload URL
type PresignPutResponse struct {
	Method  string            `json:"method" example:"PUT"`
	URL     string            `json:"url" example:"https://s3.amazonaws.com/bucket/key?..."`
	Headers map[string]string `json:"headers"`
}

// PresignGetResponse is the presigned GET URL response.
// @Description Presigned download URL
type PresignGetResponse struct {
	Method string `json:"method" example:"GET"`
	URL    string `json:"url" example:"https://s3.amazonaws.com/bucket/key?..."`
}

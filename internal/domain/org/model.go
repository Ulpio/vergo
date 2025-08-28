package org

import "time"

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerUser string    `json:"owner_user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Membership struct {
	OrgID  string `json:"org_id"`
	UserID string `json:"user_id"`
	Role   string `json:"role"` // owner | admin | member
}

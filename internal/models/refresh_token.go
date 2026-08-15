package models

import "time"

type RefreshToken struct {
	ID        string     `json:"id,omitempty"`
	UserID    string     `json:"userId"`
	TokenID   string     `json:"-"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
}
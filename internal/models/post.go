package models

import "time"

type Post struct {
	ID      string `json:"id,omitempty"`
	Content string `json:"content"`
	UserID  string `json:"userId"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type CreatePost struct {
	Content string `json:"content" validate:"required,min=3,max=400"`
}
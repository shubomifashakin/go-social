package models

import "time"

type Comment struct {
	ID        string     `json:"id,omitempty"`
	UserID 	  string 	 `json:"userId"`
	PostID 	  string 	 `json:"postId"`
	Content   string     `json:"content"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type CreateComment struct {
	Content  string   `json:"content" validate:"required,min=3,max=400"`
}
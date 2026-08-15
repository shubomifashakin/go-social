package models

import "time"

type Follower struct {
	UserID    string     `json:"userId"`
	FollowerID    string     `json:"followerId"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
}
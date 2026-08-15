package models

import "time"

type User struct {
	ID        string     `json:"id,omitempty"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Email     string     `json:"email"`
	Password  string     `json:"-"`
	Username  string     `json:"username"`
	Role      string     `json:"role"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type UserSignup struct {
    FirstName string `json:"first_name" validate:"omitempty,min=3,max=100"`
    LastName  string `json:"last_name" validate:"omitempty,min=3,max=100"`
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=8"`
    Username  string `json:"username" validate:"required,min=3,max=10"`
}

type UserLogin struct {
    Password  string `json:"password" validate:"required"`
    Username  string `json:"username" validate:"required,min=3,max=10"`
}

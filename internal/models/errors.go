package models

import "errors"

var (
	ErrDuplicateEntry = errors.New("duplicate entry")
	ErrMissingField   = errors.New("missing required field")
	ErrNotFound       = errors.New("record not found")
	ErrForeignKey     = errors.New("referenced record does not exist")
)

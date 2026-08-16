package models

type PaginatedResponse[T any] struct {
	Data        []T    `json:"data"`
	HasNextPage bool   `json:"hasNextPage"`
	Next        string `json:"next"`
}

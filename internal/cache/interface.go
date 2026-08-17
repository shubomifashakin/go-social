package cache

import (
	"context"
	"time"
)

type CacheService interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	SetNx(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error
	GetJSON(ctx context.Context, key string, dest any) error
	IsRateLimited(ctx context.Context, key string, limit int64, window time.Duration) (bool, time.Time, error)
}
package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
    return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) SetNx(ctx context.Context, key string, value any, ttl time.Duration) error {
    return c.client.SetNX(ctx, key, value, ttl).Err()
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
    return c.client.Get(ctx, key).Result()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}

func (c *Cache) Close() error {
    return c.client.Close()
}

func (c *Cache) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *Cache) GetJSON(ctx context.Context, key string, dest any) error {
    raw, err := c.client.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(raw), dest)
}



func NewRedisClient(ctx context.Context,url string, clientName string) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: url,

		MaxRetries: 5,
		WriteTimeout: time.Second * 10,
		ReadTimeout: time.Second * 10,
		ClientName: clientName,
		MinIdleConns: 5,
		MaxIdleConns: 10,
		MaxActiveConns: 50,
		ConnMaxIdleTime: time.Minute * 5,
		ConnMaxLifetime: time.Minute * 30,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Cache{
		client: client,
	},nil
}
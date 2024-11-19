package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(host, port string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
		DB:   0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %v", err)
	}

	return &RedisClient{client: client}, nil
}

func (r *RedisClient) StoreSession(ctx context.Context, userID int64, token string) error {
	key := fmt.Sprintf("session:%d", userID)
	return r.client.Set(ctx, key, token, 24*time.Hour).Err()
}

func (r *RedisClient) GetSession(ctx context.Context, userID int64) (string, error) {
	key := fmt.Sprintf("session:%d", userID)
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) DeleteSession(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("session:%d", userID)
	return r.client.Del(ctx, key).Err()
}

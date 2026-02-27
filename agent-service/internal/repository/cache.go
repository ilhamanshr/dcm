package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:generate mockgen -destination=mocks/mock_cache.go -source=cache.go ICache
type ICache interface {
	Ping(ctx context.Context) error
	SetKey(ctx context.Context, key string, value string) error
	SetKeyWithExpire(ctx context.Context, key string, value string, expireInSecond time.Duration) error
	GetKey(ctx context.Context, key string) (string, error)
	DeleteKey(ctx context.Context, key string) error
	SetNX(ctx context.Context, key string, ttl time.Duration) (bool, error)
	PubSubConn(ctx context.Context, key string) *redis.PubSub
	Publish(ctx context.Context, key, message string) error
	Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Release(ctx context.Context, key string) error
}

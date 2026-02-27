package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func NewRedisHelper(config RedisConfig) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})
	return &Redis{
		client: client,
	}
}

func (r *Redis) Ping(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *Redis) SetKey(ctx context.Context, key string, value string) error {
	_, err := r.client.Set(ctx, key, value, 0).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *Redis) SetKeyWithExpire(ctx context.Context, key string, value string, expireInSecond time.Duration) error {
	_, err := r.client.Set(ctx, key, value, expireInSecond).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *Redis) GetKey(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *Redis) DeleteKey(ctx context.Context, key string) error {
	_, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *Redis) SetNX(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	success, err := r.client.SetNX(ctx, key, "lock", ttl).Result()
	return success, err
}

func (r *Redis) PubSubConn(ctx context.Context, key string) *redis.PubSub {
	return r.client.Subscribe(ctx, key)
}

func (r *Redis) Publish(ctx context.Context, key, message string) error {
	return r.client.Publish(ctx, key, message).Err()
}

func (r *Redis) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, key, "lock", ttl).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (r *Redis) Release(ctx context.Context, key string) error {
	_, err := r.client.Del(ctx, key).Result()
	return err
}

package helpers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"dx-api/app/facades"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

// GetRedis returns the shared Redis client (singleton)
func GetRedis() *redis.Client {
	redisOnce.Do(func() {
		host := facades.Config().GetString("cache.stores.redis.host", "127.0.0.1")
		port := facades.Config().GetString("cache.stores.redis.port", "6379")
		password := facades.Config().GetString("cache.stores.redis.password", "")
		db := facades.Config().GetInt("cache.stores.redis.database", 0)

		redisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Password: password,
			DB:       db,
		})
	})
	return redisClient
}

// RedisSet sets a key with TTL
func RedisSet(key string, value string, ttl time.Duration) error {
	ctx := context.Background()
	return GetRedis().Set(ctx, key, value, ttl).Err()
}

// RedisGet retrieves a value by key
func RedisGet(key string) (string, error) {
	ctx := context.Background()
	return GetRedis().Get(ctx, key).Result()
}

// RedisDel deletes a key
func RedisDel(key string) error {
	ctx := context.Background()
	return GetRedis().Del(ctx, key).Err()
}

// RedisPing checks Redis connectivity
func RedisPing() error {
	ctx := context.Background()
	return GetRedis().Ping(ctx).Err()
}

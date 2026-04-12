package helpers

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	redis_facades "github.com/goravel/redis/facades"
)

// GetRedis returns the shared Redis client from Goravel's redis facade.
// Uses the "default" connection defined in config/database.go.
//
// Previously this created its own *redis.Client singleton, which resulted
// in two Redis clients in-process (one for Goravel's cache/queue/session
// drivers, and one for our helpers). Consolidated here so that cache,
// queue, session, online tracking, and WebSocket pub/sub all share a
// single client with consistent configuration (TLS, cluster mode, etc.).
//
// In single-node mode (which Douxue uses), the UniversalClient returned
// by the facade is a *redis.Client. Cluster mode would require different
// plumbing.
func GetRedis() *redis.Client {
	client, err := redis_facades.Instance("default")
	if err != nil {
		panic("redis facade not initialized: " + err.Error())
	}
	if c, ok := client.(*redis.Client); ok {
		return c
	}
	panic("redis facade returned non-*redis.Client (cluster mode not supported in helper)")
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

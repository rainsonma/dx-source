package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CheckRateLimit checks if a request is within the rate limit using Redis sorted sets (sliding window).
// Returns true if the request is allowed, false if rate limited.
func CheckRateLimit(key string, limit int, windowSeconds int) (bool, error) {
	ctx := context.Background()
	rdb := GetRedis()
	now := time.Now().UnixMilli()
	windowStart := now - int64(windowSeconds)*1000

	pipe := rdb.Pipeline()

	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", windowStart))

	// Count remaining entries
	countCmd := pipe.ZCard(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	count := countCmd.Val()
	if count >= int64(limit) {
		return false, nil
	}

	// Add current request
	pipe2 := rdb.Pipeline()
	pipe2.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	pipe2.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)
	_, err = pipe2.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to record rate limit entry: %w", err)
	}

	return true, nil
}

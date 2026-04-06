package helpers

import "context"

// RedisSetAdd adds members to a Redis SET.
func RedisSetAdd(key string, members ...string) error {
	ctx := context.Background()
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	return GetRedis().SAdd(ctx, key, args...).Err()
}

// RedisSetRemove removes members from a Redis SET.
func RedisSetRemove(key string, members ...string) error {
	ctx := context.Background()
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	return GetRedis().SRem(ctx, key, args...).Err()
}

// RedisSetIsMember checks if a member exists in a Redis SET.
func RedisSetIsMember(key string, member string) (bool, error) {
	ctx := context.Background()
	return GetRedis().SIsMember(ctx, key, member).Result()
}

// RedisSetCard returns the number of members in a Redis SET.
func RedisSetCard(key string) (int64, error) {
	ctx := context.Background()
	return GetRedis().SCard(ctx, key).Result()
}

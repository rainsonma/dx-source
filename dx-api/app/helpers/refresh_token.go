package helpers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/goravel/framework/facades"
)

type RefreshTokenData struct {
	UserID string `json:"user_id"`
	Guard  string `json:"guard"`
	AuthID string `json:"auth_id"`
}

// GenerateRefreshToken returns a cryptographically random 64-char hex string.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// StoreRefreshToken stores a refresh token in Redis with TTL and adds it to the user index.
func StoreRefreshToken(token, userID, guard, authID string) error {
	ctx := context.Background()
	rdb := GetRedis()
	ttl := time.Duration(facades.Config().GetInt("refresh_token.ttl", 10080)) * time.Minute

	data, err := json.Marshal(RefreshTokenData{UserID: userID, Guard: guard, AuthID: authID})
	if err != nil {
		return fmt.Errorf("failed to marshal refresh token data: %w", err)
	}

	pipe := rdb.Pipeline()
	pipe.Set(ctx, "refresh:"+token, string(data), ttl)
	pipe.SAdd(ctx, fmt.Sprintf("user_refresh:%s:%s", userID, guard), token)
	pipe.Expire(ctx, fmt.Sprintf("user_refresh:%s:%s", userID, guard), ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}
	return nil
}

// LookupRefreshToken retrieves and validates a refresh token from Redis.
func LookupRefreshToken(token string) (*RefreshTokenData, error) {
	ctx := context.Background()
	val, err := GetRedis().Get(ctx, "refresh:"+token).Result()
	if err != nil {
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}

	var data RefreshTokenData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to parse refresh token data: %w", err)
	}
	return &data, nil
}

// DeleteRefreshToken removes a refresh token from Redis and the user index.
func DeleteRefreshToken(token, userID, guard string) error {
	ctx := context.Background()
	rdb := GetRedis()

	pipe := rdb.Pipeline()
	pipe.Del(ctx, "refresh:"+token)
	pipe.SRem(ctx, fmt.Sprintf("user_refresh:%s:%s", userID, guard), token)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

// DeleteUserRefreshTokens removes all refresh tokens for a user+guard combo.
func DeleteUserRefreshTokens(userID, guard string) error {
	ctx := context.Background()
	rdb := GetRedis()
	indexKey := fmt.Sprintf("user_refresh:%s:%s", userID, guard)

	tokens, err := rdb.SMembers(ctx, indexKey).Result()
	if err != nil {
		return fmt.Errorf("failed to list user refresh tokens: %w", err)
	}

	if len(tokens) == 0 {
		return nil
	}

	pipe := rdb.Pipeline()
	for _, t := range tokens {
		pipe.Del(ctx, "refresh:"+t)
	}
	pipe.Del(ctx, indexKey)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete user refresh tokens: %w", err)
	}
	return nil
}

package realtime

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// presenceTTL is the sliding expiration window for presence sets.
// Every Add call refreshes the TTL. If a topic has no activity for
// presenceTTL, the set is garbage-collected automatically.
const presenceTTL = 15 * time.Minute

// Presence tracks which users are subscribed to which topics, across all
// dx-api instances sharing the same Redis. Backed by per-topic Redis SETs.
type Presence struct {
	client *redis.Client
}

// NewPresence constructs a Presence tracker backed by the given Redis client.
func NewPresence(client *redis.Client) *Presence {
	return &Presence{client: client}
}

func presenceKey(topic string) string { return "presence:" + topic }

// Add inserts userID into the presence set for topic and refreshes the TTL.
func (p *Presence) Add(ctx context.Context, topic, userID string) error {
	key := presenceKey(topic)
	pipe := p.client.Pipeline()
	pipe.SAdd(ctx, key, userID)
	pipe.Expire(ctx, key, presenceTTL)
	_, err := pipe.Exec(ctx)
	return err
}

// Remove deletes userID from the presence set for topic.
func (p *Presence) Remove(ctx context.Context, topic, userID string) error {
	return p.client.SRem(ctx, presenceKey(topic), userID).Err()
}

// IsPresent returns true if userID is currently in the topic's presence set.
func (p *Presence) IsPresent(ctx context.Context, topic, userID string) (bool, error) {
	return p.client.SIsMember(ctx, presenceKey(topic), userID).Result()
}

// Members returns all userIDs currently in the topic's presence set.
// Returns an empty slice (not nil) for unknown topics.
func (p *Presence) Members(ctx context.Context, topic string) ([]string, error) {
	result, err := p.client.SMembers(ctx, presenceKey(topic)).Result()
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []string{}, nil
	}
	return result, nil
}

package realtime

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedisClient(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

func TestPresence_SetAddRemoveCheck(t *testing.T) {
	client, _ := newTestRedisClient(t)
	ctx := context.Background()
	p := NewPresence(client)

	if err := p.Add(ctx, "pk:abc", "alice"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := p.Add(ctx, "pk:abc", "bob"); err != nil {
		t.Fatalf("add: %v", err)
	}

	present, err := p.IsPresent(ctx, "pk:abc", "alice")
	if err != nil || !present {
		t.Errorf("alice should be present: err=%v present=%v", err, present)
	}

	present, err = p.IsPresent(ctx, "pk:abc", "carol")
	if err != nil || present {
		t.Errorf("carol should not be present: present=%v", present)
	}

	if err := p.Remove(ctx, "pk:abc", "alice"); err != nil {
		t.Fatalf("remove: %v", err)
	}

	present, err = p.IsPresent(ctx, "pk:abc", "alice")
	if err != nil || present {
		t.Errorf("alice should be removed: present=%v", present)
	}
}

func TestPresence_Members(t *testing.T) {
	client, _ := newTestRedisClient(t)
	ctx := context.Background()
	p := NewPresence(client)

	_ = p.Add(ctx, "group:xyz", "alice")
	_ = p.Add(ctx, "group:xyz", "bob")
	_ = p.Add(ctx, "group:xyz", "carol")

	members, err := p.Members(ctx, "group:xyz")
	if err != nil {
		t.Fatalf("members: %v", err)
	}
	if len(members) != 3 {
		t.Errorf("want 3 members, got %d: %v", len(members), members)
	}
	found := map[string]bool{}
	for _, m := range members {
		found[m] = true
	}
	for _, expected := range []string{"alice", "bob", "carol"} {
		if !found[expected] {
			t.Errorf("missing %s in members", expected)
		}
	}
}

func TestPresence_AddRefreshesExpiry(t *testing.T) {
	client, mr := newTestRedisClient(t)
	ctx := context.Background()
	p := NewPresence(client)

	_ = p.Add(ctx, "pk:abc", "alice")
	ttl := mr.TTL("presence:pk:abc")
	if ttl <= 0 {
		t.Errorf("expected positive TTL, got %v", ttl)
	}

	mr.FastForward(5 * time.Minute)

	_ = p.Add(ctx, "pk:abc", "bob")
	ttl2 := mr.TTL("presence:pk:abc")
	if ttl2 <= 0 {
		t.Errorf("expected TTL refresh, got %v", ttl2)
	}
}

func TestPresence_EmptyKeyReturnsEmptyMembers(t *testing.T) {
	client, _ := newTestRedisClient(t)
	ctx := context.Background()
	p := NewPresence(client)

	members, err := p.Members(ctx, "group:nonexistent")
	if err != nil {
		t.Fatalf("members: %v", err)
	}
	if len(members) != 0 {
		t.Errorf("want empty, got %v", members)
	}
}

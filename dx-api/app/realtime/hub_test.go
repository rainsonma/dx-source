package realtime

import (
	"context"
	"sync"
	"testing"
)

// fakePubSub is an in-memory PubSub used to test Hub logic in isolation
// without bringing up Redis.
type fakePubSub struct {
	mu        sync.Mutex
	subs      map[string][]chan Event
	closed    bool
	publishes []struct {
		Topic string
		Event Event
	}
}

func newFakePubSub() *fakePubSub {
	return &fakePubSub{subs: make(map[string][]chan Event)}
}

func (p *fakePubSub) Publish(ctx context.Context, topic string, event Event) error {
	p.mu.Lock()
	p.publishes = append(p.publishes, struct {
		Topic string
		Event Event
	}{Topic: topic, Event: event})
	subs := append([]chan Event{}, p.subs[topic]...)
	p.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
	return nil
}

func (p *fakePubSub) Subscribe(topic string) (<-chan Event, func()) {
	ch := make(chan Event, 16)
	p.mu.Lock()
	p.subs[topic] = append(p.subs[topic], ch)
	p.mu.Unlock()

	unsub := func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		for i, c := range p.subs[topic] {
			if c == ch {
				p.subs[topic] = append(p.subs[topic][:i], p.subs[topic][i+1:]...)
				break
			}
		}
		close(ch)
	}
	return ch, unsub
}

func (p *fakePubSub) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	return nil
}

// allowAllAuthorizer is an Authorizer that approves every subscribe.
func allowAllAuthorizer() *Authorizer {
	return &Authorizer{
		isPkParticipant: func(ctx context.Context, userID, pkID string) (bool, error) { return true, nil },
		isGroupMember:   func(ctx context.Context, userID, groupID string) (bool, error) { return true, nil },
	}
}

func TestHub_NewHubConstructs(t *testing.T) {
	ps := newFakePubSub()
	hub := NewHub(ps, nil, allowAllAuthorizer())
	if hub == nil {
		t.Fatal("nil hub")
	}
	if len(hub.clients) != 0 {
		t.Errorf("expected empty clients, got %d", len(hub.clients))
	}
}

func TestHub_SubscribeAuthorizedAddsClient(t *testing.T) {
	ps := newFakePubSub()
	hub := NewHub(ps, nil, allowAllAuthorizer())

	c := &Client{
		hub:    hub,
		userID: "alice",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	hub.mu.Lock()
	hub.clients[c] = struct{}{}
	hub.mu.Unlock()

	err := hub.subscribe(context.Background(), c, "user:alice")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	hub.mu.RLock()
	_, inTopic := hub.topics["user:alice"][c]
	hub.mu.RUnlock()
	if !inTopic {
		t.Error("client not added to topic")
	}
	if _, has := c.topics["user:alice"]; !has {
		t.Error("client.topics missing topic")
	}
}

func TestHub_SubscribeUnauthorized(t *testing.T) {
	ps := newFakePubSub()
	authorizer := &Authorizer{
		isPkParticipant: func(ctx context.Context, userID, pkID string) (bool, error) { return false, nil },
		isGroupMember:   func(ctx context.Context, userID, groupID string) (bool, error) { return false, nil },
	}
	hub := NewHub(ps, nil, authorizer)

	c := &Client{
		hub:    hub,
		userID: "bob",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	hub.mu.Lock()
	hub.clients[c] = struct{}{}
	hub.mu.Unlock()

	err := hub.subscribe(context.Background(), c, "pk:abc")
	if err == nil {
		t.Fatal("expected authorization error")
	}

	hub.mu.RLock()
	_, inTopic := hub.topics["pk:abc"][c]
	hub.mu.RUnlock()
	if inTopic {
		t.Error("unauthorized client should not be in topic")
	}
}

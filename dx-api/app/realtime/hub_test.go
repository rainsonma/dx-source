package realtime

import (
	"context"
	"sync"
	"testing"
	"time"
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

// --- Task 4.3: Hub subscribe ref-counting ---

func TestHub_FirstSubscribeTriggersFakePubSubSubscribe(t *testing.T) {
	ps := newFakePubSub()
	hub := NewHub(ps, nil, allowAllAuthorizer())

	c1 := &Client{
		hub:    hub,
		userID: "alice",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	c2 := &Client{
		hub:    hub,
		userID: "bob",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	hub.mu.Lock()
	hub.clients[c1] = struct{}{}
	hub.clients[c2] = struct{}{}
	hub.mu.Unlock()

	_ = hub.subscribe(context.Background(), c1, "user:alice")
	_ = hub.subscribe(context.Background(), c2, "user:bob")

	ps.mu.Lock()
	subCount := len(ps.subs)
	ps.mu.Unlock()

	if subCount != 2 {
		t.Errorf("want 2 pubsub subscribes (one per user topic), got %d", subCount)
	}
}

func TestHub_SecondLocalSubscribeReusesPubSub(t *testing.T) {
	ps := newFakePubSub()
	hub := NewHub(ps, nil, allowAllAuthorizer())

	c1 := &Client{
		hub:    hub,
		userID: "alice",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	c2 := &Client{
		hub:    hub,
		userID: "bob",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	hub.mu.Lock()
	hub.clients[c1] = struct{}{}
	hub.clients[c2] = struct{}{}
	hub.mu.Unlock()

	_ = hub.subscribe(context.Background(), c1, "group:grp1")
	_ = hub.subscribe(context.Background(), c2, "group:grp1")

	ps.mu.Lock()
	grpSubs := len(ps.subs["group:grp1"])
	ps.mu.Unlock()

	if grpSubs != 1 {
		t.Errorf("want 1 pubsub channel for group:grp1 (shared), got %d", grpSubs)
	}
}

func TestHub_LastUnsubscribeClearsTopic(t *testing.T) {
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

	_ = hub.subscribe(context.Background(), c, "user:alice")
	hub.unsubscribe(c, "user:alice")

	hub.mu.RLock()
	_, topicExists := hub.topics["user:alice"]
	hub.mu.RUnlock()
	if topicExists {
		t.Error("topic should be removed after last unsubscribe")
	}
}

func TestHub_IdempotentSubscribe(t *testing.T) {
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

	_ = hub.subscribe(context.Background(), c, "user:alice")
	err := hub.subscribe(context.Background(), c, "user:alice") // second time
	if err != nil {
		t.Errorf("idempotent subscribe should succeed: %v", err)
	}

	hub.mu.RLock()
	n := len(hub.topics["user:alice"])
	hub.mu.RUnlock()
	if n != 1 {
		t.Errorf("want 1 client in topic, got %d", n)
	}
}

// --- Task 4.4: Group auto-events ---

func TestHub_SubscribeToGroupPublishesRoomMemberJoined(t *testing.T) {
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

	_ = hub.subscribe(context.Background(), c, "group:grp1")

	ps.mu.Lock()
	defer ps.mu.Unlock()
	found := false
	for _, pub := range ps.publishes {
		if pub.Topic == "group:grp1" && pub.Event.Type == "room_member_joined" {
			data, _ := pub.Event.Data.(map[string]string)
			if data["user_id"] == "alice" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("did not find room_member_joined for alice")
	}
}

func TestHub_UnsubscribeFromGroupPublishesRoomMemberLeft(t *testing.T) {
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

	_ = hub.subscribe(context.Background(), c, "group:grp1")
	hub.unsubscribe(c, "group:grp1")

	ps.mu.Lock()
	defer ps.mu.Unlock()
	found := false
	for _, pub := range ps.publishes {
		if pub.Topic == "group:grp1" && pub.Event.Type == "room_member_left" {
			data, _ := pub.Event.Data.(map[string]string)
			if data["user_id"] == "alice" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("did not find room_member_left for alice")
	}
}

func TestHub_SubscribeToUserTopicDoesNotPublishRoomEvent(t *testing.T) {
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

	_ = hub.subscribe(context.Background(), c, "user:alice")

	ps.mu.Lock()
	defer ps.mu.Unlock()
	for _, pub := range ps.publishes {
		if pub.Event.Type == "room_member_joined" || pub.Event.Type == "room_member_left" {
			t.Errorf("unexpected room event on user topic: %+v", pub)
		}
	}
}

// --- Task 4.5: Slow consumer kick ---

func TestClient_EnqueueFullChannelTriggersKick(t *testing.T) {
	// Create a fake client with a send queue of capacity 1. Fill it, then
	// enqueue another — the second should trigger slow-consumer kick.
	// The kick goroutine will call conn.Close which panics on a nil conn;
	// we guard the test via recover() to verify the closed flag transition.
	hub := NewHub(newFakePubSub(), nil, allowAllAuthorizer())
	c := &Client{
		hub:    hub,
		userID: "slow",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 1),
	}
	c.send <- Envelope{Op: OpEvent} // fill the channel

	done := make(chan struct{})
	go func() {
		defer func() { _ = recover(); close(done) }()
		c.enqueue(Envelope{Op: OpEvent, Type: "second"})
	}()
	<-done

	// Allow the kickSlowConsumer goroutine a moment to run
	time.Sleep(10 * time.Millisecond)

	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if !closed {
		t.Error("client should be marked closed after enqueue on full channel")
	}
}

// --- Task 4.6: Shutdown flag ---

func TestHub_ShutdownSetsFlag(t *testing.T) {
	hub := NewHub(newFakePubSub(), nil, allowAllAuthorizer())

	if hub.IsShuttingDown() {
		t.Error("shouldnt be shutting down before Shutdown called")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := hub.Shutdown(ctx); err != nil {
		t.Errorf("shutdown: %v", err)
	}

	if !hub.IsShuttingDown() {
		t.Error("should be shutting down after Shutdown")
	}
}

func TestClient_SessionReplacedKickDetected(t *testing.T) {
	hub := NewHub(newFakePubSub(), nil, allowAllAuthorizer())

	c := &Client{
		hub:    hub,
		userID: "alice",
		topics: map[string]struct{}{"user:alice:kick": {}},
		send:   make(chan Envelope, 16),
	}

	done := make(chan struct{})
	go func() {
		defer func() { _ = recover(); close(done) }()
		c.enqueue(Envelope{
			Op:    OpEvent,
			Topic: "user:alice:kick",
			Type:  "session_replaced",
		})
	}()
	<-done

	time.Sleep(10 * time.Millisecond)

	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if !closed {
		t.Error("client should be marked closed after session_replaced kick")
	}
}

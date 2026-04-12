package realtime

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"dx-api/app/consts"

	"github.com/coder/websocket"
	"github.com/goravel/framework/facades"
)

// Hub is the local client registry and topic router for one dx-api process.
// It knows which Clients are connected, which topics each is subscribed to,
// and how to route incoming PubSub events to the matching local clients.
type Hub struct {
	pubsub     PubSub
	presence   *Presence
	authorizer *Authorizer

	mu         sync.RWMutex
	clients    map[*Client]struct{}            // all attached clients
	topics     map[string]map[*Client]struct{} // topic -> subscribed clients
	unsubs     map[string]func()               // topic -> pubsub unsubscribe fn
	shutdownFg atomic.Bool
}

// defaultHub holds the package-level Hub atomically so concurrent
// SetDefaultHub and DefaultHub calls are race-free. Set at bootstrap via
// SetDefaultHub; read via DefaultHub.
var defaultHub atomic.Pointer[Hub]

// SetDefaultHub wires the package-level Hub used by the WS controller.
// Intended to be called once from bootstrap. Safe for concurrent readers.
func SetDefaultHub(h *Hub) {
	defaultHub.Store(h)
}

// DefaultHub returns the currently-wired Hub, or nil if SetDefaultHub has
// not been called. The WSController calls this on each upgrade.
func DefaultHub() *Hub {
	return defaultHub.Load()
}

// NewHub constructs a Hub wired to the given PubSub, Presence tracker, and
// Authorizer. All three are required (Presence may be nil for tests).
func NewHub(ps PubSub, presence *Presence, authorizer *Authorizer) *Hub {
	return &Hub{
		pubsub:     ps,
		presence:   presence,
		authorizer: authorizer,
		clients:    make(map[*Client]struct{}),
		topics:     make(map[string]map[*Client]struct{}),
		unsubs:     make(map[string]func()),
	}
}

// IsShuttingDown returns true after Shutdown has been called. The WS
// controller checks this to reject new upgrades during graceful shutdown.
func (h *Hub) IsShuttingDown() bool { return h.shutdownFg.Load() }

// Attach takes ownership of the WebSocket connection for the given user,
// starts the client's read/write loops, and blocks until the connection
// terminates. Returns any error from the read loop (including normal EOF).
//
// IMPORTANT: ctx MUST be a detached context (context.Background() or
// similar), NOT the HTTP request context. Goravel's global Timeout
// middleware cancels the request context after http.request_timeout
// (30s default), which would kill long-lived WS connections.
func (h *Hub) Attach(ctx context.Context, userID string, conn *websocket.Conn) error {
	client := newClient(h, userID, conn)

	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			facades.Log().Errorf("client panic: user=%s err=%v\n%s", userID, r, debug.Stack())
		}
		h.detach(client)
	}()

	return client.Serve(ctx)
}

// detach removes a client from the hub, unsubscribing it from all topics.
// Called from Attach's defer and from subscribe slow-consumer handling.
func (h *Hub) detach(c *Client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; !ok {
		h.mu.Unlock()
		return
	}
	delete(h.clients, c)

	// Collect topics to unsubscribe outside the lock
	topicsToRemove := make([]string, 0, len(c.topics))
	for topic := range c.topics {
		topicsToRemove = append(topicsToRemove, topic)
	}
	h.mu.Unlock()

	for _, topic := range topicsToRemove {
		h.unsubscribe(c, topic)
	}
}

// Shutdown closes all attached clients gracefully and terminates the hub.
// Called from the Goravel terminating hook on SIGTERM.
func (h *Hub) Shutdown(ctx context.Context) error {
	h.shutdownFg.Store(true)

	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	// Send server_shutdown close frame to each client
	for _, c := range clients {
		go func(client *Client) {
			if client.conn != nil {
				_ = client.conn.Close(4002, "server_shutdown")
			}
		}(c)
	}

	// Wait up to 5 seconds for clean drain
	deadline := time.After(5 * time.Second)
	for {
		h.mu.RLock()
		n := len(h.clients)
		h.mu.RUnlock()
		if n == 0 {
			break
		}
		select {
		case <-deadline:
			// Force-close stragglers
			h.mu.RLock()
			stragglers := make([]*Client, 0, len(h.clients))
			for c := range h.clients {
				stragglers = append(stragglers, c)
			}
			h.mu.RUnlock()
			for _, c := range stragglers {
				if c.conn != nil {
					_ = c.conn.CloseNow()
				}
			}
			goto done
		case <-ctx.Done():
			goto done
		case <-time.After(50 * time.Millisecond):
		}
	}
done:

	// Close the pubsub loop
	_ = h.pubsub.Close()
	return nil
}

// subscribe authorizes and adds a client to a topic. First local subscriber
// triggers a PubSub subscribe. For group topics, publishes room_member_joined.
func (h *Hub) subscribe(ctx context.Context, c *Client, topic string) error {
	if err := h.authorizer.AuthorizeSubscribe(ctx, c.userID, topic); err != nil {
		return err
	}

	h.mu.Lock()
	firstLocal := h.topics[topic] == nil
	if firstLocal {
		h.topics[topic] = make(map[*Client]struct{})
	}
	// Idempotent: if already subscribed, no-op (still ack success)
	if _, alreadySubbed := h.topics[topic][c]; alreadySubbed {
		h.mu.Unlock()
		return nil
	}
	h.topics[topic][c] = struct{}{}
	c.addTopic(topic)
	h.mu.Unlock()

	if firstLocal {
		ch, unsub := h.pubsub.Subscribe(topic)
		h.mu.Lock()
		h.unsubs[topic] = unsub
		h.mu.Unlock()
		go h.fanout(topic, ch)
	}

	// Record presence (best-effort — failures don't block)
	if h.presence != nil {
		_ = h.presence.Add(ctx, topic, c.userID)
	}

	// Auto-publish room_member_joined for group topics
	if parsed, err := ParseTopic(topic); err == nil && parsed.Kind == KindGroup {
		_ = h.pubsub.Publish(ctx, topic, Event{
			Type: "room_member_joined",
			Data: map[string]string{"user_id": c.userID},
		})
	}

	return nil
}

// unsubscribe removes a client from a topic. Last local unsubscriber
// triggers a PubSub unsubscribe. For group topics, publishes room_member_left.
func (h *Hub) unsubscribe(c *Client, topic string) {
	h.mu.Lock()
	if h.topics[topic] == nil {
		h.mu.Unlock()
		return
	}
	if _, ok := h.topics[topic][c]; !ok {
		h.mu.Unlock()
		return
	}
	delete(h.topics[topic], c)
	c.removeTopic(topic)

	var unsub func()
	lastLocal := len(h.topics[topic]) == 0
	if lastLocal {
		unsub = h.unsubs[topic]
		delete(h.topics, topic)
		delete(h.unsubs, topic)
	}
	h.mu.Unlock()

	ctx := context.Background()

	if h.presence != nil {
		_ = h.presence.Remove(ctx, topic, c.userID)
	}

	// Auto-publish room_member_left for group topics
	if parsed, err := ParseTopic(topic); err == nil && parsed.Kind == KindGroup {
		_ = h.pubsub.Publish(ctx, topic, Event{
			Type: "room_member_left",
			Data: map[string]string{"user_id": c.userID},
		})
	}

	if unsub != nil {
		unsub()
	}
}

// fanout delivers events from a PubSub subscribe channel to all currently
// subscribed clients on that topic. Runs as a goroutine per subscribed topic.
func (h *Hub) fanout(topic string, ch <-chan Event) {
	defer func() {
		if r := recover(); r != nil {
			facades.Log().Errorf("fanout panic: topic=%s err=%v\n%s", topic, r, debug.Stack())
		}
	}()

	for ev := range ch {
		h.mu.RLock()
		subs := make([]*Client, 0, len(h.topics[topic]))
		for c := range h.topics[topic] {
			subs = append(subs, c)
		}
		h.mu.RUnlock()

		env := Envelope{
			Op:    OpEvent,
			Topic: topic,
			Type:  ev.Type,
			Data:  ev.Data,
		}
		for _, c := range subs {
			c.enqueue(env)
		}
	}
}

// kickSlowConsumer is called when a client's send channel is full. It
// detaches the client and closes the connection with code 4003.
func (h *Hub) kickSlowConsumer(c *Client) {
	go func() {
		if c.conn != nil {
			_ = c.conn.Close(4003, "slow_consumer")
		}
	}()
}

// ackError builds an ack envelope from a realtimeError or generic error.
func ackError(id string, err error) Envelope {
	ok := false
	env := Envelope{Op: OpAck, ID: id, OK: &ok}
	var rtErr realtimeError
	if errors.As(err, &rtErr) {
		env.Code = rtErr.Code
		env.Message = rtErr.Message
	} else {
		env.Code = consts.CodeInternalError
		env.Message = fmt.Sprintf("internal: %v", err)
	}
	return env
}

package realtime

import (
	"context"
	"sync"
	"time"

	"dx-api/app/consts"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	sendQueueCapacity = 32
	pingInterval      = 25 * time.Second
	pingTimeout       = 5 * time.Second
	readLimitBytes    = 4096
)

// Client owns one WebSocket connection and its read/write loops.
type Client struct {
	hub    *Hub
	userID string
	conn   *websocket.Conn

	send   chan Envelope       // buffered; drained by writeLoop
	topics map[string]struct{} // topics this client subscribes to

	mu     sync.Mutex
	closed bool
}

func newClient(h *Hub, userID string, conn *websocket.Conn) *Client {
	return &Client{
		hub:    h,
		userID: userID,
		conn:   conn,
		send:   make(chan Envelope, sendQueueCapacity),
		topics: make(map[string]struct{}),
	}
}

// Serve runs the client's read and write loops. Blocks until the read loop
// returns (on disconnect, protocol error, or explicit close).
func (c *Client) Serve(ctx context.Context) error {
	c.conn.SetReadLimit(readLimitBytes)

	writeCtx, cancelWrite := context.WithCancel(ctx)
	defer cancelWrite()

	// Auto-subscribe the client to its own user:{id}:kick topic for Issue C
	_ = c.hub.subscribe(ctx, c, UserKickTopic(c.userID))

	writeDone := make(chan struct{})
	go func() {
		defer close(writeDone)
		c.writeLoop(writeCtx)
		// If writeLoop exited before readLoop (e.g., ping failure), force the
		// connection closed so readLoop's wsjson.Read unblocks. Otherwise the
		// client goroutine leaks and the hub keeps a stale entry. This is a
		// no-op if readLoop already initiated the close via cancelWrite.
		_ = c.conn.Close(websocket.StatusNormalClosure, "write loop exited")
	}()

	err := c.readLoop(ctx)

	// readLoop returned (disconnect or error) — signal writeLoop to exit
	cancelWrite()
	<-writeDone
	return err
}

// readLoop reads inbound envelopes from the WebSocket and dispatches them
// to the hub's subscribe/unsubscribe handlers.
func (c *Client) readLoop(ctx context.Context) error {
	for {
		var env Envelope
		if err := wsjson.Read(ctx, c.conn, &env); err != nil {
			return err
		}
		switch env.Op {
		case OpSubscribe:
			if err := c.hub.subscribe(ctx, c, env.Topic); err != nil {
				c.enqueue(ackError(env.ID, err))
				continue
			}
			ok := true
			c.enqueue(Envelope{Op: OpAck, ID: env.ID, OK: &ok})
		case OpUnsubscribe:
			c.hub.unsubscribe(c, env.Topic)
			ok := true
			c.enqueue(Envelope{Op: OpAck, ID: env.ID, OK: &ok})
		default:
			c.enqueue(Envelope{
				Op:      OpError,
				Code:    consts.CodeUnknownOp,
				Message: "unknown op: " + string(env.Op),
			})
		}
	}
}

// writeLoop drains the send channel and writes frames. Also sends periodic
// server-initiated pings.
func (c *Client) writeLoop(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case env, ok := <-c.send:
			if !ok {
				return
			}
			if err := wsjson.Write(ctx, c.conn, env); err != nil {
				return
			}
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// enqueue tries to add an envelope to the send channel without blocking.
// If the channel is full, the client is considered too slow and kicked.
func (c *Client) enqueue(env Envelope) {
	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if closed {
		return
	}

	// Issue C: detect session_replaced kick on the user's kick topic and
	// force-close the connection with WS code 4001 instead of delivering
	// the event. The client's onclose handler redirects to signin.
	if env.Op == OpEvent && env.Type == "session_replaced" {
		if parsed, err := ParseTopic(env.Topic); err == nil && parsed.Kind == KindUserKick && parsed.ID == c.userID {
			c.mu.Lock()
			if !c.closed {
				c.closed = true
				c.mu.Unlock()
				go func() {
					if c.conn != nil {
						_ = c.conn.Close(4001, "session_replaced")
					}
				}()
				return
			}
			c.mu.Unlock()
			return
		}
	}

	select {
	case c.send <- env:
	default:
		c.mu.Lock()
		if !c.closed {
			c.closed = true
			c.mu.Unlock()
			c.hub.kickSlowConsumer(c)
			return
		}
		c.mu.Unlock()
	}
}

// addTopic records a topic subscription on the client. Must be called under
// the hub's lock.
func (c *Client) addTopic(topic string) {
	c.topics[topic] = struct{}{}
}

// removeTopic removes a topic subscription from the client. Must be called
// under the hub's lock.
func (c *Client) removeTopic(topic string) {
	delete(c.topics, topic)
}

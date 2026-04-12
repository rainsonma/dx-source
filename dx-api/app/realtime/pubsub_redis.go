package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

// RedisPubSub is the production implementation of PubSub, backed by Redis
// PUBLISH/SUBSCRIBE. One long-lived subscribe connection per instance
// handles all topics, with per-topic ref-counting to minimize Redis traffic.
type RedisPubSub struct {
	client *redis.Client

	mu     sync.Mutex
	locals map[string]map[chan Event]struct{} // topic -> set of local channels
	refs   map[string]int                     // topic -> subscribe ref count
	pubsub *redis.PubSub                      // Redis subscribe handle

	ctx    context.Context
	cancel context.CancelFunc
	closed bool
	done   chan struct{}
}

// NewRedisPubSub constructs and starts the subscribe loop goroutine.
// The passed context is used as the parent for the internal loop; cancel it
// (or call Close) to terminate.
func NewRedisPubSub(parent context.Context, client *redis.Client) *RedisPubSub {
	ctx, cancel := context.WithCancel(parent)
	ps := &RedisPubSub{
		client: client,
		locals: make(map[string]map[chan Event]struct{}),
		refs:   make(map[string]int),
		pubsub: client.Subscribe(ctx),
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	go ps.loop()
	return ps
}

// Publish serializes event and publishes it to topic.
func (p *RedisPubSub) Publish(ctx context.Context, topic string, event Event) error {
	payload, err := json.Marshal(struct {
		Type string `json:"type"`
		Data any    `json:"data"`
	}{Type: event.Type, Data: event.Data})
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return p.client.Publish(ctx, topic, payload).Err()
}

// Subscribe registers a local channel for topic and returns it plus an
// unsubscribe function. First local subscriber triggers a Redis SUBSCRIBE.
func (p *RedisPubSub) Subscribe(topic string) (<-chan Event, func()) {
	ch := make(chan Event, 16)

	p.mu.Lock()
	if p.locals[topic] == nil {
		p.locals[topic] = make(map[chan Event]struct{})
	}
	p.locals[topic][ch] = struct{}{}
	firstLocal := p.refs[topic] == 0
	p.refs[topic]++
	p.mu.Unlock()

	if firstLocal {
		_ = p.pubsub.Subscribe(p.ctx, topic)
	}

	unsubscribe := func() {
		p.mu.Lock()
		if _, ok := p.locals[topic][ch]; !ok {
			p.mu.Unlock()
			return
		}
		delete(p.locals[topic], ch)
		p.refs[topic]--
		lastLocal := p.refs[topic] == 0
		if lastLocal {
			delete(p.locals, topic)
			delete(p.refs, topic)
		}
		// Close under the lock so it cannot race with a concurrent send
		// in loop(). The dispatch loop also holds p.mu, so we're mutually
		// exclusive.
		close(ch)
		p.mu.Unlock()

		if lastLocal {
			_ = p.pubsub.Unsubscribe(p.ctx, topic)
		}
	}

	return ch, unsubscribe
}

// Close terminates the subscribe loop and closes the Redis subscription.
// Idempotent.
func (p *RedisPubSub) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.mu.Unlock()

	p.cancel()
	_ = p.pubsub.Close()
	<-p.done
	return nil
}

// loop reads messages from the Redis subscribe connection and dispatches
// each to the registered local channels for that topic.
func (p *RedisPubSub) loop() {
	defer close(p.done)

	ch := p.pubsub.Channel()
	for msg := range ch {
		var wire struct {
			Type string `json:"type"`
			Data any    `json:"data"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &wire); err != nil {
			continue
		}
		event := Event{Type: wire.Type, Data: wire.Data}

		// Dispatch under the lock so unsubscribe() cannot close a channel
		// while we're about to send to it. The send is non-blocking
		// (select + default drop-on-full), so lock contention is bounded
		// to microseconds per message.
		p.mu.Lock()
		for c := range p.locals[msg.Channel] {
			select {
			case c <- event:
			default:
				// Drop on full — Hub handles slow-consumer policy.
			}
		}
		p.mu.Unlock()
	}
}

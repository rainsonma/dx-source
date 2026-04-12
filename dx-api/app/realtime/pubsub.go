package realtime

import (
	"context"
	"sync/atomic"
)

// PubSub is the seam between the Hub and the backing transport. Production
// uses RedisPubSub; tests use a fake declared in the test files.
type PubSub interface {
	// Publish sends event on topic. At-most-once delivery semantics.
	Publish(ctx context.Context, topic string, event Event) error

	// Subscribe returns a receive channel of events delivered on topic.
	// Subscribe is infallible by design: the local ref-counted registration
	// is a map insert, and the network subscribe is best-effort and retried
	// by the implementation's background loop. Delivery is at-most-once.
	// The returned unsubscribe function must be called to tear down the
	// subscription and release resources. Subscribe uses the PubSub's
	// constructor-time context for its lifetime, not a per-call context,
	// since subscriptions outlive the request that initiates them.
	Subscribe(topic string) (<-chan Event, func())

	// Close terminates any background goroutines and releases resources.
	Close() error
}

// defaultPubSub holds the package-level PubSub atomically so concurrent
// SetDefault and Publish calls are race-free. Set at bootstrap via
// SetDefault; read by the package-level Publish function.
var defaultPubSub atomic.Pointer[PubSub]

// SetDefault wires the package-level PubSub used by Publish. It is
// intended to be called once from bootstrap with a production RedisPubSub,
// or from TestMain / test setup with a fake. Safe for concurrent readers.
func SetDefault(ps PubSub) {
	defaultPubSub.Store(&ps)
}

// Default returns the currently-wired PubSub, or nil if SetDefault has
// not been called. Exposed so tests can inspect state; normal code should
// call Publish instead.
func Default() PubSub {
	p := defaultPubSub.Load()
	if p == nil {
		return nil
	}
	return *p
}

// Publish is the single function every service call site uses to emit an
// event on a topic. It delegates to the package-level PubSub. Returns
// ErrNotInitialized if SetDefault has not been called (bootstrap-order bug).
func Publish(ctx context.Context, topic string, event Event) error {
	ps := Default()
	if ps == nil {
		return ErrNotInitialized
	}
	return ps.Publish(ctx, topic, event)
}

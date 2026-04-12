package realtime

import "context"

// PubSub is the seam between the Hub and the backing transport. Production
// uses RedisPubSub; tests use a fake declared in the test file.
type PubSub interface {
	// Publish sends event on topic. At-most-once delivery semantics.
	Publish(ctx context.Context, topic string, event Event) error

	// Subscribe returns a receive channel of events delivered on topic.
	// The returned unsubscribe function must be called to tear down the
	// subscription and release resources.
	Subscribe(topic string) (<-chan Event, func())

	// Close terminates any background goroutines and releases resources.
	Close() error
}

// Default is the package-level PubSub used by Publish. Set at bootstrap.
var Default PubSub

// Publish is the single function every service call site uses to emit an
// event on a topic. It delegates to Default. Returns ErrNotInitialized if
// Default has not been wired (bootstrap-order bug).
func Publish(ctx context.Context, topic string, event Event) error {
	if Default == nil {
		return ErrNotInitialized
	}
	return Default.Publish(ctx, topic, event)
}

package realtime

import "errors"

// ErrNotInitialized is returned from Publish when the package-level Default
// PubSub hasn't been set (a bootstrap-order bug).
var ErrNotInitialized = errors.New("realtime: default pubsub not initialized")

// realtimeError carries both a consts.Code* integer and a user-facing message.
// It's returned from authorization and protocol validation so the hub can
// construct ack/error envelopes with the right fields.
type realtimeError struct {
	Code    int
	Message string
}

func (e realtimeError) Error() string { return e.Message }

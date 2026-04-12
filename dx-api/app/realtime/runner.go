package realtime

import (
	"context"
	"time"
)

// RealtimeRunner implements foundation.Runner so the Hub receives a clean
// Shutdown call when Goravel handles SIGTERM. Run is a no-op because the
// hub is request-driven (attached to WebSocket upgrades, not a blocking
// event loop).
type RealtimeRunner struct{}

func (r *RealtimeRunner) Signature() string { return "realtime" }

func (r *RealtimeRunner) ShouldRun() bool { return true }

func (r *RealtimeRunner) Run() error { return nil }

func (r *RealtimeRunner) Shutdown() error {
	hub := DefaultHub()
	if hub == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return hub.Shutdown(ctx)
}

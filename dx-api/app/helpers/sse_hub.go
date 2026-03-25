package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// SSEConnection wraps a single client SSE connection.
type SSEConnection struct {
	w       http.ResponseWriter
	flusher http.Flusher
	done    chan struct{}
}

// SSEHub manages group SSE connections and broadcasting.
type SSEHub struct {
	mu    sync.RWMutex
	conns map[string]map[string]*SSEConnection // groupID -> userID -> conn
}

// GroupSSEHub is the global SSE hub instance for group events.
var GroupSSEHub = &SSEHub{
	conns: make(map[string]map[string]*SSEConnection),
}

// Register adds a connection for a user in a group.
func (h *SSEHub) Register(groupID, userID string, w http.ResponseWriter) *SSEConnection {
	flusher, _ := w.(http.Flusher)
	conn := &SSEConnection{w: w, flusher: flusher, done: make(chan struct{})}

	h.mu.Lock()
	if h.conns[groupID] == nil {
		h.conns[groupID] = make(map[string]*SSEConnection)
	}
	// Close existing connection if any
	if old, ok := h.conns[groupID][userID]; ok {
		close(old.done)
	}
	h.conns[groupID][userID] = conn
	h.mu.Unlock()

	return conn
}

// Unregister removes a connection.
func (h *SSEHub) Unregister(groupID, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if group, ok := h.conns[groupID]; ok {
		delete(group, userID)
		if len(group) == 0 {
			delete(h.conns, groupID)
		}
	}
}

// Broadcast sends an event to all connected members of a group.
func (h *SSEHub) Broadcast(groupID, event string, data any) {
	jsonBytes, _ := json.Marshal(data)

	h.mu.RLock()
	defer h.mu.RUnlock()

	if group, ok := h.conns[groupID]; ok {
		for _, conn := range group {
			fmt.Fprintf(conn.w, "event: %s\ndata: %s\n\n", event, jsonBytes)
			if conn.flusher != nil {
				conn.flusher.Flush()
			}
		}
	}
}

// SendHeartbeat sends a comment line as keepalive.
func (conn *SSEConnection) SendHeartbeat() error {
	_, err := fmt.Fprintf(conn.w, ": heartbeat\n\n")
	if err != nil {
		return err
	}
	if conn.flusher != nil {
		conn.flusher.Flush()
	}
	return nil
}

// Done returns a channel that closes when the connection should end.
func (conn *SSEConnection) Done() <-chan struct{} {
	return conn.done
}

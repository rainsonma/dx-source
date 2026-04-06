package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// SSEConnection wraps a single client SSE connection.
type SSEConnection struct {
	w       http.ResponseWriter
	flusher http.Flusher
	rc      *http.ResponseController
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

// Register adds a connection for a user in a group and broadcasts join event.
func (h *SSEHub) Register(groupID, userID string, w http.ResponseWriter) *SSEConnection {
	conn := NewSSEConnection(w)

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

	// Broadcast join event to all connections (including the new one)
	h.Broadcast(groupID, "room_member_joined", map[string]string{
		"user_id": userID,
	})

	return conn
}

// Unregister removes a connection only if it matches the given pointer.
// This prevents a replaced connection (via Register) from being deleted
// by the old connection's deferred cleanup.
func (h *SSEHub) Unregister(groupID, userID string, conn *SSEConnection) {
	h.mu.Lock()
	current, exists := h.conns[groupID][userID]
	if exists && current == conn {
		delete(h.conns[groupID], userID)
		if len(h.conns[groupID]) == 0 {
			delete(h.conns, groupID)
		}
	}
	h.mu.Unlock()

	// Only broadcast leave if we actually removed the connection
	if exists && current == conn {
		h.Broadcast(groupID, "room_member_left", map[string]string{
			"user_id": userID,
		})
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

// ConnectedUserIDs returns the list of user IDs currently connected to a group.
func (h *SSEHub) ConnectedUserIDs(groupID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	group, ok := h.conns[groupID]
	if !ok {
		return nil
	}
	ids := make([]string, 0, len(group))
	for uid := range group {
		ids = append(ids, uid)
	}
	return ids
}

// SendHeartbeat sends a comment line as keepalive and extends the write deadline.
func (conn *SSEConnection) SendHeartbeat() error {
	// Extend write deadline before writing — prevents the global request_timeout from killing SSE connections.
	if conn.rc != nil {
		_ = conn.rc.SetWriteDeadline(time.Now().Add(60 * time.Second))
	}
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

// NewSSEConnection creates an SSEConnection from an http.ResponseWriter.
func NewSSEConnection(w http.ResponseWriter) *SSEConnection {
	flusher, _ := w.(http.Flusher)
	rc := http.NewResponseController(w)
	// Extend initial write deadline so the connection isn't killed before the first heartbeat.
	_ = rc.SetWriteDeadline(time.Now().Add(60 * time.Second))
	return &SSEConnection{w: w, flusher: flusher, rc: rc, done: make(chan struct{})}
}

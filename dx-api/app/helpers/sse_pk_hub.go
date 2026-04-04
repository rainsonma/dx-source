package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// PkSSEHub manages SSE connections for PK matches.
type PkSSEHub struct {
	mu    sync.RWMutex
	conns map[string]map[string]*SSEConnection // pkID -> userID -> conn
}

// PkHub is the global SSE hub instance for PK events.
var PkHub = &PkSSEHub{
	conns: make(map[string]map[string]*SSEConnection),
}

// Register adds a connection for a user in a PK match.
// Replaces any existing connection for the same user (closes old done channel).
func (h *PkSSEHub) Register(pkID, userID string, w http.ResponseWriter) *SSEConnection {
	conn := NewSSEConnection(w)

	h.mu.Lock()
	if h.conns[pkID] == nil {
		h.conns[pkID] = make(map[string]*SSEConnection)
	}
	if old, ok := h.conns[pkID][userID]; ok {
		close(old.done)
	}
	h.conns[pkID][userID] = conn
	h.mu.Unlock()

	return conn
}

// Unregister removes a connection only if it matches the given pointer.
// This prevents a replaced connection (via Register) from being deleted
// by the old connection's deferred cleanup.
func (h *PkSSEHub) Unregister(pkID, userID string, conn *SSEConnection) {
	h.mu.Lock()
	current, exists := h.conns[pkID][userID]
	if exists && current == conn {
		delete(h.conns[pkID], userID)
		if len(h.conns[pkID]) == 0 {
			delete(h.conns, pkID)
		}
	}
	h.mu.Unlock()
}

// Broadcast sends an event to all connected participants of a PK match.
func (h *PkSSEHub) Broadcast(pkID, event string, data any) {
	jsonBytes, _ := json.Marshal(data)

	h.mu.RLock()
	defer h.mu.RUnlock()

	if pk, ok := h.conns[pkID]; ok {
		for _, conn := range pk {
			fmt.Fprintf(conn.w, "event: %s\ndata: %s\n\n", event, jsonBytes)
			if conn.flusher != nil {
				conn.flusher.Flush()
			}
		}
	}
}

// IsConnected checks whether a specific user is connected to a PK match.
func (h *PkSSEHub) IsConnected(pkID, userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if pk, ok := h.conns[pkID]; ok {
		_, exists := pk[userID]
		return exists
	}
	return false
}

package helpers

import (
	"encoding/json"
	"fmt"
	"sync"
)

// NotifyHub manages lightweight SSE connections for group detail notifications.
// Unlike SSEHub, it has no presence tracking — no broadcasts on connect/disconnect.
type NotifyHub struct {
	mu    sync.RWMutex
	conns map[string]map[string]*SSEConnection // groupID -> userID -> conn
}

// GroupNotifyHub is the global notification hub instance.
var GroupNotifyHub = &NotifyHub{
	conns: make(map[string]map[string]*SSEConnection),
}

// Register adds a notification connection for a user in a group.
// Replaces any existing connection for the same user (closes old done channel).
// No broadcast on connect.
func (h *NotifyHub) Register(groupID, userID string, conn *SSEConnection) {
	h.mu.Lock()
	if h.conns[groupID] == nil {
		h.conns[groupID] = make(map[string]*SSEConnection)
	}
	if old, ok := h.conns[groupID][userID]; ok {
		close(old.done)
	}
	h.conns[groupID][userID] = conn
	h.mu.Unlock()
}

// Unregister removes a notification connection only if it matches the given pointer.
// No broadcast on disconnect.
func (h *NotifyHub) Unregister(groupID, userID string, conn *SSEConnection) {
	h.mu.Lock()
	current, exists := h.conns[groupID][userID]
	if exists && current == conn {
		delete(h.conns[groupID], userID)
		if len(h.conns[groupID]) == 0 {
			delete(h.conns, groupID)
		}
	}
	h.mu.Unlock()
}

// Notify sends a group_updated event with the given scope to all connected members.
func (h *NotifyHub) Notify(groupID, scope string) {
	data, _ := json.Marshal(map[string]string{"scope": scope})

	h.mu.RLock()
	defer h.mu.RUnlock()

	if group, ok := h.conns[groupID]; ok {
		for _, conn := range group {
			fmt.Fprintf(conn.w, "event: group_updated\ndata: %s\n\n", data)
			if conn.flusher != nil {
				conn.flusher.Flush()
			}
		}
	}
}

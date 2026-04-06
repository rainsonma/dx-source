package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

const redisOnlineUsersKey = "online_users"

// UserSSEHub manages per-user SSE connections for global notifications.
type UserSSEHub struct {
	mu    sync.RWMutex
	conns map[string]*SSEConnection // userID -> conn
}

// UserHub is the global SSE hub instance for user-level events.
var UserHub = &UserSSEHub{
	conns: make(map[string]*SSEConnection),
}

// Register adds a connection for a user and marks them online in Redis.
func (h *UserSSEHub) Register(userID string, w http.ResponseWriter) *SSEConnection {
	conn := NewSSEConnection(w)

	h.mu.Lock()
	if old, ok := h.conns[userID]; ok {
		close(old.done)
	}
	h.conns[userID] = conn
	h.mu.Unlock()

	_ = RedisSetAdd(redisOnlineUsersKey, userID)

	return conn
}

// Unregister removes a connection and marks the user offline in Redis.
func (h *UserSSEHub) Unregister(userID string, conn *SSEConnection) {
	h.mu.Lock()
	current, exists := h.conns[userID]
	if exists && current == conn {
		delete(h.conns, userID)
	}
	h.mu.Unlock()

	if exists && current == conn {
		_ = RedisSetRemove(redisOnlineUsersKey, userID)
	}
}

// SendToUser sends an SSE event to a specific user.
func (h *UserSSEHub) SendToUser(userID, event string, data any) {
	h.mu.RLock()
	conn, ok := h.conns[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}
	// Send as generic message with type field (avoids Safari named-event bug)
	payload := map[string]any{"type": event, "payload": data}
	jsonBytes, _ := json.Marshal(payload)
	fmt.Fprintf(conn.w, "data: %s\n\n", jsonBytes)
	if conn.flusher != nil {
		conn.flusher.Flush()
	}
}

// IsOnline checks if a user is in the Redis online set.
func (h *UserSSEHub) IsOnline(userID string) bool {
	ok, _ := RedisSetIsMember(redisOnlineUsersKey, userID)
	return ok
}

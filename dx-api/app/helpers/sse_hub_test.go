package helpers

import (
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHub() *SSEHub {
	return &SSEHub{
		conns: make(map[string]map[string]*SSEConnection),
	}
}

func registerTestConn(hub *SSEHub, groupID, userID string) *SSEConnection {
	w := httptest.NewRecorder()
	return hub.Register(groupID, userID, w)
}

func TestConnectedUserIDs_Empty(t *testing.T) {
	hub := newTestHub()
	ids := hub.ConnectedUserIDs("group-1")
	assert.Nil(t, ids)
}

func TestConnectedUserIDs_SingleUser(t *testing.T) {
	hub := newTestHub()
	registerTestConn(hub, "group-1", "user-1")

	ids := hub.ConnectedUserIDs("group-1")
	assert.Equal(t, []string{"user-1"}, ids)
}

func TestConnectedUserIDs_MultipleUsers(t *testing.T) {
	hub := newTestHub()
	registerTestConn(hub, "group-1", "user-1")
	registerTestConn(hub, "group-1", "user-2")
	registerTestConn(hub, "group-1", "user-3")

	ids := hub.ConnectedUserIDs("group-1")
	sort.Strings(ids)
	assert.Equal(t, []string{"user-1", "user-2", "user-3"}, ids)
}

func TestConnectedUserIDs_DifferentGroups(t *testing.T) {
	hub := newTestHub()
	registerTestConn(hub, "group-1", "user-1")
	registerTestConn(hub, "group-2", "user-2")

	ids1 := hub.ConnectedUserIDs("group-1")
	ids2 := hub.ConnectedUserIDs("group-2")
	assert.Equal(t, []string{"user-1"}, ids1)
	assert.Equal(t, []string{"user-2"}, ids2)
}

func TestRegister_ReplacesExistingConnection(t *testing.T) {
	hub := newTestHub()
	conn1 := registerTestConn(hub, "group-1", "user-1")
	conn2 := registerTestConn(hub, "group-1", "user-1")

	// conn1 should be closed (Done channel)
	select {
	case <-conn1.Done():
		// expected
	default:
		t.Fatal("old connection should be closed")
	}

	// conn2 should still be active
	ids := hub.ConnectedUserIDs("group-1")
	assert.Equal(t, []string{"user-1"}, ids)
	_ = conn2
}

func TestUnregister_MatchingPointer(t *testing.T) {
	hub := newTestHub()
	conn := registerTestConn(hub, "group-1", "user-1")

	hub.Unregister("group-1", "user-1", conn)

	ids := hub.ConnectedUserIDs("group-1")
	assert.Nil(t, ids)
}

func TestUnregister_MismatchedPointer(t *testing.T) {
	hub := newTestHub()
	oldConn := registerTestConn(hub, "group-1", "user-1")
	// Replace with new connection
	_ = registerTestConn(hub, "group-1", "user-1")

	// Unregister with old pointer — should NOT remove the new connection
	hub.Unregister("group-1", "user-1", oldConn)

	ids := hub.ConnectedUserIDs("group-1")
	assert.Equal(t, []string{"user-1"}, ids)
}

func TestBroadcast_SendsToAllConnections(t *testing.T) {
	hub := newTestHub()

	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()

	// Manually build connections using the recorder's Flusher interface
	var f1, f2 Flusher
	if fv, ok := interface{}(w1).(Flusher); ok {
		f1 = fv
	}
	if fv, ok := interface{}(w2).(Flusher); ok {
		f2 = fv
	}

	hub.mu.Lock()
	hub.conns["group-1"] = map[string]*SSEConnection{
		"user-1": {w: w1, flusher: f1, done: make(chan struct{})},
		"user-2": {w: w2, flusher: f2, done: make(chan struct{})},
	}
	hub.mu.Unlock()

	hub.Broadcast("group-1", "test_event", map[string]string{"key": "value"})

	require.Contains(t, w1.Body.String(), "event: test_event")
	require.Contains(t, w1.Body.String(), `"key":"value"`)
	require.Contains(t, w2.Body.String(), "event: test_event")
}

// Flusher matches http.Flusher for test type assertions.
type Flusher = interface{ Flush() }

func TestConnectedUserIDs_AfterDisconnect(t *testing.T) {
	hub := newTestHub()

	conn1 := registerTestConn(hub, "group-1", "user-1")
	registerTestConn(hub, "group-1", "user-2")
	registerTestConn(hub, "group-1", "user-3")

	// user-1 disconnects
	hub.Unregister("group-1", "user-1", conn1)

	ids := hub.ConnectedUserIDs("group-1")
	sort.Strings(ids)
	assert.Equal(t, []string{"user-2", "user-3"}, ids)
}

// TestReconnectDoesNotLoseConnection simulates the page navigation scenario:
// user disconnects from game room, reconnects from play page.
func TestReconnectDoesNotLoseConnection(t *testing.T) {
	hub := newTestHub()

	// 1. User connects from game room
	roomConn := registerTestConn(hub, "group-1", "user-1")

	// 2. User navigates — new connection from play page
	_ = registerTestConn(hub, "group-1", "user-1")

	// 3. Old game room connection cleanup fires
	hub.Unregister("group-1", "user-1", roomConn)

	// User should STILL be connected (via play page connection)
	ids := hub.ConnectedUserIDs("group-1")
	assert.Equal(t, []string{"user-1"}, ids)
}

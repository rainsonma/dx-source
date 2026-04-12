package realtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/redis/go-redis/v9"
)

// TestIntegration_FullRoundtrip wires a complete hub end-to-end (miniredis
// + RedisPubSub + Hub + WebSocket handler) and verifies events flow from
// Publish through to a connected client.
func TestIntegration_FullRoundtrip(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ps := NewRedisPubSub(ctx, client)
	t.Cleanup(func() { _ = ps.Close() })
	hub := NewHub(ps, NewPresence(client), allowAllAuthorizer())

	ts := httptest.NewServer(httpHandlerForTest(hub, "alice"))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	// Subscribe to user:alice
	if err := wsjson.Write(ctx, conn, Envelope{Op: OpSubscribe, Topic: "user:alice", ID: "req_1"}); err != nil {
		t.Fatalf("write subscribe: %v", err)
	}

	// Read ack
	var ack Envelope
	if err := wsjson.Read(ctx, conn, &ack); err != nil {
		t.Fatalf("read ack: %v", err)
	}
	if ack.Op != OpAck || ack.ID != "req_1" || ack.OK == nil || !*ack.OK {
		t.Fatalf("bad ack: %+v", ack)
	}

	// Give the hub's pubsub.Subscribe a moment to register with miniredis
	time.Sleep(100 * time.Millisecond)

	// Publish from "server side"
	if err := ps.Publish(ctx, UserTopic("alice"), Event{Type: "pk_invitation", Data: map[string]string{"from": "bob"}}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	// Read event
	var ev Envelope
	if err := wsjson.Read(ctx, conn, &ev); err != nil {
		t.Fatalf("read event: %v", err)
	}
	if ev.Op != OpEvent {
		t.Errorf("expected event op, got %s", ev.Op)
	}
	if ev.Topic != "user:alice" {
		t.Errorf("wrong topic: %s", ev.Topic)
	}
	if ev.Type != "pk_invitation" {
		t.Errorf("wrong type: %s", ev.Type)
	}
}

// TestIntegration_CrossInstanceDelivery simulates two dx-api instances
// connected to the same Redis. A Publish on instance A reaches a client
// connected to instance B.
func TestIntegration_CrossInstanceDelivery(t *testing.T) {
	mr := miniredis.RunT(t)
	newClient := func() *redis.Client {
		return redis.NewClient(&redis.Options{Addr: mr.Addr()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Instance A: publishes
	psA := NewRedisPubSub(ctx, newClient())
	t.Cleanup(func() { _ = psA.Close() })
	hubA := NewHub(psA, NewPresence(newClient()), allowAllAuthorizer())
	_ = hubA

	// Instance B: has the client connected
	psB := NewRedisPubSub(ctx, newClient())
	t.Cleanup(func() { _ = psB.Close() })
	hubB := NewHub(psB, NewPresence(newClient()), allowAllAuthorizer())

	tsB := httptest.NewServer(httpHandlerForTest(hubB, "bob"))
	defer tsB.Close()

	wsURL := "ws" + strings.TrimPrefix(tsB.URL, "http") + "/ws"
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	// Bob subscribes to user:bob via hub B
	_ = wsjson.Write(ctx, conn, Envelope{Op: OpSubscribe, Topic: "user:bob", ID: "s1"})
	var ack Envelope
	_ = wsjson.Read(ctx, conn, &ack)

	time.Sleep(100 * time.Millisecond)

	// Publish via instance A's pubsub — should reach Bob on instance B
	_ = psA.Publish(ctx, UserTopic("bob"), Event{Type: "notif", Data: "hi"})

	var ev Envelope
	if err := wsjson.Read(ctx, conn, &ev); err != nil {
		t.Fatalf("read cross-instance event: %v", err)
	}
	if ev.Type != "notif" {
		t.Errorf("wrong type: %s", ev.Type)
	}
}

// httpHandlerForTest creates an http.Handler that upgrades and attaches the
// connection to the given hub with the given userID. Used only in tests.
func httpHandlerForTest(hub *Hub, userID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusInternalError, "test close")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_ = hub.Attach(ctx, userID, conn)
	}
}

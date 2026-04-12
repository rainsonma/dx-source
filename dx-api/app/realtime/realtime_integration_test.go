package realtime

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

func dialWS(t *testing.T, ts *httptest.Server) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return conn
}

func writeJSON(t *testing.T, conn *websocket.Conn, v any) {
	t.Helper()
	if err := conn.WriteJSON(v); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}
}

func readJSON(t *testing.T, conn *websocket.Conn, v any) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("readJSON: %v", err)
	}
	if err := json.Unmarshal(msg, v); err != nil {
		t.Fatalf("unmarshal: %v (raw: %s)", err, msg)
	}
}

func TestIntegration_FullRoundtrip(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	ctx := t.Context()

	ps := NewRedisPubSub(ctx, client)
	t.Cleanup(func() { _ = ps.Close() })
	hub := NewHub(ps, NewPresence(client), allowAllAuthorizer())

	ts := httptest.NewServer(httpHandlerForTest(hub, "alice"))
	defer ts.Close()

	conn := dialWS(t, ts)
	defer conn.Close()

	writeJSON(t, conn, Envelope{Op: OpSubscribe, Topic: "user:alice", ID: "req_1"})

	var ack Envelope
	readJSON(t, conn, &ack)
	if ack.Op != OpAck || ack.ID != "req_1" || ack.OK == nil || !*ack.OK {
		t.Fatalf("bad ack: %+v", ack)
	}

	time.Sleep(100 * time.Millisecond)

	if err := ps.Publish(ctx, UserTopic("alice"), Event{Type: "pk_invitation", Data: map[string]string{"from": "bob"}}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	var ev Envelope
	readJSON(t, conn, &ev)
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

func TestIntegration_CrossInstanceDelivery(t *testing.T) {
	mr := miniredis.RunT(t)
	newClient := func() *redis.Client {
		return redis.NewClient(&redis.Options{Addr: mr.Addr()})
	}

	ctx := t.Context()

	psA := NewRedisPubSub(ctx, newClient())
	t.Cleanup(func() { _ = psA.Close() })

	psB := NewRedisPubSub(ctx, newClient())
	t.Cleanup(func() { _ = psB.Close() })
	hubB := NewHub(psB, NewPresence(newClient()), allowAllAuthorizer())

	tsB := httptest.NewServer(httpHandlerForTest(hubB, "bob"))
	defer tsB.Close()

	conn := dialWS(t, tsB)
	defer conn.Close()

	writeJSON(t, conn, Envelope{Op: OpSubscribe, Topic: "user:bob", ID: "s1"})
	var ack Envelope
	readJSON(t, conn, &ack)

	time.Sleep(100 * time.Millisecond)

	_ = psA.Publish(ctx, UserTopic("bob"), Event{Type: "notif", Data: "hi"})

	var ev Envelope
	readJSON(t, conn, &ev)
	if ev.Type != "notif" {
		t.Errorf("wrong type: %s", ev.Type)
	}
}

// httpHandlerForTest creates an http.Handler that upgrades via gorilla and
// attaches to the hub.
func httpHandlerForTest(hub *Hub, userID string) http.HandlerFunc {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		ctx := r.Context()
		_ = hub.Attach(ctx, userID, conn)
	}
}

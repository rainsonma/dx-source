# WebSocket Refactor — PR 1: Backend Realtime Layer + Dormant Frontend Provider

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Spec:** `docs/superpowers/specs/2026-04-11-websocket-refactor-design.md`

**Goal:** Ship backend WebSocket infrastructure (the entire `app/realtime/` package, hub, pub/sub, controller, auth, presence tracker, tests) and dormant frontend provider files, with service-layer double-publish added to all ~31 call sites and Issue C's session-replaced kick wiring. Zero user-visible change on deploy — all existing SSE flows continue to work unchanged.

**Architecture:** New `app/realtime/` Go package implementing a multiplexed WebSocket hub backed by Redis pub/sub. New `/api/ws` endpoint using `coder/websocket` with a detached context to sidestep the Goravel Gin global 30s timeout. Dormant frontend provider files created but not mounted anywhere. Service code publishes to both legacy SSE hubs AND the new `realtime.Publish` during the migration window (removed in a later PR).

**Tech Stack:** Go 1.26, Goravel v1.17.2 (framework) + goravel/gin v1.17.0 + goravel/redis v1.17.1, go-redis/v9, coder/websocket (new dep), miniredis/v2 (new test dep), TypeScript 5, Next.js 16, React 19

**PR scope:** Backend + dormant frontend files + infrastructure fix + all 8 service files double-publish + Issue C + Issue E resolution. **Not in this PR:** mounting the provider, migrating any hook, deleting any SSE code, NDJSON cutover, stale doc deletion — those are PRs 2–6.

---

## Pre-flight checklist

Before starting any task, verify:
- [ ] Working tree is clean (`git status` shows nothing staged/modified)
- [ ] Current branch is NOT `main` — create a feature branch: `git checkout -b feat/websocket-refactor-pr1`
- [ ] `go version` shows 1.26 or newer
- [ ] Docker Compose is available: `docker compose version`
- [ ] You can read the spec: `docs/superpowers/specs/2026-04-11-websocket-refactor-design.md`

---

## Phase 0 — Dependencies & environment

### Task 0.1: Add Go dependencies

**Files:**
- Modify: `dx-api/go.mod`
- Modify: `dx-api/go.sum` (auto)

- [ ] **Step 1: Add the coder/websocket dependency**

Run:
```bash
cd dx-api && go get github.com/coder/websocket@latest
```

Expected: new entries in `go.mod` under `require`, and `go.sum` updated.

- [ ] **Step 2: Add miniredis as a test dependency**

Run:
```bash
cd dx-api && go get github.com/alicebob/miniredis/v2@latest
```

Expected: miniredis added to `go.mod`. It's a pure-Go Redis mock — no external process needed for tests.

- [ ] **Step 3: Verify builds still pass with the new deps**

Run:
```bash
cd dx-api && go build ./...
```

Expected: exit 0, no output.

- [ ] **Step 4: Commit**

```bash
cd dx-api && git add go.mod go.sum
git commit -m "chore: add coder/websocket and miniredis deps for WS refactor"
```

---

## Phase 1 — Types, topic helpers, error codes

### Task 1.1: Add 4 new error code constants

**Files:**
- Modify: `dx-api/app/consts/error_code.go`

- [ ] **Step 1: Open `dx-api/app/consts/error_code.go` and locate the existing 400xx and 500xx blocks**

- [ ] **Step 2: Append the 4 new constants in their appropriate family blocks**

In the 400xx block, after `CodeInvitationNotPending = 40021`, add:

```go
	CodeInvalidEnvelope = 40022  // WS envelope missing or malformed
	CodeUnknownOp       = 40023  // WS envelope op value not recognized
	CodeInvalidTopic    = 40024  // WS topic string doesn't match known patterns
```

In the 500xx block, after `CodeEmailSendError = 50002`, add:

```go
	CodeSlowConsumer = 50003  // WS client kicked due to send queue overflow
```

- [ ] **Step 3: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: exit 0.

- [ ] **Step 4: Commit**

```bash
cd dx-api && git add app/consts/error_code.go
git commit -m "feat(consts): add WS protocol error codes (40022-40024, 50003)"
```

---

### Task 1.2: Create envelope types and tests (TDD)

**Files:**
- Create: `dx-api/app/realtime/envelope.go`
- Create: `dx-api/app/realtime/envelope_test.go`

- [ ] **Step 1: Create the test file with failing tests**

Write `dx-api/app/realtime/envelope_test.go`:

```go
package realtime

import (
	"encoding/json"
	"testing"
)

func TestEnvelope_MarshalOmitsEmptyFields(t *testing.T) {
	env := Envelope{Op: OpSubscribe, Topic: "user:alice", ID: "req_1"}
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	got := string(b)
	want := `{"op":"subscribe","topic":"user:alice","id":"req_1"}`
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestEnvelope_MarshalEventWithData(t *testing.T) {
	env := Envelope{Op: OpEvent, Topic: "pk:abc", Type: "pk_player_action", Data: map[string]string{"user_id": "alice"}}
	b, _ := json.Marshal(env)
	var back map[string]any
	_ = json.Unmarshal(b, &back)
	if back["op"] != "event" {
		t.Errorf("wrong op: %v", back["op"])
	}
	if back["topic"] != "pk:abc" {
		t.Errorf("wrong topic: %v", back["topic"])
	}
	if back["type"] != "pk_player_action" {
		t.Errorf("wrong type: %v", back["type"])
	}
}

func TestEnvelope_AckWithOKTrue(t *testing.T) {
	ok := true
	env := Envelope{Op: OpAck, ID: "req_1", OK: &ok}
	b, _ := json.Marshal(env)
	got := string(b)
	want := `{"op":"ack","id":"req_1","ok":true}`
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestEnvelope_AckWithOKFalseAndCode(t *testing.T) {
	ok := false
	env := Envelope{Op: OpAck, ID: "req_1", OK: &ok, Code: 40300, Message: "forbidden"}
	b, _ := json.Marshal(env)
	got := string(b)
	want := `{"op":"ack","id":"req_1","ok":false,"code":40300,"message":"forbidden"}`
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestEnvelope_ErrorOp(t *testing.T) {
	env := Envelope{Op: OpError, Code: 40022, Message: "envelope missing op field"}
	b, _ := json.Marshal(env)
	got := string(b)
	want := `{"op":"error","code":40022,"message":"envelope missing op field"}`
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestEnvelope_UnmarshalRoundtrip(t *testing.T) {
	original := Envelope{Op: OpSubscribe, Topic: "group:xyz", ID: "req_42"}
	b, _ := json.Marshal(original)
	var back Envelope
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Op != original.Op || back.Topic != original.Topic || back.ID != original.ID {
		t.Errorf("roundtrip mismatch: %+v vs %+v", back, original)
	}
}

func TestEvent_CarriesTypeAndData(t *testing.T) {
	e := Event{Type: "pk_player_complete", Data: map[string]int{"score": 100}}
	if e.Type != "pk_player_complete" {
		t.Errorf("wrong type: %s", e.Type)
	}
	if m, ok := e.Data.(map[string]int); !ok || m["score"] != 100 {
		t.Errorf("wrong data: %+v", e.Data)
	}
}
```

- [ ] **Step 2: Run tests, confirm they fail**

```bash
cd dx-api && go test ./app/realtime/ -run TestEnvelope -v
```

Expected: compile error "undefined: Envelope" etc. This is the red state.

- [ ] **Step 3: Create the implementation file**

Write `dx-api/app/realtime/envelope.go`:

```go
package realtime

type Op string

const (
	OpSubscribe   Op = "subscribe"
	OpUnsubscribe Op = "unsubscribe"
	OpEvent       Op = "event"
	OpAck         Op = "ack"
	OpError       Op = "error"
)

type Envelope struct {
	Op      Op     `json:"op"`
	Topic   string `json:"topic,omitempty"`
	Type    string `json:"type,omitempty"`
	Data    any    `json:"data,omitempty"`
	ID      string `json:"id,omitempty"`
	OK      *bool  `json:"ok,omitempty"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Event struct {
	Type string
	Data any
}
```

- [ ] **Step 4: Run tests, confirm they pass**

```bash
cd dx-api && go test ./app/realtime/ -run TestEnvelope -v
cd dx-api && go test ./app/realtime/ -run TestEvent -v
```

Expected: all 7 tests PASS.

- [ ] **Step 5: Commit**

```bash
cd dx-api && git add app/realtime/envelope.go app/realtime/envelope_test.go
git commit -m "feat(realtime): add WS envelope types with JSON roundtrip tests"
```

---

### Task 1.3: Create topic helpers and tests (TDD)

**Files:**
- Create: `dx-api/app/realtime/topic.go`
- Create: `dx-api/app/realtime/topic_test.go`

- [ ] **Step 1: Create the test file**

Write `dx-api/app/realtime/topic_test.go`:

```go
package realtime

import "testing"

func TestUserTopic(t *testing.T) {
	if got := UserTopic("01HZA9"); got != "user:01HZA9" {
		t.Errorf("got %s", got)
	}
}

func TestUserKickTopic(t *testing.T) {
	if got := UserKickTopic("01HZA9"); got != "user:01HZA9:kick" {
		t.Errorf("got %s", got)
	}
}

func TestPkTopic(t *testing.T) {
	if got := PkTopic("pk_123"); got != "pk:pk_123" {
		t.Errorf("got %s", got)
	}
}

func TestGroupTopic(t *testing.T) {
	if got := GroupTopic("grp_xyz"); got != "group:grp_xyz" {
		t.Errorf("got %s", got)
	}
}

func TestGroupNotifyTopic(t *testing.T) {
	if got := GroupNotifyTopic("grp_xyz"); got != "group:grp_xyz:notify" {
		t.Errorf("got %s", got)
	}
}

func TestParseTopic(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKnd TopicKind
		wantID  string
		wantErr bool
	}{
		{"user valid", "user:alice", KindUser, "alice", false},
		{"user kick valid", "user:alice:kick", KindUserKick, "alice", false},
		{"pk valid", "pk:abc", KindPk, "abc", false},
		{"group valid", "group:xyz", KindGroup, "xyz", false},
		{"group notify valid", "group:xyz:notify", KindGroupNotify, "xyz", false},
		{"empty string", "", 0, "", true},
		{"single word", "user", 0, "", true},
		{"user with empty id", "user:", 0, "", true},
		{"unknown kind", "foo:bar", 0, "", true},
		{"pk with extra segment", "pk:abc:extra", 0, "", true},
		{"group with unknown suffix", "group:xyz:applications", 0, "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := ParseTopic(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("want error, got %+v", parsed)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected err: %v", err)
				return
			}
			if parsed.Kind != tc.wantKnd {
				t.Errorf("kind: got %d want %d", parsed.Kind, tc.wantKnd)
			}
			if parsed.ID != tc.wantID {
				t.Errorf("id: got %s want %s", parsed.ID, tc.wantID)
			}
		})
	}
}

func TestParseTopicRoundtrip(t *testing.T) {
	topics := []string{
		UserTopic("alice"),
		UserKickTopic("alice"),
		PkTopic("pk123"),
		GroupTopic("grp1"),
		GroupNotifyTopic("grp1"),
	}
	for _, topic := range topics {
		parsed, err := ParseTopic(topic)
		if err != nil {
			t.Errorf("parse %s: %v", topic, err)
			continue
		}
		if parsed.ID == "" {
			t.Errorf("empty id for %s", topic)
		}
	}
}
```

- [ ] **Step 2: Run tests, confirm they fail**

```bash
cd dx-api && go test ./app/realtime/ -run "Topic|Parse" -v
```

Expected: compile errors ("undefined: UserTopic", etc.).

- [ ] **Step 3: Create the implementation file**

Write `dx-api/app/realtime/topic.go`:

```go
package realtime

import (
	"errors"
	"strings"
)

type TopicKind int

const (
	KindUnknown TopicKind = iota
	KindUser
	KindUserKick
	KindPk
	KindGroup
	KindGroupNotify
)

type ParsedTopic struct {
	Kind TopicKind
	ID   string
}

var ErrBadTopic = errors.New("realtime: bad topic format")

func UserTopic(userID string) string     { return "user:" + userID }
func UserKickTopic(userID string) string { return "user:" + userID + ":kick" }
func PkTopic(pkID string) string         { return "pk:" + pkID }
func GroupTopic(groupID string) string   { return "group:" + groupID }
func GroupNotifyTopic(groupID string) string {
	return "group:" + groupID + ":notify"
}

// ParseTopic decomposes a topic string into its Kind and the entity ID.
// Rejects malformed or unknown topics.
func ParseTopic(topic string) (ParsedTopic, error) {
	parts := strings.Split(topic, ":")
	switch {
	case len(parts) == 2 && parts[0] == "user" && parts[1] != "":
		return ParsedTopic{Kind: KindUser, ID: parts[1]}, nil
	case len(parts) == 3 && parts[0] == "user" && parts[1] != "" && parts[2] == "kick":
		return ParsedTopic{Kind: KindUserKick, ID: parts[1]}, nil
	case len(parts) == 2 && parts[0] == "pk" && parts[1] != "":
		return ParsedTopic{Kind: KindPk, ID: parts[1]}, nil
	case len(parts) == 2 && parts[0] == "group" && parts[1] != "":
		return ParsedTopic{Kind: KindGroup, ID: parts[1]}, nil
	case len(parts) == 3 && parts[0] == "group" && parts[1] != "" && parts[2] == "notify":
		return ParsedTopic{Kind: KindGroupNotify, ID: parts[1]}, nil
	default:
		return ParsedTopic{}, ErrBadTopic
	}
}
```

- [ ] **Step 4: Run tests, confirm they pass**

```bash
cd dx-api && go test ./app/realtime/ -run "Topic|Parse" -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
cd dx-api && git add app/realtime/topic.go app/realtime/topic_test.go
git commit -m "feat(realtime): add topic naming and parsing helpers"
```

---

### Task 1.4: Create error sentinels

**Files:**
- Create: `dx-api/app/realtime/errors.go`

- [ ] **Step 1: Write the file directly (no tests — it's just variable declarations)**

Write `dx-api/app/realtime/errors.go`:

```go
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
```

- [ ] **Step 2: Verify build**

```bash
cd dx-api && go build ./app/realtime/
```

Expected: exit 0.

- [ ] **Step 3: Commit**

```bash
cd dx-api && git add app/realtime/errors.go
git commit -m "feat(realtime): add error sentinel and structured error type"
```

---

## Phase 2 — Redis primitives: presence + pub/sub

### Task 2.1: Create presence helpers with miniredis tests (TDD)

**Files:**
- Create: `dx-api/app/realtime/presence.go`
- Create: `dx-api/app/realtime/presence_test.go`

- [ ] **Step 1: Create the test file**

Write `dx-api/app/realtime/presence_test.go`:

```go
package realtime

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedisClient(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

func TestPresence_SetAddRemoveCheck(t *testing.T) {
	client, _ := newTestRedisClient(t)
	ctx := context.Background()
	p := NewPresence(client)

	if err := p.Add(ctx, "pk:abc", "alice"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := p.Add(ctx, "pk:abc", "bob"); err != nil {
		t.Fatalf("add: %v", err)
	}

	present, err := p.IsPresent(ctx, "pk:abc", "alice")
	if err != nil || !present {
		t.Errorf("alice should be present: err=%v present=%v", err, present)
	}

	present, err = p.IsPresent(ctx, "pk:abc", "carol")
	if err != nil || present {
		t.Errorf("carol should not be present: present=%v", present)
	}

	if err := p.Remove(ctx, "pk:abc", "alice"); err != nil {
		t.Fatalf("remove: %v", err)
	}

	present, err = p.IsPresent(ctx, "pk:abc", "alice")
	if err != nil || present {
		t.Errorf("alice should be removed: present=%v", present)
	}
}

func TestPresence_Members(t *testing.T) {
	client, _ := newTestRedisClient(t)
	ctx := context.Background()
	p := NewPresence(client)

	_ = p.Add(ctx, "group:xyz", "alice")
	_ = p.Add(ctx, "group:xyz", "bob")
	_ = p.Add(ctx, "group:xyz", "carol")

	members, err := p.Members(ctx, "group:xyz")
	if err != nil {
		t.Fatalf("members: %v", err)
	}
	if len(members) != 3 {
		t.Errorf("want 3 members, got %d: %v", len(members), members)
	}
	// Verify all expected users are present
	found := map[string]bool{}
	for _, m := range members {
		found[m] = true
	}
	for _, expected := range []string{"alice", "bob", "carol"} {
		if !found[expected] {
			t.Errorf("missing %s in members", expected)
		}
	}
}

func TestPresence_AddRefreshesExpiry(t *testing.T) {
	client, mr := newTestRedisClient(t)
	ctx := context.Background()
	p := NewPresence(client)

	_ = p.Add(ctx, "pk:abc", "alice")
	ttl := mr.TTL("presence:pk:abc")
	if ttl <= 0 {
		t.Errorf("expected positive TTL, got %v", ttl)
	}

	// Fast-forward time in miniredis
	mr.FastForward(5 * time.Minute)

	// Add again — should refresh TTL
	_ = p.Add(ctx, "pk:abc", "bob")
	ttl2 := mr.TTL("presence:pk:abc")
	if ttl2 <= 0 {
		t.Errorf("expected TTL refresh, got %v", ttl2)
	}
}

func TestPresence_EmptyKeyReturnsEmptyMembers(t *testing.T) {
	client, _ := newTestRedisClient(t)
	ctx := context.Background()
	p := NewPresence(client)

	members, err := p.Members(ctx, "group:nonexistent")
	if err != nil {
		t.Fatalf("members: %v", err)
	}
	if len(members) != 0 {
		t.Errorf("want empty, got %v", members)
	}
}
```

- [ ] **Step 2: Run tests, confirm compile failure**

```bash
cd dx-api && go test ./app/realtime/ -run TestPresence -v
```

Expected: compile errors ("undefined: NewPresence").

- [ ] **Step 3: Create the implementation file**

Write `dx-api/app/realtime/presence.go`:

```go
package realtime

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// presenceTTL is the sliding expiration window for presence sets.
// Every Add call refreshes the TTL. If a topic has no activity for
// presenceTTL, the set is garbage-collected automatically.
const presenceTTL = 15 * time.Minute

// Presence tracks which users are subscribed to which topics, across all
// dx-api instances sharing the same Redis. Backed by per-topic Redis SETs.
type Presence struct {
	client *redis.Client
}

// NewPresence constructs a Presence tracker backed by the given Redis client.
func NewPresence(client *redis.Client) *Presence {
	return &Presence{client: client}
}

func presenceKey(topic string) string { return "presence:" + topic }

// Add inserts userID into the presence set for topic and refreshes the TTL.
func (p *Presence) Add(ctx context.Context, topic, userID string) error {
	key := presenceKey(topic)
	pipe := p.client.Pipeline()
	pipe.SAdd(ctx, key, userID)
	pipe.Expire(ctx, key, presenceTTL)
	_, err := pipe.Exec(ctx)
	return err
}

// Remove deletes userID from the presence set for topic.
func (p *Presence) Remove(ctx context.Context, topic, userID string) error {
	return p.client.SRem(ctx, presenceKey(topic), userID).Err()
}

// IsPresent returns true if userID is currently in the topic's presence set.
func (p *Presence) IsPresent(ctx context.Context, topic, userID string) (bool, error) {
	return p.client.SIsMember(ctx, presenceKey(topic), userID).Result()
}

// Members returns all userIDs currently in the topic's presence set.
// Returns an empty slice (not nil) for unknown topics.
func (p *Presence) Members(ctx context.Context, topic string) ([]string, error) {
	result, err := p.client.SMembers(ctx, presenceKey(topic)).Result()
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []string{}, nil
	}
	return result, nil
}
```

- [ ] **Step 4: Run tests, confirm they pass**

```bash
cd dx-api && go test ./app/realtime/ -run TestPresence -v
```

Expected: all 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
cd dx-api && git add app/realtime/presence.go app/realtime/presence_test.go
git commit -m "feat(realtime): add Redis-backed topic presence tracker"
```

---

### Task 2.2: Define the PubSub interface and package-level Publish

**Files:**
- Create: `dx-api/app/realtime/pubsub.go`

- [ ] **Step 1: Write the interface file (no tests — this is interface + package-level state)**

Write `dx-api/app/realtime/pubsub.go`:

```go
package realtime

import "context"

// PubSub is the seam between the Hub and the backing transport. Production
// uses RedisPubSub; tests use a fake declared in the test file.
type PubSub interface {
	// Publish sends event on topic. At-most-once delivery semantics.
	Publish(ctx context.Context, topic string, event Event) error

	// Subscribe returns a receive channel of events delivered on topic.
	// The returned unsubscribe function must be called to tear down the
	// subscription and release resources.
	Subscribe(topic string) (<-chan Event, func())

	// Close terminates any background goroutines and releases resources.
	Close() error
}

// Default is the package-level PubSub used by Publish. Set at bootstrap.
var Default PubSub

// Publish is the single function every service call site uses to emit an
// event on a topic. It delegates to Default. Returns ErrNotInitialized if
// Default has not been wired (bootstrap-order bug).
func Publish(ctx context.Context, topic string, event Event) error {
	if Default == nil {
		return ErrNotInitialized
	}
	return Default.Publish(ctx, topic, event)
}
```

- [ ] **Step 2: Verify build**

```bash
cd dx-api && go build ./app/realtime/
```

Expected: exit 0.

- [ ] **Step 3: Commit**

```bash
cd dx-api && git add app/realtime/pubsub.go
git commit -m "feat(realtime): define PubSub interface and package-level Publish"
```

---

### Task 2.3: Create RedisPubSub implementation and tests (TDD with miniredis)

**Files:**
- Create: `dx-api/app/realtime/pubsub_redis.go`
- Create: `dx-api/app/realtime/pubsub_redis_test.go`

- [ ] **Step 1: Create the test file**

Write `dx-api/app/realtime/pubsub_redis_test.go`:

```go
package realtime

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newRedisForPubSub(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

func TestRedisPubSub_PublishReceive(t *testing.T) {
	client, _ := newRedisForPubSub(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ps := NewRedisPubSub(ctx, client)
	t.Cleanup(func() { _ = ps.Close() })

	ch, unsub := ps.Subscribe("test:topic")
	defer unsub()

	// Give the subscribe goroutine a moment to register with Redis.
	time.Sleep(50 * time.Millisecond)

	if err := ps.Publish(ctx, "test:topic", Event{Type: "hello", Data: "world"}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case ev := <-ch:
		if ev.Type != "hello" {
			t.Errorf("wrong type: %s", ev.Type)
		}
		// Data is deserialized as interface{}; for a string, it'll be a string.
		if s, ok := ev.Data.(string); !ok || s != "world" {
			t.Errorf("wrong data: %#v", ev.Data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestRedisPubSub_MultipleSubscribersOnSameTopic(t *testing.T) {
	client, _ := newRedisForPubSub(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ps := NewRedisPubSub(ctx, client)
	t.Cleanup(func() { _ = ps.Close() })

	ch1, unsub1 := ps.Subscribe("test:topic")
	defer unsub1()
	ch2, unsub2 := ps.Subscribe("test:topic")
	defer unsub2()

	time.Sleep(50 * time.Millisecond)

	_ = ps.Publish(ctx, "test:topic", Event{Type: "ping", Data: nil})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		select {
		case ev := <-ch1:
			if ev.Type != "ping" {
				t.Errorf("ch1 wrong type: %s", ev.Type)
			}
		case <-time.After(2 * time.Second):
			t.Error("ch1 timeout")
		}
	}()
	go func() {
		defer wg.Done()
		select {
		case ev := <-ch2:
			if ev.Type != "ping" {
				t.Errorf("ch2 wrong type: %s", ev.Type)
			}
		case <-time.After(2 * time.Second):
			t.Error("ch2 timeout")
		}
	}()
	wg.Wait()
}

func TestRedisPubSub_UnsubscribeStopsDelivery(t *testing.T) {
	client, _ := newRedisForPubSub(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ps := NewRedisPubSub(ctx, client)
	t.Cleanup(func() { _ = ps.Close() })

	ch, unsub := ps.Subscribe("test:topic")

	time.Sleep(50 * time.Millisecond)

	unsub()

	// Give the Redis UNSUBSCRIBE a moment
	time.Sleep(50 * time.Millisecond)

	_ = ps.Publish(ctx, "test:topic", Event{Type: "shouldnt-arrive"})

	select {
	case ev, ok := <-ch:
		if ok {
			t.Errorf("unexpected event after unsubscribe: %+v", ev)
		}
		// Channel closed is OK
	case <-time.After(300 * time.Millisecond):
		// No event arrived — good
	}
}

func TestRedisPubSub_PublishContextCanceled(t *testing.T) {
	client, _ := newRedisForPubSub(t)
	ps := NewRedisPubSub(context.Background(), client)
	t.Cleanup(func() { _ = ps.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()  // cancel immediately

	err := ps.Publish(ctx, "test:topic", Event{Type: "nope"})
	if err == nil {
		t.Error("expected error from canceled context")
	}
}

func TestRedisPubSub_CloseTerminatesLoop(t *testing.T) {
	client, _ := newRedisForPubSub(t)
	ctx := context.Background()
	ps := NewRedisPubSub(ctx, client)

	if err := ps.Close(); err != nil {
		t.Errorf("close: %v", err)
	}
	// Second close should be safe (no-op)
	_ = ps.Close()
}
```

- [ ] **Step 2: Run tests, confirm compile failure**

```bash
cd dx-api && go test ./app/realtime/ -run TestRedisPubSub -v
```

Expected: "undefined: NewRedisPubSub".

- [ ] **Step 3: Create the implementation file**

Write `dx-api/app/realtime/pubsub_redis.go`:

```go
package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

// RedisPubSub is the production implementation of PubSub, backed by Redis
// PUBLISH/SUBSCRIBE. One long-lived subscribe connection per instance
// handles all topics, with per-topic ref-counting to minimize Redis traffic.
type RedisPubSub struct {
	client *redis.Client

	mu     sync.Mutex
	locals map[string]map[chan Event]struct{} // topic -> set of local channels
	refs   map[string]int                     // topic -> subscribe ref count
	pubsub *redis.PubSub                      // Redis subscribe handle

	ctx    context.Context
	cancel context.CancelFunc
	closed bool
	done   chan struct{}
}

// NewRedisPubSub constructs and starts the subscribe loop goroutine.
// The passed context is used as the parent for the internal loop; cancel it
// (or call Close) to terminate.
func NewRedisPubSub(parent context.Context, client *redis.Client) *RedisPubSub {
	ctx, cancel := context.WithCancel(parent)
	ps := &RedisPubSub{
		client: client,
		locals: make(map[string]map[chan Event]struct{}),
		refs:   make(map[string]int),
		pubsub: client.Subscribe(ctx), // empty subscribe; channels added dynamically
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	go ps.loop()
	return ps
}

// Publish serializes event and publishes it to topic.
func (p *RedisPubSub) Publish(ctx context.Context, topic string, event Event) error {
	payload, err := json.Marshal(struct {
		Type string `json:"type"`
		Data any    `json:"data"`
	}{Type: event.Type, Data: event.Data})
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return p.client.Publish(ctx, topic, payload).Err()
}

// Subscribe registers a local channel for topic and returns it plus an
// unsubscribe function. First local subscriber triggers a Redis SUBSCRIBE.
func (p *RedisPubSub) Subscribe(topic string) (<-chan Event, func()) {
	ch := make(chan Event, 16)

	p.mu.Lock()
	if p.locals[topic] == nil {
		p.locals[topic] = make(map[chan Event]struct{})
	}
	p.locals[topic][ch] = struct{}{}
	firstLocal := p.refs[topic] == 0
	p.refs[topic]++
	p.mu.Unlock()

	if firstLocal {
		_ = p.pubsub.Subscribe(p.ctx, topic)
	}

	unsubscribe := func() {
		p.mu.Lock()
		if _, ok := p.locals[topic][ch]; !ok {
			p.mu.Unlock()
			return
		}
		delete(p.locals[topic], ch)
		p.refs[topic]--
		lastLocal := p.refs[topic] == 0
		if lastLocal {
			delete(p.locals, topic)
			delete(p.refs, topic)
		}
		// Close under the lock so it cannot race with a concurrent send
		// in loop(). The dispatch loop also holds p.mu, so close and send
		// are mutually exclusive.
		close(ch)
		p.mu.Unlock()

		if lastLocal {
			_ = p.pubsub.Unsubscribe(p.ctx, topic)
		}
	}

	return ch, unsubscribe
}

// Close terminates the subscribe loop and closes the Redis subscription.
// Idempotent.
func (p *RedisPubSub) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.mu.Unlock()

	p.cancel()
	_ = p.pubsub.Close()
	<-p.done
	return nil
}

// loop reads messages from the Redis subscribe connection and dispatches
// each to the registered local channels for that topic.
func (p *RedisPubSub) loop() {
	defer close(p.done)

	ch := p.pubsub.Channel()
	for msg := range ch {
		var wire struct {
			Type string `json:"type"`
			Data any    `json:"data"`
		}
		if err := json.Unmarshal([]byte(msg.Payload), &wire); err != nil {
			// Skip malformed messages
			continue
		}
		event := Event{Type: wire.Type, Data: wire.Data}

		// Dispatch under the lock so unsubscribe() cannot close a channel
		// while we're about to send to it. The send is non-blocking
		// (select + default drop-on-full), so lock contention per message
		// is bounded to microseconds.
		p.mu.Lock()
		for c := range p.locals[msg.Channel] {
			select {
			case c <- event:
			default:
				// Drop on full — the Hub handles slow-consumer logic by
				// tracking its own client send channels. This path is a
				// last-resort backstop.
			}
		}
		p.mu.Unlock()
	}
}
```

- [ ] **Step 4: Run tests, confirm they pass**

```bash
cd dx-api && go test ./app/realtime/ -run TestRedisPubSub -v -race
```

Expected: all 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
cd dx-api && git add app/realtime/pubsub_redis.go app/realtime/pubsub_redis_test.go
git commit -m "feat(realtime): Redis PubSub impl with ref-counted subscribe"
```

---

## Phase 3 — Authorization

### Task 3.1: Create authorize helper (TDD)

**Note:** Authorization requires database access for PK and group membership checks. The test uses a facade-mocked approach; the production queries run against real Postgres via `facades.Orm()`.

**Files:**
- Create: `dx-api/app/realtime/authorize.go`
- Create: `dx-api/app/realtime/authorize_test.go`

- [ ] **Step 1: Create the authorize.go file with dependency injection for testability**

Write `dx-api/app/realtime/authorize.go`:

```go
package realtime

import (
	"context"

	"dx-api/app/consts"

	"github.com/goravel/framework/facades"
)

// Authorizer decides whether a user may subscribe to a topic.
type Authorizer struct {
	// isPkParticipant returns true if userID is a participant in pkID.
	// Separated for testability; production wires it to a DB query.
	isPkParticipant func(ctx context.Context, userID, pkID string) (bool, error)

	// isGroupMember returns true if userID is a current member of groupID.
	isGroupMember func(ctx context.Context, userID, groupID string) (bool, error)
}

// NewAuthorizer returns an Authorizer wired to production DB queries.
func NewAuthorizer() *Authorizer {
	return &Authorizer{
		isPkParticipant: pkParticipantQuery,
		isGroupMember:   groupMemberQuery,
	}
}

// AuthorizeSubscribe checks whether userID may subscribe to topic.
// Returns nil on success or a realtimeError with consts.Code* and message
// on failure.
func (a *Authorizer) AuthorizeSubscribe(ctx context.Context, userID, topic string) error {
	parsed, err := ParseTopic(topic)
	if err != nil {
		return realtimeError{Code: consts.CodeInvalidTopic, Message: "invalid topic"}
	}
	switch parsed.Kind {
	case KindUser, KindUserKick:
		if parsed.ID != userID {
			return realtimeError{Code: consts.CodeForbidden, Message: "forbidden"}
		}
		return nil
	case KindPk:
		ok, err := a.isPkParticipant(ctx, userID, parsed.ID)
		if err != nil {
			// DB failure is a 5xx, not a 403 — don't confuse operators
			// and legitimate users during a Postgres outage.
			return realtimeError{Code: consts.CodeInternalError, Message: "pk membership check failed"}
		}
		if !ok {
			return realtimeError{Code: consts.CodeForbidden, Message: "not a participant in this PK"}
		}
		return nil
	case KindGroup, KindGroupNotify:
		ok, err := a.isGroupMember(ctx, userID, parsed.ID)
		if err != nil {
			return realtimeError{Code: consts.CodeInternalError, Message: "group membership check failed"}
		}
		if !ok {
			return realtimeError{Code: consts.CodeGroupForbidden, Message: "您不在该群组中"}
		}
		return nil
	default:
		return realtimeError{Code: consts.CodeInvalidTopic, Message: "unknown topic kind"}
	}
}

// pkParticipantQuery checks whether userID is the initiator (user_id) or
// opponent (opponent_id) of the given pkID in the game_pks table.
// Table/column names verified against app/models/game_pk.go.
func pkParticipantQuery(ctx context.Context, userID, pkID string) (bool, error) {
	count, err := facades.Orm().WithContext(ctx).Query().
		Table("game_pks").
		Where("id = ?", pkID).
		Where("(user_id = ? OR opponent_id = ?)", userID, userID).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// groupMemberQuery checks whether userID is a current member of groupID in
// the game_group_members table. Table/column names verified against
// app/models/game_group_member.go.
func groupMemberQuery(ctx context.Context, userID, groupID string) (bool, error) {
	count, err := facades.Orm().WithContext(ctx).Query().
		Table("game_group_members").
		Where("game_group_id = ?", groupID).
		Where("user_id = ?", userID).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
```

**Note about the production queries**: the exact table/column names above (`pk_matches.initiator_id`, `pk_matches.opponent_id`, `group_members.user_id`) are assumed from the service layer naming conventions. Before running integration tests that hit real Postgres, verify them by reading `app/models/pk_match.go` and `app/models/group_member.go`. If column names differ, update `pkParticipantQuery` and `groupMemberQuery` accordingly. The unit tests below don't hit the DB so this is safe to commit first.

- [ ] **Step 2: Create the test file using injected fake queries**

Write `dx-api/app/realtime/authorize_test.go`:

```go
package realtime

import (
	"context"
	"errors"
	"testing"

	"dx-api/app/consts"
)

func newTestAuthorizer(
	pkParticipants map[string]map[string]bool,
	groupMembers map[string]map[string]bool,
) *Authorizer {
	return &Authorizer{
		isPkParticipant: func(ctx context.Context, userID, pkID string) (bool, error) {
			return pkParticipants[pkID][userID], nil
		},
		isGroupMember: func(ctx context.Context, userID, groupID string) (bool, error) {
			return groupMembers[groupID][userID], nil
		},
	}
}

func TestAuthorize_UserTopicSelf(t *testing.T) {
	a := newTestAuthorizer(nil, nil)
	if err := a.AuthorizeSubscribe(context.Background(), "alice", "user:alice"); err != nil {
		t.Errorf("self user topic: %v", err)
	}
}

func TestAuthorize_UserTopicOther(t *testing.T) {
	a := newTestAuthorizer(nil, nil)
	err := a.AuthorizeSubscribe(context.Background(), "alice", "user:bob")
	if err == nil {
		t.Fatal("expected forbidden")
	}
	var rtErr realtimeError
	if !errors.As(err, &rtErr) || rtErr.Code != consts.CodeForbidden {
		t.Errorf("want CodeForbidden, got %+v", err)
	}
}

func TestAuthorize_UserKickTopicSelf(t *testing.T) {
	a := newTestAuthorizer(nil, nil)
	if err := a.AuthorizeSubscribe(context.Background(), "alice", "user:alice:kick"); err != nil {
		t.Errorf("self kick topic: %v", err)
	}
}

func TestAuthorize_PkParticipant(t *testing.T) {
	a := newTestAuthorizer(
		map[string]map[string]bool{"pk_abc": {"alice": true}},
		nil,
	)
	if err := a.AuthorizeSubscribe(context.Background(), "alice", "pk:pk_abc"); err != nil {
		t.Errorf("pk participant: %v", err)
	}
}

func TestAuthorize_PkNonParticipant(t *testing.T) {
	a := newTestAuthorizer(
		map[string]map[string]bool{"pk_abc": {"alice": true}},
		nil,
	)
	err := a.AuthorizeSubscribe(context.Background(), "bob", "pk:pk_abc")
	if err == nil {
		t.Fatal("expected forbidden")
	}
	var rtErr realtimeError
	if !errors.As(err, &rtErr) || rtErr.Code != consts.CodeForbidden {
		t.Errorf("want CodeForbidden, got %+v", err)
	}
}

func TestAuthorize_GroupMember(t *testing.T) {
	a := newTestAuthorizer(
		nil,
		map[string]map[string]bool{"grp_xyz": {"alice": true}},
	)
	if err := a.AuthorizeSubscribe(context.Background(), "alice", "group:grp_xyz"); err != nil {
		t.Errorf("group member: %v", err)
	}
}

func TestAuthorize_GroupNonMember(t *testing.T) {
	a := newTestAuthorizer(
		nil,
		map[string]map[string]bool{"grp_xyz": {"alice": true}},
	)
	err := a.AuthorizeSubscribe(context.Background(), "bob", "group:grp_xyz")
	if err == nil {
		t.Fatal("expected group forbidden")
	}
	var rtErr realtimeError
	if !errors.As(err, &rtErr) || rtErr.Code != consts.CodeGroupForbidden {
		t.Errorf("want CodeGroupForbidden, got %+v", err)
	}
}

func TestAuthorize_GroupNotifyMember(t *testing.T) {
	a := newTestAuthorizer(
		nil,
		map[string]map[string]bool{"grp_xyz": {"alice": true}},
	)
	if err := a.AuthorizeSubscribe(context.Background(), "alice", "group:grp_xyz:notify"); err != nil {
		t.Errorf("group notify member: %v", err)
	}
}

func TestAuthorize_GroupNotifyNonMember(t *testing.T) {
	a := newTestAuthorizer(
		nil,
		map[string]map[string]bool{"grp_xyz": {"alice": true}},
	)
	err := a.AuthorizeSubscribe(context.Background(), "bob", "group:grp_xyz:notify")
	if err == nil {
		t.Fatal("expected group forbidden")
	}
	var rtErr realtimeError
	if !errors.As(err, &rtErr) || rtErr.Code != consts.CodeGroupForbidden {
		t.Errorf("want CodeGroupForbidden, got %+v", err)
	}
}

func TestAuthorize_InvalidTopic(t *testing.T) {
	a := newTestAuthorizer(nil, nil)
	err := a.AuthorizeSubscribe(context.Background(), "alice", "garbage")
	if err == nil {
		t.Fatal("expected invalid topic")
	}
	var rtErr realtimeError
	if !errors.As(err, &rtErr) || rtErr.Code != consts.CodeInvalidTopic {
		t.Errorf("want CodeInvalidTopic, got %+v", err)
	}
}
```

- [ ] **Step 3: Run tests, confirm they pass**

```bash
cd dx-api && go test ./app/realtime/ -run TestAuthorize -v
```

Expected: all 10 tests PASS.

- [ ] **Step 4: Verify the production file compiles**

```bash
cd dx-api && go build ./app/realtime/
```

Expected: exit 0.

- [ ] **Step 5: Check the actual table names match production**

Read these files to confirm column names used by `pkParticipantQuery` and `groupMemberQuery`:

```bash
# Verify pk_matches columns
cd dx-api && grep -n "initiator_id\|opponent_id" app/models/pk_match.go

# Verify group_members columns
cd dx-api && grep -n "group_id\|user_id" app/models/group_member.go
```

If column names in the models differ from what the queries use, update `authorize.go` accordingly before committing. If column names match, proceed.

- [ ] **Step 6: Commit**

```bash
cd dx-api && git add app/realtime/authorize.go app/realtime/authorize_test.go
git commit -m "feat(realtime): add topic subscribe authorization with unit tests"
```

---

## Phase 4 — Hub and Client

### Task 4.1: Create Hub core with attach/detach + fake PubSub for testing

**Files:**
- Create: `dx-api/app/realtime/hub.go`
- Create: `dx-api/app/realtime/hub_test.go`

- [ ] **Step 1: Create the hub.go file with core struct and Attach/detach plumbing**

Write `dx-api/app/realtime/hub.go`:

```go
package realtime

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"dx-api/app/consts"

	"github.com/coder/websocket"
	"github.com/goravel/framework/facades"
)

// Hub is the local client registry and topic router for one dx-api process.
// It knows which Clients are connected, which topics each is subscribed to,
// and how to route incoming PubSub events to the matching local clients.
type Hub struct {
	pubsub     PubSub
	presence   *Presence
	authorizer *Authorizer

	mu         sync.RWMutex
	clients    map[*Client]struct{}            // all attached clients
	topics     map[string]map[*Client]struct{} // topic -> subscribed clients
	unsubs     map[string]func()               // topic -> pubsub unsubscribe fn
	shutdownFg atomic.Bool
}

// Default_Hub is set at bootstrap. WSController.Handle calls Attach on this.
var Default_Hub *Hub

// NewHub constructs a Hub wired to the given PubSub, Presence tracker, and
// Authorizer. All three are required.
func NewHub(ps PubSub, presence *Presence, authorizer *Authorizer) *Hub {
	return &Hub{
		pubsub:     ps,
		presence:   presence,
		authorizer: authorizer,
		clients:    make(map[*Client]struct{}),
		topics:     make(map[string]map[*Client]struct{}),
		unsubs:     make(map[string]func()),
	}
}

// IsShuttingDown returns true after Shutdown has been called. The WS
// controller checks this to reject new upgrades during graceful shutdown.
func (h *Hub) IsShuttingDown() bool { return h.shutdownFg.Load() }

// Attach takes ownership of the WebSocket connection for the given user,
// starts the client's read/write loops, and blocks until the connection
// terminates. Returns any error from the read loop (including normal EOF).
//
// IMPORTANT: ctx MUST be a detached context (context.Background() or
// similar), NOT the HTTP request context. Goravel's global Timeout
// middleware cancels the request context after http.request_timeout
// (30s default), which would kill long-lived WS connections.
func (h *Hub) Attach(ctx context.Context, userID string, conn *websocket.Conn) error {
	client := newClient(h, userID, conn)

	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			facades.Log().Errorf("client panic: user=%s err=%v\n%s", userID, r, debug.Stack())
		}
		h.detach(client)
	}()

	return client.Serve(ctx)
}

// detach removes a client from the hub, unsubscribing it from all topics.
// Called from Attach's defer and from subscribe slow-consumer handling.
func (h *Hub) detach(c *Client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; !ok {
		h.mu.Unlock()
		return
	}
	delete(h.clients, c)

	// Collect topics to unsubscribe outside the lock
	topicsToRemove := make([]string, 0, len(c.topics))
	for topic := range c.topics {
		topicsToRemove = append(topicsToRemove, topic)
	}
	h.mu.Unlock()

	for _, topic := range topicsToRemove {
		h.unsubscribe(c, topic)
	}
}

// Shutdown closes all attached clients gracefully and terminates the hub.
// Called from the Goravel terminating hook on SIGTERM.
func (h *Hub) Shutdown(ctx context.Context) error {
	h.shutdownFg.Store(true)

	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	// Send server_shutdown close frame to each client
	for _, c := range clients {
		go func(client *Client) {
			_ = client.conn.Close(4002, "server_shutdown")
		}(c)
	}

	// Wait up to 5 seconds for clean drain
	deadline := time.After(5 * time.Second)
	for {
		h.mu.RLock()
		n := len(h.clients)
		h.mu.RUnlock()
		if n == 0 {
			break
		}
		select {
		case <-deadline:
			// Force-close stragglers
			h.mu.RLock()
			stragglers := make([]*Client, 0, len(h.clients))
			for c := range h.clients {
				stragglers = append(stragglers, c)
			}
			h.mu.RUnlock()
			for _, c := range stragglers {
				_ = c.conn.CloseNow()
			}
			goto done
		case <-ctx.Done():
			goto done
		case <-time.After(50 * time.Millisecond):
		}
	}
done:

	// Close the pubsub loop
	_ = h.pubsub.Close()
	return nil
}

// subscribe authorizes and adds a client to a topic. First local subscriber
// triggers a PubSub subscribe. For group topics, publishes room_member_joined.
func (h *Hub) subscribe(ctx context.Context, c *Client, topic string) error {
	if err := h.authorizer.AuthorizeSubscribe(ctx, c.userID, topic); err != nil {
		return err
	}

	h.mu.Lock()
	firstLocal := h.topics[topic] == nil
	if firstLocal {
		h.topics[topic] = make(map[*Client]struct{})
	}
	// Idempotent: if already subscribed, no-op (still ack success)
	if _, alreadySubbed := h.topics[topic][c]; alreadySubbed {
		h.mu.Unlock()
		return nil
	}
	h.topics[topic][c] = struct{}{}
	c.addTopic(topic)
	h.mu.Unlock()

	if firstLocal {
		ch, unsub := h.pubsub.Subscribe(topic)
		h.mu.Lock()
		h.unsubs[topic] = unsub
		h.mu.Unlock()
		go h.fanout(topic, ch)
	}

	// Record presence (best-effort — failures don't block)
	if h.presence != nil {
		_ = h.presence.Add(ctx, topic, c.userID)
	}

	// Auto-publish room_member_joined for group topics
	if parsed, err := ParseTopic(topic); err == nil && parsed.Kind == KindGroup {
		_ = h.pubsub.Publish(ctx, topic, Event{
			Type: "room_member_joined",
			Data: map[string]string{"user_id": c.userID},
		})
	}

	return nil
}

// unsubscribe removes a client from a topic. Last local unsubscriber
// triggers a PubSub unsubscribe. For group topics, publishes room_member_left.
func (h *Hub) unsubscribe(c *Client, topic string) {
	h.mu.Lock()
	if h.topics[topic] == nil {
		h.mu.Unlock()
		return
	}
	if _, ok := h.topics[topic][c]; !ok {
		h.mu.Unlock()
		return
	}
	delete(h.topics[topic], c)
	c.removeTopic(topic)

	var unsub func()
	lastLocal := len(h.topics[topic]) == 0
	if lastLocal {
		unsub = h.unsubs[topic]
		delete(h.topics, topic)
		delete(h.unsubs, topic)
	}
	h.mu.Unlock()

	ctx := context.Background()

	if h.presence != nil {
		_ = h.presence.Remove(ctx, topic, c.userID)
	}

	// Auto-publish room_member_left for group topics
	if parsed, err := ParseTopic(topic); err == nil && parsed.Kind == KindGroup {
		_ = h.pubsub.Publish(ctx, topic, Event{
			Type: "room_member_left",
			Data: map[string]string{"user_id": c.userID},
		})
	}

	if unsub != nil {
		unsub()
	}
}

// fanout delivers events from a PubSub subscribe channel to all currently
// subscribed clients on that topic. Runs as a goroutine per subscribed topic.
func (h *Hub) fanout(topic string, ch <-chan Event) {
	defer func() {
		if r := recover(); r != nil {
			facades.Log().Errorf("fanout panic: topic=%s err=%v\n%s", topic, r, debug.Stack())
		}
	}()

	for ev := range ch {
		h.mu.RLock()
		subs := make([]*Client, 0, len(h.topics[topic]))
		for c := range h.topics[topic] {
			subs = append(subs, c)
		}
		h.mu.RUnlock()

		env := Envelope{
			Op:    OpEvent,
			Topic: topic,
			Type:  ev.Type,
			Data:  ev.Data,
		}
		for _, c := range subs {
			c.enqueue(env)
		}
	}
}

// kickSlowConsumer is called when a client's send channel is full. It
// detaches the client and closes the connection with code 4003.
func (h *Hub) kickSlowConsumer(c *Client) {
	go func() {
		_ = c.conn.Close(4003, "slow_consumer")
	}()
}

// ackError builds an ack envelope from a realtimeError or generic error.
func ackError(id string, err error) Envelope {
	ok := false
	env := Envelope{Op: OpAck, ID: id, OK: &ok}
	var rtErr realtimeError
	if errors.As(err, &rtErr) {
		env.Code = rtErr.Code
		env.Message = rtErr.Message
	} else {
		env.Code = consts.CodeInternalError
		env.Message = fmt.Sprintf("internal: %v", err)
	}
	return env
}
```

- [ ] **Step 2: Create the hub_test.go file with a fake PubSub and initial tests**

Write `dx-api/app/realtime/hub_test.go`:

```go
package realtime

import (
	"context"
	"sync"
	"testing"
	"time"
)

// fakePubSub is an in-memory PubSub used to test Hub logic in isolation
// without bringing up Redis.
type fakePubSub struct {
	mu       sync.Mutex
	subs     map[string][]chan Event
	closed   bool
	publishes []struct {
		Topic string
		Event Event
	}
}

func newFakePubSub() *fakePubSub {
	return &fakePubSub{subs: make(map[string][]chan Event)}
}

func (p *fakePubSub) Publish(ctx context.Context, topic string, event Event) error {
	p.mu.Lock()
	p.publishes = append(p.publishes, struct {
		Topic string
		Event Event
	}{Topic: topic, Event: event})
	subs := append([]chan Event{}, p.subs[topic]...)
	p.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
	return nil
}

func (p *fakePubSub) Subscribe(topic string) (<-chan Event, func()) {
	ch := make(chan Event, 16)
	p.mu.Lock()
	p.subs[topic] = append(p.subs[topic], ch)
	p.mu.Unlock()

	unsub := func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		for i, c := range p.subs[topic] {
			if c == ch {
				p.subs[topic] = append(p.subs[topic][:i], p.subs[topic][i+1:]...)
				break
			}
		}
		close(ch)
	}
	return ch, unsub
}

func (p *fakePubSub) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	return nil
}

// allowAllAuthorizer is an Authorizer that approves every subscribe.
func allowAllAuthorizer() *Authorizer {
	return &Authorizer{
		isPkParticipant: func(ctx context.Context, userID, pkID string) (bool, error) { return true, nil },
		isGroupMember:   func(ctx context.Context, userID, groupID string) (bool, error) { return true, nil },
	}
}

func TestHub_NewHubConstructs(t *testing.T) {
	ps := newFakePubSub()
	hub := NewHub(ps, nil, allowAllAuthorizer())
	if hub == nil {
		t.Fatal("nil hub")
	}
	if len(hub.clients) != 0 {
		t.Errorf("expected empty clients, got %d", len(hub.clients))
	}
}

func TestHub_SubscribeAuthorizedAddsClient(t *testing.T) {
	ps := newFakePubSub()
	hub := NewHub(ps, nil, allowAllAuthorizer())

	// Create a fake client manually (bypassing the WebSocket)
	c := &Client{
		hub:    hub,
		userID: "alice",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	hub.mu.Lock()
	hub.clients[c] = struct{}{}
	hub.mu.Unlock()

	err := hub.subscribe(context.Background(), c, "user:alice")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	hub.mu.RLock()
	_, inTopic := hub.topics["user:alice"][c]
	hub.mu.RUnlock()
	if !inTopic {
		t.Error("client not added to topic")
	}
	if _, has := c.topics["user:alice"]; !has {
		t.Error("client.topics missing topic")
	}
}

func TestHub_SubscribeUnauthorized(t *testing.T) {
	ps := newFakePubSub()
	authorizer := &Authorizer{
		isPkParticipant: func(ctx context.Context, userID, pkID string) (bool, error) { return false, nil },
		isGroupMember:   func(ctx context.Context, userID, groupID string) (bool, error) { return false, nil },
	}
	hub := NewHub(ps, nil, authorizer)

	c := &Client{
		hub:    hub,
		userID: "bob",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	hub.mu.Lock()
	hub.clients[c] = struct{}{}
	hub.mu.Unlock()

	err := hub.subscribe(context.Background(), c, "pk:abc")
	if err == nil {
		t.Fatal("expected authorization error")
	}

	hub.mu.RLock()
	_, inTopic := hub.topics["pk:abc"][c]
	hub.mu.RUnlock()
	if inTopic {
		t.Error("unauthorized client should not be in topic")
	}
}
```

- [ ] **Step 3: Run the tests — they will fail to compile because Client doesn't exist yet**

```bash
cd dx-api && go test ./app/realtime/ -run TestHub -v
```

Expected: "undefined: Client" or similar. This is fine — we'll create Client in Task 4.2 and then re-run.

**Do NOT commit this task's Hub file yet. Continue to Task 4.2 which defines Client; then commit both together.**

---

### Task 4.2: Create Client with read/write loops

**Files:**
- Create: `dx-api/app/realtime/client.go`

- [ ] **Step 1: Write client.go**

Write `dx-api/app/realtime/client.go`:

```go
package realtime

import (
	"context"
	"sync"
	"time"

	"dx-api/app/consts"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	sendQueueCapacity = 32
	pingInterval      = 25 * time.Second
	pingTimeout       = 5 * time.Second
	readLimitBytes    = 4096
)

// Client owns one WebSocket connection and its read/write loops.
type Client struct {
	hub    *Hub
	userID string
	conn   *websocket.Conn

	send   chan Envelope       // buffered; drained by writeLoop
	topics map[string]struct{} // topics this client subscribes to

	mu     sync.Mutex
	closed bool
}

func newClient(h *Hub, userID string, conn *websocket.Conn) *Client {
	return &Client{
		hub:    h,
		userID: userID,
		conn:   conn,
		send:   make(chan Envelope, sendQueueCapacity),
		topics: make(map[string]struct{}),
	}
}

// Serve runs the client's read and write loops. Blocks until the read loop
// returns (on disconnect, protocol error, or explicit close).
func (c *Client) Serve(ctx context.Context) error {
	c.conn.SetReadLimit(readLimitBytes)

	writeCtx, cancelWrite := context.WithCancel(ctx)
	defer cancelWrite()

	// Auto-subscribe the client to its own user:{id}:kick topic for Issue C
	_ = c.hub.subscribe(ctx, c, UserKickTopic(c.userID))

	writeDone := make(chan struct{})
	go func() {
		defer close(writeDone)
		c.writeLoop(writeCtx)
	}()

	err := c.readLoop(ctx)

	// readLoop returned (disconnect or error) — signal writeLoop to exit
	cancelWrite()
	<-writeDone
	return err
}

// readLoop reads inbound envelopes from the WebSocket and dispatches them
// to the hub's subscribe/unsubscribe handlers.
func (c *Client) readLoop(ctx context.Context) error {
	for {
		var env Envelope
		if err := wsjson.Read(ctx, c.conn, &env); err != nil {
			return err
		}
		switch env.Op {
		case OpSubscribe:
			if err := c.hub.subscribe(ctx, c, env.Topic); err != nil {
				c.enqueue(ackError(env.ID, err))
				continue
			}
			ok := true
			c.enqueue(Envelope{Op: OpAck, ID: env.ID, OK: &ok})
		case OpUnsubscribe:
			c.hub.unsubscribe(c, env.Topic)
			ok := true
			c.enqueue(Envelope{Op: OpAck, ID: env.ID, OK: &ok})
		default:
			c.enqueue(Envelope{
				Op:      OpError,
				Code:    consts.CodeUnknownOp,
				Message: "unknown op: " + string(env.Op),
			})
		}
	}
}

// writeLoop drains the send channel and writes frames. Also sends periodic
// server-initiated pings.
func (c *Client) writeLoop(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case env, ok := <-c.send:
			if !ok {
				return
			}
			if err := wsjson.Write(ctx, c.conn, env); err != nil {
				return
			}
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// enqueue tries to add an envelope to the send channel without blocking.
// If the channel is full, the client is considered too slow and kicked.
func (c *Client) enqueue(env Envelope) {
	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if closed {
		return
	}

	select {
	case c.send <- env:
	default:
		c.mu.Lock()
		if !c.closed {
			c.closed = true
			c.mu.Unlock()
			c.hub.kickSlowConsumer(c)
			return
		}
		c.mu.Unlock()
	}
}

// addTopic records a topic subscription on the client. Must be called under
// the hub's lock.
func (c *Client) addTopic(topic string) {
	c.topics[topic] = struct{}{}
}

// removeTopic removes a topic subscription from the client. Must be called
// under the hub's lock.
func (c *Client) removeTopic(topic string) {
	delete(c.topics, topic)
}
```

- [ ] **Step 2: Run the existing hub tests (they should now compile)**

```bash
cd dx-api && go test ./app/realtime/ -run TestHub -v
```

Expected: 3 tests PASS.

- [ ] **Step 3: Commit Hub and Client together**

```bash
cd dx-api && git add app/realtime/hub.go app/realtime/hub_test.go app/realtime/client.go
git commit -m "feat(realtime): Hub and Client with subscribe/unsubscribe plumbing"
```

---

### Task 4.3: Add Hub tests for subscribe ref-counting and topic cleanup

**Files:**
- Modify: `dx-api/app/realtime/hub_test.go`

- [ ] **Step 1: Append new test cases to hub_test.go**

Add to `dx-api/app/realtime/hub_test.go` (after the existing tests):

```go
func TestHub_FirstSubscribeTriggersFakePubSubSubscribe(t *testing.T) {
	ps := newFakePubSub()
	hub := NewHub(ps, nil, allowAllAuthorizer())

	c1 := &Client{
		hub:    hub,
		userID: "alice",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	c2 := &Client{
		hub:    hub,
		userID: "bob",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	hub.mu.Lock()
	hub.clients[c1] = struct{}{}
	hub.clients[c2] = struct{}{}
	hub.mu.Unlock()

	_ = hub.subscribe(context.Background(), c1, "user:alice")
	_ = hub.subscribe(context.Background(), c2, "user:bob")

	ps.mu.Lock()
	subCount := len(ps.subs)
	ps.mu.Unlock()

	if subCount != 2 {
		t.Errorf("want 2 pubsub subscribes (one per user topic), got %d", subCount)
	}
}

func TestHub_SecondLocalSubscribeReusesPubSub(t *testing.T) {
	// Skip if we can't easily test this — needs custom authorizer to allow
	// multiple users to subscribe to the same topic for isolation.
	// Use a group topic since multiple members can subscribe.
	ps := newFakePubSub()
	hub := NewHub(ps, nil, allowAllAuthorizer())

	c1 := &Client{
		hub:    hub,
		userID: "alice",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	c2 := &Client{
		hub:    hub,
		userID: "bob",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	hub.mu.Lock()
	hub.clients[c1] = struct{}{}
	hub.clients[c2] = struct{}{}
	hub.mu.Unlock()

	_ = hub.subscribe(context.Background(), c1, "group:grp1")
	_ = hub.subscribe(context.Background(), c2, "group:grp1")

	ps.mu.Lock()
	grpSubs := len(ps.subs["group:grp1"])
	ps.mu.Unlock()

	if grpSubs != 1 {
		t.Errorf("want 1 pubsub channel for group:grp1 (shared), got %d", grpSubs)
	}
}

func TestHub_LastUnsubscribeClearsTopic(t *testing.T) {
	ps := newFakePubSub()
	hub := NewHub(ps, nil, allowAllAuthorizer())

	c := &Client{
		hub:    hub,
		userID: "alice",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	hub.mu.Lock()
	hub.clients[c] = struct{}{}
	hub.mu.Unlock()

	_ = hub.subscribe(context.Background(), c, "user:alice")
	hub.unsubscribe(c, "user:alice")

	hub.mu.RLock()
	_, topicExists := hub.topics["user:alice"]
	hub.mu.RUnlock()
	if topicExists {
		t.Error("topic should be removed after last unsubscribe")
	}
}

func TestHub_IdempotentSubscribe(t *testing.T) {
	ps := newFakePubSub()
	hub := NewHub(ps, nil, allowAllAuthorizer())

	c := &Client{
		hub:    hub,
		userID: "alice",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	hub.mu.Lock()
	hub.clients[c] = struct{}{}
	hub.mu.Unlock()

	_ = hub.subscribe(context.Background(), c, "user:alice")
	err := hub.subscribe(context.Background(), c, "user:alice") // second time
	if err != nil {
		t.Errorf("idempotent subscribe should succeed: %v", err)
	}

	hub.mu.RLock()
	n := len(hub.topics["user:alice"])
	hub.mu.RUnlock()
	if n != 1 {
		t.Errorf("want 1 client in topic, got %d", n)
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd dx-api && go test ./app/realtime/ -run TestHub -v -race
```

Expected: all 7 Hub tests PASS (3 original + 4 new).

- [ ] **Step 3: Commit**

```bash
cd dx-api && git add app/realtime/hub_test.go
git commit -m "test(realtime): add Hub subscribe ref-counting tests"
```

---

### Task 4.4: Add Hub test for group auto-events (room_member_joined/left)

**Files:**
- Modify: `dx-api/app/realtime/hub_test.go`

- [ ] **Step 1: Append the auto-event tests**

Add to `dx-api/app/realtime/hub_test.go`:

```go
func TestHub_SubscribeToGroupPublishesRoomMemberJoined(t *testing.T) {
	ps := newFakePubSub()
	hub := NewHub(ps, nil, allowAllAuthorizer())

	c := &Client{
		hub:    hub,
		userID: "alice",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	hub.mu.Lock()
	hub.clients[c] = struct{}{}
	hub.mu.Unlock()

	_ = hub.subscribe(context.Background(), c, "group:grp1")

	ps.mu.Lock()
	defer ps.mu.Unlock()
	found := false
	for _, pub := range ps.publishes {
		if pub.Topic == "group:grp1" && pub.Event.Type == "room_member_joined" {
			data, _ := pub.Event.Data.(map[string]string)
			if data["user_id"] == "alice" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("did not find room_member_joined for alice")
	}
}

func TestHub_UnsubscribeFromGroupPublishesRoomMemberLeft(t *testing.T) {
	ps := newFakePubSub()
	hub := NewHub(ps, nil, allowAllAuthorizer())

	c := &Client{
		hub:    hub,
		userID: "alice",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	hub.mu.Lock()
	hub.clients[c] = struct{}{}
	hub.mu.Unlock()

	_ = hub.subscribe(context.Background(), c, "group:grp1")
	hub.unsubscribe(c, "group:grp1")

	ps.mu.Lock()
	defer ps.mu.Unlock()
	found := false
	for _, pub := range ps.publishes {
		if pub.Topic == "group:grp1" && pub.Event.Type == "room_member_left" {
			data, _ := pub.Event.Data.(map[string]string)
			if data["user_id"] == "alice" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("did not find room_member_left for alice")
	}
}

func TestHub_SubscribeToUserTopicDoesNotPublishRoomEvent(t *testing.T) {
	ps := newFakePubSub()
	hub := NewHub(ps, nil, allowAllAuthorizer())

	c := &Client{
		hub:    hub,
		userID: "alice",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 16),
	}
	hub.mu.Lock()
	hub.clients[c] = struct{}{}
	hub.mu.Unlock()

	_ = hub.subscribe(context.Background(), c, "user:alice")

	ps.mu.Lock()
	defer ps.mu.Unlock()
	for _, pub := range ps.publishes {
		if pub.Event.Type == "room_member_joined" || pub.Event.Type == "room_member_left" {
			t.Errorf("unexpected room event on user topic: %+v", pub)
		}
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd dx-api && go test ./app/realtime/ -run TestHub -v -race
```

Expected: all Hub tests PASS (including new auto-event tests).

- [ ] **Step 3: Commit**

```bash
cd dx-api && git add app/realtime/hub_test.go
git commit -m "test(realtime): verify group topic auto-emits room_member_joined/left"
```

---

### Task 4.5: Add Hub slow-consumer test

**Files:**
- Modify: `dx-api/app/realtime/hub_test.go`

- [ ] **Step 1: Append the slow-consumer test**

Add to `dx-api/app/realtime/hub_test.go`:

```go
func TestClient_EnqueueFullChannelTriggersKick(t *testing.T) {
	// Create a fake client with a send queue of capacity 1. Fill it, then
	// enqueue another — the second should trigger slow-consumer kick.
	hub := NewHub(newFakePubSub(), nil, allowAllAuthorizer())
	c := &Client{
		hub:    hub,
		userID: "slow",
		topics: make(map[string]struct{}),
		send:   make(chan Envelope, 1),
	}
	// Fill the channel
	c.send <- Envelope{Op: OpEvent}

	// The kick goroutine will call conn.Close, which panics on a nil conn.
	// Guard by setting closed=true after kicking to avoid the panic path.
	// For this unit test, verify that enqueue marks closed and attempts kick
	// by checking the closed flag after the call.
	done := make(chan struct{})
	go func() {
		defer func() { recover(); close(done) }()
		c.enqueue(Envelope{Op: OpEvent, Type: "second"})
	}()
	<-done

	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if !closed {
		t.Error("client should be marked closed after enqueue on full channel")
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd dx-api && go test ./app/realtime/ -run TestClient_Enqueue -v
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
cd dx-api && git add app/realtime/hub_test.go
git commit -m "test(realtime): verify slow consumer triggers kick"
```

---

### Task 4.6: Add Hub Shutdown test

**Files:**
- Modify: `dx-api/app/realtime/hub_test.go`

- [ ] **Step 1: Append shutdown test**

Add to `dx-api/app/realtime/hub_test.go`:

```go
func TestHub_ShutdownSetsFlag(t *testing.T) {
	hub := NewHub(newFakePubSub(), nil, allowAllAuthorizer())

	if hub.IsShuttingDown() {
		t.Error("shouldnt be shutting down before Shutdown called")
	}

	// Shutdown on empty hub completes quickly
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := hub.Shutdown(ctx); err != nil {
		t.Errorf("shutdown: %v", err)
	}

	if !hub.IsShuttingDown() {
		t.Error("should be shutting down after Shutdown")
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd dx-api && go test ./app/realtime/ -run TestHub_Shutdown -v
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
cd dx-api && git add app/realtime/hub_test.go
git commit -m "test(realtime): verify Hub.Shutdown sets IsShuttingDown flag"
```

---

## Phase 5 — Infrastructure consolidation, WSController, bootstrap

### Task 5.1: Consolidate Redis client via Goravel's facades.Instance

**Files:**
- Modify: `dx-api/app/helpers/redis.go`

- [ ] **Step 1: Open `dx-api/app/helpers/redis.go` and inspect current state**

Current file creates its own `redis.NewClient(...)` singleton. We're replacing it with a wrapper around `redis_facades.Instance("default")`.

- [ ] **Step 2: Replace the file contents**

Replace `dx-api/app/helpers/redis.go` with:

```go
package helpers

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	redis_facades "github.com/goravel/redis/facades"
)

// GetRedis returns the shared Redis client from Goravel's redis facade.
// Uses the "default" connection defined in config/database.go.
// Previously this created its own *redis.Client; consolidated here so that
// cache, queue, session, online tracking, and WebSocket pub/sub all share
// a single client with consistent configuration (TLS, cluster mode, etc.).
func GetRedis() *redis.Client {
	client, err := redis_facades.Instance("default")
	if err != nil {
		panic("redis facade not initialized: " + err.Error())
	}
	// The facade returns UniversalClient; in single-node mode this is a *redis.Client
	if c, ok := client.(*redis.Client); ok {
		return c
	}
	panic("redis facade returned non-*redis.Client (cluster mode not supported in helper)")
}

// RedisSet sets a key with TTL
func RedisSet(key string, value string, ttl time.Duration) error {
	ctx := context.Background()
	return GetRedis().Set(ctx, key, value, ttl).Err()
}

// RedisGet retrieves a value by key
func RedisGet(key string) (string, error) {
	ctx := context.Background()
	return GetRedis().Get(ctx, key).Result()
}

// RedisDel deletes a key
func RedisDel(key string) error {
	ctx := context.Background()
	return GetRedis().Del(ctx, key).Err()
}

// RedisPing checks Redis connectivity
func RedisPing() error {
	ctx := context.Background()
	return GetRedis().Ping(ctx).Err()
}
```

- [ ] **Step 3: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: exit 0. If there are errors about missing imports or the `redis_facades` path, check the actual goravel/redis version's facade package path.

- [ ] **Step 4: Run all existing helper tests to make sure nothing regressed**

```bash
cd dx-api && go test ./app/helpers/... -v
```

Expected: PASS (or if no helper tests exist, just exit 0 from go test).

- [ ] **Step 5: Commit**

```bash
cd dx-api && git add app/helpers/redis.go
git commit -m "refactor(helpers): use redis_facades.Instance for shared client"
```

---

### Task 5.2: Create WSController with detached-context pattern

**Files:**
- Create: `dx-api/app/http/controllers/api/ws_controller.go`

- [ ] **Step 1: Read the existing api controller imports to match conventions**

```bash
cd dx-api && head -20 app/http/controllers/api/ai_custom_controller.go
```

This gives you the import style used by existing controllers in this directory.

- [ ] **Step 2: Create ws_controller.go**

Write `dx-api/app/http/controllers/api/ws_controller.go`:

```go
package api

import (
	"context"
	"net/http"
	"strings"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"github.com/coder/websocket"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/realtime"
)

type WSController struct{}

func NewWSController() *WSController {
	return &WSController{}
}

// Handle upgrades an HTTP request to a WebSocket and attaches the connection
// to the realtime hub. Authentication is handled by the JwtAuth middleware
// before this method is called.
//
// IMPORTANT: This handler uses a detached context (context.Background()) for
// Hub.Attach, NOT ctx.Request().Context(). The Goravel Gin global Timeout
// middleware wraps the request context with context.WithTimeout(..., 30s)
// by default; using the request context would kill the WebSocket after 30s.
// The Gin middleware's post-timeout Abort(408) is a no-op on the hijacked
// net.Conn (standard Go net/http Hijacker contract), so this pattern works
// without modifying Goravel or the timeout middleware.
//
// Graceful shutdown is driven by Hub.Shutdown via Goravel's terminating hook.
func (c *WSController) Handle(ctx contractshttp.Context) contractshttp.Response {
	// If the hub is shutting down, reject the upgrade with 503.
	if hub := realtime.DefaultHub(); hub != nil && hub.IsShuttingDown() {
		return helpers.Error(ctx, http.StatusServiceUnavailable, consts.CodeInternalError, "server shutting down")
	}

	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	origins := facades.Config().GetString("cors.allowed_origins", "")
	originPatterns := strings.Split(origins, ",")
	for i := range originPatterns {
		originPatterns[i] = strings.TrimSpace(originPatterns[i])
	}

	w := ctx.Response().Writer()
	r := ctx.Request().Origin() // underlying *http.Request

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns:  originPatterns,
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		// websocket.Accept already wrote the error response
		return nil
	}
	defer conn.Close(websocket.StatusInternalError, "server error")

	// CRITICAL: detached context — see the method doc above.
	wsCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Blocks until the connection dies (disconnect, error, shutdown).
	hub := realtime.DefaultHub()
	if hub == nil {
		return helpers.Error(ctx, http.StatusServiceUnavailable, consts.CodeInternalError, "realtime hub not initialized")
	}
	_ = hub.Attach(wsCtx, userID, conn)
	return nil
}
```

- [ ] **Step 3: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: exit 0. If `ctx.Request().Origin()` doesn't match Goravel's API, check the actual method — it might be `ctx.Request().Request()` or similar. Adjust based on the contract.

- [ ] **Step 4: Commit**

```bash
cd dx-api && git add app/http/controllers/api/ws_controller.go
git commit -m "feat(api): WSController with detached context for WebSocket upgrades"
```

---

### Task 5.3: Wire realtime.Default and Default_Hub in bootstrap

**Files:**
- Modify: `dx-api/bootstrap/app.go`

- [ ] **Step 1: Read the current bootstrap file**

```bash
cd dx-api && cat bootstrap/app.go
```

Identify where services register and where the application's initialization completes.

- [ ] **Step 2: Add the realtime wiring**

Add to `dx-api/bootstrap/app.go`. The exact insertion point depends on the file's current structure, but conceptually you want a `setupRealtime()` function called during bootstrap:

```go
// Near the top with other imports:
import (
	// ... existing imports ...
	"context"

	"dx-api/app/helpers"
	"dx-api/app/realtime"

	"github.com/goravel/framework/contracts/foundation"
)

// Add this function (typically near other init helpers):

// setupRealtime wires the realtime package's default PubSub and Default_Hub.
// Must be called after Redis is available (after the goravel/redis service
// provider has registered).
func setupRealtime(app foundation.Application) {
	redisClient := helpers.GetRedis()

	ctx := context.Background()
	pubsub := realtime.NewRedisPubSub(ctx, redisClient)
	// NOTE: SetDefault wraps the interface value in an atomic.Pointer so
	// concurrent readers never see a torn interface header. This was changed
	// from `realtime.Default = pubsub` during Task 2.2 re-review to fix a
	// latent data race. See commit aa46de7 for the rationale.
	realtime.SetDefault(pubsub)

	presence := realtime.NewPresence(redisClient)
	authorizer := realtime.NewAuthorizer()

	// NOTE: SetDefaultHub wraps the Hub in an atomic.Pointer for race-safe
	// concurrent access, matching the Task 2.2 fix pattern.
	realtime.SetDefaultHub(realtime.NewHub(pubsub, presence, authorizer))

	// Register the terminating hook for graceful shutdown.
	// The exact API may be app.Terminating(...) or app.Shutdown(...) — check
	// Goravel's foundation.Application interface.
	app.Terminating(func() error {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if hub := realtime.DefaultHub(); hub != nil {
			return hub.Shutdown(shutdownCtx)
		}
		return nil
	})
}
```

In the main bootstrap flow (wherever the app is initialized after providers register), call `setupRealtime(app)`.

**Note about the terminating hook API:** Goravel's foundation.Application interface may use `Terminating`, `Shutdown`, or expose shutdown via a different mechanism. If `app.Terminating(...)` doesn't exist, look at `framework@v1.17.2/contracts/foundation/` for the right method. If no equivalent exists, register a signal handler directly:

```go
go func() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = realtime.Default_Hub.Shutdown(shutdownCtx)
}()
```

Use whichever mechanism Goravel supports.

- [ ] **Step 3: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: exit 0. If the build fails due to the `Terminating` method not existing, fall back to the signal handler approach.

- [ ] **Step 4: Commit**

```bash
cd dx-api && git add bootstrap/app.go
git commit -m "feat(bootstrap): wire realtime.Default and Default_Hub with shutdown hook"
```

---

### Task 5.4: Register the /api/ws route

**Files:**
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Read the current api.go to find the protected (authenticated) route group**

```bash
cd dx-api && grep -n "user/events\|play-pk.*events\|middleware.JwtAuth" routes/api.go
```

Look for the section where authenticated routes are registered (e.g., under a `protected` variable after middleware attach).

- [ ] **Step 2: Add the WS route registration**

In `dx-api/routes/api.go`, find the section that declares controllers near the top and add:

```go
// Near the top where other controllers are instantiated:
wsController := api.NewWSController()
```

And in the protected route group (the one with `middleware.JwtAuth()`), add:

```go
protected.Get("/ws", wsController.Handle)
```

Place it alongside other similar registrations, for example near the existing `/user/events` route (we're not removing the SSE route in this PR — WS and SSE coexist during migration).

- [ ] **Step 3: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: exit 0.

- [ ] **Step 4: Commit**

```bash
cd dx-api && git add routes/api.go
git commit -m "feat(routes): register GET /api/ws with JwtAuth middleware"
```

---

### Task 5.5: Write the full roundtrip integration test

**Files:**
- Create: `dx-api/app/realtime/realtime_integration_test.go`

- [ ] **Step 1: Create the integration test file**

Write `dx-api/app/realtime/realtime_integration_test.go`:

```go
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

	// Handler that upgrades and attaches with a fixed user ID.
	ts := httptest.NewServer(httpHandlerForTest(hub, "alice"))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	// Client subscribes to user:alice
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

	// Allow the hub's pubsub Subscribe to register with Redis
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
			OriginPatterns: []string{"*"}, // permissive for localhost tests
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
```

Note: `net/http` is already in the import block above alongside `net/http/httptest`. If you get a compile error about an undefined `http.HandlerFunc`, double-check the import block includes both.

- [ ] **Step 2: Run the integration tests**

```bash
cd dx-api && go test ./app/realtime/ -run TestIntegration -v -race
```

Expected: both tests PASS. If they fail due to timing, increase the `time.Sleep(100ms)` in each test to 200ms.

- [ ] **Step 3: Commit**

```bash
cd dx-api && git add app/realtime/realtime_integration_test.go
git commit -m "test(realtime): full roundtrip + cross-instance integration tests"
```

---

## Phase 6 — Service call site double-publish

Each task here is one service file. The edits are mechanical: find each legacy `PkHub.X`, `UserHub.X`, `GroupSSEHub.X`, or `GroupNotifyHub.X` call, and add a `realtime.Publish(ctx, Topic(id), Event{...})` call immediately after it. Keep the legacy call.

### Task 6.1: pk_invite_service.go — 3 call sites

**Files:**
- Modify: `dx-api/app/services/api/pk_invite_service.go`

- [ ] **Step 1: List the legacy call sites**

```bash
cd dx-api && grep -n "helpers\.\(UserHub\|PkHub\|GroupSSEHub\|GroupNotifyHub\)" app/services/api/pk_invite_service.go
```

Expected: 3 matches — one `UserHub.SendToUser` (invite) and two `PkHub.Broadcast` (accept, decline).

- [ ] **Step 2: Add `realtime` import**

At the top of the file, add:

```go
"dx-api/app/realtime"
```

to the import block.

- [ ] **Step 3: For each legacy call, add a realtime.Publish alongside it**

**Invite path (around line 143 per spec Section 2):**

```go
// Before: helpers.UserHub.SendToUser(opponentID, "pk_invitation", invitePayload)
// After: (add below the legacy line)
helpers.UserHub.SendToUser(opponentID, "pk_invitation", invitePayload)
_ = realtime.Publish(ctx, realtime.UserTopic(opponentID), realtime.Event{
	Type: "pk_invitation",
	Data: invitePayload,
})
```

**Accept path (around line 211):**

```go
// After the existing PkHub.Broadcast("pk_invitation_accepted", ...)
helpers.PkHub.Broadcast(pkID, "pk_invitation_accepted", acceptedPayload)
_ = realtime.Publish(ctx, realtime.PkTopic(pkID), realtime.Event{
	Type: "pk_invitation_accepted",
	Data: acceptedPayload,
})
// Double-publish rule: also send to initiator's user topic so they're
// notified even if they haven't opened the pk-room page yet
_ = realtime.Publish(ctx, realtime.UserTopic(initiatorID), realtime.Event{
	Type: "pk_invitation_accepted",
	Data: acceptedPayload,
})
```

**Decline path (around line 259):**

```go
helpers.PkHub.Broadcast(pkID, "pk_invitation_declined", declinedPayload)
_ = realtime.Publish(ctx, realtime.PkTopic(pkID), realtime.Event{
	Type: "pk_invitation_declined",
	Data: declinedPayload,
})
_ = realtime.Publish(ctx, realtime.UserTopic(initiatorID), realtime.Event{
	Type: "pk_invitation_declined",
	Data: declinedPayload,
})
```

**Note about `ctx`:** each of these calls runs inside a service function. If the function already takes a context, use it. If not, use `context.Background()` and add `"context"` to the imports.

- [ ] **Step 4: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: exit 0.

- [ ] **Step 5: Run the existing service tests (if any)**

```bash
cd dx-api && go test ./app/services/api/ -v -run TestPkInvite 2>&1 | head -40
```

Expected: either PASS or "no tests to run". The goal is no regressions, and we're not adding new tests for service double-publish because the realtime package has its own coverage.

- [ ] **Step 6: Commit**

```bash
cd dx-api && git add app/services/api/pk_invite_service.go
git commit -m "feat(pk_invite): double-publish events to realtime topics"
```

---

### Task 6.2: game_play_pk_service.go — 6 call sites

**Files:**
- Modify: `dx-api/app/services/api/game_play_pk_service.go`

- [ ] **Step 1: List legacy call sites**

```bash
cd dx-api && grep -n "helpers\.\(UserHub\|PkHub\)" app/services/api/game_play_pk_service.go
```

Expected: ~6 matches (PkHub.Broadcast for complete, player_action, force_end, robot; UserHub.SendToUser for next_level specified).

- [ ] **Step 2: Add `realtime` import if not already present**

- [ ] **Step 3: Add a `_ = realtime.Publish(...)` next to each legacy call**

For each legacy `helpers.PkHub.Broadcast(pkID, "event_name", data)`, add:

```go
_ = realtime.Publish(ctx, realtime.PkTopic(pkID), realtime.Event{
	Type: "event_name",
	Data: data,
})
```

For `helpers.UserHub.SendToUser(userID, "event_name", data)`, add:

```go
_ = realtime.Publish(ctx, realtime.UserTopic(userID), realtime.Event{
	Type: "event_name",
	Data: data,
})
```

For the `NextPkLevel` flow (specified PK), add **both** publishes — the legacy UserHub AND the new realtime — and keep them identical in payload.

Also, in the spec we discussed replacing `PkHub.IsConnected(pkID, userID)` (in-memory presence check) with `realtime.IsPresent(ctx, ...)`. Do this:

```go
// Before:
// if helpers.PkHub.IsConnected(pkID, otherUserID) { ... }
// After:
presentOnPk, _ := realtime.Default_Hub.Presence().IsPresent(ctx, realtime.PkTopic(pkID), otherUserID)
if presentOnPk { ... }
```

This requires adding a getter to the hub. Open `hub.go` and add:

```go
// Presence returns the hub's presence tracker. Nil if the hub was constructed
// without one.
func (h *Hub) Presence() *Presence { return h.presence }
```

If the legacy `IsConnected` check is inside a function where `ctx` isn't available, use `context.Background()` and add the context import.

- [ ] **Step 4: Verify build**

```bash
cd dx-api && go build ./...
```

Expected: exit 0.

- [ ] **Step 5: Run existing tests**

```bash
cd dx-api && go test ./app/services/api/... 2>&1 | tail -20
```

Expected: no regressions.

- [ ] **Step 6: Commit**

```bash
cd dx-api && git add app/services/api/game_play_pk_service.go app/realtime/hub.go
git commit -m "feat(pk_play): double-publish to realtime + cross-instance presence check"
```

---

### Task 6.3: game_play_group_service.go — 4 call sites

**Files:**
- Modify: `dx-api/app/services/api/game_play_group_service.go`

- [ ] **Step 1: List legacy call sites**

```bash
cd dx-api && grep -n "helpers\.\(GroupSSEHub\|GroupNotifyHub\)" app/services/api/game_play_group_service.go
```

Expected: ~4 matches.

- [ ] **Step 2: Add `realtime` import**

- [ ] **Step 3: For each legacy call, add realtime.Publish alongside**

Pattern:

```go
// GroupSSEHub.Broadcast → realtime.Publish with GroupTopic
helpers.GroupSSEHub.Broadcast(groupID, "group_player_complete", payload)
_ = realtime.Publish(ctx, realtime.GroupTopic(groupID), realtime.Event{
	Type: "group_player_complete",
	Data: payload,
})

// GroupNotifyHub.Notify → realtime.Publish with GroupNotifyTopic + group_updated
helpers.GroupNotifyHub.Notify(groupID, "detail")
_ = realtime.Publish(ctx, realtime.GroupNotifyTopic(groupID), realtime.Event{
	Type: "group_updated",
	Data: map[string]string{"scope": "detail"},
})
```

- [ ] **Step 4: Verify build & test**

```bash
cd dx-api && go build ./... && go test ./app/services/api/... 2>&1 | tail -20
```

Expected: exit 0, no regressions.

- [ ] **Step 5: Commit**

```bash
cd dx-api && git add app/services/api/game_play_group_service.go
git commit -m "feat(group_play): double-publish events and notifies to realtime"
```

---

### Task 6.4: group_game_service.go — 6 call sites

**Files:**
- Modify: `dx-api/app/services/api/group_game_service.go`

- [ ] **Step 1: List legacy call sites**

```bash
cd dx-api && grep -n "helpers\.\(GroupSSEHub\|GroupNotifyHub\)" app/services/api/group_game_service.go
```

- [ ] **Step 2: Add `realtime` import**

- [ ] **Step 3: For each legacy call, add realtime.Publish alongside**

Same pattern as Task 6.3. Specifically handle:
- `StartGroupGame` — publishes `group_game_start` on GroupTopic + `group_updated:detail` on GroupNotifyTopic
- `SetGroupGame`, `ClearGroupGame` — `group_updated:detail` on GroupNotifyTopic
- `ForceEndGroupGame` — `group_game_force_end` on GroupTopic + `group_updated:detail` on notify
- `NextGroupLevel` — `group_next_level` on GroupTopic + `group_updated:detail` on notify

Also replace any `GroupSSEHub.ConnectedUserIDs(groupID)` calls with `realtime.Default_Hub.Presence().Members(ctx, realtime.GroupTopic(groupID))` where possible. This is used for winner determination.

- [ ] **Step 4: Verify build & test**

```bash
cd dx-api && go build ./... && go test ./app/services/api/... 2>&1 | tail -20
```

- [ ] **Step 5: Commit**

```bash
cd dx-api && git add app/services/api/group_game_service.go
git commit -m "feat(group_game): double-publish events to realtime + presence-based roster"
```

---

### Task 6.5: group_service.go — 1 call site

**Files:**
- Modify: `dx-api/app/services/api/group_service.go`

- [ ] **Step 1: Find the `group_dismissed` broadcast**

```bash
cd dx-api && grep -n "helpers\.GroupSSEHub" app/services/api/group_service.go
```

- [ ] **Step 2: Add realtime publish**

Next to the legacy line:

```go
helpers.GroupSSEHub.Broadcast(groupID, "group_dismissed", map[string]string{"group_id": groupID})
_ = realtime.Publish(ctx, realtime.GroupTopic(groupID), realtime.Event{
	Type: "group_dismissed",
	Data: map[string]string{"group_id": groupID},
})
```

- [ ] **Step 3: Verify build**

```bash
cd dx-api && go build ./...
```

- [ ] **Step 4: Commit**

```bash
cd dx-api && git add app/services/api/group_service.go
git commit -m "feat(group): double-publish group_dismissed to realtime"
```

---

### Task 6.6: group_member_service.go — 2 call sites

**Files:**
- Modify: `dx-api/app/services/api/group_member_service.go`

- [ ] **Step 1: Find legacy calls**

```bash
cd dx-api && grep -n "helpers\.GroupNotifyHub" app/services/api/group_member_service.go
```

- [ ] **Step 2: Add realtime publishes (KickMember, LeaveGroup paths)**

For each legacy `helpers.GroupNotifyHub.Notify(groupID, "members")` or `"detail"`, add:

```go
_ = realtime.Publish(ctx, realtime.GroupNotifyTopic(groupID), realtime.Event{
	Type: "group_updated",
	Data: map[string]string{"scope": "members"},
})
```

- [ ] **Step 3: Verify build & commit**

```bash
cd dx-api && go build ./...
cd dx-api && git add app/services/api/group_member_service.go
git commit -m "feat(group_member): double-publish notifies to realtime"
```

---

### Task 6.7: group_application_service.go — 4 call sites

**Files:**
- Modify: `dx-api/app/services/api/group_application_service.go`

- [ ] **Step 1: Find legacy calls**

```bash
cd dx-api && grep -n "helpers\.GroupNotifyHub" app/services/api/group_application_service.go
```

- [ ] **Step 2: Add realtime publishes for each**

Same pattern — for each `helpers.GroupNotifyHub.Notify(groupID, scope)`, add:

```go
_ = realtime.Publish(ctx, realtime.GroupNotifyTopic(groupID), realtime.Event{
	Type: "group_updated",
	Data: map[string]string{"scope": "applications"},  // or "members", "detail", as matches the legacy call
})
```

- [ ] **Step 3: Verify build & commit**

```bash
cd dx-api && go build ./...
cd dx-api && git add app/services/api/group_application_service.go
git commit -m "feat(group_application): double-publish notifies to realtime"
```

---

### Task 6.8: group_subgroup_service.go — 5 call sites

**Files:**
- Modify: `dx-api/app/services/api/group_subgroup_service.go`

- [ ] **Step 1: Find legacy calls**

```bash
cd dx-api && grep -n "helpers\.GroupNotifyHub" app/services/api/group_subgroup_service.go
```

- [ ] **Step 2: Add realtime publishes**

```go
_ = realtime.Publish(ctx, realtime.GroupNotifyTopic(groupID), realtime.Event{
	Type: "group_updated",
	Data: map[string]string{"scope": "subgroups"},
})
```

For each of the 5 sites (Create, Update, Delete, AssignMember, RemoveMember).

- [ ] **Step 3: Verify build & commit**

```bash
cd dx-api && go build ./...
cd dx-api && git add app/services/api/group_subgroup_service.go
git commit -m "feat(group_subgroup): double-publish notifies to realtime"
```

---

## Phase 7 — Issue C: Mid-session single-device kick

The login flow updates `user_auth:{userID}:user` in Redis. We publish a kick event on the same instance. Every Hub instance automatically subscribes every connected client to its own `user:{userID}:kick` topic (this is already done in `client.Serve` from Task 4.2). When the kick event arrives, the client receives it, and we close the connection with code 4001.

### Task 7.1: Publish kick on login

**Files:**
- Modify: `dx-api/app/services/api/auth_service.go`

- [ ] **Step 1: Find the login timestamp write at line 28**

```bash
cd dx-api && grep -n "user_auth:.*:user" app/services/api/auth_service.go
```

Expected: line ~28 is the SET, line ~184 is the DEL (logout).

- [ ] **Step 2: Add realtime import**

- [ ] **Step 3: After the RedisSet succeeds, publish the kick**

```go
// ... existing: helpers.RedisSet(fmt.Sprintf("user_auth:%s:user", userID), loginTs, ttl)
// Add after:
_ = realtime.Publish(context.Background(), realtime.UserKickTopic(userID), realtime.Event{
	Type: "session_replaced",
	Data: map[string]string{"reason": "another_device"},
})
```

**Context note:** use `context.Background()` here because this publish must outlive the HTTP request — we're notifying OTHER connections (on any instance) that may be held by goroutines unrelated to this request.

- [ ] **Step 4: Verify build**

```bash
cd dx-api && go build ./...
```

- [ ] **Step 5: Commit**

```bash
cd dx-api && git add app/services/api/auth_service.go
git commit -m "feat(auth): publish session_replaced kick on new login (Issue C)"
```

---

### Task 7.2: Hub handles kick event by force-closing client

The auto-subscribe of each client to its own `user:{userID}:kick` topic is already done in `client.Serve` (Task 4.2). When an event arrives on that topic, it flows through `fanout → enqueue → writeLoop → wsjson.Write`. We need the client to recognize the kick and close the WS with code 4001 instead of (or in addition to) writing it to the wire.

**Files:**
- Modify: `dx-api/app/realtime/client.go`

- [ ] **Step 1: Modify the `enqueue` method to detect session_replaced events**

Replace the `enqueue` method in `dx-api/app/realtime/client.go`:

```go
// enqueue tries to add an envelope to the send channel without blocking.
// If the channel is full, the client is considered too slow and kicked.
// If the envelope is a session_replaced kick event, the client is closed
// with code 4001 instead of being written to the wire.
func (c *Client) enqueue(env Envelope) {
	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if closed {
		return
	}

	// Issue C: detect session_replaced kick on the user's kick topic and
	// force-close the connection with WS code 4001.
	if env.Op == OpEvent && env.Type == "session_replaced" {
		if parsed, err := ParseTopic(env.Topic); err == nil && parsed.Kind == KindUserKick && parsed.ID == c.userID {
			c.mu.Lock()
			if !c.closed {
				c.closed = true
				c.mu.Unlock()
				go func() {
					_ = c.conn.Close(4001, "session_replaced")
				}()
				return
			}
			c.mu.Unlock()
			return
		}
	}

	select {
	case c.send <- env:
	default:
		c.mu.Lock()
		if !c.closed {
			c.closed = true
			c.mu.Unlock()
			c.hub.kickSlowConsumer(c)
			return
		}
		c.mu.Unlock()
	}
}
```

- [ ] **Step 2: Add a test for the kick detection**

Add to `dx-api/app/realtime/hub_test.go`:

```go
func TestClient_SessionReplacedKickDetected(t *testing.T) {
	hub := NewHub(newFakePubSub(), nil, allowAllAuthorizer())

	// Create a client with a mocked conn — nil conn will panic on Close,
	// so we use a test that only verifies the closed flag is set.
	c := &Client{
		hub:    hub,
		userID: "alice",
		topics: map[string]struct{}{"user:alice:kick": {}},
		send:   make(chan Envelope, 16),
	}

	// The Close call will panic on nil conn; catch it to verify the flag
	// transition is what we expected to happen.
	done := make(chan struct{})
	go func() {
		defer func() { recover(); close(done) }()
		c.enqueue(Envelope{
			Op:    OpEvent,
			Topic: "user:alice:kick",
			Type:  "session_replaced",
		})
	}()
	<-done

	// Give the goroutine scheduled by enqueue a moment to run
	time.Sleep(10 * time.Millisecond)

	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if !closed {
		t.Error("client should be marked closed after session_replaced kick")
	}
}
```

- [ ] **Step 3: Run the tests**

```bash
cd dx-api && go test ./app/realtime/ -run "TestClient_SessionReplaced" -v
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
cd dx-api && git add app/realtime/client.go app/realtime/hub_test.go
git commit -m "feat(realtime): close client with code 4001 on session_replaced kick"
```

---

## Phase 8 — Frontend: dormant provider + hooks (no tests — no framework)

**Note:** `dx-web` has no frontend test framework configured (no Jest, no Vitest, zero source-level test files). The existing SSE hooks ship without tests. PR 1 ships the new provider and hooks without tests, matching existing project convention. Setting up a frontend test framework is a separate future initiative the user can tackle when they want coverage.

### Task 8.1: Create WebSocketProvider

**Files:**
- Create: `dx-web/src/providers/websocket-provider.tsx`

- [ ] **Step 1: Create the provider file**

Write `dx-web/src/providers/websocket-provider.tsx`:

```tsx
"use client";

import { createContext, useContext, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "";

type EventHandler = (event: { type: string; data: unknown }) => void;

type WSContextValue = {
  status: "connecting" | "open" | "closed";
  subscribe: (topic: string, handler: EventHandler) => () => void;
};

const WSContext = createContext<WSContextValue | null>(null);

type Envelope = {
  op: "subscribe" | "unsubscribe" | "event" | "ack" | "error";
  topic?: string;
  type?: string;
  data?: unknown;
  id?: string;
  ok?: boolean;
  code?: number;
  message?: string;
};

function randomId(): string {
  return Math.random().toString(36).slice(2) + Date.now().toString(36);
}

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const [status, setStatus] = useState<"connecting" | "open" | "closed">("connecting");

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttempt = useRef(0);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const hasEverConnected = useRef(false);

  // topic -> set of handlers (ref-counted subscription)
  const subsRef = useRef<Map<string, Set<EventHandler>>>(new Map());
  // topic -> true once acked in the current connection
  const ackedRef = useRef<Set<string>>(new Set());

  const router = useRouter();

  useEffect(() => {
    let cancelled = false;

    const sendSubscribe = (ws: WebSocket, topic: string) => {
      const env: Envelope = { op: "subscribe", topic, id: randomId() };
      ws.send(JSON.stringify(env));
    };

    const scheduleReconnect = () => {
      const n = reconnectAttempt.current++;
      if (n >= 10) {
        toast.error("连接已断开，请刷新页面");
        return;
      }
      const base = Math.min(1000 * Math.pow(2, n), 30000);
      const jitter = Math.random() * Math.min(base * 0.3, 5000);
      const delay = base + jitter;
      reconnectTimer.current = setTimeout(connect, delay);
    };

    const routeIncoming = (env: Envelope) => {
      if (env.op === "event" && env.topic && env.type) {
        const handlers = subsRef.current.get(env.topic);
        handlers?.forEach((h) => h({ type: env.type!, data: env.data }));
        return;
      }
      if (env.op === "ack" && env.ok && env.topic) {
        ackedRef.current.add(env.topic);
        return;
      }
      // error and unsuccessful acks: ignore quietly for now. A toast can be
      // added later if useful — most failures are unusable topics that
      // components shouldn't have subscribed to anyway.
    };

    const connect = () => {
      if (cancelled) return;
      setStatus("connecting");
      const url = `${API_URL}/api/ws`;
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        setStatus("open");
        hasEverConnected.current = true;
        reconnectAttempt.current = 0;
        ackedRef.current.clear();
        // Resubscribe every topic with at least one handler
        for (const topic of subsRef.current.keys()) {
          sendSubscribe(ws, topic);
        }
      };

      ws.onmessage = (ev: MessageEvent) => {
        try {
          const env = JSON.parse(ev.data as string) as Envelope;
          routeIncoming(env);
        } catch {
          // Discard malformed frames
        }
      };

      ws.onclose = (ev: CloseEvent) => {
        wsRef.current = null;
        setStatus("closed");
        ackedRef.current.clear();

        if (cancelled) return;

        // Issue C: session replaced
        if (ev.code === 4001) {
          toast.error("您的账号已在其他设备登录");
          router.push("/auth/signin?reason=session_replaced");
          return;
        }
        // Session expired beyond refresh window
        if (ev.code === 4401) {
          router.push("/auth/signin?reason=session_expired");
          return;
        }
        // Abnormal close BEFORE ever connecting successfully → likely auth
        // failure at upgrade (no close frame in that case).
        if (ev.code === 1006 && !hasEverConnected.current) {
          router.push("/auth/signin?reason=session_expired");
          return;
        }
        // Normal close during unmount
        if (ev.code === 1000) {
          return;
        }
        // Anything else: reconnect with backoff
        scheduleReconnect();
      };

      ws.onerror = () => {
        // onclose will fire right after, letting that handler drive the flow
      };
    };

    connect();

    return () => {
      cancelled = true;
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      if (wsRef.current) {
        wsRef.current.close(1000, "unmount");
        wsRef.current = null;
      }
    };
  }, [router]);

  const subscribe = (topic: string, handler: EventHandler) => {
    let set = subsRef.current.get(topic);
    if (!set) {
      set = new Set();
      subsRef.current.set(topic, set);
      // First handler for this topic: send wire subscribe if open
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        const env: Envelope = { op: "subscribe", topic, id: randomId() };
        wsRef.current.send(JSON.stringify(env));
      }
    }
    set.add(handler);

    return () => {
      const current = subsRef.current.get(topic);
      if (!current) return;
      current.delete(handler);
      if (current.size === 0) {
        subsRef.current.delete(topic);
        ackedRef.current.delete(topic);
        if (wsRef.current?.readyState === WebSocket.OPEN) {
          const env: Envelope = { op: "unsubscribe", topic };
          wsRef.current.send(JSON.stringify(env));
        }
      }
    };
  };

  return (
    <WSContext.Provider value={{ status, subscribe }}>
      {children}
    </WSContext.Provider>
  );
}

export function useWS(): WSContextValue {
  const ctx = useContext(WSContext);
  if (!ctx) {
    throw new Error("useWS must be used within WebSocketProvider");
  }
  return ctx;
}
```

- [ ] **Step 2: Verify TypeScript builds**

```bash
cd dx-web && npm run build 2>&1 | tail -20
```

Expected: build success. If there are type errors, fix them inline. The provider isn't imported anywhere else yet, so build success = provider is self-consistent.

- [ ] **Step 3: Commit**

```bash
cd dx-web && git add src/providers/websocket-provider.tsx
git commit -m "feat(web): add WebSocketProvider with reconnect, close code handling"
```

---

### Task 8.2: Create useWebSocket re-export + useTopic helper

**Files:**
- Create: `dx-web/src/hooks/use-websocket.ts`
- Create: `dx-web/src/hooks/use-topic.ts`

- [ ] **Step 1: Create use-websocket.ts**

Write `dx-web/src/hooks/use-websocket.ts`:

```typescript
"use client";

export { useWS as useWebSocket } from "@/providers/websocket-provider";
```

- [ ] **Step 2: Create use-topic.ts**

Write `dx-web/src/hooks/use-topic.ts`:

```typescript
"use client";

import { useEffect, useRef } from "react";

import { useWS } from "@/providers/websocket-provider";

export function useTopic(
  topic: string | null,
  handlers: Record<string, (data: unknown) => void>,
): void {
  const { subscribe } = useWS();

  // Always call the latest handlers without re-subscribing
  const handlersRef = useRef(handlers);
  useEffect(() => {
    handlersRef.current = handlers;
  });

  useEffect(() => {
    if (!topic) return;
    const unsubscribe = subscribe(topic, (event) => {
      handlersRef.current[event.type]?.(event.data);
    });
    return unsubscribe;
  }, [topic, subscribe]);
}
```

- [ ] **Step 3: Verify build**

```bash
cd dx-web && npm run build 2>&1 | tail -10
```

Expected: success.

- [ ] **Step 4: Commit**

```bash
cd dx-web && git add src/hooks/use-websocket.ts src/hooks/use-topic.ts
git commit -m "feat(web): add useWebSocket re-export and useTopic helper"
```

---

### Task 8.3: Create 4 feature wrapper hooks

**Files:**
- Create: `dx-web/src/hooks/use-user-events.ts`
- Create: `dx-web/src/hooks/use-pk-events.ts`
- Create: `dx-web/src/hooks/use-group-events.ts`
- Create: `dx-web/src/hooks/use-group-notify-ws.ts`

- [ ] **Step 1: Create use-user-events.ts**

Write `dx-web/src/hooks/use-user-events.ts`:

```typescript
"use client";

import { useTopic } from "@/hooks/use-topic";

/**
 * Subscribe to the current user's global event topic (user:{userID}).
 * Receives events like pk_invitation, pk_next_level.
 *
 * The userID must be provided by the caller (e.g., from an auth context or
 * from a server-action-provided props). If null, no subscription happens.
 */
export function useUserEvents(
  userId: string | null,
  listeners: Record<string, (data: unknown) => void>,
): void {
  useTopic(userId ? `user:${userId}` : null, listeners);
}
```

- [ ] **Step 2: Create use-pk-events.ts**

Write `dx-web/src/hooks/use-pk-events.ts`:

```typescript
"use client";

import { useTopic } from "@/hooks/use-topic";

/**
 * Subscribe to events for a specific PK match (pk:{pkId}).
 * Receives events like pk_player_action, pk_player_complete, pk_force_end.
 */
export function usePkEvents(
  pkId: string | null,
  listeners: Record<string, (data: unknown) => void>,
): void {
  useTopic(pkId ? `pk:${pkId}` : null, listeners);
}
```

- [ ] **Step 3: Create use-group-events.ts**

Write `dx-web/src/hooks/use-group-events.ts`:

```typescript
"use client";

import { useTopic } from "@/hooks/use-topic";

/**
 * Subscribe to events for a specific group room (group:{groupId}).
 * Receives game events and room_member_joined/left presence events.
 *
 * Note: the existing feature-level hook at
 * features/web/groups/hooks/use-group-events.ts will be renamed to
 * use-group-room-events.ts in PR 4 to avoid shadowing this primitive-level
 * hook.
 */
export function useGroupEvents(
  groupId: string | null,
  listeners: Record<string, (data: unknown) => void>,
): void {
  useTopic(groupId ? `group:${groupId}` : null, listeners);
}
```

- [ ] **Step 4: Create use-group-notify-ws.ts**

Write `dx-web/src/hooks/use-group-notify-ws.ts`:

```typescript
"use client";

import { useTopic } from "@/hooks/use-topic";

/**
 * Subscribe to group metadata updates (group:{groupId}:notify).
 * Receives a single event type `group_updated` with a scope field in data.
 *
 * This file will be renamed to use-group-notify.ts in PR 5 once the legacy
 * SSE use-group-notify.ts is deleted — so components never have to change
 * their import path.
 */
export function useGroupNotifyWs(
  groupId: string | null,
  onScope: (scope: string) => void,
): void {
  useTopic(
    groupId ? `group:${groupId}:notify` : null,
    {
      group_updated: (data) => {
        const d = data as { scope?: string };
        if (d.scope) onScope(d.scope);
      },
    },
  );
}
```

- [ ] **Step 5: Verify build**

```bash
cd dx-web && npm run build 2>&1 | tail -10
```

Expected: success. The new hooks exist but aren't imported anywhere yet, so they're dormant.

- [ ] **Step 6: Commit**

```bash
cd dx-web && git add src/hooks/use-user-events.ts src/hooks/use-pk-events.ts src/hooks/use-group-events.ts src/hooks/use-group-notify-ws.ts
git commit -m "feat(web): add dormant WS wrapper hooks (user/pk/group/notify)"
```

---

## Phase 9 — Infrastructure: nginx prod config

### Task 9.1: Add Upgrade/Connection headers + longer timeouts in nginx.prod.conf

**Files:**
- Modify: `deploy/nginx/nginx.prod.conf`

- [ ] **Step 1: Read the current prod config**

```bash
cat /Users/rainsen/Programs/Projects/douxue/dx-source/deploy/nginx/nginx.prod.conf
```

Find the `location /api/` block.

- [ ] **Step 2: Update the /api/ location block**

Inside the `location /api/ { ... }` block, ensure these directives are present:

```nginx
location /api/ {
    proxy_pass http://api;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # WebSocket upgrade support (added for WS refactor)
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection $connection_upgrade;

    # Long timeouts for WebSocket connections
    proxy_read_timeout 3600s;
    proxy_send_timeout 3600s;

    # Keep SSE-friendly settings (will be removed in PR 6)
    proxy_buffering off;
    proxy_cache off;
}
```

And at the `http { }` level (or the top of the file, above any `server { }` block), add the `$connection_upgrade` map if not already present:

```nginx
map $http_upgrade $connection_upgrade {
    default upgrade;
    ''      close;
}
```

This map is a standard nginx idiom for WebSocket support — it only sets `Connection: upgrade` when the client actually sent an `Upgrade` header. The dev config already has this implicitly via a different mechanism; for prod we want it explicit.

- [ ] **Step 3: Validate nginx config syntax using docker compose**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && docker compose -f deploy/docker-compose.prod.yml config --quiet
```

Expected: exit 0. If the compose file has issues, they're unrelated to our nginx change.

To validate the nginx config itself, we can run nginx -t inside a container:

```bash
docker run --rm -v /Users/rainsen/Programs/Projects/douxue/dx-source/deploy/nginx/nginx.prod.conf:/etc/nginx/nginx.conf:ro nginx:alpine nginx -t
```

Expected: `nginx: the configuration file /etc/nginx/nginx.conf syntax is ok` and `test is successful`.

- [ ] **Step 4: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add deploy/nginx/nginx.prod.conf
git commit -m "fix(nginx): add Upgrade/Connection headers and 1h timeouts for /api/"
```

---

## Phase 10 — Final verification

### Task 10.1: Run the full backend test suite

**Files:** none (verification only)

- [ ] **Step 1: Run all Go tests with race detector**

```bash
cd dx-api && go test -race ./... 2>&1 | tail -40
```

Expected: all tests PASS, no race conditions detected. The realtime package tests should show ~30+ tests passing. Legacy service tests (if any) should still pass.

- [ ] **Step 2: Run `go vet`**

```bash
cd dx-api && go vet ./...
```

Expected: exit 0, no warnings.

- [ ] **Step 3: Run golangci-lint (if configured)**

```bash
cd dx-api && golangci-lint run ./... 2>&1 | tail -40
```

Expected: exit 0. If golangci-lint is not installed, skip this step. If issues are reported, fix them — the user requirement is "no lint issues".

- [ ] **Step 4: Run `go build ./...`**

```bash
cd dx-api && go build ./...
```

Expected: exit 0.

---

### Task 10.2: Run the full frontend lint and build

**Files:** none (verification only)

- [ ] **Step 1: Run lint**

```bash
cd dx-web && npm run lint 2>&1 | tail -20
```

Expected: exit 0 (no lint errors). If errors are reported in the new files, fix them — the user requirement is "no lint issues".

- [ ] **Step 2: Run build**

```bash
cd dx-web && npm run build 2>&1 | tail -20
```

Expected: build success.

---

### Task 10.3: Docker Compose dev smoke test

**Files:** none (verification only)

- [ ] **Step 1: Start the dev stack**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && docker compose -f deploy/docker-compose.dev.yml up -d --build
```

Expected: all 5 services (nginx, dx-api, dx-web, postgres, redis) start healthy. Check with:

```bash
docker compose -f deploy/docker-compose.dev.yml ps
```

- [ ] **Step 2: Verify health endpoint**

```bash
curl -s http://localhost/api/health
```

Expected: JSON response with `"code": 0` and `"data": {"db": true, "redis": true}`.

- [ ] **Step 3: Verify /api/ws accepts WebSocket upgrade (via websocat if available)**

Install websocat if needed: `brew install websocat` on macOS.

```bash
# Get a valid dx_token cookie by signing in via the web UI first, then:
websocat --header="Cookie: dx_token=<your-token>" ws://localhost/api/ws
```

Expected: the connection stays open. Now paste this subscribe frame:

```json
{"op":"subscribe","topic":"user:<your-user-id>","id":"req_1"}
```

Expected response:

```json
{"op":"ack","id":"req_1","ok":true}
```

Type the Ctrl+D to close.

- [ ] **Step 4: Trigger an actual PK invite and verify double-publish**

In a second terminal, still connected via websocat to `user:<your-user-id>`, use the existing UI (or a second logged-in account) to send a PK invite to your test user. The existing SSE path should still trigger the popup in the browser, AND the websocat session should receive the event:

```json
{"op":"event","topic":"user:<your-user-id>","type":"pk_invitation","data":{...}}
```

This confirms the double-publish is working — the legacy SSE path delivers to the browser, and the new WS path delivers to the websocat test connection.

- [ ] **Step 5: Stop the stack**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && docker compose -f deploy/docker-compose.dev.yml down
```

---

### Task 10.4: Verify prod build path

**Files:** none (verification only)

- [ ] **Step 1: Build all prod images**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && docker compose -f deploy/docker-compose.prod.yml build
```

Expected: all images build successfully. This catches any Dockerfile issues early.

---

### Task 10.5: Final self-review before opening the PR

**Files:** none (review only)

- [ ] **Step 1: Review the diff**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git log --oneline main..HEAD
```

Expected: ~30 commits covering all phases.

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git diff main..HEAD --stat | tail -20
```

- [ ] **Step 2: Confirm no unintended changes**

Spot-check with:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git diff main..HEAD -- 'dx-api/app/helpers/sse_*.go'
```

Expected: no changes to the SSE hub files (they stay functional during migration; deletion is in PR 6).

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git diff main..HEAD -- 'dx-web/src/hooks/use-*-sse.ts'
```

Expected: no changes to the existing SSE hooks (they stay in place, PR 2-5 migrate consumers one at a time, PR 6 deletes them).

- [ ] **Step 3: Regression check — the 4 SSE endpoints still work**

With the dev stack up:

```bash
curl -v -N -H "Cookie: dx_token=<your-token>" http://localhost/api/user/events 2>&1 | head -20
```

Expected: `HTTP/1.1 200 OK` with `Content-Type: text/event-stream`. This confirms the legacy SSE endpoint still serves correctly — no regressions.

- [ ] **Step 4: Push branch and open PR**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git push -u origin feat/websocket-refactor-pr1
```

Use `gh pr create` with a title like `feat: WebSocket refactor PR 1 — backend realtime layer + dormant provider` and include:
- Link to the spec: `docs/superpowers/specs/2026-04-11-websocket-refactor-design.md`
- Link to this plan: `docs/superpowers/plans/2026-04-11-websocket-refactor-pr1.md`
- The PR 1 manual verification checklist from Section 11 of the spec
- A clear "What's NOT in this PR" section listing: mounting the provider, migrating any hook, deleting any SSE code, NDJSON cutover, stale doc deletion

---

## Acceptance criteria for this PR

The PR is ready to merge when **all** of these are true:

- [ ] `go test -race ./...` in `dx-api` passes
- [ ] `go build ./...` in `dx-api` passes
- [ ] `go vet ./...` in `dx-api` passes
- [ ] `golangci-lint run ./...` in `dx-api` reports zero issues (if installed)
- [ ] `npm run lint` in `dx-web` reports zero issues
- [ ] `npm run build` in `dx-web` succeeds
- [ ] `docker compose -f deploy/docker-compose.dev.yml up` starts cleanly, health check returns 200
- [ ] `docker compose -f deploy/docker-compose.prod.yml build` succeeds
- [ ] nginx config test passes: `docker run --rm -v .../nginx.prod.conf:/etc/nginx/nginx.conf:ro nginx:alpine nginx -t`
- [ ] `websocat` can upgrade to `/api/ws`, subscribe to a user topic, receive an ack
- [ ] Triggering a PK invite shows up on both the legacy SSE path and the new WS path (double-publish working)
- [ ] All 4 existing SSE endpoints still respond with `200 OK` and `Content-Type: text/event-stream`
- [ ] The browser's existing flows (PK invite popup, group room, group notify) work exactly as before
- [ ] No files under `dx-api/app/helpers/sse_*.go` or `dx-web/src/hooks/use-*-sse.ts` are modified
- [ ] PR description links to the spec and this plan and lists the manual verification checklist

---

## What comes after this PR

Once PR 1 is merged and soaks in production for 24 hours without issues:

1. **PR 2** — Mount `WebSocketProvider` in `hall/layout.tsx` and migrate `PkInvitationProvider` from `useUserSSE` to `useUserEvents`. First user-visible change. Plan to be generated on demand.
2. **PR 3** — Migrate `usePkSSE` → `usePkEvents`
3. **PR 4** — Migrate `useGroupSSE` → `useGroupEvents` (play + room)
4. **PR 5** — Migrate `useGroupNotify` → new WS version
5. **PR 6** — Cleanup: delete SSE hubs, SSE hooks, remove double-publish, NDJSON cutover for ai-custom, delete 5 stale docs

Each of the remaining 5 PRs will get its own short plan document generated from the same spec. The patterns established here (especially the `realtime.Publish` interface and the double-publish rule) make those migrations mechanical.

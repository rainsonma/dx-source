# WebSocket Refactor Design

**Date:** 2026-04-11
**Scope:** Replace the 4 hub-based SSE real-time push connections in `dx-api` with a single multiplexed WebSocket per user, ready for horizontal scaling via Redis pub/sub. Leave the ai-custom HTTP streaming endpoints in place but rehome their framing from `text/event-stream` to `application/x-ndjson`.

**Deployment:** This project runs via Docker Compose. All infrastructure changes land under `deploy/` (nginx, docker-compose.yml variants, environment files). All manual verification uses `docker compose -f deploy/docker-compose.dev.yml up` or `... docker-compose.prod.yml up` as appropriate.

---

## 1. Executive Summary

### Problem
The `dx-api` backend uses Server-Sent Events (SSE) for all real-time push: user notifications, PK matches, group games, and group metadata updates. The current implementation has known reliability issues (`PkHub` broadcasts fail during idle phases) and is architecturally single-instance (in-memory connection maps, no cross-instance fan-out), blocking any horizontal scaling.

### Solution
A multiplexed WebSocket protocol per user, with Redis pub/sub as the cross-instance fan-out substrate. One persistent WebSocket per `/hall/*` session, carrying subscribe/unsubscribe control frames and event frames for four topic families: `user`, `pk`, `group`, `group:notify`. All hub-based SSE is replaced; ai-custom HTTP streaming is retained with NDJSON framing.

### Locked decisions

| # | Decision | Choice | Rationale summary |
|---|---|---|---|
| 1 | Transport architecture | Single multiplexed `/api/ws` per user | Matches modern best practice (Slack, Linear); one TCP connection, one auth, one reconnect path; eliminates "global SSE went idle" class of bugs |
| 2 | Scope for ai-custom | HTTP streaming retained, framing switched to NDJSON | Request/response semantics don't fit WebSocket; avoids head-of-line blocking with gameplay events; removes `text/event-stream` from codebase |
| 3 | Multi-instance model | Redis pub/sub from day 1 (M2) | Project plans horizontal scaling; avoids deferred rewrite; Redis already in the stack; negligible single-instance cost (<1ms extra per event) |
| 4 | WebSocket library | `github.com/coder/websocket` | Context-native API matches Goravel/go-redis style; built-in ping/pong; `wsjson` helper eliminates boilerplate; concurrent-safe writes |
| 5 | Migration strategy | Parallel shadow, 6 sequential PRs | Constraint: "don't break existing functions"; per-PR rollback via `git revert`; incremental verification at each phase |

### Out of scope
- Load testing beyond unit + integration
- Chaos engineering (random Redis kills, network partitions)
- E2E browser automation (Playwright/Cypress)
- Message replay or persistent event log
- Exactly-once delivery semantics
- Binary-framed messages (JSON text frames only)

---

## 2. Current State Inventory

### Backend SSE layer (all single-instance, in-memory)

| Hub | File | Endpoint | Connection key | Redis usage |
|---|---|---|---|---|
| `UserHub` | `app/helpers/sse_user_hub.go` | `GET /api/user/events` | `userID → *SSEConnection` | `SADD online_users` / `SREM` tracking only |
| `PkHub` | `app/helpers/sse_pk_hub.go` | `GET /api/play-pk/{id}/events` | `pkID → userID → *SSEConnection` | none |
| `GroupSSEHub` | `app/helpers/sse_hub.go` | `GET /api/groups/{id}/events` | `groupID → userID → *SSEConnection` | none |
| `GroupNotifyHub` | `app/helpers/sse_notify_hub.go` | `GET /api/groups/{id}/notify` | `groupID → userID → *SSEConnection` | none |

A 5th endpoint exists for online-bridging: `POST /api/user/ping` (`user_sse_controller.go`) that just writes to the `online_users` Redis set to mark a user online before their SSE connection fully establishes. This will go away in the refactor (the WS upgrade itself is the online signal).

### Frontend SSE consumers (4 core hooks + 3 wrappers)

| Core hook | File | URL | Auth | Notes |
|---|---|---|---|---|
| `useUserSSE` | `src/hooks/use-user-sse.ts` | `/api/user/events` | cookie (`withCredentials`) | Uses `onmessage` + envelope `{type, payload}` (Safari bug workaround) |
| `usePkSSE` | `src/hooks/use-pk-sse.ts` | `/api/play-pk/{pkId}/events` | cookie | Named events via `addEventListener` |
| `useGroupSSE` | `src/hooks/use-group-sse.ts` | `/api/groups/{groupId}/events` | cookie | Named events |
| `useGroupNotify` | `src/hooks/use-group-notify.ts` | `/api/groups/{groupId}/notify` | cookie | Single `group_updated` event with scope |

Feature-level wrappers (unchanged in public API by the refactor):
- `features/web/play-pk/components/pk-invitation-provider.tsx` (wraps `useUserSSE`, mounted at `app/(web)/hall/layout.tsx:13`)
- `features/web/play-pk/hooks/use-pk-play-events.ts` (wraps `usePkSSE`)
- `features/web/play-group/hooks/use-group-play-events.ts` (wraps `useGroupSSE`)
- `features/web/groups/hooks/use-group-events.ts` (wraps `useGroupSSE`)

### Deployment (Docker Compose)

| Component | Dev compose | Prod compose |
|---|---|---|
| dx-api | `docker-compose.dev.yml` with Air hot reload | `docker-compose.prod.yml`, multi-stage Alpine build |
| dx-web | `docker-compose.dev.yml` with `npm run dev` | Standalone Next.js production build |
| nginx | `nginx.dev.conf` (HMR WS support on `/`) | `nginx.prod.conf` (**missing `Upgrade`/`Connection` headers on `/api/*`** — blocker for WS) |
| Postgres | `postgres:18 + pg_partman` | same |
| Redis | `redis:7-alpine` | same |

### AI streaming endpoints (separate from the SSE hubs)

4 POST endpoints stream `text/event-stream` framing over a single HTTP request:
- `POST /api/ai-custom/break-metadata`
- `POST /api/ai-custom/generate-content-items`
- `POST /api/ai-custom/break-vocab-metadata`
- `POST /api/ai-custom/generate-vocab-content-items`

Frontend consumes via `fetch() + res.body.getReader() + TextDecoder` in `features/web/ai-custom/helpers/stream-progress.ts` — **not** `EventSource`. These stay on HTTP with NDJSON framing in the cleanup PR.

---

## 3. Issues Found During Design Exploration

### A — Stale documentation
Five doc files describe designs that do not match current code. Known-wrong references include `docs/game-word-sentence-group-rule.md:138-148` describing a `refreshAccessToken()` function, `dx_refresh` cookie, and exponential backoff in the SSE hooks — none of which exist in the current source. These will be deleted in PR 6.

Files to delete: `docs/dx-auth-design.md`, `docs/energy-bean-rule.md`, `docs/game-word-sentence-group-rule.md`, `docs/game-word-sentence-single-rule.md`, `docs/migration-plan.md`.

### B — Zero frontend error resilience on SSE disconnects
Current hooks have no `onerror`, no retry logic, no auth failure handling. A 401 from the backend on reconnect is silently retried forever by the browser's native EventSource loop with no user feedback. **Fix in this refactor:** the new `WebSocketProvider` handles close codes explicitly with redirect-to-signin for auth failures and bounded reconnect with user-visible toasts.

### C — Mid-session single-device kick is not enforced on open streams
`jwt_auth.go:47-64` enforces single-device via a Redis login-timestamp check, but only on HTTP requests. Already-open SSE/WS streams are unaware of the kick and keep running until they die for some other reason. **Fix in this refactor:** when the login flow updates `user_auth:{userID}:user` in Redis, also publish a kick event via `realtime.Publish(UserKickTopic(userID), ...)`. The hub listens and force-closes matching connections with WS close code 4001.

### D — No explicit reconnect backoff strategy today
Browser's native EventSource retry is a fixed 3-second interval. On a flaky network, this hammers the server. **Fix in this refactor:** the new hook uses exponential backoff 1s→2s→4s→8s→16s→30s (capped) with jitter, reset on successful connect, with a 10-attempt cap that surfaces "connection lost" to the user.

### E — Bundled infrastructure blockers (must fix for WS to work at all)

- **`nginx.prod.conf` missing `proxy_http_version 1.1` + `Upgrade`/`Connection` headers on `/api/*`.** WebSocket upgrade will 400 in prod without this.
- **Two Redis clients in process.** `app/helpers/redis.go:20-34` creates a separate `*redis.Client` instead of using `redis_facades.Instance("default")`. Consolidate to one client so config drift (TLS, cluster mode) can't diverge. ~15 LOC.
- **`request_timeout: 30` in `config/http.go:38` WILL kill WebSocket connections.** Verified against `goravel/gin@v1.17.0/middleware_timeout.go` — the `Timeout(timeout)` middleware is installed as the first global middleware in `goravel/gin@v1.17.0/route.go:46` (`globalMiddleware := []contractshttp.Middleware{Timeout(timeout), Cors(), Tls()}`) and wraps every request in `context.WithTimeout(ctx.Context(), timeout)`. After 30s, the wrapped context is cancelled and the request is aborted with HTTP 408. Since the WebSocket's read/write loops use that context, they'd fail after 30s.

  **Resolution (baked into the design, not deferred):** the `WSController.Handle` uses a **detached context** (derived from `context.Background()`, not from `ctx.Request().Context()`) when calling `Hub.Attach`. The Gin timeout middleware still fires after 30s, but:
  - The `Abort(408)` call is a no-op on the wire because `websocket.Accept` has already hijacked the underlying `net.Conn` (standard Go net/http `Hijacker` contract — once hijacked, the HTTP response writer is detached and status writes are discarded).
  - The middleware's `defer cancel()` fires on the request-scoped context, but we never passed that context into the WS read/write loops, so the cancellation is also a no-op.
  - Shutdown of the WS is driven by a separate cancellation signal managed by `Hub.Shutdown` (called from Goravel's `app.Terminating` hook), not by the request context.

  This means **no change to `request_timeout` is needed** — the 30s setting stays in place and continues to protect normal HTTP endpoints from slow-loris-style attacks. Only the WS handler is special-cased via its context choice, entirely inside our own code. No need to fork or modify Goravel's middleware.

---

## 4. Architecture

### High-level picture

```
┌─────────────────────────────────────────────────────────────┐
│  Browser (dx-web)                                           │
│                                                             │
│   hall/layout.tsx                                           │
│      │                                                      │
│      ▼                                                      │
│   WebSocketProvider  ─────┐                                 │
│   (one per hall mount)    │                                 │
│                           │   One WS connection             │
│   useTopic("user:{id}")   │   multiplexed with              │
│   useTopic("pk:{id}")     │   {op, topic, data} envelopes   │
│   useTopic("group:{id}")  │                                 │
│   useTopic("group:{id}:notify") ──── coder/websocket ──┐    │
└────────────────────────────────────────────────────────│────┘
                                                         │
                                              (GET /api/ws upgrade)
                                                         │
┌────────────────────────────────────────────────────────▼────┐
│  dx-api (Goravel + Gin + coder/websocket)                   │
│                                                             │
│  routes/api.go                                              │
│    GET /api/ws ──► controllers/api/ws_controller.go         │
│                     │                                       │
│                     ├─ JWT cookie auth (JwtAuth middleware) │
│                     ├─ Upgrade with coder/websocket.Accept  │
│                     └─ realtime.Default_Hub.Attach(...)     │
│                                                             │
│  app/realtime/                                              │
│  ├─ hub.go         local client registry + topic router    │
│  ├─ client.go      one per WS connection (read/write loop) │
│  ├─ pubsub.go      PubSub interface (seam)                  │
│  ├─ pubsub_redis.go  RedisPubSub implementation             │
│  ├─ presence.go    SADD/SREM/SMEMBERS for topic presence    │
│  ├─ envelope.go    wire protocol types                      │
│  ├─ topic.go       topic naming helpers                     │
│  ├─ authorize.go   per-topic subscribe authorization        │
│  └─ errors.go      sentinel errors                          │
│                                                             │
│  app/services/api/*.go  ◄─── ~30 call sites                 │
│    realtime.Publish(ctx, topic, Event{Type: ..., Data: ...})│
│                                                             │
└───────────────┬─────────────────────────────────────────────┘
                │
                ▼
         ┌──────────────┐
         │    Redis     │   pub/sub channels:
         │  (existing)  │     user:{userID}
         │              │     user:{userID}:kick
         │              │     pk:{pkID}
         │              │     group:{groupID}
         │              │     group:{groupID}:notify
         │              │   presence sets:
         │              │     presence:user:{userID}
         │              │     presence:pk:{pkID}
         │              │     presence:group:{groupID}
         │              │     ...
         └──────────────┘
```

### One-sentence flow
Service code calls `realtime.Publish(ctx, topic, event)` → event lands on a Redis pub/sub channel → every dx-api instance subscribed to that channel picks it up → each instance's local Hub fans it out to every Client subscribed to that topic → the Client writes a JSON envelope frame to its WebSocket → the browser's `useTopic` hook delivers it to the registered handler.

### Component responsibilities

**Hub** — owns the local client registry and topic→client routing for one dx-api process. Knows which `*Client`s are connected and which topics each has subscribed to. Does not know about Redis or WebSocket framing. Pure mediator between clients and pub/sub.

**Client** — owns one WebSocket connection. Two goroutines per client: read loop (handles inbound subscribe/unsubscribe ops) and write loop (drains a buffered send channel, periodic server ping via `coder/websocket`'s `c.Ping`). Concurrent-safe because all writes go through the channel.

**PubSub interface** — the seam between the Hub and the backing transport. Two implementations: `RedisPubSub` (production) and a fake declared in test files. Business code only calls `realtime.Publish(ctx, topic, event)`, which delegates to `Default.Publish(...)`.

**RedisPubSub** — owns one long-lived Redis subscribe connection per dx-api instance. One loop goroutine reads messages from Redis and fans them out to local subscribers. Ref-counts topics: first local subscriber triggers `SUBSCRIBE`; last unsubscribe triggers `UNSUBSCRIBE`. Reconnect on Redis drop is handled internally by `go-redis/v9`.

**Service code** — exactly one new dependency: `realtime.Publish`. Replaces all `Hub.Broadcast/SendToUser/Notify` calls. During the migration window (PRs 1–5), service code calls **both** the legacy hub AND `realtime.Publish` — the "double-publish" pattern that makes MIG2 safe.

### Frontend shape
One primitive (`WebSocketProvider`), one helper (`useTopic`), four migrated wrapper hooks with identical public APIs to the current SSE hooks. The existing per-feature hooks (`usePkPlayEvents`, `useGroupPlayEvents`, `useGroupEvents`, `PkInvitationProvider`) switch their inner import from `useXxxSSE` to `useXxxEvents` — components using them don't change.

### What the architecture does NOT touch
- PostgreSQL schemas, GORM models, migrations (realtime state is Redis + in-memory only)
- PG partitioning + code-level FK constraint strategy (realtime layer deals in opaque ID strings)
- Game play business logic (only the "push event to browser" lines change)
- REST endpoints (all POST/GET stays HTTP)
- ai-custom endpoints (HTTP streaming, framing switched to NDJSON)
- JWT auth system (cookie auth reused as-is at the WS handshake)

---

## 5. Wire Protocol

### Envelope

```go
type Envelope struct {
    Op      Op     `json:"op"`
    Topic   string `json:"topic,omitempty"`
    Type    string `json:"type,omitempty"`
    Data    any    `json:"data,omitempty"`
    ID      string `json:"id,omitempty"`       // correlation id for subscribe/unsubscribe
    OK      *bool  `json:"ok,omitempty"`       // on ack
    Code    int    `json:"code,omitempty"`     // on ack/error, aligned with consts.Code*
    Message string `json:"message,omitempty"`  // on ack/error
}
```

### Op types

| Op | Direction | Purpose |
|---|---|---|
| `subscribe` | c → s | Subscribe to a topic |
| `unsubscribe` | c → s | Remove a subscription |
| `event` | s → c | Pushed event on a subscribed topic |
| `ack` | s → c | Acknowledge a subscribe/unsubscribe with success or failure |
| `error` | s → c | Out-of-band / protocol-level error |

The client never sends game actions over the WebSocket. Game actions stay REST.

### Topic families

| Topic pattern | Scope | Authorization rule |
|---|---|---|
| `user:{userID}` | One user globally | Only if `userID == currentUser.id` |
| `user:{userID}:kick` | One user (kick channel) | Same as above; server publishes to this on single-device replace |
| `pk:{pkID}` | One PK match | Only if user is a participant (DB check) |
| `group:{groupID}` | One group room | Only if user is a member (DB check) |
| `group:{groupID}:notify` | Group metadata | Only if user is a member (DB check) |

### Event type catalog (1:1 with current SSE)

**`user:{userID}`**: `pk_invitation`, `pk_next_level`
**`pk:{pkID}`**: `pk_invitation_accepted`, `pk_invitation_declined`, `pk_player_complete`, `pk_player_action`, `pk_force_end`
**`group:{groupID}`**: `group_game_start`, `group_player_complete`, `group_player_action`, `group_game_force_end`, `group_next_level`, `group_dismissed`, `room_member_joined`, `room_member_left`
**`group:{groupID}:notify`**: `group_updated` (with `data.scope` in `detail|members|applications|subgroups`)

17 event types total, all identical to the current SSE event names so the frontend migration is a transport swap, not a schema change.

### Subscribe semantics

- **First local subscribe on an instance** triggers a Redis `SUBSCRIBE` for that topic (via `RedisPubSub`).
- **Last local unsubscribe on an instance** triggers `UNSUBSCRIBE`.
- **Idempotent**: re-subscribing to a topic you're already on returns a successful ack.
- **Authorization enforced at subscribe time**, not publish time. Publish-side code never checks topic authorization — only subscribed (authorized) clients can receive events.
- **Presence tracked in Redis** via `SADD presence:{topic} {userID}` on subscribe and `SREM` on unsubscribe, with sliding `EXPIRE presence:{topic} 900` to self-heal stale entries from crashed instances.

### Heartbeat

Server-initiated pings via `coder/websocket`'s built-in `c.Ping(ctx)`, every 25 seconds (configurable), with a 5-second context timeout. If ping fails, the write loop closes the connection. Client pongs are handled automatically by the browser's WebSocket implementation. **No application-level ping/pong ops in the envelope protocol** — pings are at the WebSocket protocol layer below the JSON envelope.

25s is chosen to stay well under nginx's `proxy_read_timeout` (which we'll set to 1h+ for `/api/ws`) and under intermediate load balancer timeouts.

### Upgrade handshake

- **Auth**: `JwtAuth` middleware runs on the HTTP upgrade request. Valid token → upgrade proceeds; expired token within 14-day refresh window → middleware transparently refreshes and sets fresh cookie on the 101 response; expired beyond refresh window or missing cookie → 401 before upgrade.
- **Origin check**: `websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: ...})` populated from `CORS_ALLOWED_ORIGINS` env var. Cross-origin upgrades are rejected at the protocol level.
- **Compression**: disabled (`CompressionDisabled`). Marginal payload sizes + historical CVE class.
- **Read limit**: `c.SetReadLimit(4096)` — client frames are tiny subscribe/unsubscribe ops, 4 KB is ample and caps memory exposure to a misbehaving client.

### Token expiry during a long-lived connection

Auth is checked at upgrade time only. Once upgraded, the connection runs under the user context captured at upgrade until it dies — matches current SSE semantics exactly. If the user's access token expires mid-session, the socket stays alive; only on next reconnect does the middleware run again and (within the 14-day refresh window) auto-refresh transparently via `facades.Auth.Guard("user").Refresh()` at `jwt_auth.go:26-33`.

**No `refreshAccessToken()` function exists in the frontend.** No `dx_refresh` cookie exists. The prior design docs describing these are stale and will be deleted in PR 6. Token refresh is entirely server-side, driven by the middleware.

### Things deliberately NOT in the protocol

- Sequence numbers (`seq`). No per-topic ordering guarantees; clients re-query state via REST on reconnect.
- Request-response RPC over the socket. HTTP handles all requests/responses; WebSocket is push-only.
- Binary frames. JSON text frames only.

---

## 6. Backend Package Layout

### Directory tree

```
dx-api/app/realtime/
├── envelope.go               wire protocol types
├── topic.go                  topic naming + parsing helpers
├── authorize.go              per-topic subscribe authorization
├── pubsub.go                 PubSub interface + package-level Default + Publish()
├── pubsub_redis.go           RedisPubSub implementation
├── hub.go                    Hub struct: client registry + topic router + presence
├── client.go                 Client struct: read/write loops, one per WS
├── presence.go               Redis SET helpers: SADD/SREM/SMEMBERS/EXPIRE
├── errors.go                 sentinel errors
│
├── envelope_test.go
├── topic_test.go
├── authorize_test.go
├── pubsub_redis_test.go      miniredis
├── hub_test.go               fake PubSub in test file
├── client_test.go            httptest + coder/websocket
├── presence_test.go          miniredis
└── realtime_integration_test.go   full roundtrip + cross-instance simulation

dx-api/app/http/controllers/api/
└── ws_controller.go          /api/ws upgrade handler + test

dx-api/bootstrap/app.go       wire realtime.Default + Default_Hub + Terminating hook
```

### Key exported symbols

**`realtime.Publish(ctx, topic, event) error`** — the single function every service call site uses.
**`realtime.Default PubSub`** — package-level, set at bootstrap. `Publish` delegates here.
**`realtime.Default_Hub *Hub`** — package-level, set at bootstrap. The `WSController` calls `Attach` on this.
**`realtime.NewRedisPubSub(ctx, client) *RedisPubSub`** — constructor; starts the subscribe loop goroutine.
**`realtime.NewHub(pubsub) *Hub`** — constructor; no side effects until `Attach` is called.
**`realtime.UserTopic(userID) string`** — topic helpers, one per family.
**`realtime.AuthorizeSubscribe(ctx, userID, topic) error`** — returns `realtimeError` with `Code` and `Message`.
**`realtime.IsPresent(ctx, topic, userID) (bool, error)`** — cross-instance topic presence check.
**`realtime.PresentOnTopic(ctx, topic) ([]string, error)`** — returns all userIDs present on a topic (replaces in-memory hub walks for winner determination).
**`Hub.Attach(ctx, userID, conn) error`** — called by `WSController.Handle` after successful upgrade. Blocks until the connection dies. **Must be called with a detached context, not `ctx.Request().Context()`** — see the controller sketch below.
**`Hub.Shutdown(ctx) error`** — graceful shutdown entry point, registered as Goravel terminating hook.

### WebSocket controller sketch (with the detached-context fix for Issue E)

```go
// dx-api/app/http/controllers/api/ws_controller.go
package api

import (
    "context"
    "net/http"
    "strings"

    contractshttp "github.com/goravel/framework/contracts/http"
    "github.com/coder/websocket"
    "github.com/goravel/framework/facades"

    "dx-api/app/consts"
    "dx-api/app/helpers"
    "dx-api/app/realtime"
)

type WSController struct{}

func NewWSController() *WSController { return &WSController{} }

func (c *WSController) Handle(ctx contractshttp.Context) contractshttp.Response {
    userID, err := facades.Auth(ctx).Guard("user").ID()
    if err != nil || userID == "" {
        return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
    }

    origins := facades.Config().GetString("cors.allowed_origins", "")
    w := ctx.Response().Writer()
    r := ctx.Request().Origin() // *http.Request

    conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
        OriginPatterns:  strings.Split(origins, ","),
        CompressionMode: websocket.CompressionDisabled,
    })
    if err != nil {
        // websocket.Accept already wrote the error response
        return nil
    }
    defer conn.Close(websocket.StatusInternalError, "server error")

    // CRITICAL: use a detached context, NOT ctx.Request().Context().
    // Goravel's global Timeout middleware (goravel/gin/middleware_timeout.go)
    // cancels the request context after http.request_timeout (30s by default),
    // which would kill the WebSocket. A background-derived context is immune;
    // shutdown is driven separately via Hub.Shutdown.
    wsCtx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Hub.Attach blocks until the connection dies (disconnect, error, shutdown).
    // The hub is responsible for close-code handling on its side.
    _ = realtime.Default_Hub.Attach(wsCtx, userID, conn)
    return nil
}
```

Key points:
- The HTTP upgrade response (the `101 Switching Protocols`) is written by `websocket.Accept`, which hijacks the underlying `net.Conn`. After that point, the Gin `ResponseWriter` is detached.
- When the Gin timeout middleware fires its `Abort(408)` after 30s, it's a no-op on the hijacked connection — status writes go to /dev/null.
- The middleware's `defer cancel()` on the timeout context is also a no-op because we never passed that context into the hub.
- The WS stays alive as long as the user stays connected and the process is running. Shutdown is handled by `Hub.Shutdown` via Goravel's terminating hook, which iterates and closes all clients cleanly with WS close code 4002.

### Double-publish during migration

During PRs 1–5, every existing service call site gets **two** publish lines side by side — the legacy hub call (old path) and `realtime.Publish` (new path):

```go
// Example: game_play_group_service.go
func StartGroupGame(ctx, groupID, ...) error {
    // ... business logic ...
    
    // Legacy path — delivers to SSE subscribers still on the old hooks
    helpers.GroupSSEHub.Broadcast(groupID, "group_game_start", startEvent)
    
    // New path — delivers to WebSocket subscribers on the migrated hooks
    _ = realtime.Publish(ctx, realtime.GroupTopic(groupID), realtime.Event{
        Type: "group_game_start",
        Data: startEvent,
    })
    
    return nil
}
```

Both lines live in the code until PR 6 (cleanup), which removes the legacy call. Easy to grep (`helpers.GroupSSEHub.Broadcast`) and delete.

### Call sites to migrate

| Service file | Legacy calls | Count |
|---|---|---|
| `app/services/api/pk_invite_service.go` | `UserHub.SendToUser`, `PkHub.Broadcast` × 2 | 3 |
| `app/services/api/game_play_pk_service.go` | `PkHub.Broadcast` × 4, `UserHub.SendToUser` × 2 | 6 |
| `app/services/api/game_play_group_service.go` | `GroupSSEHub.Broadcast` × 2, `GroupNotifyHub.Notify` × 1 + fire-and-forget goroutine | 4 |
| `app/services/api/group_game_service.go` | `GroupSSEHub.Broadcast` × 3, `GroupNotifyHub.Notify` × 3 | 6 |
| `app/services/api/group_service.go` | `GroupSSEHub.Broadcast` × 1 | 1 |
| `app/services/api/group_member_service.go` | `GroupNotifyHub.Notify` × 2 | 2 |
| `app/services/api/group_application_service.go` | `GroupNotifyHub.Notify` × 4 | 4 |
| `app/services/api/group_subgroup_service.go` | `GroupNotifyHub.Notify` × 5 | 5 |

**Total**: ~31 call sites, mechanical edits. Each adds one `realtime.Publish` line next to the legacy call during PR 1; each removes the legacy line during PR 6.

### Double-publish rule for invitation-style events
Events that must reach a user regardless of their current page — today's `UserHub.SendToUser` callers — publish to **two** topics: the context topic (e.g., `pk:{pkID}`) AND the recipient's user topic (e.g., `user:{bobID}`). This preserves the existing "PK accept guarantees delivery even if recipient hasn't opened the pk-room page yet" behavior. Concretely, `AcceptPkInvite`, `DeclinePkInvite`, `InvitePk`, and `NextPkLevel (specified PK)` all double-publish. Enumerated in the implementation plan.

---

## 7. Frontend Design

### New files

```
dx-web/src/providers/websocket-provider.tsx    WebSocketProvider + WSContext + useWS()
dx-web/src/hooks/use-websocket.ts                re-exports useWS for consistency
dx-web/src/hooks/use-topic.ts                    useTopic(topic, handlers) primitive
dx-web/src/hooks/use-user-events.ts              wraps useTopic("user:{id}", ...)
dx-web/src/hooks/use-pk-events.ts                wraps useTopic("pk:{id}", ...)
dx-web/src/hooks/use-group-events.ts             wraps useTopic("group:{id}", ...)
dx-web/src/hooks/use-group-notify.ts             replaces the old file, wraps useTopic("group:{id}:notify", ...)
```

### Provider responsibilities (`websocket-provider.tsx`)

- Opens one WebSocket connection on mount at `${NEXT_PUBLIC_API_URL}/api/ws`.
- Exposes `subscribe(topic, handler) -> unsubscribe()` via React context.
- Tracks subscriptions in a ref-counted map (multiple `useTopic` calls with the same topic share one wire subscription).
- On first handler for a topic: sends `{op: "subscribe", topic, id}` over the wire.
- On last handler removed: sends `{op: "unsubscribe", topic}`.
- Handles close codes explicitly:
  - **4001** (session_replaced): toast + redirect to `/auth/signin?reason=session_replaced`, no reconnect
  - **4401** (session_expired): redirect to `/auth/signin?reason=session_expired`, no reconnect
  - **1000** (normal close): no action (provider unmounting)
  - **other**: schedule reconnect with exponential backoff + jitter
- Reconnect backoff: 1s → 2s → 4s → 8s → 16s → 30s (capped), with `jitter = random(0, base * 0.3)`, reset on `onopen`, cap at 10 attempts (surface "connection lost, please refresh" toast and stop).
- On successful reconnect (`onopen`), re-sends subscribe frames for every topic in the active subscription map — so mid-connection subscriptions survive reconnects transparently.
- Closes WebSocket with code 1000 on provider unmount.

### Helper (`use-topic.ts`)

```typescript
export function useTopic(
  topic: string | null,
  handlers: Record<string, (data: unknown) => void>,
): void {
  const { subscribe } = useWS();
  const handlersRef = useRef(handlers);
  useEffect(() => { handlersRef.current = handlers; });

  useEffect(() => {
    if (!topic) return;
    const unsubscribe = subscribe(topic, (event) => {
      handlersRef.current[event.type]?.(event.data);
    });
    return unsubscribe;
  }, [topic, subscribe]);
}
```

Ref-based handler update pattern (same as current SSE hooks) so handlers can change without re-subscribing.

### Wrapper hooks (same public API as today's SSE hooks)

Each is ~15 lines. Examples:

```typescript
// use-user-events.ts
export function useUserEvents(listeners: Record<string, (data: unknown) => void>): void {
  const userId = useCurrentUser()?.id;
  useTopic(userId ? `user:${userId}` : null, listeners);
}

// use-pk-events.ts
export function usePkEvents(
  pkId: string | null,
  listeners: Record<string, (data: unknown) => void>,
): void {
  useTopic(pkId ? `pk:${pkId}` : null, listeners);
}
```

### Name collision handling

The feature-level hook `useGroupEvents` at `features/web/groups/hooks/use-group-events.ts` shares a name with the new primitive-level `useGroupEvents` at `src/hooks/`. The feature hook is renamed to `useGroupRoomEvents` (or similar) during PR 4 to avoid shadowing. One-time cost, contained in PR 4.

### Hall layout change

```tsx
// dx-web/src/app/(web)/hall/layout.tsx (PR 2 change)
import { WebSocketProvider } from "@/providers/websocket-provider";
import { PkInvitationProvider } from "@/features/web/play-pk/components/pk-invitation-provider";

export default function HallLayout({ children }) {
  return (
    <WebSocketProvider>
      <PkInvitationProvider>
        {children}
      </PkInvitationProvider>
    </WebSocketProvider>
  );
}
```

`PkInvitationProvider` sits **inside** `WebSocketProvider` because it consumes the context.

### The `POST /api/user/ping` call goes away

`use-user-sse.ts:15` today calls `POST /api/user/ping` on mount to mark the user online early. In the new design, the WS upgrade itself is the online signal — the hub calls `RedisSetAdd("online_users", userID)` on `Attach` and `SRem` on `detach`. No separate ping endpoint needed. Removed in PR 2 (when `use-user-sse.ts` is deleted). The backend `/api/user/ping` route is deleted in PR 6.

---

## 8. Data Flow Scenarios

### Scenario 1 — PK invitation → accept → play (two-machine cross-instance)

Alice is on Machine 1, Bob is on Machine 2. Both have `WebSocketProvider` mounted and their respective `user:{id}` topics subscribed. Machine 1 is subscribed to Redis channel `user:alice` and Machine 2 to `user:bob`.

1. **Alice clicks "Invite Bob"** → `POST /api/pk/invite` → Machine 1 `pk_invite_service.InvitePk`.
2. Service does DB work (creates pk row), then publishes to **both** `pk:{pkID}` and `user:{bobID}` (double-publish rule).
3. Redis fans out `user:bob` to Machine 2 → Machine 2's hub → Bob's Client → WebSocket frame → `PkInvitationProvider` state update → modal appears.
4. **Bob clicks Accept** → `POST /api/pk/{pkID}/accept` → routed to any instance (say Machine 2) → service publishes `pk_invitation_accepted` to **both** `pk:{pkID}` and `user:{aliceID}`.
5. Alice's handler on her user topic navigates her to `/hall/play-pk/{pkID}`. The pk-play page mounts, its `usePkEvents(pkID)` subscribes to `pk:{pkID}` → Machine 1 sends `SUBSCRIBE pk:{pkID}` to Redis.
6. Bob's frontend navigates similarly → Machine 2 also subscribes.
7. **Gameplay**: each `POST /api/play-pk/{pkID}/answer` publishes `pk_player_action` on `pk:{pkID}`. Redis fans out to both Machine 1 and Machine 2. Both players see opponent action animations.
8. **`CompletePk`** publishes `pk_player_complete` → both see result panel.
9. **`NextPkLevel` presence check**: service calls `realtime.IsPresent(ctx, PkTopic(pkID), otherUserID)` — uses the Redis `SISMEMBER presence:pk:{pkID} {otherUserID}` check, cross-instance-consistent. If opponent still present, stays in specified PK; otherwise falls back to robot.

### Scenario 2 — Group game

Members A, B, C on `group:abc`, all subscribed to `group:abc` topic.

1. **Owner starts game** → `StartGroupGame` publishes `group_game_start` on `group:abc` and `group_updated{scope:"detail"}` on `group:abc:notify`.
2. All members receive `group_game_start`, navigate to play page, `useGroupPlayEvents` mounts, subscribes to `group:abc` (reuses existing subscription, just another local handler).
3. **Player answers** → `POST /api/play-group/{id}/answer` → publishes `group_player_action` on `group:abc`.
4. **Level complete** → `GroupPlayCompleteLevel` publishes `group_player_complete` on `group:abc` + `group_updated{scope:"detail"}` on notify.
5. **Winner determination on mid-game leave**: service calls `realtime.PresentOnTopic(ctx, GroupTopic(groupID))` to get the authoritative list of still-connected members across all instances, replacing the in-memory `GroupSSEHub.ConnectedUserIDs` walk.
6. **Member leaves room** → WS closes (or `useGroupEvents` unmounts) → hub calls `unsubscribe`, which (for `KindGroup` topics) publishes `room_member_left` to `group:abc`. Other members receive the event and refetch member avatars.

### Scenario 3 — Group metadata update

1. **Owner kicks a member** → `KickMember` publishes `group_updated{scope:"members"}` and `group_updated{scope:"detail"}` to `group:{id}:notify`.
2. All members on the group detail page (subscribed to the notify topic) receive the events → SWR refetches the relevant cache scopes.

### Scenarios unchanged

**Single game play**: zero WS involvement. REST only. Unchanged.
**AI-custom**: 4 POST streaming endpoints stay HTTP with NDJSON framing (cleanup in PR 6). Unchanged in user-visible behavior.

### Design elements added during scenario walk-through

1. **Double-publish rule for invitation-style events** (preserves current guaranteed-delivery for PK invites, accept, decline, specified next-level).
2. **Topic presence tracking via Redis SET** (`presence:{topic}`) — cross-instance `IsPresent` and `PresentOnTopic` helpers; sliding `EXPIRE` for self-healing.
3. **Group auto-events** (`room_member_joined` / `room_member_left`) emitted by the Hub when a client subscribes/unsubscribes to a `KindGroup` topic. Preserves current SSE behavior through the new transport.

All three additions live in `hub.go` + `presence.go` (~50 LOC total).

---

## 9. Error Handling, Reconnect, and Graceful Shutdown

### WebSocket close code catalog

| Code | Kind | Sender | When | Client action |
|---|---|---|---|---|
| 1000 | Normal | Either side | Clean shutdown | No action |
| 1001 | Going away | Server | Close without explicit shutdown ceremony | Reconnect w/ backoff |
| 1006 | Abnormal | (No close frame) | Network drop or upgrade 401 | Reconnect if prior onopen; else → signin |
| 1008 | Policy violation | Server | Client protocol violation | Reconnect + log (bug) |
| 1009 | Message too big | Server | Client exceeded 4 KB read limit | Reconnect + log |
| 1011 | Internal error | Server | Hub panic or unexpected bug | Reconnect w/ backoff |
| 4001 | session_replaced | Server | Single-device kick (Issue C) | **Redirect to signin**, no reconnect |
| 4002 | server_shutdown | Server | SIGTERM graceful shutdown | Reconnect (lands on surviving instance) |
| 4003 | slow_consumer | Server | Client send queue overflow | Reconnect + log frequency |
| 4401 | session_expired | Server | Auth beyond refresh window (rare — usually 401 at HTTP) | Redirect to signin |

Application error codes inside `ack`/`error` envelopes are a separate layer, aligned with `consts.Code*` integer constants (see Appendix A).

### Server-side error handling per component

**RedisPubSub — subscribe loop resilience**
- `go-redis/v9` handles TCP reconnect internally, auto-resubscribes channels after recovery.
- Messages published during a Redis outage window are lost (fire-and-forget semantics). Clients recover state via REST re-query after reconnect.
- Publish side also fails during outage; service code logs and continues. Business state is durable in Postgres; only the realtime push is lost.

**Publish error handling at call sites**
```go
if err := realtime.Publish(ctx, topic, event); err != nil {
    facades.Log().WithContext(ctx).Warnf("realtime publish failed: topic=%s type=%s err=%v", topic, event.Type, err)
}
```
For high-frequency events (`pk_player_action`), use `_ = realtime.Publish(...)` with no log to avoid noise.

**Slow consumer handling (backpressure)**
- `Client.send` is a buffered channel, capacity 32.
- `Client.enqueue` is non-blocking; if full, closes the client with WS close code 4003 and detaches.
- Client reconnects, re-subscribes, re-queries state via REST.
- Industry-standard pattern — better to force a reconnect than accumulate unbounded memory per slow client.

**Panic isolation**
- `Client.Serve` wraps read/write loops in `defer recover` + log + detach.
- Hub fanout goroutines same treatment.
- One panic does not take down the hub.

**Bootstrap order**
```go
// bootstrap/app.go
redisClient, _ := redis_facades.Instance("default")
ctx := context.Background()
realtime.Default = realtime.NewRedisPubSub(ctx, redisClient)
realtime.Default_Hub = realtime.NewHub(realtime.Default)
// ... HTTP server + routes ...
app.Terminating(func() error {
    return realtime.Default_Hub.Shutdown(context.Background())
})
```
Bootstrap test asserts `realtime.Default != nil` after init. Calls to `realtime.Publish` before init return `ErrNotInitialized` as a defensive nil-guard.

### Graceful shutdown sequence

1. Stop accepting new WS upgrades (`/api/ws` handler checks a `shuttingDown` flag, returns 503 if set).
2. Send close frame code 4002 to all attached clients.
3. Wait up to 5 seconds for clean drain.
4. Force-close stragglers via `conn.CloseNow()`.
5. Close RedisPubSub loop (cancel context, `pubsub.Close()`).
6. Process exits.

In multi-instance deployments: rolling deploys work cleanly because disconnected clients reconnect via backoff and land on a healthy instance via the load balancer.

### Client-side dispatch (in `WebSocketProvider`)

Already covered in Section 7. Key points:
- Close code switch with explicit handling for 4001, 4401, 1000, default (backoff reconnect)
- 10-attempt reconnect cap with user-visible "connection lost" toast
- Replays all active subscriptions on successful reconnect
- Event loss during the reconnect window is acceptable; clients reconcile via REST re-query

### What we don't handle

- Message de-duplication (handlers must be idempotent — set state, don't increment)
- Persistent event log for replay (out of scope)
- Exactly-once delivery (out of scope)
- Cross-topic event ordering (each topic is ordered internally per RedisPubSub semantics; cross-topic has no guarantee)

---

## 10. Testing Strategy

### Philosophy
- TDD for hub logic, authorization, pub/sub roundtrip
- Integration tests for cross-boundary components (`miniredis` for RedisPubSub, `httptest` for WS controller)
- Evidence before claims — no PR marked ready without observed output from its verification checklist

### Unit tests (backend)

| File | What it covers |
|---|---|
| `envelope_test.go` | Marshal/unmarshal roundtrip per op type, omitempty correctness, required field validation, `OK *bool` edge cases |
| `topic_test.go` | Helper roundtrip, `ParseTopic` rejects malformed topics, `group:x:notify` resolves to `KindGroupNotify` not `KindGroup` |
| `authorize_test.go` | Authorization matrix (table-driven): user self/other, pk participant/non-participant, group member/non-member, invalid topic |
| `pubsub_redis_test.go` | Publish/Subscribe roundtrip via miniredis, ref-counting, reconnect after Redis kill/restart, context cancellation, Close terminates goroutine |
| `hub_test.go` | Attach/detach, topic ref-counting, slow consumer kick (fill channel → assert close 4003), concurrent subscribe lock correctness, auto-events on group subscribe/unsubscribe, panic recovery in fanout goroutines |
| `client_test.go` | Read loop subscribe/unsubscribe dispatch, unknown op → error envelope, write loop drain, ping ticker via time mock, read limit 1009, graceful close |
| `presence_test.go` | SADD/SREM/SMEMBERS correctness, sliding EXPIRE behavior, IsPresent / PresentOnTopic helpers |

### Integration tests

- **`realtime_integration_test.go` — full roundtrip**: httptest server + coder/ws upgrade + RedisPubSub (miniredis) → publish → assert event delivery via WebSocket client.
- **Cross-instance simulation**: two Hub/PubSub pairs sharing one miniredis; assert event published on instance A reaches a client connected to instance B.

### Bootstrap tests

- `realtime.Default` initialized after bootstrap
- Shutdown hook registered and invocable
- Calls to `realtime.Publish` before bootstrap return `ErrNotInitialized`

### Frontend tests

*To be aligned with the project's existing Jest/Vitest setup during PR 2 implementation.*

- `WebSocketProvider` mount/unmount, reconnect backoff schedule via fake timers, close code dispatch, slow consumer kick end-to-end
- `useTopic` ref-update without re-subscription, shared subscription for multiple handlers
- Wrapper hook smoke tests (verify correct topic string and handler pass-through)

### Manual verification per MIG2 PR

Each PR has a specific checklist (see Section 11). Key commands:

```bash
# Run from deploy/ for docker-compose manual verification
docker compose -f deploy/docker-compose.dev.yml up -d
# Then exercise flows in the browser

# Before any PR merge
cd dx-api && go test -race ./... && go build ./... && golangci-lint run ./...
cd dx-web && npm run lint && npm run build
```

### Explicitly out of scope

- Load testing (separate future initiative)
- Chaos engineering
- Playwright/Cypress E2E (manual checklists substitute)
- Performance regression benchmarks

---

## 11. Migration Plan (6 PRs, MIG2 Parallel Shadow)

### Sequencing

```
PR 1 — Backend + provider (dormant)          ← NO user-visible change
PR 2 — Mount provider + useUserEvents         ← first user-visible change
PR 3 — usePkEvents migration
PR 4 — useGroupEvents migration (play + room)
PR 5 — useGroupNotify migration
PR 6 — Cleanup: delete SSE, remove double-publish, ai-custom NDJSON, stale docs
```

Each PR is independently deployable and revertible. During PRs 1–5, service code double-publishes. PR 6 removes double-publish atomically.

### PR 1 — Backend realtime layer + dormant frontend provider

**Goal**: ship backend infrastructure and frontend provider file (not yet mounted). Zero user-visible change.

**Backend added**:
- `app/realtime/*` package (all files from Section 6)
- `app/http/controllers/api/ws_controller.go`
- `app/realtime/realtime_integration_test.go`

**Backend modified**:
- `app/consts/error_code.go` — 4 new constants (Appendix A)
- `app/helpers/redis.go` — consolidate to `redis_facades.Instance("default")`, fixing the dual-client issue
- `bootstrap/app.go` — wire `realtime.Default`, `Default_Hub`, terminating hook
- `routes/api.go` — register `protected.Get("/ws", wsController.Handle)`
- `config/http.go` — probe `request_timeout`; if it kills WS upgrades, exempt `/api/ws` or raise the timeout
- Service call sites (~31 lines added for double-publish across 8 files, enumerated in Section 6)
- Login service wherever it updates `user_auth:{userID}:user` in Redis — add `realtime.Publish(ctx, UserKickTopic(userID), Event{Type:"session_replaced"})` for Issue C

**Infrastructure modified**:
- `deploy/nginx/nginx.prod.conf` — add `proxy_http_version 1.1`, `Upgrade`, `Connection` headers for `/api/*`; set `proxy_read_timeout` and `proxy_send_timeout` to 1h for WS
- Verify `deploy/docker-compose.prod.yml` doesn't add its own timeout middleware that would conflict

**Frontend added (dormant — not mounted anywhere)**:
- `src/providers/websocket-provider.tsx`
- `src/hooks/use-websocket.ts`
- `src/hooks/use-topic.ts`
- `src/hooks/use-user-events.ts`
- `src/hooks/use-pk-events.ts`
- `src/hooks/use-group-events.ts`
- `src/hooks/use-group-notify-ws.ts`

**Dependencies**:
- `dx-api/go.mod`: add `github.com/coder/websocket`, `github.com/alicebob/miniredis/v2` (test)

**Tests added**: all unit + integration tests from Section 10.

**Manual verification** (run via docker-compose):
```bash
docker compose -f deploy/docker-compose.dev.yml up -d
```
- [ ] `go test -race ./... && go build ./... && golangci-lint run ./...`
- [ ] `npm run lint && npm run build`
- [ ] Health check `/api/health` returns 200 with db=true, redis=true
- [ ] Existing SSE flows work end-to-end (nothing visibly changed)
- [ ] Manual WS test via `websocat ws://localhost/api/ws` (with dev JWT cookie via `--header`): upgrade succeeds, subscribe to `user:{your-id}` returns ack
- [ ] Trigger a real PK invite from a second account — observe it arriving via both SSE (existing flow) AND the websocat connection (double-publish working)
- [ ] Run a second dx-api container locally, connect websocat to each, verify cross-instance delivery
- [ ] Kill & restart Redis container, verify publishers resume within ~2s

**Rollback**: `git revert <pr1>`. SSE flows still work (never changed). Zero user impact.

**Size**: ~2000 LOC added, ~25 new files.

### PR 2 — Mount WebSocketProvider + migrate useUserEvents

**Goal**: first user-visible change. Mount provider at hall layout, switch `PkInvitationProvider` to `useUserEvents`.

**Frontend modified**:
- `src/app/(web)/hall/layout.tsx` — wrap children in `<WebSocketProvider>`
- `src/features/web/play-pk/components/pk-invitation-provider.tsx` — import `useUserEvents` instead of `useUserSSE`, remove `POST /api/user/ping` call

**Frontend deleted**:
- `src/hooks/use-user-sse.ts`

**Tests added**: `WebSocketProvider` tests (mount, reconnect backoff, close code dispatch, slow consumer)

**Manual verification**:
```bash
docker compose -f deploy/docker-compose.dev.yml up -d
```
- [ ] PR 1 regression passes (all existing flows)
- [ ] Browser devtools Network → WS shows `/api/ws` open, subscribe frame to `user:{id}`
- [ ] Second account sends PK invite → modal appears within 1-2s
- [ ] Accept → `pk_next_level` auto-navigates to play-pk page
- [ ] Network toggle in devtools: disable 3s, re-enable → WS reconnects with backoff, next invite still delivers
- [ ] Issue C test: log in with same account in second browser → first browser redirects to `/auth/signin?reason=session_replaced`
- [ ] Legacy `/api/user/events` endpoint still reachable via curl (not yet deleted; retired in PR 6)

**Rollback**: `git revert <pr2>`. Users may see a 1-2s reconnect glitch during revert deploy.

**Size**: ~100 LOC modified, 1 file deleted.

### PR 3 — usePkEvents migration

**Goal**: move PK play events (pk-room + play-pk pages) from SSE to WS.

**Frontend modified**:
- `src/features/web/play-pk/hooks/use-pk-play-events.ts` — inner import `usePkSSE` → `usePkEvents`
- `src/app/(web)/hall/pk-room/[id]/page.tsx` — inner import `usePkSSE` → `usePkEvents`

**Frontend deleted**:
- `src/hooks/use-pk-sse.ts`

**Tests added**: unit test for `use-pk-play-events.ts` after migration

**Manual verification**:
- [ ] PRs 1-2 regression passes
- [ ] Full specified PK: invite → accept → play → complete → next level → force-end
- [ ] Random PK: robot actions arrive via WS
- [ ] PK presence check: disconnect one browser mid-match, opponent gets robot fallback (tests `realtime.IsPresent`)
- [ ] PK force-end delivers `pk_force_end` to both players

**Rollback**: `git revert <pr3>`. PK reverts to SSE.

**Size**: ~50 LOC modified, 1 file deleted.

### PR 4 — useGroupEvents migration (play + room)

**Goal**: move group play + room waiting area events to WS.

**Frontend modified**:
- `src/features/web/play-group/hooks/use-group-play-events.ts` — inner import `useGroupSSE` → `useGroupEvents`
- `src/features/web/groups/hooks/use-group-events.ts` — **rename file to `use-group-room-events.ts`** to avoid shadowing the primitive, update inner import
- `src/features/web/groups/components/group-game-room.tsx` — update import to new hook name

**Frontend deleted**:
- `src/hooks/use-group-sse.ts`

**Tests added**: unit tests for the two migrated wrappers

**Manual verification**:
- [ ] PRs 1-3 regression passes
- [ ] Full group game: owner sets → members enter → start → play → complete → next level → force-end
- [ ] Room presence: join/leave events update avatars (auto-emit on subscribe/unsubscribe)
- [ ] Winner determination on mid-game disconnect (tests `realtime.PresentOnTopic`)
- [ ] `group_dismissed` navigates members back to groups list

**Rollback**: `git revert <pr4>`. Group gameplay reverts to SSE.

**Size**: ~80 LOC modified, 1 file deleted, 1 file renamed.

### PR 5 — useGroupNotify migration

**Goal**: move group metadata updates to WS.

**Frontend modified**:
- `src/features/web/groups/components/group-detail-content.tsx` — uses new hook (same public name after rename)

**Frontend deleted**:
- `src/hooks/use-group-notify.ts` (the old one)

**Frontend renamed**:
- `src/hooks/use-group-notify-ws.ts` → `src/hooks/use-group-notify.ts` (new hook takes the stable name)

**Tests added**: unit test for renamed hook

**Manual verification**:
- [ ] PRs 1-4 regression passes
- [ ] Kick member → members list refreshes (`scope:"members"`)
- [ ] Apply to group → application list refreshes (`scope:"applications"`)
- [ ] Create/update/delete subgroup → tree refreshes (`scope:"subgroups"`)
- [ ] Set/clear game on group → detail refreshes (`scope:"detail"`)

**Rollback**: `git revert <pr5>`. Notify reverts to SSE.

**Size**: ~40 LOC modified, 1 file deleted, 1 file renamed.

### PR 6 — Cleanup

**Goal**: delete legacy code, remove double-publish, convert ai-custom to NDJSON, delete stale docs. **Atomic deletion.**

**Backend deleted**:
- `app/helpers/sse.go`
- `app/helpers/sse_hub.go`
- `app/helpers/sse_pk_hub.go`
- `app/helpers/sse_user_hub.go`
- `app/helpers/sse_notify_hub.go`
- (Any `sse_*_test.go` files)
- `app/http/controllers/api/user_sse_controller.go`
- `app/http/controllers/api/group_notify_controller.go`
- `Events()` methods from `game_play_pk_controller.go` and `group_game_controller.go` (keep the rest of each controller)

**Backend modified**:
- All service call sites from Section 6 — remove the legacy `Hub.Broadcast/SendToUser/Notify` lines, keep only `realtime.Publish`
- `routes/api.go` — remove SSE route registrations:
  - `/user/events`, `/user/ping`, `/groups/{id}/events`, `/groups/{id}/notify`, `/play-pk/{id}/events`
- `app/http/controllers/api/ai_custom_controller.go` — swap SSE framing (`data: ...\n\n`) for NDJSON (`...\n`); change content-type header
- `app/http/controllers/api/ai_custom_vocab_controller.go` — same treatment
- Create `app/helpers/ndjson.go` with a minimal writer helper replacing the deleted `sse.go`

**Frontend modified**:
- `src/features/web/ai-custom/helpers/stream-progress.ts` — update line parser for `application/x-ndjson` instead of SSE framing

**Stale docs deleted**:
- `docs/dx-auth-design.md`
- `docs/energy-bean-rule.md`
- `docs/game-word-sentence-group-rule.md`
- `docs/game-word-sentence-single-rule.md`
- `docs/migration-plan.md`

Before deletion, defensive grep across the entire repo to ensure nothing references these paths:
```bash
grep -rn "dx-auth-design\|energy-bean-rule\|game-word-sentence-group-rule\|game-word-sentence-single-rule\|docs/migration-plan" .
```
If anything references them (README links, CI configs), update or remove those references first.

**Tests**:
- Delete SSE-specific tests that no longer apply
- `go test -race ./...` + `npm run build && npm run lint` pass

**Manual verification**:
```bash
docker compose -f deploy/docker-compose.dev.yml down
docker compose -f deploy/docker-compose.dev.yml up --build -d
docker compose -f deploy/docker-compose.prod.yml build  # verify prod build path still works
```
- [ ] PRs 1-5 regression passes end-to-end
- [ ] `grep -rn "PkHub\|GroupSSEHub\|GroupNotifyHub\|UserHub\b\|SSEConnection\|sse_hub\.go\|sse_pk_hub\.go\|sse_user_hub\.go\|sse_notify_hub\.go" dx-api/app/ dx-api/routes/` → zero hits
- [ ] `grep -rn "useUserSSE\|usePkSSE\|useGroupSSE\|use-user-sse\|use-pk-sse\|use-group-sse" dx-web/src/` → zero hits
- [ ] `grep -rn "text/event-stream" dx-api/ dx-web/` → zero hits
- [ ] 5 stale doc files are gone
- [ ] Full PK flow works end-to-end
- [ ] Full group flow works end-to-end
- [ ] AI-custom break-metadata works end-to-end with NDJSON framing (word + vocab variants)
- [ ] `docker compose -f deploy/docker-compose.prod.yml build` succeeds

**Rollback** (asymmetric — call out explicitly): reverting PR 6 leaves service code calling only `realtime.Publish` (double-publish was already removed). A clean `git revert` puts back the old hub files but the legacy calls are missing, so no events reach SSE clients. **Correct rollback is either forward-fix or cascade revert** (revert PR 6 + manually re-add legacy publish calls in services). Recommended safety net: 24-hour staging soak before merging PR 6.

**Size**: ~1500 LOC deleted, ~100 LOC modified (NDJSON + route cleanup), 12 files deleted (code + docs).

### Totals across the 6 PRs

| PR | LOC added | LOC removed | New files | Deleted files | Risk |
|---|---|---|---|---|---|
| 1 | ~2000 | ~10 | ~25 | 0 | Medium |
| 2 | ~80 | ~50 | 0 | 1 | Low |
| 3 | ~30 | ~40 | 0 | 1 | Low |
| 4 | ~50 | ~50 | 0 | 1 | Low |
| 5 | ~25 | ~35 | 0 | 1 (rename) | Low |
| 6 | ~100 | ~1500 | 0 | 12 | Medium |

**Calendar estimate**: 2-3 weeks focused work with staging soak between PRs. Could compress to ~1 week by stacking PRs and skipping soak — not recommended given the user constraint "don't break existing functions".

### Per-PR execution discipline

For each PR:
1. TDD — tests first
2. Implementation
3. Local `go test -race ./... && go build ./...` + `npm run build && npm run lint`
4. Manual verification via `docker compose -f deploy/docker-compose.dev.yml up`
5. Self-review of diff
6. Commit (one logical commit per PR, or 2-3 if clean progression)
7. Push + open PR
8. Code review (via code-reviewer agent per project workflow); address CRITICAL + HIGH
9. Merge
10. Deploy to staging, manual test
11. Monitor production for 24h before proceeding to next PR

---

## 12. Non-Goals (Explicitly Deferred)

- **Load testing** — separate future initiative
- **Chaos engineering** — random Redis kills, network partitions
- **Playwright/Cypress E2E** — manual checklists substitute
- **Performance regression benchmarks** — qualitative monitoring only
- **Message replay / persistent event log** — clients re-query state via REST on reconnect
- **Exactly-once delivery** — at-most-once is sufficient for gameplay
- **Binary frames** — JSON text only
- **Horizontal Redis scaling (Sentinel/Cluster)** — single Redis with managed service option is the path
- **TLS termination in dx-api** — nginx handles TLS (already configured)
- **Proactive token refresh for long-lived WS** — deferred until concrete need; current model checks auth at upgrade and re-checks only on reconnect

---

## Appendix A — Error Code Additions

Additions to `dx-api/app/consts/error_code.go`:

```go
// 400xx: Validation (append to existing block)
CodeInvalidEnvelope = 40022  // WS envelope missing or malformed
CodeUnknownOp       = 40023  // WS envelope op value not recognized
CodeInvalidTopic    = 40024  // WS topic string doesn't match known patterns

// 500xx: Server Error (append to existing block)
CodeSlowConsumer    = 50003  // WS client kicked due to send queue overflow
```

All other error conditions reuse existing constants:
- `CodeForbidden = 40300` — generic topic authorization failure
- `CodeGroupForbidden = 40301` — non-member tries to subscribe to group topic
- `CodeUnauthorized = 40100` — upgrade-time auth failure (HTTP 401 before upgrade)
- `CodeSessionReplaced = 40104` — single-device kick (also sent via WS close code 4001)
- `CodeInvalidToken = 40102` — token parse failure at upgrade

---

## Appendix B — Deploy Directory Touchpoints

Files under `deploy/` that change or need verification during the refactor:

| File | Change | PR |
|---|---|---|
| `deploy/nginx/nginx.prod.conf` | Add `Upgrade`/`Connection` headers for `/api/*`, raise `proxy_read_timeout`/`proxy_send_timeout` to 1h | PR 1 |
| `deploy/nginx/nginx.dev.conf` | Verify HMR WebSocket routing doesn't conflict with `/api/ws` (should be fine — different paths) | PR 1 (verify only) |
| `deploy/docker-compose.dev.yml` | No change expected | All PRs (verify regression) |
| `deploy/docker-compose.prod.yml` | No change expected | PR 6 (verify build) |
| `deploy/env/.env.dev` | No new env vars (CORS_ALLOWED_ORIGINS reused for WS origin check) | — |
| `deploy/env/.env.example` | No new env vars | — |
| `deploy/postgres/init/*.sql` | No change | — |

Verification commands for each PR:
```bash
# Dev manual testing
docker compose -f deploy/docker-compose.dev.yml up -d

# Prod build path validation (PR 1 + PR 6)
docker compose -f deploy/docker-compose.prod.yml build
```

---

## Appendix C — Dependency Additions

`dx-api/go.mod`:
```
require (
    github.com/coder/websocket v1.8.x
    github.com/alicebob/miniredis/v2 v2.x  // test only
)
```

`dx-web/package.json`: no new deps (browser native WebSocket used).

---

## Appendix D — Minor Implementation-Time Lookups

These are narrow technical lookups that happen naturally during implementation and don't affect any architectural decision in this design:

1. **Exact table names for participant/member checks** in `authorize.go`. Will be discovered by reading `pk_invite_service.go` / `group_member_service.go` during PR 1 implementation.
2. **Exact Goravel terminating hook API.** `app.Terminating(...)` is the assumed form; verify against `bootstrap/app.go` conventions during PR 1.
3. **Frontend test framework** — Jest or Vitest? Check `dx-web/package.json` during PR 2 implementation and write tests matching existing conventions.

*Resolved during design:* the `request_timeout` middleware behavior was verified against `goravel/gin@v1.17.0` source and the resolution is baked into Section 3 (blocker fix) and Section 6 (detached context in `WSController`).

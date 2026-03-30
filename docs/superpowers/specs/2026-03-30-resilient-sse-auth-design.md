# Resilient SSE Auth for Group Workflows

## Problem

The two SSE hooks (`use-group-events.ts`, `use-group-play-events.ts`) capture the JWT access token once at mount time and embed it in the EventSource URL. The access token expires after 10 minutes (`JWT_TTL=10`). After expiry:

1. **Active connection drops** — when the network hiccups or server restarts, EventSource tries to reconnect with the same expired token URL → 401 forever
2. **New connection fails** — page refresh loses the in-memory token, hook bails out (`if (!token) return`) → SSE never established
3. **Cascading failures** — room presence breaks, game_start / level_complete / next_level events are never received, players get stuck on waiting screens

The refresh token (7-day TTL, `dx_refresh` cookie) is alive but no code path uses it for SSE connections.

## Scope

**In scope:**
- Create a shared resilient SSE hook (`useGroupSSE`)
- Refactor `use-group-events.ts` and `use-group-play-events.ts` to use it
- Handle token refresh on SSE auth failure
- Handle reconnection with exponential backoff
- Handle missing token on mount (page refresh)

**Out of scope:**
- Backend changes (none needed)
- The play-single 404 error (separate stale session issue)
- AI streaming SSE in `stream-progress.ts` (uses fetch, not EventSource; has its own error handling)

## Design

### New shared hook: `useGroupSSE`

Location: `dx-web/src/hooks/use-group-sse.ts`

```typescript
function useGroupSSE(
  groupId: string | null,
  listeners: Record<string, (data: unknown) => void>
): void
```

**Parameters:**
- `groupId` — group to connect to; `null` disables the connection
- `listeners` — map of SSE event names to handler functions (e.g., `{ "group_game_start": handler }`)

**Lifecycle:**

```
Mount (groupId set)
  │
  ▼
getAccessToken()
  │
  ├─ null ──► refreshAccessToken() ──► retry connect()
  │
  ├─ valid ──► new EventSource(url)
  │              │
  │              ├─ onopen ──► reset retryCount to 0
  │              │
  │              ├─ event ──► call listeners[event](data)
  │              │
  │              └─ onerror ──► close EventSource
  │                              │
  │                              ▼
  │                        retryCount < MAX_RETRIES?
  │                         ├─ yes ──► wait(backoff) ──► refreshAccessToken() ──► connect()
  │                         └─ no  ──► give up (stop retrying)
  │
Unmount
  │
  ▼
close EventSource, clear timers, set disposed=true
```

**Backoff schedule:** `min(1000 * 2^(retryCount-1), 30000)` ms → 1s, 2s, 4s, 8s, 16s, 30s, 30s...

**Max retries:** 10 — after this, stop retrying. The user is likely logged out or has a persistent network issue.

**Key behaviors:**
- `listenersRef` pattern: handlers are always read from ref, so the hook doesn't re-establish SSE on handler changes (same as current hooks)
- `disposed` flag: prevents stale async callbacks from creating connections after unmount
- On refresh failure: stop retrying (the `refreshAccessToken` function handles redirect to `/auth/signin` internally)

### Refactored hooks

Both hooks become thin wrappers that map their typed handler props to the `listeners` record.

**`use-group-events.ts`** — no API change:
```typescript
export function useGroupEvents(groupId: string | null, handlers: GroupEventHandlers) {
  const handlersRef = useRef(handlers);
  handlersRef.current = handlers;

  const listeners = useMemo(() => ({
    group_game_start: (data: unknown) => handlersRef.current.onGameStart?.(data as GroupGameStartEvent),
    group_level_complete: (data: unknown) => handlersRef.current.onLevelComplete?.(data as GroupLevelCompleteEvent),
    group_game_force_end: (data: unknown) => handlersRef.current.onForceEnd?.(data as GroupForceEndEvent),
    room_member_joined: (data: unknown) => handlersRef.current.onRoomMemberJoined?.(data as RoomMemberEvent),
    room_member_left: (data: unknown) => handlersRef.current.onRoomMemberLeft?.(data as RoomMemberEvent),
  }), []);

  useGroupSSE(groupId, listeners);
}
```

**`use-group-play-events.ts`** — no API change:
```typescript
export function useGroupPlayEvents(groupId: string | null, handlers: GroupPlayEventHandlers) {
  const handlersRef = useRef(handlers);
  handlersRef.current = handlers;

  const listeners = useMemo(() => ({
    group_level_complete: (data: unknown) => handlersRef.current.onLevelComplete?.(data as GroupLevelCompleteEvent),
    group_game_force_end: (data: unknown) => handlersRef.current.onForceEnd?.(data as GroupForceEndEvent),
    group_next_level: (data: unknown) => handlersRef.current.onNextLevel?.(data as GroupNextLevelEvent),
    group_player_complete: (data: unknown) => handlersRef.current.onPlayerComplete?.(data as GroupPlayerCompleteEvent),
  }), []);

  useGroupSSE(groupId, listeners);
}
```

### Backend compatibility

No backend changes needed. The existing SSE infrastructure handles reconnection gracefully:

- `SSEHub.Register()` (`sse_hub.go:38-39`): If a user reconnects while already connected, it closes the old connection (`close(old.done)`) and replaces it — no duplicate connections
- `SSEHub.Unregister()` (`sse_hub.go:56-58`): Only removes if the pointer matches the current connection — a replaced connection's deferred cleanup won't accidentally delete the new one
- `ParseJWTUserID()`: Validates whatever token is in the URL — works the same with a fresh token
- Heartbeats (30s): Keep the connection alive between events; the new hook doesn't interfere with this

### EventSource behavior on 401

Important: when the server returns HTTP 401 (non-200, Content-Type is `application/json` not `text/event-stream`), the browser:
1. Fires the `error` event
2. Sets `readyState` to `CLOSED`
3. Does NOT auto-reconnect (auto-reconnect only happens for network-level drops on an established connection)

So the hook must handle all reconnection manually, which gives us full control over token refresh timing.

### Edge cases

| Case | Behavior |
|------|----------|
| Token null on mount (page refresh) | `refreshAccessToken()` called first → cookie sends `dx_refresh` → new access token → connect |
| Token expires during active connection | Connection stays alive (JWT only checked at connect time). On next network blip → onerror → refresh → reconnect |
| Two rapid page navigations | First hook cleans up (`disposed = true`), second hook connects fresh |
| `refreshAccessToken()` fails (logged out) | It internally redirects to `/auth/signin` — hook stops retrying |
| Server restart / deploy | Existing connection drops → onerror → backoff → refresh → reconnect |
| User kicked (code 40104) | `refreshAccessToken()` detects code 40104 → alerts → redirects to signin |
| Multiple rapid errors | Exponential backoff prevents hammering the server |
| `groupId` changes | Effect cleanup closes old EventSource, new one connects to new group |

## Files Changed

| File | Change |
|------|--------|
| `dx-web/src/hooks/use-group-sse.ts` | **New** — shared resilient SSE hook |
| `dx-web/src/features/web/groups/hooks/use-group-events.ts` | Refactor to thin wrapper around `useGroupSSE` |
| `dx-web/src/features/web/play-group/hooks/use-group-play-events.ts` | Refactor to thin wrapper around `useGroupSSE` |
| `docs/game-lsrw-group-rule.md` | Update SSE section to document reconnection behavior |

## Testing

- Sign in → enter game room → wait 11+ minutes → verify room presence still correct (SSE reconnected)
- Start group game → play for 11+ minutes → verify level_complete event still received
- Refresh page while in game room → verify SSE reconnects after token refresh
- Open browser DevTools Network tab → verify reconnection attempts use fresh tokens (different JWT in URL)
- Kill the Go server → restart → verify SSE reconnects within backoff window

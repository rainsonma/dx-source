# Group Detail Real-Time Notifications

## Problem

The group detail page has no real-time updates. When a member applies to join, the owner must refresh to see the new application in the 待审批 modal. When the owner sets/clears the current game, other members must refresh to see the updated game section and 进入教室 button. Member and subgroup list changes also require a manual refresh.

## Solution

Add a separate, lightweight notification SSE endpoint for the group detail page. When any group state changes, the backend broadcasts a `group_updated` event with a `scope` field. The frontend maps the scope to SWR cache keys and triggers revalidation. No payload — just a poke to refetch.

## Design Decisions

- **Separate hub from game room SSE** — The existing `GroupSSEHub` doubles as room presence tracking (`room_member_joined`/`room_member_left` broadcasts on connect/disconnect). Connecting from the detail page would pollute room presence and break game mechanics. A separate `GroupNotifyHub` has no presence semantics.
- **Separate endpoint** — `GET /api/groups/{id}/notify?token=JWT` alongside the existing `GET /api/groups/{id}/events?token=JWT`. Different hubs, different purposes.
- **SWR revalidation, not payload push** — The SSE event carries only `{ "scope": "applications" }`, not the actual data. The frontend triggers `mutate()` on the corresponding SWR key, which refetches from the API. Single source of truth stays in the API response. Simpler, less error-prone.
- **All members connect** — Every group member viewing the detail page gets the notification SSE. Owner additionally benefits from application notifications. Non-members are rejected (same auth pattern as game room SSE).
- **One event type, scope-based routing** — A single `group_updated` event type with a `scope` string. The frontend maps scope → SWR key(s). Easy to extend.

## Backend

### New: GroupNotifyHub (`app/helpers/sse_notify_hub.go`)

Global instance `GroupNotifyHub`, same concurrency pattern as `GroupSSEHub`:

```go
type NotifyHub struct {
    mu    sync.RWMutex
    conns map[string]map[string]*SSEConnection // groupID → userID → connection
}

var GroupNotifyHub = &NotifyHub{
    conns: make(map[string]map[string]*SSEConnection),
}
```

Methods:
- `Register(groupID, userID string, conn *SSEConnection)` — Adds connection, replaces existing for same user. **No broadcast** (no presence tracking).
- `Unregister(groupID, userID string, conn *SSEConnection)` — Removes connection with pointer equality check. **No broadcast**.
- `Notify(groupID, scope string)` — Broadcasts `group_updated` event with `{"scope": "<scope>"}` to all connections in the group. Fire-and-forget.

Heartbeat: 30-second keepalive comments (reuse `SSEConnection.SendHeartbeat()`).

### New: Notify Endpoint (`app/http/controllers/api/group_notify_controller.go`)

`GET /api/groups/{id}/notify?token=JWT`

Handler logic (same pattern as `Events()` in `group_game_controller.go`):
1. Parse JWT from `token` query param via `ParseJWTUserID()`
2. Verify user is a `GameGroupMember` of this group
3. Set SSE response headers
4. Create `SSEConnection`, register in `GroupNotifyHub`
5. Heartbeat loop (30s) + listen for client disconnect
6. On disconnect: `GroupNotifyHub.Unregister()`

### New Route (`routes/api.go`)

```go
// Group notification SSE (query-param auth, no middleware)
router.Get("/api/groups/{id}/notify", groupNotifyController.Notify)
```

Placed alongside the existing `/api/groups/{id}/events` route (public, query-param auth).

### Modified: Service Functions — Add Notify() Calls

Each mutation adds fire-and-forget `Notify()` calls after the operation succeeds:

**`group_application_service.go`:**

| Function | Notify scopes |
|----------|---------------|
| `ApplyToGroup()` | `"applications"` |
| `CancelApplication()` | `"applications"` |
| `HandleApplication()` — accept | `"applications"`, `"members"`, `"detail"` |
| `HandleApplication()` — reject | `"applications"` |

**`group_game_service.go`:**

| Function | Notify scopes |
|----------|---------------|
| `SetGroupGame()` | `"detail"` |
| `ClearGroupGame()` | `"detail"` |
| `StartGroupGame()` | `"detail"` |
| `ForceEndGroupGame()` | `"detail"` |

**`group_member_service.go`:**

| Function | Notify scopes |
|----------|---------------|
| `LeaveGroup()` | `"members"`, `"detail"` |
| `KickMember()` | `"members"`, `"detail"` |

Note: `JoinByCode()` calls `ApplyToGroup()` internally, so the `"applications"` notify from `ApplyToGroup` covers it.

**`group_subgroup_service.go`:**

| Function | Notify scopes |
|----------|---------------|
| `CreateSubgroup()` | `"subgroups"` |
| `DeleteSubgroup()` | `"subgroups"`, `"detail"` |
| `UpdateSubgroup()` | `"subgroups"` |
| `AssignSubgroupMembers()` | `"subgroups"` |
| `RemoveSubgroupMember()` | `"subgroups"` |

## Frontend

### New: useGroupNotify Hook (`src/hooks/use-group-notify.ts`)

Same EventSource pattern as `use-group-sse.ts`:
- Connects to `${API_URL}/api/groups/{groupId}/notify?token={jwt}`
- Token from `getToken()`, refresh via `refreshAccessToken()` on 401
- Exponential backoff reconnection (1s → 30s, max 10 retries)
- Listens for `group_updated` event
- Parses `{ scope: string }` from event data
- Calls provided callback with the scope string
- Cleanup on unmount

```typescript
function useGroupNotify(groupId: string, onUpdate: (scope: string) => void): void
```

### Modified: group-detail-content.tsx

Wire up the hook:

```typescript
useGroupNotify(id, (scope) => {
  if (scope === "applications") mutate(`/api/groups/${id}/applications`)
  if (scope === "members") mutate(`/api/groups/${id}/members`)
  if (scope === "subgroups") mutate(`/api/groups/${id}/subgroups`)
  if (scope === "detail") mutate(`/api/groups/${id}`)
})
```

No new components, no new state, no new types beyond the hook.

## Scopes Reference

| Scope | SWR key invalidated | What updates |
|-------|---------------------|--------------|
| `applications` | `/api/groups/{id}/applications` | 待审批 badge count + modal list |
| `members` | `/api/groups/{id}/members` | 群成员 list |
| `subgroups` | `/api/groups/{id}/subgroups` | 群小组 list |
| `detail` | `/api/groups/{id}` | Game section, 进入教室 button, member count, all group info |

## Files Changed

### Backend (dx-api)

| File | Change |
|------|--------|
| `app/helpers/sse_notify_hub.go` | **New** — `GroupNotifyHub` |
| `app/http/controllers/api/group_notify_controller.go` | **New** — SSE notify endpoint |
| `routes/api.go` | **Modified** — Add notify route |
| `app/services/api/group_application_service.go` | **Modified** — Add `Notify()` calls |
| `app/services/api/group_game_service.go` | **Modified** — Add `Notify()` calls |
| `app/services/api/group_member_service.go` | **Modified** — Add `Notify()` calls |
| `app/services/api/group_subgroup_service.go` | **Modified** — Add `Notify()` calls |

### Frontend (dx-web)

| File | Change |
|------|--------|
| `src/hooks/use-group-notify.ts` | **New** — notification SSE hook |
| `src/features/web/groups/components/group-detail-content.tsx` | **Modified** — wire `useGroupNotify` |

### Docs

| File | Change |
|------|--------|
| `docs/game-lsrw-group-rule.md` | **Modified** — Add notification SSE section |

## What Stays Untouched

- `GroupSSEHub` and `sse_hub.go` — unchanged
- Game room SSE endpoint and `group_game_controller.go` `Events()` — unchanged
- All existing components, types, actions — unchanged
- Game play flow — unchanged
- All existing SSE events (`group_game_start`, `room_member_joined`, etc.) — unchanged

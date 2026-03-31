# Group Detail Real-Time Notifications — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a lightweight notification SSE system so the group detail page updates in real time when applications, game config, members, or subgroups change.

**Architecture:** A separate `GroupNotifyHub` (no presence tracking) with its own SSE endpoint `/api/groups/{id}/notify`. Backend mutations call `Notify(groupID, scope)` fire-and-forget. Frontend hook maps scope to SWR cache key and triggers revalidation.

**Tech Stack:** Go/Goravel (backend SSE hub + endpoint), React/Next.js + SWR (frontend hook)

---

### Task 1: GroupNotifyHub — Backend SSE Hub

**Files:**
- Create: `dx-api/app/helpers/sse_notify_hub.go`

- [ ] **Step 1: Create the NotifyHub struct and global instance**

```go
// dx-api/app/helpers/sse_notify_hub.go
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
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 3: Commit**

```bash
cd dx-api
git add app/helpers/sse_notify_hub.go
git commit -m "feat: add GroupNotifyHub for group detail notifications"
```

---

### Task 2: Notify SSE Endpoint — Backend Controller & Route

**Files:**
- Create: `dx-api/app/http/controllers/api/group_notify_controller.go`
- Modify: `dx-api/routes/api.go:78-80`

- [ ] **Step 1: Create the notify controller**

```go
// dx-api/app/http/controllers/api/group_notify_controller.go
package api

import (
	nethttp "net/http"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/helpers"
	"dx-api/app/models"
)

type GroupNotifyController struct{}

func NewGroupNotifyController() *GroupNotifyController {
	return &GroupNotifyController{}
}

// Notify establishes a persistent SSE connection for group detail notifications.
func (c *GroupNotifyController) Notify(ctx contractshttp.Context) contractshttp.Response {
	token := ctx.Request().Query("token", "")
	if token == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, 0, "missing token")
	}

	userID, err := helpers.ParseJWTUserID(token)
	if err != nil {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, 0, "invalid token")
	}

	groupID := ctx.Request().Route("id")

	var member models.GameGroupMember
	if err := facades.Orm().Query().Where("game_group_id", groupID).
		Where("user_id", userID).First(&member); err != nil || member.ID == "" {
		return helpers.Error(ctx, nethttp.StatusForbidden, 0, "not a group member")
	}

	w := ctx.Response().Writer()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	if f, ok := w.(nethttp.Flusher); ok {
		f.Flush()
	}

	conn := helpers.NewSSEConnection(w)
	helpers.GroupNotifyHub.Register(groupID, userID, conn)
	defer helpers.GroupNotifyHub.Unregister(groupID, userID, conn)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	clientGone := ctx.Request().Origin().Context().Done()

	for {
		select {
		case <-clientGone:
			return nil
		case <-conn.Done():
			return nil
		case <-ticker.C:
			if err := conn.SendHeartbeat(); err != nil {
				return nil
			}
		}
	}
}
```

- [ ] **Step 2: Add NewSSEConnection constructor to sse_hub.go**

Add at the end of `dx-api/app/helpers/sse_hub.go` (after the `Done()` method, line 122):

```go
// NewSSEConnection creates an SSEConnection from an http.ResponseWriter.
func NewSSEConnection(w http.ResponseWriter) *SSEConnection {
	flusher, _ := w.(http.Flusher)
	return &SSEConnection{w: w, flusher: flusher, done: make(chan struct{})}
}
```

Also update `SSEHub.Register()` to use this constructor. Replace lines 30-31:

```go
// Before:
flusher, _ := w.(http.Flusher)
conn := &SSEConnection{w: w, flusher: flusher, done: make(chan struct{})}

// After:
conn := NewSSEConnection(w)
```

- [ ] **Step 3: Add the route to api.go**

In `dx-api/routes/api.go`, after the existing SSE events route (line 80), add:

```go
// Group detail notification SSE (query-param auth, not JWT middleware)
groupNotifyController := apicontrollers.NewGroupNotifyController()
router.Get("/groups/{id}/notify", groupNotifyController.Notify)
```

- [ ] **Step 4: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 5: Commit**

```bash
cd dx-api
git add app/http/controllers/api/group_notify_controller.go app/helpers/sse_hub.go routes/api.go
git commit -m "feat: add /api/groups/{id}/notify SSE endpoint"
```

---

### Task 3: Add Notify Calls to Application Service

**Files:**
- Modify: `dx-api/app/services/api/group_application_service.go`

- [ ] **Step 1: Add import for helpers package**

At the top of `dx-api/app/services/api/group_application_service.go`, add `"dx-api/app/helpers"` to the import block (if not already present).

- [ ] **Step 2: Add Notify to ApplyToGroup**

After the `Create(&app)` call succeeds (line 57, before `return app.ID, nil`), add:

```go
helpers.GroupNotifyHub.Notify(groupID, "applications")
```

- [ ] **Step 3: Add Notify to CancelApplication**

After the `Delete` call succeeds (line 73, before `return nil`), add:

```go
helpers.GroupNotifyHub.Notify(groupID, "applications")
```

- [ ] **Step 4: Add Notify to HandleApplication — accept path**

After the transaction succeeds (line 183, after the closing `})` of the transaction, before the closing `}`), restructure slightly. The current code returns the transaction result directly. Change to capture the error, notify on success, then return:

Replace lines 167-183:
```go
// Before:
return facades.Orm().Transaction(func(tx orm.Query) error {
    // ... transaction body ...
})

// After:
if err := facades.Orm().Transaction(func(tx orm.Query) error {
    // ... transaction body unchanged ...
}); err != nil {
    return err
}
helpers.GroupNotifyHub.Notify(groupID, "applications")
helpers.GroupNotifyHub.Notify(groupID, "members")
helpers.GroupNotifyHub.Notify(groupID, "detail")
return nil
```

- [ ] **Step 5: Add Notify to HandleApplication — reject path**

After the reject update succeeds (line 188, before `return nil` on line 190), add:

```go
helpers.GroupNotifyHub.Notify(groupID, "applications")
```

- [ ] **Step 6: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 7: Commit**

```bash
cd dx-api
git add app/services/api/group_application_service.go
git commit -m "feat: notify group detail on application changes"
```

---

### Task 4: Add Notify Calls to Game Service

**Files:**
- Modify: `dx-api/app/services/api/group_game_service.go`
- Modify: `dx-api/app/services/api/game_play_group_service.go`

- [ ] **Step 1: Add import for helpers package**

In `dx-api/app/services/api/group_game_service.go`, add `"dx-api/app/helpers"` to imports if not already present (it's already imported for `GroupSSEHub` usage — verify).

- [ ] **Step 2: Add Notify to SetGroupGame**

After the `Update` call succeeds (line 113, before `return nil` on line 114), add:

```go
helpers.GroupNotifyHub.Notify(groupID, "detail")
```

- [ ] **Step 3: Add Notify to ClearGroupGame**

After the `Exec` call succeeds (line 135, before `return nil` on line 136), add:

```go
helpers.GroupNotifyHub.Notify(groupID, "detail")
```

- [ ] **Step 4: Add Notify to StartGroupGame**

After the `Broadcast` call (line 344, before `return nil` on line 346), add:

```go
helpers.GroupNotifyHub.Notify(groupID, "detail")
```

- [ ] **Step 5: Add Notify to ForceEndGroupGame**

After the `Broadcast` call (line 420, before `return results, nil` on line 422), add:

```go
helpers.GroupNotifyHub.Notify(groupID, "detail")
```

- [ ] **Step 6: Add Notify to auto-end in game_play_group_service.go**

In `dx-api/app/services/api/game_play_group_service.go`, after the auto-end `is_playing = false` update (line 411), add:

```go
helpers.GroupNotifyHub.Notify(*session.GameGroupID, "detail")
```

Make sure `"dx-api/app/helpers"` is imported (it's already imported for `GroupSSEHub.Broadcast` — verify).

- [ ] **Step 7: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 8: Commit**

```bash
cd dx-api
git add app/services/api/group_game_service.go app/services/api/game_play_group_service.go
git commit -m "feat: notify group detail on game config and play state changes"
```

---

### Task 5: Add Notify Calls to Member Service

**Files:**
- Modify: `dx-api/app/services/api/group_member_service.go`

- [ ] **Step 1: Add import for helpers package**

Add `"dx-api/app/helpers"` to imports in `dx-api/app/services/api/group_member_service.go`.

- [ ] **Step 2: Add Notify to KickMember**

After `removeMemberFromGroup` succeeds (line 130, before the closing `}`), restructure to capture the error:

```go
// Before:
return removeMemberFromGroup(groupID, targetUserID)

// After:
if err := removeMemberFromGroup(groupID, targetUserID); err != nil {
    return err
}
helpers.GroupNotifyHub.Notify(groupID, "members")
helpers.GroupNotifyHub.Notify(groupID, "detail")
return nil
```

- [ ] **Step 3: Add Notify to LeaveGroup**

After `removeMemberFromGroup` succeeds (line 139, before the closing `}`), restructure:

```go
// Before:
return removeMemberFromGroup(groupID, userID)

// After:
if err := removeMemberFromGroup(groupID, userID); err != nil {
    return err
}
helpers.GroupNotifyHub.Notify(groupID, "members")
helpers.GroupNotifyHub.Notify(groupID, "detail")
return nil
```

- [ ] **Step 4: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 5: Commit**

```bash
cd dx-api
git add app/services/api/group_member_service.go
git commit -m "feat: notify group detail on member changes"
```

---

### Task 6: Add Notify Calls to Subgroup Service

**Files:**
- Modify: `dx-api/app/services/api/group_subgroup_service.go`

- [ ] **Step 1: Add import for helpers package**

Add `"dx-api/app/helpers"` to imports in `dx-api/app/services/api/group_subgroup_service.go`.

- [ ] **Step 2: Add Notify to CreateSubgroup**

In `CreateSubgroup`, after the `Create(&sub)` call succeeds (before `return sub.ID, nil`), add:

```go
helpers.GroupNotifyHub.Notify(groupID, "subgroups")
```

- [ ] **Step 3: Add Notify to UpdateSubgroup**

After the `Update("name", name)` call succeeds (line 142, before `return nil` on line 143), add:

```go
helpers.GroupNotifyHub.Notify(groupID, "subgroups")
```

- [ ] **Step 4: Add Notify to DeleteSubgroup**

`DeleteSubgroup` returns the transaction result directly. Restructure:

```go
// Before:
return facades.Orm().Transaction(func(tx orm.Query) error {
    // ... transaction body ...
})

// After:
if err := facades.Orm().Transaction(func(tx orm.Query) error {
    // ... transaction body unchanged ...
}); err != nil {
    return err
}
helpers.GroupNotifyHub.Notify(groupID, "subgroups")
helpers.GroupNotifyHub.Notify(groupID, "detail")
return nil
```

- [ ] **Step 5: Add Notify to AssignSubgroupMembers**

Same pattern — restructure the transaction return:

```go
// Before:
return facades.Orm().Transaction(func(tx orm.Query) error {
    // ... transaction body ...
})

// After:
if err := facades.Orm().Transaction(func(tx orm.Query) error {
    // ... transaction body unchanged ...
}); err != nil {
    return err
}
helpers.GroupNotifyHub.Notify(groupID, "subgroups")
return nil
```

- [ ] **Step 6: Add Notify to RemoveSubgroupMember**

After the `Delete` call succeeds (line 262, before `return nil` on line 263), add:

```go
helpers.GroupNotifyHub.Notify(groupID, "subgroups")
```

- [ ] **Step 7: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 8: Commit**

```bash
cd dx-api
git add app/services/api/group_subgroup_service.go
git commit -m "feat: notify group detail on subgroup changes"
```

---

### Task 7: useGroupNotify Hook — Frontend

**Files:**
- Create: `dx-web/src/hooks/use-group-notify.ts`

- [ ] **Step 1: Create the notification hook**

```typescript
// dx-web/src/hooks/use-group-notify.ts
"use client";

import { useEffect, useRef } from "react";
import { getToken, refreshAccessToken } from "@/lib/api-client";

const MAX_RETRIES = 10;

function backoffDelay(retryCount: number): number {
  return Math.min(1000 * Math.pow(2, retryCount - 1), 30000);
}

export function useGroupNotify(
  groupId: string | null,
  onUpdate: (scope: string) => void
): void {
  const callbackRef = useRef(onUpdate);
  callbackRef.current = onUpdate;

  useEffect(() => {
    if (!groupId) return;

    let eventSource: EventSource | null = null;
    let retryCount = 0;
    let retryTimer: ReturnType<typeof setTimeout> | null = null;
    let disposed = false;

    function connect(token: string) {
      if (disposed) return;

      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "";
      const url = `${apiUrl}/api/groups/${groupId}/notify?token=${encodeURIComponent(token)}`;

      eventSource = new EventSource(url);

      eventSource.addEventListener("group_updated", (e: MessageEvent) => {
        try {
          const data = JSON.parse(e.data) as { scope: string };
          callbackRef.current(data.scope);
        } catch {
          // Discard malformed SSE messages
        }
      });

      eventSource.onopen = () => {
        retryCount = 0;
      };

      eventSource.onerror = () => {
        if (disposed) return;
        eventSource?.close();
        eventSource = null;
        scheduleReconnect();
      };
    }

    function scheduleReconnect() {
      if (disposed) return;
      retryCount++;
      if (retryCount > MAX_RETRIES) return;

      const delay = backoffDelay(retryCount);
      retryTimer = setTimeout(() => {
        if (disposed) return;
        refreshAndConnect();
      }, delay);
    }

    function refreshAndConnect() {
      if (disposed) return;
      refreshAccessToken()
        .then((token) => {
          if (!disposed) connect(token);
        })
        .catch(() => {
          scheduleReconnect();
        });
    }

    // Initial connection: use existing token or refresh first
    const token = getToken();
    if (token) {
      connect(token);
    } else {
      refreshAndConnect();
    }

    return () => {
      disposed = true;
      if (retryTimer) clearTimeout(retryTimer);
      eventSource?.close();
      eventSource = null;
    };
  }, [groupId]);
}
```

- [ ] **Step 2: Verify lint**

Run: `cd dx-web && npx eslint src/hooks/use-group-notify.ts`
Expected: No lint errors.

- [ ] **Step 3: Commit**

```bash
cd dx-web
git add src/hooks/use-group-notify.ts
git commit -m "feat: add useGroupNotify hook for group detail SSE"
```

---

### Task 8: Wire useGroupNotify into Group Detail Page

**Files:**
- Modify: `dx-web/src/features/web/groups/components/group-detail-content.tsx`

- [ ] **Step 1: Add import**

Add at the top of `group-detail-content.tsx` (with the other hook imports):

```typescript
import { useGroupNotify } from "@/hooks/use-group-notify";
```

- [ ] **Step 2: Add the hook call**

Inside the `GroupDetailContent` component, after the existing SWR hooks (after line 77, after the `appsData` SWR call), add:

```typescript
useGroupNotify(id, (scope) => {
  if (scope === "applications") swrMutate(`/api/groups/${id}/applications`);
  if (scope === "members") swrMutate(`/api/groups/${id}/members`);
  if (scope === "subgroups") swrMutate(`/api/groups/${id}/subgroups`);
  if (scope === "detail") swrMutate(`/api/groups/${id}`);
});
```

Note: `swrMutate` is already imported (line 33) and used in this file (line 97). The function accepts SWR key prefixes and invalidates all matching cache entries.

- [ ] **Step 3: Verify lint**

Run: `cd dx-web && npx eslint src/features/web/groups/components/group-detail-content.tsx`
Expected: No lint errors.

- [ ] **Step 4: Verify build**

Run: `cd dx-web && npm run build`
Expected: Build succeeds with no errors.

- [ ] **Step 5: Commit**

```bash
cd dx-web
git add src/features/web/groups/components/group-detail-content.tsx
git commit -m "feat: wire useGroupNotify for real-time group detail updates"
```

---

### Task 9: Update Group Rules Documentation

**Files:**
- Modify: `docs/game-lsrw-group-rule.md`

- [ ] **Step 1: Add Notification SSE section**

In `docs/game-lsrw-group-rule.md`, after the "SSE Events" table (after line 524), add a new section:

```markdown
## Group Detail Notifications

### Notification SSE Endpoint

A separate, lightweight SSE endpoint for pushing real-time updates to the group detail page. Unlike the game room SSE (`/api/groups/{id}/events`), this endpoint has **no presence tracking** — connecting and disconnecting does not broadcast any events and does not affect game mechanics.

**Endpoint:** `GET /api/groups/{id}/notify?token=JWT`

- Query-param JWT auth (same as game room SSE)
- Members only — non-members receive 403
- 30-second heartbeat keepalive
- Exponential backoff reconnection (1s → 30s, max 10 retries)

### Event Format

A single event type `group_updated` with a scope field:

```json
{ "scope": "applications" }
```

### Scopes

| Scope | Trigger | What updates on group detail page |
|-------|---------|----------------------------------|
| `applications` | Member applies, cancels, or owner accepts/rejects | 待审批 badge count + modal list |
| `members` | Member accepted, kicked, or leaves | 群成员 list |
| `subgroups` | Subgroup created, updated, deleted, or members assigned/removed | 群小组 list |
| `detail` | Game set/cleared/started/ended, member count changed, subgroup deleted | Group info card, 当前课程游戏 section, 进入教室 button |

### Frontend Behavior

The group detail page connects to the notification SSE on mount. When a `group_updated` event arrives, the frontend invalidates the corresponding SWR cache key, triggering a refetch from the API. No payload is pushed — the SSE event is a lightweight "poke" to revalidate.
```

- [ ] **Step 2: Add notify endpoint to API Endpoints table**

In the "Group Management" API table (around line 488), add a new row:

```markdown
| GET | `/api/groups/{id}/notify?token=JWT` | SSE connection for group detail notifications (query-param auth) |
```

- [ ] **Step 3: Add group_updated to SSE Events table**

In the "SSE Events" table (around line 524), add a new row:

```markdown
| `group_updated` | Any group state mutation (see Notification Scopes) | `{ scope }` — triggers SWR revalidation on group detail page |
```

- [ ] **Step 4: Commit**

```bash
git add docs/game-lsrw-group-rule.md
git commit -m "docs: add group detail notification SSE to group rules"
```

---

### Task 10: Full Integration Verification

**Files:** None (verification only)

- [ ] **Step 1: Verify Go backend builds and passes vet**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: No errors.

- [ ] **Step 2: Verify frontend builds with no lint errors**

Run: `cd dx-web && npm run lint && npm run build`
Expected: No errors.

- [ ] **Step 3: Manual smoke test**

1. Start backend: `cd dx-api && go run .`
2. Start frontend: `cd dx-web && npm run dev`
3. Open group detail page as owner in browser
4. In a second browser/incognito, apply to join the group
5. Verify: The owner's 待审批 badge count updates without refresh
6. Accept the application
7. Verify: The member list updates for both browsers without refresh
8. Set a course game as owner
9. Verify: The member's 当前课程游戏 section and 进入教室 button update without refresh

- [ ] **Step 4: Final commit (if any fixes needed)**

```bash
git add -A
git commit -m "fix: address integration issues from smoke test"
```

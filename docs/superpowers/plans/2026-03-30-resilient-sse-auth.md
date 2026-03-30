# Resilient SSE Auth Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix SSE connections that die after 10 minutes (JWT expiry) by adding auto-refresh and reconnection to the two group SSE hooks.

**Architecture:** Create a shared `useGroupSSE` hook that manages EventSource lifecycle with token refresh on auth failure and exponential backoff on reconnect. Both existing SSE hooks become thin typed wrappers — their public APIs stay identical.

**Tech Stack:** React hooks, browser EventSource API, existing `refreshAccessToken()` from `api-client.ts`

---

### Task 1: Create the shared `useGroupSSE` hook

**Files:**
- Create: `dx-web/src/hooks/use-group-sse.ts`

- [ ] **Step 1: Create the hook file**

```typescript
"use client";

import { useEffect, useRef } from "react";
import { getAccessToken, refreshAccessToken } from "@/lib/api-client";

const MAX_RETRIES = 10;

function backoffDelay(retryCount: number): number {
  return Math.min(1000 * Math.pow(2, retryCount - 1), 30000);
}

export function useGroupSSE(
  groupId: string | null,
  listeners: Record<string, (data: unknown) => void>
): void {
  const listenersRef = useRef(listeners);
  listenersRef.current = listeners;

  useEffect(() => {
    if (!groupId) return;

    let eventSource: EventSource | null = null;
    let retryCount = 0;
    let retryTimer: ReturnType<typeof setTimeout> | null = null;
    let disposed = false;

    function attachListeners(es: EventSource) {
      for (const eventName of Object.keys(listenersRef.current)) {
        es.addEventListener(eventName, (e: MessageEvent) => {
          const data: unknown = JSON.parse(e.data);
          listenersRef.current[eventName]?.(data);
        });
      }
    }

    function connect(token: string) {
      if (disposed) return;

      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "";
      const url = `${apiUrl}/api/groups/${groupId}/events?token=${encodeURIComponent(token)}`;

      eventSource = new EventSource(url);

      attachListeners(eventSource);

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
          // refreshAccessToken handles redirect to /auth/signin on failure
        });
    }

    // Initial connection: use existing token or refresh first
    const token = getAccessToken();
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

- [ ] **Step 2: Verify the file compiles**

Run: `cd dx-web && npx tsc --noEmit src/hooks/use-group-sse.ts 2>&1 | head -20`

If tsc has trouble with the single file, run the full project check:

Run: `cd dx-web && npx tsc --noEmit 2>&1 | head -20`

Expected: no errors related to `use-group-sse.ts`

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/hooks/use-group-sse.ts
git commit -m "feat: add resilient SSE hook with auto-refresh and reconnection"
```

---

### Task 2: Refactor `useGroupEvents` to use `useGroupSSE`

**Files:**
- Modify: `dx-web/src/features/web/groups/hooks/use-group-events.ts`

- [ ] **Step 1: Replace the file contents**

Replace the entire file with:

```typescript
"use client";

import { useRef, useMemo } from "react";
import { useGroupSSE } from "@/hooks/use-group-sse";
import type {
  GroupGameStartEvent,
  GroupLevelCompleteEvent,
  GroupForceEndEvent,
  RoomMemberEvent,
} from "../types/group";

type GroupEventHandlers = {
  onGameStart?: (event: GroupGameStartEvent) => void;
  onLevelComplete?: (event: GroupLevelCompleteEvent) => void;
  onForceEnd?: (event: GroupForceEndEvent) => void;
  onRoomMemberJoined?: (event: RoomMemberEvent) => void;
  onRoomMemberLeft?: (event: RoomMemberEvent) => void;
};

export function useGroupEvents(
  groupId: string | null,
  handlers: GroupEventHandlers
) {
  const handlersRef = useRef(handlers);
  handlersRef.current = handlers;

  const listeners = useMemo(() => ({
    group_game_start: (data: unknown) =>
      handlersRef.current.onGameStart?.(data as GroupGameStartEvent),
    group_level_complete: (data: unknown) =>
      handlersRef.current.onLevelComplete?.(data as GroupLevelCompleteEvent),
    group_game_force_end: (data: unknown) =>
      handlersRef.current.onForceEnd?.(data as GroupForceEndEvent),
    room_member_joined: (data: unknown) =>
      handlersRef.current.onRoomMemberJoined?.(data as RoomMemberEvent),
    room_member_left: (data: unknown) =>
      handlersRef.current.onRoomMemberLeft?.(data as RoomMemberEvent),
  }), []);

  useGroupSSE(groupId, listeners);
}
```

- [ ] **Step 2: Verify the consumer still compiles**

Run: `cd dx-web && npx tsc --noEmit 2>&1 | grep -i "use-group-events\|group-game-room" | head -10`

Expected: no errors. The `group-game-room.tsx` consumer calls `useGroupEvents(groupId, { onGameStart, onRoomMemberJoined, onRoomMemberLeft })` — the API is unchanged.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/groups/hooks/use-group-events.ts
git commit -m "refactor: use resilient SSE hook in useGroupEvents"
```

---

### Task 3: Refactor `useGroupPlayEvents` to use `useGroupSSE`

**Files:**
- Modify: `dx-web/src/features/web/play-group/hooks/use-group-play-events.ts`

- [ ] **Step 1: Replace the file contents**

Replace the entire file with:

```typescript
"use client";

import { useRef, useMemo } from "react";
import { useGroupSSE } from "@/hooks/use-group-sse";
import type {
  GroupLevelCompleteEvent,
  GroupForceEndEvent,
  GroupNextLevelEvent,
  GroupPlayerCompleteEvent,
} from "../types/group-play";

type GroupPlayEventHandlers = {
  onLevelComplete?: (event: GroupLevelCompleteEvent) => void;
  onForceEnd?: (event: GroupForceEndEvent) => void;
  onNextLevel?: (event: GroupNextLevelEvent) => void;
  onPlayerComplete?: (event: GroupPlayerCompleteEvent) => void;
};

export function useGroupPlayEvents(
  groupId: string | null,
  handlers: GroupPlayEventHandlers
) {
  const handlersRef = useRef(handlers);
  handlersRef.current = handlers;

  const listeners = useMemo(() => ({
    group_level_complete: (data: unknown) =>
      handlersRef.current.onLevelComplete?.(data as GroupLevelCompleteEvent),
    group_game_force_end: (data: unknown) =>
      handlersRef.current.onForceEnd?.(data as GroupForceEndEvent),
    group_next_level: (data: unknown) =>
      handlersRef.current.onNextLevel?.(data as GroupNextLevelEvent),
    group_player_complete: (data: unknown) =>
      handlersRef.current.onPlayerComplete?.(data as GroupPlayerCompleteEvent),
  }), []);

  useGroupSSE(groupId, listeners);
}
```

- [ ] **Step 2: Verify the consumer still compiles**

Run: `cd dx-web && npx tsc --noEmit 2>&1 | grep -i "use-group-play-events\|group-play-shell" | head -10`

Expected: no errors. The `group-play-shell.tsx` consumer calls `useGroupPlayEvents(groupId, { onLevelComplete, onForceEnd, onNextLevel, onPlayerComplete })` — the API is unchanged.

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/play-group/hooks/use-group-play-events.ts
git commit -m "refactor: use resilient SSE hook in useGroupPlayEvents"
```

---

### Task 4: Full build verification

**Files:** (none modified — verification only)

- [ ] **Step 1: Run full TypeScript check**

Run: `cd dx-web && npx tsc --noEmit 2>&1 | tail -5`

Expected: no type errors

- [ ] **Step 2: Run ESLint**

Run: `cd dx-web && npm run lint 2>&1 | tail -10`

Expected: no new errors (pre-existing warnings are fine)

- [ ] **Step 3: Run dev build**

Run: `cd dx-web && npm run build 2>&1 | tail -10`

Expected: build succeeds

- [ ] **Step 4: Commit docs changes**

```bash
git add docs/superpowers/specs/2026-03-30-resilient-sse-auth-design.md
git add docs/superpowers/plans/2026-03-30-resilient-sse-auth.md
git add docs/game-lsrw-group-rule.md
git commit -m "docs: add resilient SSE auth spec, plan, and update group rules"
```

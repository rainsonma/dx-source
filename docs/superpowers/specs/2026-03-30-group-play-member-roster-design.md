# Group Play Member Roster — Design Spec

**Date:** 2026-03-30

## Overview

Display all group members below the progress bar in `GroupPlayTopBar` during gameplay. Solo mode shows a flat list; team mode groups members under subgroup labels. Each avatar shows a live completion indicator via a new per-player SSE event.

## Backend Changes

### 1. Extend `GroupGameStartEvent` payload

Add a `participants` field to the existing `group_game_start` SSE event.

**Solo mode shape:**

```json
{
  "...existing fields...",
  "participants": {
    "mode": "group_solo",
    "members": [
      { "user_id": "abc", "user_name": "Alice" },
      { "user_id": "def", "user_name": "Bob" }
    ]
  }
}
```

**Team mode shape:**

```json
{
  "...existing fields...",
  "participants": {
    "mode": "group_team",
    "teams": [
      {
        "subgroup_id": "sg1",
        "subgroup_name": "Team Alpha",
        "members": [
          { "user_id": "abc", "user_name": "Alice" },
          { "user_id": "def", "user_name": "Bob" }
        ]
      },
      {
        "subgroup_id": "sg2",
        "subgroup_name": "Team Beta",
        "members": [
          { "user_id": "ghi", "user_name": "Carol" }
        ]
      }
    ]
  }
}
```

**Source:** Connected users from `GroupSSEHub.ConnectedUserIDs()` at game start time. For team mode, each connected user is matched to their subgroup via `game_subgroup_members` table.

**Files:** `dx-api/app/services/api/group_game_service.go`

### 2. New SSE event: `group_player_complete`

Broadcast when an individual player completes a level, **before** the existing `CheckAndDetermineWinner` call in `GroupPlayCompleteLevel`.

**Payload:**

```json
{
  "user_id": "abc",
  "user_name": "Alice",
  "game_level_id": "level123"
}
```

Fires for every individual completion. The existing `group_level_complete` event continues to fire only when all connected players finish.

**Files:** `dx-api/app/services/api/game_play_group_service.go`

## Frontend Changes

### 3. State management — `useGroupPlayStore`

**New state fields:**

- `participants`: Roster from `group_game_start`. Union type:
  - Solo: `{ mode: "group_solo"; members: { user_id: string; user_name: string }[] }`
  - Team: `{ mode: "group_team"; teams: { subgroup_id: string; subgroup_name: string; members: { user_id: string; user_name: string }[] }[] }`
- `completedPlayerIds`: `string[]` — user IDs who completed the current level (array, not Set, for Zustand compatibility)

**New actions:**

- `setParticipants(data)` — stores participants from SSE start event
- `addCompletedPlayer(userId)` — appends user to `completedPlayerIds` (deduplicates)

**Existing actions updated:**

- `setGroupResult` / `clearGroupPhase` — also resets `completedPlayerIds`

**Files:** `dx-web/src/features/web/play-group/hooks/use-group-play-store.ts`

### 4. Data flow across navigation boundary

The `group_game_start` SSE event is received in `group-game-room.tsx`. The play-group page loads via `router.push`. Participants are passed across this boundary via `sessionStorage`:

1. `group-game-room.tsx` receives `group_game_start` → stores `participants` in `sessionStorage` keyed by `groupId`
2. `group-play-loading-screen.tsx` reads from `sessionStorage` during init → stores in `useGroupPlayStore` via `setParticipants`
3. `sessionStorage` entry cleared after read

**Files:**
- `dx-web/src/features/web/groups/components/group-game-room.tsx`
- `dx-web/src/features/web/play-group/components/group-play-loading-screen.tsx`

### 5. SSE event handling

Add `group_player_complete` listener in `use-group-play-events.ts`:

- New handler: `onPlayerComplete(event)` → calls `addCompletedPlayer(event.user_id)`
- Only processes events matching the current `game_level_id`

**Files:** `dx-web/src/features/web/play-group/hooks/use-group-play-events.ts`

### 6. Types

**New types in `group-play.ts`:**

```typescript
export type ParticipantMember = {
  user_id: string;
  user_name: string;
};

export type SoloParticipants = {
  mode: "group_solo";
  members: ParticipantMember[];
};

export type TeamParticipants = {
  mode: "group_team";
  teams: {
    subgroup_id: string;
    subgroup_name: string;
    members: ParticipantMember[];
  }[];
};

export type Participants = SoloParticipants | TeamParticipants;

export type GroupPlayerCompleteEvent = {
  user_id: string;
  user_name: string;
  game_level_id: string;
};
```

**Files:** `dx-web/src/features/web/play-group/types/group-play.ts`

### 7. UI rendering — `GroupPlayTopBar`

**Location:** Below the progress bar, above the stats section in the player panel (absolute-positioned, `w-56 md:w-64`).

**Solo mode layout:**

```
[progress bar]
─────────────────
👤 👤 👤✓ 👤 👤✓
─────────────────
[stats]
```

Horizontal flex-wrap row of `Avatar size="sm"` (24px). Each avatar:
- Background: `getAvatarColor(user_id)` from `@/lib/avatar`
- Fallback text: first character of `user_name`, uppercase, white, bold
- Completed: `AvatarBadge` with `bg-green-500` and `Check` icon (lucide, ~8px)
- Current player: `ring-2 ring-teal-500` to highlight "you"

**Team mode layout:**

```
[progress bar]
─────────────────
Team Alpha
👤 👤✓ 👤
Team Beta
👤✓ 👤 👤✓
─────────────────
[stats]
```

Each subgroup: small muted label (`text-[10px] text-muted-foreground font-medium`) + flex-wrap avatar row. Subgroups separated by `space-y-1.5`.

**Overflow:** `max-h-24 overflow-y-auto` to prevent panel from growing too tall with many members.

**Props change:** `GroupPlayTopBar` receives `playerId: string` (current user ID) to identify the current player for the ring highlight.

**Files:** `dx-web/src/features/web/play-group/components/group-play-top-bar.tsx`

### 8. Update `GroupGameStartEvent` type (frontend)

Add `participants` field to the existing `GroupGameStartEvent` type in `dx-web/src/features/web/groups/types/group.ts`.

**Files:** `dx-web/src/features/web/groups/types/group.ts`

## Documentation

Update `docs/game-lsrw-group-rule.md` if the new SSE event and member roster behavior warrant documentation.

## Files Changed Summary

| File | Change |
|------|--------|
| `dx-api/app/services/api/group_game_service.go` | Extend `GroupGameStartEvent` with participants |
| `dx-api/app/services/api/game_play_group_service.go` | Broadcast `group_player_complete` SSE event |
| `dx-web/src/features/web/play-group/types/group-play.ts` | New types |
| `dx-web/src/features/web/play-group/hooks/use-group-play-store.ts` | New state + actions |
| `dx-web/src/features/web/play-group/hooks/use-group-play-events.ts` | New SSE listener |
| `dx-web/src/features/web/play-group/components/group-play-top-bar.tsx` | Render member roster |
| `dx-web/src/features/web/play-group/components/group-play-shell.tsx` | Pass `playerId` prop |
| `dx-web/src/features/web/play-group/components/group-play-loading-screen.tsx` | Read participants from sessionStorage |
| `dx-web/src/features/web/groups/components/group-game-room.tsx` | Store participants in sessionStorage |
| `dx-web/src/features/web/groups/types/group.ts` | Update `GroupGameStartEvent` type |
| `docs/game-lsrw-group-rule.md` | Document new SSE event and roster behavior |

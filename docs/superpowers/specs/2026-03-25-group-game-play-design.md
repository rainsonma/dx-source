# Group Game Play — Connecting Groups to Real Game Play

**Date:** 2026-03-25
**Status:** Draft

## Overview

Connect the existing game group feature to actual game play. Groups currently set a `current_game_id` and `game_mode` (solo/team) but have no link to game sessions. This design adds synchronized, owner-initiated group play with per-answer time limits, winner tracking, and real-time SSE coordination.

## Game Modes

- **Solo** — individual competition within the group. Highest score wins.
- **Team** — subgroup-based competition. Highest sum of member scores wins for the subgroup.

## Schema Changes

### Alter `game_groups`

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `answer_time_limit` | INTEGER | 10 | Seconds per answer during group play (valid range: 5–60) |
| `is_playing` | BOOLEAN | false | Whether a group round is currently in progress |

### Alter `game_group_members`

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `last_won_at` | TIMESTAMP | NULL | Last time this member was on the winning side |

### Alter `game_subgroups`

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `last_won_at` | TIMESTAMP | NULL | Last time this subgroup won (team mode only) |

### Alter `game_session_totals`

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `game_group_id` | UUID (FK → game_groups) | NULL | Links session to group competition |
| `game_subgroup_id` | UUID (FK → game_subgroups) | NULL | Links session to subgroup (team mode only, NULL in solo) |

**Index:** `(game_group_id)` WHERE `game_group_id IS NOT NULL`

**Unique active session index change:** The existing unique index on `(user_id, game_id, degree, COALESCE(pattern, ''))` WHERE `ended_at IS NULL` must be extended to include `COALESCE(game_group_id, '')`. This allows a user to have both a regular solo session and a group session for the same game+degree+pattern simultaneously.

### Alter `game_session_levels`

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `game_group_id` | TEXT (FK → game_groups) | NULL | Links level session to group |
| `game_subgroup_id` | TEXT (FK → game_subgroups) | NULL | Links level session to subgroup (team mode only, NULL in solo) |

**Index:** `(game_group_id, game_level_id)` WHERE `game_group_id IS NOT NULL`

**Note:** `game_group_id` and `game_subgroup_id` on `game_session_levels` is intentional denormalization — duplicated from the parent `game_session_totals` for query efficiency in winner determination, avoiding a JOIN on every level completion check.

### Behavior Change

- Remove `ErrCannotLeaveOwned` guard in both `LeaveGroup` and `KickMember` — owner can leave or be removed (including self-removal via kick endpoint) like any member
- Owner is still auto-added as member on group creation (member_count starts at 1)
- Owner shows a remove button in 群成员 panel (calls the kick endpoint on self)
- Ownership does not transfer — if owner removes themselves, they remain the owner (can still manage the group, start games) but don't participate as a player
- `SetGroupGame` and `ClearGroupGame` must reject with error if `is_playing` is true — cannot change the game mid-round
- `DeleteGroup` must reject if `is_playing` is true. For ended sessions referencing a deleted group, FKs use `ON DELETE SET NULL`.

## SSE Infrastructure

The existing `SSEWriter` in `helpers/sse.go` is a request-scoped response writer. Group play requires a server-push model where the backend can send events to connected members at any time.

### SSE Connection Endpoint

`GET /api/groups/{id}/events` — members establish a persistent SSE connection.

- Authenticated via JWT (same as all protected endpoints)
- Server holds the connection open, sends heartbeat pings every 30s
- On disconnect, the connection is removed from the registry
- Client auto-reconnects using standard `EventSource` retry behavior

### Connection Registry

In-memory map: `map[groupID]map[userID]*SSEConnection`

- On connect: register `(groupID, userID) → connection`
- On disconnect: remove from map
- On event broadcast: iterate all connections for the group, write to each
- Thread-safe with `sync.RWMutex`

**Note:** In-memory registry is sufficient for single-server deployment. If scaling to multiple servers, replace with Redis pub/sub.

## Game Start Flow

1. Owner clicks "开始游戏" on group detail page (below 当前课程游戏)
2. Owner sees game mode card panel (same as game detail page) — picks degree + pattern
3. Backend validates:
   - User is group owner
   - Group has `current_game_id` set
   - `is_playing` is false (prevents duplicate starts)
4. Backend sets `is_playing = true` on the group
5. SSE pushes `group_game_start` event to all group members
6. Members' clients receive the event → navigate to game play loading screen
7. During loading screen, each member's client calls `POST /api/sessions/start` with additional `game_group_id` parameter
8. Backend creates `game_session_total` with `game_group_id` set. If team mode, backend resolves `game_subgroup_id` from `game_subgroup_members`. If team mode and user isn't in a subgroup → reject with error.
9. Each `game_session_level` inherits `game_group_id` and `game_subgroup_id` from its parent `game_session_total`

**Session creation with `game_group_id`:** When `game_group_id` is present in the `StartSession` request, the backend must NOT resume an existing non-group session. The `findActiveSession` lookup must include `game_group_id` in its filter, so group sessions and regular sessions are treated as distinct. The extended unique index ensures no conflicts.

**Level progression:** All members start at level 1 and play through all levels sequentially. The owner does not pick a specific level. Each level is a competition checkpoint — members wait between levels for results before advancing together.

**Owner participation:** Owner is also a member (unless they removed themselves). If they are a member, their client navigates to the loading screen and creates a session the same way. If not a member, they just initiate the round.

**Offline members:** Only members with an active SSE connection receive the event. No retry/queue — offline members miss the round.

## Answer Time Limit Enforcement

During group play, each answer has an `answer_time_limit` countdown (default 10s, range 5–60s):

- Frontend renders a countdown timer per question
- Timer expires without answer → client auto-calls `POST /api/sessions/{id}/skips` (treated as skip)
- Timer expires with incomplete answer → client auto-calls `POST /api/sessions/{id}/answers` with the incomplete input (scored normally — incorrect if wrong)
- The `answer_time_limit` is passed to the client via the SSE `group_game_start` event payload
- Regular (non-group) play is unaffected — no timer unless playing in a group context

## Winner Determination

### Trigger

After each member completes a level, check if they are the last one:

1. Count all participating members (active `game_session_totals` with this `game_group_id` where `ended_at IS NULL`)
2. Count completed level sessions for this `game_group_id` + `game_level_id` (where level `ended_at IS NOT NULL`)
3. If completed count == participant count → all done → determine winner

**Concurrency safety:** Use an atomic `UPDATE ... RETURNING` pattern — increment a completed counter (or use `SELECT COUNT(...) FOR UPDATE`) within a transaction to ensure exactly one request triggers winner determination when two members complete simultaneously.

This triggers in three scenarios:
1. All members finish within time — last member completes naturally
2. Time's up on the last answer — per-answer timer forces auto-skip/submit, level completes
3. Owner force-ends — all active `game_session_totals` AND their child `game_session_levels` with `ended_at IS NULL` are set to `ended_at = now()`, then winner calculated from whatever scores exist

### Solo Mode (`game_subgroup_id IS NULL`)

1. Query all `game_session_levels` WHERE `game_group_id = ? AND game_level_id = ? AND ended_at IS NOT NULL`
2. Rank by `score` descending
3. Highest score → update that member's `last_won_at` on `game_group_members`
4. Tie-break: earlier level `ended_at` wins

### Team Mode (`game_subgroup_id IS NOT NULL`)

1. Query all `game_session_levels` WHERE `game_group_id = ? AND game_level_id = ? AND ended_at IS NOT NULL`
2. Group by `game_subgroup_id`, sum scores per subgroup
3. Highest sum → set `last_won_at` on that `game_subgroups` record + all participating members of that subgroup (those with completed level sessions) get `last_won_at` on `game_group_members`
4. Tie-break: subgroup with the earliest MAX(`ended_at`) among its members wins (the subgroup whose last member finished first)

### `last_won_at` Semantics

- Updates every time a new level completes and a winner is determined
- Reflects the most recent win, not a history
- Resets implicitly by being overwritten with a newer timestamp

### Round End

When the last level of the game is completed (or force-ended), set `is_playing = false` on the group. The round is over.

## SSE Events

### `group_game_start`

Owner initiates a round → pushed to all group members.

```json
{
  "event": "group_game_start",
  "data": {
    "game_group_id": "uuid",
    "game_id": "uuid",
    "game_name": "string",
    "game_mode": "solo",
    "degree": "string",
    "pattern": "string",
    "answer_time_limit": 10
  }
}
```

### `group_level_complete`

Last member finishes a level (or force-ended) → pushed to all participants.

**Solo result:**
```json
{
  "event": "group_level_complete",
  "data": {
    "game_level_id": "uuid",
    "mode": "solo",
    "winner": {
      "user_id": "uuid",
      "user_name": "string",
      "score": 42
    }
  }
}
```

**Team result:**
```json
{
  "event": "group_level_complete",
  "data": {
    "game_level_id": "uuid",
    "mode": "team",
    "winner": {
      "subgroup_id": "uuid",
      "subgroup_name": "string",
      "total_score": 128,
      "members": [
        { "user_id": "uuid", "user_name": "string", "score": 45 },
        { "user_id": "uuid", "user_name": "string", "score": 42 },
        { "user_id": "uuid", "user_name": "string", "score": 41 }
      ]
    }
  }
}
```

### `group_game_force_end`

Owner force-ends → pushed to all participants with results for all completed levels.

```json
{
  "event": "group_game_force_end",
  "data": {
    "results": [
      { "game_level_id": "uuid", "mode": "solo", "winner": { ... } }
    ]
  }
}
```

## Client States During Group Play

| State | UI |
|-------|-----|
| Playing | Normal game UI with answer countdown timer |
| Finished level, others still playing | "等待其他选手完成..." waiting screen |
| All complete | Result panel with winner info (via SSE) |
| More levels remain | "下一关" button to advance |
| Last level complete | Final results screen |
| Force-ended by owner | Final results screen |

## New API Endpoints

### `GET /api/groups/{id}/events`

SSE connection endpoint for group members. Authenticated, persistent connection with 30s heartbeat.

### `POST /api/groups/{id}/start-game`

Owner starts a group game round.

**Request:**
```json
{
  "degree": "intermediate",
  "pattern": "write"
}
```

**Behavior:**
- Validates user is group owner
- Validates group has `current_game_id` and `game_mode` set
- Validates `is_playing` is false
- Sets `is_playing = true`
- Pushes `group_game_start` SSE to all members
- Returns success

### `POST /api/groups/{id}/force-end`

Owner force-ends the current round.

**Behavior:**
- Validates user is group owner
- Validates `is_playing` is true
- Ends all active `game_session_totals` with this `game_group_id` (set `ended_at = now()`)
- Ends all active child `game_session_levels` with this `game_group_id` (set `ended_at = now()`)
- Calculates winners for each level that has completed sessions
- Sets `is_playing = false`
- Pushes `group_game_force_end` SSE to all participants
- Returns results

## Non-Goals

- No round history table — active sessions define the round
- No level-wide time limit — per-answer timer naturally forces level completion
- No offline member catch-up — miss the SSE, miss the round
- No group leaderboard history — `last_won_at` is a simple "current champion" marker
- No ownership transfer — owner remains owner even if they leave as a member

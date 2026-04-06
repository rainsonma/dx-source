# PK Specified Opponent Design

Date: 2026-04-06
Status: Approved

## Overview

Add a "指定对手" (specified opponent) feature to the PK play mode, allowing users to invite a specific online user to a real-time PK match. This extends the existing robot-based PK system (now labeled "随机对手") with real player-vs-player gameplay.

## Scope

### In Scope

- Restyle 对手强度 selector to left-right layout matching 起始关卡
- Add ShadCN Tabs (随机对手 / 指定对手) to GameModeCard PK panel
- Redis online user set for presence tracking (reusable across features)
- Global user SSE channel (`UserSSEHub`) for real-time notifications
- Username verification endpoint (exists + online + VIP)
- PK invitation flow (invite / accept / decline)
- PK room waiting page for specified opponent
- Bottom-right slide-up invitation popup for the opponent
- Real PvP gameplay using the existing PK SSE and scoring infrastructure

### Out of Scope

- Changes to 随机对手 (random/robot) flow — remains unchanged
- Online user count display (future use of the Redis set)
- Chat or messaging between PK players
- Match history for specified PK

## Architecture

### Flow Overview

```
Initiator                          System                          Opponent
   |                                  |                               |
   | Click 双人PK                     |                               |
   | Select 指定对手 tab              |                               |
   | Enter username + click 验证      |                               |
   |--------------------------------->|                               |
   |   POST /api/users/verify-online  |                               |
   |   Check: exists + online + VIP   |                               |
   |<---------------------------------|                               |
   | See green check + nickname       |                               |
   |                                  |                               |
   | Select 起始关卡 + click 开始 PK  |                               |
   |--------------------------------->|                               |
   |   POST /api/play-pk/invite       |                               |
   |   Create game_pk (specified)     |                               |
   |   Create initiator session       |                               |
   |   Push pk_invitation SSE --------|------------------------------>|
   |<---------------------------------|                               |
   | Navigate to /hall/pk-room/{pkId} |                               |
   | Connect to PK SSE               |                               |
   | Show waiting UI (30s timeout)    |     See invitation popup      |
   |                                  |     (slide-up bottom-right)   |
   |                                  |                               |
   |                                  |           Click 接受          |
   |                                  |<------------------------------|
   |                                  |   POST /api/play-pk/invite/{pkId}/accept
   |                                  |   VIP re-check (safety net)   |
   |                                  |   Create opponent session     |
   |   pk_invitation_accepted SSE <---|                               |
   |                                  |------------------------------>|
   | Show opponent joined             |     Navigate to PK room       |
   | 1s countdown                     |     1s countdown              |
   | Redirect to play-pk page         |     Redirect to play-pk page  |
   |                                  |                               |
   | === Normal PK gameplay (existing SSE + scoring) ===              |
```

### Decline / Timeout Flow

```
Decline:
  Opponent clicks 拒绝
  → POST /api/play-pk/invite/{pkId}/decline
  → invitation_status = declined, is_playing = false, end initiator session
  → pk_invitation_declined SSE to initiator
  → Initiator sees "对方已拒绝" + 返回 button

Timeout (30s):
  → Popup auto-dismisses on opponent side
  → Room page shows "对方未响应" + 返回 button
  → Backend expires: invitation_status = expired, is_playing = false, end session
```

## Schema Changes

### game_pks table — add columns

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `pk_type` | text | `'random'` | `random` (robot) or `specified` (real PvP) |
| `invitation_status` | text | NULL | `pending` / `accepted` / `declined` / `expired` (NULL for random) |

Migration: Add columns with defaults so existing rows get `pk_type = 'random'`, `invitation_status = NULL`.

## Backend Components

### 1. Redis Online Set

- Key: `online_users` (Redis SET)
- `SADD online_users {userID}` — on global SSE connect
- `SREM online_users {userID}` — on global SSE disconnect
- `SISMEMBER online_users {userID}` — single user online check
- `SCARD online_users` — total online count (future use)

### 2. Global User SSE Hub (`sse_user_hub.go`)

```go
type UserSSEHub struct {
    mu    sync.RWMutex
    conns map[string]*SSEConnection // userID -> conn
}

var UserHub = &UserSSEHub{
    conns: make(map[string]*SSEConnection),
}
```

Methods:
- `Register(userID, w)` — SADD Redis, store conn, replace old conn
- `Unregister(userID, conn)` — SREM Redis, remove conn (pointer match)
- `SendToUser(userID, event, data)` — send SSE event to specific user
- `IsOnline(userID)` — check Redis SISMEMBER
- Heartbeat every 30s

### 3. Global SSE Endpoint

- `GET /api/user/events` — JWT protected
- Registers user in `UserSSEHub`
- Blocks until disconnect (same pattern as PkHub SSE endpoint)

### 4. Verify Online Endpoint

- `POST /api/users/verify-online`
- Request: `{ "username": string }`
- Logic:
  1. Find user by username — not found → 404 "用户不存在"
  2. Check cannot be self — 400 "不能挑战自己"
  3. `SISMEMBER online_users {userID}` — offline → 200 `{ is_online: false }` "对方不在线"
  4. `IsVipActive(userID)` — not VIP → 200 `{ is_online: true, is_vip: false }` "对方会员已过期"
  5. All pass → 200 `{ user_id, nickname, is_online: true, is_vip: true }`

### 5. Invitation Endpoints

**POST /api/play-pk/invite**

Request:
```json
{
  "game_id": "uuid",
  "game_level_id": "uuid",
  "degree": "beginner|intermediate|advanced",
  "pattern": "write|null",
  "opponent_id": "uuid"
}
```

Logic:
1. VIP check (initiator)
2. Verify opponent online + VIP (Redis + DB)
3. Check no existing active PK for initiator on this game
4. Create `game_pks` record: `pk_type=specified`, `invitation_status=pending`, `is_playing=true`
5. Create initiator's `game_session`
6. Push `pk_invitation` SSE event to opponent via `UserHub.SendToUser()`
7. Return `{ pk_id, session_id, game_level_id }`

`pk_invitation` SSE event payload:
```json
{
  "pk_id": "uuid",
  "game_id": "uuid",
  "game_name": "连词成句",
  "level_name": "第一关",
  "initiator_id": "uuid",
  "initiator_name": "rainson"
}
```

**POST /api/play-pk/invite/{pkId}/accept**

Logic:
1. Verify caller is opponent on this PK
2. VIP re-check (safety net for edge case)
3. Set `invitation_status = accepted`
4. Create opponent's `game_session`
5. Broadcast `pk_invitation_accepted` via `PkHub` to the room
6. Return `{ session_id, game_id, game_level_id, degree, pattern }`

**POST /api/play-pk/invite/{pkId}/decline**

Logic:
1. Verify caller is opponent on this PK
2. Set `invitation_status = declined`, `is_playing = false`
3. End initiator's session (`ended_at = now`)
4. Broadcast `pk_invitation_declined` via `PkHub` to the room
5. Return success

### 6. PK Service Changes

In `StartPk()` — no changes. Only called for `random` flow.

In `CompletePk()` — no changes needed. Already player-agnostic.

In `NextPkLevel()`:
- If original PK was `pk_type=specified`, new PK also gets `pk_type=specified`, `invitation_status=accepted`
- Skip robot spawning
- Both players transition via existing `pk_player_complete` SSE event

In `EndPk()` / `OnPkDisconnect()`:
- Works as-is for both random and specified

### 7. Invitation Expiry

When room timeout fires (30s), frontend calls `POST /api/play-pk/{pkId}/end`.
Backend: sets `invitation_status = expired`, `is_playing = false`, ends sessions.

No background job needed — expiry is client-driven (same as existing PK end flow).

## Frontend Components

### 1. GameModeCard UI Changes (`game-mode-card.tsx`)

**Restyle 对手强度**: Replace button-bar with left-right row (icon + label + Select dropdown), matching 起始关卡.

**Add Tabs** (below 游戏方式, PK mode only):
```
<Tabs defaultValue="random" onValueChange={setPkTab}>
  <TabsList>
    <TabsTrigger value="random">随机对手</TabsTrigger>
    <TabsTrigger value="specified">指定对手</TabsTrigger>
  </TabsList>
  <TabsContent value="random">
    对手强度 (icon + label + Select)
    起始关卡 (icon + label + Select)
  </TabsContent>
  <TabsContent value="specified">
    用户名 Input + 验证 Button (inline group)
    Verification result message
    起始关卡 (icon + label + Select)
  </TabsContent>
</Tabs>
```

**State management**:
- `pkTab: "random" | "specified"` — active tab
- `specifiedUsername: string` — input value
- `verifyResult: { userId, nickname, isOnline, isVip } | null` — verify response
- `verifyError: string | null` — error message
- Tab switch resets the other tab's state
- "开始 PK" disabled in specified tab until verify passes

**handlePkStart() changes**:
- If `pkTab === "random"`: existing flow (navigate to play-pk page)
- If `pkTab === "specified"`: call invite endpoint → navigate to `/hall/pk-room/{pkId}`

### 2. Verify Action (`session.action.ts`)

New action:
```typescript
export async function verifyOpponentAction(username: string) {
  // POST /api/users/verify-online { username }
  // Returns { user_id, nickname, is_online, is_vip } or error
}
```

### 3. Global User SSE Hook (`use-user-sse.ts`)

New hook in `dx-web/src/hooks/`:
```typescript
export function useUserSSE(onPkInvitation: (event: PkInvitationEvent) => void)
```

- Connects to `GET /api/user/events` with credentials
- Listens for `pk_invitation` event
- Returns cleanup function

### 4. PK Invitation Popup Component

New component in `dx-web/src/features/web/play-pk/components/pk-invitation-popup.tsx`:
- Slides up from bottom-right with animation
- Shows: initiator name, game name, level name
- Buttons: 接受 (green) / 拒绝 (muted)
- 接受 → calls accept action → navigate to PK room
- 拒绝 → calls decline action → dismiss
- Auto-dismiss after 30s

### 5. PK Invitation Provider

New provider in root `(web)/layout.tsx`:
- Wraps `useUserSSE()` + popup state management
- Manages popup show/hide lifecycle
- Only active for authenticated users

### 6. PK Room Page (`/hall/pk-room/[id]/page.tsx`)

New page:
- Fetches PK details on mount
- Connects to `/api/play-pk/{pkId}/events` SSE (reuses PkHub)
- UI: game info header, two player slots (initiator filled, opponent pulsing)
- SSE listeners:
  - `pk_invitation_accepted` → show opponent avatar → 1s countdown → redirect to play-pk
  - `pk_invitation_declined` → show "对方已拒绝" + 返回 button
- 30s timer → show "对方未响应" + 返回 button → call end PK
- 取消 button → call end PK → navigate back

### 7. Play Page Changes (`play-pk/[id]/page.tsx`)

- Accept new `pkId` and `sessionId` search params
- When `pkId` + `sessionId` present: skip `startPkAction()` in loading screen
- Loading screen uses `sessionId` directly, calls `restoreSessionDataAction()` + `fetchLevelContentAction()`
- Everything else unchanged

### 8. Session ID Propagation

Both players need their `session_id` to enter the play page. The PK room → play-pk redirect includes all params:

```
/hall/play-pk/{gameId}?degree=X&pattern=P&level=L&pkId=xxx&sessionId=yyy
```

- **Initiator**: gets `session_id` from `POST /api/play-pk/invite` response, stores in component state during room wait
- **Opponent**: gets `session_id` from `POST /api/play-pk/invite/{pkId}/accept` response, stores before navigating to room

The PK room page also needs a `GET /api/play-pk/{pkId}/details` endpoint (or reuse invite response data) so the opponent's room page can display game info + both player names without extra props drilling. This lightweight endpoint returns:

```json
{
  "pk_id": "uuid",
  "game_id": "uuid",
  "game_name": "string",
  "level_name": "string",
  "degree": "string",
  "initiator_id": "uuid",
  "initiator_name": "string",
  "opponent_id": "uuid",
  "opponent_name": "string",
  "invitation_status": "pending|accepted"
}
```

## API Route Summary

| Method | Route | Auth | New? |
|--------|-------|------|------|
| GET | `/api/user/events` | JWT | New — global SSE |
| POST | `/api/users/verify-online` | JWT | New — username verify |
| POST | `/api/play-pk/invite` | JWT | New — send invitation |
| POST | `/api/play-pk/invite/{pkId}/accept` | JWT | New — accept |
| POST | `/api/play-pk/invite/{pkId}/decline` | JWT | New �� decline |
| GET | `/api/play-pk/{pkId}/details` | JWT | New — PK room info |

All existing PK endpoints remain unchanged.

## SSE Events Summary

| Event | Hub | Direction | Payload |
|-------|-----|-----------|---------|
| `pk_invitation` | UserHub | Server → Opponent | `{ pk_id, game_id, game_name, level_name, initiator_id, initiator_name }` |
| `pk_invitation_accepted` | PkHub | Server → Initiator | `{ pk_id, opponent_id, opponent_name }` |
| `pk_invitation_declined` | PkHub | Server → Initiator | `{ pk_id }` |

Existing PK SSE events (`pk_player_action`, `pk_player_complete`, `pk_force_end`) unchanged.

## Constraints

- Both initiator and opponent must have active VIP
- Both checked at 验证 time; opponent re-checked at accept time (safety net)
- Only one active PK per user per game (existing unique constraint)
- Random opponent flow completely unchanged
- Robot is never spawned for specified PK
- 30-second invitation timeout, client-driven expiry

## File Changes Summary

### New Files

| File | Purpose |
|------|---------|
| `dx-api/app/helpers/sse_user_hub.go` | Global user SSE hub + Redis online set |
| `dx-api/app/http/controllers/api/user_sse_controller.go` | Global SSE endpoint |
| `dx-api/app/http/controllers/api/user_verify_controller.go` | Verify online endpoint |
| `dx-api/app/http/controllers/api/pk_invite_controller.go` | Invitation endpoints |
| `dx-api/app/services/api/pk_invite_service.go` | Invitation business logic |
| `dx-api/app/http/requests/api/pk_invite_request.go` | Invitation request validation |
| `dx-api/app/http/controllers/api/pk_detail_controller.go` | PK room details endpoint |
| `dx-api/database/migrations/2026040600000x_add_pk_type_columns.go` | Schema migration |
| `dx-web/src/hooks/use-user-sse.ts` | Global SSE hook |
| `dx-web/src/features/web/play-pk/components/pk-invitation-popup.tsx` | Invitation popup |
| `dx-web/src/features/web/play-pk/actions/invite.action.ts` | Invite/accept/decline actions |
| `dx-web/src/app/(web)/hall/pk-room/[id]/page.tsx` | PK room page |

### Modified Files

| File | Changes |
|------|---------|
| `dx-web/src/features/web/play-core/components/game-mode-card.tsx` | Tabs, restyle difficulty, verify input |
| `dx-web/src/app/(web)/layout.tsx` | Add UserSSE provider + invitation popup |
| `dx-web/src/app/(web)/hall/play-pk/[id]/page.tsx` | Handle pkId param, skip startPkAction |
| `dx-web/src/features/web/play-pk/components/pk-play-loading-screen.tsx` | Skip start when pkId provided |
| `dx-api/app/services/api/game_play_pk_service.go` | NextPkLevel: handle specified type |
| `dx-api/app/models/game_pk.go` | Add PkType, InvitationStatus fields |
| `dx-api/routes/api.go` | Add new routes |
| `dx-api/bootstrap/migrations.go` | Register new migration |

# LSRW Group Game Rules

## Overview

Group game play is a competitive feature where multiple members of a game group play the same game simultaneously under timed conditions. The group owner controls the game setup and start. Winners are determined per level by comparing scores.

## Group Setup

### Creating a Group

1. Any user can create a group via the "创建学习群" button on the groups list page
2. On creation:
   - Group is assigned a unique 8-character invite code
   - A QR code image is generated (non-blocking — fails gracefully)
   - The creator becomes the owner AND first member (`member_count = 1`)
   - `level_time_limit` defaults to 10 minutes
   - `is_playing` defaults to false
3. The owner can leave/be removed from their own group (ownership persists, but they won't participate as a player)

### Adding Members

Members can join a group through:

| Method | Flow |
|--------|------|
| Invite link (`/g/{code}`) | User visits link → clicks "Join" → `POST /api/groups/join/{code}` → auto-accepted |
| Application | User applies → `POST /api/groups/{id}/apply` → owner accepts/rejects in 待审批 panel |

- When accepted: `GameGroupMember` created, `member_count` incremented
- A user cannot be a member twice (unique constraint on `game_group_id + user_id`)
- Rejected users can re-apply

### Removing Members

- Owner can kick any member (including themselves) via the remove button in 群成员 panel
- Non-owners can leave via "退出群组" button
- Removing a member also removes them from all subgroups within the group
- `member_count` is decremented

### Subgroups (Team Mode Only)

- Owner can create subgroups with name, description, and display order
- Members are assigned to subgroups by the owner (select → "分配到小组" dropdown)
- A member can only be in one subgroup per group (auto-removed from previous on reassign)
- Subgroups cannot be modified while `is_playing = true`

## Game Configuration

### Setting the Current Game

Owner opens "设置群课程游戏" dialog from the group detail page:

1. **Search**: Find published games by name (debounced search, shows latest 3 by default)
2. **Select game**: Radio-button selection from search results
3. **Select starting level**: Dropdown of game levels (defaults to first level, shown as "起始关卡")
4. **Set level time limit**: 1-60 minutes per level (default 10, shown as "每关卡限时")
5. **Set game mode**: Solo/单人 (`group_solo`) or Team/小组 (`group_team`) toggle
6. **Confirm**: `PUT /api/groups/{id}/game` with `game_id`, `game_mode`, `level_time_limit`, `start_game_level_id`

Constraints:
- Owner only
- Cannot change game while `is_playing = true`
- Game must be published

### Clearing the Current Game

- Owner clicks "清除" button → confirms → `DELETE /api/groups/{id}/game`
- Clears `current_game_id`, `game_mode`, and `start_game_level_id` (sets to NULL)
- Cannot clear while `is_playing = true`

### Updating Level Time Limit

- Can be set during "设置群课程游戏" or via group update (编辑群组)
- Validation: integer, range 1-60 minutes
- Cannot be 0

## Game Room

### Entering the Game Room

- All members see a "进入教室" button on the group detail page — it is **always visible** regardless of game state
- Button has 3 states:

| Condition | Text | Style |
|-----------|------|-------|
| No game set | 教室未开放 | Disabled, muted |
| Game set, `is_playing = false` | 进入教室 | Active teal, navigates to `/hall/groups/{id}/room` |
| Game set, `is_playing = true`, owner | 进入教室（管理） | Active teal, navigates to room (owner needs force-end access) |
| Game set, `is_playing = true`, non-owner | 游戏中... | Disabled, muted |

### Room Access During Gameplay

- When `is_playing = true`, **non-owners are redirected** from the room page back to the group detail page
- The **owner retains room access** during gameplay for the "游戏中，强制结束" (force-end) button
- Non-owners cannot enter the room during gameplay from any path (button disabled + room page redirect)

### Game Room Page

Displays:
- Back button (CircleArrowLeft icon, top-left corner of card)
- Group name and member count
- Current game name with mode badge (单人/小组) and starting level name (if set)
- **Room member presence**: Avatars of all members currently in the room, showing "已进入教室（N/M）"
  - When all members are present, shows "全员到齐" label
- Waiting message:
  - All members present + owner: "全员到齐，可以开始"
  - All members present + non-owner: "等待群主开始"
  - Members missing: "等待成员进入教室..."
- **Owner controls**: "开始" (teal) button or "游戏中，强制结束" (red) button
  - "开始" button is **disabled** until all members are present (全员到齐)

### SSE Connection & Room Presence

- On mount, the game room establishes an `EventSource` connection to `GET /api/groups/{id}/events?token={JWT}`
- JWT is passed as query parameter (browser EventSource cannot set Authorization headers)
- Backend validates JWT via `ParseJWTUserID()`, verifies group membership
- Connection stays open with 30-second heartbeat pings
- On disconnect, connection is removed from the SSE hub (only if it matches the current connection — prevents race conditions during EventSource reconnects)
- **Disconnect during gameplay**: A disconnected player's session is **not** ended — they remain "in the game" as an inactive participant. Winner determination (`CheckAndDetermineWinner`) only counts **connected** players (via SSE hub) as participants, so disconnected players do not block remaining players from seeing results. The disconnected player gets no credit for the level. On disconnect, `RecheckGroupWinners` is called to re-run winner determination for all in-progress levels — this unblocks waiting players immediately when a leaver disconnects.
- **Room presence tracking**: The SSE hub's connection registry serves as the source of truth for who is in the room
  - When a member connects (enters the room): backend broadcasts `room_member_joined` SSE event to all connections
  - When a member disconnects (leaves the room): backend broadcasts `room_member_left` SSE event to remaining connections
  - Frontend fetches initial member list via `GET /api/groups/{id}/room-members` on mount
  - On join/leave events, frontend re-fetches the member list to update avatars
  - The `GET /api/groups/{id}/room-members` endpoint reads connected user IDs from the SSE hub and returns their names

### SSE Auth Resilience

Both SSE hooks (`useGroupEvents` for the game room, `useGroupPlayEvents` for gameplay) use a shared `useGroupSSE` hook that handles token expiration and reconnection:

- **Token refresh on auth failure**: When EventSource gets a 401 (expired JWT), the hook closes the connection, calls `refreshAccessToken()` to obtain a fresh access token via the `dx_refresh` cookie, then creates a new EventSource with the fresh token
- **Missing token on page refresh**: The access token is in-memory only. On page refresh, the hook detects a null token, calls `refreshAccessToken()` first, then connects
- **Exponential backoff**: Reconnection delays follow `min(1000 * 2^(n-1), 30000)` ms — 1s, 2s, 4s, 8s, 16s, 30s — to avoid hammering the server
- **Transient network failures**: If the token refresh call itself fails (e.g., WiFi drops briefly), the hook schedules another backoff retry instead of giving up — the redirect-to-signin path navigates away before the retry fires
- **Max retries**: After 10 consecutive failures, the hook stops retrying
- **Backoff reset**: On successful connection (`onopen`), the retry counter resets to 0
- **Backend safety**: `SSEHub.Register()` replaces existing connections for the same user (closes old, registers new), so reconnection never creates duplicate connections

### Entry Point

- Members enter the room via the "进入教室" button on the group detail page
- The button is always visible but only active (teal, clickable) when a game is set and (`is_playing = false` OR user is the owner)
- During gameplay, the owner sees "进入教室（管理）" to access the force-end button; non-owners see "游戏中..." (disabled)

## Starting the Game

### Owner Actions

1. Owner clicks "开始" in the game room (only enabled when all members are present)
2. "开始游戏" dialog opens with:
   - **Degree selection**: 初级 (beginner), 中级 (intermediate), 高级 (advanced)
     - No 练习 (practice) mode for group games
   - **Pattern selection** (required): 听 (listen), 说 (speak), 读 (read), 写 (write, default)
3. Owner clicks "开始游戏" button (teal)

### Backend Validation (`POST /api/groups/{id}/start-game`)

| Check | Error |
|-------|-------|
| User is group owner | "无权操作此学习群" |
| `is_playing` is false | "游戏正在进行中" |
| `current_game_id` is set | "未设置当前游戏" |
| `game_mode` is set | "未设置游戏模式" |
| ≥ 2 members | "至少需要2名成员才能开始竞赛" |
| Team mode: ≥ 2 subgroups with members | "至少需要2个小组才能开始小组竞赛" |
| Team mode: equal member count per subgroup | "每个小组的成员数量必须相等" |

### On Success

1. Backend ends any stale sessions from a previous round (active `game_session_levels` and `game_session_totals` for this group get `ended_at = NOW`) — this ensures each round starts with fresh sessions
2. Backend sets `is_playing = true` on the group
3. Backend broadcasts SSE event `group_game_start` to all group members:
   ```json
   {
     "game_group_id": "uuid",
     "game_id": "uuid",
     "game_name": "游戏名称",
     "game_mode": "group_solo",
     "degree": "intermediate",
     "pattern": "write",
     "level_time_limit": 10,
     "level_id": "uuid",
     "level_name": "Level 1",
     "participants": {
       "mode": "group_solo",
       "members": [
         { "user_id": "uuid", "user_name": "张三" },
         { "user_id": "uuid", "user_name": "李四" }
       ]
     }
   }
   ```
3. All members' clients (on the game room page) receive the event
4. Clients auto-navigate to `/hall/play-group/{gameId}?groupId={groupId}&degree={degree}&pattern={pattern}&levelTimeLimit={minutes}&gameMode={gameMode}&level={levelId}`

### Members Not in the Game Room

- Members who are not on the game room page do not receive the SSE event
- They miss the game — no retry or notification mechanism

## Game Play

### Loading Screen

Sequential loading steps (with progress bar 0-100%):

1. **Start session** (25%): `POST /api/play-group/start` with `game_id`, `degree`, `pattern`, `level_id`, `game_group_id`
   - Creates `GameSessionTotal` with `game_group_id` and `game_subgroup_id` (team mode)
   - For team mode: resolves subgroup membership. If user not in a subgroup → error "member not in any subgroup"
2. **Start level** (50%): `POST /api/play-group/{id}/levels/start` with `game_level_id`, `degree`, `pattern`
   - Creates `GameSessionLevel` inheriting `game_group_id` and `game_subgroup_id`
3. **Restore data** (75%): If resuming, fetches accumulated stats
4. **Fetch content** (100%): `GET /api/games/{gameId}/levels/{levelId}/content?degree={degree}`
   - Same shared content endpoint as single play

Minimum display time: 1.2 seconds

### Playing Phase

- Game mode component renders (GameLsrw, GameVocabMatch, etc.)
- Shared game components read state from `useGameStore` (play-core)
- Actions are injected via `GamePlayContext`:
  - `recordAnswer` → `POST /api/play-group/{id}/answers`
  - `recordSkip` → `POST /api/play-group/{id}/skips`
  - `markAsReview` → `POST /api/tracking/review` (shared endpoint)
  - `markAsMastered` → `POST /api/tracking/master` (shared endpoint)

### Scoring

Same scoring rules as single play:

| Event | Points |
|-------|--------|
| Correct answer | +1 base |
| 3-combo | +3 bonus |
| 5-combo | +5 bonus |
| 10-combo | +10 bonus |
| Wrong answer | 0 (combo resets) |
| Skip | 0 (combo resets, neutral) |

Combo cycle resets after 10 consecutive correct answers.

### Level Time Limit

- Displayed in the center of the top bar as "Group: MM:SS" with a Clock icon
- Counts down from `level_time_limit * 60` seconds
- Turns red when ≤ 60 seconds remaining
- When timer reaches 00:00:
  1. `setPhase("result")` is called
  2. `completeAndWait()` triggers level completion via API
  3. All players' timers expire simultaneously (same start time, same duration)

### Top Bar

| Element | Description |
|---------|-------------|
| Back button | Opens exit confirmation modal |
| Level name | Current level display |
| Center timer | "Group: MM:SS" countdown (teal, red at ≤60s) |
| Action buttons | Settings, Reset, Report, Fullscreen |
| Player panel | Expandable: avatar, score, combo, progress bar, **member roster with completion indicators**, stats |

### Member Roster (During Play)

- Displayed below the progress bar in the player panel
- **Solo mode**: Flat row of member avatars (ShadCN `Avatar size="sm"` with deterministic color from user ID)
- **Team mode**: Subgroup name labels with member avatars grouped underneath
- **Completion indicator**: Green checkmark badge (`AvatarBadge`) appears on a player's avatar when they complete the level
- **Current player highlight**: Ring-2 teal border on the current user's avatar
- **Data source**: Participant roster embedded in `group_game_start` SSE event, stored via sessionStorage across navigation
- **Live updates**: `group_player_complete` SSE event triggers badge appearance in real-time
- **Reset**: Completion indicators reset on level transition (`group_next_level`)

### Live Stats Feed (During Play)

- The stats section (跳过, 得分, 连击) displays a **real-time activity feed** of all group members' actions
- Each row shows the last player who performed that action type, with a whoosh animation bar
- **跳过**: Whooshes when any player skips, showing the player's name
- **得分**: Whooshes when any player answers correctly, showing the player's name
- **连击**: Whooshes when any player hits a combo bonus (3, 5, or 10 streak), showing "player ×N"
- Actions are broadcast via `group_player_action` SSE event from the backend after each answer/skip
- Wrong answers do not trigger any broadcast
- Latest action wins — rapid successive actions replace the previous display immediately
- Self-events (current player's own actions) are NOT filtered — all members see a consistent feed

### Play Time Tracking

- Active play time accumulates only during `phase=playing` with no overlay
- Synced to server with each answer/skip submission
- On tab close: `navigator.sendBeacon` flushes via `POST /api/play-group/{id}/sync-playtime`

## Level Completion

### When a Player Finishes

When all content items are answered/skipped, or when the timer expires:

1. Client calls `completeLevelAction(sessionId, levelId, { score, maxCombo, totalItems })` — retries once on failure to prevent the player from being stuck on the waiting screen
2. Backend `GroupPlayCompleteLevel`:
   - Marks `GameSessionLevel.ended_at = NOW`
   - Calculates accuracy (`correct_count / total_items`)
   - Awards EXP if accuracy ≥ 60% threshold (+10 EXP)
   - Updates session totals, level stats, game stats, user EXP
3. Backend calls `CheckAndDetermineWinner(groupID, levelID)` — errors are logged but do not fail the completion response

### Winner Determination

**Concurrency safety**: Uses `SELECT ... FOR UPDATE` to lock participant rows, preventing duplicate winner calculations.

**Completion check**:
1. Lock active sessions for this group (`game_session_totals WHERE ended_at IS NULL`)
2. Count only **connected** participants (cross-referenced with SSE hub) — disconnected players are ignored
3. **Deduplicate** connected participant IDs: a user with multiple active session totals is counted once so that `participantCount` matches the `COUNT(DISTINCT user_id)` used for completion counting
4. Count completed level sessions from connected players, scoped to active sessions (`gst.ended_at IS NULL`) — prevents stale completions from previous rounds being counted
5. If `completed < connected participants` → return nil (still waiting)
6. If all connected players done → determine winner (scoped to current round's active session IDs)
7. On SSE disconnect, `RecheckGroupWinners` re-runs this check for all in-progress levels (scoped to active sessions)

**Session scoping**: The completion count uses `gst.ended_at IS NULL` (via JOIN) to scope to the current round's active sessions. Winner determination queries filter by `game_session_total_id IN (active session IDs)`. This prevents old completed sessions from previous rounds polluting the completion count or winner scores when replaying the same game. Force-end collects session IDs before ending them so the scope is preserved.

**Participant deduplication**: Connected participant IDs are deduplicated with a seen-set before counting. This ensures `participantCount` matches the `COUNT(DISTINCT gst.user_id)` used in the completion query. Without this, a user with stale duplicate sessions would inflate the participant count, causing the winner check to never succeed.

**Winner query deduplication**: All winner queries use `DISTINCT ON (user_id)` to handle players with multiple completed level sessions (e.g., from restarts). Only the highest score per user is kept.

**Solo mode winner**:
- Query completed level sessions for this group + level, scoped to current round (deduplicated per user, best score)
- Rank by `score DESC`, tie-break by `ended_at ASC` (earlier finish wins)
- Winner's `game_group_members.last_won_at` is updated

**Team mode winner**:
- Sum best scores per user per subgroup, scoped to current round (deduplicated)
- Highest sum wins, tie-break by `MAX(ended_at) ASC` (team whose last member finished earliest wins)
- Winning subgroup's `game_subgroups.last_won_at` is updated
- All participating members of the winning subgroup get `game_group_members.last_won_at` updated

### SSE Broadcast

When winner is determined, backend broadcasts `group_level_complete` to all group members:

**Solo result**:
```json
{
  "game_level_id": "uuid",
  "mode": "group_solo",
  "winner": {
    "user_id": "uuid",
    "user_name": "张三",
    "score": 42
  },
  "participants": [
    { "user_id": "uuid", "user_name": "张三", "score": 42 },
    { "user_id": "uuid", "user_name": "李四", "score": 38 },
    { "user_id": "uuid", "user_name": "王五", "score": 35 }
  ]
}
```

**Team result**:
```json
{
  "game_level_id": "uuid",
  "mode": "group_team",
  "winner": {
    "subgroup_id": "uuid",
    "subgroup_name": "A组",
    "total_score": 128,
    "members": [
      { "user_id": "uuid", "user_name": "张三", "score": 45 },
      { "user_id": "uuid", "user_name": "李四", "score": 42 },
      { "user_id": "uuid", "user_name": "王五", "score": 41 }
    ]
  },
  "participants": [
    { "user_id": "uuid", "user_name": "张三", "score": 45 },
    { "user_id": "uuid", "user_name": "李四", "score": 42 },
    { "user_id": "uuid", "user_name": "王五", "score": 41 },
    { "user_id": "uuid", "user_name": "赵六", "score": 38 }
  ],
  "teams": [
    { "subgroup_id": "uuid", "subgroup_name": "A组", "total_score": 128, "members": [...] },
    { "subgroup_id": "uuid", "subgroup_name": "B组", "total_score": 112, "members": [...] }
  ]
}
```

### Auto-End Round

After winner broadcast, if this was the last level in the game:
- Backend counts active levels for the game; only proceeds if the count query succeeds and returns > 0
- Checks `played_levels_count + 1 >= total_levels` (uses pre-increment value + 1)
- If yes: sets `game_group.is_playing = false`
- If the count query fails, `is_playing` is left unchanged to avoid prematurely ending the round

## Client State During Game

| State | UI |
|-------|-----|
| Loading | Loading screen with progress bar |
| Playing | Game component with countdown timer in top bar |
| Result (waiting) | Player avatar + game info card with mode badge and level name, spinner with "好厉害！请耐心等待其他选手完成...", teal "返回" button |
| Result (received) | Podium result panel with level name subtitle — teal-themed stepped podium for top 3, ranked list for remaining, all participant avatars. Buttons: "下一关" (solid) + "返回" (outline) if more levels remain, or single "结束" button on last level |

### State Transitions

```
Loading → Playing (initSession)
Playing → Result (all items done OR timer expires)
Result → Waiting Screen (completeAndWait called via useEffect)
Waiting Screen → Result Panel (SSE group_level_complete received)
Result Panel → Group Detail (user clicks "返回" or "结束")
Result Panel → Next Level Loading (any participant clicks "下一关" → SSE group_next_level → all navigate)
```

### Next Level

When the result panel shows and there are more levels:
1. Any participant can click "下一关" button
2. Frontend calls `POST /api/groups/{id}/next-level` with `current_level_id`
3. Backend finds the next active level by `order`, uses Redis cache guard for idempotency
4. Backend broadcasts `group_next_level` SSE event with next level info
5. All participants' clients receive the event and navigate to the play-group page with the new level
6. This triggers a fresh loading screen → session start → play cycle

If no next level exists, the button shows "结束" instead (links back to group page).

## Force End

### Owner Force End

1. Owner clicks "游戏中，强制结束" (red button) in the game room
2. Backend `ForceEndGroupGame`:
   - Validates owner, validates `is_playing = true`
   - Collects active session IDs before ending (scopes winner queries to current round)
   - Ends all active `game_session_levels` (sets `ended_at = NOW`)
   - Ends all active `game_session_totals` (sets `ended_at = NOW`)
   - Collects distinct level IDs with completed sessions from the current round's sessions
   - Calls `DetermineWinnerForLevel(sessionIDs)` for each (skips participant count check — sessions already ended)
   - Sets `is_playing = false`
   - Broadcasts SSE event `group_game_force_end` with results array:
     ```json
     {
       "results": [
         { "game_level_id": "uuid", "mode": "group_solo", "winner": {...}, "participants": [...] }
       ]
     }
     ```
3. **Player-side behavior** (all players currently on the play-group page):
   - SSE hook (`useGroupPlayEvents`) receives `group_game_force_end` event
   - `onForceEnd` handler fires:
     - Takes the last level result from the results array
     - Sets it as the group result via `setGroupResult()`
     - Transitions phase to "result" via `setPhase("result")`
   - The game stops immediately — regardless of whether the player was mid-answer, waiting, or idle
   - The group result panel displays with the podium and "返回" button
   - Players who already finished and are on the waiting screen also transition to the result panel

### Group Dismissal

- Owner dismisses the group via "解散群组" button on the group detail page → `POST /api/groups/{id}/dismiss`
- If `is_playing = true`, the backend automatically force-ends the game first (broadcasts `group_game_force_end`), then sets `dismissed_at = NOW()` and broadcasts `group_dismissed`
- Dismissed groups return 404 on all endpoints — they no longer appear in the groups list
- Members in the game room or mid-gameplay receive the `group_dismissed` SSE event and are redirected to the groups list page
- This action is irreversible

## API Endpoints

### Group Management

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/groups/{id}/events?token=JWT` | SSE connection (query-param auth) |
| POST | `/api/groups/{id}/start-game` | Owner starts game round |
| POST | `/api/groups/{id}/force-end` | Owner force-ends game |
| GET | `/api/groups/{id}/room-members` | List members currently in the game room |
| PUT | `/api/groups/{id}/game` | Set current game + mode + time limit + start level |
| DELETE | `/api/groups/{id}/game` | Clear current game |
| POST | `/api/groups/{id}/next-level` | Any member triggers next level (SSE broadcast) |

### Group Play Session

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/play-group/start` | Start/resume session |
| POST | `/api/play-group/{id}/levels/start` | Start level |
| POST | `/api/play-group/{id}/levels/{levelId}/complete` | Complete level |
| POST | `/api/play-group/{id}/answers` | Record answer |
| POST | `/api/play-group/{id}/skips` | Record skip |
| POST | `/api/play-group/{id}/sync-playtime` | Sync play time |
| GET | `/api/play-group/{id}/restore` | Restore session data |
| PUT | `/api/play-group/{id}/content-item` | Update resume point |

### Shared Endpoints (Used by Both Single and Group Play)

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/games/{id}/levels/{levelId}/content?degree=X` | Fetch level content |
| POST | `/api/tracking/master` | Mark content as mastered |
| POST | `/api/tracking/unknown` | Mark content as unknown |
| POST | `/api/tracking/review` | Mark content for review |

## SSE Events

| Event | Trigger | Payload |
|-------|---------|---------|
| `group_game_start` | Owner starts game | Game info, degree, pattern, time limit, level_id, level_name, participants roster |
| `group_player_complete` | Individual player completes a level | `{ user_id, user_name, game_level_id }` |
| `group_player_action` | Player records correct answer, skip, or combo bonus | `{ user_id, user_name, action, combo_streak? }` |
| `group_level_complete` | All connected participants finish a level (or re-check after disconnect) | Winner result (solo or team) |
| `group_next_level` | Any participant triggers next level | game_group_id, game_id, level_id, level_name, degree, pattern, level_time_limit |
| `group_game_force_end` | Owner force-ends game | Array of level results |
| `group_dismissed` | Owner dismisses the group | `{ group_id }` |
| `room_member_joined` | Member enters game room (SSE connects) | `{ user_id }` |
| `room_member_left` | Member leaves game room (SSE disconnects) | `{ user_id }` |

## Database Schema (Group Play Fields)

### game_groups

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `level_time_limit` | INTEGER | 10 | Minutes per level (1-60) |
| `is_playing` | BOOLEAN | false | Whether a game round is in progress |
| `start_game_level_id` | VARCHAR | NULL | Starting level ID (set via 设置群课程游戏) |
| `dismissed_at` | TIMESTAMP | NULL | When the group was dismissed (NULL = active) |

### game_group_members

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `last_won_at` | TIMESTAMP | NULL | Last time this member won (solo or team) |
| `is_last_winner` | BOOLEAN | computed | Whether this member won the most recent group game (not stored, computed in query) |

### game_subgroups

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `last_won_at` | TIMESTAMP | NULL | Last time this subgroup won (team mode) |
| `is_last_winner` | BOOLEAN | computed | Whether this subgroup won the most recent group game (not stored, computed in query) |

### game_session_totals (additional columns)

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `game_group_id` | UUID (FK) | NULL | Links session to group |
| `game_subgroup_id` | UUID (FK) | NULL | Links session to subgroup (team mode, NULL in solo) |

### game_session_levels (additional columns)

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `game_group_id` | UUID (FK) | NULL | Links level session to group |
| `game_subgroup_id` | UUID (FK) | NULL | Links level session to subgroup (team mode, NULL in solo) |

### Custom Indexes

| Index | Table | Columns | Condition |
|-------|-------|---------|-----------|
| `idx_game_session_totals_unique_active_regular` | game_session_totals | user_id, game_id, degree, COALESCE(pattern, '') | ended_at IS NULL AND game_group_id IS NULL |
| `idx_game_session_totals_unique_active_group` | game_session_totals | user_id, game_id, degree, COALESCE(pattern, ''), game_group_id | ended_at IS NULL AND game_group_id IS NOT NULL |
| `idx_game_session_totals_group` | game_session_totals | game_group_id | game_group_id IS NOT NULL |
| `idx_game_session_levels_group_level` | game_session_levels | game_group_id, game_level_id | game_group_id IS NOT NULL |

## Winner Crown Badge

On the group detail page, a small amber crown icon (Lucide `Crown`, `h-3.5 w-3.5 text-amber-400`) is displayed next to the name of the latest winner:

- **群成员 list (Solo or Team mode)**: Members whose `last_won_at` equals the maximum `last_won_at` across all members in the group show a crown. In solo mode this is one member; in team mode this is all members of the winning subgroup (they share the same timestamp).
- **群小组 list (Team mode)**: The subgroup whose `last_won_at` equals the maximum `last_won_at` across all subgroups in the group shows a crown.

The `is_last_winner` boolean is computed server-side in the list API queries using:

```sql
CASE WHEN last_won_at IS NOT NULL
     AND last_won_at = (SELECT MAX(last_won_at) FROM <table> WHERE game_group_id = ?)
THEN true ELSE false END
```

If no member or subgroup has ever won (`last_won_at` is NULL for all), no crowns are displayed.

## Frontend Architecture

```
features/web/
├── play-core/           ← Shared game engine
│   ├── context/         ← GamePlayContext (dependency injection)
│   ├── components/      ← Game mode components, modals, shared UI
│   ├── hooks/           ← useGameStore, useLsrw, useGameTimer, etc.
│   ├── helpers/         ← Scoring logic
│   ├── actions/         ← Shared tracking/content actions
│   └── types/           ← Spelling types
│
├── play-single/         ← Solo game play
│   ├── components/      ← Shell, loading screen, top bar, result card
│   └── actions/         ← /api/play-single/* session actions
│
├── play-group/          ← Group game play (fully isolated)
│   ├── components/      ← Shell, loading screen, top bar, waiting, result
│   ├── hooks/           ← useGroupPlayStore, useGroupPlayEvents (SSE)
│   ├── actions/         ← /api/play-group/* session actions
│   └── types/           ← GroupLevelCompleteEvent, SoloWinner, TeamWinner
│
└── groups/              ← Group management (no game play logic)
    ├── components/      ← Group detail, game room, set-game dialog, etc.
    ├── hooks/           ← useGroupEvents (SSE for room)
    ├── actions/         ← Group API actions
    └── types/           ← Group types
```

### Dependency Injection (GamePlayContext)

Game mode components (GameLsrw, etc.) do NOT import API actions directly. Instead:

1. Each shell (play-single, play-group) wraps content with `<GamePlayProvider actions={...}>`
2. The provider supplies implementation-specific action functions
3. Game components call `useGamePlayActions()` to get the injected actions
4. This ensures group play components call `/api/play-group/*` endpoints, not `/api/play-single/*`

Injected actions:
- `recordAnswer` — record an answer
- `recordSkip` — record a skip
- `markAsReview` — mark content for review
- `completeLevel` — complete a level
- `endSession` — end a session
- `restartLevel` — restart a level

## Differences from Single Play

| Aspect | Single Play | Group Play |
|--------|------------|------------|
| Route | `/hall/play-single/{gameId}` | `/hall/play-group/{gameId}` |
| API prefix | `/api/play-single/*` | `/api/play-group/*` |
| Session creation | No group fields | Sets `game_group_id`, `game_subgroup_id` |
| Timer | Elapsed time (count up) | Level countdown (count down from limit) |
| Level completion | Shows result card | Waits for all connected participants, shows winner |
| Winner determination | N/A | Solo: highest score. Team: highest subgroup sum |
| SSE | N/A | group_game_start, group_level_complete, group_next_level, group_game_force_end |
| Entry point | Game detail page | Group game room (SSE-triggered) |
| Degree options | practice, beginner, intermediate, advanced | beginner, intermediate, advanced (no practice) |
| Pattern | Optional | Required (default: write) |
| EXP threshold | 60% accuracy | 60% accuracy (same) |

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
- Clears `current_game_id` and `game_mode` (sets to NULL)
- Cannot clear while `is_playing = true`

### Updating Level Time Limit

- Can be set during "设置群课程游戏" or via group update (编辑群组)
- Validation: integer, range 1-60 minutes
- Cannot be 0

## Game Room

### Entering the Game Room

- All members see "进入课程游戏" button (teal) on the group detail page below the current game card
- Button navigates to `/hall/groups/{id}/room`

### Game Room Page

Displays:
- Back button (CircleArrowLeft icon, top-left corner of card)
- Group name and member count
- Current game name with mode badge (单人/小组)
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
- On disconnect, connection is removed from the SSE hub
- **Room presence tracking**: The SSE hub's connection registry serves as the source of truth for who is in the room
  - When a member connects (enters the room): backend broadcasts `room_member_joined` SSE event to all connections
  - When a member disconnects (leaves the room): backend broadcasts `room_member_left` SSE event to remaining connections
  - Frontend fetches initial member list via `GET /api/groups/{id}/room-members` on mount
  - On join/leave events, frontend re-fetches the member list to update avatars
  - The `GET /api/groups/{id}/room-members` endpoint reads connected user IDs from the SSE hub and returns their names

### Entry Point

- Members enter the room via the "进入教室" button (teal) on the group detail page
- The button is visible below the current game card when a game is set

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

1. Backend sets `is_playing = true` on the group
2. Backend broadcasts SSE event `group_game_start` to all group members:
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
     "level_name": "Level 1"
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
| Player panel | Expandable: avatar, score, combo, progress bar, stats |

### Play Time Tracking

- Active play time accumulates only during `phase=playing` with no overlay
- Synced to server with each answer/skip submission
- On tab close: `navigator.sendBeacon` flushes via `POST /api/play-group/{id}/sync-playtime`

## Level Completion

### When a Player Finishes

When all content items are answered/skipped, or when the timer expires:

1. Client calls `completeLevelAction(sessionId, levelId, { score, maxCombo, totalItems })`
2. Backend `GroupPlayCompleteLevel`:
   - Marks `GameSessionLevel.ended_at = NOW`
   - Calculates accuracy (`correct_count / total_items`)
   - Awards EXP if accuracy ≥ 60% threshold (+10 EXP)
   - Updates session totals, level stats, game stats, user EXP
3. Backend calls `CheckAndDetermineWinner(groupID, levelID)`

### Winner Determination

**Concurrency safety**: Uses `SELECT ... FOR UPDATE` to lock participant rows, preventing duplicate winner calculations.

**Completion check**:
1. Count active sessions for this group (participants)
2. Count completed level sessions for this group + level
3. If `completed < participants` → return nil (still waiting)
4. If all done → determine winner

**Solo mode winner**:
- Query all completed level sessions for this group + level
- Rank by `score DESC`, tie-break by `ended_at ASC` (earlier finish wins)
- Winner's `game_group_members.last_won_at` is updated

**Team mode winner**:
- Sum scores per subgroup
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
- Backend checks `played_levels_count >= total_levels`
- If yes: sets `game_group.is_playing = false`

## Client State During Game

| State | UI |
|-------|-----|
| Loading | Loading screen with progress bar |
| Playing | Game component with countdown timer in top bar |
| Result (waiting) | Player avatar + game info card with mode badge, spinner with "好厉害！请耐心等待其他选手完成...", teal "返回" button |
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
   - Ends all active `game_session_levels` (sets `ended_at = NOW`)
   - Ends all active `game_session_totals` (sets `ended_at = NOW`)
   - Collects all distinct level IDs with completed sessions
   - Calls `DetermineWinnerForLevel()` for each (skips participant count check — sessions already ended)
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

### Group Deletion While Playing

- Cannot delete a group while `is_playing = true`
- Error: "game is in progress"

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
| `group_game_start` | Owner starts game | Game info, degree, pattern, time limit, level_id, level_name |
| `group_level_complete` | All participants finish a level | Winner result (solo or team) |
| `group_next_level` | Any participant triggers next level | game_group_id, game_id, level_id, level_name, degree, pattern, level_time_limit |
| `group_game_force_end` | Owner force-ends game | Array of level results |
| `room_member_joined` | Member enters game room (SSE connects) | `{ user_id }` |
| `room_member_left` | Member leaves game room (SSE disconnects) | `{ user_id }` |

## Database Schema (Group Play Fields)

### game_groups

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `level_time_limit` | INTEGER | 10 | Minutes per level (1-60) |
| `is_playing` | BOOLEAN | false | Whether a game round is in progress |
| `start_game_level_id` | VARCHAR | NULL | Starting level ID (set via 设置群课程游戏) |

### game_group_members

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `last_won_at` | TIMESTAMP | NULL | Last time this member won (solo or team) |

### game_subgroups

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `last_won_at` | TIMESTAMP | NULL | Last time this subgroup won (team mode) |

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
| Level completion | Shows result card | Waits for all participants, shows winner |
| Winner determination | N/A | Solo: highest score. Team: highest subgroup sum |
| SSE | N/A | group_game_start, group_level_complete, group_game_force_end |
| Entry point | Game detail page | Group game room (SSE-triggered) |
| Degree options | practice, beginner, intermediate, advanced | beginner, intermediate, advanced (no practice) |
| Pattern | Optional | Required (default: write) |
| EXP threshold | 60% accuracy | 60% accuracy (same) |

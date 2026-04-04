# PK Robot Mode Design

PK mode lets users compete against a robot opponent in real-time. The robot is a randomly selected mock user (`is_mock = true`) that simulates gameplay via a backend goroutine. The user doesn't know it's a robot — it looks and behaves like a real opponent. All play data is recorded normally, appearing on stats and leaderboards.

## Data Model

### New table: `game_pks`

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID V7 (PK) | Match identifier |
| `user_id` | string (FK users) | Real user who initiated |
| `opponent_id` | string (FK users) | Idle mock user, randomly selected |
| `game_id` | string (FK games) | Game being played |
| `degree` | string | beginner / intermediate / advanced |
| `pattern` | string (nullable) | listen / speak / read / write |
| `robot_difficulty` | string | easy / normal / hard |
| `current_level_id` | string (nullable, FK game_levels) | Tracks level progression |
| `is_playing` | bool | Match in progress |
| `last_winner_id` | string (nullable, FK users) | Last level's winner |
| `created_at` | timestamp | |
| `updated_at` | timestamp | |

### New columns on existing tables

- `game_session_totals.game_pk_id` — nullable, FK to `game_pks`
- `game_session_levels.game_pk_id` — nullable, FK to `game_pks`

Both the human's and robot's sessions link to the same `game_pks` record via `game_pk_id`, mirroring the `game_group_id` pattern.

## Constants

### Difficulty presets (`consts/pk.go`)

| Param | Easy | Normal | Hard |
|-------|------|--------|------|
| Accuracy | 50-70% | 70-85% | 85-95% |
| Min answer delay | 3s | 2s | 1s |
| Max answer delay | 6s | 4s | 3s |
| Combo break chance | 50% | 30% | 10% |

Constants: `PkDifficultyEasy = "easy"`, `PkDifficultyNormal = "normal"`, `PkDifficultyHard = "hard"`.

## Opponent Selection

1. Query `users WHERE is_mock = true` excluding any with an active PK (`game_pks.is_playing = true AND opponent_id = user.id`)
2. Pick one randomly
3. If none available, auto-create a new mock user:
   - `is_mock = true`
   - Username: random English first name (lowercase), e.g. `james`, `emma`, `sarah`
   - Nickname: random mix of English name, Chinese name, or hybrid, with `-` and `_` mixed in, e.g. `Emma_Li`, `David-Chen`, `小明-wang`, `Lily_张`, `小红`
   - Random avatar (deterministic from user ID)
4. Use the selected/created mock user as the opponent

A PK match holds one robot for its entire duration. The robot is freed when `game_pks.is_playing` is set to `false`.

## Backend Architecture

### New files

| File | Purpose |
|------|---------|
| `consts/pk.go` | Difficulty presets with accuracy, delay, combo params |
| `models/game_pk.go` | GamePk model struct |
| `controllers/api/game_play_pk_controller.go` | PK endpoints |
| `services/api/game_play_pk_service.go` | PK session lifecycle + robot goroutine |
| `services/api/pk_winner_service.go` | Per-level winner determination (always 2 players) |
| `services/api/mock_user_service.go` | Find idle mock user or auto-create |
| `helpers/sse_pk_hub.go` | SSE hub scoped to PK matches |
| `requests/api/pk_request.go` | Request validation structs |

### Routes (`routes/api.go`)

All routes require user JWT + VIP guard (same as group play).

```
POST   /api/play-pk/start                        — Start PK match
POST   /api/play-pk/{id}/levels/start             — Start level
POST   /api/play-pk/{id}/levels/{levelId}/complete — Complete level
POST   /api/play-pk/{id}/answers                  — Record answer
POST   /api/play-pk/{id}/skips                    — Record skip
POST   /api/play-pk/{id}/sync-playtime            — Sync playtime
GET    /api/play-pk/{id}/restore                  — Restore session state
PUT    /api/play-pk/{id}/content-item             — Update current content item
POST   /api/play-pk/{id}/end                      — End PK match
POST   /api/play-pk/{id}/next-level               — Advance to next level
POST   /api/play-pk/{id}/pause                    — Pause robot
POST   /api/play-pk/{id}/resume                   — Resume robot
GET    /api/play-pk/{id}/events                   — SSE connection
```

### SSE Hub (`helpers/sse_pk_hub.go`)

Lightweight copy of `GroupSSEHub`, scoped to PK matches. Max 2 connections per match (human + optional monitoring). Same `Register`, `Unregister`, `Broadcast`, `ConnectedUserIDs`, `SendHeartbeat` methods.

When the human's SSE connection drops:
1. Cancel the robot goroutine context
2. Auto-end both sessions
3. Set `game_pks.is_playing = false`
4. Robot is freed

### SSE Events

| Event | When | Data |
|-------|------|------|
| `pk_player_action` | Robot scores / combos / skips | `{ user_id, user_name, action, combo_streak? }` |
| `pk_player_complete` | Robot finishes level | `{ user_id, user_name, game_level_id }` |
| `pk_level_complete` | Both finished, winner determined | `{ game_level_id, winner, participants }` |
| `pk_force_end` | Match force-ended | `{ results }` |
| `pk_next_level` | Next level triggered | `{ level_id, level_name, degree, pattern }` |
| `pk_timeout_warning` | 4m30s after robot finishes level | `{ countdown: 30 }` |
| `pk_timeout` | 5m00s after robot finishes level | Auto-end level, robot wins |
| `pk_pause` | Human paused | Robot goroutine pauses |
| `pk_resume` | Human resumed | Robot goroutine resumes |

## Robot Goroutine

### Lifecycle per level

```
PK start / next-level
  → create robot's game_session_totals + game_session_levels (game_pk_id linked)
  → fetch content items for level
  → spawn goroutine with cancellable context

Goroutine loop:
  for each content item:
    1. Check pause channel — block if paused
    2. select:
       - ctx.Done() → cleanup, exit
       - time.After(randomDelay) → continue
    3. Decide correct/wrong (roll against accuracy threshold)
    4. If on combo streak, roll combo break chance
    5. Write game_records entry (same table as real players)
    6. Update robot's game_session_levels stats (score, combo, counts)
    7. Update robot's game_session_totals stats
    8. Broadcast pk_player_action SSE event

  Complete robot's level session (set ended_at)
  Broadcast pk_player_complete

  Wait phase:
    select:
    - ctx.Done() → exit (human finished or disconnected)
    - time.After(4m30s) → broadcast pk_timeout_warning
      then select:
      - ctx.Done() → exit
      - time.After(30s) → broadcast pk_timeout, auto-end level, robot wins
```

### Realistic behavior

- Variable timing with jitter (not perfectly uniform)
- Occasional wrong answers even on Hard difficulty
- Combo streaks break naturally via combo break chance
- Longer words/sentences get slightly longer delay
- Scoring follows same `consts/scoring.go` logic (base score + combo bonus)

### Pause / Resume

Robot goroutine checks a pause channel between answers. When paused, it blocks on the channel until resumed. The 5-minute timeout timer also pauses.

### Cancellation triggers

| Trigger | What happens |
|---------|-------------|
| Human completes + winner determined | Context cancelled, goroutine exits |
| Human SSE disconnects | Context cancelled, auto-end PK match |
| 5-minute timeout | Auto-end current level, robot wins |
| Human clicks 退出游戏 | Context cancelled, match ends |

## Winner Determination

### Per-level (simplified — always 2 players)

No need for `FOR UPDATE` locking or connected-player counting.

| Scenario | Winner |
|----------|--------|
| Both completed | Higher score wins; earlier `ended_at` breaks ties |
| Only robot completed (timeout) | Robot wins |
| Only human completed (robot cancelled) | Human wins |

### After winner determined

1. Update `game_pks.last_winner_id` with winner's user ID
2. Broadcast `pk_level_complete` with winner info and both scores

### Stats recording

- Both players' sessions write to `game_session_totals`, `game_session_levels`, `game_records`
- Both contribute to `game_stats_totals` and `game_stats_levels`
- Robot's stats appear on leaderboards naturally (mock user looks real)
- Human earns EXP normally (10 EXP per level if accuracy >= 60%)

## PK Match End Conditions

| Condition | Result |
|-----------|--------|
| Human clicks 退出游戏 | `is_playing = false`, cancel robot goroutine, navigate to game details |
| All levels completed | `is_playing = false` on last level completion |
| Human SSE disconnects | Auto-end, `is_playing = false`, robot freed |
| Timeout on any level | Auto-end that level (robot wins), match continues if more levels |

## Frontend Architecture

### New files

| File | Purpose |
|------|---------|
| `app/(web)/hall/play-pk/[id]/page.tsx` | Route page |
| `features/web/play-pk/components/pk-play-shell.tsx` | Main PK shell (mirrors group-play-shell) |
| `features/web/play-pk/components/pk-play-top-bar.tsx` | Top bar with pause, settings, exit |
| `features/web/play-pk/components/pk-play-loading-screen.tsx` | Loading screen with opponent info |
| `features/web/play-pk/components/pk-play-result-panel.tsx` | Result panel with podium (2 players) |
| `features/web/play-pk/components/pk-play-waiting-screen.tsx` | Waiting for robot to finish |
| `features/web/play-pk/hooks/use-pk-play-store.ts` | Zustand store (mirrors group play store) |
| `features/web/play-pk/hooks/use-pk-play-events.ts` | SSE event listeners for PK |
| `features/web/play-pk/actions/session.action.ts` | Server actions for PK endpoints |
| `features/web/play-pk/types/pk-play.ts` | TypeScript types for PK events |

### Modified files

| File | Change |
|------|--------|
| `features/web/games/components/hero-card.tsx` | Add PK button before 群组 button |
| `features/web/games/components/game-mode-card.tsx` | Add difficulty selector (only shown when opened in PK mode, not for regular single play) |

### Button order on HeroCard

```
[ 开始游戏 ]  [ PK ]  [ 群组 ]  [ 收藏 ]
```

### PK button flow

1. Click PK on HeroCard → opens `GameModeCard` with degree + pattern + difficulty selector (Easy / Normal / Hard)
2. Select options → navigate to `/hall/play-pk/{gameId}?degree=X&pattern=Y&difficulty=Z`
3. `PkPlayShell` mounts → calls `startPkAction()` → backend creates match, spawns robot goroutine
4. Same game component renders (GameWordSentence, etc.) — shared from `play-core`
5. SSE shows robot's live progress (score/combo via `pk_player_action` events)
6. Human finishes → waiting screen (if robot still playing) or result panel
7. If robot finishes first → human keeps playing, sees robot completed notification
8. At 4m30s after robot finishes → timeout countdown modal: "你将在 30 秒后输掉本场 PK" with live countdown
9. At 5m00s → auto-end level, robot wins
10. Result panel: podium with 2 players, "下一关" / "结束" buttons

### PkPlayTopBar

- Pause button → POST `/api/play-pk/{id}/pause` → robot pauses, overlay shown
- Resume → POST `/api/play-pk/{id}/resume` → robot resumes, overlay dismissed
- 退出游戏 → confirm modal → POST `/api/play-pk/{id}/end` → match ends, navigate to game details

### Reused from existing code (no changes)

- All game mode components (`play-core/components/game-*.tsx`)
- `useGameStore` core state
- `GamePlayProvider` context (record answer, skip, complete level)
- Scoring helpers (`play-core/helpers/scoring.ts`)
- Timer, fullscreen, settings hooks
- `SpellingInputRow` and other input components

## Scope

- Available for all game modes (word-sentence, vocab-battle, vocab-match, vocab-elimination, listening-challenge)
- Requires VIP (same guard as group play)
- No changes to existing single play or group play functionality

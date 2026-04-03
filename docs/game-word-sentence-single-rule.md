# Word-Sentence Game Rules

## Scoring

### Base Score

Each correct answer earns **1 point**.

### Combo Bonuses

Consecutive correct answers trigger combo bonuses:

| Combo Streak | Bonus |
|-------------|-------|
| 3 | +3 |
| 5 | +5 |
| 10 | +10 |

- A wrong answer or skip resets the combo counter to 0
- After reaching 10-combo, the cycle resets and starts from 3 again
- Bonuses trigger once at each threshold per cycle

### Combo Example

20 consecutive correct answers:

```
Cycle 1 (answers 1-10):
  10 x 1 = 10 (base)
  + 3 (3-combo) + 5 (5-combo) + 10 (10-combo) = 18 bonus
  Cycle subtotal: 28

Cycle 2 (answers 11-20):
  10 x 1 = 10 (base)
  + 3 + 5 + 10 = 18 bonus
  Cycle subtotal: 28

Total: 56 points
```

### Skip Behavior

- Skipping an item (Tab key or 跳过 button) is treated as neutral
- Skipped items do NOT create a GameRecord or UserReview entry
- Skip count is tracked on GameSessionLevel and GameSession
- Skipping resets the combo counter but does not count as a wrong answer
- Skipped items reduce accuracy (counted in the denominator, not the numerator)

## Experience Points (EXP)

### Earning EXP

| Event | EXP | Condition |
|-------|-----|-----------|
| Complete a level with 60%+ accuracy | +10 | Each time threshold is met |

- EXP is granted every time a level is completed with 60%+ accuracy (correct / total items in the level)
- Replaying a completed level earns EXP again as long as 60% accuracy is achieved
- EXP accumulates on `User.exp` in real-time during gameplay

## Game Session Model

### Structure

- **GameSessionTotal**: One record per play attempt of a game at a specific `degree + pattern` combination
- **GameSessionLevel**: One record per level attempt within a session
- **GameRecord**: Individual answer record, unique per `gameSessionLevelId + contentItemId`

### Key Rules

- Each `degree + pattern` combination corresponds to one active `GameSessionTotal` per game per user
- Each level attempt within a session corresponds to one `GameSessionLevel`
- A user may have **multiple active game sessions simultaneously** (different degrees and/or patterns)
- Level sessions may be left active (orphaned) when switching to a different level — users can return to them later
- All play records are stored per level session, even on replay (the same content item can appear in multiple level sessions)
- Resume point within a level is tracked on `GameSessionLevel.currentContentItemId`
- Current level is tracked on `GameSessionTotal.currentLevelId`

### Session Lifecycle

| State | Description |
|-------|-------------|
| Active (`endedAt = null`) | User is playing or can resume |
| Ended (`endedAt` set) | Session finished (completed or force-ended) |

- A `GameSessionTotal` ends when the user completes the last level, or when a new session is needed for the same `degree + pattern`
- A `GameSessionLevel` ends when the user completes the level, or when the user restarts/replays the level (a new level session is created)
- Restarting a level only ends the **level session**, not the total session — cumulative stats in the total session are preserved

## Game Start/Resume Strategy

### Degree + Pattern (Game Mode Selection)

- **Degree**: practice, beginner, intermediate, advanced
- **Pattern** (Word-Sentence only): listen, speak, read, write (default: write)
- The degree determines which content types are included in the level
- The degree + pattern combination determines which session to create or resume

### Session Scope

- All active session checks (`CheckAnyActiveSession`, `CheckActiveSession`, `CheckActiveLevelSession`) and `StartSession` exclude group sessions by filtering `game_group_id IS NULL`
- Group game sessions are never resumable from the game detail page — they are managed exclusively through the group play flow
- The hall session progress list also excludes group sessions

### Entry Point 1: Hero Section Button (no level specified)

**Hero button label** (server-side, resolved at page load):

| Condition | Button Label |
|-----------|-------------|
| No active session for any degree+pattern | 开始游戏 |
| All sessions ended | 开始游戏 |
| Active session exists | 继续学习「{currentLevelId level name}」 |

**Behavior** (always opens GameModeCard modal):

| Condition | Action |
|-----------|--------|
| No active session for this degree+pattern | Create new session, start from first level |
| All sessions for this degree+pattern ended | Create new session, start from first level |
| Active session exists for this degree+pattern | Resume at session's `currentLevelId` |
| User clicks 重新开始 on game mode card | End current level session, create new level session for the **same** level (total session survives) |

When an active session exists, the GameModeCard modal pre-selects the session's degree+pattern. Users can change the selection, which triggers a re-check for the newly selected degree+pattern combination.

**Game mode card buttons:** Show 继续游戏/重新开始 only if an active **level session** exists for the game session currentLevelId degree+pattern. Otherwise show 开始游戏.

### Entry Point 2: Level Grid Button (specific level selected)

| Condition | Action                                                                                                                          |
|-----------|---------------------------------------------------------------------------------------------------------------------------------|
| No active session for this degree+pattern | Create new session, start from selected level                                                                                   |
| All sessions for this degree+pattern ended | Create new session, start from selected level                                                                                   |
| Active session + no level session for selected level | Create new level session within existing session                                                                                |
| Active session + all level sessions for selected level ended | Create new level session from selected level within existing session                                                            |
| Active session + active level session for selected level | Resume it                                                                                                                       |
| User clicks 重新开始 on game mode card | End current level session, create new level session for **selected** level (total session survives, cumulative stats preserved) |

**Game mode card buttons:** Show 继续游戏/重新开始 only if an active **level session** exists for the selected level within the active game session. Otherwise show 开始游戏 (which creates a new level session within the existing session if one exists).

### Entry Point 3: 再来一局 (replay same level from result panel)

| Condition | Action |
|-----------|--------|
| No active session (safety fallback) | Create new session, start from same level |
| All sessions ended (e.g., was last level) | Create new session, start from same level |
| Active session (non-last level) | Create new level session for same level within existing session |

### Entry Point 4: 下一关 (next level from result panel)

| Condition | Action |
|-----------|--------|
| Last level in the game | Hide the 下一关 button |
| No active session (safety fallback) | Create new session, start from next level |
| All sessions ended | Create new session, start from next level |
| Active session | Create new level session for next level within existing session |

### Entry Point 5: Page Refresh During Playing

Resume the same level session from the saved resume point (`GameSessionLevel.currentContentItemId`).

### Entry Point 6: Page Refresh During Result Panel

| Condition | Action |
|-----------|--------|
| No active session (safety fallback) | Create new session, start from same level |
| All sessions ended | Create new session, start from same level |
| Active session | Create new level session for same level within existing session |

### Entry Point 7: In-Game Reset Button (重置 in game top bar)

| Condition | Action |
|-----------|--------|
| Active session + active level session | Show reset confirmation modal → end current level session, create new level session for same level |

Flow: Click 重置 button → GameResetModal opens → user clicks 确认 → `restartLevelSessionAction` ends current level session → client state resets → navigate to play page → loading screen → fresh start. The total session survives with cumulative stats preserved.

### Consistency Rule

All content displayed across the game detail page, game loading screen, game playing page, and game level result page must be consistent — the level shown, the degree, the pattern, and the session state must all match.

## Progress Tracking

- **Per session** (`GameSessionTotal`): score, exp earned, max combo, correct/wrong/skip counts, play time, total levels count (snapshot at creation), played levels count (incremented on each level completion)
- **Per level session** (`GameSessionLevel`): score, exp, max combo, correct/wrong/skip counts, resume point (`currentContentItemId`), degree, pattern, play time, total items count (snapshot at creation, filtered by degree), played items count (incremented on each answer, not on skip)
- **Per record** (`GameRecord`): linked to `gameSessionLevelId`, stores individual answer per content item per level session attempt
  - `duration`: wall-clock seconds from item display to final word submission (rounded to nearest second)
  - Duration includes time spent on wrong attempts (retries)
  - Duration does NOT pause when overlays are shown (unlike session-level `playTime`)
  - Server clamps duration to 0–3600 seconds
  - Skipped items have no `GameRecord`, so no duration is tracked
- **Per level lifetime** (`GameStatsLevel`): highest score, total scores, total play time, completion count — updated on each level completion
- **Per game lifetime** (`GameStatsTotal`): highest score, total scores, total sessions, total EXP, total play time, completion count — `totalPlayTime`, `totalScores`, `totalExp`, `highestScore` updated per level completion; `totalSessions` updated on session creation; `completionCount` updated on session end
- **Per user global**: total EXP (`User.exp`)

## Play Time Tracking

- Active play time is tracked per session (`GameSessionTotal.playTime`) and per level (`GameSessionLevel.playTime`)
- Play time accumulates only while `phase=playing` and no overlay is shown (pauses excluded)
- Play time is synced to the database with each answer/skip submission
- On tab close, `navigator.sendBeacon` flushes the latest play time
- On resume, play time is restored from the server
- At level/session completion, play time feeds into `GameStatsLevel.totalPlayTime` and `GameStatsTotal.totalPlayTime`

## Level Access Control

### VIP Gating

- **Level 1** is free for all users (free, paid, and expired)
- **Levels 2+** require an active VIP membership

### VIP Definition

A user is "active VIP" if:

- `grade == "lifetime"` (never expires), OR
- `grade != "free"` AND `vipDueAt` is not null AND `vipDueAt > now()`

### Enforcement

- **Backend**: `StartSession`, `StartLevel`, `AdvanceLevel`, and `GetLevelContent` all check VIP status when the target level is not the first level. Returns error code `40302` (`CodeVipRequired`) with HTTP 403 if the user is not VIP.
- **Frontend**: The `LevelGrid` component shows lock icons on levels 2+ for non-VIP users. Clicking a locked level opens an upgrade dialog directing to `/purchase/membership`. The play page redirects non-VIP users back to the game detail page if they attempt to access a non-first level via URL.

### First Level Determination

The "first level" is the active level with the lowest `order` value for the game. This is queried as: `SELECT ... FROM game_levels WHERE game_id = ? AND is_active = true ORDER BY "order" ASC LIMIT 1`.

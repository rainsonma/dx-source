# VIP Level Gating Design

## Overview

Gate game levels behind VIP membership. Level 1 is free for all users. Levels 2+ require active VIP. Group features (except list view) require VIP entirely.

## VIP Status Logic

A user is "active VIP" if:

- `grade == "lifetime"` (never expires), OR
- `grade != "free"` AND `vipDueAt != nil` AND `vipDueAt > now()`

Everything else is non-VIP (free grade or expired membership).

## Backend Changes

### New Files

**`dx-api/app/services/api/vip_service.go`**:

- `IsVipActive(userID string) (bool, error)` — fetches user by ID, applies VIP logic
- `isFirstLevel(query, gameID, levelID string) (bool, error)` — checks if a level is the first active level of a game (lowest `order`)

### Modified Files

**`dx-api/app/services/api/errors.go`**:

- Add `ErrVipRequired = errors.New("升级会员解锁此功能")`

**`dx-api/app/consts/error_code.go`**:

- Add `CodeVipRequired = 40302`

**`dx-api/app/services/api/game_play_single_service.go`**:

- `StartSession`: If `levelID` is provided and is not the first level, call `IsVipActive`. Return `ErrVipRequired` if not VIP.
- `StartLevel`: If `gameLevelID` is not the first level, require VIP.
- `AdvanceLevel`: If `nextLevelID` is not the first level, require VIP.

**`dx-api/app/services/api/content_service.go`**:

- `GetLevelContent`: If the level is not the first level of its game, require VIP.

**`dx-api/app/services/api/group_service.go`**:

- `CreateGroup`, `GetGroupDetail`, `UpdateGroup`, `DismissGroup`: Require VIP.
- `ListGroups`, `GetGroupByInviteCode`: No guard (viewable by all).

**`dx-api/app/services/api/group_member_service.go`**:

- `JoinByCode`, `KickMember`, `LeaveGroup`, `ListGroupMembers`: Require VIP.

**`dx-api/app/services/api/group_game_service.go`**:

- `SetGroupGame`, `ClearGroupGame`, `StartGroupGame`, `ForceEndGroupGame`, `NextGroupLevel`, `SearchGamesForGroup`: Require VIP.

**`dx-api/app/services/api/game_play_group_service.go`**:

- All functions require VIP: `GroupPlayStartSession`, `GroupPlayStartLevel`, `GroupPlayCompleteLevel`, `GroupPlayRecordAnswer`, `GroupPlayRecordSkip`, `GroupPlaySyncPlayTime`, `GroupPlayRestoreSessionData`, `GroupPlayUpdateContentItem`.

**Controllers** — map `ErrVipRequired` to HTTP response:

- `game_play_single_controller.go`: Add to `mapSessionError` and `Start`/`StartLevel`/`AdvanceLevel`.
- `game_play_group_controller.go`: Add to all handlers.
- `group_controller.go`, `group_member_controller.go`, `group_game_controller.go`: Add mapping.
- `content_controller.go`: Add mapping.
- All map to: `HTTP 403`, code `40302`, message `"升级会员解锁此功能"`.

## Frontend Changes

### New Files

**`dx-web/src/lib/vip.ts`**:

```ts
export function isVipActive(grade: UserGrade, vipDueAt: string | null): boolean
```

Pure function. Same logic as backend.

**`dx-web/src/features/web/games/components/upgrade-dialog.tsx`**:

- shadcn `AlertDialog` with teal theme
- Props: `open`, `onOpenChange`, `message` (customizable description)
- Title: "解锁全部关卡" (or contextual)
- Description: configurable via `message` prop
- Cancel: "稍后再说"
- Action (teal bg): "立即升级" → navigates to `/purchase/membership`

### Modified Files

**`dx-web/src/features/web/games/components/level-grid.tsx`**:

- Add `isVip: boolean` prop to `LevelGrid`
- If `isVip`: all levels render as `"current"` (teal border, clickable)
- If not VIP: level 1 is `"current"`, rest are `"locked"` (existing logic with `completedLevels: 0`)
- Locked levels get `onClick` that opens `UpgradeDialog`

**`dx-web/src/features/web/games/components/game-detail-content.tsx`**:

- Add `isVip: boolean` prop, pass through to `LevelGrid`

**`dx-web/src/app/(web)/hall/(main)/games/[id]/page.tsx`**:

- Fetch user profile in existing `Promise.all` to get `grade` + `vip_due_at`
- Compute `isVip` via `isVipActive()`
- Pass `isVip` to `GameDetailContent`

**`dx-web/src/app/(web)/hall/play-single/[id]/page.tsx`**:

- Fetch user profile, compute `isVip`
- If not VIP and requested `level` param is not the first level: redirect to game detail page with toast "升级会员解锁全部关卡"

**Group pages** (groups list, group detail, invite page):

- Fetch user profile, compute `isVip`
- "创建学习群" button: opens `UpgradeDialog` if not VIP
- Group card click: opens `UpgradeDialog` instead of navigating if not VIP
- Invite page join button: opens `UpgradeDialog` if not VIP
- Group-specific dialog message: "升级会员即可创建和加入学习群，与同学一起学习"

## Documentation Updates

**`docs/game-word-sentence-single-rule.md`** — new section "Level Access Control":

- Level 1 free for all users
- Levels 2+ require active VIP membership
- VIP definition: `grade == "lifetime"` OR (`grade != "free"` AND `vipDueAt > now`)
- Backend error code `40302` for unauthorized access
- Frontend shows upgrade dialog on locked level click

**`docs/game-word-sentence-group-rule.md`** — new section "VIP Requirement":

- All group operations require active VIP (create, join, play, etc.)
- Only groups list page and invite info are viewable by free users
- Backend error code `40302` for unauthorized group actions
- Frontend shows upgrade dialog on blocked actions

## Files Summary

### Backend (dx-api)

| File | Change |
|------|--------|
| `app/services/api/vip_service.go` | NEW — `IsVipActive`, `isFirstLevel` |
| `app/services/api/errors.go` | Add `ErrVipRequired` |
| `app/consts/error_code.go` | Add `CodeVipRequired = 40302` |
| `app/services/api/game_play_single_service.go` | VIP check in `StartSession`, `StartLevel`, `AdvanceLevel` |
| `app/services/api/game_play_group_service.go` | VIP check in all 8 functions |
| `app/services/api/group_service.go` | VIP check in `CreateGroup`, `GetGroupDetail`, `UpdateGroup`, `DismissGroup` |
| `app/services/api/group_member_service.go` | VIP check in `JoinByCode`, `KickMember`, `LeaveGroup`, `ListGroupMembers` |
| `app/services/api/group_game_service.go` | VIP check in all 6 functions |
| `app/services/api/content_service.go` | VIP check in `GetLevelContent` |
| `app/http/controllers/api/game_play_single_controller.go` | Map `ErrVipRequired` → 403/40302 |
| `app/http/controllers/api/game_play_group_controller.go` | Map `ErrVipRequired` → 403/40302 |
| `app/http/controllers/api/group_controller.go` | Map `ErrVipRequired` → 403/40302 |
| `app/http/controllers/api/group_member_controller.go` | Map `ErrVipRequired` → 403/40302 |
| `app/http/controllers/api/group_game_controller.go` | Map `ErrVipRequired` → 403/40302 |
| `app/http/controllers/api/content_controller.go` | Map `ErrVipRequired` → 403/40302 |

### Frontend (dx-web)

| File | Change |
|------|--------|
| `src/lib/vip.ts` | NEW — `isVipActive()` |
| `src/features/web/games/components/upgrade-dialog.tsx` | NEW — teal AlertDialog |
| `src/features/web/games/components/level-grid.tsx` | Add `isVip` prop, locked level onClick |
| `src/features/web/games/components/game-detail-content.tsx` | Pass `isVip` through |
| `src/app/(web)/hall/(main)/games/[id]/page.tsx` | Fetch profile, compute `isVip` |
| `src/app/(web)/hall/play-single/[id]/page.tsx` | VIP guard redirect |
| Group list/detail/invite pages | VIP gating on create, join, navigate |

### Docs

| File | Change |
|------|--------|
| `docs/game-word-sentence-single-rule.md` | Add "Level Access Control" section |
| `docs/game-word-sentence-group-rule.md` | Add "VIP Requirement" section |

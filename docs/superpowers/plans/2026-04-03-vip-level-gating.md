# VIP Level Gating Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Gate game levels 2+ behind VIP membership, lock free/expired users out of group features, with backend enforcement and frontend UX.

**Architecture:** Add `IsVipActive` helper + `isFirstLevel` check in Go service layer. Guard all game-play and group endpoints. Frontend computes VIP status client-side, shows teal upgrade AlertDialog on locked actions.

**Tech Stack:** Go/Goravel (backend), Next.js/React/shadcn (frontend), PostgreSQL

**Spec:** `docs/superpowers/specs/2026-04-03-vip-level-gating-design.md`

---

### Task 1: Backend — Error sentinel and error code

**Files:**
- Modify: `dx-api/app/services/api/errors.go:59`
- Modify: `dx-api/app/consts/error_code.go:32`

- [ ] **Step 1: Add error sentinel**

In `dx-api/app/services/api/errors.go`, add after `ErrInvalidProduct`:

```go
ErrVipRequired = errors.New("升级会员解锁此功能")
```

- [ ] **Step 2: Add error code**

In `dx-api/app/consts/error_code.go`, add after `CodeGroupForbidden = 40301`:

```go
CodeVipRequired = 40302
```

- [ ] **Step 3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/services/api/errors.go dx-api/app/consts/error_code.go
git commit -m "feat: add ErrVipRequired sentinel and CodeVipRequired error code"
```

---

### Task 2: Backend — VIP service helper

**Files:**
- Create: `dx-api/app/services/api/vip_service.go`

- [ ] **Step 1: Create vip_service.go**

```go
package api

import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

// IsVipActive checks whether the user has an active VIP membership.
// Returns true if grade is "lifetime", or grade is not "free" and vipDueAt > now.
func IsVipActive(userID string) (bool, error) {
	var user models.User
	if err := facades.Orm().Query().Select("id", "grade", "vip_due_at").
		Where("id", userID).First(&user); err != nil {
		return false, fmt.Errorf("failed to find user: %w", err)
	}
	if user.ID == "" {
		return false, ErrUserNotFound
	}
	return checkVipActive(user), nil
}

// checkVipActive applies VIP logic to a loaded user.
func checkVipActive(user models.User) bool {
	if user.Grade == consts.UserGradeLifetime {
		return true
	}
	if user.Grade == consts.UserGradeFree {
		return false
	}
	return user.VipDueAt != nil && user.VipDueAt.StdTime().After(time.Now())
}

// isFirstLevel checks whether the given levelID is the first active level of a game.
func isFirstLevel(query orm.Query, gameID, levelID string) (bool, error) {
	var first models.GameLevel
	if err := query.Where("game_id", gameID).Where("is_active", true).
		Order("\"order\" asc").First(&first); err != nil || first.ID == "" {
		return false, ErrNoGameLevels
	}
	return first.ID == levelID, nil
}

// requireVipForLevel checks VIP status if the level is not the first level.
// Returns nil if the user can access the level, ErrVipRequired otherwise.
func requireVipForLevel(userID, gameID, levelID string) error {
	query := facades.Orm().Query()
	first, err := isFirstLevel(query, gameID, levelID)
	if err != nil {
		return err
	}
	if first {
		return nil
	}
	vip, err := IsVipActive(userID)
	if err != nil {
		return fmt.Errorf("failed to check VIP status: %w", err)
	}
	if !vip {
		return ErrVipRequired
	}
	return nil
}

// requireVip checks VIP status unconditionally.
// Returns nil if the user is VIP, ErrVipRequired otherwise.
func requireVip(userID string) error {
	vip, err := IsVipActive(userID)
	if err != nil {
		return fmt.Errorf("failed to check VIP status: %w", err)
	}
	if !vip {
		return ErrVipRequired
	}
	return nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/vip_service.go
git commit -m "feat: add VIP service helpers (IsVipActive, isFirstLevel, requireVipForLevel, requireVip)"
```

---

### Task 3: Backend — Guard single-player game play endpoints

**Files:**
- Modify: `dx-api/app/services/api/game_play_single_service.go:133,557,730`
- Modify: `dx-api/app/http/controllers/api/game_play_single_controller.go:24,94,145,396`

- [ ] **Step 1: Guard StartSession**

In `dx-api/app/services/api/game_play_single_service.go`, function `StartSession` (line 133), add VIP check right after finding the first level (after line 141, before checking existing active session):

```go
// VIP guard: non-first levels require active VIP
targetLevelID := levelID
if targetLevelID == nil {
	targetLevelID = &firstLevel.ID
}
if *targetLevelID != firstLevel.ID {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
}
```

- [ ] **Step 2: Guard StartLevel**

In `StartLevel` (line 557), add VIP check after `verifyOwnership` and before `UpsertLevelStats`. Need to look up the game ID from the session:

```go
// VIP guard: non-first levels require active VIP
var session models.GameSessionTotal
if err := facades.Orm().Query().Select("id", "game_id").Where("id", sessionID).First(&session); err != nil || session.ID == "" {
	return nil, ErrSessionNotFound
}
if err := requireVipForLevel(userID, session.GameID, gameLevelID); err != nil {
	return nil, err
}
```

- [ ] **Step 3: Guard AdvanceLevel**

In `AdvanceLevel` (line 730), add VIP check after `verifyOwnership`:

```go
// VIP guard: non-first levels require active VIP
var session models.GameSessionTotal
if err := facades.Orm().Query().Select("id", "game_id").Where("id", sessionID).First(&session); err != nil || session.ID == "" {
	return ErrSessionNotFound
}
if err := requireVipForLevel(userID, session.GameID, nextLevelID); err != nil {
	return err
}
```

- [ ] **Step 4: Add ErrVipRequired mapping to mapSessionError**

In `dx-api/app/http/controllers/api/game_play_single_controller.go`, function `mapSessionError` (line 396), add before the default case:

```go
if errors.Is(err, services.ErrVipRequired) {
	return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
}
```

- [ ] **Step 5: Add ErrVipRequired mapping to Start handler**

In the `Start` method (line 24), the error switch after `services.StartSession` call (line 36), add a new case before the internal error fallback:

```go
if errors.Is(err, services.ErrVipRequired) {
	return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
}
```

- [ ] **Step 6: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 7: Commit**

```bash
git add dx-api/app/services/api/game_play_single_service.go dx-api/app/http/controllers/api/game_play_single_controller.go
git commit -m "feat: add VIP guard to single-player game play (StartSession, StartLevel, AdvanceLevel)"
```

---

### Task 4: Backend — Guard content endpoint

**Files:**
- Modify: `dx-api/app/services/api/content_service.go:27`
- Modify: `dx-api/app/http/controllers/api/content_controller.go:18`

- [ ] **Step 1: Guard GetLevelContent**

In `dx-api/app/services/api/content_service.go`, change `GetLevelContent` signature to accept `userID`:

```go
func GetLevelContent(userID, gameLevelID string, degree string) ([]ContentItemData, error) {
```

Add VIP check at the top of the function, before the query:

```go
// VIP guard: non-first levels require active VIP
var level models.GameLevel
if err := facades.Orm().Query().Select("id", "game_id").Where("id", gameLevelID).First(&level); err != nil || level.ID == "" {
	return nil, ErrLevelNotFound
}
if err := requireVipForLevel(userID, level.GameID, gameLevelID); err != nil {
	return nil, err
}
```

- [ ] **Step 2: Update controller to pass userID**

In `dx-api/app/http/controllers/api/content_controller.go`, the `LevelContent` method already extracts `userID`. Change the service call (line 34) to pass it:

```go
items, err := services.GetLevelContent(userID, levelID, req.Degree)
```

Add error mapping after the service call:

```go
if err != nil {
	if errors.Is(err, services.ErrVipRequired) {
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
	}
	if errors.Is(err, services.ErrLevelNotFound) {
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "关卡不存在")
	}
	return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get level content")
}
```

Add `"errors"` to the import block.

- [ ] **Step 3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/services/api/content_service.go dx-api/app/http/controllers/api/content_controller.go
git commit -m "feat: add VIP guard to GetLevelContent endpoint"
```

---

### Task 5: Backend — Guard group services

**Files:**
- Modify: `dx-api/app/services/api/group_service.go`
- Modify: `dx-api/app/services/api/group_member_service.go`
- Modify: `dx-api/app/services/api/group_game_service.go`
- Modify: `dx-api/app/services/api/game_play_group_service.go`

- [ ] **Step 1: Guard group_service.go**

Add `if err := requireVip(userID); err != nil { return ..., err }` as the first line in each of these functions (adjust return types accordingly):

- `CreateGroup` (returns `*CreateGroupResult, error`) — add after function signature, before existing logic:
  ```go
  if err := requireVip(userID); err != nil {
  	return nil, err
  }
  ```
- `GetGroupDetail` (returns `*GroupDetail, error`) — same pattern
- `UpdateGroup` (returns `error`) — add:
  ```go
  if err := requireVip(userID); err != nil {
  	return err
  }
  ```
- `DismissGroup` (returns `error`) — same pattern

Do NOT guard `ListGroups` or `GetGroupByInviteCode`.

- [ ] **Step 2: Guard group_member_service.go**

Add `requireVip(userID)` as first line in:

- `JoinByCode` (returns `string, error`):
  ```go
  if err := requireVip(userID); err != nil {
  	return "", err
  }
  ```
- `KickMember` (returns `error`):
  ```go
  if err := requireVip(userID); err != nil {
  	return err
  }
  ```
- `LeaveGroup` (returns `error`) — same as KickMember
- `ListGroupMembers` (returns `[]MemberItem, string, bool, error`):
  ```go
  if err := requireVip(userID); err != nil {
  	return nil, "", false, err
  }
  ```

- [ ] **Step 3: Guard group_game_service.go**

Add `requireVip(userID)` as first line in:

- `SearchGamesForGroup` — this function doesn't take `userID`. The controller does ownership verification before calling it. Add VIP check in the controller `SearchGames` method instead (see Task 6).
- `SetGroupGame` (returns `error`):
  ```go
  if err := requireVip(userID); err != nil {
  	return err
  }
  ```
- `ClearGroupGame` (returns `error`) — same
- `StartGroupGame` (returns `error`) — same
- `ForceEndGroupGame` (returns `[]LevelWinnerResult, error`):
  ```go
  if err := requireVip(userID); err != nil {
  	return nil, err
  }
  ```
- `NextGroupLevel` (returns `error`) — same as SetGroupGame

- [ ] **Step 4: Guard game_play_group_service.go**

Add `requireVip(userID)` as first line in all 8 functions:

- `GroupPlayStartSession` (returns `*GroupPlayStartSessionResult, error`):
  ```go
  if err := requireVip(userID); err != nil {
  	return nil, err
  }
  ```
- `GroupPlayStartLevel` — same pattern (returns `*GroupPlayStartLevelResult, error`)
- `GroupPlayCompleteLevel` — same pattern (returns `*GroupPlayCompleteLevelResult, error`)
- `GroupPlayRecordAnswer` (returns `error`):
  ```go
  if err := requireVip(userID); err != nil {
  	return err
  }
  ```
- `GroupPlayRecordSkip` — same as RecordAnswer
- `GroupPlaySyncPlayTime` — same as RecordAnswer
- `GroupPlayRestoreSessionData` (returns `*GroupPlayRestoreSessionResult, error`):
  ```go
  if err := requireVip(userID); err != nil {
  	return nil, err
  }
  ```
- `GroupPlayUpdateContentItem` (returns `error`) — same as RecordAnswer

- [ ] **Step 5: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/services/api/group_service.go dx-api/app/services/api/group_member_service.go dx-api/app/services/api/group_game_service.go dx-api/app/services/api/game_play_group_service.go
git commit -m "feat: add VIP guard to all group and group-play services"
```

---

### Task 6: Backend — Guard group controllers (error mapping)

**Files:**
- Modify: `dx-api/app/http/controllers/api/group_controller.go:210`
- Modify: `dx-api/app/http/controllers/api/group_member_controller.go`
- Modify: `dx-api/app/http/controllers/api/group_game_controller.go:25,251`
- Modify: `dx-api/app/http/controllers/api/game_play_group_controller.go:247`

- [ ] **Step 1: Add to mapGroupError**

In `dx-api/app/http/controllers/api/group_controller.go`, `mapGroupError` (line 210), add a new case before the default:

```go
case errors.Is(err, services.ErrVipRequired):
	return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
```

- [ ] **Step 2: Add to mapGroupGameError**

In `dx-api/app/http/controllers/api/group_game_controller.go`, `mapGroupGameError` (line 251), add before default:

```go
case errors.Is(err, services.ErrVipRequired):
	return helpers.Error(ctx, nethttp.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
```

Also add VIP check in the `SearchGames` controller method (line 25). After ownership verification (line 38) and before the service call:

```go
if err := services.RequireVipExported(userID); err != nil {
	return mapGroupGameError(ctx, err)
}
```

Wait — `requireVip` is unexported. We need to either export it or handle it differently. Since the controller already uses `mapGroupGameError` and the service returns the error, the simplest approach: add VIP check in the `SearchGamesForGroup` service itself. But it doesn't take `userID`.

Better approach: add `userID` param to `SearchGamesForGroup`:

In `dx-api/app/services/api/group_game_service.go`, change signature:

```go
func SearchGamesForGroup(userID, query string, limit int) ([]GroupGameSearchItem, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
```

In `dx-api/app/http/controllers/api/group_game_controller.go` `SearchGames` method, update the call (line 42):

```go
items, err := services.SearchGamesForGroup(userID, q, limit)
if err != nil {
	return mapGroupGameError(ctx, err)
}
```

Remove the separate error handling lines 43-45 (old error handling) since `mapGroupGameError` now covers all errors.

- [ ] **Step 3: Add to mapGroupPlayError**

In `dx-api/app/http/controllers/api/game_play_group_controller.go`, `mapGroupPlayError` (line 247), add before the default return:

```go
if errors.Is(err, services.ErrVipRequired) {
	return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
}
```

- [ ] **Step 4: Add to group_member_controller.go**

The group member controller uses `mapGroupError` from `group_controller.go` (already updated in Step 1). Verify that all methods in `group_member_controller.go` use `mapGroupError` for error mapping. If any use inline error handling, add:

```go
if errors.Is(err, services.ErrVipRequired) {
	return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
}
```

- [ ] **Step 5: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 6: Run go vet**

Run: `cd dx-api && go vet ./...`
Expected: No errors

- [ ] **Step 7: Commit**

```bash
git add dx-api/app/http/controllers/api/group_controller.go dx-api/app/http/controllers/api/group_game_controller.go dx-api/app/http/controllers/api/game_play_group_controller.go dx-api/app/http/controllers/api/group_member_controller.go dx-api/app/services/api/group_game_service.go
git commit -m "feat: add VIP error mapping to all group controllers"
```

---

### Task 7: Frontend — VIP helper

**Files:**
- Create: `dx-web/src/lib/vip.ts`

- [ ] **Step 1: Create vip.ts**

```ts
import type { UserGrade } from "@/consts/user-grade";

/**
 * Check whether the user has an active VIP membership.
 * VIP = lifetime grade, or paid grade with vipDueAt in the future.
 */
export function isVipActive(grade: UserGrade, vipDueAt: string | null): boolean {
  if (grade === "lifetime") return true;
  if (grade === "free") return false;
  if (!vipDueAt) return false;
  return new Date(vipDueAt) > new Date();
}
```

- [ ] **Step 2: Verify lint**

Run: `cd dx-web && npx eslint src/lib/vip.ts`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/lib/vip.ts
git commit -m "feat: add isVipActive frontend helper"
```

---

### Task 8: Frontend — UpgradeDialog component

**Files:**
- Create: `dx-web/src/features/web/games/components/upgrade-dialog.tsx`

- [ ] **Step 1: Create upgrade-dialog.tsx**

```tsx
"use client";

import { useRouter } from "next/navigation";
import { Crown } from "lucide-react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

interface UpgradeDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: string;
  message?: string;
}

export function UpgradeDialog({
  open,
  onOpenChange,
  title = "解锁全部关卡",
  message = "升级会员即可畅玩所有关卡，享受完整学习体验",
}: UpgradeDialogProps) {
  const router = useRouter();

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <div className="flex items-center gap-2">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-teal-600/10">
              <Crown className="h-[18px] w-[18px] text-teal-600" />
            </div>
            <AlertDialogTitle className="text-lg">{title}</AlertDialogTitle>
          </div>
          <AlertDialogDescription className="text-[13px] leading-relaxed">
            {message}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>稍后再说</AlertDialogCancel>
          <AlertDialogAction
            className="bg-teal-600 hover:bg-teal-700"
            onClick={() => router.push("/purchase/membership")}
          >
            立即升级
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
```

- [ ] **Step 2: Verify lint**

Run: `cd dx-web && npx eslint src/features/web/games/components/upgrade-dialog.tsx`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/games/components/upgrade-dialog.tsx
git commit -m "feat: add teal-themed UpgradeDialog component"
```

---

### Task 9: Frontend — LevelGrid VIP support

**Files:**
- Modify: `dx-web/src/features/web/games/components/level-grid.tsx`

- [ ] **Step 1: Add isVip prop and upgrade dialog state**

Update the `LevelGrid` component to accept `isVip` and manage the upgrade dialog:

```tsx
import { useState } from "react";
import { Lock, Star } from "lucide-react";
import { UpgradeDialog } from "@/features/web/games/components/upgrade-dialog";
```

Change the `LevelGrid` props:

```tsx
export function LevelGrid({
  levels,
  completedLevels,
  isVip,
  onLevelClick,
}: {
  levels: { id: string; name: string; order: number }[];
  completedLevels: number;
  isVip: boolean;
  onLevelClick?: (levelId: string, levelName: string) => void;
}) {
```

- [ ] **Step 2: Add upgrade dialog state and update level status logic**

Inside `LevelGrid`, add state and update the status computation:

```tsx
const [upgradeOpen, setUpgradeOpen] = useState(false);
const totalLevels = levels.length;
```

In the `.map()`, change the status logic:

```tsx
{levels.map((lv, idx) => {
  const levelNum = idx + 1;
  const status = isVip
    ? "current"
    : levelNum <= completedLevels
      ? "completed"
      : levelNum === completedLevels + 1
        ? "current"
        : "locked";

  return (
    <LevelCell
      key={lv.id}
      level={levelNum}
      name={lv.name}
      status={status}
      onClick={
        status === "locked"
          ? () => setUpgradeOpen(true)
          : onLevelClick
            ? () => onLevelClick(lv.id, lv.name)
            : undefined
      }
    />
  );
})}
```

Add the `UpgradeDialog` at the end of the component, inside the outer `<div>`:

```tsx
<UpgradeDialog open={upgradeOpen} onOpenChange={setUpgradeOpen} />
```

- [ ] **Step 3: Make locked LevelCell clickable**

Currently the locked `LevelCell` has no cursor or click handler. Update the locked case (the default return in `LevelCell`) to support `onClick`:

```tsx
return (
  <div
    onClick={onClick}
    className={`relative flex h-[67px] w-[67px] flex-col items-center justify-center gap-0.5 rounded-[10px] bg-muted${onClick ? " cursor-pointer" : ""}`}
  >
    <div className="absolute -left-[5px] -top-[5px] flex h-4 w-4 items-center justify-center rounded-full bg-amber-500">
      <Lock className="h-[9px] w-[9px] text-white" />
    </div>
    <span className="text-lg font-extrabold text-muted-foreground">{level}</span>
    <span className="text-[9px] font-medium text-muted-foreground">
      {shortName}
    </span>
  </div>
);
```

- [ ] **Step 4: Verify lint**

Run: `cd dx-web && npx eslint src/features/web/games/components/level-grid.tsx`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add dx-web/src/features/web/games/components/level-grid.tsx
git commit -m "feat: add VIP-based level locking to LevelGrid with upgrade dialog"
```

---

### Task 10: Frontend — Wire VIP into game detail page

**Files:**
- Modify: `dx-web/src/features/web/games/components/game-detail-content.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/games/[id]/page.tsx`

- [ ] **Step 1: Add isVip prop to GameDetailContent**

In `dx-web/src/features/web/games/components/game-detail-content.tsx`, add `isVip` to the `game` prop type:

```tsx
interface GameDetailContentProps {
  game: {
    id: string;
    name: string;
    description: string;
    mode: string;
    coverUrl: string | null;
    levelCount: number;
    playerCount: string;
    levels: Level[];
    completedLevels: number;
    isVip: boolean;
  };
```

Pass `isVip` to `LevelGrid`:

```tsx
<LevelGrid
  levels={game.levels}
  completedLevels={game.completedLevels}
  isVip={game.isVip}
  onLevelClick={handleLevelClick}
/>
```

- [ ] **Step 2: Fetch profile and compute isVip in game detail page**

In `dx-web/src/app/(web)/hall/(main)/games/[id]/page.tsx`:

Add import:

```tsx
import { isVipActive } from "@/lib/vip";
import type { UserGrade } from "@/consts/user-grade";
```

In the `load()` function, add profile fetch to the `Promise.all` (line 133). Change the existing `Promise.all` to include profile:

```tsx
const [activeSessionRes, statsRes, favoritedRes, profileRes] = await Promise.all([
  sessionApi.checkAnyActive(mapped.id),
  apiClient.get<GameStats>(`/api/games/${mapped.id}/stats`),
  apiClient.get<{ favorited: boolean }>(`/api/games/${mapped.id}/favorited`),
  apiClient.get<{ grade: string; vip_due_at: string | null }>("/api/user/profile"),
]);
```

After the existing variable declarations, compute VIP status:

```tsx
const profile = profileRes.code === 0 ? profileRes.data : null;
const isVip = profile
  ? isVipActive(profile.grade as UserGrade, profile.vip_due_at)
  : false;
```

Add a state for `isVip`:

```tsx
const [vip, setVip] = useState(false);
```

Set it in the load function:

```tsx
setVip(isVip);
```

Pass it in the `GameDetailContent` props:

```tsx
<GameDetailContent
  game={{
    id: game.id,
    name: game.name,
    description: game.description ?? "",
    mode: game.mode,
    coverUrl: game.coverUrl,
    levelCount: game.levelCount,
    playerCount: String(myStats?.totalSessions ?? 0),
    levels: game.levels,
    completedLevels: 0,
    isVip: vip,
  }}
```

- [ ] **Step 3: Verify lint**

Run: `cd dx-web && npx eslint src/features/web/games/components/game-detail-content.tsx src/app/\(web\)/hall/\(main\)/games/\[id\]/page.tsx`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/games/components/game-detail-content.tsx dx-web/src/app/\(web\)/hall/\(main\)/games/\[id\]/page.tsx
git commit -m "feat: wire VIP status into game detail page and level grid"
```

---

### Task 11: Frontend — Guard play-single page

**Files:**
- Modify: `dx-web/src/app/(web)/hall/play-single/[id]/page.tsx`

- [ ] **Step 1: Add VIP guard**

Add imports:

```tsx
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { isVipActive } from "@/lib/vip";
import type { UserGrade } from "@/consts/user-grade";
```

Update the `ApiProfileData` type to include VIP fields:

```tsx
type ApiProfileData = {
  id?: string;
  nickname?: string | null;
  username?: string;
  avatarUrl?: string | null;
  grade?: string;
  vip_due_at?: string | null;
};
```

In the `load()` function, after fetching profile and game data, add VIP check before `setLoaded(true)`:

```tsx
// VIP guard: redirect if non-VIP tries to access non-first level
const profile = profileRes.code === 0 ? profileRes.data : null;
const vip = profile
  ? isVipActive((profile.grade ?? "free") as UserGrade, profile.vip_due_at ?? null)
  : false;

if (!vip && level) {
  const firstLevel = g.levels?.[0];
  if (firstLevel && level !== firstLevel.id) {
    toast.error("升级会员解锁全部关卡");
    router.replace(`/hall/games/${id}`);
    return;
  }
}
```

Add `router` from `useRouter()` at the top of the component (it already uses `useSearchParams`, add `useRouter`).

- [ ] **Step 2: Verify lint**

Run: `cd dx-web && npx eslint src/app/\(web\)/hall/play-single/\[id\]/page.tsx`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/app/\(web\)/hall/play-single/\[id\]/page.tsx
git commit -m "feat: add VIP guard redirect on play-single page for non-first levels"
```

---

### Task 12: Frontend — Guard group pages

**Files:**
- Modify: `dx-web/src/features/web/groups/components/group-list-content.tsx`
- Modify: `dx-web/src/features/web/groups/components/group-card.tsx`
- Modify: `dx-web/src/features/web/groups/components/group-invite-content.tsx`

- [ ] **Step 1: Add VIP state to GroupListContent**

In `dx-web/src/features/web/groups/components/group-list-content.tsx`, add imports:

```tsx
import { apiClient } from "@/lib/api-client";
import { isVipActive } from "@/lib/vip";
import type { UserGrade } from "@/consts/user-grade";
import { UpgradeDialog } from "@/features/web/games/components/upgrade-dialog";
```

Add state inside `GroupListContent`:

```tsx
const [isVip, setIsVip] = useState(true); // optimistic default
const [upgradeOpen, setUpgradeOpen] = useState(false);
```

Add useEffect to fetch VIP status:

```tsx
useEffect(() => {
  apiClient.get<{ grade: string; vip_due_at: string | null }>("/api/user/profile").then((res) => {
    if (res.code === 0 && res.data) {
      setIsVip(isVipActive(res.data.grade as UserGrade, res.data.vip_due_at));
    }
  });
}, []);
```

- [ ] **Step 2: Guard create button**

Change the "创建学习群" button onClick:

```tsx
onClick={() => isVip ? setCreateOpen(true) : setUpgradeOpen(true)}
```

- [ ] **Step 3: Guard group card actions**

Pass `isVip` and `onUpgrade` to `GroupCard`:

```tsx
<GroupCard
  key={group.id}
  group={group}
  isMember={group.is_member}
  isVip={isVip}
  onJoin={() => isVip ? setApplyTarget(group) : setUpgradeOpen(true)}
  onUpgrade={() => setUpgradeOpen(true)}
/>
```

- [ ] **Step 4: Add UpgradeDialog to the component**

Add at the end of the fragment, after the apply confirm dialog:

```tsx
<UpgradeDialog
  open={upgradeOpen}
  onOpenChange={setUpgradeOpen}
  title="会员专属功能"
  message="升级会员即可创建和加入学习群，与同学一起学习"
/>
```

- [ ] **Step 5: Update GroupCard to support VIP gating**

In `dx-web/src/features/web/groups/components/group-card.tsx`, add props:

```tsx
interface GroupCardProps {
  group: Group;
  isMember?: boolean;
  isVip?: boolean;
  highlighted?: boolean;
  onJoin?: () => void;
  onUpgrade?: () => void;
}
```

Update the component signature:

```tsx
export function GroupCard({ group, isMember = true, isVip = true, highlighted = false, onJoin, onUpgrade }: GroupCardProps) {
```

Change the member link: if not VIP, clicking the card opens upgrade dialog instead of navigating:

```tsx
if (isMember) {
  if (isVip) {
    return <Link href={`/hall/groups/${group.id}`} className="block">{cardContent}</Link>;
  }
  return <div className="block cursor-pointer" onClick={onUpgrade}>{cardContent}</div>;
}
return <div className="block">{cardContent}</div>;
```

- [ ] **Step 6: Guard invite page**

In `dx-web/src/features/web/groups/components/group-invite-content.tsx`, add imports:

```tsx
import { useState, useEffect } from "react";
import { apiClient } from "@/lib/api-client";
import { isVipActive } from "@/lib/vip";
import type { UserGrade } from "@/consts/user-grade";
import { UpgradeDialog } from "@/features/web/games/components/upgrade-dialog";
```

Add VIP state inside `GroupInviteContent` (only if logged in):

```tsx
const [isVip, setIsVip] = useState(true);
const [upgradeOpen, setUpgradeOpen] = useState(false);

useEffect(() => {
  if (!isLoggedIn) return;
  apiClient.get<{ grade: string; vip_due_at: string | null }>("/api/user/profile").then((res) => {
    if (res.code === 0 && res.data) {
      setIsVip(isVipActive(res.data.grade as UserGrade, res.data.vip_due_at));
    }
  });
}, [isLoggedIn]);
```

Change the logged-in button from directly calling `handleApply` to checking VIP first:

```tsx
{isLoggedIn ? (
  <button
    type="button"
    onClick={() => isVip ? handleApply() : setUpgradeOpen(true)}
    disabled={applying}
    className="flex w-full items-center justify-center gap-2 rounded-xl bg-teal-600 py-3 text-sm font-semibold text-white hover:bg-teal-700 disabled:opacity-60"
  >
    {applying && <Loader2 className="h-4 w-4 animate-spin" />}
    申请加入
  </button>
) : (
```

Add `UpgradeDialog` at the end of the outer div:

```tsx
<UpgradeDialog
  open={upgradeOpen}
  onOpenChange={setUpgradeOpen}
  title="会员专属功能"
  message="升级会员即可创建和加入学习群，与同学一起学习"
/>
```

- [ ] **Step 7: Verify lint**

Run: `cd dx-web && npx eslint src/features/web/groups/components/group-list-content.tsx src/features/web/groups/components/group-card.tsx src/features/web/groups/components/group-invite-content.tsx`
Expected: No errors

- [ ] **Step 8: Commit**

```bash
git add dx-web/src/features/web/groups/components/group-list-content.tsx dx-web/src/features/web/groups/components/group-card.tsx dx-web/src/features/web/groups/components/group-invite-content.tsx
git commit -m "feat: add VIP gating to group list, group card, and invite pages"
```

---

### Task 13: Frontend — Build verification

**Files:** None (verification only)

- [ ] **Step 1: Run full frontend build**

Run: `cd dx-web && npm run build`
Expected: BUILD SUCCESS with no errors

- [ ] **Step 2: Run lint**

Run: `cd dx-web && npm run lint`
Expected: No lint errors

- [ ] **Step 3: Run backend build and vet**

Run: `cd dx-api && go build ./... && go vet ./...`
Expected: No errors

- [ ] **Step 4: Commit if any fixes were needed**

Only if previous steps required fixes.

---

### Task 14: Documentation updates

**Files:**
- Modify: `docs/game-word-sentence-single-rule.md`
- Modify: `docs/game-word-sentence-group-rule.md`

- [ ] **Step 1: Add Level Access Control section to single-rule doc**

Append to the end of `docs/game-word-sentence-single-rule.md`:

```markdown

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
```

- [ ] **Step 2: Add VIP Requirement section to group-rule doc**

Append to the end of `docs/game-word-sentence-group-rule.md`:

```markdown

## VIP Requirement

### Access Control

All group operations require an active VIP membership, with two exceptions:

| Operation | VIP Required |
|-----------|-------------|
| View groups list (`ListGroups`) | No |
| View invite info (`GetGroupByInviteCode`) | No |
| All other operations | Yes |

This includes: creating groups, joining groups, viewing group details, managing members, setting games, starting/playing group games, and all group play recording endpoints.

### Enforcement

- **Backend**: Every guarded service function calls `requireVip(userID)` at the top. Returns error code `40302` (`CodeVipRequired`) with HTTP 403 if the user is not VIP.
- **Frontend**: The groups list page gates the "创建学习群" button, group card navigation, and join/apply actions behind VIP status. The invite page gates the "申请加入" button. All show an upgrade dialog directing to `/purchase/membership`.

### VIP Definition

Same as game level access control: `grade == "lifetime"` OR (`grade != "free"` AND `vipDueAt > now()`).
```

- [ ] **Step 3: Commit**

```bash
git add docs/game-word-sentence-single-rule.md docs/game-word-sentence-group-rule.md
git commit -m "docs: add VIP level access control and group VIP requirement sections"
```

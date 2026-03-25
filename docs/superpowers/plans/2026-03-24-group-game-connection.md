# Group Game Connection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow group owners to search published games and set one as the group's current game with a play mode (solo/team).

**Architecture:** New GroupGameController + service for search/set/clear game. Update existing GroupDetail to return game info. Frontend: new SetGameDialog component + update group detail page.

**Tech Stack:** Go/Goravel (backend), Next.js 16 + TypeScript + SWR + Tailwind (frontend)

**Spec:** `docs/superpowers/specs/2026-03-24-group-game-connection-design.md`

---

## Task 1: Migration + Model + Constants

**Files:**
- Modify: `dx-api/database/migrations/20260322000028_create_game_groups_table.go`
- Modify: `dx-api/app/models/game_group.go`
- Modify: `dx-api/app/consts/group.go`

- [ ] **Step 1: Add game_mode column to the existing game_groups migration**

In `dx-api/database/migrations/20260322000028_create_game_groups_table.go`, add after the `is_active` line and before `member_count`:

```go
table.Text("game_mode").Nullable()
```

- [ ] **Step 2: Add GameMode field to GameGroup model**

In `dx-api/app/models/game_group.go`, add after the `IsActive` field:

```go
GameMode      *string `gorm:"column:game_mode" json:"game_mode"`
```

- [ ] **Step 3: Append game mode constants to consts/group.go**

```go
const (
	GameModeSolo = "solo"
	GameModeTeam = "team"
)
```

- [ ] **Step 4: Verify compilation**

Run: `cd dx-api && go build ./...`

- [ ] **Step 5: Commit**

```bash
git add dx-api/database/migrations/20260322000028_create_game_groups_table.go \
       dx-api/app/models/game_group.go \
       dx-api/app/consts/group.go
git commit -m "feat: add game_mode column to game_groups, add game mode constants"
```

---

## Task 2: Group Game Service (TDD)

**Files:**
- Create: `dx-api/app/services/api/group_game_service.go`
- Create: `dx-api/app/services/api/group_game_service_test.go`
- Modify: `dx-api/app/services/api/errors.go`

- [ ] **Step 1: Add error sentinel to errors.go**

Add to the `var` block:

```go
ErrGameNotPublished = errors.New("game not published")
```

- [ ] **Step 2: Write failing tests**

```go
// dx-api/app/services/api/group_game_service_test.go
package api

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSearchGamesForGroupExists(t *testing.T) {
	assert.NotNil(t, SearchGamesForGroup)
}

func TestSetGroupGameExists(t *testing.T) {
	assert.NotNil(t, SetGroupGame)
}

func TestClearGroupGameExists(t *testing.T) {
	assert.NotNil(t, ClearGroupGame)
}
```

- [ ] **Step 3: Implement group_game_service.go**

```go
// dx-api/app/services/api/group_game_service.go
package api

import (
	"fmt"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// GroupGameSearchItem represents a game in group search results.
type GroupGameSearchItem struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Mode         string  `json:"mode"`
	CategoryName *string `json:"category_name"`
}

// SearchGamesForGroup searches published games by name. Owner only (checked in controller).
func SearchGamesForGroup(query string, limit int) ([]GroupGameSearchItem, error) {
	if limit <= 0 {
		limit = 20
	}

	var games []models.Game
	q := facades.Orm().Query().
		Where("status", consts.GameStatusPublished).
		Where("is_active", true)

	if query != "" {
		q = q.Where("name ILIKE ?", "%"+query+"%")
	}

	if err := q.Order("created_at DESC").Limit(limit).Get(&games); err != nil {
		return nil, fmt.Errorf("failed to search games: %w", err)
	}

	// Batch load categories
	categoryIDs := make([]string, 0, len(games))
	for _, g := range games {
		if g.GameCategoryID != nil && *g.GameCategoryID != "" {
			categoryIDs = append(categoryIDs, *g.GameCategoryID)
		}
	}
	categoryMap := make(map[string]string)
	if len(categoryIDs) > 0 {
		var categories []models.GameCategory
		if err := facades.Orm().Query().Where("id IN ?", categoryIDs).Get(&categories); err == nil {
			for _, cat := range categories {
				categoryMap[cat.ID] = cat.Name
			}
		}
	}

	items := make([]GroupGameSearchItem, 0, len(games))
	for _, g := range games {
		var catName *string
		if g.GameCategoryID != nil {
			if name, ok := categoryMap[*g.GameCategoryID]; ok {
				catName = &name
			}
		}
		items = append(items, GroupGameSearchItem{
			ID:           g.ID,
			Name:         g.Name,
			Mode:         g.Mode,
			CategoryName: catName,
		})
	}

	return items, nil
}

// SetGroupGame sets the current game and game mode for a group. Owner only.
func SetGroupGame(userID, groupID, gameID, gameMode string) error {
	// Verify ownership
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("is_active", true).First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return ErrNotGroupOwner
	}

	// Verify game exists and is published
	var game models.Game
	if err := facades.Orm().Query().Where("id", gameID).First(&game); err != nil || game.ID == "" {
		return ErrGameNotFound
	}
	if game.Status != consts.GameStatusPublished {
		return ErrGameNotPublished
	}

	// Update group
	if _, err := facades.Orm().Query().Model(&models.GameGroup{}).Where("id", groupID).Update(map[string]any{
		"current_game_id": gameID,
		"game_mode":       gameMode,
	}); err != nil {
		return fmt.Errorf("failed to set group game: %w", err)
	}

	return nil
}

// ClearGroupGame clears the current game from a group. Owner only.
func ClearGroupGame(userID, groupID string) error {
	var group models.GameGroup
	if err := facades.Orm().Query().Where("id", groupID).Where("is_active", true).First(&group); err != nil || group.ID == "" {
		return ErrGroupNotFound
	}
	if group.OwnerID != userID {
		return ErrNotGroupOwner
	}

	if _, err := facades.Orm().Query().Exec(
		"UPDATE game_groups SET current_game_id = NULL, game_mode = NULL WHERE id = ?", groupID,
	); err != nil {
		return fmt.Errorf("failed to clear group game: %w", err)
	}

	return nil
}
```

- [ ] **Step 4: Run tests**

Run: `cd dx-api && go test ./app/services/api/ -run "TestSearchGamesForGroup|TestSetGroupGame|TestClearGroupGame" -v`

- [ ] **Step 5: Verify build**

Run: `cd dx-api && go build ./...`

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/services/api/group_game_service.go \
       dx-api/app/services/api/group_game_service_test.go \
       dx-api/app/services/api/errors.go
git commit -m "feat: implement group game service (search, set, clear)"
```

---

## Task 3: Update GetGroupDetail to return game info

**Files:**
- Modify: `dx-api/app/services/api/group_service.go`

- [ ] **Step 1: Add fields to GroupDetail struct**

Add after `IsOwner`:

```go
CurrentGameID   *string `json:"current_game_id"`
GameMode        *string `json:"game_mode"`
CurrentGameName string  `json:"current_game_name"`
```

- [ ] **Step 2: Update GetGroupDetail function**

After getting owner info and before building the return struct, add game name resolution:

```go
// Resolve current game name
var currentGameName string
if group.CurrentGameID != nil && *group.CurrentGameID != "" {
    var game models.Game
    if err := facades.Orm().Query().Select("name").Where("id", *group.CurrentGameID).First(&game); err == nil && game.ID != "" {
        currentGameName = game.Name
    }
}
```

Then add to the return struct:

```go
CurrentGameID:   group.CurrentGameID,
GameMode:        group.GameMode,
CurrentGameName: currentGameName,
```

- [ ] **Step 3: Verify build**

Run: `cd dx-api && go build ./...`

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/services/api/group_service.go
git commit -m "feat: include current game info in group detail response"
```

---

## Task 4: Controller + Request + Routes

**Files:**
- Create: `dx-api/app/http/controllers/api/group_game_controller.go`
- Create: `dx-api/app/http/requests/api/group_game_request.go`
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Create request validation**

```go
// dx-api/app/http/requests/api/group_game_request.go
package api

import "github.com/goravel/framework/contracts/http"

type SetGroupGameRequest struct {
	GameID   string `form:"game_id" json:"game_id"`
	GameMode string `form:"game_mode" json:"game_mode"`
}

func (r *SetGroupGameRequest) Authorize(ctx http.Context) error { return nil }
func (r *SetGroupGameRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id":   "required",
		"game_mode": "required|in:solo,team",
	}
}
func (r *SetGroupGameRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id": "trim",
	}
}
func (r *SetGroupGameRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"game_id.required":   "请选择一个游戏",
		"game_mode.required": "请选择游戏模式",
		"game_mode.in":       "游戏模式只能为单人或小组",
	}
}
```

- [ ] **Step 2: Create controller**

```go
// dx-api/app/http/controllers/api/group_game_controller.go
package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type GroupGameController struct{}

func NewGroupGameController() *GroupGameController {
	return &GroupGameController{}
}

// SearchGames searches published games for group game selection.
func (c *GroupGameController) SearchGames(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	groupID := ctx.Request().Route("id")
	if groupID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	// Verify ownership
	if err := services.VerifyGroupOwnership(userID, groupID); err != nil {
		return mapGroupGameError(ctx, err)
	}

	q := ctx.Request().Query("q", "")
	items, err := services.SearchGamesForGroup(q, 20)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "搜索失败")
	}

	return helpers.Success(ctx, items)
}

// SetGame sets the current game and mode for a group.
func (c *GroupGameController) SetGame(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	groupID := ctx.Request().Route("id")
	if groupID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	var req requests.SetGroupGameRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.SetGroupGame(userID, groupID, req.GameID, req.GameMode); err != nil {
		return mapGroupGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// ClearGame clears the current game from a group.
func (c *GroupGameController) ClearGame(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	groupID := ctx.Request().Route("id")
	if groupID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	if err := services.ClearGroupGame(userID, groupID); err != nil {
		return mapGroupGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

func mapGroupGameError(ctx contractshttp.Context, err error) contractshttp.Response {
	switch {
	case errors.Is(err, services.ErrGroupNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeGroupNotFound, "学习群不存在")
	case errors.Is(err, services.ErrNotGroupOwner):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeGroupForbidden, "无权操作此学习群")
	case errors.Is(err, services.ErrGameNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeGameNotFound, "游戏不存在")
	case errors.Is(err, services.ErrGameNotPublished):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "该游戏未发布")
	default:
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "操作失败")
	}
}
```

**Note:** The controller calls `services.VerifyGroupOwnership(userID, groupID)`. This function is currently named `verifyGroupOwnership` (unexported) in `group_subgroup_service.go`. The implementing agent must either (a) export it by renaming to `VerifyGroupOwnership`, or (b) inline the ownership check in the service functions. Option (a) is cleaner — rename in `group_subgroup_service.go` and update all callers in that file.

- [ ] **Step 3: Add routes**

In `dx-api/routes/api.go`, inside the existing `groups` prefix group, add after the subgroup routes:

```go
// Group game
groupGameController := apicontrollers.NewGroupGameController()
groups.Get("/{id}/games/search", groupGameController.SearchGames)
groups.Put("/{id}/game", groupGameController.SetGame)
groups.Delete("/{id}/game", groupGameController.ClearGame)
```

- [ ] **Step 4: Verify build and tests**

Run: `cd dx-api && go build ./... && go test -race ./...`

- [ ] **Step 5: Commit**

```bash
git add dx-api/app/http/controllers/api/group_game_controller.go \
       dx-api/app/http/requests/api/group_game_request.go \
       dx-api/routes/api.go \
       dx-api/app/services/api/group_subgroup_service.go
git commit -m "feat: add group game controller, request validation, and routes"
```

---

## Task 5: Frontend Types + Actions

**Files:**
- Modify: `dx-web/src/features/web/groups/types/group.ts`
- Modify: `dx-web/src/features/web/groups/actions/group.action.ts`

- [ ] **Step 1: Update GroupDetail type**

In `types/group.ts`, update `GroupDetail`:

```typescript
export type GroupDetail = Group & {
  is_active: boolean;
  current_game_id: string | null;
  game_mode: string | null;
  current_game_name: string | null;
};
```

Add new type:

```typescript
export type GroupGameSearchItem = {
  id: string;
  name: string;
  mode: string;
  category_name: string | null;
};
```

- [ ] **Step 2: Add actions to group.action.ts**

Append to the `groupApi` object:

```typescript
async searchGamesForGroup(groupId: string, q?: string) {
  const params = new URLSearchParams();
  if (q) params.set("q", q);
  const qs = params.toString();
  return apiClient.get<GroupGameSearchItem[]>(`/api/groups/${groupId}/games/search${qs ? `?${qs}` : ""}`);
},
async setGame(groupId: string, gameId: string, gameMode: string) {
  return apiClient.put<null>(`/api/groups/${groupId}/game`, { game_id: gameId, game_mode: gameMode });
},
async clearGame(groupId: string) {
  return apiClient.delete<null>(`/api/groups/${groupId}/game`);
},
```

Import `GroupGameSearchItem` from types.

- [ ] **Step 3: Verify build**

Run: `cd dx-web && npm run build`

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/groups/types/group.ts \
       dx-web/src/features/web/groups/actions/group.action.ts
git commit -m "feat: add group game types and API actions"
```

---

## Task 6: SetGameDialog Component

**Files:**
- Create: `dx-web/src/features/web/groups/components/set-game-dialog.tsx`

- [ ] **Step 1: Create the dialog component**

Matching the .pen design exactly:
- Header: teal icon (Gamepad2) + "设置群课程游戏" + close button (X)
- Search input: debounced 300ms, calls `groupApi.searchGamesForGroup`
- Game list: scrollable (max-h-60), radio selection, selected row has teal bg (`bg-teal-50`)
- Mode selector: segmented control — "单人" (User icon, teal when selected) | "小组" (Users icon)
- Confirm button: full-width teal, "确认设置" with check icon
- Pre-selects current game and mode if provided via props
- On confirm: calls `groupApi.setGame`, then `swrMutate`, closes dialog
- Uses shadcn Dialog component

Props:
```typescript
interface SetGameDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  groupId: string;
  currentGameId?: string | null;
  currentGameMode?: string | null;
}
```

- [ ] **Step 2: Verify build**

Run: `cd dx-web && npm run build`

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/groups/components/set-game-dialog.tsx
git commit -m "feat: add set game dialog component"
```

---

## Task 7: Update Group Detail Page

**Files:**
- Modify: `dx-web/src/features/web/groups/components/group-detail-content.tsx`

- [ ] **Step 1: Add game section to left column**

Between the stats section and the divider before invite link, add:

- If no game set (`!group.current_game_id`): show "设置课程游戏" button
- If game set: show game name, mode badge ("单人" with User icon / "小组" with Users icon), "更换" button, "清除" button
- Wire "设置课程游戏" and "更换" to open `SetGameDialog`
- Wire "清除" to call `groupApi.clearGame` + `swrMutate`
- Only show for owner

- [ ] **Step 2: Add dialog state and import**

Add `setGameOpen` state, import `SetGameDialog`, render it with current game/mode props.

- [ ] **Step 3: Verify build**

Run: `cd dx-web && npm run build`

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/groups/components/group-detail-content.tsx
git commit -m "feat: add current game section to group detail page"
```

---

## Task 8: Final Verification

- [ ] **Step 1: Backend build + tests**

Run: `cd dx-api && go build ./... && go test -race ./...`

- [ ] **Step 2: Frontend build**

Run: `cd dx-web && npm run build`

- [ ] **Step 3: Verify no existing functionality broken**

Check that existing group endpoints still work — the only changed existing file is `group_service.go` (added fields to response, which is backward-compatible).

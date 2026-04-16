# Private Games Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an `is_private` boolean to games, letting creators mark course games as private so they are hidden from all public game listings.

**Architecture:** New boolean column on `games` table, threaded through the existing CRUD flow (request → controller → service → model). Public query paths gain a `WHERE is_private = false` filter. Frontend adds a ShadCN Switch to create/edit forms and passes the value through existing actions.

**Tech Stack:** Go/Goravel (backend), Next.js 16 + ShadCN/Radix (frontend), PostgreSQL

---

## File Map

### Backend (dx-api)

| Action | File |
|--------|------|
| Modify | `database/migrations/20260322000016_create_games_table.go` — add `is_private` column |
| Modify | `app/models/game.go` — add `IsPrivate` field |
| Modify | `app/http/requests/api/course_game_request.go` — add `IsPrivate` to create/update requests |
| Modify | `app/http/controllers/api/course_game_controller.go` — pass `isPrivate` to service |
| Modify | `app/services/api/course_game_service.go` — accept `isPrivate` in Create/Update, include in detail response |
| Modify | `app/services/api/game_service.go` — filter `is_private = false` in 4 public queries |

### Frontend (dx-web)

| Action | File |
|--------|------|
| Modify | `src/features/web/ai-custom/components/create-course-form.tsx` — add Switch |
| Modify | `src/features/web/ai-custom/components/edit-game-dialog.tsx` — add Switch with default value |
| Modify | `src/features/web/ai-custom/actions/course-game.action.ts` — pass `isPrivate` in create/update |
| Modify | `src/features/web/ai-custom/components/course-detail-content.tsx` — map `isPrivate` from API |
| Modify | `src/features/web/ai-custom/components/game-hero-card.tsx` — add `isPrivate` to game type and pass to edit dialog |

---

## Task 1: Add `is_private` column to games migration

**Files:**
- Modify: `dx-api/database/migrations/20260322000016_create_games_table.go:33`

- [ ] **Step 1: Add the column definition**

In the `Up()` method, after the `is_selective` line (line 33), add:

```go
table.Boolean("is_private").Default(false)
```

And after the existing indexes (before the closing `})`), add:

```go
table.Index("is_private")
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: clean build, no errors

- [ ] **Step 3: Commit**

```bash
git add dx-api/database/migrations/20260322000016_create_games_table.go
git commit -m "feat(api): add is_private column to games migration"
```

---

## Task 2: Add `IsPrivate` field to Game model

**Files:**
- Modify: `dx-api/app/models/game.go:20`

- [ ] **Step 1: Add the field**

After the `IsSelective` field (line 20), add:

```go
IsPrivate   bool    `gorm:"column:is_private" json:"is_private"`
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: clean build, no errors

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/models/game.go
git commit -m "feat(api): add IsPrivate field to Game model"
```

---

## Task 3: Add `isPrivate` to request validation

**Files:**
- Modify: `dx-api/app/http/requests/api/course_game_request.go:11-18,54-61`

- [ ] **Step 1: Add field to CreateGameRequest**

Add to the `CreateGameRequest` struct (after the `CoverID` field on line 17):

```go
IsPrivate bool `form:"isPrivate" json:"isPrivate"`
```

No validation rule needed — Go zero-value `false` is the correct default for a bool.

- [ ] **Step 2: Add field to UpdateGameRequest**

Add to the `UpdateGameRequest` struct (after the `CoverID` field on line 60):

```go
IsPrivate bool `form:"isPrivate" json:"isPrivate"`
```

- [ ] **Step 3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: clean build, no errors

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/http/requests/api/course_game_request.go
git commit -m "feat(api): add isPrivate to create/update game requests"
```

---

## Task 4: Thread `isPrivate` through service layer

**Files:**
- Modify: `dx-api/app/services/api/course_game_service.go:168,199,648-661`

- [ ] **Step 1: Update CreateGame signature and body**

Change the `CreateGame` function signature (line 168) from:

```go
func CreateGame(userID, name string, description *string, mode string, categoryID, pressID, coverID *string) (string, error) {
```

to:

```go
func CreateGame(userID, name string, description *string, mode string, categoryID, pressID, coverID *string, isPrivate bool) (string, error) {
```

In the `game := models.Game{...}` literal (line 174-186), add after `Status`:

```go
IsPrivate:  isPrivate,
```

- [ ] **Step 2: Update UpdateGame signature and body**

Change the `UpdateGame` function signature (line 199) from:

```go
func UpdateGame(userID, gameID, name string, description *string, mode string, categoryID, pressID, coverID *string) error {
```

to:

```go
func UpdateGame(userID, gameID, name string, description *string, mode string, categoryID, pressID, coverID *string, isPrivate bool) error {
```

In the `Update(map[string]any{...})` call (lines 212-219), add:

```go
"is_private": isPrivate,
```

- [ ] **Step 3: Add `IsPrivate` to CourseGameDetailData response**

Add to the `CourseGameDetailData` struct (after the `Status` field on line 48):

```go
IsPrivate      bool                  `json:"isPrivate"`
```

In `GetCourseGameDetail`, in the return struct (lines 648-661), add after `Status`:

```go
IsPrivate:      game.IsPrivate,
```

- [ ] **Step 4: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: compilation errors in controller (expected — we haven't updated the call sites yet)

- [ ] **Step 5: Commit**

```bash
git add dx-api/app/services/api/course_game_service.go
git commit -m "feat(api): thread isPrivate through course game service"
```

---

## Task 5: Update controller to pass `isPrivate`

**Files:**
- Modify: `dx-api/app/http/controllers/api/course_game_controller.go:77,111`

- [ ] **Step 1: Update Create handler**

Change line 77 from:

```go
gameID, err := services.CreateGame(userID, req.Name, req.Description, req.GameMode, categoryID, pressID, req.CoverID)
```

to:

```go
gameID, err := services.CreateGame(userID, req.Name, req.Description, req.GameMode, categoryID, pressID, req.CoverID, req.IsPrivate)
```

- [ ] **Step 2: Update Update handler**

Change line 111 from:

```go
err = services.UpdateGame(userID, gameID, req.Name, req.Description, req.GameMode, categoryID, pressID, req.CoverID)
```

to:

```go
err = services.UpdateGame(userID, gameID, req.Name, req.Description, req.GameMode, categoryID, pressID, req.CoverID, req.IsPrivate)
```

- [ ] **Step 3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: clean build, no errors

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/http/controllers/api/course_game_controller.go
git commit -m "feat(api): pass isPrivate from controller to service"
```

---

## Task 6: Filter private games from public queries

**Files:**
- Modify: `dx-api/app/services/api/game_service.go:71-73,222-224,303-306,358-360`

- [ ] **Step 1: Filter in ListPublishedGames**

After line 73 (the `Where("is_active", true)` line), add:

```go
Where("is_private", false)
```

So the query builder chain becomes:

```go
query := facades.Orm().Query().
    Where("status", consts.GameStatusPublished).
    Where("is_active", true).
    Where("is_private", false)
```

- [ ] **Step 2: Filter in SearchGames**

In the `SearchGames` function (lines 222-224), add `.Where("is_private", false)` to the query chain. The query becomes:

```go
if err := facades.Orm().Query().
    Where("status", consts.GameStatusPublished).
    Where("is_active", true).
    Where("is_private", false).
    Where("name ILIKE ?", "%"+queryStr+"%").
```

- [ ] **Step 3: Filter in GetPlayedGames**

In the `GetPlayedGames` function (lines 303-306), add `.Where("is_private", false)` to the games loading query:

```go
if err := facades.Orm().Query().
    Where("id IN ?", gameIDs).
    Where("status", consts.GameStatusPublished).
    Where("is_active", true).
    Where("is_private", false).
    Get(&games); err != nil {
```

- [ ] **Step 4: Filter in GetGameDetail**

In the `GetGameDetail` function (lines 358-360), add `.Where("is_private", false)`:

```go
if err := facades.Orm().Query().
    Where("id", gameID).
    Where("status", consts.GameStatusPublished).
    Where("is_active", true).
    Where("is_private", false).
    First(&game); err != nil {
```

- [ ] **Step 5: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: clean build, no errors

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/services/api/game_service.go
git commit -m "feat(api): filter private games from all public queries"
```

---

## Task 7: Add Switch to create course form

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/components/create-course-form.tsx`
- Modify: `dx-web/src/features/web/ai-custom/actions/course-game.action.ts:37-44`

- [ ] **Step 1: Add Switch import to create-course-form.tsx**

Add to the imports (after the existing component imports):

```tsx
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
```

- [ ] **Step 2: Add `isPrivate` state**

In the component body (after line 34, the `const [pressId, setPressId]` line), add:

```tsx
const [isPrivate, setIsPrivate] = useState(false);
```

- [ ] **Step 3: Add Switch UI**

After the cover image `</div>` block (after line 187) and before the closing `</div>` of the form fields container, add:

```tsx
{/* Private switch */}
<div className="flex items-center justify-between rounded-[10px] border border-border px-4 py-3">
  <Label htmlFor="isPrivate" className="text-sm font-medium text-foreground">
    私有
    <span className="ml-1 text-xs font-normal text-muted-foreground">
      (仅自己可见)
    </span>
  </Label>
  <Switch
    id="isPrivate"
    checked={isPrivate}
    onCheckedChange={setIsPrivate}
  />
</div>
<input type="hidden" name="isPrivate" value={isPrivate ? "true" : "false"} />
```

- [ ] **Step 4: Update create action to pass isPrivate**

In `course-game.action.ts`, in the `createCourseGameAction` function body (lines 37-44), add `isPrivate` to the body object:

```ts
isPrivate: formData.get("isPrivate") === "true",
```

- [ ] **Step 5: Verify no lint issues**

Run: `cd dx-web && npm run lint`
Expected: no errors

- [ ] **Step 6: Commit**

```bash
git add dx-web/src/features/web/ai-custom/components/create-course-form.tsx dx-web/src/features/web/ai-custom/actions/course-game.action.ts
git commit -m "feat(web): add private switch to create course game form"
```

---

## Task 8: Add Switch to edit game dialog

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/components/edit-game-dialog.tsx`
- Modify: `dx-web/src/features/web/ai-custom/actions/course-game.action.ts:166-173`
- Modify: `dx-web/src/features/web/ai-custom/components/course-detail-content.tsx:13-27,46-66`
- Modify: `dx-web/src/features/web/ai-custom/components/game-hero-card.tsx:39-52,222-230`

- [ ] **Step 1: Add `isPrivate` to `defaultValues` type in edit-game-dialog.tsx**

In the `EditGameDialogProps` type (line 37-44), add after `coverUrl`:

```ts
isPrivate: boolean;
```

- [ ] **Step 2: Add Switch import to edit-game-dialog.tsx**

Add to the imports:

```tsx
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
```

- [ ] **Step 3: Add `isPrivate` state to edit-game-dialog.tsx**

In the component body (after line 57, the `useUpdateCourseGame` line), add:

```tsx
const [isPrivate, setIsPrivate] = useState(defaultValues.isPrivate);
```

Add `useState` to the existing React import (line 1 — it's not currently imported in this file). The import should become:

```tsx
import { useState } from "react";
```

- [ ] **Step 4: Add Switch UI to edit-game-dialog.tsx**

After the cover image `</div>` block (after line 248) and before the closing `</div>` of the form fields container, add:

```tsx
{/* Private switch */}
<div className="flex items-center justify-between rounded-[10px] border border-border px-4 py-3">
  <Label htmlFor="editIsPrivate" className="text-sm font-medium text-foreground">
    私有
    <span className="ml-1 text-xs font-normal text-muted-foreground">
      (仅自己可见)
    </span>
  </Label>
  <Switch
    id="editIsPrivate"
    checked={isPrivate}
    onCheckedChange={setIsPrivate}
  />
</div>
<input type="hidden" name="isPrivate" value={isPrivate ? "true" : "false"} />
```

- [ ] **Step 5: Update updateCourseGameAction to pass isPrivate**

In `course-game.action.ts`, in the `updateCourseGameAction` function body (lines 166-173), add `isPrivate` to the body object:

```ts
isPrivate: formData.get("isPrivate") === "true",
```

- [ ] **Step 6: Add `isPrivate` to RawGameDetail type in course-detail-content.tsx**

In the `RawGameDetail` type (lines 13-28), add after `coverId`:

```ts
isPrivate?: boolean;
```

- [ ] **Step 7: Map `isPrivate` in course-detail-content.tsx**

In the `mapGameDetail` function return object (lines 46-66), add after `coverId`:

```ts
isPrivate: raw.isPrivate ?? false,
```

- [ ] **Step 8: Add `isPrivate` to GameHeroCardProps game type**

In `game-hero-card.tsx`, in the `GameHeroCardProps` type (lines 39-53), add after `coverId`:

```ts
isPrivate: boolean;
```

- [ ] **Step 9: Pass `isPrivate` to EditGameDialog defaultValues**

In `game-hero-card.tsx`, in the `<EditGameDialog>` JSX (lines 222-230), add after `coverUrl`:

```tsx
isPrivate: game.isPrivate,
```

- [ ] **Step 10: Verify no lint issues**

Run: `cd dx-web && npm run lint`
Expected: no errors

- [ ] **Step 11: Commit**

```bash
git add dx-web/src/features/web/ai-custom/components/edit-game-dialog.tsx dx-web/src/features/web/ai-custom/actions/course-game.action.ts dx-web/src/features/web/ai-custom/components/course-detail-content.tsx dx-web/src/features/web/ai-custom/components/game-hero-card.tsx
git commit -m "feat(web): add private switch to edit game dialog"
```

---

## Task 9: Verify full build

**Files:** None (verification only)

- [ ] **Step 1: Backend build**

Run: `cd dx-api && go build ./...`
Expected: clean build

- [ ] **Step 2: Frontend lint**

Run: `cd dx-web && npm run lint`
Expected: no errors

- [ ] **Step 3: Frontend build**

Run: `cd dx-web && npm run build`
Expected: clean build

- [ ] **Step 4: Commit spec and plan docs**

```bash
git add docs/superpowers/specs/2026-04-16-private-games-design.md docs/superpowers/plans/2026-04-16-private-games.md
git commit -m "docs: add private games design spec and implementation plan"
```

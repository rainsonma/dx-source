# Course Game Detail: Owner + UpdatedAt Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `GET /api/course-games/{id}` include the game owner (`user: {id, username}`) and `updatedAt` so the 课程游戏信息 card on the ai-custom detail page stops showing `?` + `未知` for 作者 and `Invalid Date` for 修改时间.

**Architecture:** Pure backend change. One Go file, one new struct, one function extended. Add an explicit users lookup by primary key — same soft-reference pattern the function already uses for the cover image. Populate two new fields on the JSON response. No frontend changes needed; `mapGameDetail` and `GameInfoCard` already consume these fields.

**Tech Stack:** Go 1.x, Goravel framework, GORM, PostgreSQL.

**Spec:** `docs/superpowers/specs/2026-04-17-course-game-detail-owner-updatedat-design.md`

**Verification model:** No existing `course_game_service_test.go` file — creating new test scaffolding for two new fields is over-engineering (per spec). Verification is `go build ./...`, `go vet ./...`, and manual browser verification against the spec's test plan.

---

## File Structure

Single file modified:

| File | Role |
| ---- | ---- |
| `dx-api/app/services/api/course_game_service.go` | Add `CourseGameOwnerData` type; extend `CourseGameDetailData` with `User` and `UpdatedAt`; populate both in `GetCourseGameDetail`. |

No new files, no model changes, no route changes, no request validation changes, no frontend changes.

---

## Task 1: Extend response struct and populate owner + updatedAt

**Files:**
- Modify: `dx-api/app/services/api/course_game_service.go`

- [ ] **Step 1: Add `CourseGameOwnerData` struct**

Open `dx-api/app/services/api/course_game_service.go`. Locate the `CourseGameLevelData` struct at lines 58–65:

```go
// CourseGameLevelData represents a level in a course game.
type CourseGameLevelData struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    Description *string `json:"description"`
    Order       float64 `json:"order"`
    ItemCount   int64   `json:"itemCount"`
}
```

Immediately after this struct (new lines at line 66+), add the new type:

```go

// CourseGameOwnerData represents the minimal creator info shown on a game detail.
type CourseGameOwnerData struct {
    ID       string `json:"id"`
    Username string `json:"username"`
}
```

The leading blank line preserves one-line separation between top-level declarations, matching the rest of the file.

- [ ] **Step 2: Extend `CourseGameDetailData` with `User` and `UpdatedAt`**

Locate `CourseGameDetailData` at lines 42–56:

```go
// CourseGameDetailData represents a course game with levels for editing.
type CourseGameDetailData struct {
    ID             string                `json:"id"`
    Name           string                `json:"name"`
    Description    *string               `json:"description"`
    Mode           string                `json:"mode"`
    Status         string                `json:"status"`
    IsPrivate      bool                  `json:"isPrivate"`
    GameCategoryID *string               `json:"gameCategoryId"`
    GamePressID    *string               `json:"gamePressId"`
    CoverID        *string               `json:"coverId"`
    CoverURL       *string               `json:"coverUrl"`
    Levels         []CourseGameLevelData `json:"levels"`
    CreatedAt      any                   `json:"createdAt"`
}
```

Replace with:

```go
// CourseGameDetailData represents a course game with levels for editing.
type CourseGameDetailData struct {
    ID             string                `json:"id"`
    Name           string                `json:"name"`
    Description    *string               `json:"description"`
    Mode           string                `json:"mode"`
    Status         string                `json:"status"`
    IsPrivate      bool                  `json:"isPrivate"`
    GameCategoryID *string               `json:"gameCategoryId"`
    GamePressID    *string               `json:"gamePressId"`
    CoverID        *string               `json:"coverId"`
    CoverURL       *string               `json:"coverUrl"`
    Levels         []CourseGameLevelData `json:"levels"`
    User           *CourseGameOwnerData  `json:"user"`
    CreatedAt      any                   `json:"createdAt"`
    UpdatedAt      any                   `json:"updatedAt"`
}
```

Key points:
- `User` is a pointer (nullable in JSON) → matches frontend's `{ id, username } | null` contract and allows graceful `null` when the owner row is missing.
- `UpdatedAt` uses `any` to mirror `CreatedAt`'s existing type, preserving the serialization path the frontend already parses via `new Date(...)`.

- [ ] **Step 3: Fetch the owner inside `GetCourseGameDetail`**

Locate `GetCourseGameDetail` at lines 608–665. The function currently (lines 628–649) looks like:

```go
    // Load cover URL
    var coverURL *string
    if game.CoverID != nil && *game.CoverID != "" {
        var image models.Image
        if err := facades.Orm().Query().Where("id", *game.CoverID).First(&image); err == nil && image.ID != "" {
            coverURL = &image.Url
        }
    }

    levelData := make([]CourseGameLevelData, 0, len(levels))
```

Insert the owner lookup between the cover-URL block and the `levelData` line so the two soft-reference lookups sit together. The new block to insert right before `levelData := ...`:

```go
    // Load owner (soft reference — code-level FK; graceful if missing)
    var owner *CourseGameOwnerData
    if game.UserID != nil && *game.UserID != "" {
        var u models.User
        if err := facades.Orm().Query().Where("id", *game.UserID).First(&u); err == nil && u.ID != "" {
            owner = &CourseGameOwnerData{ID: u.ID, Username: u.Username}
        }
    }

```

So the section becomes:

```go
    // Load cover URL
    var coverURL *string
    if game.CoverID != nil && *game.CoverID != "" {
        var image models.Image
        if err := facades.Orm().Query().Where("id", *game.CoverID).First(&image); err == nil && image.ID != "" {
            coverURL = &image.Url
        }
    }

    // Load owner (soft reference — code-level FK; graceful if missing)
    var owner *CourseGameOwnerData
    if game.UserID != nil && *game.UserID != "" {
        var u models.User
        if err := facades.Orm().Query().Where("id", *game.UserID).First(&u); err == nil && u.ID != "" {
            owner = &CourseGameOwnerData{ID: u.ID, Username: u.Username}
        }
    }

    levelData := make([]CourseGameLevelData, 0, len(levels))
```

- [ ] **Step 4: Populate `User` and `UpdatedAt` in the returned struct**

Locate the existing return (currently lines 651–664):

```go
    return &CourseGameDetailData{
        ID:             game.ID,
        Name:           game.Name,
        Description:    game.Description,
        Mode:           game.Mode,
        Status:         game.Status,
        IsPrivate:      game.IsPrivate,
        GameCategoryID: game.GameCategoryID,
        GamePressID:    game.GamePressID,
        CoverID:        game.CoverID,
        CoverURL:       coverURL,
        Levels:         levelData,
        CreatedAt:      game.CreatedAt,
    }, nil
```

Replace with:

```go
    return &CourseGameDetailData{
        ID:             game.ID,
        Name:           game.Name,
        Description:    game.Description,
        Mode:           game.Mode,
        Status:         game.Status,
        IsPrivate:      game.IsPrivate,
        GameCategoryID: game.GameCategoryID,
        GamePressID:    game.GamePressID,
        CoverID:        game.CoverID,
        CoverURL:       coverURL,
        Levels:         levelData,
        User:           owner,
        CreatedAt:      game.CreatedAt,
        UpdatedAt:      game.UpdatedAt,
    }, nil
```

The field ordering here matches the struct definition's new order (User between Levels and CreatedAt, UpdatedAt at the end), keeping the struct-literal and struct-decl visually aligned.

- [ ] **Step 5: Run `gofmt` / `goimports` check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api
gofmt -l app/services/api/course_game_service.go
```

Expected: no output (file is already formatted). If the file path is printed, run `gofmt -w app/services/api/course_game_service.go` to rewrite, then re-check.

- [ ] **Step 6: Build the whole API module**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api
go build ./...
```

Expected: exit 0, no output (or only clean log lines). If the build fails with "undeclared name" or type mismatch, review Step 2's field list carefully.

- [ ] **Step 7: Run `go vet`**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api
go vet ./...
```

Expected: exit 0, no output. If vet flags struct-tag or field-alignment issues, fix them inline before moving on.

- [ ] **Step 8: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/services/api/course_game_service.go
git commit -m "fix(api): include owner and updatedAt in course-game detail"
```

Verify the commit contains exactly one file:

```bash
git show --stat HEAD
```

Expected: one file changed (`dx-api/app/services/api/course_game_service.go`), with additions for the new type, the two new struct fields, the owner lookup block, and the two new values in the return literal.

---

## Task 2: Manual browser verification

**Files:** none — verification only.

- [ ] **Step 1: Start the API**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api
go run .
```

Wait for the server to report listening on `http://localhost:3001`.

- [ ] **Step 2: Start the web dev server**

In another terminal:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run dev
```

Wait for the dev server to report listening on `http://localhost:3000`.

- [ ] **Step 3: Sign in as a VIP user**

- Open `http://localhost:3000`, sign in with a VIP-enabled account.
- Navigate to `/hall/ai-custom`.

- [ ] **Step 4: Open an existing course game (or create one)**

- If you already have a course game from earlier testing, click its card to open the detail page.
- Otherwise, click "创建课程游戏", fill the form with minimum viable values (name, game mode, category if required), and open the new game's detail page.

- [ ] **Step 5: Verify the 作者 row**

On the 课程游戏信息 card (right side, or bottom on mobile), confirm:
- The avatar letter is the first letter of your username, uppercased (e.g., `R` for "rainson"). NOT `?`.
- The name next to it is your username. NOT `未知`.

- [ ] **Step 6: Verify the 修改时间 row**

Confirm:
- 创建时间 shows a proper `YYYY/MM/DD` date.
- 修改时间 shows a proper `YYYY/MM/DD` date. NOT `Invalid Date`.
- For a freshly created game, 创建时间 and 修改时间 will be the same date — this is correct.

- [ ] **Step 7: Verify 修改时间 advances on edit**

- Click the 编辑 button on the hero card. Change the name or description. Save.
- Reload the detail page (or go back to the list and reopen it).
- Confirm 修改时间 now shows today's date, unchanged-or-advanced depending on whether an edit crossed midnight local time.

- [ ] **Step 8: Regression spot-check**

Confirm the rest of the detail page still works:
- Hero card shows the correct name, mode, category, press.
- Levels panel lists levels.
- Publish/Withdraw buttons still work (no JS error on page load).
- Network tab: `GET /api/course-games/{id}` returns HTTP 200 with `data.user` populated and `data.updatedAt` as an ISO date string.

- [ ] **Step 9: Stop dev servers**

Stop both `go run .` and `npm run dev` processes (Ctrl-C) once verification is complete.

- [ ] **Step 10: No commit — verification task**

Task 2 produces no file changes.

---

## Task 3: Final checks and merge

**Files:** none — repo hygiene.

- [ ] **Step 1: Confirm branch state**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git status
git log --oneline -3
```

Expected:
- Working tree clean.
- Most recent commits include the fix commit from Task 1 (`fix(api): include owner and updatedAt in course-game detail`) and the spec commit (`docs: spec for course-game detail owner and updatedAt fix`).

- [ ] **Step 2: Decide branch strategy**

If Task 1's commit landed directly on `main`, there is nothing to merge — go to Step 4.

If Task 1's commit was made on a feature branch (e.g., `feat/course-game-owner-updatedat`), fast-forward merge to `main`:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git checkout main
git merge --ff-only feat/course-game-owner-updatedat
git branch -d feat/course-game-owner-updatedat
```

Expected: fast-forward merge succeeds, feature branch deleted locally.

- [ ] **Step 3: Final `go build` sanity check from main**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api
go build ./...
```

Expected: exit 0, no output.

- [ ] **Step 4: Report completion**

Summarize to the user:
- What shipped: `/api/course-games/{id}` now includes `user: {id, username}` and `updatedAt`.
- UI effect: the 作者 row on the 课程游戏信息 card now shows the creator's avatar letter + username (not `?` + `未知`); 修改时间 now shows a real date (not `Invalid Date`).
- Verified: `go build`, `go vet`, and manual browser walkthrough including edit-then-reload.
- Where: on `main`, pending push.

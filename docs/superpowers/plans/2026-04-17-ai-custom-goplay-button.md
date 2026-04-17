# AI Custom 去玩 Button Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a "去玩" button on the AI Custom course-game detail hero and on each published card in the AI Custom grid, routing to `/hall/games/{id}`; fix the API so the author of a private published game can view their own game on that page.

**Architecture:** Two small React component changes on the web side. Three coordinated backend changes on the API: widen `GetGameDetail` to take the caller's user ID, relax the `is_private` filter to `(is_private = false OR user_id = ?)`, and move `/api/games/{id}` behind JWT auth so `Auth().ID()` is populated. All callers of `/api/games/{id}` are already authenticated in the frontend, so the route move is a no-op for them.

**Tech Stack:** Next.js 16 (App Router, client components), TailwindCSS, lucide-react; Goravel (Go), GORM, Postgres.

**Spec:** `docs/superpowers/specs/2026-04-17-ai-custom-goplay-button-design.md`

---

## File Structure

Files created/modified, each with a clear single responsibility:

Backend (dx-api):
- `app/services/api/game_service.go` — `GetGameDetail` gains `userID` parameter and an ownership-aware filter.
- `app/http/controllers/api/game_controller.go` — `Detail` reads authenticated user and passes it to the service.
- `routes/api.go` — move `games.Get("/{id}", …)` out of the public group and into the JWT-protected group.
- `app/services/api/game_service_test.go` (new) — compile-time signature pin for `GetGameDetail`.

Frontend (dx-web):
- `src/features/web/ai-custom/components/game-hero-card.tsx` — inserts a primary 去玩 Link button into the action row when the game is published.
- `src/features/web/ai-custom/components/game-card-item.tsx` — renders a second, split layout for published cards so 进入 and 去玩 are two real buttons side by side; non-published cards are untouched.

---

## Task 1: Backend — `GetGameDetail` becomes ownership-aware

**Files:**
- Create: `dx-api/app/services/api/game_service_test.go`
- Modify: `dx-api/app/services/api/game_service.go` (lines 350–360)
- Modify: `dx-api/app/http/controllers/api/game_controller.go` (lines 78–93)

- [ ] **Step 1: Write a compile-time signature pin for the new two-argument `GetGameDetail`**

Create the new test file `dx-api/app/services/api/game_service_test.go`:

```go
package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Pins the expected two-argument signature at compile time so downstream
// callers (controllers) cannot drift from it silently.
func TestGetGameDetailFunctionExists(t *testing.T) {
	assert.NotNil(t, GetGameDetail)
	var _ func(string, string) (*GameDetailData, error) = GetGameDetail
}
```

- [ ] **Step 2: Run the test and confirm it fails to compile**

Run from `dx-api/`:

```bash
go test ./app/services/api/ -run TestGetGameDetailFunctionExists -v
```

Expected: compile error pointing at `game_service_test.go` — something like `cannot use GetGameDetail (value of type func(string) (*GameDetailData, error)) as func(string, string) (*GameDetailData, error) value`. This is the RED phase — the signature mismatch is intentional.

- [ ] **Step 3: Update `GetGameDetail` to take `userID` and relax the `is_private` filter**

In `dx-api/app/services/api/game_service.go`, replace:

```go
// GetGameDetail returns full game detail with levels.
func GetGameDetail(gameID string) (*GameDetailData, error) {
	var game models.Game
	if err := facades.Orm().Query().
		Where("id", gameID).
		Where("status", consts.GameStatusPublished).
		Where("is_active", true).
		Where("is_private", false).
		First(&game); err != nil {
		return nil, fmt.Errorf("failed to find game: %w", err)
	}
```

with:

```go
// GetGameDetail returns full game detail with levels.
// Pass an empty userID for anonymous callers (public games only); when
// userID is populated, private games are also returned if it matches the owner.
func GetGameDetail(gameID string, userID string) (*GameDetailData, error) {
	var game models.Game
	q := facades.Orm().Query().
		Where("id", gameID).
		Where("status", consts.GameStatusPublished).
		Where("is_active", true)
	if userID == "" {
		q = q.Where("is_private", false)
	} else {
		q = q.Where("(is_private = ? OR user_id = ?)", false, userID)
	}
	if err := q.First(&game); err != nil {
		return nil, fmt.Errorf("failed to find game: %w", err)
	}
```

Nothing else in the function changes. The rest of the function (levels, cover, category, press, author lookups) is unaffected.

Notes on the filter:
- The anonymous branch must avoid `user_id = ?` entirely. `user_id` is a Postgres `uuid` column, so a literal `''` fails the cast at parse time (`ERROR: invalid input syntax for type uuid: ""`) before any row is evaluated. Branching on empty `userID` keeps anonymous calls on a public-only predicate.
- When `userID` is a real UUID, the parenthesized `OR` correctly matches either public games or the caller's own games.
- Non-owners of a private game continue to get `ErrGameNotFound` (the guard below the query returns that when `game.ID == ""`).

- [ ] **Step 4: Update `GameController.Detail` to read the authenticated user and pass it in**

In `dx-api/app/http/controllers/api/game_controller.go`, replace the `Detail` method:

```go
// Detail returns full game detail with levels.
func (c *GameController) Detail(ctx contractshttp.Context) contractshttp.Response {
	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id is required")
	}

	detail, err := services.GetGameDetail(gameID)
	if err != nil {
		if errors.Is(err, services.ErrGameNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeGameNotFound, "游戏不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get game detail")
	}

	return helpers.Success(ctx, detail)
}
```

with:

```go
// Detail returns full game detail with levels.
// Ownership-aware: the authenticated user can view their own private published games.
func (c *GameController) Detail(ctx contractshttp.Context) contractshttp.Response {
	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id is required")
	}

	userID, _ := facades.Auth(ctx).Guard("user").ID()

	detail, err := services.GetGameDetail(gameID, userID)
	if err != nil {
		if errors.Is(err, services.ErrGameNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeGameNotFound, "游戏不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get game detail")
	}

	return helpers.Success(ctx, detail)
}
```

No new imports are needed — `facades` is already imported at the top of the file. If the import block somehow lacks `facades`, verify lines 1–14 include `"github.com/goravel/framework/facades"`.

- [ ] **Step 5: Run the test and a full build**

From `dx-api/`:

```bash
go build ./...
go test ./app/services/api/ -run TestGetGameDetailFunctionExists -v
```

Expected: build succeeds; test passes (`ok  dx-api/app/services/api`).

- [ ] **Step 6: Commit**

```bash
cd dx-api
git add app/services/api/game_service.go app/services/api/game_service_test.go app/http/controllers/api/game_controller.go
git commit -m "feat(api): let game author view own private published game detail"
```

---

## Task 2: Backend — Move `/api/games/{id}` behind JWT

**Files:**
- Modify: `dx-api/routes/api.go` (public `/games` group near line 46–56; protected block starts around line 92)

- [ ] **Step 1: Remove the detail route from the public `/games` group**

In `dx-api/routes/api.go`, replace:

```go
		// Public game routes
		gameController := &apicontrollers.GameController{}
		router.Prefix("/games").Group(func(games route.Router) {
			games.Get("/", gameController.List)
			games.Get("/search", gameController.Search)
			games.Get("/{id}", gameController.Detail)
		})
```

with:

```go
		// Public game routes (list + search only; detail is protected below)
		gameController := &apicontrollers.GameController{}
		router.Prefix("/games").Group(func(games route.Router) {
			games.Get("/", gameController.List)
			games.Get("/search", gameController.Search)
		})
```

- [ ] **Step 2: Add the detail route inside the protected block**

Still in `dx-api/routes/api.go`, locate this existing block inside `router.Middleware(middleware.JwtAuth()).Group(func(protected route.Router) { ... })`:

```go
			// Protected game routes
			contentController := &apicontrollers.ContentController{}
			gameStatsController := apicontrollers.NewGameStatsController()
			userFavoriteController := apicontrollers.NewUserFavoriteController()
			protected.Get("/games/played", gameController.Played)
			protected.Get("/games/{id}/stats", gameStatsController.Stats)
			protected.Get("/games/{id}/favorited", userFavoriteController.Favorited)
			protected.Get("/games/{id}/levels/{levelId}/content", contentController.LevelContent)
```

Change it to add the detail route first:

```go
			// Protected game routes
			contentController := &apicontrollers.ContentController{}
			gameStatsController := apicontrollers.NewGameStatsController()
			userFavoriteController := apicontrollers.NewUserFavoriteController()
			protected.Get("/games/{id}", gameController.Detail)
			protected.Get("/games/played", gameController.Played)
			protected.Get("/games/{id}/stats", gameStatsController.Stats)
			protected.Get("/games/{id}/favorited", userFavoriteController.Favorited)
			protected.Get("/games/{id}/levels/{levelId}/content", contentController.LevelContent)
```

Goravel's router evaluates more specific routes (`/games/played`, `/games/{id}/stats`, etc.) independently of registration order because `{id}` is a parameter segment — but keeping the `/{id}` catch-all near the top of its group mirrors existing conventions in the file.

- [ ] **Step 3: Build and verify**

From `dx-api/`:

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
cd dx-api
git add routes/api.go
git commit -m "fix(api): require JWT for GET /api/games/{id}"
```

---

## Task 3: Frontend — Hero 去玩 button

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/components/game-hero-card.tsx` (imports at lines 1–28; action row at lines 164–213)

- [ ] **Step 1: Add `Link` and `Play` imports**

In `dx-web/src/features/web/ai-custom/components/game-hero-card.tsx`, at the top of the file.

Replace:

```tsx
"use client";

import { useState } from "react";
import {
  Send,
  Undo2,
  Pencil,
  Trash2,
  Loader2,
} from "lucide-react";
```

with:

```tsx
"use client";

import { useState } from "react";
import Link from "next/link";
import {
  Send,
  Undo2,
  Pencil,
  Trash2,
  Loader2,
  Play,
} from "lucide-react";
```

- [ ] **Step 2: Insert the 去玩 Link as the first action when published**

Still in `game-hero-card.tsx`. Find the action-buttons block that currently starts with `{canPublish && ( …发布… )}` and then flows into `{isPublished && ( …撤回… )}`. Replace:

```tsx
            {/* Action buttons */}
            <div className="flex flex-wrap items-center gap-3">
              {canPublish && (
                <button
                  type="button"
                  onClick={() => setPublishOpen(true)}
                  className="flex items-center gap-2 rounded-xl bg-gradient-to-b from-teal-500 to-teal-700 px-6 py-2.5"
                >
                  <Send className="h-4 w-4 text-white" />
                  <span className="text-sm font-semibold text-white">
                    发布
                  </span>
                </button>
              )}              {isPublished && (
                <button
                  type="button"
                  onClick={() => setWithdrawOpen(true)}
                  className="flex items-center gap-2 rounded-xl border border-amber-200 bg-amber-50 px-6 py-2.5"
                >
                  <Undo2 className="h-4 w-4 text-amber-600" />
                  <span className="text-sm font-semibold text-amber-600">
                    撤回
                  </span>
                </button>
              )}
```

with:

```tsx
            {/* Action buttons */}
            <div className="flex flex-wrap items-center gap-3">
              {canPublish && (
                <button
                  type="button"
                  onClick={() => setPublishOpen(true)}
                  className="flex items-center gap-2 rounded-xl bg-gradient-to-b from-teal-500 to-teal-700 px-6 py-2.5"
                >
                  <Send className="h-4 w-4 text-white" />
                  <span className="text-sm font-semibold text-white">
                    发布
                  </span>
                </button>
              )}
              {isPublished && (
                <Link
                  href={`/hall/games/${game.id}`}
                  className="flex items-center gap-2 rounded-xl bg-gradient-to-b from-teal-500 to-teal-700 px-6 py-2.5"
                >
                  <Play className="h-4 w-4 text-white" />
                  <span className="text-sm font-semibold text-white">
                    去玩
                  </span>
                </Link>
              )}
              {isPublished && (
                <button
                  type="button"
                  onClick={() => setWithdrawOpen(true)}
                  className="flex items-center gap-2 rounded-xl border border-amber-200 bg-amber-50 px-6 py-2.5"
                >
                  <Undo2 className="h-4 w-4 text-amber-600" />
                  <span className="text-sm font-semibold text-amber-600">
                    撤回
                  </span>
                </button>
              )}
```

The rest of the file (编辑, 删除 buttons, dialogs) stays untouched.

- [ ] **Step 3: Lint**

From `dx-web/`:

```bash
npm run lint
```

Expected: no new warnings or errors in `game-hero-card.tsx`.

- [ ] **Step 4: Commit**

```bash
cd dx-web
git add src/features/web/ai-custom/components/game-hero-card.tsx
git commit -m "feat(web): add 去玩 button on ai-custom detail hero"
```

---

## Task 4: Frontend — Card restructure with 进入 + 去玩

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/components/game-card-item.tsx` (whole file)

Non-published cards keep the current one-big-Link layout exactly. Published cards render a different structure so the two buttons are real interactive elements, not nested anchors.

- [ ] **Step 1: Replace the file content**

Overwrite `dx-web/src/features/web/ai-custom/components/game-card-item.tsx` with:

```tsx
import Link from "next/link";
import { Play } from "lucide-react";

import { GAME_MODE_LABELS, type GameMode } from "@/consts/game-mode";

type GameCard = {
  id: string;
  name: string;
  description?: string | null;
  mode: string;
  status: string;
  createdAt?: string | Date | null;
  coverUrl?: string | null;
  levelCount?: number;
  // Legacy Prisma shape compatibility
  cover?: { url: string } | null;
  _count?: { levels: number };
};

const coverColors = [
  "bg-gradient-to-br from-teal-500 to-emerald-600",
  "bg-gradient-to-br from-indigo-500 to-purple-600",
  "bg-gradient-to-br from-rose-500 to-pink-600",
  "bg-gradient-to-br from-amber-500 to-orange-600",
  "bg-gradient-to-br from-cyan-500 to-blue-600",
  "bg-gradient-to-br from-fuchsia-500 to-violet-600",
];

function pickCoverColor(id: string) {
  let hash = 0;
  for (let i = 0; i < id.length; i++) {
    hash = (hash * 31 + id.charCodeAt(i)) | 0;
  }
  return coverColors[Math.abs(hash) % coverColors.length];
}

type StatusVariant = "published" | "withdraw" | "draft";

const statusStyles: Record<StatusVariant, { bg: string; label: string }> = {
  published: { bg: "bg-green-600", label: "已发布" },
  withdraw: { bg: "bg-amber-600", label: "已撤回" },
  draft: { bg: "bg-slate-500", label: "未发布" },
};

export function GameCardItem({ game, asDiv, onClick }: { game: GameCard; asDiv?: boolean; onClick?: () => void }) {
  const status = (game.status === "published" ? "published" : game.status === "withdraw" ? "withdraw" : "draft") as StatusVariant;
  const s = statusStyles[status];
  const modeLabel = GAME_MODE_LABELS[game.mode as GameMode] ?? game.mode;
  const isPublished = status === "published";
  const detailHref = `/hall/ai-custom/${game.id}`;
  const playHref = `/hall/games/${game.id}`;
  const dateStr = game.createdAt
    ? new Date(game.createdAt).toLocaleDateString("zh-CN", {
        year: "numeric",
        month: "2-digit",
      })
    : "";

  const cover = (
    <div className={`relative flex h-[120px] items-center justify-center ${(game.coverUrl || game.cover?.url) ? "bg-border" : pickCoverColor(game.id)}`}>
      {(game.coverUrl || game.cover?.url) ? (
        /* eslint-disable-next-line @next/next/no-img-element */
        <img
          src={game.coverUrl ?? game.cover?.url ?? ""}
          alt={game.name}
          className="h-full w-full object-cover"
        />
      ) : (
        <span className="text-2xl font-bold text-white/80">{modeLabel}</span>
      )}
      <div className="absolute inset-x-0 top-0 flex h-9 items-start justify-between px-2 pt-2">
        <span
          className={`rounded-md px-2 py-0.5 text-[10px] font-semibold text-white ring-1 ring-white/30 drop-shadow-md ${s.bg}`}
        >
          {s.label}
        </span>
      </div>
    </div>
  );

  const infoBlock = (
    <div className="flex flex-1 flex-col gap-2 p-3 pb-2">
      <div className="flex flex-col gap-1">
        <span className="text-sm font-bold text-foreground">{game.name}</span>
        {game.description && (
          <span className="line-clamp-2 text-[11px] leading-snug text-muted-foreground">
            {game.description}
          </span>
        )}
      </div>
      <div className="flex flex-col gap-0.5">
        <span className="text-[11px] text-muted-foreground">
          {game.levelCount ?? game._count?.levels ?? 0} 个学习单元
        </span>
        <span className="text-[11px] text-muted-foreground">{dateStr}</span>
      </div>
    </div>
  );

  const modeChip = (
    <span className="rounded-[10px] bg-teal-50 px-2 py-0.5 text-[11px] font-medium text-teal-600">
      {modeLabel}
    </span>
  );

  const enterChipContent = (
    <>
      <Play className="h-3 w-3" />
      进入
    </>
  );

  const playChipContent = (
    <>
      <Play className="h-3 w-3" />
      去玩
    </>
  );

  // Non-published: identical to previous behavior — whole card is one Link, footer has a single 进入 visual chip.
  if (!isPublished) {
    const legacyBody = (
      <>
        {cover}
        <div className="flex flex-1 flex-col justify-between gap-2 p-3">
          <div className="flex flex-col gap-1">
            <span className="text-sm font-bold text-foreground">{game.name}</span>
            {game.description && (
              <span className="line-clamp-2 text-[11px] leading-snug text-muted-foreground">
                {game.description}
              </span>
            )}
          </div>
          <div className="flex flex-col gap-0.5">
            <span className="text-[11px] text-muted-foreground">
              {game.levelCount ?? game._count?.levels ?? 0} 个学习单元
            </span>
            <span className="text-[11px] text-muted-foreground">{dateStr}</span>
          </div>
          <div className="flex items-center justify-between">
            {modeChip}
            <span className="flex items-center gap-1 rounded-md bg-teal-600 px-3 py-1 text-[11px] font-semibold text-white">
              {enterChipContent}
            </span>
          </div>
        </div>
      </>
    );

    if (asDiv) {
      return (
        <div
          onClick={onClick}
          className="flex cursor-pointer flex-col overflow-hidden rounded-xl border border-border bg-card transition-shadow hover:shadow-md"
        >
          {legacyBody}
        </div>
      );
    }

    return (
      <Link
        href={detailHref}
        className="flex flex-col overflow-hidden rounded-xl border border-border bg-card transition-shadow hover:shadow-md"
      >
        {legacyBody}
      </Link>
    );
  }

  // Published: split layout — cover + info wrap in an inner Link to the detail page,
  // footer has two real interactive buttons (进入 → detail, 去玩 → play). Non-VIP (asDiv)
  // renders buttons without their own handlers so the outer onClick bubbles.
  const enterButtonVip = (
    <Link
      href={detailHref}
      className="flex items-center gap-1 rounded-md bg-teal-50 px-3 py-1 text-[11px] font-semibold text-teal-600"
    >
      {enterChipContent}
    </Link>
  );

  const playButtonVip = (
    <Link
      href={playHref}
      className="flex items-center gap-1 rounded-md bg-teal-600 px-3 py-1 text-[11px] font-semibold text-white"
    >
      {playChipContent}
    </Link>
  );

  const enterButtonGated = (
    <button
      type="button"
      className="flex items-center gap-1 rounded-md bg-teal-50 px-3 py-1 text-[11px] font-semibold text-teal-600"
    >
      {enterChipContent}
    </button>
  );

  const playButtonGated = (
    <button
      type="button"
      className="flex items-center gap-1 rounded-md bg-teal-600 px-3 py-1 text-[11px] font-semibold text-white"
    >
      {playChipContent}
    </button>
  );

  const footer = (
    <div className="flex items-center justify-between px-3 pb-3">
      {modeChip}
      <div className="flex items-center gap-2">
        {asDiv ? enterButtonGated : enterButtonVip}
        {asDiv ? playButtonGated : playButtonVip}
      </div>
    </div>
  );

  if (asDiv) {
    return (
      <div
        onClick={onClick}
        className="flex cursor-pointer flex-col overflow-hidden rounded-xl border border-border bg-card transition-shadow hover:shadow-md"
      >
        {cover}
        {infoBlock}
        {footer}
      </div>
    );
  }

  return (
    <div className="flex flex-col overflow-hidden rounded-xl border border-border bg-card transition-shadow hover:shadow-md">
      <Link href={detailHref} className="flex flex-1 flex-col">
        {cover}
        {infoBlock}
      </Link>
      {footer}
    </div>
  );
}
```

Reasoning notes embedded above via comments. Non-published path keeps the exact original `flex-1 flex-col justify-between gap-2 p-3` layout so its visual output is unchanged.

- [ ] **Step 2: Lint**

From `dx-web/`:

```bash
npm run lint
```

Expected: clean — no new errors or warnings on `game-card-item.tsx`.

- [ ] **Step 3: Commit**

```bash
cd dx-web
git add src/features/web/ai-custom/components/game-card-item.tsx
git commit -m "feat(web): add 去玩 button next to 进入 on published ai-custom cards"
```

---

## Task 5: End-to-end manual verification

No automated UI tests exist in this project for these components. Verify the golden path and edge cases by hand.

**Files:** none (runtime verification only).

- [ ] **Step 1: Boot both servers**

Two terminals:

```bash
cd dx-api && go run .      # port 3001
```

```bash
cd dx-web && npm run dev   # port 3000
```

- [ ] **Step 2: Log in as a VIP user with at least one course-game in each status**

Sign in. Visit `http://localhost:3000/hall/ai-custom`.

If you don't already have coverage, create three course-games to exercise each branch:
1. One public, published (use an existing game or publish a new one with the 私有 switch OFF)
2. One private, published (publish a new game with 私有 switch ON)
3. One draft (just create it, don't publish)

- [ ] **Step 3: Verify the grid page (`/hall/ai-custom`)**

- Draft card → whole card clickable, lands on `/hall/ai-custom/{id}`. Footer shows the single 进入 chip. **No 去玩 button.**
- Withdrawn card → same as draft: single 进入 chip, no 去玩.
- Published (public) card → footer shows two chips: 进入 + 去玩. Clicking body → detail page. Clicking 进入 → detail page. Clicking 去玩 → `/hall/games/{id}` loads normally.
- Published (private) card → same layout as public published. Clicking 去玩 → `/hall/games/{id}` loads normally **for the owner** (no 404).
- Open a DevTools incognito window, log in as a **different** user, and visit `/hall/games/{privateGameId}` directly → expect "游戏不存在" / 404 (private game not leaked).

- [ ] **Step 4: Verify the hero page (`/hall/ai-custom/{id}`)**

- Draft / withdrawn → no 去玩 button in the hero action row.
- Published (public or private, as owner) → hero shows `去玩 → 撤回 → 编辑 (disabled)`. Click 去玩 → `/hall/games/{id}` loads.

- [ ] **Step 5: Verify non-VIP gating on published cards**

Impersonate a non-VIP user (or temporarily flip VIP off in the DB for your account). Reload `/hall/ai-custom`. On a published card:
- Click body → upgrade dialog opens.
- Click 进入 button → upgrade dialog opens.
- Click 去玩 button → upgrade dialog opens.

No navigation should happen; the parent onClick fires for all clicks because 进入/去玩 render as plain `<button>` elements with no handlers of their own.

- [ ] **Step 6: Final sanity sweep**

```bash
cd dx-api && go build ./... && go test ./app/services/api/ -run TestGetGameDetailFunctionExists -v
cd dx-web && npm run lint
```

All three green.

- [ ] **Step 7: If manual verification surfaces an issue, do not silently amend**

Fix the underlying cause, add or extend a test if applicable, commit a new corrective commit. Do not mark this plan complete while any of the above checks fails.

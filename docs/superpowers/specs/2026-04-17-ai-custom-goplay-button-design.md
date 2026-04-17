# AI Custom 去玩 Button

## Summary

Add a "去玩" (Go Play) entry point from a user's own course-games into the regular game play page (`/hall/games/{id}`) once the game is published. Two surfaces get the button:

1. Hero section on `/hall/ai-custom/[id]` (course-game detail page)
2. Footer of each published card on `/hall/ai-custom` (grid page)

Also fix a pre-existing gap: the author cannot currently view their own private published course-game on `/hall/games/{id}` because `/api/games/{id}` unconditionally filters out private games. This contradicts the intent recorded in `2026-04-16-private-games-design.md`, which specifies that only **non-owner access** to private games should be blocked.

## Scope

- **In scope:** the three changes above. Frontend UI on two ai-custom surfaces; backend ownership-aware detail query.
- **Out of scope:** any change to public list/search behavior (`ListPublishedGames`, `SearchGames`, `GetPlayedGames` keep `is_private = false` filtering). Any redesign of the play page itself.

## Behavior

- "published ready for playing" means `status === "published"`. Existing publish validation already enforces active levels, content items, generated items, and vocab batch-size rules — `status === "published"` is a sufficient readiness signal.
- 去玩 always navigates to `/hall/games/{game.id}` (Option A: the existing public detail/play page — rules, stats, "start play" CTA). Not a direct session start.

## Frontend (dx-web)

### `features/web/ai-custom/components/game-hero-card.tsx`

When `isPublished`, render a primary-CTA `<Link>` as the **first** button in the existing action-row flex:

- `href={` `/hall/games/${game.id}` `}`
- Style: `flex items-center gap-2 rounded-xl bg-gradient-to-b from-teal-500 to-teal-700 px-6 py-2.5` (matches the existing 发布 button styling)
- Icon: `Play` from `lucide-react` (`h-4 w-4 text-white`)
- Label: `去玩` (`text-sm font-semibold text-white`)

Action-row order when published: **去玩 → 撤回 → 编辑** (编辑 stays disabled).

### `features/web/ai-custom/components/game-card-item.tsx`

Current layout (unchanged for draft/withdraw): whole card is one `<Link>` to `/hall/ai-custom/{id}`; the `进入` chip is a visual-only `<span>`.

For **published** cards, restructure:

- Outer wrapper becomes a `<div>` (no outer `<Link>`), same classes otherwise. Nested `<a>` elements would be invalid HTML once two Link buttons sit inside.
- Cover + body block wrapped in an inner `<Link href="/hall/ai-custom/{id}">` so the "enter detail" area remains clickable the same way.
- Footer row keeps the mode chip on the left and replaces the single `进入` span with two small real buttons on the right, sharing the same footprint so they read as a paired pill set:
  - `进入` — `<Link href="/hall/ai-custom/{id}">`, `flex items-center gap-1 rounded-md bg-teal-50 px-3 py-1 text-[11px] font-semibold text-teal-600`, `Play` icon (`h-3 w-3`).
  - `去玩` — `<Link href="/hall/games/{id}">`, `flex items-center gap-1 rounded-md bg-teal-600 px-3 py-1 text-[11px] font-semibold text-white`, `Play` icon (`h-3 w-3`). Matches the current single-chip `进入` footprint so the CTA remains familiar.

Non-VIP (`asDiv` mode): when rendered as a gated `<div onClick={openUpgrade}>`, render 进入 and 去玩 as `<button type="button">` elements without their own handlers so the parent `onClick` still fires. No direct navigation happens until the user upgrades.

Draft and withdraw cards render exactly as today — no extra button, outer Link preserved.

## Backend (dx-api)

### `routes/api.go`

Move `games.Get("/{id}", gameController.Detail)` out of the public `/games` group and into the JWT-protected block. Keep `games.Get("/", ...)` (list) and `games.Get("/search", ...)` public.

All existing callers of `/api/games/{id}` already send the `dx_token` cookie (verified: every reference in `dx-web/src` uses `apiClient.get`, which attaches the cookie). No anonymous caller exists, so no behavior regression.

### `app/http/controllers/api/game_controller.go`

`GameController.Detail` reads the authenticated user:

```go
userID, _ := facades.Auth(ctx).Guard("user").ID()
```

Passes it down to the service. The JWT middleware guarantees auth, so the ID is always populated for this route — but the service still treats an empty userID as anonymous for safety.

### `app/services/api/game_service.go`

Change signature:

```go
func GetGameDetail(gameID string, userID string) (*GameDetailData, error)
```

Branch the filter on empty `userID`:

```go
q := facades.Orm().Query().
    Where("id", gameID).
    Where("status", consts.GameStatusPublished).
    Where("is_active", true)
if userID == "" {
    q = q.Where("is_private", false)
} else {
    q = q.Where("(is_private = ? OR user_id = ?)", false, userID)
}
```

The anonymous branch must avoid binding `user_id = ?` because `user_id` is a Postgres `uuid` column — passing `""` fails the cast at parse time (`ERROR: invalid input syntax for type uuid: ""`) before any row is evaluated, returning 500 to every anonymous caller. Branching keeps anonymous callers on the public-only predicate. Authenticated callers (UUID populated) get the `OR` clause that surfaces public games plus their own games.

Non-owners of a private game still get `ErrGameNotFound` — no info leak.

Other query paths (`ListPublishedGames`, `SearchGames`, `GetPlayedGames`) are unchanged.

## Data Flow

```
[published ai-custom card / hero]
        │  去玩
        ▼
/hall/games/{id}
        │  GET /api/games/{id}  (JWT)
        ▼
GameController.Detail
        │  userID = Auth().ID()
        ▼
services.GetGameDetail(gameID, userID)
        │  WHERE status=published AND is_active AND (is_private=false OR user_id=userID)
        ▼
returns detail | ErrGameNotFound
```

## Error Handling

- Frontend: plain `<Link>` navigation; no new client-side failure modes.
- Backend: unchanged error codes. `ErrGameNotFound` → `CodeGameNotFound` (404). DB errors → `CodeInternalError` (500).
- Ownership check runs inside SQL — no TOCTOU window.

## Testing

- Go: add an existence stub for `GetGameDetail` in a new `game_service_test.go` (matches existing "function-exists" style used in `group_service_test.go`), confirming the two-arg signature compiles.
- `go build ./...` clean.
- `npm run lint` clean.

Manual verification checklist (browser):
- Draft card → unchanged: whole card click → ai-custom detail. No 去玩.
- Withdraw card → unchanged. No 去玩.
- Published **public** card → body click → ai-custom detail; 进入 → ai-custom detail; 去玩 → `/hall/games/{id}` loads normally.
- Published **private** card (as owner) → 去玩 → `/hall/games/{id}` loads normally (previously 404).
- Published private game opened by a different user via direct URL → 404 (unchanged).
- Hero 去玩 visible only when `isPublished`; click → `/hall/games/{id}`.
- Non-VIP on published cards → both 进入 and 去玩 trigger the upgrade dialog (not navigation).

## Files Touched

Frontend:
- `dx-web/src/features/web/ai-custom/components/game-hero-card.tsx`
- `dx-web/src/features/web/ai-custom/components/game-card-item.tsx`

Backend:
- `dx-api/routes/api.go`
- `dx-api/app/http/controllers/api/game_controller.go`
- `dx-api/app/services/api/game_service.go`
- `dx-api/app/services/api/game_service_test.go` (new, existence stub)

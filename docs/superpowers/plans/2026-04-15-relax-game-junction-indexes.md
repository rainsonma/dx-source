# Relax Game Junction Unique Indexes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert the two partial unique indexes on `game_metas` / `game_items` to non-unique, 3-column indexes keyed on `(game_level_id, content_*_id, deleted_at)`, enabling M:N between levels and content, and fix the latent reorder-path bug that the uniqueness guarantee was hiding.

**Architecture:** Pure dx-api/dx-web code + migration cleanup. Split into two commits: (1) reorder fix across 6 files (backend service/controller/request + frontend api-client/action/component), (2) schema relaxation across 3 files (new migration + edit existing migration + register in bootstrap). After the second commit, the user takes a pre-migration `pg_dump` snapshot and runs `go run . artisan migrate` on the live DB to apply the index swap, followed by manual smoke tests.

**Tech Stack:** Go 1.22, Goravel framework, PostgreSQL; Next.js 16 + TypeScript on the frontend.

---

## Context for the Implementer

### What you're doing at a high level

The junction tables `game_metas` and `game_items` currently have partial UNIQUE indexes `(game_level_id, content_*_id) WHERE deleted_at IS NULL`. This forbids the same `content_meta` / `content_item` from being linked to the same level more than once — which blocks both cross-level content reuse (in principle, works today) AND in-level repetition (forbidden today). The user wants M:N.

Relaxing the unique indexes exposes one latent bug: `ReorderMetadata` and `ReorderContentItems` in `dx-api/app/services/api/course_content_service.go` issue UPDATE statements scoped only by `content_*_id`, not by `game_level_id`. Under the old unique constraint, at most one junction row per `(level, content)` existed, so each UPDATE matched exactly one row by accident. The moment M:N data exists (any duplicate junction row), those UPDATEs would silently rewrite `"order"` on every level that references the same content, trampling unrelated level state.

So the work is: (1) fix reorder to include `game_level_id` in the WHERE clause, threading the level through the service function signature and — for the content-item reorder path — adding the field to the request DTO and frontend payload since it's currently missing there; (2) replace the partial unique indexes with non-unique, non-partial 3-column indexes.

**Order matters.** The reorder fix must land *first*. If commits land in reverse order, there's a theoretical window where the unique constraint is dropped but reorder is still latently broken. No M:N data exists during that window, so no actual corruption, but defensively the reorder fix is a precondition for the schema relaxation.

### What you will NOT touch (explicitly out of scope)

- `DeleteContentItem` / `DeleteMetadata` / `DeleteAllLevelContent` — these currently delete junctions by `content_*_id` globally, which is correct on today's 1:1 data but will over-delete once M:N data exists. Rewriting them with reference counting is the dedup task's job (`docs/superpowers/specs/2026-04-14-content-meta-dedup-design.md`). Touching them here would bleed scope.
- `SaveMetadataBatch` dedup logic — same, owned by the dedup task.
- Frontend delete-route URL changes (`/levels/{levelId}/`) — same.
- `idx_game_metas_level_order` / `idx_game_items_level_order` — already non-unique, they serve a different query pattern with `deleted_at` in the middle of the key specifically for index range scan + ORDER BY. Leave them alone.
- `dx-api/app/models/game_meta.go` and `game_item.go` — no struct changes needed; removing the `deleted_at` predicate from the index does not change the model's fields.

### Working directory

All absolute paths in this plan assume the repo root is `/Users/rainsen/Programs/Projects/douxue/dx-source`. When a command says `cd dx-api`, assume it's `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api`.

### Branch

The user's convention is to work locally on `main` for small changes (per `feedback_git_workflow.md` memory). Stay on `main`. Do not create a feature branch unless the user explicitly asks.

### TDD note

This plan does not include automated tests. The repo currently has no tests for `game_metas` / `game_items`, and adding a test suite for this one bug fix would be out of scope (the dedup task will add the feature test file). Verification is via build + vet + fresh-migrate + manual smoke tests.

---

## File Structure

### Phase 1 — Reorder fix (6 files, 1 commit)

| File | Change type |
|---|---|
| `dx-api/app/services/api/course_content_service.go` | edit (2 functions) |
| `dx-api/app/http/controllers/api/course_game_controller.go` | edit (2 call sites) |
| `dx-api/app/http/requests/api/course_game_request.go` | edit (1 struct) |
| `dx-web/src/lib/api-client.ts` | edit (1 function) |
| `dx-web/src/features/web/ai-custom/actions/course-game.action.ts` | edit (1 action) |
| `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx` | edit (1 call site) |

### Phase 2 — Schema relaxation (3 files, 1 commit)

| File | Change type |
|---|---|
| `dx-api/database/migrations/20260414000002_add_game_junction_partial_indexes.go` | edit (Up + Down) |
| `dx-api/database/migrations/20260415000001_relax_game_junction_indexes.go` | new |
| `dx-api/bootstrap/migrations.go` | edit (register new migration) |

### Phase 3 — Migration run and verification on live DB (no file changes, user-executed)

Pre-migration backup → `go run . artisan migrate` → schema inspection → row-count invariants → reorder smoke test → single-play smoke test.

---

## Task 1: Fix ReorderMetadata — scope UPDATE by game_level_id

**Rationale:** The service function `ReorderMetadata` currently issues `UPDATE game_metas SET "order" = ? WHERE content_meta_id = ? AND deleted_at IS NULL`, which under the current unique constraint matches exactly one row by accident. Adding `game_level_id = ?` to the WHERE clause makes it correct under any junction cardinality. On today's 1:1 data, the added clause matches the same one row — no behavior regression.

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go:169-195`

- [ ] **Step 1: Edit `ReorderMetadata` — add `gameLevelID` parameter, scope UPDATE**

Replace the entire function (lines 168-195 in the current file) with:

```go
// ReorderMetadata updates the order of a content metadata entry within a level.
func ReorderMetadata(userID, gameID, gameLevelID, metaID string, newOrder float64) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return err
	}

	if game.Status == consts.GameStatusPublished {
		return ErrGamePublished
	}

	// Verify meta belongs to this game
	if err := verifyMetaBelongsToGame(metaID, gameID); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`UPDATE game_metas SET "order" = ?
		   WHERE game_level_id = ? AND content_meta_id = ? AND deleted_at IS NULL`,
		newOrder, gameLevelID, metaID,
	); err != nil {
		return fmt.Errorf("failed to reorder metadata: %w", err)
	}

	return nil
}
```

Changes from the original:
- Signature: `ReorderMetadata(userID, gameID, metaID string, newOrder float64)` → `ReorderMetadata(userID, gameID, gameLevelID, metaID string, newOrder float64)`.
- Comment: `"updates the order of a content metadata entry."` → `"updates the order of a content metadata entry within a level."`.
- SQL: added `game_level_id = ?` to the WHERE clause.
- Exec args: added `gameLevelID` as the second-to-last argument (Postgres positional binding order: `newOrder`, `gameLevelID`, `metaID`).

Note: `go build ./...` will fail at the end of this task because the controller still calls the old 4-arg signature. That's expected — Task 2 fixes it.

---

## Task 2: Fix ReorderContentItems — scope UPDATE by game_level_id

**Rationale:** Same pattern as Task 1, for the sibling function that reorders content items.

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go:438-464`

- [ ] **Step 1: Edit `ReorderContentItems` — add `gameLevelID` parameter, scope UPDATE**

Replace the entire function (lines 438-464 in the current file) with:

```go
// ReorderContentItems updates the order of a content item within a level.
func ReorderContentItems(userID, gameID, gameLevelID, itemID string, newOrder float64) error {
	if err := requireVip(userID); err != nil {
		return err
	}
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return err
	}

	if game.Status == consts.GameStatusPublished {
		return ErrGamePublished
	}

	if err := verifyItemBelongsToGame(itemID, gameID); err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Exec(
		`UPDATE game_items SET "order" = ?
		   WHERE game_level_id = ? AND content_item_id = ? AND deleted_at IS NULL`,
		newOrder, gameLevelID, itemID,
	); err != nil {
		return fmt.Errorf("failed to reorder content item: %w", err)
	}

	return nil
}
```

Changes from the original:
- Signature: `ReorderContentItems(userID, gameID, itemID string, newOrder float64)` → `ReorderContentItems(userID, gameID, gameLevelID, itemID string, newOrder float64)`.
- Comment: `"updates the order of a content item."` → `"updates the order of a content item within a level."`.
- SQL: added `game_level_id = ?` to the WHERE clause.
- Exec args: added `gameLevelID` as the second-to-last argument.

---

## Task 3: Update controller to thread gameLevelID for both reorder endpoints

**Rationale:** The service functions now require `gameLevelID`. The controllers already have access to it: `ReorderMetadataRequest` already carries `GameLevelID` (line 168 of the request file), and Task 4 will add the same field to `ReorderContentItemRequest`. This task just plumbs the value through both controller functions.

**Files:**
- Modify: `dx-api/app/http/controllers/api/course_game_controller.go:294` and `:410`

- [ ] **Step 1: Edit `ReorderMetadata` controller call site (around line 294)**

Find this line:

```go
	if err := services.ReorderMetadata(userID, gameID, req.MetaID, req.NewOrder); err != nil {
```

Replace with:

```go
	if err := services.ReorderMetadata(userID, gameID, req.GameLevelID, req.MetaID, req.NewOrder); err != nil {
```

No other changes to the controller function — `req.GameLevelID` was already being validated by the request struct (see `course_game_request.go` lines 168 and 177).

- [ ] **Step 2: Edit `ReorderContentItems` controller call site (around line 410)**

Find this line:

```go
	if err := services.ReorderContentItems(userID, gameID, req.ItemID, req.NewOrder); err != nil {
```

Replace with:

```go
	if err := services.ReorderContentItems(userID, gameID, req.GameLevelID, req.ItemID, req.NewOrder); err != nil {
```

Note: at this point the build is still broken — `req.GameLevelID` does not yet exist on `ReorderContentItemRequest`. Task 4 adds it.

---

## Task 4: Add GameLevelID to ReorderContentItemRequest

**Rationale:** The request struct currently has only `ItemID` and `NewOrder`. Adding `GameLevelID` with matching validation rules mirrors the existing `ReorderMetadataRequest` (same file, lines 167-190).

**Files:**
- Modify: `dx-api/app/http/requests/api/course_game_request.go:260-279`

- [ ] **Step 1: Replace `ReorderContentItemRequest` and its methods**

Find the struct and its three methods (approximately lines 258-279 in the current file):

```go
// ---------- ReorderContentItemRequest ----------

type ReorderContentItemRequest struct {
	ItemID   string  `json:"itemId"`
	NewOrder float64 `json:"newOrder"`
}

func (r *ReorderContentItemRequest) Authorize(ctx http.Context) error { return nil }
func (r *ReorderContentItemRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"itemId":   "required|uuid",
		"newOrder": "required|min:0",
	}
}
func (r *ReorderContentItemRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"itemId.required":   "请指定内容项",
		"itemId.uuid":       "无效的内容项ID",
		"newOrder.required": "请指定排序位置",
		"newOrder.min":      "排序位置不能为负数",
	}
}
```

Replace with:

```go
// ---------- ReorderContentItemRequest ----------

type ReorderContentItemRequest struct {
	GameLevelID string  `json:"gameLevelId"`
	ItemID      string  `json:"itemId"`
	NewOrder    float64 `json:"newOrder"`
}

func (r *ReorderContentItemRequest) Authorize(ctx http.Context) error { return nil }
func (r *ReorderContentItemRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"gameLevelId": "required|uuid",
		"itemId":      "required|uuid",
		"newOrder":    "required|min:0",
	}
}
func (r *ReorderContentItemRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"gameLevelId.required": "请指定关卡",
		"gameLevelId.uuid":     "无效的关卡ID",
		"itemId.required":      "请指定内容项",
		"itemId.uuid":          "无效的内容项ID",
		"newOrder.required":    "请指定排序位置",
		"newOrder.min":         "排序位置不能为负数",
	}
}
```

Changes:
- Struct: added `GameLevelID string \`json:"gameLevelId"\`` as the first field (matches the visual layout of `ReorderMetadataRequest` lines 167-171).
- Rules: added `"gameLevelId": "required|uuid"` (matches `ReorderMetadataRequest` line 177).
- Messages: added the two `gameLevelId` validation messages in Chinese (matches `ReorderMetadataRequest` lines 185-186).

---

## Task 5: Verify backend build is clean after reorder fix

**Rationale:** After Tasks 1-4, the three backend files are internally consistent. Running the compiler catches any typo or missed call site before moving to the frontend.

**Files:** none (verification only)

- [ ] **Step 1: Build check**

Run:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: exits 0 with no output.

If a compile error surfaces (for example: `cannot use req.GameLevelID (type ... undefined)` or `too few arguments in call to services.ReorderMetadata`), re-check Tasks 1-4 for whichever file is flagged. The most likely misses are:
- Task 3 Step 1 or Step 2 missed one of the two call sites → add the argument.
- Task 4 didn't get saved → re-apply the replacement.

- [ ] **Step 2: Vet check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./...
```

Expected: exits 0 with no output.

---

## Task 6: Update frontend api-client — send gameLevelId in reorderContentItems payload

**Rationale:** The backend now rejects requests missing `gameLevelId`. The frontend's api-client wrapper currently builds the payload from only `itemId` and `newOrder`; extend it to accept and forward `gameLevelId`. TypeScript will enforce that every caller supplies the new field.

**Files:**
- Modify: `dx-web/src/lib/api-client.ts:654-657`

- [ ] **Step 1: Edit `reorderContentItems` method**

Find this method (around lines 653-657):

```typescript
  /** Reorder content items */
  async reorderContentItems(gameId: string, data: { itemId: string; newOrder: number }) {
    return apiClient.put<null>(`/api/course-games/${gameId}/content-items/reorder`, data);
  },
```

Replace with:

```typescript
  /** Reorder content items */
  async reorderContentItems(
    gameId: string,
    data: { gameLevelId: string; itemId: string; newOrder: number }
  ) {
    return apiClient.put<null>(`/api/course-games/${gameId}/content-items/reorder`, data);
  },
```

Changes:
- `data` parameter type: added `gameLevelId: string` as the first field.
- Method body unchanged — `data` is forwarded verbatim, and adding a field to it automatically includes it in the JSON payload.

---

## Task 7: Update reorderItemAction to accept and pass gameLevelId

**Rationale:** The server action wrapper used by React components needs the new field. Placing `gameLevelId` right after `gameId` in the argument list matches the existing `reorderMetaAction` shape.

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/actions/course-game.action.ts:249-265`

- [ ] **Step 1: Edit `reorderItemAction`**

Find this function (around lines 249-265):

```typescript
/** Reorder a content item via Go API. */
export async function reorderItemAction(
  gameId: string,
  itemId: string,
  newOrder: number
): Promise<SimpleActionResult> {
  try {
    const res = await apiClient.put<null>(
      `/api/course-games/${gameId}/content-items/reorder`,
      { itemId, newOrder }
    );
    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "排序失败" };
  }
}
```

Replace with:

```typescript
/** Reorder a content item via Go API. */
export async function reorderItemAction(
  gameId: string,
  gameLevelId: string,
  itemId: string,
  newOrder: number
): Promise<SimpleActionResult> {
  try {
    const res = await apiClient.put<null>(
      `/api/course-games/${gameId}/content-items/reorder`,
      { gameLevelId, itemId, newOrder }
    );
    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "排序失败" };
  }
}
```

Changes:
- Added `gameLevelId: string` as the second positional argument (after `gameId`, matching `reorderMetaAction` at line 209-211 which uses the same `gameId, gameLevelId, ...` order).
- Payload object: added `gameLevelId` as the first property.

Note: `npm run lint` will fail at the end of this task because `level-units-panel.tsx` still calls the old 3-arg signature. That's expected — Task 8 fixes it.

---

## Task 8: Update level-units-panel call site to pass levelId

**Rationale:** The component already has `levelId` in scope as a prop (used at line 205 when calling `reorderMetaAction`). Just passing it along to `reorderItemAction` makes the two call shapes consistent.

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx:344-348`

- [ ] **Step 1: Edit the `reorderItemAction` call site**

Find this call (around lines 344-348):

```typescript
        const result = await reorderItemAction(
          gameId,
          active.id as string,
          newOrder
```

Replace with:

```typescript
        const result = await reorderItemAction(
          gameId,
          levelId,
          active.id as string,
          newOrder
```

Changes:
- Added `levelId` as the second argument.
- The closing `);` on the next line and the surrounding try/catch block stay exactly as they were.

Note: `levelId` is a component prop (the same one passed to `reorderMetaAction` at line 205), so it's already in scope in this function body. No import changes needed.

---

## Task 9: Verify frontend lint is clean after reorder fix

**Rationale:** Catches any missed call site via the TypeScript type system.

**Files:** none (verification only)

- [ ] **Step 1: Lint check**

Run:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run lint
```

Expected: exits 0 with no errors. Warnings are OK (if the project has pre-existing lint warnings, they're not introduced by this task).

If the lint fails with a type error like `Expected 4 arguments, but got 3` or `Type '{ itemId: string; newOrder: number }' is missing the following properties from type '{ gameLevelId: string; ... }'`, re-check Tasks 6-8 for whichever file is flagged.

---

## Task 10: Commit the reorder fix

**Rationale:** All 6 files are in a consistent end state, backend builds, frontend lints. This is the right moment to commit Phase 1 as a single atomic change.

**Files:** all 6 from Tasks 1-8.

- [ ] **Step 1: Stage and commit**

Run from the repo root (`/Users/rainsen/Programs/Projects/douxue/dx-source`):

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && \
git add dx-api/app/services/api/course_content_service.go \
        dx-api/app/http/controllers/api/course_game_controller.go \
        dx-api/app/http/requests/api/course_game_request.go \
        dx-web/src/lib/api-client.ts \
        dx-web/src/features/web/ai-custom/actions/course-game.action.ts \
        dx-web/src/features/web/ai-custom/components/level-units-panel.tsx && \
git commit -m "$(cat <<'EOF'
fix(api): scope reorder paths to game_level_id for M:N safety

Both ReorderMetadata and ReorderContentItems issued UPDATEs scoped
only by content_*_id, which worked by accident under the partial
unique index that forbade more than one junction row per
(level, content). Scoping the UPDATE to game_level_id too makes
reorder correct under the upcoming M:N relaxation.

Also threads gameLevelId through the content-item reorder request
(which previously lacked the field) and its frontend callers.
The meta reorder path already carried gameLevelId end-to-end; only
the controller needed to thread it into the service.

No user-visible behavior change on current 1:1 data — the added
game_level_id = ? matches the same one junction row that the old
content_*_id = ? alone matched.
EOF
)"
```

Expected: single commit created. `git log -1 --stat` should list exactly the 6 files.

- [ ] **Step 2: Verify the commit**

Run:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git log -1 --stat
```

Expected output includes: the commit subject `fix(api): scope reorder paths to game_level_id for M:N safety`, and a file list showing exactly 6 files touched (3 backend + 3 frontend), all marked as modifications (no additions or deletions of whole files).

---

## Task 11: Edit migration 20260414000002 to use the new index shape

**Rationale:** The existing migration `20260414000002_add_game_junction_partial_indexes.go` was last written (commit `6e7e36d`) with the partial unique indexes. Editing it now produces the final (non-unique, 3-column) shape directly on fresh-migrate paths, while leaving live-DB paths unaffected (Goravel tracks applied migrations by signature and won't re-run it). The `Down()` function also updates to match the new `Up()` so rollback on a fresh DB drops the new names.

**Files:**
- Modify: `dx-api/database/migrations/20260414000002_add_game_junction_partial_indexes.go:13-41`

- [ ] **Step 1: Replace both `Up()` and `Down()` bodies**

Find the `Up()` and `Down()` functions (current content below):

```go
func (r *M20260414000002AddGameJunctionPartialIndexes) Up() error {
	indexes := []string{
		`CREATE UNIQUE INDEX idx_game_metas_level_meta_unique ON game_metas (game_level_id, content_meta_id) WHERE deleted_at IS NULL`,
		`CREATE INDEX idx_game_metas_level_order ON game_metas (game_level_id, deleted_at, "order")`,
		`CREATE UNIQUE INDEX idx_game_items_level_item_unique ON game_items (game_level_id, content_item_id) WHERE deleted_at IS NULL`,
		`CREATE INDEX idx_game_items_level_order ON game_items (game_level_id, deleted_at, "order")`,
	}
	for _, sql := range indexes {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260414000002AddGameJunctionPartialIndexes) Down() error {
	indexes := []string{
		`DROP INDEX IF EXISTS idx_game_items_level_order`,
		`DROP INDEX IF EXISTS idx_game_items_level_item_unique`,
		`DROP INDEX IF EXISTS idx_game_metas_level_order`,
		`DROP INDEX IF EXISTS idx_game_metas_level_meta_unique`,
	}
	for _, sql := range indexes {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
```

Replace with:

```go
func (r *M20260414000002AddGameJunctionPartialIndexes) Up() error {
	indexes := []string{
		`CREATE INDEX idx_game_metas_level_meta ON game_metas (game_level_id, content_meta_id, deleted_at)`,
		`CREATE INDEX idx_game_metas_level_order ON game_metas (game_level_id, deleted_at, "order")`,
		`CREATE INDEX idx_game_items_level_item ON game_items (game_level_id, content_item_id, deleted_at)`,
		`CREATE INDEX idx_game_items_level_order ON game_items (game_level_id, deleted_at, "order")`,
	}
	for _, sql := range indexes {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260414000002AddGameJunctionPartialIndexes) Down() error {
	indexes := []string{
		`DROP INDEX IF EXISTS idx_game_items_level_order`,
		`DROP INDEX IF EXISTS idx_game_items_level_item`,
		`DROP INDEX IF EXISTS idx_game_metas_level_order`,
		`DROP INDEX IF EXISTS idx_game_metas_level_meta`,
	}
	for _, sql := range indexes {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
```

Changes:
- Line 1 of `Up()` indexes slice: `CREATE UNIQUE INDEX idx_game_metas_level_meta_unique ON game_metas (game_level_id, content_meta_id) WHERE deleted_at IS NULL` → `CREATE INDEX idx_game_metas_level_meta ON game_metas (game_level_id, content_meta_id, deleted_at)`.
- Line 3 of `Up()` indexes slice: analogous for `game_items_level_item`.
- The two `_level_order` lines (lines 2 and 4) are **unchanged** — they're already non-unique and serve a different query pattern.
- `Down()` index names: `idx_game_items_level_item_unique` → `idx_game_items_level_item`, `idx_game_metas_level_meta_unique` → `idx_game_metas_level_meta`.
- Struct declaration, `Signature()`, and imports are untouched.

---

## Task 12: Create new migration 20260415000001_relax_game_junction_indexes

**Rationale:** The live DB already has `20260414000002` applied with the partial unique indexes. This new migration transitions it: creates the new non-unique 3-column indexes first (old uniqueness still enforced, no gap), then drops the old partial unique indexes. Idempotent via `IF NOT EXISTS` / `IF EXISTS` so it's safe to re-run and also becomes a no-op on fresh-migrate paths where Task 11's edit already produced the final shape.

**Files:**
- Create: `dx-api/database/migrations/20260415000001_relax_game_junction_indexes.go`

- [ ] **Step 1: Create the new file**

Create `dx-api/database/migrations/20260415000001_relax_game_junction_indexes.go` with this exact content:

```go
package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260415000001RelaxGameJunctionIndexes struct{}

func (r *M20260415000001RelaxGameJunctionIndexes) Signature() string {
	return "20260415000001_relax_game_junction_indexes"
}

func (r *M20260415000001RelaxGameJunctionIndexes) Up() error {
	stmts := []string{
		// Step 1: Create the new non-unique 3-column indexes BEFORE dropping
		//         the old unique ones. Columns are indexed throughout, and the
		//         old uniqueness is still enforced while the new indexes build.
		`CREATE INDEX IF NOT EXISTS idx_game_metas_level_meta
		   ON game_metas (game_level_id, content_meta_id, deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_game_items_level_item
		   ON game_items (game_level_id, content_item_id, deleted_at)`,
		// Step 2: Drop the old partial unique indexes. Columns remain covered
		//         by the new indexes from step 1.
		`DROP INDEX IF EXISTS idx_game_metas_level_meta_unique`,
		`DROP INDEX IF EXISTS idx_game_items_level_item_unique`,
	}
	for _, sql := range stmts {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260415000001RelaxGameJunctionIndexes) Down() error {
	stmts := []string{
		// Recreate the original partial unique indexes. This will fail if the
		// application has created multi-junction M:N data post-Up — expected,
		// Down is for dev-time rollback before any M:N data exists.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_metas_level_meta_unique
		   ON game_metas (game_level_id, content_meta_id)
		   WHERE deleted_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_game_items_level_item_unique
		   ON game_items (game_level_id, content_item_id)
		   WHERE deleted_at IS NULL`,
		// Drop the new non-unique indexes.
		`DROP INDEX IF EXISTS idx_game_metas_level_meta`,
		`DROP INDEX IF EXISTS idx_game_items_level_item`,
	}
	for _, sql := range stmts {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
```

Note: this matches the pattern of the existing sibling migrations `20260405000003_add_game_session_indexes.go`, `20260405000006_add_game_pk_indexes.go`, and `20260414000002_add_game_junction_partial_indexes.go` — pure raw SQL through `facades.Orm().Query().Exec()`, no `facades.Schema()` calls, with a loop over a slice of SQL strings.

---

## Task 13: Register the new migration in bootstrap

**Rationale:** Goravel's migration runner reads the list from `bootstrap/migrations.go`. The new struct must be appended for the migration to be discovered.

**Files:**
- Modify: `dx-api/bootstrap/migrations.go:60`

- [ ] **Step 1: Add the struct registration**

Find this line in `Migrations()`:

```go
		&migrations.M20260414000002AddGameJunctionPartialIndexes{},
	}
```

Replace with:

```go
		&migrations.M20260414000002AddGameJunctionPartialIndexes{},
		&migrations.M20260415000001RelaxGameJunctionIndexes{},
	}
```

The closing `}` of the slice (which sits on the same line as the new entry's trailing comma in the replaced block, or on the next line depending on current file layout) stays exactly as it was. Do not reorder any other migrations.

---

## Task 14: Verify backend build is clean after schema changes

**Rationale:** Compilation catches any typo in the new struct name, signature mismatch, or missing import. Does not run migrations.

**Files:** none (verification only)

- [ ] **Step 1: Build check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: exits 0 with no output.

Typical failure modes if something is wrong:
- `undefined: migrations.M20260415000001RelaxGameJunctionIndexes` → Task 12 file path or struct name mismatch, or Task 13 registered with a typo.
- `imported and not used: github.com/goravel/framework/facades` in the new file → struct declared but `Up()`/`Down()` bodies missing. Re-check Task 12.

- [ ] **Step 2: Vet check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./...
```

Expected: exits 0 with no output.

---

## Task 15: Commit the schema relaxation

**Rationale:** Phase 2's three files are all in a consistent end state, backend compiles. This is a clean atomic commit for the schema-only change.

**Files:** the 3 files from Tasks 11-13.

- [ ] **Step 1: Stage and commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && \
git add dx-api/database/migrations/20260414000002_add_game_junction_partial_indexes.go \
        dx-api/database/migrations/20260415000001_relax_game_junction_indexes.go \
        dx-api/bootstrap/migrations.go && \
git commit -m "$(cat <<'EOF'
refactor(api): relax game junction unique indexes to non-unique

Replace the partial unique indexes idx_game_metas_level_meta_unique
and idx_game_items_level_item_unique with non-unique, non-partial
indexes keyed on (game_level_id, content_*_id, deleted_at).

Unlocks M:N between game_levels and content_metas/content_items —
the same content can now appear multiple times within a level and
across multiple levels without duplicating underlying content rows.

New migration 20260415000001_relax_game_junction_indexes transitions
the live DB by creating the new indexes first, then dropping the old
partial unique ones. Idempotent via IF NOT EXISTS / IF EXISTS.

The existing 20260414000002 is also edited so fresh-migrate paths
reach the same end state directly (Goravel doesn't re-run already-
applied migrations, so the edit only affects fresh DBs).

The two *_level_order composite indexes are untouched — they already
serve their own query pattern.
EOF
)"
```

Expected: single commit created.

- [ ] **Step 2: Verify the commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git log -1 --stat
```

Expected output: commit subject `refactor(api): relax game junction unique indexes to non-unique`, and a file list showing exactly 3 files — 2 `dx-api/database/migrations/*.go` and `dx-api/bootstrap/migrations.go`, with one being `create mode` (the new migration) and the other two `modify`.

---

## Task 16: Pre-migration backup of the live DB

**Rationale:** Changing indexes on a table with ~1.22M rows is low-risk (it doesn't touch any row), but the user explicitly keeps DB snapshots in `/Users/rainsen/Programs/Projects/douxue/db-backup/` and this is standard practice. Skipping the snapshot is not worth the minute it takes.

**Files:** none (creates a backup file outside the repo)

- [ ] **Step 1: Confirm no server is running against the DB**

Before running the migration, stop any active `go run .` / `air` process that has the `douxue` database open. A live writer is not catastrophic (the `CREATE INDEX` takes a brief SHARE lock that blocks writes, and everything resumes after), but it's cleaner to shut the server down first.

```bash
# If you have a running process, find and stop it. For example:
pgrep -f 'go run \.' || echo 'no go run process'
pgrep -f 'air' || echo 'no air process'
```

If either prints a PID, stop that process with the appropriate method (Ctrl+C in the terminal that started it).

- [ ] **Step 2: Take the snapshot**

```bash
mkdir -p /Users/rainsen/Programs/Projects/douxue/db-backup && \
pg_dump -h localhost -U postgres douxue \
  > /Users/rainsen/Programs/Projects/douxue/db-backup/dx-pre-relax-indexes-$(date +%Y%m%d-%H%M%S).sql
```

Expected: the command exits 0 silently.

- [ ] **Step 3: Verify the dump is non-empty**

```bash
ls -lah /Users/rainsen/Programs/Projects/douxue/db-backup/dx-pre-relax-indexes-*.sql | tail -1
```

Expected: the most recent file is at least several hundred KB (with 1.22M `game_items` rows, the full dump will be hundreds of MB or more). If it's under ~10 KB, the dump failed silently — do NOT proceed; re-check `pg_dump` credentials and the DB name.

---

## Task 17: Run the new migration against the live DB

**Rationale:** This is the moment the live DB's indexes swap. All prior tasks set the code up; this applies it.

**Files:** none (runs artisan command)

- [ ] **Step 1: Run migrate**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go run . artisan migrate
```

Expected output: a line like:
```
INFO  Running: 20260415000001_relax_game_junction_indexes
INFO  Migrated: 20260415000001_relax_game_junction_indexes
```
or similar (exact format depends on Goravel version). No error lines. The older migrations are already tracked as applied and will not re-run.

If the command prints `relation "..._unique" does not exist` — that means the live DB for some reason already has the new state. Not necessarily a bug, but stop and investigate before proceeding. You can confirm the current state with `psql -h localhost -U postgres -d douxue -c '\d game_items'` and compare to the expected output in Task 18.

If the command prints `column "..." does not exist` — that would indicate a typo in the SQL. Re-check Task 12 carefully.

---

## Task 18: Verify the live DB schema matches the target

**Rationale:** Reading the actual Postgres state confirms the migration did what we think it did.

**Files:** none (read-only psql queries)

- [ ] **Step 1: Inspect `game_metas` indexes**

```bash
psql -h localhost -U postgres -d douxue -c '\d game_metas'
```

Expected indexes list (the `\d` output has a "Indexes:" section listing them):

- `game_metas_pkey` — PRIMARY KEY on `id`
- `game_metas_content_meta_id_index` (or similar) — single-column blueprint index from the original create migration
- `game_metas_created_at_index` — single-column blueprint index
- `game_metas_game_id_index` — single-column blueprint index
- `idx_game_metas_level_meta` — `btree (game_level_id, content_meta_id, deleted_at)` — the NEW index, non-unique
- `idx_game_metas_level_order` — `btree (game_level_id, deleted_at, "order")` — unchanged

**Must NOT be present:**
- `idx_game_metas_level_meta_unique` — the old partial UNIQUE index

- [ ] **Step 2: Inspect `game_items` indexes**

```bash
psql -h localhost -U postgres -d douxue -c '\d game_items'
```

Expected indexes list, analogous to Step 1:

- `game_items_pkey`, `game_items_content_item_id_index`, `game_items_created_at_index`, `game_items_game_id_index`
- `idx_game_items_level_item` — `btree (game_level_id, content_item_id, deleted_at)` — NEW, non-unique
- `idx_game_items_level_order` — `btree (game_level_id, deleted_at, "order")` — unchanged

**Must NOT be present:** `idx_game_items_level_item_unique`.

- [ ] **Step 3: Row-count invariant check**

```bash
psql -h localhost -U postgres -d douxue -c 'SELECT COUNT(*) AS live_game_items FROM game_items WHERE deleted_at IS NULL;'
psql -h localhost -U postgres -d douxue -c 'SELECT COUNT(*) AS live_game_metas FROM game_metas WHERE deleted_at IS NULL;'
```

Expected: `live_game_items` is around 1.22M (matches the pre-migration count exactly — if you took note of it earlier, compare; otherwise it should be in the low millions). `live_game_metas` is 0 (the junction was populated on game_items only, not game_metas, per the prior junction refactor).

If either count is 0 when it shouldn't be, or dramatically different from the expected range, STOP and investigate. This should not happen — `CREATE INDEX` and `DROP INDEX` do not touch row data — but a catastrophic failure somewhere else in the migration pipeline could look like this, and catching it here is much safer than catching it after a smoke test deletes something.

---

## Task 19: Manual smoke tests — reorder and single-play

**Rationale:** Exercises the reorder fix end-to-end and confirms the index swap didn't subtly break any query plan. These are browser-manual tests; an agent cannot run them.

**Files:** none

- [ ] **Step 1: Start the backend and frontend**

In separate terminals:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go run .
```

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run dev
```

Expected: backend starts on port 3001 with no error on the startup logs (pay attention to the migration lines — all previous migrations should be "already applied", no new ones should run since Task 17 already applied `20260415000001`). Frontend starts on port 3000.

- [ ] **Step 2: Log in as a VIP user and navigate to the course editor**

Open `http://localhost:3000` in a browser, log in with a VIP account (reorder requires VIP per `requireVip` in both reorder service functions), and navigate to AI-custom → pick a game → pick a level that has at least 2 metas and 2 content items.

- [ ] **Step 3: Reorder a meta within the level**

Drag one meta from position 1 to position 2 (or vice versa). The UI should update immediately. Refresh the page — the new order should persist.

Expected: the order change is visible after refresh. The backend server log (the `go run .` terminal) should show a successful PUT `/api/course-games/.../metadata/reorder` response with no error.

If reorder fails with a 4xx error, check the browser devtools Network tab for the request body. It should contain `{"gameLevelId": "...", "metaId": "...", "newOrder": ...}`. If `gameLevelId` is missing, Task 7 or Task 8 wasn't saved.

- [ ] **Step 4: Reorder a content item within the level**

Same dance: drag one item to a new position within its meta. Refresh and confirm order persists.

Expected: same as Step 3, but for the item reorder path. The PUT body should be `{"gameLevelId": "...", "itemId": "...", "newOrder": ...}`.

- [ ] **Step 5: Single-play the reordered level**

From the hall, start a single-play session on the level you just reordered. Play through at least 5 items.

Expected: items appear in the new order (i.e., the order change from Steps 3-4 is reflected in the play sequence). No 5xx errors on the server log. Session completes normally.

- [ ] **Step 6: Spot-check: reorder a different level's content to confirm isolation**

Navigate to a DIFFERENT level (one that contains different content) and reorder its metas/items. Confirm its order changes persist. This is specifically the case that the old buggy reorder would have silently corrupted if any content were shared — under today's 1:1 data the behavior is identical, but exercising it confirms the fix didn't break the happy path.

Expected: second level's reorder works. First level (from Steps 3-4) still shows the orders you set earlier.

---

## Self-Review Checklist (for the implementer, after all tasks)

- [ ] Phase 1 commit (`fix(api): scope reorder paths...`) exists and touches exactly 6 files.
- [ ] Phase 2 commit (`refactor(api): relax game junction unique indexes...`) exists and touches exactly 3 files, one of which is a `create mode` for `20260415000001_relax_game_junction_indexes.go`.
- [ ] `ReorderMetadata` and `ReorderContentItems` in `course_content_service.go` both take `gameLevelID` as the third parameter and scope the UPDATE by it.
- [ ] `ReorderContentItemRequest` has `GameLevelID` as its first field with `required|uuid` validation.
- [ ] `reorderItemAction` in `course-game.action.ts` takes `gameLevelId` as the second argument and sends it in the payload.
- [ ] `level-units-panel.tsx` passes `levelId` to `reorderItemAction` at its only call site.
- [ ] `go build ./...` and `go vet ./...` in `dx-api` are both clean.
- [ ] `npm run lint` in `dx-web` is clean.
- [ ] `psql ... -c '\d game_metas'` shows `idx_game_metas_level_meta` with key `(game_level_id, content_meta_id, deleted_at)` and does NOT show `idx_game_metas_level_meta_unique`.
- [ ] `psql ... -c '\d game_items'` shows `idx_game_items_level_item` similarly, does NOT show `idx_game_items_level_item_unique`.
- [ ] `SELECT COUNT(*) FROM game_items WHERE deleted_at IS NULL` matches the pre-migration count (no row loss).
- [ ] Pre-migration `pg_dump` file exists under `/Users/rainsen/Programs/Projects/douxue/db-backup/` and is non-trivial in size.
- [ ] Drag-reorder of meta works and persists after refresh.
- [ ] Drag-reorder of content item works and persists after refresh.
- [ ] Single-play on the reordered level respects the new order.

If any item is unchecked, fix before declaring the task done.

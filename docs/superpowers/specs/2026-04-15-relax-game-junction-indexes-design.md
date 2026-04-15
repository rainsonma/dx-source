---
title: Relax Game Junction Unique Indexes for M:N Capability
date: 2026-04-15
status: approved
related:
  - dx-api/database/migrations/20260414000002_add_game_junction_partial_indexes.go
  - dx-api/app/services/api/course_content_service.go
  - dx-api/app/http/controllers/api/course_game_controller.go
  - dx-api/app/http/requests/api/course_game_request.go
  - dx-web/src/lib/api-client.ts
  - dx-web/src/features/web/ai-custom/actions/course-game.action.ts
  - dx-web/src/features/web/ai-custom/components/level-units-panel.tsx
---

# Relax Game Junction Unique Indexes for M:N Capability

## Purpose

Enable a many-to-many relationship between `game_levels` and `content_metas` / `content_items` by relaxing the two partial unique indexes on the junction tables (`game_metas`, `game_items`) from `UNIQUE` to non-unique, and simultaneously fix the one latent bug that dropping the uniqueness guarantee exposes: both `ReorderMetadata` and `ReorderContentItems` currently issue UPDATE statements without a `game_level_id` scope, which works only because the unique constraint makes at most one junction row per `(level, content_id)` exist.

After this task, the database will permit duplicate junction rows (same level, same content, appearing more than once) and reordering will correctly scope to one level. The task does **not** implement dedup-on-save, reference-counted deletes, or the frontend delete-route rewrite — those remain the separate `2026-04-14-content-meta-dedup-design.md` task.

## Background

The junction tables `game_metas` and `game_items` were added on 2026-04-14 (commit `4f9f60e`) and their partial indexes were moved to a sibling migration on 2026-04-15 (commit `6e7e36d`) after discovering that mixing `facades.Schema().Create()` and `facades.Orm().Query().Exec()` in one Goravel migration fails on a fresh DB. The current partial indexes (created by `20260414000002_add_game_junction_partial_indexes.go`) are:

```sql
CREATE UNIQUE INDEX idx_game_metas_level_meta_unique
  ON game_metas (game_level_id, content_meta_id)
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_game_items_level_item_unique
  ON game_items (game_level_id, content_item_id)
  WHERE deleted_at IS NULL;
```

The uniqueness was a holdover from the pre-junction era when `content_metas.game_level_id` was a direct 1:N FK. Now that the junction is in place, lifting uniqueness unlocks two related capabilities:

1. **Cross-level reuse** — the same `content_meta` / `content_item` row can be linked from multiple levels without duplicating underlying content. (The dedup-on-save task will actively create reuse; this task just removes the structural obstacle.)
2. **In-level repetition** — the same meta/item can appear more than once inside a single level (e.g., drilling the same word three times in a sequence). Currently forbidden by the unique constraint.

A comprehensive audit of the backend (summarized in the companion research report) found that the only code path that silently depends on the uniqueness guarantee is the reorder SQL in `course_content_service.go`:

```go
// ReorderMetadata — line 187
UPDATE game_metas SET "order" = ? WHERE content_meta_id = ? AND deleted_at IS NULL

// ReorderContentItems — line 456
UPDATE game_items SET "order" = ? WHERE content_item_id = ? AND deleted_at IS NULL
```

Both UPDATE statements lack a `game_level_id` clause. Under the current unique constraint, at most one junction row per content ID can exist across all levels (a single meta linked to a single level), so these updates match exactly one row in practice. The instant an M:N relationship creates a second junction row with the same content ID (in any level), these UPDATEs will silently rewrite the `"order"` value on **every** level that references the same content — corrupting order across unrelated levels. This is the one "reference caused by the index change" that must be fixed in the same task as the relaxation.

Every other junction-table call site (save paths, AI-custom break / generate paths, game-play read paths, course import CLI, `DeleteGame` / `DeleteLevel`) is already structured in a way that tolerates M:N. The delete paths in `course_content_service.go` (`DeleteContentItem`, `DeleteMetadata`, `DeleteAllLevelContent`) are correct under today's 1:1 data but will over-delete once M:N data exists; rewriting them with reference counting is the explicit scope of the dedup task, not this one.

## Scope

### In scope

- Schema migration that replaces the two partial unique indexes with non-unique, non-partial indexes including `deleted_at` as a third key column.
- Edit to the existing `20260414000002_add_game_junction_partial_indexes.go` so fresh-migrate paths produce the final shape directly.
- `ReorderMetadata`: add `gameLevelID` parameter, scope the UPDATE to `game_level_id = ?`, plumb the value from the controller (the `ReorderMetadataRequest` already carries `GameLevelID`).
- `ReorderContentItems`: same treatment, plus add `GameLevelID` to `ReorderContentItemRequest` (currently missing), plus the small frontend change to send the field in the payload.
- Verification on live data: pre-migration `pg_dump` snapshot, schema inspection, row-count invariants, reorder smoke test, single-play smoke test.

### Out of scope

- Dedup-on-save logic in `SaveMetadataBatch` — owned by `2026-04-14-content-meta-dedup-design.md`.
- Reference-counted delete rewrite for `DeleteContentItem` / `DeleteMetadata` / `DeleteAllLevelContent` — same.
- Frontend delete-route rewrite (`/levels/{levelId}/` in delete URLs) — same.
- Adding a dedup lookup index on `content_metas (source_type, source_data)` — same.
- Backfill / cleanup of any pre-existing data. None is needed: the existing rows are compliant with the old unique constraint by construction, so the new non-unique index builds cleanly.
- Changes to `idx_game_metas_level_order` / `idx_game_items_level_order`. These composite indexes on `(game_level_id, deleted_at, "order")` are already non-unique and serve a different query pattern (index range scan with pre-sorted `"order"` output); leaving them untouched.

## Decisions (from brainstorming)

| Question | Decision |
|---|---|
| Unique → non-unique or keep unique with different shape? | Non-unique — M:N is the explicit goal. |
| Partial (`WHERE deleted_at IS NULL`) or full index? | Full. `deleted_at` moves from WHERE clause into the key columns (see next row). |
| Column order for the new index | `(game_level_id, content_meta_id, deleted_at)` for `game_metas`, analogous for `game_items`. Rationale: leftmost-prefix `(level, content_id)` handles the link-existence and "how many live repetitions" queries directly; trailing `deleted_at` gives Postgres a clean extension when the query also filters by it. Roughly 10 MB larger than a 2-col index on `game_items` — trivial. |
| Index naming | Drop the `_unique` suffix since it's no longer accurate: `idx_game_metas_level_meta`, `idx_game_items_level_item`. |
| Migration ordering | Create new indexes first, then drop old ones (both wrapped in `IF NOT EXISTS` / `IF EXISTS` so the migration is idempotent and safe on fresh + existing DB paths). |
| Scope of code changes in this task | Only the reorder path (the one latent bug uniqueness was hiding). Delete/save changes belong to the dedup task. |
| Commit granularity | Two commits — reorder fix first, schema relaxation second. Reorder fix is a precondition for the schema relaxation being safe. |
| Backup strategy | Take a `pg_dump` snapshot before running the new migration on the current live DB. Stored in `/Users/rainsen/Programs/Projects/douxue/db-backup/`. |

## Schema changes

### New migration: `dx-api/database/migrations/20260415000001_relax_game_junction_indexes.go`

Pattern: matches existing raw-SQL index migrations (`20260405000003_add_game_session_indexes.go`, `20260405000006_add_game_pk_indexes.go`, `20260414000002_add_game_junction_partial_indexes.go`). Pure raw SQL through `facades.Orm().Query().Exec()`; no `facades.Schema()` calls. Idempotent.

**`Up()`:**

```go
func (r *M20260415000001RelaxGameJunctionIndexes) Up() error {
    stmts := []string{
        // Step 1: Create new non-unique 3-column indexes BEFORE dropping the old
        //         unique ones. The columns are indexed throughout, and the old
        //         uniqueness constraint is still enforced while the new index is
        //         being built — no window of unenforced uniqueness.
        `CREATE INDEX IF NOT EXISTS idx_game_metas_level_meta
           ON game_metas (game_level_id, content_meta_id, deleted_at)`,
        `CREATE INDEX IF NOT EXISTS idx_game_items_level_item
           ON game_items (game_level_id, content_item_id, deleted_at)`,
        // Step 2: Drop the old partial unique indexes. Columns are still covered
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
```

**`Down()`:**

```go
func (r *M20260415000001RelaxGameJunctionIndexes) Down() error {
    stmts := []string{
        // Recreate the original partial unique indexes.
        `CREATE UNIQUE INDEX IF NOT EXISTS idx_game_metas_level_meta_unique
           ON game_metas (game_level_id, content_meta_id)
           WHERE deleted_at IS NULL`,
        `CREATE UNIQUE INDEX IF NOT EXISTS idx_game_items_level_item_unique
           ON game_items (game_level_id, content_item_id)
           WHERE deleted_at IS NULL`,
        // Drop the non-unique indexes.
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

Down() will **fail** if the application has created M:N data (multiple rows with identical `(level, content_id)` and `deleted_at IS NULL`) by the time of rollback. This is expected and acceptable: Down() is a development-time safety net for "revert the migration before anyone creates M:N data," not a production escape hatch. Once M:N data exists, the recovery path is `pg_restore` from a pre-migration snapshot.

### Edit: `20260414000002_add_game_junction_partial_indexes.go`

Replace the `indexes` slice in `Up()` so fresh-migrate paths produce the final shape directly, without ever creating the unique version. The two `_level_order` composite indexes are untouched.

Before (current):

```go
indexes := []string{
    `CREATE UNIQUE INDEX idx_game_metas_level_meta_unique ON game_metas (game_level_id, content_meta_id) WHERE deleted_at IS NULL`,
    `CREATE INDEX idx_game_metas_level_order ON game_metas (game_level_id, deleted_at, "order")`,
    `CREATE UNIQUE INDEX idx_game_items_level_item_unique ON game_items (game_level_id, content_item_id) WHERE deleted_at IS NULL`,
    `CREATE INDEX idx_game_items_level_order ON game_items (game_level_id, deleted_at, "order")`,
}
```

After:

```go
indexes := []string{
    `CREATE INDEX idx_game_metas_level_meta ON game_metas (game_level_id, content_meta_id, deleted_at)`,
    `CREATE INDEX idx_game_metas_level_order ON game_metas (game_level_id, deleted_at, "order")`,
    `CREATE INDEX idx_game_items_level_item ON game_items (game_level_id, content_item_id, deleted_at)`,
    `CREATE INDEX idx_game_items_level_order ON game_items (game_level_id, deleted_at, "order")`,
}
```

`Down()` is also updated so the dropped names match the new `Up()`:

```go
indexes := []string{
    `DROP INDEX IF EXISTS idx_game_items_level_order`,
    `DROP INDEX IF EXISTS idx_game_items_level_item`,
    `DROP INDEX IF EXISTS idx_game_metas_level_order`,
    `DROP INDEX IF EXISTS idx_game_metas_level_meta`,
}
```

### Edit: `dx-api/bootstrap/migrations.go`

Register the new migration immediately after the existing `M20260414000002AddGameJunctionPartialIndexes{}` entry:

```go
&migrations.M20260414000001CreateGameMetasAndGameItemsTables{},
&migrations.M20260414000002AddGameJunctionPartialIndexes{},
&migrations.M20260415000001RelaxGameJunctionIndexes{},   // NEW
```

### Convergence on both migrate paths

**Fresh DB path** (user's previous reset-plus-backup-restore path, or any other fresh migrate):
- `20260414000002` runs with the edited indexes → creates `idx_game_metas_level_meta` and `idx_game_items_level_item` directly.
- `20260415000001` runs → `CREATE INDEX IF NOT EXISTS` skips both (already exist); `DROP INDEX IF EXISTS` skips both (they never existed as the `_unique` names). Migration is a no-op and commits successfully.
- End state: only the non-unique indexes, with three-column key.

**Existing DB path** (current live DB, `20260414000002` already applied):
- `20260414000002` is tracked as applied and not re-run. The edit to its file does not affect anything.
- `20260415000001` runs → `CREATE INDEX IF NOT EXISTS` creates both new indexes (they don't exist yet on the live DB); `DROP INDEX IF EXISTS` drops both old `_unique` indexes (they still exist on the live DB). Commits.
- End state: same as fresh path — only the non-unique indexes.

Both converge. The migration is idempotent and re-runnable on either path.

### Data safety on the existing DB

Changing an index is **not a data migration**. Postgres B-tree indexes reference underlying rows via CTIDs but never rewrite row bytes. Every `game_items` row (~1.22M) and every `game_metas` row (~0) is preserved exactly as-is through the migration. The migration only builds new B-tree structures and drops the old ones.

Creating a **non-unique** index cannot fail on existing data, because there is no constraint to violate. Postgres reads every row and inserts its key columns into the new B-tree unconditionally. Even if the existing data had (hypothetically) accumulated duplicate `(level, content_id)` pairs somehow, the new index would absorb them without error.

Locking (for awareness — development DB doesn't require `CONCURRENTLY`):
- `CREATE INDEX` (without `CONCURRENTLY`) takes a `SHARE` lock on the target table — blocks writes, allows reads. On `game_items` (~1.22M rows), expect a few seconds. On `game_metas` (~0 rows), milliseconds.
- `DROP INDEX` takes an `ACCESS EXCLUSIVE` lock — blocks both reads and writes. Duration: milliseconds.

For a production-grade rollout with live traffic, these would become `CREATE INDEX CONCURRENTLY` / `DROP INDEX CONCURRENTLY` — but that requires the migration runner to NOT wrap statements in a transaction (CONCURRENTLY cannot run inside a txn block). Goravel's migration behavior here should be confirmed before any production run. For the user's current dev environment, the default locking is fine.

## Code changes

### `dx-api/app/services/api/course_content_service.go`

**`ReorderMetadata` (currently lines 169-195)** — add `gameLevelID` parameter, scope the UPDATE:

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

**`ReorderContentItems` (currently lines 439-464)** — same pattern:

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

Both changes are purely additive to the WHERE clause on top of a new required parameter. On current 1:1 data, behavior is identical — the added `game_level_id = ?` matches exactly the same one row that the old `content_*_id = ?` alone matched. No user-visible behavior change on today's data; M:N-safety starts the moment a duplicate junction row exists.

### `dx-api/app/http/controllers/api/course_game_controller.go`

**`ReorderMetadata` controller (around line 294)** — thread `req.GameLevelID` (field already exists on the request):

```go
if err := services.ReorderMetadata(userID, gameID, req.GameLevelID, req.MetaID, req.NewOrder); err != nil {
    return mapCourseGameError(ctx, err)
}
```

**`ReorderContentItems` controller (around line 410)** — thread `req.GameLevelID` (field added to the request below):

```go
if err := services.ReorderContentItems(userID, gameID, req.GameLevelID, req.ItemID, req.NewOrder); err != nil {
    return mapCourseGameError(ctx, err)
}
```

### `dx-api/app/http/requests/api/course_game_request.go`

**`ReorderContentItemRequest` (currently lines 260-279)** — add `GameLevelID`:

```go
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

`ReorderMetadataRequest` already has `GameLevelID` with matching validation rules — no changes there.

### `dx-web/src/lib/api-client.ts`

**`reorderContentItems` (currently line 655)** — extend the `data` shape to include `gameLevelId`:

```typescript
/** Reorder content items */
async reorderContentItems(
  gameId: string,
  data: { gameLevelId: string; itemId: string; newOrder: number }
) {
  return apiClient.put<null>(`/api/course-games/${gameId}/content-items/reorder`, data);
},
```

### `dx-web/src/features/web/ai-custom/actions/course-game.action.ts`

**`reorderItemAction` (currently lines 250-264)** — add `gameLevelId` parameter, pass it in the payload. Place it right after `gameId` to match the existing `reorderMetaAction` signature shape.

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

### `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx`

**Call site (currently line 344-347)** — `levelId` is already in scope as a component prop (used at line 205 for `reorderMetaAction`). Pass it as the second argument:

```typescript
const result = await reorderItemAction(
  gameId,
  levelId,
  active.id as string,
  newOrder
);
```

### Files touched

| File | Change |
|---|---|
| `dx-api/database/migrations/20260414000002_add_game_junction_partial_indexes.go` | edit (rename + non-unique + 3-col) |
| `dx-api/database/migrations/20260415000001_relax_game_junction_indexes.go` | new |
| `dx-api/bootstrap/migrations.go` | edit (register new migration) |
| `dx-api/app/services/api/course_content_service.go` | edit (2 functions) |
| `dx-api/app/http/controllers/api/course_game_controller.go` | edit (2 call sites) |
| `dx-api/app/http/requests/api/course_game_request.go` | edit (add field to 1 struct) |
| `dx-web/src/lib/api-client.ts` | edit (1 function) |
| `dx-web/src/features/web/ai-custom/actions/course-game.action.ts` | edit (1 action) |
| `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx` | edit (1 call site) |

Total: 9 files. 3 backend source files + 3 migration files + 3 frontend files. Split across two commits (reorder fix: 6 files; index change: 3 files).

### Commit plan

**Commit 1 — `fix(api): scope reorder paths to game_level_id for M:N safety`**

- `dx-api/app/services/api/course_content_service.go`
- `dx-api/app/http/controllers/api/course_game_controller.go`
- `dx-api/app/http/requests/api/course_game_request.go`
- `dx-web/src/lib/api-client.ts`
- `dx-web/src/features/web/ai-custom/actions/course-game.action.ts`
- `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx`

Commit message body: "Both ReorderMetadata and ReorderContentItems issued UPDATEs scoped only by content_*_id, which worked by accident under the partial unique index that forbade more than one junction row per (level, content). Scoping the UPDATE to game_level_id too makes reorder correct under the upcoming M:N relaxation. Also threads gameLevelId through the content-item reorder request (which previously lacked the field) and its frontend callers."

**Commit 2 — `refactor(api): relax game junction unique indexes to non-unique`**

- `dx-api/database/migrations/20260414000002_add_game_junction_partial_indexes.go`
- `dx-api/database/migrations/20260415000001_relax_game_junction_indexes.go`
- `dx-api/bootstrap/migrations.go`

Commit message body: "Replace the partial unique indexes idx_game_metas_level_meta_unique and idx_game_items_level_item_unique with non-unique, non-partial indexes keyed on (game_level_id, content_*_id, deleted_at). Unlocks M:N between game_levels and content_metas/content_items — the same content can now appear multiple times within a level and across multiple levels. New migration 20260415000001 transitions the live DB; the existing 20260414000002 is edited so fresh-migrate paths reach the same end state directly. Idempotent on both paths via IF NOT EXISTS / IF EXISTS."

## Verification

### Automated

1. **Backend build and vet (after commit 1 and commit 2):**
   ```bash
   cd dx-api && go build ./... && go vet ./...
   ```
   Both must exit 0. Compile errors catch any missed caller of `ReorderMetadata` / `ReorderContentItems` with the old 4-arg signature.

2. **Backend tests (after commit 1):**
   ```bash
   cd dx-api && go test -race ./...
   ```
   No existing tests assert uniqueness (verified in the exploration phase); the test suite should continue to pass.

3. **Frontend lint (after commit 1):**
   ```bash
   cd dx-web && npm run lint
   ```
   TypeScript will flag any missed caller of `reorderItemAction` with the old 3-arg signature.

### Manual — database migration on the live DB

4. **Pre-migration snapshot:**
   ```bash
   pg_dump -h localhost -U postgres douxue \
     > /Users/rainsen/Programs/Projects/douxue/db-backup/dx-pre-relax-indexes-$(date +%Y%m%d-%H%M%S).sql
   ```
   Verify the dump file is non-empty (at least a few hundred KB) before proceeding.

5. **Run the new migration:**
   ```bash
   cd dx-api && go run . artisan migrate
   ```
   Expected output: one new "Migrated: 20260415000001_relax_game_junction_indexes" line. No errors. The older migrations are already tracked and are not re-run.

6. **Schema inspection:**
   ```sql
   \d game_metas
   \d game_items
   ```
   Expected `game_metas` indexes: `game_metas_pkey` (PK), `idx_game_metas_level_meta` on `(game_level_id, content_meta_id, deleted_at)`, `idx_game_metas_level_order` on `(game_level_id, deleted_at, "order")`, plus the per-column indexes created via blueprint (`game_id`, `content_meta_id`, `created_at`). The `_unique` name must be gone.
   Expected `game_items` indexes: analogous — `game_items_pkey`, `idx_game_items_level_item` on `(game_level_id, content_item_id, deleted_at)`, `idx_game_items_level_order` on `(game_level_id, deleted_at, "order")`, plus `game_id`, `content_item_id`, `created_at`. No `_unique`.

7. **Row-count invariant:**
   ```sql
   SELECT COUNT(*) FROM game_items WHERE deleted_at IS NULL;
   SELECT COUNT(*) FROM game_metas WHERE deleted_at IS NULL;
   ```
   Both must match their pre-migration values exactly (no row was touched by an index swap).

### Manual — application smoke tests

8. **Reorder meta within a level:** Open the course editor, drag-reorder a meta within a level. Expected: the new order persists (refresh confirms). Open a different level containing different content; expected: that level's order is unchanged (this specifically guards against the pre-fix cross-level corruption — it wouldn't fire under today's 1:1 data, but verifies the WHERE-clause fix didn't break the happy path).

9. **Reorder content item within a level:** Same — drag-reorder an item, refresh, confirm.

10. **Single-play on reordered level:** Start a single-play session on the reordered level. Confirm items appear in the new order. Play through at least 5 items without errors.

11. **Save new AI-custom meta:** Confirm that adding new metadata via the AI-custom save flow still works unchanged (this confirms the save paths are unaffected by the index change).

## Risks & mitigations

| Risk | Mitigation |
|---|---|
| Pre-migration backup missing or corrupt | Verification step 4 explicitly takes and size-checks the dump before running the migration |
| Migration crashes halfway (e.g., first `CREATE INDEX` succeeds, second fails) | `IF NOT EXISTS` / `IF EXISTS` make the migration re-runnable; next `artisan migrate` attempt picks up where it left off without error |
| Some caller of `ReorderMetadata` or `ReorderContentItems` missed in the edit, still passes the old 4-arg signature | Go compiler fails the build; `go build ./...` is a mandatory step before the commit |
| Frontend caller of `reorderItemAction` missed in the edit, still passes 3 args | TypeScript compiler fails; `npm run lint` (or `tsc --noEmit`) is mandatory before the commit |
| `Down()` can't rebuild the unique index if M:N data has been created post-migration | Documented; `Down()` is a development-time rollback only. For production, recovery is `pg_restore` from the pre-migration snapshot |
| Migration takes longer than expected due to index rebuild on `game_items` (~1.22M rows) | Expected duration is a few seconds at most; the `SHARE` lock blocks writes only during the scan, reads remain available. No user impact in dev environment |
| Concurrent writes to `game_items` during the migration | Blocked by the `SHARE` lock; resume automatically after `CREATE INDEX` finishes. No data loss |
| Cross-task contention with the upcoming dedup task | The dedup task already anticipates this change (its spec explicitly plans to drop the unique partial indexes). After this task lands, the dedup task's migration becomes a pure add-the-dedup-lookup-index operation, simpler than originally specified. No conflict |
| `deleted_at` in the key column vs. the WHERE clause changes query planner behavior unexpectedly | The 3-col key index is strictly more flexible than a 2-col index; any query that used the old partial index is served equally well or better by the new non-partial index. Verification step 8-10 exercises the hot path empirically |

## Out of scope / future work

- **Dedup-on-save (`SaveMetadataBatch`)** — rewrites the save loop to reuse existing `content_metas` / `content_items` owned by the same user. Spec at `docs/superpowers/specs/2026-04-14-content-meta-dedup-design.md`.
- **Reference-counted deletes** — rewrites `DeleteContentItem` / `DeleteMetadata` / `DeleteAllLevelContent` to only soft-delete underlying content rows when no live junction row remains anywhere. Spec: same as above.
- **Frontend delete-route rewrite** — changes two delete URLs to include `/levels/{levelId}/` so the backend can scope deletes to a single level. Spec: same as above.
- **`idx_content_metas_dedup_lookup`** — index on `content_metas (source_type, source_data) WHERE deleted_at IS NULL` to accelerate the dedup query. Spec: same as above.
- **`CREATE INDEX CONCURRENTLY` for production rollouts** — not needed for the current dev environment. Add as a deploy-time switch if this work ever moves to a live production DB with concurrent writers.
- **Test coverage for the new junction semantics** — the current repo has no `game_metas` / `game_items` unit tests. Adding them is desirable but out of scope for this task; the dedup task's spec already includes a test file (`dx-api/tests/feature/course_content_dedup_test.go`) that will exercise the new semantics end-to-end.

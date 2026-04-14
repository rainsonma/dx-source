# Reintroduce Junction Tables for Content Reuse — Design Spec

**Date:** 2026-04-14
**Scope:** `dx-api` backend only (no frontend changes, no API shape changes)
**Status:** Approved

## Purpose

Add `game_metas` and `game_items` as many-to-many junction tables between `game_levels` and `content_metas`/`content_items`, so that content units can be reused across multiple levels and games in a future task without duplicating the underlying `content_items` rows (currently 1,220,803 rows).

After this task, junction tables are the authoritative source for which content belongs to which level. The content tables (`content_metas`, `content_items`) become pure content stores — no per-level attributes.

**This task introduces the structure and migrates the existing 1:1 data into the junction tables.** The actual reuse (cross-level linking of existing content) is a follow-up task.

## Background

This is the **second attempt** at junction tables:

1. **Apr 7, 2026** (commit `07c244d`): `game_metas`/`game_items` added, with backfill + code rewrites.
2. **Apr 9, 2026** (commits `9d716eb`, `f62beb7`, `f3da121`): junction tables **removed**, `game_level_id` put back directly on content tables. Reason documented in `docs/superpowers/specs/2026-04-09-remove-junction-tables-design.md` — at the time, simpler queries and no orphan cleanup outweighed the reuse benefit.
3. **Apr 14, 2026** (this spec): reintroduce junction tables, now that content reuse is the explicit goal.

The empty migration stub `dx-api/database/migrations/20260407000001_create_game_junction_tables.go` is left in place (it was already registered in some environments) and is unrelated to the new migrations.

## Current State

- `content_metas`: 0 rows. Columns include `id`, `game_level_id`, `source_from`, `source_type`, `source_data`, `translation`, `is_break_done`, `order`, timestamps, `deleted_at`.
- `content_items`: **1,220,803 rows**. Columns include `id`, `game_level_id`, `content_meta_id` (nullable), `content`, `content_type`, `uk_audio_id`, `us_audio_id`, `definition`, `translation`, `explanation`, `items` (jsonb), `structure` (jsonb), `order`, `tags`, `is_active`, timestamps, `deleted_at`.
- Neither table has a DB-level FK (code-level constraints only, for Postgres partition compatibility).
- `content_items.is_active` is used as a **read filter** in 4 call sites but **never set to `false`** by any code path — effectively dead.
- Six content-authoring models (`Game`, `GameLevel`, `ContentMeta`, `ContentItem`, plus future `GameMeta`, `GameItem`) use `orm.SoftDeletes`. Joined tables need explicit `AND deleted_at IS NULL` in raw JOINs since the ORM soft-delete auto-filter only covers the primary model.

## Goals

- Introduce `game_metas` and `game_items` junction tables with proper indexes for efficient reads.
- Backfill existing content (1,220,803 content_items + 0 content_metas) into junction tables as 1:1 links so all current behavior is preserved.
- Drop `game_level_id`, `order`, and `is_active` from content tables — junction tables become the authoritative source for these per-level attributes.
- Update every backend service, controller helper, and CLI command that reads or writes content so they use junction tables.
- Zero API contract changes — the frontend continues to pass `gameLevelId` and receive the same response shapes.
- Zero broken functionality: single play, PK, group play, ai-custom (sentence + vocab), course editor CRUD, publish/withdraw, import_courses all keep working.

## Non-Goals

- Implementing the actual reuse feature (UX for linking existing content to another level). That is a follow-up task.
- Changing the `content_items` PK or adding Postgres partitioning. Partitioning is a future concern and does not block this work.
- Frontend changes — intentionally out of scope.
- Rewriting the `20260407000001_create_game_junction_tables.go` stub (already registered as a no-op in some DBs).

## Schema

### Modified: `content_metas` (0 rows, easy migration)

- **REMOVE columns:** `game_level_id`, `order`
- **REMOVE indexes:** the `order` index
- **KEEP:** `id`, `source_from`, `source_type`, `source_data`, `translation`, `is_break_done`, `created_at`, `updated_at`, `deleted_at`
- **KEEP indexes:** `source_from`, `source_type`, `created_at`

### Modified: `content_items` (1,220,803 rows)

- **REMOVE columns:** `game_level_id`, `order`, `is_active`
- **REMOVE indexes:** the `order` and `is_active` indexes
- **KEEP:** `id`, `content_meta_id` (nullable, preserved as **immutable parsing lineage** — "this item was parsed from that source meta"), `content`, `content_type`, `uk_audio_id`, `us_audio_id`, `definition`, `translation`, `explanation`, `items` (jsonb), `structure` (jsonb), `tags` (text[]), `created_at`, `updated_at`, `deleted_at`
- **KEEP indexes:** `content_meta_id`, `uk_audio_id`, `us_audio_id`, `content_type`, `created_at`

### New: `game_metas`

```
id                UUID PK
game_id           UUID              -- denormalized, avoids JOIN to game_levels for game-wide queries
game_level_id     UUID
content_meta_id   UUID
order             double precision  -- per-level meta ordering (authoritative)
created_at        timestamptz
updated_at        timestamptz
deleted_at        timestamptz NULL  -- soft delete

UNIQUE (game_level_id, content_meta_id) WHERE deleted_at IS NULL
INDEX (game_level_id, deleted_at, "order")     -- editor load
INDEX (content_meta_id)                         -- inverse lookup, reuse detection
INDEX (game_id)
INDEX (created_at)
```

### New: `game_items`

```
id                UUID PK
game_id           UUID              -- denormalized
game_level_id     UUID
content_item_id   UUID
order             double precision  -- per-level item ordering (authoritative)
created_at        timestamptz
updated_at        timestamptz
deleted_at        timestamptz NULL

UNIQUE (game_level_id, content_item_id) WHERE deleted_at IS NULL
INDEX (game_level_id, deleted_at, "order")     -- play hot path; covers countLevelItems + GetLevelContent
INDEX (content_item_id)                         -- inverse lookup, reuse detection
INDEX (game_id)
INDEX (created_at)
```

**No `is_active` field on `game_items`.** The existing `content_items.is_active` is dead — removing it everywhere is a no-op change to behavior. Soft-delete (`deleted_at`) handles the only real disabling semantics: linking vs unlinking a content item from a level.

## Migrations

Three new migration files, registered in `dx-api/bootstrap/migrations.go` in order:

### Migration 1: `20260414000001_create_game_metas_and_game_items_tables.go`

DDL only. Uses `facades.Schema().Create()` with `HasTable()` idempotency guards (matching the existing migration style in the repo). Creates both tables with all columns, indexes, and soft-delete columns. Uses raw `Exec` for the partial unique indexes (`WHERE deleted_at IS NULL`) since Goravel's Blueprint doesn't expose partial indexes.

Down: `DropIfExists` both tables.

### Migration 2: `20260414000002_backfill_junction_tables.go`

Data move only. Two idempotent `INSERT ... SELECT ... ON CONFLICT DO NOTHING` statements. Joins to `game_levels` to derive `game_id` (since content tables don't have it directly).

```sql
-- content_metas → game_metas (0 rows expected, but re-runnable)
INSERT INTO game_metas (id, game_id, game_level_id, content_meta_id, "order", created_at, updated_at)
SELECT gen_random_uuid(), gl.game_id, cm.game_level_id, cm.id, cm."order", cm.created_at, cm.updated_at
FROM content_metas cm
JOIN game_levels gl ON gl.id = cm.game_level_id AND gl.deleted_at IS NULL
WHERE cm.deleted_at IS NULL
ON CONFLICT DO NOTHING;

-- content_items → game_items (~1.22M rows)
INSERT INTO game_items (id, game_id, game_level_id, content_item_id, "order", created_at, updated_at)
SELECT gen_random_uuid(), gl.game_id, ci.game_level_id, ci.id, ci."order", ci.created_at, ci.updated_at
FROM content_items ci
JOIN game_levels gl ON gl.id = ci.game_level_id AND gl.deleted_at IS NULL
WHERE ci.deleted_at IS NULL
ON CONFLICT DO NOTHING;
```

The `INSERT ... SELECT` for 1.22M rows runs as a single statement — Postgres handles this efficiently with sequential scan on `content_items` and bulk insert into the empty `game_items`. Expected runtime on modern hardware: tens of seconds to a few minutes. The target table has no indexes to maintain *during* the insert (Postgres builds indexes but the `UNIQUE` partial index is small). `ON CONFLICT DO NOTHING` makes the migration re-runnable safely.

Down: `DELETE FROM game_items; DELETE FROM game_metas;` (safe because in-migration-only data).

### Migration 3: `20260414000003_drop_legacy_columns_from_content_tables.go`

DDL only. Drops the legacy per-level columns from content tables. `ALTER TABLE DROP COLUMN` in Postgres is O(1) (marks the column as dropped in the catalog; space is reclaimed lazily on row rewrites).

```
ALTER TABLE content_metas DROP COLUMN game_level_id;
ALTER TABLE content_metas DROP COLUMN "order";

ALTER TABLE content_items DROP COLUMN game_level_id;
ALTER TABLE content_items DROP COLUMN "order";
ALTER TABLE content_items DROP COLUMN is_active;
```

Down: re-add the columns. The down path cannot restore prior values (that would require reading game_metas/game_items), but re-creates the shape so an emergency downgrade can run.

## Models

### New: `dx-api/app/models/game_meta.go`

```go
package models

import "github.com/goravel/framework/database/orm"

type GameMeta struct {
    orm.Timestamps
    orm.SoftDeletes
    ID            string  `gorm:"column:id;primaryKey" json:"id"`
    GameID        string  `gorm:"column:game_id" json:"game_id"`
    GameLevelID   string  `gorm:"column:game_level_id" json:"game_level_id"`
    ContentMetaID string  `gorm:"column:content_meta_id" json:"content_meta_id"`
    Order         float64 `gorm:"column:order" json:"order"`
}

func (g *GameMeta) TableName() string { return "game_metas" }
```

### New: `dx-api/app/models/game_item.go`

```go
package models

import "github.com/goravel/framework/database/orm"

type GameItem struct {
    orm.Timestamps
    orm.SoftDeletes
    ID            string  `gorm:"column:id;primaryKey" json:"id"`
    GameID        string  `gorm:"column:game_id" json:"game_id"`
    GameLevelID   string  `gorm:"column:game_level_id" json:"game_level_id"`
    ContentItemID string  `gorm:"column:content_item_id" json:"content_item_id"`
    Order         float64 `gorm:"column:order" json:"order"`
}

func (g *GameItem) TableName() string { return "game_items" }
```

### Modified: `dx-api/app/models/content_meta.go`

- Remove `GameLevelID string` field
- Remove `Order float64` field

### Modified: `dx-api/app/models/content_item.go`

- Remove `GameLevelID string` field
- Remove `Order float64` field
- Remove `IsActive bool` field
- Keep `ContentMetaID *string` (immutable parsing lineage)

## Code Changes

Reads and writes that currently go through `content_metas.game_level_id` / `content_items.game_level_id` must now go via junction tables. Joined tables in raw JOINs need explicit `AND deleted_at IS NULL` because Goravel's soft-delete auto-filter only covers the primary model.

### Write path — every content creation gets a sibling junction row

**`dx-api/app/services/api/course_content_service.go`**
- `SaveMetadataBatch()` — after inserting each `ContentMeta`, insert a matching `GameMeta` with `GameID` (derived once from the level), `GameLevelID`, and `Order` computed from max existing `game_metas.order` for that level.
- `InsertContentItem()` — after inserting a `ContentItem`, insert a matching `GameItem`. Order computed from max `game_items.order` for the level.
- `calculateInsertionOrder()` — query `MAX("order") FROM game_items WHERE game_level_id = ?` (was `FROM content_items`).

**`dx-api/app/services/api/ai_custom_service.go`**
- `processBreakMeta()` — when creating each `ContentItem` from AI-parsed units, also create a `GameItem`. `GameID` and `GameLevelID` are passed in or derived once per call.
- Queries in `BreakMetadata()` / `GenerateContentItems()` that look up metas/items by `game_level_id` now go through `game_metas` / `game_items`.

**`dx-api/app/services/api/ai_custom_vocab_service.go`**
- Same pattern as `ai_custom_service.go` — `processVocabBreakMeta()` creates a `GameItem` alongside each `ContentItem`.

**`dx-api/app/console/commands/import_courses.go`**
- Reintroduce `createGameItemsBatch()` and `createGameMetasBatch()` helpers (these existed on the original junction branch) to insert junction rows in the same transaction as content rows.
- `insertLevels()` calls both helpers after content inserts.

### Read path — `WHERE game_level_id = ?` becomes `JOIN game_items/game_metas`

**`dx-api/app/services/api/content_service.go` — `GetLevelContent()`**

```go
query.Model(&models.ContentItem{}).
    Select("content_items.*").
    Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
    Where("gi.game_level_id", gameLevelID).
    Where("content_items.deleted_at IS NULL").   // required: auto-filter doesn't cover joined primary
    // optional degree filter via ci.content_type IN ?
    OrderBy("gi.\"order\"")
```

No more `is_active` filter (column removed). Audio loads continue via `uk_audio_id` / `us_audio_id` batch loads, unchanged.

**`dx-api/app/services/api/game_play_single_service.go` — `countLevelItems()`**

Shared by single play, PK play, and group play. The single function drives all three.

```go
q := query.Model(&models.GameItem{}).
    Select("COUNT(*)").
    Join("JOIN content_items ci ON ci.id = game_items.content_item_id AND ci.deleted_at IS NULL").
    Where("game_items.game_level_id", gameLevelID)
if allowedTypes, ok := consts.DegreeContentTypes[degree]; ok && allowedTypes != nil {
    q = q.Where("ci.content_type IN ?", allowedTypes)
}
// Count()
```

Index `game_items(game_level_id, deleted_at, order)` covers this; the JOIN to `content_items` is a PK lookup per row.

**`dx-api/app/services/api/game_play_pk_service.go` — `spawnRobotForLevel()`**

Same JOIN pattern for robot content fetch. No independent `content_items.game_level_id` query.

**`dx-api/app/services/api/ai_custom_service.go` / `ai_custom_vocab_service.go`**

`BreakMetadata()`, `BreakVocabMetadata()`, `GenerateContentItems()`, `GenerateVocabContentItems()`:
- Replace `SELECT cm.* FROM content_metas cm WHERE cm.game_level_id = ?` with `SELECT cm.* FROM content_metas cm JOIN game_metas gm ON gm.content_meta_id = cm.id AND gm.deleted_at IS NULL WHERE gm.game_level_id = ? AND cm.deleted_at IS NULL`.
- Pending-item queries (`items IS NULL`) similarly JOIN through `game_items`.

**`dx-api/app/services/api/course_content_service.go` — `GetContentItemsByMeta()`**

Returns metas grouped with their items for the editor view. Joins `game_metas` for the level's metas, then `content_items` (via `content_meta_id`) plus `game_items` (via `game_level_id + content_item_id`) for the level-specific item order.

**`verifyMetaBelongsToGame()` / `verifyItemBelongsToGame()`**
- Verify via `game_metas`/`game_items` with a `game_levels` join to check `game_id`:
  ```sql
  SELECT 1 FROM game_metas gm
  JOIN game_levels gl ON gl.id = gm.game_level_id AND gl.deleted_at IS NULL
  WHERE gm.content_meta_id = ? AND gl.game_id = ? AND gm.deleted_at IS NULL
  LIMIT 1
  ```

**`dx-api/app/services/api/course_game_service.go` — `PublishGame()`, `GetCourseGameDetail()`**
- `PublishGame()` validation: "each level has items" + "every item has `items IS NOT NULL`" — both re-expressed as junction JOINs. Remove `is_active` filter.
- `GetCourseGameDetail()` per-level item count: `SELECT COUNT(*) FROM game_items WHERE game_level_id = ? AND deleted_at IS NULL`.

### Delete path — soft-delete junction rows alongside content rows

Pre-reuse, the 1:1 relationship holds: every `content_item` has exactly one `game_item`. Deletes cascade through both.

**`course_content_service.go`**
- `DeleteContentItem()` — soft-delete `game_items` row by `content_item_id + game_level_id`, then soft-delete `content_items`. If the parent meta has no remaining items (count via `game_items` JOIN), reset `content_metas.is_break_done = false`.
- `DeleteAllLevelContent()` — soft-delete `game_items WHERE game_level_id = ?`, soft-delete `game_metas WHERE game_level_id = ?`, then soft-delete `content_items` / `content_metas` whose IDs appear in the just-unlinked junction rows. 1:1 assumption holds, so all affected rows are owned by this level.
- `DeleteMetadata()` — soft-delete `game_metas` row, soft-delete `game_items` rows for that meta + level (via JOIN on `content_items.content_meta_id`), soft-delete the content rows.

**`course_game_service.go`**
- `DeleteGame()` — collect level IDs, then soft-delete `game_items WHERE game_level_id IN (?)`, `game_metas WHERE game_level_id IN (?)`, `content_items WHERE id IN (<collected>)`, `content_metas WHERE id IN (<collected>)`.
- `DeleteLevel()` — same scoped to one level.

**`import_courses.go` — `forceCleanup()`**
- Orphan cleanup targets junction rows too. `UPDATE game_items SET deleted_at = NOW() WHERE game_level_id IN (SELECT id FROM game_levels WHERE game_id = ? AND deleted_at IS NOT NULL)` and the analogous `game_metas` update.

### Files touched

| File | Change type |
|---|---|
| `app/models/content_meta.go` | edit (remove fields) |
| `app/models/content_item.go` | edit (remove fields) |
| `app/models/game_meta.go` | new |
| `app/models/game_item.go` | new |
| `app/services/api/content_service.go` | edit (read path) |
| `app/services/api/course_content_service.go` | edit (read + write + delete, ~8 functions) |
| `app/services/api/course_game_service.go` | edit (~4 functions) |
| `app/services/api/ai_custom_service.go` | edit (~3 functions) |
| `app/services/api/ai_custom_vocab_service.go` | edit (~3 functions) |
| `app/services/api/game_play_single_service.go` | edit (`countLevelItems`) |
| `app/services/api/game_play_pk_service.go` | edit (`spawnRobotForLevel`) |
| `app/console/commands/import_courses.go` | edit (`insertLevels`, `forceCleanup`, add batch helpers) |
| `database/migrations/20260414000001_create_game_metas_and_game_items_tables.go` | new |
| `database/migrations/20260414000002_backfill_junction_tables.go` | new |
| `database/migrations/20260414000003_drop_legacy_columns_from_content_tables.go` | new |
| `bootstrap/migrations.go` | edit (register 3 new migrations) |

Total: **4 model edits, 7 service edits, 1 CLI command edit, 3 new migrations, 1 bootstrap edit** = 16 files.

Frontend (`dx-web`): **no changes**. API response shapes are identical; the frontend continues to pass `gameLevelId` and receive content items.

## Query Efficiency

The two hot paths are `countLevelItems` (called on every session start) and `GetLevelContent` (called when loading a level for play).

Both query plans after this change:
```
game_items (game_level_id, deleted_at, order) index
  → index range scan (≤ few hundred rows per level typically)
  → JOIN content_items via PK (ci.id)
    → constant-time lookup per row
  → optional filter on ci.content_type (small result set, no scan)
  → LEFT JOIN audios via PK on uk_audio_id / us_audio_id
```

No sequential scans on the 1.22M-row `content_items` table. No N+1. The total work is bounded by the number of items in the level, which is typically < 200.

Postgres `EXPLAIN ANALYZE` should confirm this in verification.

## Verification

### Migration verification (per-step, on a staging copy)

1. **After migration 1:** `\d game_items` and `\d game_metas` — confirm columns, indexes, and partial unique constraints exist.
2. **After migration 2:**
   - `SELECT COUNT(*) FROM content_items WHERE deleted_at IS NULL` — baseline count (should match 1,220,803 minus any soft-deleted).
   - `SELECT COUNT(*) FROM game_items WHERE deleted_at IS NULL` — should match the content_items count exactly.
   - `SELECT COUNT(*) FROM content_metas WHERE deleted_at IS NULL` → should be 0; `SELECT COUNT(*) FROM game_metas WHERE deleted_at IS NULL` → also 0.
   - Spot-check 10 random content_items: each has exactly one `game_items` row with the same `"order"`, `created_at`, `game_level_id`; and the `game_id` matches the linked level's `game_id`.
   - `SELECT COUNT(*) FROM game_items gi LEFT JOIN content_items ci ON ci.id = gi.content_item_id WHERE ci.id IS NULL` — should be 0 (no orphan junction rows).
3. **After migration 3:** `\d content_items` and `\d content_metas` — confirm dropped columns are gone.

### Code verification

- `go build ./...` clean
- `go vet ./...` clean
- `staticcheck ./...` clean (if installed)
- `go test -race ./...` passes
- Smoke tests via the API:
  - Single play: start session → record 5 answers → complete level → restart → end session
  - AI custom sentence: generate-metadata → save → break-metadata (SSE) → generate-content-items (SSE) → publish
  - AI custom vocab: generate-vocab → save → break-vocab-metadata (SSE) → generate-vocab-content-items (SSE) → publish
  - Course editor: create game → create level → add content_meta → add content_item → reorder → delete → cascade cleanup
  - PK: create robot PK → play → verify content loads correctly
  - Group play: create group → start game → load content → verify same items as single play

### Row-count invariant

Before migration, after migration 2, after migration 3:
```
SELECT COUNT(*) FROM content_items WHERE deleted_at IS NULL
= SELECT COUNT(*) FROM game_items   WHERE deleted_at IS NULL
```

Must hold at every step.

## Rollout

1. Back up the production DB (critical — 1.22M-row table is touched).
2. Deploy during a low-traffic window. Concurrent writes to content during migration 2 are not expected since the ai-custom paths are admin-triggered and can be paused briefly.
3. Migrations run at app startup (Goravel style). Monitor the startup log for the three new migration signatures.
4. Watch error rates on `/api/play-single/*`, `/api/ai-custom/*`, `/api/course-games/*` after startup.
5. If rollback is needed:
   - Migration 3 down: re-add columns (empty; data would need to be re-derived from junction tables, which is possible but requires a separate script).
   - Migration 2 down: `DELETE FROM game_items; DELETE FROM game_metas`.
   - Migration 1 down: drop both tables.
   - Code must also roll back in the same deploy — runtime cannot straddle junction vs. direct FK states.

## Risks & Mitigations

| Risk | Mitigation |
|---|---|
| Migration 2 takes too long on 1.22M rows | Single statement, no triggers, no indexes-built-twice; expect minutes at most. Run during maintenance window. |
| Code update misses a `WHERE game_level_id = ?` on content tables | After migration 3 the column is gone — compilation fails. Hard dependency. |
| A raw JOIN forgets `AND gi.deleted_at IS NULL` | Grep for all new `JOIN game_items`/`JOIN game_metas` in the PR and manually verify each has the deleted_at clause (matches the pattern from the prior soft-delete refactor). |
| `countLevelItems` returns a different count than before | Spot-check pre/post migration counts for a handful of levels; they must match exactly. |
| Orphaned `game_items` with missing `content_items` | Migration 2 uses a JOIN that guarantees every junction row points to a valid content row. Add verification query (see above). |
| Concurrency: an admin creates content mid-migration | Short maintenance window; admin paths are low-traffic. Optional: add a read-only flag during migration 2. |
| Existing empty stub migration `20260407000001` confuses someone | Keep it untouched with a comment clarifying it's a historical no-op. |

## Open Questions

None remaining after brainstorming. All design choices confirmed with the user:

- Target: junction-only reads, drop `game_level_id` from content tables ✓
- Per-level attributes (`order`, `is_active`): move to junction, remove `is_active` entirely as it's dead code ✓
- Migration shape: three separate files (create → backfill → drop) ✓
- Content reuse itself deferred to a later task ✓

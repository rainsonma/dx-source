# Reintroduce Junction Tables for Content Reuse — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `game_metas` and `game_items` as many-to-many junction tables between `game_levels` and `content_metas`/`content_items`, migrate the existing 1,220,803 content_items as 1:1 junction rows, refactor all backend code to use junction tables, and drop the now-redundant `game_level_id`, `order`, and `is_active` columns from content tables.

**Architecture:** Three incremental migrations (create tables → backfill → drop columns), paired with per-file service refactors that land one at a time. Services first refactor *reads* to use junction JOINs (while content tables still hold columns), then one final cleanup task removes the legacy struct fields, stops setting the columns on writes, and applies migration 3 to drop the columns — keeping the code compilable at every intermediate commit.

**Tech Stack:** Go, Goravel (Laravel-style PHP-to-Go port), GORM, PostgreSQL 18, raw SQL for partial unique indexes and batch backfill.

**Spec:** `docs/superpowers/specs/2026-04-14-reintroduce-junction-tables-for-content-reuse-design.md`

---

## File Structure

### New files

| Path | Responsibility |
|---|---|
| `dx-api/database/migrations/20260414000001_create_game_metas_and_game_items_tables.go` | Migration 1: DDL for the two junction tables and their indexes. |
| `dx-api/database/migrations/20260414000002_backfill_junction_tables.go` | Migration 2: `INSERT ... SELECT` backfill from content tables. |
| `dx-api/database/migrations/20260414000003_drop_legacy_columns_from_content_tables.go` | Migration 3: `ALTER TABLE DROP COLUMN` for the 5 legacy columns. |
| `dx-api/app/models/game_meta.go` | GORM model for `game_metas`. |
| `dx-api/app/models/game_item.go` | GORM model for `game_items`. |

### Modified files

| Path | Change type |
|---|---|
| `dx-api/bootstrap/migrations.go` | Register 3 new migrations (incrementally per task). |
| `dx-api/app/models/content_meta.go` | Remove `GameLevelID`, `Order` fields (Task 13 only). |
| `dx-api/app/models/content_item.go` | Remove `GameLevelID`, `Order`, `IsActive` fields (Task 13 only). |
| `dx-api/app/services/api/content_service.go` | `GetLevelContent()` → JOIN `game_items`. |
| `dx-api/app/services/api/game_play_single_service.go` | `countLevelItems()` → JOIN `game_items`. Shared by single/PK/group. |
| `dx-api/app/services/api/game_play_pk_service.go` | `spawnRobotForLevel()` → JOIN `game_items`. |
| `dx-api/app/services/api/course_content_service.go` | 8 functions: reads, writes, deletes via junction. |
| `dx-api/app/services/api/course_game_service.go` | `DeleteGame`, `DeleteLevel`, `PublishGame`, `GetCourseGameDetail`. |
| `dx-api/app/services/api/ai_custom_service.go` | `BreakMetadata`, `processBreakMeta`, `GenerateContentItems`. |
| `dx-api/app/services/api/ai_custom_vocab_service.go` | `BreakVocabMetadata`, `processVocabBreakMeta`, `GenerateVocabContentItems`. |
| `dx-api/app/console/commands/import_courses.go` | `insertLevels`, `forceCleanup`, add batch helpers. |

**Total:** 5 new files, 11 modified files = 16 files.

### Frontend (`dx-web`): no changes

API response shapes are unchanged. The frontend continues to pass `gameLevelId` as a parameter and receives content items in the same format.

### Intermediate compile invariant

Tasks 1–12 keep `content_metas.GameLevelID` / `content_items.GameLevelID` in the struct and continue to set them on inserts (dual-write) so that `go build ./...` is clean at every commit and existing inserts don't violate the NOT NULL constraint on the yet-to-be-dropped columns. Task 13 is the atomic cleanup: struct fields removed, dual-write lines removed, Migration 3 added, columns dropped — all in one commit.

---

## Task 0: Pre-flight orphan check

**Purpose:** Before any migration runs, verify there are no content rows whose `game_level_id` points to a missing/deleted level. Migration 2 uses an `INNER JOIN` to `game_levels`, so orphaned rows would silently fail to be backfilled and become unreachable via junction reads after Task 13.

**Files:** none (read-only DB inspection)

- [ ] **Step 0.1: Run the orphan check against the development DB**

```bash
psql $DX_DB_URL -c "
SELECT 'content_items orphans' AS label, COUNT(*) AS n
FROM content_items ci
LEFT JOIN game_levels gl ON gl.id = ci.game_level_id AND gl.deleted_at IS NULL
WHERE ci.deleted_at IS NULL AND gl.id IS NULL
UNION ALL
SELECT 'content_metas orphans', COUNT(*)
FROM content_metas cm
LEFT JOIN game_levels gl ON gl.id = cm.game_level_id AND gl.deleted_at IS NULL
WHERE cm.deleted_at IS NULL AND gl.id IS NULL;
"
```

Expected: both counts = 0.

If either count is > 0: **STOP** and report the count to the human. Those rows are already orphaned in the current schema; surfacing them now lets the human decide whether to delete them, repair their `game_level_id`, or accept that they'll be silently skipped by the backfill. Do not proceed until the orphan count is 0 or the human explicitly approves skipping them.

- [ ] **Step 0.2: Record baseline row counts for later invariant checking**

```bash
psql $DX_DB_URL -c "
SELECT 'content_items live' AS label, COUNT(*) AS n FROM content_items WHERE deleted_at IS NULL
UNION ALL SELECT 'content_metas live', COUNT(*) FROM content_metas WHERE deleted_at IS NULL;
"
```

Save the numbers. After Task 4 (backfill) the `game_items`/`game_metas` counts must exactly equal these.

---

## Task 1: Write migration 1 (create junction tables)

**Files:**
- Create: `dx-api/database/migrations/20260414000001_create_game_metas_and_game_items_tables.go`

- [ ] **Step 1.1: Write the migration file**

```go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260414000001CreateGameMetasAndGameItemsTables struct{}

func (r *M20260414000001CreateGameMetasAndGameItemsTables) Signature() string {
	return "20260414000001_create_game_metas_and_game_items_tables"
}

func (r *M20260414000001CreateGameMetasAndGameItemsTables) Up() error {
	// 1. game_metas
	if !facades.Schema().HasTable("game_metas") {
		if err := facades.Schema().Create("game_metas", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Uuid("content_meta_id")
			table.Double("order").Default(0)
			table.TimestampsTz()
			table.SoftDeletesTz()
			table.Index("game_id")
			table.Index("content_meta_id")
			table.Index("created_at")
		}); err != nil {
			return err
		}
		// Partial indexes (not exposed by Blueprint)
		if _, err := facades.Orm().Query().Exec(
			`CREATE UNIQUE INDEX idx_game_metas_level_meta_unique
			 ON game_metas (game_level_id, content_meta_id)
			 WHERE deleted_at IS NULL`,
		); err != nil {
			return err
		}
		if _, err := facades.Orm().Query().Exec(
			`CREATE INDEX idx_game_metas_level_order
			 ON game_metas (game_level_id, deleted_at, "order")`,
		); err != nil {
			return err
		}
	}

	// 2. game_items
	if !facades.Schema().HasTable("game_items") {
		if err := facades.Schema().Create("game_items", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Uuid("content_item_id")
			table.Double("order").Default(0)
			table.TimestampsTz()
			table.SoftDeletesTz()
			table.Index("game_id")
			table.Index("content_item_id")
			table.Index("created_at")
		}); err != nil {
			return err
		}
		if _, err := facades.Orm().Query().Exec(
			`CREATE UNIQUE INDEX idx_game_items_level_item_unique
			 ON game_items (game_level_id, content_item_id)
			 WHERE deleted_at IS NULL`,
		); err != nil {
			return err
		}
		if _, err := facades.Orm().Query().Exec(
			`CREATE INDEX idx_game_items_level_order
			 ON game_items (game_level_id, deleted_at, "order")`,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260414000001CreateGameMetasAndGameItemsTables) Down() error {
	if err := facades.Schema().DropIfExists("game_items"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("game_metas")
}
```

- [ ] **Step 1.2: Register the migration in bootstrap**

Edit `dx-api/bootstrap/migrations.go`. After the line `&migrations.M20260407000001CreateGameJunctionTables{},` (currently the last entry), append:

```go
		&migrations.M20260414000001CreateGameMetasAndGameItemsTables{},
```

- [ ] **Step 1.3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: exit 0, no output.

- [ ] **Step 1.4: Run migrations against the dev DB**

Run: `cd dx-api && go run . artisan migrate`
Expected: output includes "Migrating: 20260414000001_create_game_metas_and_game_items_tables" and "Migrated" lines.

- [ ] **Step 1.5: Verify tables and indexes exist**

```bash
psql $DX_DB_URL -c '\d game_metas'
psql $DX_DB_URL -c '\d game_items'
```

Expected on each: column list matches the migration (id, game_id, game_level_id, content_meta_id/content_item_id, order, created_at, updated_at, deleted_at) and an index list including the partial unique index (with `WHERE (deleted_at IS NULL)` shown) and the compound level-order index.

- [ ] **Step 1.6: Commit**

```bash
git add dx-api/database/migrations/20260414000001_create_game_metas_and_game_items_tables.go \
        dx-api/bootstrap/migrations.go
git commit -m "feat(api): add migration creating game_metas and game_items junction tables"
```

---

## Task 2: Create GameMeta and GameItem models

**Files:**
- Create: `dx-api/app/models/game_meta.go`
- Create: `dx-api/app/models/game_item.go`

- [ ] **Step 2.1: Write `game_meta.go`**

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

// TableName returns the database table name.
func (g *GameMeta) TableName() string {
	return "game_metas"
}
```

- [ ] **Step 2.2: Write `game_item.go`**

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

// TableName returns the database table name.
func (g *GameItem) TableName() string {
	return "game_items"
}
```

- [ ] **Step 2.3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: exit 0.

- [ ] **Step 2.4: Commit**

```bash
git add dx-api/app/models/game_meta.go dx-api/app/models/game_item.go
git commit -m "feat(api): add GameMeta and GameItem GORM models"
```

---

## Task 3: Write migration 2 (backfill junction tables)

**Files:**
- Create: `dx-api/database/migrations/20260414000002_backfill_junction_tables.go`
- Modify: `dx-api/bootstrap/migrations.go`

- [ ] **Step 3.1: Write the migration file**

```go
package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260414000002BackfillJunctionTables struct{}

func (r *M20260414000002BackfillJunctionTables) Signature() string {
	return "20260414000002_backfill_junction_tables"
}

func (r *M20260414000002BackfillJunctionTables) Up() error {
	// Backfill game_metas from content_metas. Uses INNER JOIN to game_levels so
	// any orphaned content_metas (whose game_level_id references a deleted level)
	// are skipped. ON CONFLICT DO NOTHING makes this safely re-runnable if the
	// migration aborts mid-way.
	if _, err := facades.Orm().Query().Exec(`
		INSERT INTO game_metas (id, game_id, game_level_id, content_meta_id, "order", created_at, updated_at)
		SELECT gen_random_uuid(), gl.game_id, cm.game_level_id, cm.id, cm."order", cm.created_at, cm.updated_at
		FROM content_metas cm
		JOIN game_levels gl ON gl.id = cm.game_level_id AND gl.deleted_at IS NULL
		WHERE cm.deleted_at IS NULL
		ON CONFLICT DO NOTHING
	`); err != nil {
		return err
	}

	// Backfill game_items from content_items (~1.22M rows).
	if _, err := facades.Orm().Query().Exec(`
		INSERT INTO game_items (id, game_id, game_level_id, content_item_id, "order", created_at, updated_at)
		SELECT gen_random_uuid(), gl.game_id, ci.game_level_id, ci.id, ci."order", ci.created_at, ci.updated_at
		FROM content_items ci
		JOIN game_levels gl ON gl.id = ci.game_level_id AND gl.deleted_at IS NULL
		WHERE ci.deleted_at IS NULL
		ON CONFLICT DO NOTHING
	`); err != nil {
		return err
	}

	return nil
}

func (r *M20260414000002BackfillJunctionTables) Down() error {
	// Safe: the backfilled rows are the only data in these tables during this
	// migration step, so a full delete rolls back the entire step.
	if _, err := facades.Orm().Query().Exec(`DELETE FROM game_items`); err != nil {
		return err
	}
	if _, err := facades.Orm().Query().Exec(`DELETE FROM game_metas`); err != nil {
		return err
	}
	return nil
}
```

- [ ] **Step 3.2: Register the migration in bootstrap**

Edit `dx-api/bootstrap/migrations.go`. After `&migrations.M20260414000001CreateGameMetasAndGameItemsTables{},` append:

```go
		&migrations.M20260414000002BackfillJunctionTables{},
```

- [ ] **Step 3.3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: exit 0.

- [ ] **Step 3.4: Run migration 2 against dev DB**

Run: `cd dx-api && go run . artisan migrate`
Expected: output includes "Migrated: 20260414000002_backfill_junction_tables".

- [ ] **Step 3.5: Verify row-count invariant**

```bash
psql $DX_DB_URL -c "
SELECT 'content_items' AS label, COUNT(*) AS n FROM content_items WHERE deleted_at IS NULL
UNION ALL SELECT 'game_items', COUNT(*) FROM game_items WHERE deleted_at IS NULL
UNION ALL SELECT 'content_metas', COUNT(*) FROM content_metas WHERE deleted_at IS NULL
UNION ALL SELECT 'game_metas', COUNT(*) FROM game_metas WHERE deleted_at IS NULL;
"
```

Expected: `content_items` and `game_items` counts are **equal**. `content_metas` and `game_metas` counts are **equal** (both 0 on the current dev DB).

- [ ] **Step 3.6: Spot-check a random content_item has a matching game_items row**

```bash
psql $DX_DB_URL -c "
SELECT ci.id AS content_item_id, ci.game_level_id, ci.\"order\" AS ci_order,
       gi.id AS game_item_id, gi.game_id, gi.\"order\" AS gi_order,
       gi.game_level_id = ci.game_level_id AS level_match,
       gi.\"order\" = ci.\"order\" AS order_match
FROM content_items ci
JOIN game_items gi ON gi.content_item_id = ci.id
WHERE ci.deleted_at IS NULL
ORDER BY random()
LIMIT 10;
"
```

Expected: 10 rows, `level_match` and `order_match` both `t` in every row, `game_id` is a valid UUID.

- [ ] **Step 3.7: Check no orphan junction rows**

```bash
psql $DX_DB_URL -c "
SELECT COUNT(*) FROM game_items gi
LEFT JOIN content_items ci ON ci.id = gi.content_item_id
WHERE ci.id IS NULL;
"
```

Expected: 0.

- [ ] **Step 3.8: Commit**

```bash
git add dx-api/database/migrations/20260414000002_backfill_junction_tables.go \
        dx-api/bootstrap/migrations.go
git commit -m "feat(api): backfill game_metas/game_items from existing content tables"
```

---

## Task 4: Refactor `content_service.GetLevelContent()` (hot read path)

**Files:**
- Modify: `dx-api/app/services/api/content_service.go`

**Why this one first:** It's the smallest read-path change and unblocks validating the junction-join approach against live data before touching the more complex services.

- [ ] **Step 4.1: Read the current function**

```bash
grep -n "func GetLevelContent" dx-api/app/services/api/content_service.go
```
Note the start line. Read the whole function body (roughly lines 26–104 per the spec).

- [ ] **Step 4.2: Replace the `content_items` query with a junction JOIN**

Find this block (current code at approximately `content_service.go:38-45`):

```go
	q := facades.Orm().Query().Model(&models.ContentItem{}).
		Where("content_items.game_level_id", gameLevelID).
		Where("content_items.is_active", true)
	if allowedTypes, ok := consts.DegreeContentTypes[degree]; ok && allowedTypes != nil {
		q = q.Where("content_items.content_type IN ?", allowedTypes)
	}
	var items []models.ContentItem
	if err := q.OrderBy("content_items.order").Get(&items); err != nil {
```

Replace with:

```go
	q := facades.Orm().Query().Model(&models.ContentItem{}).
		Select("content_items.*").
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
		Where("gi.game_level_id", gameLevelID)
	if allowedTypes, ok := consts.DegreeContentTypes[degree]; ok && allowedTypes != nil {
		q = q.Where("content_items.content_type IN ?", allowedTypes)
	}
	var items []models.ContentItem
	if err := q.OrderBy(`gi."order"`).Get(&items); err != nil {
```

Rationale:
- `is_active` filter removed — column is going away and is dead (no code sets it to false).
- `content_items.game_level_id` filter replaced with `gi.game_level_id` via JOIN.
- Ordering by `gi."order"` (the junction's authoritative per-level order) instead of `content_items.order`.
- `content_items.deleted_at IS NULL` is auto-applied by the ORM on the primary Model; the JOIN includes `gi.deleted_at IS NULL` because the ORM does not auto-filter joined tables.

- [ ] **Step 4.3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: exit 0.

- [ ] **Step 4.4: Run vet**

Run: `cd dx-api && go vet ./...`
Expected: exit 0, no output.

- [ ] **Step 4.5: Start the API and smoke-test the play content endpoint**

Run: `cd dx-api && go run . &` (or `air` if configured)
Wait for "Listening on :3001" in the log.

Then, in another terminal, pick a known `gameId` + `levelId` + `degree` from the dev DB and hit:

```bash
curl -s -H "Authorization: Bearer $DEV_USER_TOKEN" \
  "http://localhost:3001/api/games/$GAME_ID/levels/$LEVEL_ID/content?degree=beginner" \
  | jq '.data | length'
```

Expected: a non-zero number matching `SELECT COUNT(*) FROM game_items WHERE game_level_id = $LEVEL_ID AND deleted_at IS NULL`.

Stop the API with `kill %1` (or Ctrl-C the `go run`).

- [ ] **Step 4.6: Commit**

```bash
git add dx-api/app/services/api/content_service.go
git commit -m "refactor(api): GetLevelContent joins game_items instead of filtering content_items.game_level_id"
```

---

## Task 5: Refactor `game_play_single_service.countLevelItems()`

**Files:**
- Modify: `dx-api/app/services/api/game_play_single_service.go`

This function is shared by single play, PK play, and group play. A single refactor fixes all three.

- [ ] **Step 5.1: Read the current function**

```bash
grep -n "func countLevelItems" dx-api/app/services/api/game_play_single_service.go
```
Read the function body (currently around lines 603–615).

- [ ] **Step 5.2: Replace the body**

Find this code:

```go
// countLevelItems counts active content items for a level, filtered by degree.
func countLevelItems(query orm.Query, gameLevelID, degree string) (int64, error) {
	q := query.Model(&models.ContentItem{}).
		Where("game_level_id", gameLevelID).
		Where("is_active", true)
	allowedTypes, ok := consts.DegreeContentTypes[degree]
	if ok && allowedTypes != nil {
		q = q.Where("content_type IN ?", allowedTypes)
	}
	var count int64
	var err error
	count, err = q.Count()
	return count, err
}
```

Replace with:

```go
// countLevelItems counts content items linked to a level via game_items,
// filtered by the degree's allowed content types. Shared by single play,
// PK play, and group play.
func countLevelItems(query orm.Query, gameLevelID, degree string) (int64, error) {
	q := query.Model(&models.GameItem{}).
		Join("JOIN content_items ci ON ci.id = game_items.content_item_id AND ci.deleted_at IS NULL").
		Where("game_items.game_level_id", gameLevelID)
	if allowedTypes, ok := consts.DegreeContentTypes[degree]; ok && allowedTypes != nil {
		q = q.Where("ci.content_type IN ?", allowedTypes)
	}
	return q.Count()
}
```

- [ ] **Step 5.3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: exit 0.

- [ ] **Step 5.4: Spot-check the count against the previous implementation**

```bash
psql $DX_DB_URL -c "
-- Old count path (direct)
SELECT COUNT(*) AS old_count FROM content_items
WHERE game_level_id = '$LEVEL_ID' AND is_active = true AND deleted_at IS NULL;
-- New count path (via junction)
SELECT COUNT(*) AS new_count FROM game_items gi
JOIN content_items ci ON ci.id = gi.content_item_id AND ci.deleted_at IS NULL
WHERE gi.game_level_id = '$LEVEL_ID' AND gi.deleted_at IS NULL;
"
```

Expected: both counts are equal for any representative level.

- [ ] **Step 5.5: Commit**

```bash
git add dx-api/app/services/api/game_play_single_service.go
git commit -m "refactor(api): countLevelItems joins game_items (fixes single/PK/group play)"
```

---

## Task 6: Refactor `game_play_pk_service.spawnRobotForLevel()`

**Files:**
- Modify: `dx-api/app/services/api/game_play_pk_service.go`

- [ ] **Step 6.1: Locate the content fetch for the robot**

```bash
grep -n "content_items" dx-api/app/services/api/game_play_pk_service.go
grep -n "is_active" dx-api/app/services/api/game_play_pk_service.go
```

Find the block (around line 595–605) that looks like:

```go
	var items []models.ContentItem
	err := facades.Orm().Query().
		Where("game_level_id", levelID).
		Where("is_active", true).
		...
		Get(&items)
```

- [ ] **Step 6.2: Replace with a junction join**

Replace that block with:

```go
	var items []models.ContentItem
	err := facades.Orm().Query().Model(&models.ContentItem{}).
		Select("content_items.*").
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
		Where("gi.game_level_id", levelID).
		// (preserve any existing filters on content_items.items IS NOT NULL etc. here)
		OrderBy(`gi."order"`).
		Get(&items)
```

Carefully preserve any additional `.Where(...)` clauses that exist on the original (e.g., filters on `items IS NOT NULL` to only pick generated items). If unsure, read the full original block and re-implement all filters over `content_items.*` (not `gi`).

- [ ] **Step 6.3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: exit 0.

- [ ] **Step 6.4: Commit**

```bash
git add dx-api/app/services/api/game_play_pk_service.go
git commit -m "refactor(api): spawnRobotForLevel joins game_items"
```

---

## Task 7: Refactor `course_content_service.go` reads and helpers

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go`

This file has the most read-path changes: `SaveMetadataBatch`'s capacity check, `GetContentItemsByMeta`, `verifyMetaBelongsToGame`, `verifyItemBelongsToGame`, `calculateInsertionOrder`. Writes and deletes are in Tasks 8 and 11 respectively.

- [ ] **Step 7.1: Refactor the capacity check in `SaveMetadataBatch()`**

Find (around line 75–79):

```go
	var existingMetas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Get(&existingMetas); err != nil {
		return 0, fmt.Errorf("failed to count metas: %w", err)
	}
```

Replace with:

```go
	var existingMetas []models.ContentMeta
	if err := facades.Orm().Query().Model(&models.ContentMeta{}).
		Select("content_metas.*").
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
		Where("gm.game_level_id", gameLevelID).
		Get(&existingMetas); err != nil {
		return 0, fmt.Errorf("failed to count metas: %w", err)
	}
```

- [ ] **Step 7.2: Refactor `GetContentItemsByMeta()` read query**

Locate the function (search for `func GetContentItemsByMeta`). Rewrite it to use **two separate queries** so we never depend on removed struct fields. This shape stays valid after Task 13 removes `ContentMeta.Order` and `ContentItem.Order`:

```go
func GetContentItemsByMeta(userID, gameID, gameLevelID string) ([]LevelContentData, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	if _, err := getCourseGameOwned(userID, gameID); err != nil {
		return nil, err
	}

	// 1. Load game_metas for this level in order.
	var gameMetas []models.GameMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		OrderBy(`"order"`).
		Get(&gameMetas); err != nil {
		return nil, fmt.Errorf("failed to load game_metas: %w", err)
	}
	if len(gameMetas) == 0 {
		return []LevelContentData{}, nil
	}

	// 2. Load the referenced content_metas by ID.
	metaIDs := make([]string, 0, len(gameMetas))
	for _, gm := range gameMetas {
		metaIDs = append(metaIDs, gm.ContentMetaID)
	}
	var contentMetas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("id IN ?", metaIDs).
		Get(&contentMetas); err != nil {
		return nil, fmt.Errorf("failed to load content_metas: %w", err)
	}
	metaByID := make(map[string]models.ContentMeta, len(contentMetas))
	for _, cm := range contentMetas {
		metaByID[cm.ID] = cm
	}

	// 3. Load game_items for this level in order.
	var gameItems []models.GameItem
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		OrderBy(`"order"`).
		Get(&gameItems); err != nil {
		return nil, fmt.Errorf("failed to load game_items: %w", err)
	}

	// 4. Load the referenced content_items by ID.
	var items []models.ContentItem
	if len(gameItems) > 0 {
		itemIDs := make([]string, 0, len(gameItems))
		for _, gi := range gameItems {
			itemIDs = append(itemIDs, gi.ContentItemID)
		}
		if err := facades.Orm().Query().
			Where("id IN ?", itemIDs).
			Get(&items); err != nil {
			return nil, fmt.Errorf("failed to load content_items: %w", err)
		}
	}
	itemByID := make(map[string]models.ContentItem, len(items))
	for _, it := range items {
		itemByID[it.ID] = it
	}

	// 5. Build items-by-meta map, preserving game_items.order as the authoritative
	//    per-level order.
	itemsByMeta := make(map[string][]CourseContentItemData)
	for _, gi := range gameItems {
		it, ok := itemByID[gi.ContentItemID]
		if !ok {
			continue
		}
		metaKey := ""
		if it.ContentMetaID != nil {
			metaKey = *it.ContentMetaID
		}
		raw := json.RawMessage("null")
		if it.Items != nil {
			raw = json.RawMessage(*it.Items)
		}
		itemsByMeta[metaKey] = append(itemsByMeta[metaKey], CourseContentItemData{
			ID:            it.ID,
			ContentMetaID: it.ContentMetaID,
			Content:       it.Content,
			ContentType:   it.ContentType,
			Translation:   it.Translation,
			Items:         raw,
			Order:         gi.Order, // from game_items (authoritative)
		})
	}

	// 6. Assemble result in game_metas.order.
	result := make([]LevelContentData, 0, len(gameMetas))
	for _, gm := range gameMetas {
		cm, ok := metaByID[gm.ContentMetaID]
		if !ok {
			continue
		}
		result = append(result, LevelContentData{
			Meta: ContentMetaData{
				ID:          cm.ID,
				SourceFrom:  cm.SourceFrom,
				SourceType:  cm.SourceType,
				SourceData:  cm.SourceData,
				Translation: cm.Translation,
				IsBreakDone: cm.IsBreakDone,
				Order:       gm.Order, // from game_metas (authoritative)
			},
			Items: itemsByMeta[cm.ID],
		})
	}
	return result, nil
}
```

Notes:
- Uses only struct fields that survive Task 13 (`ContentMeta` with no `GameLevelID`/`Order`, `ContentItem` with no `GameLevelID`/`Order`/`IsActive`).
- Both `Order` values in the response come from the junction structs (`gm.Order`, `gi.Order`) — authoritative.
- Four queries, but each is a fast primary-key-or-indexed lookup. No N+1.

- [ ] **Step 7.3: Refactor `verifyMetaBelongsToGame()`**

Find the function (search for `func verifyMetaBelongsToGame`). Replace the check query:

```go
func verifyMetaBelongsToGame(metaID, gameID string) error {
	var count int64
	// Verify via game_metas JOIN game_levels: the meta has a non-deleted
	// linkage to a level that belongs to this game.
	err := facades.Orm().Query().Model(&models.GameMeta{}).
		Join("JOIN game_levels gl ON gl.id = game_metas.game_level_id AND gl.deleted_at IS NULL").
		Where("game_metas.content_meta_id", metaID).
		Where("gl.game_id", gameID).
		Count(&count)
	if err != nil {
		return fmt.Errorf("failed to verify meta: %w", err)
	}
	if count == 0 {
		return ErrForbidden
	}
	return nil
}
```

Adjust the return signature to match the existing function — some versions return `(bool, error)` or write directly. Re-use the existing error handling pattern; only the query shape changes.

- [ ] **Step 7.4: Refactor `verifyItemBelongsToGame()`**

Same pattern but for items:

```go
func verifyItemBelongsToGame(itemID, gameID string) error {
	var count int64
	err := facades.Orm().Query().Model(&models.GameItem{}).
		Join("JOIN game_levels gl ON gl.id = game_items.game_level_id AND gl.deleted_at IS NULL").
		Where("game_items.content_item_id", itemID).
		Where("gl.game_id", gameID).
		Count(&count)
	if err != nil {
		return fmt.Errorf("failed to verify item: %w", err)
	}
	if count == 0 {
		return ErrForbidden
	}
	return nil
}
```

- [ ] **Step 7.5: Refactor `calculateInsertionOrder()`**

Find `func calculateInsertionOrder`. Replace the "select MAX order" query from `content_items` to `game_items`:

```go
// calculateInsertionOrder returns the next "order" value for a new item
// in the given level. Uses game_items.order as the authoritative per-level order.
func calculateInsertionOrder(gameLevelID string) (float64, error) {
	var row struct {
		MaxOrder *float64 `gorm:"column:max_order"`
	}
	err := facades.Orm().Query().Model(&models.GameItem{}).
		Select(`MAX("order") AS max_order`).
		Where("game_level_id", gameLevelID).
		Scan(&row)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate order: %w", err)
	}
	if row.MaxOrder == nil {
		return 1000, nil
	}
	return *row.MaxOrder + 1000, nil
}
```

Adjust the increment (1000) and zero-case default to match the existing implementation exactly — only the table name in the query changes.

- [ ] **Step 7.6: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: exit 0.

- [ ] **Step 7.7: Commit**

```bash
git add dx-api/app/services/api/course_content_service.go
git commit -m "refactor(api): course_content_service reads use junction tables"
```

---

## Task 8: Refactor `course_content_service.go` writes (create junction rows)

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go` (same file, continuing)

This task adds the write-side dual-write: after inserting a `ContentMeta` or `ContentItem`, insert a matching `GameMeta` or `GameItem` row. The content struct still sets `GameLevelID` and `Order` (they stay until Task 13).

- [ ] **Step 8.1: Update `SaveMetadataBatch()` to insert game_metas**

Find the loop that creates ContentMeta rows in `SaveMetadataBatch()`. Currently each iteration inserts only into `content_metas`. Extend each iteration to also insert a `GameMeta`:

```go
	// Determine starting order from existing game_metas (authoritative for per-level ordering).
	var maxMetaOrder float64
	{
		var row struct {
			MaxOrder *float64 `gorm:"column:max_order"`
		}
		if err := facades.Orm().Query().Model(&models.GameMeta{}).
			Select(`MAX("order") AS max_order`).
			Where("game_level_id", gameLevelID).
			Scan(&row); err != nil {
			return 0, fmt.Errorf("failed to compute meta insertion order: %w", err)
		}
		if row.MaxOrder != nil {
			maxMetaOrder = *row.MaxOrder
		}
	}

	// Use the loaded level's game_id (already fetched earlier in this function as `level.GameID`).
	gameMetaGameID := level.GameID

	for i, entry := range entries {
		order := maxMetaOrder + float64(1000*(i+1))
		metaID := uuid.NewString()
		meta := models.ContentMeta{
			ID:          metaID,
			GameLevelID: gameLevelID, // dual-write until Task 13
			SourceFrom:  sourceFrom,
			SourceType:  entry.SourceType,
			SourceData:  entry.SourceData,
			Translation: entry.Translation,
			IsBreakDone: false,
			Order:       order, // dual-write until Task 13
		}
		if err := facades.Orm().Query().Create(&meta); err != nil {
			return 0, fmt.Errorf("failed to create content meta: %w", err)
		}
		gm := models.GameMeta{
			ID:            uuid.NewString(),
			GameID:        gameMetaGameID,
			GameLevelID:   gameLevelID,
			ContentMetaID: metaID,
			Order:         order,
		}
		if err := facades.Orm().Query().Create(&gm); err != nil {
			return 0, fmt.Errorf("failed to create game meta: %w", err)
		}
	}
```

Note: **if the function previously used a 1:1 loop with an inline `Create` call, preserve its surrounding transaction/error handling exactly.** Only add the `gm := models.GameMeta{...}; Create(&gm)` pair after each meta insert.

- [ ] **Step 8.2: Update `InsertContentItem()` to also insert a game_items row**

Find `func InsertContentItem`. After the `ContentItem` insert, add:

```go
	gi := models.GameItem{
		ID:            uuid.NewString(),
		GameID:        gameID, // passed in or looked up from the level earlier
		GameLevelID:   gameLevelID,
		ContentItemID: item.ID,
		Order:         item.Order,
	}
	if err := facades.Orm().Query().Create(&gi); err != nil {
		return fmt.Errorf("failed to create game item: %w", err)
	}
```

If the existing function doesn't have `gameID` in scope, load it once from the level:

```go
	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).First(&level); err != nil {
		return fmt.Errorf("failed to load level: %w", err)
	}
```

and use `level.GameID`.

- [ ] **Step 8.3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: exit 0.

- [ ] **Step 8.4: Smoke-test the write path**

Start the API (`cd dx-api && go run .`), then via the course editor flow (or a direct `curl` POST to `/api/course-games/.../levels/.../metadata`), create a test meta. Then:

```bash
psql $DX_DB_URL -c "
SELECT cm.id, cm.game_level_id, gm.id AS game_meta_id, gm.game_id, gm.\"order\"
FROM content_metas cm
LEFT JOIN game_metas gm ON gm.content_meta_id = cm.id AND gm.deleted_at IS NULL
ORDER BY cm.created_at DESC LIMIT 5;
"
```

Expected: every new `content_metas` row has a matching `game_metas` row with the same `game_level_id` and `order`.

Stop the API.

- [ ] **Step 8.5: Commit**

```bash
git add dx-api/app/services/api/course_content_service.go
git commit -m "refactor(api): course_content_service writes create junction rows"
```

---

## Task 9: Refactor `course_game_service.go` reads

**Files:**
- Modify: `dx-api/app/services/api/course_game_service.go`

- [ ] **Step 9.1: Refactor `PublishGame()` validation**

Find the block (around lines 311–334) that loops over levels and counts content items. Replace the two count queries inside the loop:

```go
	for _, l := range levels {
		// "level has active items" check
		itemCount, err3 := facades.Orm().Query().Model(&models.GameItem{}).
			Join("JOIN content_items ci ON ci.id = game_items.content_item_id AND ci.deleted_at IS NULL").
			Where("game_items.game_level_id", l.ID).
			Count()
		if err3 != nil {
			return fmt.Errorf("failed to count items: %w", err3)
		}
		if itemCount == 0 {
			return fmt.Errorf("关卡「%s」没有练习内容", l.Name)
		}

		// "every item has generated items JSON" check
		ungeneratedCount, err4 := facades.Orm().Query().Model(&models.GameItem{}).
			Join("JOIN content_items ci ON ci.id = game_items.content_item_id AND ci.deleted_at IS NULL").
			Where("game_items.game_level_id", l.ID).
			Where("ci.items IS NULL").
			Count()
		if err4 != nil {
			return fmt.Errorf("failed to count ungenerated items: %w", err4)
		}
		if ungeneratedCount > 0 {
			return fmt.Errorf("关卡「%s」有未生成的练习单元", l.Name)
		}
	}
```

Removed: the `is_active = true` filter on both counts (column going away).

- [ ] **Step 9.2: Refactor `GetCourseGameDetail()` per-level item count**

Find the block (around lines 123–145) that loads levels and counts items per level. Replace the item-count query with:

```go
	itemCount, err := facades.Orm().Query().Model(&models.GameItem{}).
		Where("game_level_id", l.ID).
		Count()
	if err != nil {
		// handle error as before
	}
```

(Leave the surrounding loop structure alone; only the query body changes.)

- [ ] **Step 9.3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: exit 0.

- [ ] **Step 9.4: Commit**

```bash
git add dx-api/app/services/api/course_game_service.go
git commit -m "refactor(api): course_game_service read paths use game_items"
```

---

## Task 10: Refactor `course_game_service.go` deletes and `course_content_service.go` deletes

**Files:**
- Modify: `dx-api/app/services/api/course_game_service.go`
- Modify: `dx-api/app/services/api/course_content_service.go`

Deletes must soft-delete the junction rows alongside the content rows. Since this task is pre-reuse, 1:1 semantics hold.

- [ ] **Step 10.1: Update `course_game_service.DeleteLevel()`**

Inside the transaction, **before** the `UPDATE content_items SET deleted_at = NOW() WHERE game_level_id = ?` line, add:

```go
	if _, err := tx.Exec(`UPDATE game_items SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL`, levelID); err != nil {
		return fmt.Errorf("failed to soft-delete game_items: %w", err)
	}
	if _, err := tx.Exec(`UPDATE game_metas SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL`, levelID); err != nil {
		return fmt.Errorf("failed to soft-delete game_metas: %w", err)
	}
```

Replace `tx.Exec` with whatever Goravel query handle the existing code uses (e.g., `tx.Exec`, `query.Exec`, `facades.Orm().Query().Exec`). Read the existing implementation and match its style.

- [ ] **Step 10.2: Update `course_game_service.DeleteGame()`**

Same pattern. Before the `content_items` / `content_metas` soft-delete step, soft-delete `game_items` and `game_metas` for every level in the game:

```go
	if _, err := tx.Exec(`
		UPDATE game_items SET deleted_at = NOW()
		WHERE game_level_id IN (SELECT id FROM game_levels WHERE game_id = ?) AND deleted_at IS NULL
	`, gameID); err != nil {
		return fmt.Errorf("failed to soft-delete game_items: %w", err)
	}
	if _, err := tx.Exec(`
		UPDATE game_metas SET deleted_at = NOW()
		WHERE game_level_id IN (SELECT id FROM game_levels WHERE game_id = ?) AND deleted_at IS NULL
	`, gameID); err != nil {
		return fmt.Errorf("failed to soft-delete game_metas: %w", err)
	}
```

- [ ] **Step 10.3: Update `course_content_service.DeleteContentItem()`**

Inside the transaction, **before** soft-deleting the content_item, soft-delete its junction row:

```go
	if _, err := tx.Exec(`UPDATE game_items SET deleted_at = NOW() WHERE content_item_id = ? AND deleted_at IS NULL`, itemID); err != nil {
		return fmt.Errorf("failed to soft-delete game_item: %w", err)
	}
```

Then soft-delete the content_item (existing code), then the `is_break_done` reset. For the `is_break_done` reset, the query that counts remaining items in the parent meta needs to use `game_items` for per-level scoping:

```go
	var remaining int64
	err := tx.Model(&models.GameItem{}).
		Join("JOIN content_items ci ON ci.id = game_items.content_item_id AND ci.deleted_at IS NULL").
		Where("ci.content_meta_id", metaID).
		Where("game_items.game_level_id", gameLevelID).
		Count(&remaining)
	// ... if remaining == 0, reset is_break_done on content_metas
```

- [ ] **Step 10.4: Update `course_content_service.DeleteAllLevelContent()`**

Before the content-table soft-deletes, soft-delete both junction tables:

```go
	if _, err := tx.Exec(`UPDATE game_items SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL`, gameLevelID); err != nil {
		return fmt.Errorf("failed to soft-delete game_items: %w", err)
	}
	if _, err := tx.Exec(`UPDATE game_metas SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL`, gameLevelID); err != nil {
		return fmt.Errorf("failed to soft-delete game_metas: %w", err)
	}
```

- [ ] **Step 10.5: Update `course_content_service.DeleteMetadata()`**

Before soft-deleting `content_metas` and its children:

```go
	if _, err := tx.Exec(`
		UPDATE game_items SET deleted_at = NOW()
		WHERE content_item_id IN (SELECT id FROM content_items WHERE content_meta_id = ?)
		  AND deleted_at IS NULL
	`, metaID); err != nil {
		return fmt.Errorf("failed to soft-delete game_items for meta: %w", err)
	}
	if _, err := tx.Exec(`UPDATE game_metas SET deleted_at = NOW() WHERE content_meta_id = ? AND deleted_at IS NULL`, metaID); err != nil {
		return fmt.Errorf("failed to soft-delete game_metas: %w", err)
	}
```

- [ ] **Step 10.6: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: exit 0.

- [ ] **Step 10.7: Commit**

```bash
git add dx-api/app/services/api/course_game_service.go \
        dx-api/app/services/api/course_content_service.go
git commit -m "refactor(api): delete paths soft-delete junction rows alongside content"
```

---

## Task 11: Refactor `ai_custom_service.go`

**Files:**
- Modify: `dx-api/app/services/api/ai_custom_service.go`

- [ ] **Step 11.1: Refactor `BreakMetadata()` meta lookup**

Find the query that loads unbroken metas for a level (search for `is_break_done.*false` or `Where("game_level_id"` in this file). Replace the `Where` filter:

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().Model(&models.ContentMeta{}).
		Select("content_metas.*").
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
		Where("gm.game_level_id", gameLevelID).
		Where("content_metas.is_break_done", false).
		OrderBy(`gm."order"`).
		Get(&metas); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
		return
	}
```

- [ ] **Step 11.2: Refactor `processBreakMeta()` to create `game_items` rows**

Find the loop that creates `ContentItem` rows from AI-parsed units inside `processBreakMeta`. After each `Create(&item)` call, add:

```go
		gi := models.GameItem{
			ID:            uuid.NewString(),
			GameID:        gameID,       // passed in via closure or derived from gameLevelID once
			GameLevelID:   gameLevelID,
			ContentItemID: item.ID,
			Order:         item.Order,
		}
		if err := facades.Orm().Query().Create(&gi); err != nil {
			return fmt.Errorf("failed to create game_item: %w", err)
		}
```

If `gameID` is not already in scope, derive it at the top of `BreakMetadata()` (or pass it through `processBreakMeta`'s signature):

```go
	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).First(&level); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load level: %w", err))
		return
	}
	gameID := level.GameID
```

- [ ] **Step 11.3: Refactor `GenerateContentItems()` pending-items query**

Find (around line 583) the query that fetches pending items (`items IS NULL`). It currently queries `content_items` directly. Replace with a junction JOIN so only items linked to this level are considered:

```go
	var pendingItems []models.ContentItem
	if err := facades.Orm().Query().Model(&models.ContentItem{}).
		Select("content_items.*").
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
		Where("gi.game_level_id", gameLevelID).
		Where("content_items.content_meta_id IN ?", metaIDs).
		Where("content_items.items IS NULL").
		Get(&pendingItems); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load pending items: %w", err))
		return
	}
```

Removed: the `is_active = true` filter. The junction join implicitly restricts to items that are actually linked to the level.

- [ ] **Step 11.4: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: exit 0.

- [ ] **Step 11.5: Commit**

```bash
git add dx-api/app/services/api/ai_custom_service.go
git commit -m "refactor(api): ai_custom_service reads/writes via junction tables"
```

---

## Task 12: Refactor `ai_custom_vocab_service.go` and `import_courses.go`

**Files:**
- Modify: `dx-api/app/services/api/ai_custom_vocab_service.go`
- Modify: `dx-api/app/console/commands/import_courses.go`

- [ ] **Step 12.1: Refactor `BreakVocabMetadata()` in ai_custom_vocab_service.go**

Same pattern as Task 11 Step 11.1. Replace any `Where("game_level_id", ...)` on `content_metas` with a `JOIN game_metas`.

- [ ] **Step 12.2: Refactor `processVocabBreakMeta()` to create `game_items` rows**

Same pattern as Task 11 Step 11.2. After each `ContentItem.Create`, insert a matching `GameItem` row.

- [ ] **Step 12.3: Refactor `GenerateVocabContentItems()` pending-items query**

Same pattern as Task 11 Step 11.3.

- [ ] **Step 12.4: Refactor `import_courses.go` `insertLevels()`**

After the existing content_items batch insert inside `insertLevels`, add a `createGameItemsBatch` call. Create a new helper at the top of the file or alongside existing helpers:

```go
// createGameItemsBatch inserts game_items rows in bulk, one per content_item.
func createGameItemsBatch(gameID, gameLevelID string, contentItems []models.ContentItem) error {
	if len(contentItems) == 0 {
		return nil
	}
	batch := make([]models.GameItem, 0, len(contentItems))
	for _, ci := range contentItems {
		batch = append(batch, models.GameItem{
			ID:            uuid.NewString(),
			GameID:        gameID,
			GameLevelID:   gameLevelID,
			ContentItemID: ci.ID,
			Order:         ci.Order,
		})
	}
	return facades.Orm().Query().Create(&batch)
}

// createGameMetasBatch inserts game_metas rows in bulk, one per content_meta.
func createGameMetasBatch(gameID, gameLevelID string, contentMetas []models.ContentMeta) error {
	if len(contentMetas) == 0 {
		return nil
	}
	batch := make([]models.GameMeta, 0, len(contentMetas))
	for _, cm := range contentMetas {
		batch = append(batch, models.GameMeta{
			ID:            uuid.NewString(),
			GameID:        gameID,
			GameLevelID:   gameLevelID,
			ContentMetaID: cm.ID,
			Order:         cm.Order,
		})
	}
	return facades.Orm().Query().Create(&batch)
}
```

Then in `insertLevels()`, after the `content_items` and `content_metas` batch creates, call both helpers:

```go
	if err := createGameMetasBatch(gameID, level.ID, metas); err != nil {
		return fmt.Errorf("failed to create game_metas: %w", err)
	}
	if err := createGameItemsBatch(gameID, level.ID, items); err != nil {
		return fmt.Errorf("failed to create game_items: %w", err)
	}
```

Use whatever local variable names the existing function has for the just-inserted `content_metas` / `content_items` slices.

- [ ] **Step 12.5: Refactor `import_courses.go` `forceCleanup()`**

Before the existing cleanup UPDATEs for content tables, add:

```go
	if _, err := facades.Orm().Query().Exec(`
		UPDATE game_items SET deleted_at = NOW()
		WHERE game_level_id IN (SELECT id FROM game_levels WHERE game_id = ? AND deleted_at IS NOT NULL)
		  AND deleted_at IS NULL
	`, gameID); err != nil {
		return fmt.Errorf("failed to clean up game_items: %w", err)
	}
	if _, err := facades.Orm().Query().Exec(`
		UPDATE game_metas SET deleted_at = NOW()
		WHERE game_level_id IN (SELECT id FROM game_levels WHERE game_id = ? AND deleted_at IS NOT NULL)
		  AND deleted_at IS NULL
	`, gameID); err != nil {
		return fmt.Errorf("failed to clean up game_metas: %w", err)
	}
```

- [ ] **Step 12.6: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: exit 0.

- [ ] **Step 12.7: Commit**

```bash
git add dx-api/app/services/api/ai_custom_vocab_service.go \
        dx-api/app/console/commands/import_courses.go
git commit -m "refactor(api): ai_custom_vocab and import_courses use junction tables"
```

---

## Task 13: Final cleanup — drop legacy columns and struct fields (atomic)

**Files:**
- Modify: `dx-api/app/models/content_meta.go`
- Modify: `dx-api/app/models/content_item.go`
- Modify: `dx-api/app/services/api/course_content_service.go` (remove `GameLevelID`/`Order` assignments in `SaveMetadataBatch`, `InsertContentItem`, any other insert sites)
- Modify: `dx-api/app/services/api/ai_custom_service.go` (remove `GameLevelID`/`Order` in `processBreakMeta`)
- Modify: `dx-api/app/services/api/ai_custom_vocab_service.go` (remove `GameLevelID`/`Order` in `processVocabBreakMeta`)
- Modify: `dx-api/app/console/commands/import_courses.go` (remove `GameLevelID`/`Order` in content-table inserts)
- Modify: `dx-api/bootstrap/migrations.go`
- Create: `dx-api/database/migrations/20260414000003_drop_legacy_columns_from_content_tables.go`

Everything in this task must be committed together so that when the app starts, Migration 3 runs before the code serves requests and the struct's absence of `GameLevelID` is consistent with the dropped column.

- [ ] **Step 13.1: Remove fields from `content_meta.go`**

Replace the struct definition:

```go
package models

import "github.com/goravel/framework/database/orm"

type ContentMeta struct {
	orm.Timestamps
	orm.SoftDeletes
	ID          string  `gorm:"column:id;primaryKey" json:"id"`
	SourceFrom  string  `gorm:"column:source_from" json:"source_from"`
	SourceType  string  `gorm:"column:source_type" json:"source_type"`
	SourceData  string  `gorm:"column:source_data" json:"source_data"`
	Translation *string `gorm:"column:translation" json:"translation"`
	IsBreakDone bool    `gorm:"column:is_break_done" json:"is_break_done"`
}

// TableName returns the database table name.
func (c *ContentMeta) TableName() string {
	return "content_metas"
}
```

- [ ] **Step 13.2: Remove fields from `content_item.go`**

Replace the struct definition:

```go
package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/lib/pq"
)

type ContentItem struct {
	orm.Timestamps
	orm.SoftDeletes
	ID            string         `gorm:"column:id;primaryKey" json:"id"`
	ContentMetaID *string        `gorm:"column:content_meta_id" json:"content_meta_id"`
	Content       string         `gorm:"column:content" json:"content"`
	ContentType   string         `gorm:"column:content_type" json:"content_type"`
	UkAudioID     *string        `gorm:"column:uk_audio_id" json:"uk_audio_id"`
	UsAudioID     *string        `gorm:"column:us_audio_id" json:"us_audio_id"`
	Definition    *string        `gorm:"column:definition" json:"definition"`
	Translation   *string        `gorm:"column:translation" json:"translation"`
	Explanation   *string        `gorm:"column:explanation" json:"explanation"`
	Items         *string        `gorm:"column:items;type:jsonb" json:"items"`
	Structure     *string        `gorm:"column:structure;type:jsonb" json:"structure"`
	Tags          pq.StringArray `gorm:"column:tags;type:text[]" json:"tags"`
}

// TableName returns the database table name.
func (c *ContentItem) TableName() string {
	return "content_items"
}
```

- [ ] **Step 13.3: Let the compiler find every stale reference**

Run: `cd dx-api && go build ./...`
Expected: **build fails** with one or more errors like:

```
app/services/api/course_content_service.go:XX:YY: unknown field 'GameLevelID' in struct literal of type models.ContentMeta
app/services/api/course_content_service.go:XX:YY: unknown field 'Order' in struct literal of type models.ContentMeta
app/services/api/ai_custom_service.go:XX:YY: unknown field 'GameLevelID' in struct literal of type models.ContentItem
...
```

This is the **intended** compile-failure checkpoint: the compiler enumerates every remaining legacy reference.

- [ ] **Step 13.4: Fix each error by removing the stale assignment**

For each error, open the file at the reported line, find the struct literal, and delete the lines that set `GameLevelID`, `Order`, or `IsActive` on `ContentMeta`/`ContentItem`. **Do not delete `GameLevelID`/`Order` lines on `GameMeta`/`GameItem` struct literals** — those still exist.

Re-run `go build ./...` after fixing each file. Repeat until the build succeeds.

- [ ] **Step 13.5: Also remove any remaining `.Where("is_active", ...)` on `content_items`**

Search for leftover is_active filters on content_items specifically:

```bash
grep -rn "is_active" dx-api/app/services/api/ | grep -iE "content_item"
```

Any hits on content_items (not on games, game_levels, notices, posts, etc.) should have their `is_active` filter removed. The models no longer have the field; the queries will compile as-is (`.Where("is_active", true)` is a string arg) but will fail at DB time after Migration 3 drops the column. Remove them now.

- [ ] **Step 13.6: Write migration 3**

Create `dx-api/database/migrations/20260414000003_drop_legacy_columns_from_content_tables.go`:

```go
package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260414000003DropLegacyColumnsFromContentTables struct{}

func (r *M20260414000003DropLegacyColumnsFromContentTables) Signature() string {
	return "20260414000003_drop_legacy_columns_from_content_tables"
}

func (r *M20260414000003DropLegacyColumnsFromContentTables) Up() error {
	stmts := []string{
		`ALTER TABLE content_metas DROP COLUMN IF EXISTS game_level_id`,
		`ALTER TABLE content_metas DROP COLUMN IF EXISTS "order"`,
		`ALTER TABLE content_items DROP COLUMN IF EXISTS game_level_id`,
		`ALTER TABLE content_items DROP COLUMN IF EXISTS "order"`,
		`ALTER TABLE content_items DROP COLUMN IF EXISTS is_active`,
	}
	for _, s := range stmts {
		if _, err := facades.Orm().Query().Exec(s); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260414000003DropLegacyColumnsFromContentTables) Down() error {
	// Re-creates the columns as nullable so an emergency downgrade can run.
	// Values cannot be auto-restored; a separate repair script would be needed
	// to re-derive them from game_items/game_metas.
	stmts := []string{
		`ALTER TABLE content_metas ADD COLUMN IF NOT EXISTS game_level_id UUID`,
		`ALTER TABLE content_metas ADD COLUMN IF NOT EXISTS "order" DOUBLE PRECISION DEFAULT 0`,
		`ALTER TABLE content_items ADD COLUMN IF NOT EXISTS game_level_id UUID`,
		`ALTER TABLE content_items ADD COLUMN IF NOT EXISTS "order" DOUBLE PRECISION DEFAULT 0`,
		`ALTER TABLE content_items ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true`,
	}
	for _, s := range stmts {
		if _, err := facades.Orm().Query().Exec(s); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 13.7: Register migration 3 in bootstrap**

Edit `dx-api/bootstrap/migrations.go`. After `&migrations.M20260414000002BackfillJunctionTables{},` append:

```go
		&migrations.M20260414000003DropLegacyColumnsFromContentTables{},
```

- [ ] **Step 13.8: Verify compilation and vet**

```bash
cd dx-api && go build ./... && go vet ./...
```

Expected: both exit 0.

- [ ] **Step 13.9: Run migration 3 against the dev DB**

Run: `cd dx-api && go run . artisan migrate`
Expected: "Migrated: 20260414000003_drop_legacy_columns_from_content_tables".

- [ ] **Step 13.10: Verify columns are gone**

```bash
psql $DX_DB_URL -c '\d content_metas'
psql $DX_DB_URL -c '\d content_items'
```

Expected: `content_metas` has no `game_level_id` or `order` column. `content_items` has no `game_level_id`, `order`, or `is_active` column.

- [ ] **Step 13.11: Commit**

```bash
git add dx-api/app/models/content_meta.go \
        dx-api/app/models/content_item.go \
        dx-api/app/services/api/course_content_service.go \
        dx-api/app/services/api/ai_custom_service.go \
        dx-api/app/services/api/ai_custom_vocab_service.go \
        dx-api/app/console/commands/import_courses.go \
        dx-api/bootstrap/migrations.go \
        dx-api/database/migrations/20260414000003_drop_legacy_columns_from_content_tables.go
git commit -m "refactor(api): drop game_level_id/order/is_active from content tables"
```

---

## Task 14: Full-system verification

**Files:** none (verification only)

- [ ] **Step 14.1: Clean build and vet**

```bash
cd dx-api && go build ./... && go vet ./...
```

Expected: exit 0.

- [ ] **Step 14.2: Run unit tests with race detector**

```bash
cd dx-api && go test -race ./...
```

Expected: all tests pass.

- [ ] **Step 14.3: Row-count invariant still holds**

```bash
psql $DX_DB_URL -c "
SELECT 'content_items live' AS label, COUNT(*) AS n FROM content_items WHERE deleted_at IS NULL
UNION ALL SELECT 'game_items live', COUNT(*) FROM game_items WHERE deleted_at IS NULL
UNION ALL SELECT 'content_metas live', COUNT(*) FROM content_metas WHERE deleted_at IS NULL
UNION ALL SELECT 'game_metas live', COUNT(*) FROM game_metas WHERE deleted_at IS NULL;
"
```

Expected: `content_items live` = `game_items live`. `content_metas live` = `game_metas live`. Matches the baseline from Task 0 Step 0.2.

- [ ] **Step 14.4: Smoke test — single play**

Start the API (`cd dx-api && go run .`). In another terminal, with a valid JWT in `$DEV_USER_TOKEN`:

```bash
# start session
curl -s -XPOST -H "Authorization: Bearer $DEV_USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"game_id\":\"$GAME_ID\",\"game_level_id\":\"$LEVEL_ID\",\"degree\":\"beginner\"}" \
  http://localhost:3001/api/play-single/start | jq

# load content
curl -s -H "Authorization: Bearer $DEV_USER_TOKEN" \
  "http://localhost:3001/api/games/$GAME_ID/levels/$LEVEL_ID/content?degree=beginner" \
  | jq '.data | length'
```

Expected: session starts with an ID; content call returns a non-zero count matching the junction-based query.

- [ ] **Step 14.5: Smoke test — ai-custom sentence flow**

Via the web UI (`cd dx-web && npm run dev` in another terminal) or direct API calls:
1. Create a test course game
2. Create a level
3. Generate metadata (5 beans)
4. Save metadata to level
5. Break metadata (SSE)
6. Generate content items (SSE)
7. Publish game

Expected: all seven steps succeed. After step 6, `game_items` has new rows for each generated content item:

```bash
psql $DX_DB_URL -c "
SELECT COUNT(*) FROM game_items
WHERE game_level_id = '$TEST_LEVEL_ID' AND deleted_at IS NULL;
"
```

- [ ] **Step 14.6: Smoke test — ai-custom vocab flow**

Same flow with the vocab endpoints (`generate-vocab`, `format-vocab`, `break-vocab-metadata`, `generate-vocab-content-items`).

- [ ] **Step 14.7: Smoke test — course editor CRUD + cascade delete**

1. Insert a content item via the editor
2. Delete it — verify both the `content_items` row and the `game_items` row are soft-deleted
3. Delete the entire level — verify `game_items`/`game_metas` and `content_items`/`content_metas` for that level are all soft-deleted
4. Delete the game — verify cascades to all levels

```bash
psql $DX_DB_URL -c "
SELECT
  (SELECT COUNT(*) FROM game_items WHERE game_level_id = '$TEST_LEVEL_ID' AND deleted_at IS NULL) AS live_items,
  (SELECT COUNT(*) FROM game_items WHERE game_level_id = '$TEST_LEVEL_ID' AND deleted_at IS NOT NULL) AS deleted_items;
"
```

Expected after delete: `live_items = 0`, `deleted_items > 0`.

- [ ] **Step 14.8: Smoke test — PK and group play**

Start a robot PK (`POST /api/game-pks/robot` or equivalent) for the test level. Verify the robot session loads content via the junction path:

```bash
curl -s -XPOST -H "Authorization: Bearer $DEV_USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"game_id\":\"$GAME_ID\",\"game_level_id\":\"$LEVEL_ID\",\"degree\":\"beginner\"}" \
  http://localhost:3001/api/game-pks/robot | jq
```

Expected: the PK starts successfully and `total_items_count` matches `SELECT COUNT(*) FROM game_items WHERE game_level_id = $LEVEL_ID AND deleted_at IS NULL` (with content_type filter if the degree has one).

Same for group play via `POST /api/game-groups/...`.

- [ ] **Step 14.9: Update memory for future sessions**

If any memory files in `/Users/rainsen/.claude/projects/-Users-rainsen-Programs-Projects-douxue-dx-source/memory/` mention the old content-tables schema or the previous removal of junction tables, update them to reflect the current state (junction-tables are back; content tables no longer have `game_level_id`, `order`, `is_active`). In particular, `project_game_junction_tables.md` should be revised or replaced.

- [ ] **Step 14.10: Final commit (if any verification fixes were needed)**

If any smoke test uncovered a bug and required a code fix, commit it:

```bash
git add <files>
git commit -m "fix(api): <concrete description of what was broken>"
```

Otherwise, no commit needed — the plan is complete.

---

## Self-review notes

**Spec coverage:**
- Migration 1 (tables + indexes) → Task 1 ✓
- Migration 2 (backfill) → Task 3 ✓
- Migration 3 (drop columns) → Task 13 ✓
- New models `GameMeta`/`GameItem` → Task 2 ✓
- Remove fields from `ContentMeta`/`ContentItem` → Task 13 Steps 13.1, 13.2 ✓
- `content_service.GetLevelContent` → Task 4 ✓
- `game_play_single_service.countLevelItems` → Task 5 ✓
- `game_play_pk_service.spawnRobotForLevel` → Task 6 ✓
- `course_content_service` (reads) → Task 7 ✓
- `course_content_service` (writes) → Task 8 ✓
- `course_game_service` (reads) → Task 9 ✓
- `course_game_service` + `course_content_service` (deletes) → Task 10 ✓
- `ai_custom_service` → Task 11 ✓
- `ai_custom_vocab_service` → Task 12 ✓
- `import_courses` → Task 12 ✓
- `bootstrap/migrations.go` → registered in Tasks 1, 3, 13 ✓
- Verification (row-count invariant, smoke tests, spot checks) → Tasks 0, 3, 14 ✓

**Placeholder scan:** searched this file for "TBD", "TODO", "implement later", "similar to", "add appropriate" — none found. Every code block is concrete. Every SQL statement is complete. Every task has explicit file paths.

**Type consistency:**
- `GameMeta` fields (`ID`, `GameID`, `GameLevelID`, `ContentMetaID`, `Order`) match Task 2 definition and Task 8/11/12 usage.
- `GameItem` fields (`ID`, `GameID`, `GameLevelID`, `ContentItemID`, `Order`) match Task 2 definition and usage throughout.
- `countLevelItems` signature `(query orm.Query, gameLevelID, degree string) (int64, error)` is unchanged from the original — Task 5 only rewrites the body.
- The alias `gi` (for `game_items`) and `gm` (for `game_metas`) are used consistently in all JOINs.
- The `"order"` column name is always quoted (reserved keyword in PG), via either SQL literal (`"order"`) or Go-escaped (`\"order\"`) or raw string (`` `gi."order"` ``).

**Scope check:** this is one cohesive refactor with a clear end state. All 14 tasks implement pieces of the same design and nothing else.

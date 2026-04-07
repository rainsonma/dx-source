# Soft Delete Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Goravel-native soft delete to 6 content authoring tables so deleting games/levels/content no longer orphans user progress data.

**Architecture:** Embed `orm.SoftDeletes` in 6 models, add `SoftDeletesTz()` to their existing migrations, update 4 delete functions to use soft-delete-aware orphan cleanup, add `AND deleted_at IS NULL` to 14 raw JOINs, and add `WithTrashed()` to 3 enrichment queries.

**Tech Stack:** Go 1.26, Goravel ORM (orm.SoftDeletes, WithTrashed, ForceDelete), PostgreSQL 18

**Spec:** `docs/superpowers/specs/2026-04-07-soft-delete-design.md`

**Branch:** `feat/game-junction-tables` (continuing existing work)

---

## File Map

| Action | File | Purpose |
|--------|------|---------|
| Modify | `dx-api/app/models/game.go` | Add `orm.SoftDeletes` |
| Modify | `dx-api/app/models/game_level.go` | Add `orm.SoftDeletes` |
| Modify | `dx-api/app/models/content_meta.go` | Add `orm.SoftDeletes` |
| Modify | `dx-api/app/models/content_item.go` | Add `orm.SoftDeletes` |
| Modify | `dx-api/app/models/game_meta.go` | Add `orm.SoftDeletes` + goravel orm import |
| Modify | `dx-api/app/models/game_item.go` | Add `orm.SoftDeletes` + goravel orm import |
| Modify | `dx-api/database/migrations/20260322000016_create_games_table.go:31` | Add `SoftDeletesTz()` |
| Modify | `dx-api/database/migrations/20260322000027_create_game_levels_table.go:27` | Add `SoftDeletesTz()` |
| Modify | `dx-api/database/migrations/20260322000036_create_content_metas_table.go:26` | Add `SoftDeletesTz()` |
| Modify | `dx-api/database/migrations/20260322000037_create_content_items_table.go:33` | Add `SoftDeletesTz()` |
| Modify | `dx-api/database/migrations/20260407000001_create_game_junction_tables.go:23,40` | Add `SoftDeletesTz()` |
| Modify | `dx-api/app/services/api/course_content_service.go:77,201,215,309,416-500,531,551` | Delete funcs + JOIN filters |
| Modify | `dx-api/app/services/api/course_game_service.go:235-274,308,320,412-452,516` | Delete funcs + JOIN filters |
| Modify | `dx-api/app/services/api/content_service.go:41` | JOIN filter |
| Modify | `dx-api/app/services/api/game_play_single_service.go:606` | JOIN filter |
| Modify | `dx-api/app/services/api/game_play_pk_service.go:595` | JOIN filter |
| Modify | `dx-api/app/services/api/ai_custom_service.go:326,572` | JOIN filters |
| Modify | `dx-api/app/services/api/user_master_service.go:105,123` | WithTrashed() |
| Modify | `dx-api/app/services/api/favorite_service.go:76` | WithTrashed() |

---

### Task 1: Add soft delete to models

**Files:**
- Modify: `dx-api/app/models/game.go:6`
- Modify: `dx-api/app/models/game_level.go:9`
- Modify: `dx-api/app/models/content_meta.go:6`
- Modify: `dx-api/app/models/content_item.go:9`
- Modify: `dx-api/app/models/game_meta.go:3-5`
- Modify: `dx-api/app/models/game_item.go:3-5`

- [ ] **Step 1: Add `orm.SoftDeletes` to Game model**

In `dx-api/app/models/game.go`, add `orm.SoftDeletes` after `orm.Timestamps`:

```go
type Game struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string  `gorm:"column:id;primaryKey" json:"id"`
```

- [ ] **Step 2: Add `orm.SoftDeletes` to GameLevel model**

In `dx-api/app/models/game_level.go`, add `orm.SoftDeletes` after `orm.Timestamps`:

```go
type GameLevel struct {
	orm.Timestamps
	orm.SoftDeletes
	ID           string         `gorm:"column:id;primaryKey" json:"id"`
```

- [ ] **Step 3: Add `orm.SoftDeletes` to ContentMeta model**

In `dx-api/app/models/content_meta.go`, add `orm.SoftDeletes` after `orm.Timestamps`:

```go
type ContentMeta struct {
	orm.Timestamps
	orm.SoftDeletes
	ID         string `gorm:"column:id;primaryKey" json:"id"`
```

- [ ] **Step 4: Add `orm.SoftDeletes` to ContentItem model**

In `dx-api/app/models/content_item.go`, add `orm.SoftDeletes` after `orm.Timestamps`:

```go
type ContentItem struct {
	orm.Timestamps
	orm.SoftDeletes
	ID            string         `gorm:"column:id;primaryKey" json:"id"`
```

- [ ] **Step 5: Add `orm.SoftDeletes` to GameMeta model**

In `dx-api/app/models/game_meta.go`, replace the import and add `orm.SoftDeletes`:

```go
import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type GameMeta struct {
	orm.SoftDeletes
	ID            string    `gorm:"column:id;primaryKey" json:"id"`
```

- [ ] **Step 6: Add `orm.SoftDeletes` to GameItem model**

In `dx-api/app/models/game_item.go`, replace the import and add `orm.SoftDeletes`:

```go
import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type GameItem struct {
	orm.SoftDeletes
	ID            string    `gorm:"column:id;primaryKey" json:"id"`
```

- [ ] **Step 7: Verify compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: clean build (orm.SoftDeletes just adds a DeletedAt field — no breaking changes).

- [ ] **Step 8: Commit**

```bash
git add dx-api/app/models/game.go dx-api/app/models/game_level.go dx-api/app/models/content_meta.go dx-api/app/models/content_item.go dx-api/app/models/game_meta.go dx-api/app/models/game_item.go
git commit -m "feat: add orm.SoftDeletes to 6 content authoring models"
```

---

### Task 2: Add SoftDeletesTz() to existing migrations

**Files:**
- Modify: `dx-api/database/migrations/20260322000016_create_games_table.go:31`
- Modify: `dx-api/database/migrations/20260322000027_create_game_levels_table.go:27`
- Modify: `dx-api/database/migrations/20260322000036_create_content_metas_table.go:26`
- Modify: `dx-api/database/migrations/20260322000037_create_content_items_table.go:33`
- Modify: `dx-api/database/migrations/20260407000001_create_game_junction_tables.go:23,40`

- [ ] **Step 1: Add SoftDeletesTz() to games migration**

In `dx-api/database/migrations/20260322000016_create_games_table.go`, add after `table.TimestampsTz()` (line 31):

```go
			table.TimestampsTz()
			table.SoftDeletesTz()
```

- [ ] **Step 2: Add SoftDeletesTz() to game_levels migration**

In `dx-api/database/migrations/20260322000027_create_game_levels_table.go`, add after `table.TimestampsTz()` (line 27):

```go
			table.TimestampsTz()
			table.SoftDeletesTz()
```

- [ ] **Step 3: Add SoftDeletesTz() to content_metas migration**

In `dx-api/database/migrations/20260322000036_create_content_metas_table.go`, add after `table.TimestampsTz()` (line 26):

```go
			table.TimestampsTz()
			table.SoftDeletesTz()
```

- [ ] **Step 4: Add SoftDeletesTz() to content_items migration**

In `dx-api/database/migrations/20260322000037_create_content_items_table.go`, add after `table.TimestampsTz()` (line 33):

```go
			table.TimestampsTz()
			table.SoftDeletesTz()
```

- [ ] **Step 5: Add SoftDeletesTz() to game_metas in junction migration**

In `dx-api/database/migrations/20260407000001_create_game_junction_tables.go`, add after the `TimestampTz("created_at")` line (line 23) for game_metas:

```go
			table.TimestampTz("created_at").UseCurrent()
			table.SoftDeletesTz()
```

- [ ] **Step 6: Add SoftDeletesTz() to game_items in junction migration**

In the same file, add after the `TimestampTz("created_at")` line (line 40) for game_items:

```go
			table.TimestampTz("created_at").UseCurrent()
			table.SoftDeletesTz()
```

- [ ] **Step 7: Verify compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: clean build.

- [ ] **Step 8: Commit**

```bash
git add dx-api/database/migrations/20260322000016_create_games_table.go dx-api/database/migrations/20260322000027_create_game_levels_table.go dx-api/database/migrations/20260322000036_create_content_metas_table.go dx-api/database/migrations/20260322000037_create_content_items_table.go dx-api/database/migrations/20260407000001_create_game_junction_tables.go
git commit -m "feat: add SoftDeletesTz() to 6 content authoring table migrations"
```

---

### Task 3: Add AND deleted_at IS NULL to all 14 JOIN clauses

**Files:**
- Modify: `dx-api/app/services/api/content_service.go:41`
- Modify: `dx-api/app/services/api/course_game_service.go:308,320,516`
- Modify: `dx-api/app/services/api/course_content_service.go:77,201,215,309,531,551`
- Modify: `dx-api/app/services/api/game_play_single_service.go:606`
- Modify: `dx-api/app/services/api/game_play_pk_service.go:595`
- Modify: `dx-api/app/services/api/ai_custom_service.go:326,572`

Every raw JOIN to `game_items` or `game_metas` needs the soft-delete filter appended to the ON clause.

- [ ] **Step 1: content_service.go — GetLevelContent() (line 41)**

```go
// Before
Join("JOIN game_items gi ON gi.content_item_id = content_items.id").

// After
Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
```

- [ ] **Step 2: course_game_service.go — PublishGame() item count (line 308)**

```go
// Before
Join("JOIN game_items gi ON gi.content_item_id = content_items.id").

// After
Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
```

- [ ] **Step 3: course_game_service.go — PublishGame() ungenerated count (line 320)**

```go
// Before
Join("JOIN game_items gi ON gi.content_item_id = content_items.id").

// After
Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
```

- [ ] **Step 4: course_game_service.go — GetCourseGameDetail() level item count (line 516)**

```go
// Before
Join("JOIN game_items gi ON gi.content_item_id = content_items.id").

// After
Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
```

- [ ] **Step 5: course_content_service.go — SaveMetadataBatch() capacity check (line 77)**

```go
// Before
Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id").

// After
Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
```

- [ ] **Step 6: course_content_service.go — GetContentItemsByMeta() metas query (line 201)**

```go
// Before
Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id").

// After
Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
```

- [ ] **Step 7: course_content_service.go — GetContentItemsByMeta() items query (line 215)**

```go
// Before
Join("JOIN game_items gi ON gi.content_item_id = content_items.id").

// After
Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
```

- [ ] **Step 8: course_content_service.go — InsertContentItem() item count (line 309)**

```go
// Before
Join("JOIN game_items gi ON gi.content_item_id = content_items.id").

// After
Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
```

- [ ] **Step 9: course_content_service.go — calculateInsertionOrder() last item (line 531)**

```go
// Before
Join("JOIN game_items gi ON gi.content_item_id = content_items.id").

// After
Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
```

- [ ] **Step 10: course_content_service.go — calculateInsertionOrder() all items (line 551)**

```go
// Before
Join("JOIN game_items gi ON gi.content_item_id = content_items.id").

// After
Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
```

- [ ] **Step 11: game_play_single_service.go — countLevelItems() (line 606)**

```go
// Before
Join("JOIN game_items gi ON gi.content_item_id = content_items.id").

// After
Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
```

- [ ] **Step 12: game_play_pk_service.go — spawnRobotForLevel() (line 595)**

```go
// Before
Join("JOIN game_items gi ON gi.content_item_id = content_items.id").

// After
Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
```

- [ ] **Step 13: ai_custom_service.go — BreakMetadata() (line 326)**

```go
// Before
Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id").

// After
Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
```

- [ ] **Step 14: ai_custom_service.go — GenerateContentItems() (line 572)**

```go
// Before
Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id").

// After
Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
```

- [ ] **Step 15: Verify compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: clean build.

- [ ] **Step 16: Run go vet**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./...
```

Expected: clean.

- [ ] **Step 17: Commit**

```bash
git add dx-api/app/services/api/content_service.go dx-api/app/services/api/course_game_service.go dx-api/app/services/api/course_content_service.go dx-api/app/services/api/game_play_single_service.go dx-api/app/services/api/game_play_pk_service.go dx-api/app/services/api/ai_custom_service.go
git commit -m "fix: add deleted_at IS NULL filter to all 14 junction table JOINs"
```

---

### Task 4: Update delete functions to use soft delete

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go:416-500`
- Modify: `dx-api/app/services/api/course_game_service.go:235-274,435-452`

- [ ] **Step 1: Wrap DeleteContentItem in a transaction**

In `dx-api/app/services/api/course_content_service.go`, replace the body of `DeleteContentItem` (lines 416-451). Keep the guards unchanged, replace everything from the junction delete onward:

```go
// DeleteContentItem removes a single content item.
func DeleteContentItem(userID, gameID, itemID string) error {
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

	return facades.Orm().Transaction(func(tx orm.Query) error {
		// Soft-delete junction row
		if _, err := tx.
			Where("content_item_id", itemID).Where("game_id", gameID).
			Delete(&models.GameItem{}); err != nil {
			return fmt.Errorf("failed to delete game item: %w", err)
		}

		// Count active references (auto-excludes soft-deleted)
		remaining, _ := tx.Model(&models.GameItem{}).
			Where("content_item_id", itemID).Count()
		if remaining == 0 {
			if _, err := tx.Where("id", itemID).Delete(&models.ContentItem{}); err != nil {
				return fmt.Errorf("failed to delete content item: %w", err)
			}
		}
		return nil
	})
}
```

- [ ] **Step 2: Update DeleteAllLevelContent orphan SQL**

In `dx-api/app/services/api/course_content_service.go`, replace the two raw SQL Exec calls inside the transaction in `DeleteAllLevelContent` (lines 487-496):

```go
		// Soft-delete orphaned content (not referenced by any active game)
		if _, err := tx.Exec(
			"UPDATE content_items SET deleted_at = NOW() WHERE deleted_at IS NULL AND id NOT IN (SELECT content_item_id FROM game_items WHERE deleted_at IS NULL)",
		); err != nil {
			return fmt.Errorf("failed to delete orphaned content items: %w", err)
		}
		if _, err := tx.Exec(
			"UPDATE content_metas SET deleted_at = NOW() WHERE deleted_at IS NULL AND id NOT IN (SELECT content_meta_id FROM game_metas WHERE deleted_at IS NULL)",
		); err != nil {
			return fmt.Errorf("failed to delete orphaned content metas: %w", err)
		}
```

- [ ] **Step 3: Update DeleteLevel orphan SQL**

In `dx-api/app/services/api/course_game_service.go`, replace the two raw SQL Exec calls inside the transaction in `DeleteLevel` (lines 442-447):

```go
		if _, err := tx.Exec("UPDATE content_items SET deleted_at = NOW() WHERE deleted_at IS NULL AND id NOT IN (SELECT content_item_id FROM game_items WHERE deleted_at IS NULL)"); err != nil {
			return fmt.Errorf("failed to delete orphaned content items: %w", err)
		}
		if _, err := tx.Exec("UPDATE content_metas SET deleted_at = NOW() WHERE deleted_at IS NULL AND id NOT IN (SELECT content_meta_id FROM game_metas WHERE deleted_at IS NULL)"); err != nil {
			return fmt.Errorf("failed to delete orphaned content metas: %w", err)
		}
```

- [ ] **Step 4: Update DeleteGame orphan SQL**

In `dx-api/app/services/api/course_game_service.go`, replace the two raw SQL Exec calls inside the transaction in `DeleteGame` (lines 255-260):

```go
			if _, err := tx.Exec("UPDATE content_items SET deleted_at = NOW() WHERE deleted_at IS NULL AND id NOT IN (SELECT content_item_id FROM game_items WHERE deleted_at IS NULL)"); err != nil {
				return fmt.Errorf("failed to delete orphaned content items: %w", err)
			}
			if _, err := tx.Exec("UPDATE content_metas SET deleted_at = NOW() WHERE deleted_at IS NULL AND id NOT IN (SELECT content_meta_id FROM game_metas WHERE deleted_at IS NULL)"); err != nil {
				return fmt.Errorf("failed to delete orphaned content metas: %w", err)
			}
```

- [ ] **Step 5: Verify compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: clean build.

- [ ] **Step 6: Run go vet**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./...
```

Expected: clean.

- [ ] **Step 7: Commit**

```bash
git add dx-api/app/services/api/course_content_service.go dx-api/app/services/api/course_game_service.go
git commit -m "refactor: delete functions use soft delete with orphan cleanup"
```

---

### Task 5: Add WithTrashed() to enrichment queries

**Files:**
- Modify: `dx-api/app/services/api/user_master_service.go:105,123`
- Modify: `dx-api/app/services/api/favorite_service.go:76`

- [ ] **Step 1: Add WithTrashed() to batchLoadContentItems (line 105)**

In `dx-api/app/services/api/user_master_service.go`, change line 105:

```go
// Before
facades.Orm().Query().Where("id IN ?", ids).Get(&items)

// After
facades.Orm().Query().WithTrashed().Where("id IN ?", ids).Get(&items)
```

- [ ] **Step 2: Add WithTrashed() to batchLoadGameNames (line 123)**

In the same file, change line 123:

```go
// Before
facades.Orm().Query().Where("id IN ?", ids).Get(&games)

// After
facades.Orm().Query().WithTrashed().Where("id IN ?", ids).Get(&games)
```

- [ ] **Step 3: Add WithTrashed() to ListFavorites game load (line 76)**

In `dx-api/app/services/api/favorite_service.go`, change line 76:

```go
// Before
facades.Orm().Query().Where("id IN ?", gameIDs).Get(&games)

// After
facades.Orm().Query().WithTrashed().Where("id IN ?", gameIDs).Get(&games)
```

- [ ] **Step 4: Verify compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: clean build.

- [ ] **Step 5: Run go vet**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./...
```

Expected: clean.

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/services/api/user_master_service.go dx-api/app/services/api/favorite_service.go
git commit -m "fix: add WithTrashed() to enrichment queries for soft-deleted content"
```

---

### Task 6: Final verification

- [ ] **Step 1: Full build**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: clean.

- [ ] **Step 2: go vet**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./...
```

Expected: clean.

- [ ] **Step 3: Run tests**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./...
```

Expected: all pass (or no test failures).

- [ ] **Step 4: Review all changes**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git diff main...HEAD --stat
```

Verify changed files match the file map.

- [ ] **Step 5: Verify no untracked files**

```bash
git status
```

Expected: clean working tree on `feat/game-junction-tables` branch.

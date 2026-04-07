# Game Junction Tables Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `game_metas` and `game_items` junction tables to decouple content from games, enabling future cross-game content sharing.

**Architecture:** Two new junction tables bridge games/levels to content_metas/content_items. All content queries switch from direct `game_level_id` filtering to JOINing through junction tables. Existing data is migrated. New writes no longer set `game_level_id` on content tables.

**Tech Stack:** Go 1.26, Goravel ORM, PostgreSQL 18, GORM

**Spec:** `docs/superpowers/specs/2026-04-07-game-junction-tables-design.md`

**Branch:** `feat/game-junction-tables`

---

## File Map

| Action | File | Purpose |
|--------|------|---------|
| Create | `dx-api/database/migrations/20260407000001_create_game_junction_tables.go` | Migration: create tables, backfill data, alter columns |
| Create | `dx-api/app/models/game_meta.go` | GameMeta model |
| Create | `dx-api/app/models/game_item.go` | GameItem model |
| Modify | `dx-api/app/models/content_meta.go` | GameLevelID becomes `*string` |
| Modify | `dx-api/app/models/content_item.go` | GameLevelID becomes `*string` |
| Modify | `dx-api/bootstrap/migrations.go:59` | Register new migration |
| Modify | `dx-api/app/services/api/content_service.go:27-104` | GetLevelContent — JOIN through game_items |
| Modify | `dx-api/app/services/api/game_play_single_service.go:604-612` | countLevelItems — JOIN through game_items |
| Modify | `dx-api/app/services/api/course_content_service.go:51-543` | All CRUD — use junction tables |
| Modify | `dx-api/app/services/api/ai_custom_service.go:306-473,542-746` | BreakMetadata/GenerateContentItems — thread gameID, write junction rows |

---

### Task 1: Create feature branch

- [ ] **Step 1: Create and switch to feature branch**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git checkout -b feat/game-junction-tables
```

- [ ] **Step 2: Verify branch**

```bash
git branch --show-current
```

Expected: `feat/game-junction-tables`

---

### Task 2: Create migration file

**Files:**
- Create: `dx-api/database/migrations/20260407000001_create_game_junction_tables.go`
- Modify: `dx-api/bootstrap/migrations.go:59`

- [ ] **Step 1: Create the migration file**

```go
// dx-api/database/migrations/20260407000001_create_game_junction_tables.go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260407000001CreateGameJunctionTables struct{}

func (r *M20260407000001CreateGameJunctionTables) Signature() string {
	return "20260407000001_create_game_junction_tables"
}

func (r *M20260407000001CreateGameJunctionTables) Up() error {
	// 1. Create game_metas table
	if !facades.Schema().HasTable("game_metas") {
		if err := facades.Schema().Create("game_metas", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Uuid("content_meta_id")
			table.TimestampTz("created_at").UseCurrent()
			table.Index("game_level_id")
			table.Index("content_meta_id")
		}); err != nil {
			return err
		}
		// Composite unique index
		if _, err := facades.Orm().Query().Exec(
			"CREATE UNIQUE INDEX idx_game_metas_unique ON game_metas (game_id, game_level_id, content_meta_id)",
		); err != nil {
			return err
		}
	}

	// 2. Create game_items table
	if !facades.Schema().HasTable("game_items") {
		if err := facades.Schema().Create("game_items", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Uuid("content_item_id")
			table.TimestampTz("created_at").UseCurrent()
			table.Index("game_level_id")
			table.Index("content_item_id")
		}); err != nil {
			return err
		}
		if _, err := facades.Orm().Query().Exec(
			"CREATE UNIQUE INDEX idx_game_items_unique ON game_items (game_id, game_level_id, content_item_id)",
		); err != nil {
			return err
		}
	}

	// 3. Backfill game_metas from existing content_metas
	if _, err := facades.Orm().Query().Exec(`
		INSERT INTO game_metas (id, game_id, game_level_id, content_meta_id, created_at)
		SELECT gen_random_uuid(), gl.game_id, cm.game_level_id, cm.id, cm.created_at
		FROM content_metas cm
		JOIN game_levels gl ON gl.id = cm.game_level_id
		WHERE cm.game_level_id IS NOT NULL
		ON CONFLICT DO NOTHING
	`); err != nil {
		return err
	}

	// 4. Backfill game_items from existing content_items
	if _, err := facades.Orm().Query().Exec(`
		INSERT INTO game_items (id, game_id, game_level_id, content_item_id, created_at)
		SELECT gen_random_uuid(), gl.game_id, ci.game_level_id, ci.id, ci.created_at
		FROM content_items ci
		JOIN game_levels gl ON gl.id = ci.game_level_id
		WHERE ci.game_level_id IS NOT NULL
		ON CONFLICT DO NOTHING
	`); err != nil {
		return err
	}

	// 5. Make game_level_id nullable on content tables
	if _, err := facades.Orm().Query().Exec(
		"ALTER TABLE content_metas ALTER COLUMN game_level_id DROP NOT NULL",
	); err != nil {
		return err
	}
	if _, err := facades.Orm().Query().Exec(
		"ALTER TABLE content_items ALTER COLUMN game_level_id DROP NOT NULL",
	); err != nil {
		return err
	}

	return nil
}

func (r *M20260407000001CreateGameJunctionTables) Down() error {
	// Restore NOT NULL (only safe if no NULLs exist yet)
	_, _ = facades.Orm().Query().Exec(
		"ALTER TABLE content_metas ALTER COLUMN game_level_id SET NOT NULL",
	)
	_, _ = facades.Orm().Query().Exec(
		"ALTER TABLE content_items ALTER COLUMN game_level_id SET NOT NULL",
	)
	_ = facades.Schema().DropIfExists("game_items")
	_ = facades.Schema().DropIfExists("game_metas")
	return nil
}
```

- [ ] **Step 2: Register migration in bootstrap/migrations.go**

Add to the end of the slice in `dx-api/bootstrap/migrations.go`, before the closing `}` at line 60:

```go
		&migrations.M20260407000001CreateGameJunctionTables{},
```

- [ ] **Step 3: Verify compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add dx-api/database/migrations/20260407000001_create_game_junction_tables.go dx-api/bootstrap/migrations.go
git commit -m "feat: add game_metas and game_items junction tables migration"
```

---

### Task 3: Create new models + update existing models

**Files:**
- Create: `dx-api/app/models/game_meta.go`
- Create: `dx-api/app/models/game_item.go`
- Modify: `dx-api/app/models/content_meta.go:8`
- Modify: `dx-api/app/models/content_item.go:11`

- [ ] **Step 1: Create GameMeta model**

```go
// dx-api/app/models/game_meta.go
package models

import "time"

type GameMeta struct {
	ID            string    `gorm:"column:id;primaryKey" json:"id"`
	GameID        string    `gorm:"column:game_id" json:"game_id"`
	GameLevelID   string    `gorm:"column:game_level_id" json:"game_level_id"`
	ContentMetaID string    `gorm:"column:content_meta_id" json:"content_meta_id"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

func (g *GameMeta) TableName() string {
	return "game_metas"
}
```

- [ ] **Step 2: Create GameItem model**

```go
// dx-api/app/models/game_item.go
package models

import "time"

type GameItem struct {
	ID            string    `gorm:"column:id;primaryKey" json:"id"`
	GameID        string    `gorm:"column:game_id" json:"game_id"`
	GameLevelID   string    `gorm:"column:game_level_id" json:"game_level_id"`
	ContentItemID string    `gorm:"column:content_item_id" json:"content_item_id"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

func (g *GameItem) TableName() string {
	return "game_items"
}
```

- [ ] **Step 3: Update ContentMeta.GameLevelID to pointer**

In `dx-api/app/models/content_meta.go`, change line 8 from:

```go
GameLevelID string  `gorm:"column:game_level_id" json:"game_level_id"`
```

to:

```go
GameLevelID *string `gorm:"column:game_level_id" json:"game_level_id"`
```

- [ ] **Step 4: Update ContentItem.GameLevelID to pointer**

In `dx-api/app/models/content_item.go`, change line 11 from:

```go
GameLevelID   string         `gorm:"column:game_level_id" json:"game_level_id"`
```

to:

```go
GameLevelID   *string        `gorm:"column:game_level_id" json:"game_level_id"`
```

- [ ] **Step 5: Verify compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: compile errors in services that use `GameLevelID` as a string — this is expected and will be fixed in subsequent tasks.

- [ ] **Step 6: Commit models only (fix compile errors in later tasks)**

```bash
git add dx-api/app/models/game_meta.go dx-api/app/models/game_item.go dx-api/app/models/content_meta.go dx-api/app/models/content_item.go
git commit -m "feat: add GameMeta/GameItem models, make GameLevelID nullable"
```

---

### Task 4: Update content_service.go — gameplay content fetching

**Files:**
- Modify: `dx-api/app/services/api/content_service.go:27-104`

- [ ] **Step 1: Update GetLevelContent to JOIN through game_items**

Replace the query block at lines 40-52 with:

```go
	query := facades.Orm().Query().
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id").
		Where("gi.game_level_id", gameLevelID).
		Where("content_items.is_active", true)

	// If degree is defined and has specific content type restrictions, filter by them
	if hasDegree && allowedTypes != nil {
		query = query.Where("content_items.content_type IN ?", allowedTypes)
	}

	var items []models.ContentItem
	if err := query.Order("content_items.\"order\" ASC").Get(&items); err != nil {
		return nil, fmt.Errorf("failed to get level content: %w", err)
	}
```

- [ ] **Step 2: Update the VIP guard level lookup**

The VIP guard at lines 29-35 reads `level.GameID` from `game_levels`. This still works because we query the `game_levels` table directly — no change needed here.

- [ ] **Step 3: Verify compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: may still have errors from other files — that's fine, this file should compile cleanly.

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/services/api/content_service.go
git commit -m "refactor: GetLevelContent queries through game_items junction table"
```

---

### Task 5: Update game_play_single_service.go — countLevelItems

**Files:**
- Modify: `dx-api/app/services/api/game_play_single_service.go:604-612`

- [ ] **Step 1: Update countLevelItems to JOIN through game_items**

Replace lines 604-612 with:

```go
func countLevelItems(query orm.Query, gameLevelID, degree string) (int64, error) {
	q := query.Model(&models.ContentItem{}).
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id").
		Where("gi.game_level_id", gameLevelID).
		Where("content_items.is_active", true)
	allowedTypes, ok := consts.DegreeContentTypes[degree]
	if ok && allowedTypes != nil {
		q = q.Where("content_items.content_type IN ?", allowedTypes)
	}
	return q.Count()
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/game_play_single_service.go
git commit -m "refactor: countLevelItems queries through game_items junction table"
```

---

### Task 6: Update course_content_service.go — read operations

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go`

- [ ] **Step 1: Update GetContentItemsByMeta (lines 167-251)**

Replace the metas query at lines 187-190 with:

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id").
		Where("gm.game_level_id", gameLevelID).
		Order("content_metas.\"order\" ASC").
		Get(&metas); err != nil {
		return nil, fmt.Errorf("failed to load metas: %w", err)
	}
```

Replace the items query at lines 199-204 with:

```go
	var items []models.ContentItem
	if err := facades.Orm().Query().
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id").
		Where("gi.game_level_id", gameLevelID).
		Where("content_items.is_active", true).
		Order("content_items.\"order\" ASC").
		Get(&items); err != nil {
		return nil, fmt.Errorf("failed to load items: %w", err)
	}
```

- [ ] **Step 2: Update verifyMetaBelongsToGame (lines 444-462)**

Replace entire function with:

```go
func verifyMetaBelongsToGame(metaID, gameID string) error {
	var gm models.GameMeta
	if err := facades.Orm().Query().
		Where("content_meta_id", metaID).
		Where("game_id", gameID).
		First(&gm); err != nil || gm.ID == "" {
		return ErrMetaNotFound
	}
	return nil
}
```

- [ ] **Step 3: Update verifyItemBelongsToGame (lines 465-483)**

Replace entire function with:

```go
func verifyItemBelongsToGame(itemID, gameID string) error {
	var gi models.GameItem
	if err := facades.Orm().Query().
		Where("content_item_id", itemID).
		Where("game_id", gameID).
		First(&gi); err != nil || gi.ID == "" {
		return ErrContentItemNotFound
	}
	return nil
}
```

- [ ] **Step 4: Update calculateInsertionOrder (lines 486-543)**

Replace the items query at lines 509-513 with:

```go
	var items []models.ContentItem
	if err := facades.Orm().Query().
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id").
		Where("gi.game_level_id", gameLevelID).
		Order("content_items.\"order\" ASC").
		Get(&items); err != nil {
		return 0, fmt.Errorf("failed to load items: %w", err)
	}
```

Also update the "append at end" query at lines 489-494:

```go
	if referenceItemID == "" {
		var lastItem models.ContentItem
		if err := facades.Orm().Query().
			Join("JOIN game_items gi ON gi.content_item_id = content_items.id").
			Where("gi.game_level_id", gameLevelID).
			Order("content_items.\"order\" DESC").
			First(&lastItem); err != nil || lastItem.ID == "" {
			return 1000, nil
		}
		return lastItem.Order + 1000, nil
	}
```

- [ ] **Step 5: Verify compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/services/api/course_content_service.go
git commit -m "refactor: course content reads query through junction tables"
```

---

### Task 7: Update course_content_service.go — write operations

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go`

- [ ] **Step 1: Update SaveMetadataBatch (lines 51-138)**

Add `"dx-api/app/models"` to imports if not already present. Also add `"github.com/google/uuid"` if not present.

Replace the existing capacity check query at lines 74-77. The capacity check now queries through game_metas:

```go
	// Check existing capacity via junction table
	var existingMetas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id").
		Where("gm.game_level_id", gameLevelID).
		Get(&existingMetas); err != nil {
		return 0, fmt.Errorf("failed to count metas: %w", err)
	}
```

Replace the meta creation loop at lines 120-135 to skip `GameLevelID` and add junction row:

```go
	for i, e := range entries {
		id := uuid.Must(uuid.NewV7()).String()
		meta := models.ContentMeta{
			ID:          id,
			SourceFrom:  sourceFrom,
			SourceType:  e.SourceType,
			SourceData:  e.SourceData,
			Translation: e.Translation,
			IsBreakDone: false,
			Order:       maxOrder + float64((i+1)*1000),
		}
		if err := facades.Orm().Query().Create(&meta); err != nil {
			return 0, fmt.Errorf("failed to create content meta: %w", err)
		}

		gm := models.GameMeta{
			ID:            uuid.Must(uuid.NewV7()).String(),
			GameID:        gameID,
			GameLevelID:   gameLevelID,
			ContentMetaID: id,
		}
		if err := facades.Orm().Query().Create(&gm); err != nil {
			return 0, fmt.Errorf("failed to create game meta: %w", err)
		}
	}
```

- [ ] **Step 2: Update InsertContentItem (lines 254-328)**

Replace the item count check at lines 289-295 to query through game_metas:

```go
	itemCount, err2 := facades.Orm().Query().Model(&models.ContentItem{}).
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id").
		Where("gi.game_level_id", gameLevelID).
		Where("content_items.content_meta_id", contentMetaID).
		Count()
	if err2 != nil {
		return nil, fmt.Errorf("failed to count items: %w", err2)
	}
```

Replace the item creation at lines 303-316 to skip `GameLevelID` and add junction row:

```go
	id := uuid.Must(uuid.NewV7()).String()
	item := models.ContentItem{
		ID:            id,
		ContentMetaID: &contentMetaID,
		Content:       content,
		ContentType:   contentType,
		Translation:   translation,
		Order:         order,
		IsActive:      true,
	}

	if err := facades.Orm().Query().Create(&item); err != nil {
		return nil, fmt.Errorf("failed to create content item: %w", err)
	}

	gi := models.GameItem{
		ID:            uuid.Must(uuid.NewV7()).String(),
		GameID:        gameID,
		GameLevelID:   gameLevelID,
		ContentItemID: id,
	}
	if err := facades.Orm().Query().Create(&gi); err != nil {
		return nil, fmt.Errorf("failed to create game item: %w", err)
	}
```

- [ ] **Step 3: Update DeleteContentItem (lines 384-406)**

Add junction row deletion before the content item deletion:

```go
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

	// Delete junction row
	if _, err := facades.Orm().Query().
		Where("content_item_id", itemID).Where("game_id", gameID).
		Delete(&models.GameItem{}); err != nil {
		return fmt.Errorf("failed to delete game item: %w", err)
	}

	// Delete content item if no other games reference it
	var remaining int64
	remaining, _ = facades.Orm().Query().Model(&models.GameItem{}).
		Where("content_item_id", itemID).Count()
	if remaining == 0 {
		if _, err := facades.Orm().Query().Where("id", itemID).Delete(&models.ContentItem{}); err != nil {
			return fmt.Errorf("failed to delete content item: %w", err)
		}
	}

	return nil
}
```

- [ ] **Step 4: Update DeleteAllLevelContent (lines 409-441)**

Replace the transaction at lines 432-440:

```go
	return facades.Orm().Transaction(func(tx orm.Query) error {
		// Delete junction rows for this level
		if _, err := tx.Where("game_level_id", gameLevelID).Where("game_id", gameID).Delete(&models.GameItem{}); err != nil {
			return fmt.Errorf("failed to delete game items: %w", err)
		}
		if _, err := tx.Where("game_level_id", gameLevelID).Where("game_id", gameID).Delete(&models.GameMeta{}); err != nil {
			return fmt.Errorf("failed to delete game metas: %w", err)
		}

		// Delete orphaned content (not referenced by any game)
		if _, err := tx.Exec(
			"DELETE FROM content_items WHERE id NOT IN (SELECT content_item_id FROM game_items)",
		); err != nil {
			return fmt.Errorf("failed to delete orphaned content items: %w", err)
		}
		if _, err := tx.Exec(
			"DELETE FROM content_metas WHERE id NOT IN (SELECT content_meta_id FROM game_metas)",
		); err != nil {
			return fmt.Errorf("failed to delete orphaned content metas: %w", err)
		}

		return nil
	})
```

- [ ] **Step 5: Verify compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/services/api/course_content_service.go
git commit -m "refactor: course content writes use junction tables"
```

---

### Task 8: Update ai_custom_service.go — BreakMetadata + processBreakMeta

**Files:**
- Modify: `dx-api/app/services/api/ai_custom_service.go:306-473`

- [ ] **Step 1: Update BreakMetadata to query metas through game_metas and pass gameID**

At lines 311-315, `verifyLevelOwnership` already returns `game`. Capture `game.ID`:

```go
	game, level, err := verifyLevelOwnership(userID, gameLevelID)
	if err != nil {
		writeSSEError(writer, err)
		return
	}
	if game.Status == consts.GameStatusPublished {
		writeSSEError(writer, ErrGamePublished)
		return
	}
	_ = level
	gameID := game.ID
```

Replace the metas query at lines 323-328:

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id").
		Where("gm.game_level_id", gameLevelID).
		Where("content_metas.is_break_done", false).
		Order("content_metas.\"order\" ASC").
		Get(&metas); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
		return
	}
```

Update the goroutine call at line 376 to pass `gameID` and `gameLevelID`:

```go
			success := processBreakMeta(m, gameID, gameLevelID)
```

- [ ] **Step 2: Update processBreakMeta signature and add junction row writes**

Change function signature at line 408 from:

```go
func processBreakMeta(meta models.ContentMeta, gameLevelID string) bool {
```

to:

```go
func processBreakMeta(meta models.ContentMeta, gameID, gameLevelID string) bool {
```

Replace the item creation at lines 450-461 to skip `GameLevelID` and add junction row:

```go
		item := models.ContentItem{
			ID:            id,
			ContentMetaID: &metaID,
			Content:       unit.Content,
			ContentType:   unit.ContentType,
			Translation:   translation,
			Order:         baseOrder + float64(i*10),
			IsActive:      true,
		}
		if err := facades.Orm().Query().Create(&item); err != nil {
			return false
		}

		gi := models.GameItem{
			ID:            uuid.Must(uuid.NewV7()).String(),
			GameID:        gameID,
			GameLevelID:   gameLevelID,
			ContentItemID: id,
		}
		if err := facades.Orm().Query().Create(&gi); err != nil {
			return false
		}
```

- [ ] **Step 3: Verify compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/services/api/ai_custom_service.go
git commit -m "refactor: BreakMetadata writes junction rows via game_items"
```

---

### Task 9: Update ai_custom_service.go — GenerateContentItems + processGenItems

**Files:**
- Modify: `dx-api/app/services/api/ai_custom_service.go:542-746`

- [ ] **Step 1: Update GenerateContentItems to query through game_metas**

After `verifyLevelOwnership` at line 547, capture `gameID`:

```go
	game, level, err := verifyLevelOwnership(userID, gameLevelID)
	if err != nil {
		writeSSEError(writer, err)
		return
	}
	if game.Status == consts.GameStatusPublished {
		writeSSEError(writer, ErrGamePublished)
		return
	}
	_ = level
	gameID := game.ID
```

Replace the metas query at lines 559-564:

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id").
		Where("gm.game_level_id", gameLevelID).
		Where("content_metas.is_break_done", true).
		Order("content_metas.\"order\" ASC").
		Get(&metas); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
		return
	}
```

The pending items query at lines 583-591 uses `WHERE content_meta_id IN ?` which queries by meta ID directly — **no change needed** since it doesn't use `game_level_id`.

`processGenItems` only updates existing items' `items` JSON column — **no change needed** to its signature or body.

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/ai_custom_service.go
git commit -m "refactor: GenerateContentItems queries through game_metas junction table"
```

---

### Task 10: Fix remaining compile errors from GameLevelID pointer change

**Files:**
- Modify: any files that reference `ContentMeta.GameLevelID` or `ContentItem.GameLevelID` as plain string

- [ ] **Step 1: Search for all remaining usages of GameLevelID on content models**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./... 2>&1
```

Review each compile error. Common fixes:

- `meta.GameLevelID` used as string → now needs nil check or was removed in previous tasks
- `item.GameLevelID` used as string → same treatment

For any remaining usages in `ai_custom_service.go` or `course_content_service.go` that were missed (e.g., the capacity check loop that reads `m.SourceType` — that doesn't touch `GameLevelID`, so it's fine):

Fix each error by either:
1. Removing the usage (if it was replaced by a junction table query)
2. Dereferencing with a nil check if still needed

- [ ] **Step 2: Verify clean compilation**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: no errors.

- [ ] **Step 3: Run go vet**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./...
```

Expected: no issues.

- [ ] **Step 4: Commit**

```bash
git add -u dx-api/
git commit -m "fix: resolve compile errors from GameLevelID pointer change"
```

---

### Task 11: Verify no lint issues

- [ ] **Step 1: Run full build**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: clean.

- [ ] **Step 2: Run go vet**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./...
```

Expected: clean.

- [ ] **Step 3: Run tests with race detector**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./...
```

Expected: all pass (or no tests exist yet — either way, no failures).

- [ ] **Step 4: Verify frontend still builds**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run build
```

Expected: clean build (no frontend changes were made).

---

### Task 12: Final review commit

- [ ] **Step 1: Review all changes**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git diff main...HEAD --stat
```

Verify the changed files match the expected scope from the file map.

- [ ] **Step 2: Verify no untracked files left behind**

```bash
git status
```

Expected: clean working tree on `feat/game-junction-tables` branch.

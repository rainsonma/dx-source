# Remove Junction Tables Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove `game_metas` and `game_items` junction tables, restore `game_level_id` directly on `content_metas` and `content_items`, and add `is_selective` to `games`.

**Architecture:** Replace all JOIN-based queries through junction tables with direct `WHERE game_level_id = ?` queries. Remove all GameMeta/GameItem record creation/deletion from write paths. Simplify cascade-delete logic to direct soft-deletes by `game_level_id`.

**Tech Stack:** Go / Goravel / GORM / PostgreSQL

**Build safety:** Tasks are ordered so the build passes after every task. Model fields are added first (non-breaking), then service files updated (junction models still exist), then junction model files deleted last.

---

### Task 1: Add fields to ContentMeta, ContentItem, and Game models

**Files:**
- Modify: `dx-api/app/models/content_meta.go`
- Modify: `dx-api/app/models/content_item.go`
- Modify: `dx-api/app/models/game.go`

- [ ] **Step 1: Add `GameLevelID` to ContentMeta model**

In `dx-api/app/models/content_meta.go`, add the field after `ID`:

```go
type ContentMeta struct {
	orm.Timestamps
	orm.SoftDeletes
	ID          string  `gorm:"column:id;primaryKey" json:"id"`
	GameLevelID string  `gorm:"column:game_level_id" json:"game_level_id"`
	SourceFrom  string  `gorm:"column:source_from" json:"source_from"`
	SourceType  string  `gorm:"column:source_type" json:"source_type"`
	SourceData  string  `gorm:"column:source_data" json:"source_data"`
	Translation *string `gorm:"column:translation" json:"translation"`
	IsBreakDone bool    `gorm:"column:is_break_done" json:"is_break_done"`
	Order       float64 `gorm:"column:order" json:"order"`
}
```

- [ ] **Step 2: Add `GameLevelID` to ContentItem model**

In `dx-api/app/models/content_item.go`, add the field after `ID`:

```go
type ContentItem struct {
	orm.Timestamps
	orm.SoftDeletes
	ID            string         `gorm:"column:id;primaryKey" json:"id"`
	GameLevelID   string         `gorm:"column:game_level_id" json:"game_level_id"`
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
	Order         float64        `gorm:"column:order" json:"order"`
	Tags          pq.StringArray `gorm:"column:tags;type:text[]" json:"tags"`
	IsActive      bool           `gorm:"column:is_active" json:"is_active"`
}
```

- [ ] **Step 3: Add `IsSelective` to Game model**

In `dx-api/app/models/game.go`, add the field after `Status`:

```go
type Game struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string  `gorm:"column:id;primaryKey" json:"id"`
	Name           string  `gorm:"column:name" json:"name"`
	Description    *string `gorm:"column:description" json:"description"`
	UserID         *string `gorm:"column:user_id" json:"user_id"`
	Mode           string  `gorm:"column:mode" json:"mode"`
	GameCategoryID *string `gorm:"column:game_category_id" json:"game_category_id"`
	GamePressID    *string `gorm:"column:game_press_id" json:"game_press_id"`
	Icon           *string `gorm:"column:icon" json:"icon"`
	CoverID        *string `gorm:"column:cover_id" json:"cover_id"`
	Order          float64 `gorm:"column:order" json:"order"`
	IsActive       bool    `gorm:"column:is_active" json:"is_active"`
	Status         string  `gorm:"column:status" json:"status"`
	IsSelective    bool    `gorm:"column:is_selective" json:"is_selective"`
}
```

- [ ] **Step 4: Verify build**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: SUCCESS (new fields are additive, nothing breaks)

---

### Task 2: Update `content_service.go` — `GetLevelContent()`

**Files:**
- Modify: `dx-api/app/services/api/content_service.go:40-43`

- [ ] **Step 1: Replace game_items JOIN with direct game_level_id filter**

Replace lines 40-43:

```go
	query := facades.Orm().Query().
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
		Where("gi.game_level_id", gameLevelID).
		Where("content_items.is_active", true)
```

With:

```go
	query := facades.Orm().Query().
		Where("content_items.game_level_id", gameLevelID).
		Where("content_items.is_active", true)
```

- [ ] **Step 2: Verify build**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: SUCCESS

---

### Task 3: Update `game_play_single_service.go` — `countLevelItems()`

This function is shared by single play, PK play, and group play.

**Files:**
- Modify: `dx-api/app/services/api/game_play_single_service.go:604-614`

- [ ] **Step 1: Replace game_items JOIN with direct game_level_id filter**

Replace the entire `countLevelItems` function (lines 604-614):

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
	return q.Count()
}
```

- [ ] **Step 2: Verify build**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: SUCCESS

---

### Task 4: Update `game_play_pk_service.go` — `spawnRobotForLevel()`

**Files:**
- Modify: `dx-api/app/services/api/game_play_pk_service.go:593-603`

- [ ] **Step 1: Replace game_items JOIN in robot content fetch**

Replace lines 593-603:

```go
	// Fetch content items for the level
	contentTypes := consts.DegreeContentTypes[degree]
	contentQuery := query.Model(&models.ContentItem{}).
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
		Where("gi.game_level_id", gameLevelID).
		Where("content_items.is_active", true)
	if len(contentTypes) > 0 {
		contentQuery = contentQuery.Where("content_items.content_type IN ?", contentTypes)
	}
	var items []models.ContentItem
	if err := contentQuery.Order("content_items.\"order\" asc").Get(&items); err != nil || len(items) == 0 {
```

With:

```go
	// Fetch content items for the level
	contentTypes := consts.DegreeContentTypes[degree]
	contentQuery := query.Model(&models.ContentItem{}).
		Where("game_level_id", gameLevelID).
		Where("is_active", true)
	if len(contentTypes) > 0 {
		contentQuery = contentQuery.Where("content_type IN ?", contentTypes)
	}
	var items []models.ContentItem
	if err := contentQuery.Order("\"order\" asc").Get(&items); err != nil || len(items) == 0 {
```

- [ ] **Step 2: Verify build**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: SUCCESS

---

### Task 5: Update `course_content_service.go` — all functions

This is the largest change. 8 functions need updating.

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go`

- [ ] **Step 1: Update `SaveMetadataBatch()` — replace game_metas JOIN + remove GameMeta creation**

Replace lines 75-81 (existing metas query):

```go
	var existingMetas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
		Where("gm.game_level_id", gameLevelID).
		Get(&existingMetas); err != nil {
		return 0, fmt.Errorf("failed to count metas: %w", err)
	}
```

With:

```go
	var existingMetas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Get(&existingMetas); err != nil {
		return 0, fmt.Errorf("failed to count metas: %w", err)
	}
```

Replace lines 131-155 (create metas loop):

```go
	// Create metas in batch
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

With:

```go
	// Create metas in batch
	for i, e := range entries {
		id := uuid.Must(uuid.NewV7()).String()
		meta := models.ContentMeta{
			ID:          id,
			GameLevelID: gameLevelID,
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
	}
```

- [ ] **Step 2: Update `GetContentItemsByMeta()` — replace both JOINs**

Replace lines 207-213 (metas query):

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
		Where("gm.game_level_id", gameLevelID).
		Order("content_metas.\"order\" ASC").
		Get(&metas); err != nil {
		return nil, fmt.Errorf("failed to load metas: %w", err)
	}
```

With:

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Order("\"order\" ASC").
		Get(&metas); err != nil {
		return nil, fmt.Errorf("failed to load metas: %w", err)
	}
```

Replace lines 221-228 (items query):

```go
	var items []models.ContentItem
	if err := facades.Orm().Query().
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
		Where("gi.game_level_id", gameLevelID).
		Where("content_items.is_active", true).
		Order("content_items.\"order\" ASC").
		Get(&items); err != nil {
		return nil, fmt.Errorf("failed to load items: %w", err)
	}
```

With:

```go
	var items []models.ContentItem
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("is_active", true).
		Order("\"order\" ASC").
		Get(&items); err != nil {
		return nil, fmt.Errorf("failed to load items: %w", err)
	}
```

- [ ] **Step 3: Update `InsertContentItem()` — replace JOIN, set GameLevelID, remove GameItem creation**

Replace lines 315-323 (item count query):

```go
	itemCount, err2 := facades.Orm().Query().Model(&models.ContentItem{}).
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
		Where("gi.game_level_id", gameLevelID).
		Where("content_items.content_meta_id", contentMetaID).
		Count()
	if err2 != nil {
		return nil, fmt.Errorf("failed to count items: %w", err2)
	}
	if itemCount >= int64(MaxItemsPerMeta) {
		return nil, ErrItemLimitExceeded
	}
```

With:

```go
	itemCount, err2 := facades.Orm().Query().Model(&models.ContentItem{}).
		Where("game_level_id", gameLevelID).
		Where("content_meta_id", contentMetaID).
		Count()
	if err2 != nil {
		return nil, fmt.Errorf("failed to count items: %w", err2)
	}
	if itemCount >= int64(MaxItemsPerMeta) {
		return nil, ErrItemLimitExceeded
	}
```

Replace lines 334-356 (create item + GameItem):

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

With:

```go
	id := uuid.Must(uuid.NewV7()).String()
	item := models.ContentItem{
		ID:            id,
		GameLevelID:   gameLevelID,
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
```

- [ ] **Step 4: Update `DeleteContentItem()` — remove GameItem deletion, simplify**

Replace the entire transaction body (lines 440-472):

```go
	return facades.Orm().Transaction(func(tx orm.Query) error {
		// Soft-delete junction row
		if _, err := tx.Where("content_item_id", itemID).Where("game_id", gameID).Delete(&models.GameItem{}); err != nil {
			return fmt.Errorf("failed to delete game item: %w", err)
		}

		// Soft-delete content item if no other game references it
		if _, err := tx.Exec(
			"UPDATE content_items SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM game_items WHERE content_item_id = ? AND deleted_at IS NULL)",
			itemID, itemID,
		); err != nil {
			return fmt.Errorf("failed to delete content item: %w", err)
		}

		// Reset is_break_done when meta has no remaining items in this game
		if _, err := tx.Exec(
			`UPDATE content_metas SET is_break_done = false
			 WHERE id = (SELECT content_meta_id FROM content_items WHERE id = ?)
			   AND deleted_at IS NULL
			   AND NOT EXISTS (
			     SELECT 1 FROM content_items ci
			     JOIN game_items gi ON gi.content_item_id = ci.id AND gi.deleted_at IS NULL
			     WHERE ci.content_meta_id = content_metas.id
			       AND ci.deleted_at IS NULL
			       AND gi.game_id = ?
			   )`,
			itemID, gameID,
		); err != nil {
			return fmt.Errorf("failed to reset meta break status: %w", err)
		}
		return nil
	})
```

With:

```go
	return facades.Orm().Transaction(func(tx orm.Query) error {
		// Soft-delete content item
		if _, err := tx.Exec(
			"UPDATE content_items SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL",
			itemID,
		); err != nil {
			return fmt.Errorf("failed to delete content item: %w", err)
		}

		// Reset is_break_done when meta has no remaining active items
		if _, err := tx.Exec(
			`UPDATE content_metas SET is_break_done = false
			 WHERE id = (SELECT content_meta_id FROM content_items WHERE id = ?)
			   AND deleted_at IS NULL
			   AND NOT EXISTS (
			     SELECT 1 FROM content_items
			     WHERE content_meta_id = content_metas.id
			       AND deleted_at IS NULL
			   )`,
			itemID,
		); err != nil {
			return fmt.Errorf("failed to reset meta break status: %w", err)
		}
		return nil
	})
```

- [ ] **Step 5: Update `DeleteAllLevelContent()` — simplify to direct deletes**

Replace the entire transaction body (lines 498-528):

```go
	return facades.Orm().Transaction(func(tx orm.Query) error {
		// Soft-delete junction rows for this level
		if _, err := tx.Where("game_level_id", gameLevelID).Where("game_id", gameID).Delete(&models.GameItem{}); err != nil {
			return fmt.Errorf("failed to delete game items: %w", err)
		}
		if _, err := tx.Where("game_level_id", gameLevelID).Where("game_id", gameID).Delete(&models.GameMeta{}); err != nil {
			return fmt.Errorf("failed to delete game metas: %w", err)
		}

		// Soft-delete orphaned content scoped to this level only
		if _, err := tx.Exec(
			`UPDATE content_items SET deleted_at = NOW()
			 WHERE deleted_at IS NULL
			   AND id IN (SELECT gi.content_item_id FROM game_items gi WHERE gi.game_level_id = ? AND gi.game_id = ? AND gi.deleted_at IS NOT NULL)
			   AND NOT EXISTS (SELECT 1 FROM game_items WHERE content_item_id = content_items.id AND deleted_at IS NULL)`,
			gameLevelID, gameID,
		); err != nil {
			return fmt.Errorf("failed to delete orphaned content items: %w", err)
		}
		if _, err := tx.Exec(
			`UPDATE content_metas SET deleted_at = NOW()
			 WHERE deleted_at IS NULL
			   AND id IN (SELECT gm.content_meta_id FROM game_metas gm WHERE gm.game_level_id = ? AND gm.game_id = ? AND gm.deleted_at IS NOT NULL)
			   AND NOT EXISTS (SELECT 1 FROM game_metas WHERE content_meta_id = content_metas.id AND deleted_at IS NULL)`,
			gameLevelID, gameID,
		); err != nil {
			return fmt.Errorf("failed to delete orphaned content metas: %w", err)
		}

		return nil
	})
```

With:

```go
	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Exec(
			"UPDATE content_items SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL",
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to delete content items: %w", err)
		}
		if _, err := tx.Exec(
			"UPDATE content_metas SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL",
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to delete content metas: %w", err)
		}
		return nil
	})
```

- [ ] **Step 6: Update `DeleteMetadata()` — simplify to direct deletes**

Replace the entire transaction body (lines 549-581):

```go
	return facades.Orm().Transaction(func(tx orm.Query) error {
		// Soft-delete game_items whose content_items belong to this meta
		if _, err := tx.Exec(
			"UPDATE game_items SET deleted_at = NOW() WHERE game_id = ? AND deleted_at IS NULL AND content_item_id IN (SELECT id FROM content_items WHERE content_meta_id = ? AND deleted_at IS NULL)",
			gameID, metaID,
		); err != nil {
			return fmt.Errorf("failed to delete game items for meta: %w", err)
		}

		// Soft-delete game_metas junction row
		if _, err := tx.Where("content_meta_id", metaID).Where("game_id", gameID).Delete(&models.GameMeta{}); err != nil {
			return fmt.Errorf("failed to delete game meta: %w", err)
		}

		// Soft-delete orphaned content items (NOT EXISTS uses index, avoids full scan)
		if _, err := tx.Exec(
			"UPDATE content_items SET deleted_at = NOW() WHERE content_meta_id = ? AND deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM game_items WHERE content_item_id = content_items.id AND deleted_at IS NULL)",
			metaID,
		); err != nil {
			return fmt.Errorf("failed to delete orphaned content items: %w", err)
		}

		// Soft-delete orphaned content meta
		if _, err := tx.Exec(
			"UPDATE content_metas SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM game_metas WHERE content_meta_id = content_metas.id AND deleted_at IS NULL)",
			metaID,
		); err != nil {
			return fmt.Errorf("failed to delete orphaned content meta: %w", err)
		}

		return nil
	})
```

With:

```go
	return facades.Orm().Transaction(func(tx orm.Query) error {
		// Soft-delete content items belonging to this meta
		if _, err := tx.Exec(
			"UPDATE content_items SET deleted_at = NOW() WHERE content_meta_id = ? AND deleted_at IS NULL",
			metaID,
		); err != nil {
			return fmt.Errorf("failed to delete content items for meta: %w", err)
		}

		// Soft-delete the content meta
		if _, err := tx.Exec(
			"UPDATE content_metas SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL",
			metaID,
		); err != nil {
			return fmt.Errorf("failed to delete content meta: %w", err)
		}

		return nil
	})
```

- [ ] **Step 7: Update `verifyMetaBelongsToGame()` — query via game_levels instead of game_metas**

Replace lines 584-593:

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

With:

```go
func verifyMetaBelongsToGame(metaID, gameID string) error {
	var meta models.ContentMeta
	if err := facades.Orm().Query().
		Where("id", metaID).
		Where("game_level_id IN (SELECT id FROM game_levels WHERE game_id = ? AND deleted_at IS NULL)", gameID).
		First(&meta); err != nil || meta.ID == "" {
		return ErrMetaNotFound
	}
	return nil
}
```

- [ ] **Step 8: Update `verifyItemBelongsToGame()` — query via game_levels instead of game_items**

Replace lines 596-605:

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

With:

```go
func verifyItemBelongsToGame(itemID, gameID string) error {
	var item models.ContentItem
	if err := facades.Orm().Query().
		Where("id", itemID).
		Where("game_level_id IN (SELECT id FROM game_levels WHERE game_id = ? AND deleted_at IS NULL)", gameID).
		First(&item); err != nil || item.ID == "" {
		return ErrContentItemNotFound
	}
	return nil
}
```

- [ ] **Step 9: Update `calculateInsertionOrder()` — replace game_items JOINs**

Replace lines 611-616 (last item query):

```go
		if err := facades.Orm().Query().
			Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
			Where("gi.game_level_id", gameLevelID).
			Order("content_items.\"order\" DESC").
			First(&lastItem); err != nil || lastItem.ID == "" {
			return 1000, nil
		}
```

With:

```go
		if err := facades.Orm().Query().
			Where("game_level_id", gameLevelID).
			Order("\"order\" DESC").
			First(&lastItem); err != nil || lastItem.ID == "" {
			return 1000, nil
		}
```

Replace lines 631-636 (items list query):

```go
	if err := facades.Orm().Query().
		Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
		Where("gi.game_level_id", gameLevelID).
		Order("content_items.\"order\" ASC").
		Get(&items); err != nil {
		return 0, fmt.Errorf("failed to load items: %w", err)
	}
```

With:

```go
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Order("\"order\" ASC").
		Get(&items); err != nil {
		return 0, fmt.Errorf("failed to load items: %w", err)
	}
```

- [ ] **Step 10: Remove unused imports from `course_content_service.go`**

The `models.GameMeta` and `models.GameItem` references are now gone. The file should no longer need the `"github.com/google/uuid"` import for GameMeta/GameItem UUIDs, but it still uses `uuid` for ContentMeta and ContentItem IDs. Keep the import. No import changes needed.

- [ ] **Step 11: Verify build**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: SUCCESS

---

### Task 6: Update `course_game_service.go` — delete/publish/detail functions

**Files:**
- Modify: `dx-api/app/services/api/course_game_service.go`

- [ ] **Step 1: Update `DeleteGame()` — remove junction cascade**

Replace the transaction body (lines 240-291):

```go
	return facades.Orm().Transaction(func(tx orm.Query) error {
		// Get level IDs for cascade
		var levels []models.GameLevel
		if err := tx.Where("game_id", gameID).Get(&levels); err != nil {
			return fmt.Errorf("failed to load levels: %w", err)
		}

		levelIDs := make([]string, 0, len(levels))
		for _, l := range levels {
			levelIDs = append(levelIDs, l.ID)
		}

		// Cascade soft-delete junction rows, then orphaned content
		if len(levelIDs) > 0 {
			if _, err := tx.Where("game_id", gameID).Delete(&models.GameItem{}); err != nil {
				return fmt.Errorf("failed to delete game items: %w", err)
			}
			if _, err := tx.Where("game_id", gameID).Delete(&models.GameMeta{}); err != nil {
				return fmt.Errorf("failed to delete game metas: %w", err)
			}
			if _, err := tx.Exec(
				`UPDATE content_items SET deleted_at = NOW()
				 WHERE deleted_at IS NULL
				   AND id IN (SELECT gi.content_item_id FROM game_items gi WHERE gi.game_id = ? AND gi.deleted_at IS NOT NULL)
				   AND NOT EXISTS (SELECT 1 FROM game_items WHERE content_item_id = content_items.id AND deleted_at IS NULL)`,
				gameID,
			); err != nil {
				return fmt.Errorf("failed to delete orphaned content items: %w", err)
			}
			if _, err := tx.Exec(
				`UPDATE content_metas SET deleted_at = NOW()
				 WHERE deleted_at IS NULL
				   AND id IN (SELECT gm.content_meta_id FROM game_metas gm WHERE gm.game_id = ? AND gm.deleted_at IS NOT NULL)
				   AND NOT EXISTS (SELECT 1 FROM game_metas WHERE content_meta_id = content_metas.id AND deleted_at IS NULL)`,
				gameID,
			); err != nil {
				return fmt.Errorf("failed to delete orphaned content metas: %w", err)
			}
		}

		// Soft-delete levels
		if _, err := tx.Where("game_id", gameID).Delete(&models.GameLevel{}); err != nil {
			return fmt.Errorf("failed to delete levels: %w", err)
		}

		// Soft-delete game
		if _, err := tx.Where("id", gameID).Delete(&models.Game{}); err != nil {
			return fmt.Errorf("failed to delete game: %w", err)
		}

		return nil
	})
```

With:

```go
	return facades.Orm().Transaction(func(tx orm.Query) error {
		// Get level IDs for cascade
		var levels []models.GameLevel
		if err := tx.Where("game_id", gameID).Get(&levels); err != nil {
			return fmt.Errorf("failed to load levels: %w", err)
		}

		levelIDs := make([]string, 0, len(levels))
		for _, l := range levels {
			levelIDs = append(levelIDs, l.ID)
		}

		// Cascade soft-delete content
		if len(levelIDs) > 0 {
			if _, err := tx.Exec(
				"UPDATE content_items SET deleted_at = NOW() WHERE game_level_id IN ? AND deleted_at IS NULL",
				levelIDs,
			); err != nil {
				return fmt.Errorf("failed to delete content items: %w", err)
			}
			if _, err := tx.Exec(
				"UPDATE content_metas SET deleted_at = NOW() WHERE game_level_id IN ? AND deleted_at IS NULL",
				levelIDs,
			); err != nil {
				return fmt.Errorf("failed to delete content metas: %w", err)
			}
		}

		// Soft-delete levels
		if _, err := tx.Where("game_id", gameID).Delete(&models.GameLevel{}); err != nil {
			return fmt.Errorf("failed to delete levels: %w", err)
		}

		// Soft-delete game
		if _, err := tx.Where("id", gameID).Delete(&models.Game{}); err != nil {
			return fmt.Errorf("failed to delete game: %w", err)
		}

		return nil
	})
```

- [ ] **Step 2: Update `DeleteLevel()` — remove junction cascade**

Replace the transaction body (lines 452-482):

```go
	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Where("game_level_id", levelID).Where("game_id", gameID).Delete(&models.GameItem{}); err != nil {
			return fmt.Errorf("failed to delete game items: %w", err)
		}
		if _, err := tx.Where("game_level_id", levelID).Where("game_id", gameID).Delete(&models.GameMeta{}); err != nil {
			return fmt.Errorf("failed to delete game metas: %w", err)
		}
		// Scoped orphan cleanup — only items/metas from this level
		if _, err := tx.Exec(
			`UPDATE content_items SET deleted_at = NOW()
			 WHERE deleted_at IS NULL
			   AND id IN (SELECT gi.content_item_id FROM game_items gi WHERE gi.game_level_id = ? AND gi.game_id = ? AND gi.deleted_at IS NOT NULL)
			   AND NOT EXISTS (SELECT 1 FROM game_items WHERE content_item_id = content_items.id AND deleted_at IS NULL)`,
			levelID, gameID,
		); err != nil {
			return fmt.Errorf("failed to delete orphaned content items: %w", err)
		}
		if _, err := tx.Exec(
			`UPDATE content_metas SET deleted_at = NOW()
			 WHERE deleted_at IS NULL
			   AND id IN (SELECT gm.content_meta_id FROM game_metas gm WHERE gm.game_level_id = ? AND gm.game_id = ? AND gm.deleted_at IS NOT NULL)
			   AND NOT EXISTS (SELECT 1 FROM game_metas WHERE content_meta_id = content_metas.id AND deleted_at IS NULL)`,
			levelID, gameID,
		); err != nil {
			return fmt.Errorf("failed to delete orphaned content metas: %w", err)
		}
		if _, err := tx.Where("id", levelID).Delete(&models.GameLevel{}); err != nil {
			return fmt.Errorf("failed to delete level: %w", err)
		}
		return nil
	})
```

With:

```go
	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Exec(
			"UPDATE content_items SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL",
			levelID,
		); err != nil {
			return fmt.Errorf("failed to delete content items: %w", err)
		}
		if _, err := tx.Exec(
			"UPDATE content_metas SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL",
			levelID,
		); err != nil {
			return fmt.Errorf("failed to delete content metas: %w", err)
		}
		if _, err := tx.Where("id", levelID).Delete(&models.GameLevel{}); err != nil {
			return fmt.Errorf("failed to delete level: %w", err)
		}
		return nil
	})
```

- [ ] **Step 3: Update `PublishGame()` — replace game_items JOINs**

Replace lines 324-348 (the level validation loop):

```go
	for _, l := range levels {
		itemCount, err3 := facades.Orm().Query().Model(&models.ContentItem{}).
			Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
			Where("gi.game_level_id", l.ID).
			Where("content_items.is_active", true).
			Count()
		if err3 != nil {
			return fmt.Errorf("failed to count items: %w", err3)
		}
		if itemCount == 0 {
			return fmt.Errorf("关卡「%s」没有练习内容", l.Name)
		}

		ungeneratedCount, err4 := facades.Orm().Query().Model(&models.ContentItem{}).
			Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
			Where("gi.game_level_id", l.ID).
			Where("content_items.is_active", true).
			Where("content_items.items IS NULL").
			Count()
		if err4 != nil {
			return fmt.Errorf("failed to count ungenerated items: %w", err4)
		}
		if ungeneratedCount > 0 {
			return fmt.Errorf("关卡「%s」有未生成的练习单元", l.Name)
		}
	}
```

With:

```go
	for _, l := range levels {
		itemCount, err3 := facades.Orm().Query().Model(&models.ContentItem{}).
			Where("game_level_id", l.ID).
			Where("is_active", true).
			Count()
		if err3 != nil {
			return fmt.Errorf("failed to count items: %w", err3)
		}
		if itemCount == 0 {
			return fmt.Errorf("关卡「%s」没有练习内容", l.Name)
		}

		ungeneratedCount, err4 := facades.Orm().Query().Model(&models.ContentItem{}).
			Where("game_level_id", l.ID).
			Where("is_active", true).
			Where("items IS NULL").
			Count()
		if err4 != nil {
			return fmt.Errorf("failed to count ungenerated items: %w", err4)
		}
		if ungeneratedCount > 0 {
			return fmt.Errorf("关卡「%s」有未生成的练习单元", l.Name)
		}
	}
```

- [ ] **Step 4: Update `GetCourseGameDetail()` — replace game_items JOIN for item count**

Replace lines 545-549:

```go
		itemCount, _ := facades.Orm().Query().Model(&models.ContentItem{}).
			Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
			Where("gi.game_level_id", l.ID).
			Where("content_items.is_active", true).
			Count()
```

With:

```go
		itemCount, _ := facades.Orm().Query().Model(&models.ContentItem{}).
			Where("game_level_id", l.ID).
			Where("is_active", true).
			Count()
```

- [ ] **Step 5: Verify build**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: SUCCESS

---

### Task 7: Update `ai_custom_service.go` — break/generate functions

**Files:**
- Modify: `dx-api/app/services/api/ai_custom_service.go`

- [ ] **Step 1: Update `BreakMetadata()` — replace game_metas JOIN**

Replace lines 325-332:

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
		Where("gm.game_level_id", gameLevelID).
		Where("content_metas.is_break_done", false).
		Order("content_metas.\"order\" ASC").
		Get(&metas); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
```

With:

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("is_break_done", false).
		Order("\"order\" ASC").
		Get(&metas); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
```

- [ ] **Step 2: Update `processBreakMeta()` — set GameLevelID, remove GameItem creation**

Replace lines 452-473 (the item creation + GameItem creation inside the for loop):

```go
		id := uuid.Must(uuid.NewV7()).String()
		metaID := meta.ID
		var translation *string
		if unit.Translation != "" {
			translation = &unit.Translation
		}

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

With:

```go
		id := uuid.Must(uuid.NewV7()).String()
		metaID := meta.ID
		var translation *string
		if unit.Translation != "" {
			translation = &unit.Translation
		}

		item := models.ContentItem{
			ID:            id,
			GameLevelID:   gameLevelID,
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
```

- [ ] **Step 3: Update `GenerateContentItems()` — replace game_metas JOIN**

Replace lines 571-579:

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
		Where("gm.game_level_id", gameLevelID).
		Where("content_metas.is_break_done", true).
		Order("content_metas.\"order\" ASC").
		Get(&metas); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
		return
	}
```

With:

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("is_break_done", true).
		Order("\"order\" ASC").
		Get(&metas); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
		return
	}
```

- [ ] **Step 4: Verify build**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: SUCCESS

---

### Task 8: Update `ai_custom_vocab_service.go` — break/generate functions

**Files:**
- Modify: `dx-api/app/services/api/ai_custom_vocab_service.go`

- [ ] **Step 1: Update `BreakVocabMetadata()` — replace game_metas JOIN**

Replace lines 200-207:

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
		Where("gm.game_level_id", gameLevelID).
		Where("content_metas.is_break_done", false).
		Order("content_metas.\"order\" ASC").
		Get(&metas); err != nil {
		writeVocabSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
```

With:

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("is_break_done", false).
		Order("\"order\" ASC").
		Get(&metas); err != nil {
		writeVocabSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
```

- [ ] **Step 2: Update `processVocabBreakMeta()` — set GameLevelID, remove GameItem creation**

Replace lines 279-303:

```go
	contentType := "word"
	if strings.Contains(strings.TrimSpace(meta.SourceData), " ") {
		contentType = "phrase"
	}

	id := uuid.Must(uuid.NewV7()).String()
	metaID := meta.ID

	item := models.ContentItem{
		ID:            id,
		ContentMetaID: &metaID,
		Content:       meta.SourceData,
		ContentType:   contentType,
		Translation:   meta.Translation,
		Order:         meta.Order + 10,
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

With:

```go
	contentType := "word"
	if strings.Contains(strings.TrimSpace(meta.SourceData), " ") {
		contentType = "phrase"
	}

	id := uuid.Must(uuid.NewV7()).String()
	metaID := meta.ID

	item := models.ContentItem{
		ID:            id,
		GameLevelID:   gameLevelID,
		ContentMetaID: &metaID,
		Content:       meta.SourceData,
		ContentType:   contentType,
		Translation:   meta.Translation,
		Order:         meta.Order + 10,
		IsActive:      true,
	}
	if err := facades.Orm().Query().Create(&item); err != nil {
		return false
	}
```

- [ ] **Step 3: Update `GenerateVocabContentItems()` — replace game_metas JOIN**

Replace lines 336-343:

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
		Where("gm.game_level_id", gameLevelID).
		Where("content_metas.is_break_done", true).
		Order("content_metas.\"order\" ASC").
		Get(&metas); err != nil {
		writeVocabSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
```

With:

```go
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("is_break_done", true).
		Order("\"order\" ASC").
		Get(&metas); err != nil {
		writeVocabSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
```

- [ ] **Step 4: Verify build**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: SUCCESS

---

### Task 9: Update `import_courses.go` — remove GameItem batch creation

**Files:**
- Modify: `dx-api/app/console/commands/import_courses.go`

- [ ] **Step 1: Update `insertLevels()` — set GameLevelID on ContentItem, remove createGameItemsBatch calls**

Replace lines 350-396 (the content item creation loop inside `insertLevels`):

```go
		// Build content items in batches
		var batch []models.ContentItem
		for _, item := range level.Sentences {
			items, err := transformItems(item.Content, item.WordDetails)
			if err != nil {
				return fmt.Errorf("failed to transform items for %q: %w", item.Content, err)
			}

			structure, err := transformStructure(item.SentenceStructure)
			if err != nil {
				return fmt.Errorf("failed to transform structure for %q: %w", item.Content, err)
			}

			ci := models.ContentItem{
				ID:          uuid.Must(uuid.NewV7()).String(),
				Content:     item.Content,
				ContentType: item.Type,
				Translation: &item.Chinese,
				Items:       &items,
				Structure:   structure,
				Order:       float64(item.SortOrder * 1000),
				IsActive:    true,
			}
			batch = append(batch, ci)

			if len(batch) >= batchSize {
				if err := tx.Create(&batch); err != nil {
					return fmt.Errorf("failed to batch create content items: %w", err)
				}
				if err := createGameItemsBatch(tx, gameID, levelID, batch); err != nil {
					return err
				}
				batch = batch[:0]
			}
		}

		// Flush remaining
		if len(batch) > 0 {
			if err := tx.Create(&batch); err != nil {
				return fmt.Errorf("failed to batch create remaining content items: %w", err)
			}
			if err := createGameItemsBatch(tx, gameID, levelID, batch); err != nil {
				return err
			}
		}
```

With:

```go
		// Build content items in batches
		var batch []models.ContentItem
		for _, item := range level.Sentences {
			items, err := transformItems(item.Content, item.WordDetails)
			if err != nil {
				return fmt.Errorf("failed to transform items for %q: %w", item.Content, err)
			}

			structure, err := transformStructure(item.SentenceStructure)
			if err != nil {
				return fmt.Errorf("failed to transform structure for %q: %w", item.Content, err)
			}

			ci := models.ContentItem{
				ID:          uuid.Must(uuid.NewV7()).String(),
				GameLevelID: levelID,
				Content:     item.Content,
				ContentType: item.Type,
				Translation: &item.Chinese,
				Items:       &items,
				Structure:   structure,
				Order:       float64(item.SortOrder * 1000),
				IsActive:    true,
			}
			batch = append(batch, ci)

			if len(batch) >= batchSize {
				if err := tx.Create(&batch); err != nil {
					return fmt.Errorf("failed to batch create content items: %w", err)
				}
				batch = batch[:0]
			}
		}

		// Flush remaining
		if len(batch) > 0 {
			if err := tx.Create(&batch); err != nil {
				return fmt.Errorf("failed to batch create remaining content items: %w", err)
			}
		}
```

- [ ] **Step 2: Delete `createGameItemsBatch()` function**

Delete the entire function (lines 432-446):

```go
// createGameItemsBatch creates game_items junction rows for a batch of content items.
func createGameItemsBatch(tx orm.Query, gameID, levelID string, items []models.ContentItem) error {
	var batch []models.GameItem
	for _, ci := range items {
		batch = append(batch, models.GameItem{
			ID:            uuid.Must(uuid.NewV7()).String(),
			GameID:        gameID,
			GameLevelID:   levelID,
			ContentItemID: ci.ID,
		})
	}
	if err := tx.Create(&batch); err != nil {
		return fmt.Errorf("failed to batch create game items: %w", err)
	}
	return nil
}
```

- [ ] **Step 3: Update `forceCleanup()` — remove junction table cleanup**

Replace lines 231-271:

```go
func forceCleanup(categoryID string, names []string) (int, error) {
	query := facades.Orm().Query()
	var games []models.Game
	if err := query.
		Where("game_category_id", categoryID).
		Where("name IN ?", names).
		Get(&games); err != nil {
		return 0, fmt.Errorf("failed to query games for cleanup: %w", err)
	}

	for _, game := range games {
		var levels []models.GameLevel
		if err := query.Where("game_id", game.ID).Get(&levels); err != nil {
			return 0, fmt.Errorf("failed to query levels for game %s: %w", game.ID, err)
		}

		// Delete junction rows and orphaned content for all levels
		if _, err := query.Where("game_id", game.ID).Delete(&models.GameItem{}); err != nil {
			return 0, fmt.Errorf("failed to delete game items for game %s: %w", game.ID, err)
		}
		if _, err := query.Where("game_id", game.ID).Delete(&models.GameMeta{}); err != nil {
			return 0, fmt.Errorf("failed to delete game metas for game %s: %w", game.ID, err)
		}
		if _, err := query.Exec("UPDATE content_items SET deleted_at = NOW() WHERE deleted_at IS NULL AND id NOT IN (SELECT content_item_id FROM game_items WHERE deleted_at IS NULL)"); err != nil {
			return 0, fmt.Errorf("failed to delete orphaned content items: %w", err)
		}
		if _, err := query.Exec("UPDATE content_metas SET deleted_at = NOW() WHERE deleted_at IS NULL AND id NOT IN (SELECT content_meta_id FROM game_metas WHERE deleted_at IS NULL)"); err != nil {
			return 0, fmt.Errorf("failed to delete orphaned content metas: %w", err)
		}

		if _, err := query.Where("game_id", game.ID).Delete(&models.GameLevel{}); err != nil {
			return 0, fmt.Errorf("failed to delete levels for game %s: %w", game.ID, err)
		}

		if _, err := query.Where("id", game.ID).Delete(&models.Game{}); err != nil {
			return 0, fmt.Errorf("failed to delete game %s: %w", game.ID, err)
		}
	}

	return len(games), nil
}
```

With:

```go
func forceCleanup(categoryID string, names []string) (int, error) {
	query := facades.Orm().Query()
	var games []models.Game
	if err := query.
		Where("game_category_id", categoryID).
		Where("name IN ?", names).
		Get(&games); err != nil {
		return 0, fmt.Errorf("failed to query games for cleanup: %w", err)
	}

	for _, game := range games {
		var levels []models.GameLevel
		if err := query.Where("game_id", game.ID).Get(&levels); err != nil {
			return 0, fmt.Errorf("failed to query levels for game %s: %w", game.ID, err)
		}

		levelIDs := make([]string, 0, len(levels))
		for _, l := range levels {
			levelIDs = append(levelIDs, l.ID)
		}

		// Delete content by game_level_id
		if len(levelIDs) > 0 {
			if _, err := query.Exec(
				"UPDATE content_items SET deleted_at = NOW() WHERE game_level_id IN ? AND deleted_at IS NULL",
				levelIDs,
			); err != nil {
				return 0, fmt.Errorf("failed to delete content items for game %s: %w", game.ID, err)
			}
			if _, err := query.Exec(
				"UPDATE content_metas SET deleted_at = NOW() WHERE game_level_id IN ? AND deleted_at IS NULL",
				levelIDs,
			); err != nil {
				return 0, fmt.Errorf("failed to delete content metas for game %s: %w", game.ID, err)
			}
		}

		if _, err := query.Where("game_id", game.ID).Delete(&models.GameLevel{}); err != nil {
			return 0, fmt.Errorf("failed to delete levels for game %s: %w", game.ID, err)
		}

		if _, err := query.Where("id", game.ID).Delete(&models.Game{}); err != nil {
			return 0, fmt.Errorf("failed to delete game %s: %w", game.ID, err)
		}
	}

	return len(games), nil
}
```

- [ ] **Step 4: Verify build**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: SUCCESS

---

### Task 10: Delete junction model files and clean up migration

**Files:**
- Delete: `dx-api/app/models/game_meta.go`
- Delete: `dx-api/app/models/game_item.go`
- Modify: `dx-api/database/migrations/20260407000001_create_game_junction_tables.go`

- [ ] **Step 1: Verify no remaining references to GameMeta or GameItem**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && grep -r "GameMeta\|GameItem\|game_metas\|game_items" --include="*.go" app/ | grep -v "_test.go"`
Expected: NO OUTPUT (zero matches)

If any matches remain, fix them before proceeding.

- [ ] **Step 2: Delete `dx-api/app/models/game_meta.go`**

```bash
rm dx-api/app/models/game_meta.go
```

- [ ] **Step 3: Delete `dx-api/app/models/game_item.go`**

```bash
rm dx-api/app/models/game_item.go
```

- [ ] **Step 4: Make junction migration a no-op**

Replace the entire content of `dx-api/database/migrations/20260407000001_create_game_junction_tables.go`:

```go
package migrations

type M20260407000001CreateGameJunctionTables struct{}

func (r *M20260407000001CreateGameJunctionTables) Signature() string {
	return "20260407000001_create_game_junction_tables"
}

func (r *M20260407000001CreateGameJunctionTables) Up() error {
	return nil
}

func (r *M20260407000001CreateGameJunctionTables) Down() error {
	return nil
}
```

- [ ] **Step 5: Verify build**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: SUCCESS

- [ ] **Step 6: Run go vet**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./...`
Expected: SUCCESS (zero warnings)

---

### Task 11: Final verification

- [ ] **Step 1: Verify no lint issues**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./... && go vet ./...`
Expected: Both pass with zero errors/warnings

- [ ] **Step 2: Verify no stale references**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && grep -r "GameMeta\|GameItem\|game_metas\|game_items" --include="*.go" | grep -v "_test.go" | grep -v "20260407000001"`
Expected: NO OUTPUT

- [ ] **Step 3: Verify no unused imports**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./... 2>&1 | grep "imported and not used"`
Expected: NO OUTPUT

- [ ] **Step 4: Verify frontend builds**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run build`
Expected: SUCCESS (no frontend changes, but verify nothing broke)

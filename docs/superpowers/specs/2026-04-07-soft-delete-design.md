# Soft Delete for Content Entities

Add Goravel-native soft delete to content authoring tables so that deleting games, levels, and content items no longer orphans user progress, game records, or tracking data.

## Motivation

Currently all delete operations hard-delete rows. Since the project uses code-level FK constraints (no DB-level cascades, to support PostgreSQL partitions), deleting a content item, game level, or game leaves orphaned references in 12+ dependent tables (game_records, game_sessions, user_masters, user_unknowns, user_reviews, user_favorites, game_reports, game_pks, etc.).

With Goravel's built-in `orm.SoftDeletes`, deleted rows get a `deleted_at` timestamp instead of being removed. All ORM queries automatically filter them out. Historical/tracking queries use `WithTrashed()` to still resolve references.

## Tables Getting Soft Delete

| Table | Deletion Pattern |
|---|---|
| `games` | Creator soft-deletes |
| `game_levels` | Creator soft-deletes, or cascade from game delete |
| `content_metas` | Cascade from level/game delete, or orphan cleanup |
| `content_items` | Creator soft-deletes single item, or cascade/orphan cleanup |
| `game_metas` | Junction — cascade soft-delete with level/game |
| `game_items` | Junction — cascade soft-delete with level/game |

### Tables NOT Getting Soft Delete

| Table | Reason |
|---|---|
| `user_masters`, `user_unknowns`, `user_reviews`, `user_favorites` | User hard-deletes own entries, no cascade risk |
| `game_sessions`, `game_records`, `game_reports`, `game_pks` | Never user-deleted; future age-based archival (hard delete) |

## Model Changes

Add `orm.SoftDeletes` to 6 models. All already embed `orm.Timestamps`; the junction models (`GameMeta`, `GameItem`) need the goravel orm import added.

```go
// game.go
type Game struct {
    orm.Timestamps
    orm.SoftDeletes
    // ... existing fields unchanged
}

// game_level.go
type GameLevel struct {
    orm.Timestamps
    orm.SoftDeletes
    // ... existing fields unchanged
}

// content_meta.go
type ContentMeta struct {
    orm.Timestamps
    orm.SoftDeletes
    // ... existing fields unchanged
}

// content_item.go
type ContentItem struct {
    orm.Timestamps
    orm.SoftDeletes
    // ... existing fields unchanged
}

// game_meta.go — add goravel orm import
type GameMeta struct {
    orm.SoftDeletes
    ID            string    `gorm:"column:id;primaryKey" json:"id"`
    GameID        string    `gorm:"column:game_id" json:"game_id"`
    GameLevelID   string    `gorm:"column:game_level_id" json:"game_level_id"`
    ContentMetaID string    `gorm:"column:content_meta_id" json:"content_meta_id"`
    CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

// game_item.go — add goravel orm import
type GameItem struct {
    orm.SoftDeletes
    ID            string    `gorm:"column:id;primaryKey" json:"id"`
    GameID        string    `gorm:"column:game_id" json:"game_id"`
    GameLevelID   string    `gorm:"column:game_level_id" json:"game_level_id"`
    ContentItemID string    `gorm:"column:content_item_id" json:"content_item_id"`
    CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}
```

## Migration Changes

Add `table.SoftDeletesTz()` directly to the existing CREATE TABLE migrations for all 6 tables. No new migration file — DB will be reset fresh.

## Delete Function Changes

### DeleteContentItem (course_content_service.go)

Wrap in transaction (fixes existing bug). ORM `.Delete()` auto-soft-deletes. Orphan check `.Count()` auto-excludes soft-deleted junction rows.

```go
func DeleteContentItem(userID, gameID, itemID string) error {
    // ... existing guards unchanged ...

    return facades.Orm().Transaction(func(tx orm.Query) error {
        // Soft-deletes junction row (automatic via orm.SoftDeletes)
        if _, err := tx.
            Where("content_item_id", itemID).Where("game_id", gameID).
            Delete(&models.GameItem{}); err != nil {
            return fmt.Errorf("failed to delete game item: %w", err)
        }

        // Count() auto-excludes soft-deleted junction rows
        remaining, _ := tx.Model(&models.GameItem{}).
            Where("content_item_id", itemID).Count()
        if remaining == 0 {
            // Soft-deletes content item (automatic)
            if _, err := tx.Where("id", itemID).Delete(&models.ContentItem{}); err != nil {
                return fmt.Errorf("failed to delete content item: %w", err)
            }
        }
        return nil
    })
}
```

### DeleteAllLevelContent (course_content_service.go)

ORM `.Delete()` calls unchanged (auto-soft-delete). Raw SQL orphan cleanup changes from `DELETE FROM` to `UPDATE SET deleted_at`:

```sql
-- Before
DELETE FROM content_items WHERE id NOT IN (SELECT content_item_id FROM game_items)

-- After
UPDATE content_items SET deleted_at = NOW() WHERE deleted_at IS NULL AND id NOT IN (SELECT content_item_id FROM game_items WHERE deleted_at IS NULL)

-- Same pattern for content_metas
UPDATE content_metas SET deleted_at = NOW() WHERE deleted_at IS NULL AND id NOT IN (SELECT content_meta_id FROM game_metas WHERE deleted_at IS NULL)
```

### DeleteLevel (course_game_service.go)

Same raw SQL changes as DeleteAllLevelContent. ORM calls unchanged.

### DeleteGame (course_game_service.go)

Same raw SQL changes. ORM calls for deleting junctions, levels, and game record unchanged (auto-soft-delete).

## JOIN Filter Additions

Goravel auto-filters `deleted_at IS NULL` only on the primary model. Raw JOINs to junction tables need explicit filtering. Add `AND gi.deleted_at IS NULL` or `AND gm.deleted_at IS NULL` to each JOIN clause.

```go
// Before
Join("JOIN game_items gi ON gi.content_item_id = content_items.id")

// After
Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL")
```

### All 14 JOINs

| # | File | Function | Junction |
|---|---|---|---|
| 1 | content_service.go:41 | GetLevelContent() | game_items |
| 2 | course_game_service.go:308 | PublishGame() item count | game_items |
| 3 | course_game_service.go:320 | PublishGame() ungenerated count | game_items |
| 4 | course_game_service.go:516 | GetCourseGameDetail() level item count | game_items |
| 5 | course_content_service.go:77 | SaveMetadataBatch() capacity check | game_metas |
| 6 | course_content_service.go:201 | GetContentItemsByMeta() metas query | game_metas |
| 7 | course_content_service.go:215 | GetContentItemsByMeta() items query | game_items |
| 8 | course_content_service.go:309 | InsertContentItem() item count | game_items |
| 9 | course_content_service.go:531 | calculateInsertionOrder() last item | game_items |
| 10 | course_content_service.go:551 | calculateInsertionOrder() all items | game_items |
| 11 | game_play_single_service.go:606 | countLevelItems() | game_items |
| 12 | game_play_pk_service.go:595 | spawnRobotForLevel() | game_items |
| 13 | ai_custom_service.go:326 | BreakMetadata() | game_metas |
| 14 | ai_custom_service.go:572 | GenerateContentItems() | game_metas |

## WithTrashed() Additions

Enrichment queries that load soft-deleted entities by ID for historical/tracking display.

```go
// Before
facades.Orm().Query().Where("id IN ?", ids).Get(&items)

// After
facades.Orm().Query().WithTrashed().Where("id IN ?", ids).Get(&items)
```

| # | File | Function | Model | Used By |
|---|---|---|---|---|
| 1 | user_master_service.go:105 | batchLoadContentItems() | ContentItem | mastered, unknown, reviews (shared) |
| 2 | user_master_service.go:123 | batchLoadGameNames() | Game | mastered, unknown, reviews (shared) |
| 3 | favorite_service.go:76 | ListFavorites() | Game | favorites list |

### What Does NOT Need WithTrashed()

- **Gameplay queries** (GetLevelContent, countLevelItems, spawnRobotForLevel) — published games can't have soft-deleted content (published guard prevents deletion)
- **Authoring queries** — soft-deleted content should stay hidden from the creator
- **GetPlayedGames()** — already filters by `status = published`, which excludes withdrawn/soft-deleted games

## Files NOT Changed

- All frontend code (dx-web) — API contract unchanged
- Docker/deploy configuration — unchanged
- Game session, record, report services — no soft delete on those tables
- User tracking services logic (mark/unmark) — no soft delete on those tables
- Published game guard — still blocks deletion of published games

## Safety

- All existing ORM queries automatically filter soft-deleted rows (zero manual WHERE clauses)
- Published game guard unchanged — content can only be soft-deleted from unpublished/withdrawn games
- Transaction wrapping on DeleteContentItem fixes existing race condition bug
- Soft delete is reversible via `Restore()` if needed in future
- No data loss — all historical references still resolve via `WithTrashed()`

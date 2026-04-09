# Remove Junction Tables — Restore game_level_id to Content Tables

**Date**: 2026-04-09
**Status**: Approved

## Context

The project previously moved to junction tables (`game_metas`, `game_items`) to link content to game levels. After careful consideration, the decision is to revert: put `game_level_id` back directly on `content_metas` and `content_items`, and remove the junction tables entirely. This simplifies queries, eliminates orphan cleanup logic, and aligns with the PostgreSQL partitioning strategy (code-level FK constraints, no DB-level FKs).

The DB migrations already have `game_level_id` on both content tables. The existing data has already been handled correctly with `game_level_id` populated. An `is_selective` field has also been added to the `games` table migration.

## Scope

### Models (5 file changes)

1. **`models/content_meta.go`** — Add `GameLevelID string` field
2. **`models/content_item.go`** — Add `GameLevelID string` field
3. **`models/game.go`** — Add `IsSelective bool` field
4. **`models/game_meta.go`** — Delete file
5. **`models/game_item.go`** — Delete file

### Backend Services (8 files)

#### 1. `content_service.go` — `GetLevelContent()`

**Before**:
```go
Join("JOIN game_items gi ON gi.content_item_id = content_items.id AND gi.deleted_at IS NULL").
Where("gi.game_level_id", gameLevelID)
```

**After**:
```go
Where("content_items.game_level_id", gameLevelID)
```

#### 2. `course_content_service.go` — 8 functions

**`SaveMetadataBatch()`**:
- Set `GameLevelID: gameLevelID` on ContentMeta directly
- Remove GameMeta creation entirely

**`GetContentItemsByMeta()`**:
- Replace `JOIN game_metas gm ...` with `WHERE content_metas.game_level_id = ?`
- Replace `JOIN game_items gi ...` with `WHERE content_items.game_level_id = ?`

**`InsertContentItem()`**:
- Set `GameLevelID: gameLevelID` on ContentItem directly
- Remove GameItem creation entirely
- Replace game_items JOIN in item count query with `WHERE game_level_id = ?`

**`DeleteContentItem()`**:
- Remove GameItem soft-delete step
- Simplify to direct soft-delete of content_item
- is_break_done reset: check remaining items via `WHERE content_meta_id = content_metas.id AND game_level_id = (SELECT game_level_id FROM content_items WHERE id = ?) AND deleted_at IS NULL`

**`DeleteAllLevelContent()`**:
- Remove junction table soft-deletes
- Direct: `UPDATE content_items SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL`
- Direct: `UPDATE content_metas SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL`

**`DeleteMetadata()`**:
- Remove game_items and game_metas soft-deletes
- Direct soft-delete content_items by `content_meta_id`
- Direct soft-delete content_meta by ID

**`verifyMetaBelongsToGame()`**:
- Query: `SELECT id FROM content_metas WHERE id = ? AND game_level_id IN (SELECT id FROM game_levels WHERE game_id = ? AND deleted_at IS NULL) AND deleted_at IS NULL`

**`verifyItemBelongsToGame()`**:
- Query: `SELECT id FROM content_items WHERE id = ? AND game_level_id IN (SELECT id FROM game_levels WHERE game_id = ? AND deleted_at IS NULL) AND deleted_at IS NULL`

**`calculateInsertionOrder()`**:
- Replace game_items JOIN with `WHERE content_items.game_level_id = ?`

#### 3. `course_game_service.go` — 4 functions

**`DeleteGame()`**:
- Remove GameItem/GameMeta soft-deletes
- Collect level IDs, then: `UPDATE content_items SET deleted_at = NOW() WHERE game_level_id IN (?) AND deleted_at IS NULL`
- Same for content_metas

**`DeleteLevel()`**:
- Remove GameItem/GameMeta soft-deletes
- Direct: `UPDATE content_items SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL`
- Direct: `UPDATE content_metas SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL`

**`PublishGame()`**:
- Replace game_items JOINs with `WHERE content_items.game_level_id = ?`

**`GetCourseGameDetail()`**:
- Replace game_items JOIN for item count with `WHERE game_level_id = ?`

#### 4. `ai_custom_service.go` — 3 functions

**`BreakMetadata()`**:
- Replace game_metas JOIN with `WHERE content_metas.game_level_id = ?`

**`processBreakMeta()`**:
- Set `GameLevelID: gameLevelID` on ContentItem
- Remove GameItem creation

**`GenerateContentItems()`**:
- Replace game_metas JOIN with `WHERE content_metas.game_level_id = ?`

#### 5. `ai_custom_vocab_service.go` — 3 functions

**`BreakVocabMetadata()`**:
- Replace game_metas JOIN with `WHERE content_metas.game_level_id = ?`

**`processVocabBreakMeta()`**:
- Set `GameLevelID: gameLevelID` on ContentItem
- Remove GameItem creation

**`GenerateVocabContentItems()`**:
- Replace game_metas JOIN with `WHERE content_metas.game_level_id = ?`

#### 6. `game_play_single_service.go` — 1 function

**`countLevelItems()`**:
- Replace game_items JOIN with `WHERE content_items.game_level_id = ?`
- This function is shared by PK and group play services

#### 7. `game_play_pk_service.go` — 1 function

**`spawnRobotForLevel()`**:
- Replace game_items JOIN with `WHERE content_items.game_level_id = ?` for robot content fetch

#### 8. `import_courses.go` — 3 functions

**`insertLevels()`**:
- Set `GameLevelID: levelID` on each ContentItem
- Remove `createGameItemsBatch()` calls

**`createGameItemsBatch()`**:
- Delete this function entirely

**`forceCleanup()`**:
- Remove GameItem/GameMeta deletion
- Replace orphan cleanup with direct: `UPDATE content_items SET deleted_at = NOW() WHERE game_level_id IN (SELECT id FROM game_levels WHERE game_id = ? AND deleted_at IS NOT NULL) AND deleted_at IS NULL`
- Same pattern for content_metas

### Frontend Changes

**None.** API response shapes are unchanged. The frontend passes `gameLevelId` as parameters and receives content items — it never interacts with junction tables.

### Migration File

The junction tables migration (`20260407000001_create_game_junction_tables.go`) should be kept but made a no-op (tables may still exist in DB but are unused). No new migration needed since `game_level_id` is already in the content table migrations.

## Correctness Guarantees

1. **Every read query** that used `JOIN game_items/game_metas` is replaced with direct `WHERE game_level_id = ?`
2. **Every write operation** that created GameItem/GameMeta records now sets `GameLevelID` on the content record directly
3. **Every delete operation** that cascaded through junction tables now directly soft-deletes content by `game_level_id`
4. **The `countLevelItems()` function** (shared by single play, PK, and group play) is updated once, fixing all three workflows
5. **SSE workflows** (break metadata, generate items) use the same simplified queries
6. **No API response shapes change** — frontend continues to work without modification
7. **Build verification**: `go build ./...` and `go vet ./...` must pass with zero errors after all changes

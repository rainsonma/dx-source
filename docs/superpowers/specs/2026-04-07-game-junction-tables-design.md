# Game Junction Tables Design

Decouple content from games so the same `content_metas` and `content_items` can be shared across multiple game modes (word-sentence, vocab-battle, vocab-match, vocab-elimination).

## Motivation

Currently `content_metas` and `content_items` are directly bound to a specific `game_level_id`. If two games want the same English content, each must have its own copy. By introducing junction tables, we can create one game, populate its content, then later add another game mode that shares the same content rows.

## Approach

Add two junction tables (`game_metas`, `game_items`) that bridge games/levels to content. Change all content queries and writes to go through the junction tables. Migrate existing data. Do not break any existing functionality.

## Schema

### New Tables

```sql
CREATE TABLE game_metas (
    id            TEXT PRIMARY KEY,
    game_id       TEXT NOT NULL,
    game_level_id TEXT NOT NULL,
    content_meta_id TEXT NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT now()
);
CREATE UNIQUE INDEX idx_game_metas_unique ON game_metas (game_id, game_level_id, content_meta_id);
CREATE INDEX idx_game_metas_game_level ON game_metas (game_level_id);
CREATE INDEX idx_game_metas_content_meta ON game_metas (content_meta_id);

CREATE TABLE game_items (
    id            TEXT PRIMARY KEY,
    game_id       TEXT NOT NULL,
    game_level_id TEXT NOT NULL,
    content_item_id TEXT NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT now()
);
CREATE UNIQUE INDEX idx_game_items_unique ON game_items (game_id, game_level_id, content_item_id);
CREATE INDEX idx_game_items_game_level ON game_items (game_level_id);
CREATE INDEX idx_game_items_content_item ON game_items (content_item_id);
```

### Existing Table Alterations

```sql
ALTER TABLE content_metas ALTER COLUMN game_level_id DROP NOT NULL;
ALTER TABLE content_items ALTER COLUMN game_level_id DROP NOT NULL;
```

`game_level_id` stays on both tables but becomes nullable. New writes set it to NULL. Existing data retains its value. The column will be dropped entirely after migration verification.

### Data Migration

Populate junction tables from existing data in the same migration:

```sql
INSERT INTO game_metas (id, game_id, game_level_id, content_meta_id, created_at)
SELECT gen_random_uuid(), gl.game_id, cm.game_level_id, cm.id, cm.created_at
FROM content_metas cm
JOIN game_levels gl ON gl.id = cm.game_level_id;

INSERT INTO game_items (id, game_id, game_level_id, content_item_id, created_at)
SELECT gen_random_uuid(), gl.game_id, ci.game_level_id, ci.id, ci.created_at
FROM content_items ci
JOIN game_levels gl ON gl.id = ci.game_level_id;
```

## New Go Models

```go
// app/models/game_meta.go
type GameMeta struct {
    ID            string    `gorm:"column:id;primaryKey" json:"id"`
    GameID        string    `gorm:"column:game_id" json:"game_id"`
    GameLevelID   string    `gorm:"column:game_level_id" json:"game_level_id"`
    ContentMetaID string    `gorm:"column:content_meta_id" json:"content_meta_id"`
    CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

// app/models/game_item.go
type GameItem struct {
    ID            string    `gorm:"column:id;primaryKey" json:"id"`
    GameID        string    `gorm:"column:game_id" json:"game_id"`
    GameLevelID   string    `gorm:"column:game_level_id" json:"game_level_id"`
    ContentItemID string    `gorm:"column:content_item_id" json:"content_item_id"`
    CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}
```

### Existing Model Changes

```go
// content_meta.go — GameLevelID becomes nullable
GameLevelID *string `gorm:"column:game_level_id" json:"game_level_id"`

// content_item.go — GameLevelID becomes nullable
GameLevelID *string `gorm:"column:game_level_id" json:"game_level_id"`
```

## Query Changes

All content queries switch from direct `game_level_id` filtering to JOINing through junction tables. Order stays on `content_metas` and `content_items`.

### Read Patterns

| Location | Before | After |
|----------|--------|-------|
| `content_service.go` `GetLevelContent()` | `WHERE game_level_id = ?` | `JOIN game_items gi ON gi.content_item_id = content_items.id WHERE gi.game_level_id = ?` |
| `game_play_single_service.go` `countLevelItems()` | `WHERE game_level_id = ?` | Same JOIN pattern |
| `course_content_service.go` `GetContentItemsByMeta()` | `WHERE game_level_id = ?` on both metas and items | JOIN through `game_metas` / `game_items` |
| `course_content_service.go` `calculateInsertionOrder()` | `WHERE game_level_id = ?` | JOIN through `game_items` |
| `course_content_service.go` `verifyMetaBelongsToGame()` | Two queries: meta -> level -> game | Single query: `game_metas WHERE content_meta_id = ? AND game_id = ?` |
| `course_content_service.go` `verifyItemBelongsToGame()` | Two queries: item -> level -> game | Single query: `game_items WHERE content_item_id = ? AND game_id = ?` |

### Write Patterns

| Location | Before | After |
|----------|--------|-------|
| `course_content_service.go` `SaveMetadataBatch()` | Create `content_meta` with `game_level_id` | Create `content_meta` (no `game_level_id`) + create `game_metas` row |
| `ai_custom_service.go` `processBreakMeta()` | Create `content_item` with `game_level_id` | Create `content_item` (no `game_level_id`) + create `game_items` row |
| `course_content_service.go` `InsertContentItem()` | Create `content_item` with `game_level_id` | Create `content_item` (no `game_level_id`) + create `game_items` row |

### Delete Patterns

| Location | Before | After |
|----------|--------|-------|
| `course_content_service.go` `DeleteContentItem()` | Delete `content_item` by id | Delete `game_items` row + delete `content_item` |
| `course_content_service.go` `DeleteAllLevelContent()` | Delete items then metas by `game_level_id` | Delete `game_items` + `game_metas` by level, then delete orphaned content rows |

## Function Signature Changes

`processBreakMeta` and `processGenItems` in `ai_custom_service.go` currently don't receive `gameID`. The parameter needs to be threaded through so junction rows can be created. No logic change, just signature change.

## Files NOT Changed

- `user_master_service.go` / `user_unknown_service.go` / `user_review_service.go` — query by `content_item_id` directly
- `game_records` — references `content_item_id` directly
- `game_sessions` — references `current_content_item_id` directly
- All frontend code (dx-web) — API request/response shapes are unchanged
- AI prompts and processing logic in `ai_custom_service.go` — unchanged
- Docker/deploy configuration — unchanged

## Safety

- Work on a feature branch (`feat/game-junction-tables`)
- Existing `game_level_id` columns kept (nullable) until post-migration verification
- Data migration is additive (no data deleted or modified)
- Unique indexes on junction tables prevent duplicate rows
- All affected queries are mechanical transformations (JOIN pattern)
- No lint issues introduced

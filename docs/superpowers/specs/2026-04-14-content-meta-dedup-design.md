---
title: Content Meta & Item Deduplication on Save
date: 2026-04-14
status: approved
related:
  - dx-api/app/services/api/course_content_service.go
  - dx-api/database/migrations/20260414000001_create_game_metas_and_game_items_tables.go
  - dx-api/database/migrations/20260414000003_drop_legacy_columns_from_content_tables.go
---

# Content Meta & Item Deduplication on Save

## Goal

When a user adds new metadata to a level on the AI-Custom page, the backend must:

1. Reuse an existing `content_metas` row when the user already owns an identical one (saves DB space, avoids redundant rows).
2. Reuse the existing `content_items` rows of any reused meta that has already been broken down — so the new level inherits the breakdown for free.
3. Always create new `game_metas` / `game_items` junction rows linking the (existing or new) underlying content into the target level — including allowing a meta or item to repeat within the same level if the user wants it.

This is purely a server-side change. The save endpoint contract, frontend dialogs, and server actions remain unchanged.

## Background

The recent refactor on `feat/game-junction-tables` introduced two junction tables:

- `game_metas (game_id, game_level_id, content_meta_id, order)` — links a level to a meta
- `game_items (game_id, game_level_id, content_item_id, order)` — links a level to a content item

Both junctions were created with a partial unique index `(game_level_id, content_*_id) WHERE deleted_at IS NULL` — a holdover from the pre-junction 1:1 era. We are now opening the design up so the same underlying meta/item can appear (a) in multiple levels and (b) repeated within a single level.

## Decisions (from brainstorming)

| Question | Decision |
|---|---|
| Match identity for dedup | `(source_type, source_data, translation)` — `source_from` is NOT part of identity |
| Empty translation handling | NULL ≡ empty string; otherwise exact match |
| Dedup search scope | Per-user only — find candidates among `content_metas` reachable via the current user's own games |
| In-level repetition | Allowed — drop the unique partial indexes on both junctions, replace with non-unique partial indexes |
| Delete strategy | Reference-counted soft delete: junction rows are deleted unconditionally for the current scope; underlying rows are only soft-deleted when no live junction rows remain anywhere |
| Items reuse when meta is already broken down | Reuse content_items via fresh `game_items` junction rows — no row duplication |
| Frontend changes | None — endpoint contract and response shape are unchanged |

## Architecture

### Schema migration

One new migration: `20260415000001_relax_junction_unique_indexes_and_add_dedup_index.go`

`Up()`:

```sql
-- game_metas: drop unique, create non-unique
DROP INDEX IF EXISTS idx_game_metas_level_meta_unique;
CREATE INDEX idx_game_metas_level_meta
  ON game_metas (game_level_id, content_meta_id)
  WHERE deleted_at IS NULL;

-- game_items: drop unique, create non-unique
DROP INDEX IF EXISTS idx_game_items_level_item_unique;
CREATE INDEX idx_game_items_level_item
  ON game_items (game_level_id, content_item_id)
  WHERE deleted_at IS NULL;

-- speed up dedup lookups on content_metas
CREATE INDEX IF NOT EXISTS idx_content_metas_dedup_lookup
  ON content_metas (source_type, source_data)
  WHERE deleted_at IS NULL;
```

`Down()` reverses: re-creates the two unique partial indexes and drops `idx_content_metas_dedup_lookup`.

`Down()` will fail if the table has multiple `game_metas` (or `game_items`) rows that share the same `(level, content_*_id)` after the new code starts producing intentional duplicates. That's expected — `Down()` is for migration rollback during development, not a permanent escape hatch.

### Save flow — `SaveMetadataBatch`

File: `dx-api/app/services/api/course_content_service.go`

The function keeps the same signature:

```go
func SaveMetadataBatch(userID, gameID, gameLevelID string, entries []MetadataEntry, sourceFrom string) (int, error)
```

Lines 137-163 (the create loop) are replaced. The rest of the function (VIP check, ownership, capacity validation, max order computation) stays identical.

**Step 1 — build dedup keys for the batch.**

```go
type metaKey struct {
    SourceType string
    SourceData string
    Translation string // normalized: NULL -> ""
}

func makeKey(e MetadataEntry) metaKey {
    t := ""
    if e.Translation != nil {
        t = *e.Translation
    }
    return metaKey{e.SourceType, e.SourceData, t}
}
```

**Step 2 — query existing user-owned candidates.**

```go
type existingMetaRow struct {
    ID          string
    SourceType  string
    SourceData  string
    Translation string  // normalized to "" via COALESCE
    IsBreakDone bool
}

// Distinct (source_type, source_data) pairs from the batch.
sourceTypes, sourceData := distinctTypesAndData(entries)

var rows []existingMetaRow
err := facades.Orm().Query().Raw(
    `SELECT DISTINCT cm.id, cm.source_type, cm.source_data,
            COALESCE(cm.translation, '') AS translation, cm.is_break_done
       FROM content_metas cm
       JOIN game_metas gm ON gm.content_meta_id = cm.id AND gm.deleted_at IS NULL
       JOIN game_levels gl ON gl.id = gm.game_level_id AND gl.deleted_at IS NULL
       JOIN games g ON g.id = gl.game_id AND g.deleted_at IS NULL
      WHERE cm.deleted_at IS NULL
        AND g.user_id = ?
        AND cm.source_type IN ?
        AND cm.source_data IN ?`,
    userID, sourceTypes, sourceData,
).Scan(&rows)
```

Build `existingByKey map[metaKey]existingMetaRow` — first row wins per key.

**Step 3 — process each entry.**

Inside one transaction (`facades.Orm().Transaction(...)`). Note: the current implementation does NOT wrap its loop in a transaction, so a partial failure can leave orphaned `content_metas` rows. The new implementation closes that gap as a side effect — strictly an improvement, not a behavior regression.

```go
itemsByMetaCache := map[string][]string{}    // metaID -> []contentItemID
var maxItemOrder *float64                    // lazily loaded for the level
itemsAddedSoFar := 0

for i, e := range entries {
    key := makeKey(e)

    var metaID string
    if existing, ok := existingByKey[key]; ok {
        metaID = existing.ID
    } else {
        metaID = uuid.Must(uuid.NewV7()).String()
        meta := models.ContentMeta{
            ID:          metaID,
            SourceFrom:  sourceFrom,
            SourceType:  e.SourceType,
            SourceData:  e.SourceData,
            Translation: e.Translation,
            IsBreakDone: false,
        }
        if err := tx.Create(&meta); err != nil {
            return fmt.Errorf("failed to create content meta: %w", err)
        }
        // Add to map so subsequent within-batch identical entries reuse this row.
        existingByKey[key] = existingMetaRow{
            ID: metaID, SourceType: e.SourceType, SourceData: e.SourceData,
            Translation: keyTranslation(e), IsBreakDone: false,
        }
    }

    // Always create a junction row.
    gm := models.GameMeta{
        ID:            uuid.Must(uuid.NewV7()).String(),
        GameID:        level.GameID,
        GameLevelID:   gameLevelID,
        ContentMetaID: metaID,
        Order:         maxOrder + float64((i+1)*1000),
    }
    if err := tx.Create(&gm); err != nil {
        return fmt.Errorf("failed to create game meta: %w", err)
    }

    // If reusing a broken-down meta, also create game_items rows.
    if existing, ok := existingByKey[key]; ok && existing.IsBreakDone {
        if err := reuseItemsIntoLevel(tx, existing.ID, gameLevelID, level.GameID,
            itemsByMetaCache, &maxItemOrder, &itemsAddedSoFar); err != nil {
            return err
        }
    }
}
```

`reuseItemsIntoLevel`:
1. If `itemsByMetaCache[metaID]` is unset, query `SELECT id FROM content_items WHERE content_meta_id = ? AND deleted_at IS NULL ORDER BY id` and cache.
2. If `maxItemOrder` is unset, query `SELECT COALESCE(MAX("order"), 0) FROM game_items WHERE game_level_id = ? AND deleted_at IS NULL` and cache.
3. For each cached item ID, insert a `game_items` row with `order = *maxItemOrder + float64((*itemsAddedSoFar + j + 1) * 1000)`. Increment `*itemsAddedSoFar` after the loop.

**Step 4 — return `len(entries)`** (every input entry produced exactly one new junction row, regardless of dedup).

### Delete flow — reference-counted

Three functions in the same file are updated. All run inside `facades.Orm().Transaction(...)` and follow the same pattern: delete the junction(s) for the requested scope, then count remaining live junctions for the underlying row, and only soft-delete the underlying row when zero remain.

#### `DeleteContentItem(userID, gameID, gameLevelID, itemID)`

Add `gameLevelID` parameter. Update controller to plumb it from route params (route already includes `:levelId`).

```go
1. Verify item belongs to game (existing helper still works).
2. Soft-delete game_items WHERE content_item_id = ? AND game_level_id = ? AND deleted_at IS NULL
   (deletes ALL repetitions of this item in the level — parallel to DeleteMetadata).
3. Count live game_items WHERE content_item_id = ? AND deleted_at IS NULL
4. If 0 → soft-delete content_items WHERE id = ?
5. Reset is_break_done if this LEVEL has no remaining game_items for the meta (existing logic, unchanged).
```

#### `DeleteMetadata(userID, gameID, gameLevelID, metaID)`

Add `gameLevelID` parameter.

```go
1. Verify meta belongs to game.
2. Collect content_item_ids referenced by THIS level for THIS meta:
     SELECT gi.content_item_id
       FROM game_items gi
       JOIN content_items ci ON ci.id = gi.content_item_id AND ci.deleted_at IS NULL
      WHERE ci.content_meta_id = ?
        AND gi.game_level_id = ?
        AND gi.deleted_at IS NULL
3. Soft-delete the game_items rows for that level + meta.
4. Soft-delete game_metas WHERE content_meta_id = ? AND game_level_id = ? AND deleted_at IS NULL
   (deletes ALL repetitions of this meta in the level).
5. For each collected content_item_id: count live game_items; if 0 → soft-delete the content_item.
6. Count live game_metas for this content_meta_id across all levels; if 0 → soft-delete the content_meta.
```

Subtle UX behavior: if the user repeated a meta inside one level, "delete this meta from the level" removes all repetitions. The list-row UI doesn't currently expose per-repetition delete; if it ever does, we'll add a separate "delete by junction id" path.

#### `DeleteAllLevelContent(userID, gameID, gameLevelID)`

```go
1. Verify level belongs to game.
2. Collect distinct content_meta_ids and content_item_ids referenced by this level
   (BEFORE soft-deleting the junctions).
3. Soft-delete game_items WHERE game_level_id = ? AND deleted_at IS NULL
4. Soft-delete game_metas WHERE game_level_id = ? AND deleted_at IS NULL
5. For each collected content_item_id: count live game_items; if 0 → soft-delete.
6. For each collected content_meta_id: count live game_metas; if 0 → soft-delete.
```

Steps 5-6 are O(N) SELECTs in the worst case. `DeleteAllLevelContent` is rare and bulk; acceptable. Optimizable to a single `LEFT JOIN ... WHERE NOT EXISTS` if profiling shows it matters.

### `verifyMetaBelongsToGame` and `verifyItemBelongsToGame`

No changes. Both already use the junction and check "is there at least one row binding this content to this game?" — the answer remains correct under reuse.

## Frontend impact

**None.** The endpoint contract, request payload, and response shape are unchanged. `AddMetadataDialog`, `AddVocabDialog`, `saveMetadataAction`, and `parseMetadataText` need no edits. The level content refresh after save (SWR re-fetch) automatically renders the deduped state, including any pre-existing breakdown items inherited via reuse.

The two delete controllers (`DeleteContentItem`, `DeleteMetadata`) need to plumb `gameLevelID` from their route params (already in the URL) into the service call. No frontend payload changes; existing route shapes are preserved.

## Data safety

This change touches the delete path, which is data-destruction code. Required precautions:

1. **Pre-migration backup.** Before running `go run . migrate` in any environment that has live data, take a `pg_dump` snapshot to `/Users/rainsen/Programs/Projects/douxue/db-backup/dx-YYYYMMDD-HHMMSS.sql.gz`. The implementation plan must include this as an explicit step ahead of the migration. Verify the dump is non-empty before proceeding.
2. **Migration is index-only.** No row data is touched; no column types change. Worst-case rollback is `down` then restore from the backup if needed.
3. **Save changes are additive.** Dedup logic only avoids creating new rows; it never modifies or deletes existing data.
4. **Delete changes are strictly more conservative than today.** The new logic preserves underlying rows that the old logic would have cascade-deleted. Any bug in reference counting that errs on the cautious side leaves orphans (cleanable later); a bug that errs on the destructive side could prematurely soft-delete shared content. The test suite (Section 5.2 below) exercises both directions explicitly to catch this.
5. **Soft-delete only.** All deletion paths set `deleted_at = NOW()` rather than hard-deleting. A buggy reference count that soft-deletes too aggressively can be recovered by clearing `deleted_at`.

## Testing surface

Go test file: `dx-api/tests/feature/course_content_dedup_test.go`

### Save / dedup

Note: each test starts from a clean DB state for the user (no pre-existing dedup matches in the user's other games unless the test setup explicitly seeds them).

1. Fresh entries (no dedup target) → all created; junction count == entries count.
2. Saving identical entries to a second game by the same user reuses `content_metas` from the first game (verify by row counts).
3. Within-batch repetition: two identical entries → one new `content_metas`, two new `game_metas`.
4. Translation matching: NULL ↔ "" treated as equivalent; otherwise exact match.
5. `source_type` differentiation: same `source_data` with different `source_type` does NOT dedup.
6. Reusing a meta with `is_break_done = true` also creates `game_items` rows in the new level for all the meta's `content_items`, with monotonically increasing order appended after the level's pre-save max.
7. Cross-user isolation: User A's content is NOT visible to User B's dedup query.
8. Capacity check still enforces existing limits — entries count toward capacity even when deduped.
9. Word-sentence ratio rule still passes/fails correctly with deduped entries.

### Delete / reference counting
10. `DeleteMetadata` from level L1 leaves a duplicate-reused meta in level L2 intact.
11. `DeleteMetadata` from the only level using a meta DOES soft-delete the underlying `content_metas` and its items.
12. `DeleteContentItem` leaves a shared content_item alive when other levels still reference it.
13. `DeleteAllLevelContent` correctly reference-counts both metas and items.
14. `is_break_done` reset still works correctly after delete (existing behavior preserved).

### Migration smoke
15. After running `Up()`, the unique indexes are gone and inserting two `game_metas` with the same `(game_level_id, content_meta_id)` succeeds. After `Down()` (with no duplicates present), the unique indexes return.

## Performance considerations

- Dedup query is a single SELECT per save, scoped to the user via the join chain. With `idx_content_metas_dedup_lookup` and the existing junction indexes, lookup is fast even for users with thousands of metas.
- Save loop is O(N) inserts (same as today). Items reuse query is amortized via `itemsByMetaCache`, so O(1) extra queries per distinct reused meta in the batch.
- Delete reference counts are O(N) extra COUNT queries in the bulk path. Acceptable for `DeleteAllLevelContent`; cheap for the per-row paths.

## Risk register

| Risk | Mitigation |
|---|---|
| Migration drops unique index but new code isn't deployed yet | Migration is purely additive at runtime — old code keeps working unchanged after the unique index is dropped (it never relied on the constraint to enforce semantics). |
| Reference counting bug deletes shared content | Test cases 10-14 specifically exercise both directions; all changes are soft-deletes so recoverable. |
| Concurrent saves create duplicate `content_metas` rows | Accepted as rare; no unique constraint is added to enforce. Future saves dedup against the first row. |
| Dropping `Down()` may fail under intentional duplicates | Documented: `Down()` is for development rollback, not production. |
| Existing data with the old unique constraint already in place | No conflict — the old data is by definition compliant with the unique constraint, so dropping the constraint is non-destructive. |

## Out of scope

- Per-repetition delete (deleting one repetition of a meta within a level by junction ID) — not currently exposed in the UI.
- Cross-user content sharing — explicitly rejected; dedup is per-user.
- Adding a unique partial index on `content_metas (source_type, source_data, COALESCE(translation, ''))` to harden against concurrent races — could be a follow-up if it becomes a real problem.
- Backfill / cleanup of any pre-existing duplicate `content_metas` rows that were created before this change. Not needed: from this point forward, new saves will dedup; old data is left as-is.

---
title: Squash Legacy Content Migrations
date: 2026-04-15
status: approved
related:
  - dx-api/database/migrations/20260322000036_create_content_metas_table.go
  - dx-api/database/migrations/20260322000037_create_content_items_table.go
  - dx-api/database/migrations/20260407000001_create_game_junction_tables.go
  - dx-api/database/migrations/20260414000002_backfill_junction_tables.go
  - dx-api/database/migrations/20260414000003_drop_legacy_columns_from_content_tables.go
  - dx-api/bootstrap/migrations.go
---

# Squash Legacy Content Migrations

## Purpose

Collapse `content_metas` / `content_items` to their final junction-era shape by removing the two transition migrations (backfill + drop-legacy-columns) and the orphaned no-op stub from an earlier junction-tables attempt. With the junction-tables feature merged to `main` and a post-junction backup available for re-import, the transition migrations no longer serve any purpose — they only existed to move data from the legacy columns into the new junction tables. Resetting the local DB and restoring the backup produces the final schema directly, without the transition chain.

## Background

The junction-tables refactor (`feat/game-junction-tables`, merged via commit `4f9f60e`) was designed to be applied to a live database holding ~1.22M `content_items` rows. That required three sequential migrations:

1. `20260414000001_create_game_metas_and_game_items_tables` — create the new junction tables (still needed).
2. `20260414000002_backfill_junction_tables` — copy `content_items.game_level_id/order` and `content_metas.game_level_id/order` into `game_items`/`game_metas`.
3. `20260414000003_drop_legacy_columns_from_content_tables` — drop `game_level_id`, `order`, `is_active` from the content tables.

A fourth file, `20260407000001_create_game_junction_tables.go`, is a no-op stub left over from an earlier junction-tables attempt that was reverted on Apr 9, 2026 (see `2026-04-14-reintroduce-junction-tables-for-content-reuse-design.md`, "Background"). It was kept because some DBs had already recorded its original signature and changing or removing the file would confuse the migration tracker on those environments.

The current operation (fresh DB + backup restore) means:

- No environment has any of these legacy signatures recorded (`migrate:fresh` wipes the tracking table).
- The backup was taken *after* the junction feature was in place, so the dump already contains `game_metas`/`game_items` rows and has no legacy columns on the content tables.
- Running 414002 against the clean schema would fail (it selects from columns that no longer exist).
- Running 414003 against the clean schema would be a no-op (nothing to drop).

So the transition chain is not just unnecessary — it is actively broken in the fresh-DB path. It must go.

## Goals

- Delete the two transition migrations (`20260414000002_backfill_junction_tables.go`, `20260414000003_drop_legacy_columns_from_content_tables.go`) that only move legacy data into junction tables and then drop the legacy columns.
- Delete the no-op stub `20260407000001_create_game_junction_tables.go` that survives from an earlier, reverted junction-tables attempt.
- Collapse `20260322000036_create_content_metas_table.go` and `20260322000037_create_content_items_table.go` to their final shape (the shape the models already reflect).
- Unregister the three deleted migrations from `bootstrap/migrations.go` so the build stays green.
- Leave `20260414000001_create_game_metas_and_game_items_tables.go` untouched — it is still the sole creator of the junction tables.
- Zero application-code changes. The models (`content_meta.go`, `content_item.go`) already match the target shape.

## Non-Goals

- Changing junction-tables runtime behavior.
- Touching models, services, controllers, or the frontend.
- Preserving any legacy migration history for downstream environments (none exist — the user's local DB is being reset and there is no staging/prod running the old chain).

## Current State

### `20260322000036_create_content_metas_table.go`

```
id (PK), game_level_id, source_from, source_type, source_data,
translation, is_break_done, order, timestamps, soft-deletes
Indexes: source_from, source_type, order, created_at
```

### `20260322000037_create_content_items_table.go`

```
id (PK), game_level_id, content_meta_id (nullable), content, content_type,
uk_audio_id, us_audio_id, definition, translation, explanation,
items (json), structure (json), order, tags (text[]), is_active,
timestamps, soft-deletes
Indexes: content_meta_id, uk_audio_id, us_audio_id, content_type,
         order, is_active, created_at
```

### `20260407000001_create_game_junction_tables.go`

No-op. Both `Up()` and `Down()` return `nil`.

### `20260414000002_backfill_junction_tables.go`

Two `INSERT … SELECT … ON CONFLICT DO NOTHING` statements copying `content_*` → `game_*` via joins to `game_levels`.

### `20260414000003_drop_legacy_columns_from_content_tables.go`

Five `ALTER TABLE … DROP COLUMN IF EXISTS …` statements.

### `bootstrap/migrations.go`

Registers all three deleted files in the `Migrations()` slice, alongside the rest.

### Model files (already in final shape — no edits needed)

- `app/models/content_meta.go`: no `GameLevelID`, no `Order` field.
- `app/models/content_item.go`: no `GameLevelID`, no `Order`, no `IsActive` field.

## Changes

### 1. Delete three migration files

- `dx-api/database/migrations/20260407000001_create_game_junction_tables.go`
- `dx-api/database/migrations/20260414000002_backfill_junction_tables.go`
- `dx-api/database/migrations/20260414000003_drop_legacy_columns_from_content_tables.go`

### 2. Edit `20260322000036_create_content_metas_table.go`

Remove the following lines from the `Up()` blueprint:

- `table.Uuid("game_level_id")`
- `table.Double("order").Default(0)`
- `table.Index("order")`

Final column set: `id` (PK), `source_from`, `source_type`, `source_data`, `translation` (nullable), `is_break_done`, `created_at`, `updated_at`, `deleted_at`.
Final index set: `source_from`, `source_type`, `created_at`.

### 3. Edit `20260322000037_create_content_items_table.go`

Remove the following lines from the `Up()` blueprint:

- `table.Uuid("game_level_id")`
- `table.Double("order").Default(0)`
- `table.Boolean("is_active").Default(true)`
- `table.Index("order")`
- `table.Index("is_active")`

Final column set: `id` (PK), `content_meta_id` (nullable), `content`, `content_type`, `uk_audio_id` (nullable), `us_audio_id` (nullable), `definition`, `translation`, `explanation`, `items` (jsonb), `structure` (jsonb), `tags` (text[]), `created_at`, `updated_at`, `deleted_at`.
Final index set: `content_meta_id`, `uk_audio_id`, `us_audio_id`, `content_type`, `created_at`.

### 4. Edit `bootstrap/migrations.go`

Remove these three lines from the `Migrations()` slice:

- `&migrations.M20260407000001CreateGameJunctionTables{},`
- `&migrations.M20260414000002BackfillJunctionTables{},`
- `&migrations.M20260414000003DropLegacyColumnsFromContentTables{},`

Leave `&migrations.M20260414000001CreateGameMetasAndGameItemsTables{},` in place — the junction tables are still created there.

### Files touched

| File | Change |
|---|---|
| `dx-api/database/migrations/20260322000036_create_content_metas_table.go` | edit |
| `dx-api/database/migrations/20260322000037_create_content_items_table.go` | edit |
| `dx-api/database/migrations/20260407000001_create_game_junction_tables.go` | delete |
| `dx-api/database/migrations/20260414000002_backfill_junction_tables.go` | delete |
| `dx-api/database/migrations/20260414000003_drop_legacy_columns_from_content_tables.go` | delete |
| `dx-api/bootstrap/migrations.go` | edit |

Total: 3 deletes, 3 edits. Zero application-code changes. Frontend unaffected.

## Verification

1. **Build is clean:** `go build ./...` inside `dx-api` — catches any leftover references to the deleted migration structs (there shouldn't be any outside `bootstrap/migrations.go`).
2. **Vet is clean:** `go vet ./...`.
3. **Fresh migrate:** drop the local DB, create an empty DB, run the server (migrations run on startup in Goravel). Expected: no errors. Only the remaining migrations execute.
4. **Schema check:**
   - `\d content_metas` → no `game_level_id`, no `order` column; no `order` index.
   - `\d content_items` → no `game_level_id`, no `order`, no `is_active` columns; no `order`/`is_active` indexes.
   - `\d game_metas` → exists, with the partial unique index `(game_level_id, content_meta_id) WHERE deleted_at IS NULL`.
   - `\d game_items` → exists, with the analogous partial unique index.
5. **Backup restore:** load the backup into the fresh schema. Expected: no column-mismatch errors (backup is already in the final shape).
6. **Smoke tests:** start server, log in, load one level, play single for ~5 answers, save an AI-custom meta, verify course editor loads a game and a level.

## Risks & Mitigations

| Risk | Mitigation |
|---|---|
| `bootstrap/migrations.go` still references a deleted struct after the edit | `go build ./...` fails loudly — must pass before proceeding. |
| A forgotten service references `content_items.game_level_id` or `.order` or `.is_active` | Models already lack these fields; any such reference would already be a compile error today. Junction feature is merged and working, so this is a settled state. Grep for `game_level_id`, `"order"` on content tables anyway as belt-and-suspenders. |
| User's local backup is actually pre-junction (disagreeing with the brainstorming answer) | User confirmed backup is post-junction. Not a real risk here. |
| The 414001 file quietly depends on one of the deleted migrations running first | It doesn't — 414001 is pure DDL that creates the junction tables independently. |
| An orphan environment (CI, a teammate's laptop) has the old legacy-column chain already recorded | User is the primary developer on this repo (per CLAUDE.md monorepo structure); no teammate migration state is at risk. If an older env ever needs re-sync, a fresh DB import is the documented path. |

## Open Questions

None. All decisions confirmed during brainstorming.

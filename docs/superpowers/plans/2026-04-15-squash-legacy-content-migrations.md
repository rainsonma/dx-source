# Squash Legacy Content Migrations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Collapse `content_metas` / `content_items` migrations to their final junction-era shape by deleting the two transition migrations (backfill + drop-legacy-columns), deleting the orphaned `20260407000001` no-op stub, editing `20260322000036` and `20260322000037` to remove the legacy `game_level_id` / `order` / `is_active` columns, and unregistering the three deleted migrations from `bootstrap/migrations.go`.

**Architecture:** Pure migration-layer cleanup. No application code, no model, no frontend changes. The models (`content_meta.go`, `content_item.go`) already match the target shape. Total surface area: 3 file deletes + 3 file edits, all under `dx-api/database/migrations/` and `dx-api/bootstrap/`.

**Tech Stack:** Go 1.22, Goravel framework, PostgreSQL.

---

## Context for the Implementer

This is a cleanup task, not a feature build. The work is trivial in volume but has one subtle ordering concern you need to understand before touching any file.

**The ordering concern.** The three changes are tightly coupled:
- `20260414000002_backfill_junction_tables.go` runs SQL that reads `content_metas.game_level_id`, `content_metas."order"`, `content_items.game_level_id`, `content_items."order"`. If you remove those columns from the `CREATE TABLE` migrations (36/37) without also deleting 414002, running `go run . artisan migrate` on a fresh DB will succeed through 36/37/414001, then **fail at runtime** on 414002 with a "column does not exist" error.
- `20260414000003_drop_legacy_columns_from_content_tables.go` would still succeed (it uses `DROP COLUMN IF EXISTS`), but that's irrelevant — 414002 has already failed by that point.

So the safe order is: **delete the transition migrations (and unregister them) BEFORE editing migrations 36/37.** That way, at every intermediate commit, a fresh `migrate` run will either (a) still create the legacy columns and not try to touch them afterward (after Task 1) or (b) reach the final target shape (after Tasks 2 and 3).

**Why TDD structure is minimal here.** There is no unit-level behavior to test — we are removing CREATE-TABLE column declarations. The only verifications that make sense are (1) `go build ./...` passes after each edit (catches dangling struct references in `bootstrap/migrations.go`), and (2) a full fresh-migrate works end-to-end at the end.

**Working directory and branch.** The user's convention is to work locally on `main` or a short-lived feature branch and push only `main`. For this trivial change, stay on `main`. All commands below assume the current working directory is the repo root (`/Users/rainsen/Programs/Projects/douxue/dx-source`).

**Do not touch these files.** They are in the target shape already and any changes here are out of scope:
- `dx-api/database/migrations/20260414000001_create_game_metas_and_game_items_tables.go` — still the sole creator of `game_metas` / `game_items`.
- `dx-api/app/models/content_meta.go` — already lacks `GameLevelID` / `Order`.
- `dx-api/app/models/content_item.go` — already lacks `GameLevelID` / `Order` / `IsActive`.
- Any file under `dx-api/app/services/`, `dx-api/app/http/controllers/`, or `dx-web/`.

---

## File Structure

### Files to delete (3)

| Path | Current content | Why delete |
|---|---|---|
| `dx-api/database/migrations/20260407000001_create_game_junction_tables.go` | No-op stub. `Up()` and `Down()` both `return nil`. | Orphan from an earlier, reverted junction-tables attempt. Only kept because some DBs had its old signature recorded; irrelevant on a fresh DB. |
| `dx-api/database/migrations/20260414000002_backfill_junction_tables.go` | Two `INSERT … SELECT` statements moving `content_*` rows into `game_*` junction tables. | Backup already contains post-junction data; fresh DB has no legacy rows to backfill. |
| `dx-api/database/migrations/20260414000003_drop_legacy_columns_from_content_tables.go` | Five `ALTER TABLE … DROP COLUMN` statements. | After Tasks 2 and 3 the columns never exist, so there is nothing to drop. |

### Files to edit (3)

| Path | What changes |
|---|---|
| `dx-api/bootstrap/migrations.go` | Remove the three struct registrations for the deleted files. |
| `dx-api/database/migrations/20260322000036_create_content_metas_table.go` | Remove `game_level_id` column, `order` column, and the `order` index. |
| `dx-api/database/migrations/20260322000037_create_content_items_table.go` | Remove `game_level_id`, `order`, and `is_active` columns plus the `order` and `is_active` indexes. |

Total: 3 deletes, 3 edits. Zero application-code changes.

---

## Task 1: Delete transition migrations and unregister from bootstrap

**Rationale:** Remove the migration files that will cause runtime failures once the legacy columns are gone from 36/37 (Tasks 2 and 3). Doing this first keeps every intermediate commit in a state where `go run . artisan migrate` on a fresh DB still succeeds end-to-end.

**Files:**
- Delete: `dx-api/database/migrations/20260407000001_create_game_junction_tables.go`
- Delete: `dx-api/database/migrations/20260414000002_backfill_junction_tables.go`
- Delete: `dx-api/database/migrations/20260414000003_drop_legacy_columns_from_content_tables.go`
- Modify: `dx-api/bootstrap/migrations.go`

- [ ] **Step 1: Delete the three migration files**

Run:

```bash
rm dx-api/database/migrations/20260407000001_create_game_junction_tables.go
rm dx-api/database/migrations/20260414000002_backfill_junction_tables.go
rm dx-api/database/migrations/20260414000003_drop_legacy_columns_from_content_tables.go
```

Expected: no output; all three files no longer exist.

- [ ] **Step 2: Unregister the three migrations from `bootstrap/migrations.go`**

Open `dx-api/bootstrap/migrations.go`. Remove these three lines from the `Migrations()` slice:

```go
		&migrations.M20260407000001CreateGameJunctionTables{},
```

```go
		&migrations.M20260414000002BackfillJunctionTables{},
		&migrations.M20260414000003DropLegacyColumnsFromContentTables{},
```

After the edit, the tail of the `Migrations()` slice should look like this (keeping the surrounding entries for context):

```go
		&migrations.M20260405000004CreateGameRecordsTable{},
		&migrations.M20260405000005CreateGamePksTable{},
		&migrations.M20260405000006AddGamePkIndexes{},
		&migrations.M20260414000001CreateGameMetasAndGameItemsTables{},
	}
}
```

Note the `407001` line is gone (it used to sit between `406` and `414001`) and the `414002` / `414003` lines are gone (they used to sit after `414001`).

- [ ] **Step 3: Verify the build is clean**

Run:

```bash
cd dx-api && go build ./...
```

Expected: exits 0 with no output. Any dangling reference to the deleted migration structs (only `bootstrap/migrations.go` should reference them) will surface here as a compile error such as:

```
bootstrap/migrations.go:60:18: undefined: migrations.M20260407000001CreateGameJunctionTables
```

If you see this, you missed one of the lines in Step 2 — go back and remove it.

- [ ] **Step 4: Verify `go vet` is clean**

Run:

```bash
cd dx-api && go vet ./...
```

Expected: exits 0 with no output.

- [ ] **Step 5: Commit**

```bash
git add dx-api/database/migrations/20260407000001_create_game_junction_tables.go \
        dx-api/database/migrations/20260414000002_backfill_junction_tables.go \
        dx-api/database/migrations/20260414000003_drop_legacy_columns_from_content_tables.go \
        dx-api/bootstrap/migrations.go
git commit -m "refactor(api): delete legacy-to-junction transition migrations

The backfill and drop-legacy-columns migrations only existed to move a
live DB from the pre-junction schema to the junction schema. With a
post-junction backup being re-imported into a fresh DB, they have no
purpose. Also remove the orphaned 20260407000001 no-op stub."
```

Expected: commit succeeds. `git status` shows a clean tree (or only Tasks 2/3 work-in-progress files if you are reading ahead).

---

## Task 2: Edit migration 36 — remove legacy columns from `content_metas`

**Rationale:** With the transition migrations gone (Task 1), the content-metas create migration can safely drop its legacy `game_level_id` / `order` declarations. After this edit, a fresh migrate creates `content_metas` in its final shape.

**Files:**
- Modify: `dx-api/database/migrations/20260322000036_create_content_metas_table.go`

- [ ] **Step 1: Replace the `Up()` body**

Open `dx-api/database/migrations/20260322000036_create_content_metas_table.go`. Replace the entire `Up()` function with:

```go
func (r *M20260322000036CreateContentMetasTable) Up() error {
	if !facades.Schema().HasTable("content_metas") {
		return facades.Schema().Create("content_metas", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Text("source_from").Default("")
			table.Text("source_type").Default("")
			table.Text("source_data").Default("")
			table.Text("translation").Nullable()
			table.Boolean("is_break_done").Default(false)
			table.TimestampsTz()
			table.SoftDeletesTz()
			table.Index("source_from")
			table.Index("source_type")
			table.Index("created_at")
		})
	}
	return nil
}
```

Changes vs. the original `Up()`:
- Removed: `table.Uuid("game_level_id")`
- Removed: `table.Double("order").Default(0)`
- Removed: `table.Index("order")`

Everything else (`id` PK, `source_from`, `source_type`, `source_data`, `translation`, `is_break_done`, timestamps, soft-deletes, and the `source_from` / `source_type` / `created_at` indexes) stays exactly as-is.

- [ ] **Step 2: Verify the build is clean**

Run:

```bash
cd dx-api && go build ./...
```

Expected: exits 0 with no output.

- [ ] **Step 3: Verify `go vet` is clean**

Run:

```bash
cd dx-api && go vet ./...
```

Expected: exits 0 with no output.

- [ ] **Step 4: Commit**

```bash
git add dx-api/database/migrations/20260322000036_create_content_metas_table.go
git commit -m "refactor(api): remove legacy columns from content_metas migration

Drop game_level_id, order column, and the order index from the
content_metas create migration. These were only needed for the
pre-junction-tables schema; the junction tables (game_metas) are now
the authoritative source for level membership and ordering."
```

Expected: commit succeeds.

---

## Task 3: Edit migration 37 — remove legacy columns from `content_items`

**Rationale:** Same reasoning as Task 2 but for `content_items`. This also removes the already-dead `is_active` column (never written to by any code path).

**Files:**
- Modify: `dx-api/database/migrations/20260322000037_create_content_items_table.go`

- [ ] **Step 1: Replace the `Up()` body**

Open `dx-api/database/migrations/20260322000037_create_content_items_table.go`. Replace the entire `Up()` function with:

```go
func (r *M20260322000037CreateContentItemsTable) Up() error {
	if !facades.Schema().HasTable("content_items") {
		return facades.Schema().Create("content_items", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("content_meta_id").Nullable()
			table.Text("content").Default("")
			table.Text("content_type").Default("")
			table.Uuid("uk_audio_id").Nullable()
			table.Uuid("us_audio_id").Nullable()
			table.Text("definition").Nullable()
			table.Text("translation").Nullable()
			table.Text("explanation").Nullable()
			table.Json("items").Nullable()
			table.Json("structure").Nullable()
			table.Column("tags", "text[]").Nullable()
			table.TimestampsTz()
			table.SoftDeletesTz()
			table.Index("content_meta_id")
			table.Index("uk_audio_id")
			table.Index("us_audio_id")
			table.Index("content_type")
			table.Index("created_at")
		})
	}
	return nil
}
```

Changes vs. the original `Up()`:
- Removed: `table.Uuid("game_level_id")`
- Removed: `table.Double("order").Default(0)`
- Removed: `table.Boolean("is_active").Default(true)`
- Removed: `table.Index("order")`
- Removed: `table.Index("is_active")`

Everything else stays exactly as-is: `id` PK, `content_meta_id` (nullable, immutable parsing lineage), `content`, `content_type`, `uk_audio_id` / `us_audio_id` (nullable), `definition`, `translation`, `explanation`, `items` (jsonb), `structure` (jsonb), `tags` (text[]), timestamps, soft-deletes, and the `content_meta_id` / `uk_audio_id` / `us_audio_id` / `content_type` / `created_at` indexes.

- [ ] **Step 2: Verify the build is clean**

Run:

```bash
cd dx-api && go build ./...
```

Expected: exits 0 with no output.

- [ ] **Step 3: Verify `go vet` is clean**

Run:

```bash
cd dx-api && go vet ./...
```

Expected: exits 0 with no output.

- [ ] **Step 4: Commit**

```bash
git add dx-api/database/migrations/20260322000037_create_content_items_table.go
git commit -m "refactor(api): remove legacy columns from content_items migration

Drop game_level_id, order, is_active columns (and their indexes) from
the content_items create migration. Level membership and per-level
ordering now live in game_items; is_active was dead code already."
```

Expected: commit succeeds.

---

## Task 4: End-to-end verification on a fresh database

**Rationale:** Each earlier task was verified by `go build` and `go vet`, which catch Go-level issues but not SQL-level issues. This task runs the full migrate chain on a clean DB and confirms the final schema matches expectations. It also does a quick smoke test against a couple of hot paths.

**Files:** none — this is verification only.

> **Note:** Do not run Steps 1–3 on your normal development database if it has work-in-progress data. Use a throwaway DB or accept that the data will be wiped. The user's stated plan is exactly this flow: drop the local DB, migrate fresh, import backup — so this is the expected sequence.

- [ ] **Step 1: Drop and recreate the local database**

Run (from any directory; adjust credentials if your local setup differs):

```bash
psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS douxue;"
psql -h localhost -U postgres -c "CREATE DATABASE douxue;"
```

Expected: `DROP DATABASE` and `CREATE DATABASE` both succeed. If `douxue` is actively connected by another process (e.g., a running server), the drop will fail — stop any running `go run .` / `air` process first.

- [ ] **Step 2: Run the full migration chain**

```bash
cd dx-api && go run . artisan migrate
```

Expected: output includes a "Migrated" line for each registered migration, ending with `20260414000001_create_game_metas_and_game_items_tables`. You should **not** see any lines for `20260407000001_create_game_junction_tables`, `20260414000002_backfill_junction_tables`, or `20260414000003_drop_legacy_columns_from_content_tables`. No errors.

If you see a "column does not exist" error, one of Tasks 1–3 was done out of order and there is still a live reference to a legacy column somewhere. Re-read the "ordering concern" in the context section above, then `git log --oneline -5` to see which tasks have been committed.

- [ ] **Step 3: Inspect the final schema**

Run:

```bash
psql -h localhost -U postgres -d douxue -c "\d content_metas"
psql -h localhost -U postgres -d douxue -c "\d content_items"
psql -h localhost -U postgres -d douxue -c "\d game_metas"
psql -h localhost -U postgres -d douxue -c "\d game_items"
```

Expected `content_metas` columns: `id`, `source_from`, `source_type`, `source_data`, `translation`, `is_break_done`, `created_at`, `updated_at`, `deleted_at`. No `game_level_id`, no `order`.
Expected `content_metas` indexes: primary key on `id`, plus indexes on `source_from`, `source_type`, `created_at`. No `order` index.

Expected `content_items` columns: `id`, `content_meta_id`, `content`, `content_type`, `uk_audio_id`, `us_audio_id`, `definition`, `translation`, `explanation`, `items`, `structure`, `tags`, `created_at`, `updated_at`, `deleted_at`. No `game_level_id`, no `order`, no `is_active`.
Expected `content_items` indexes: primary key on `id`, plus indexes on `content_meta_id`, `uk_audio_id`, `us_audio_id`, `content_type`, `created_at`. No `order` or `is_active` indexes.

Expected `game_metas` columns: `id`, `game_id`, `game_level_id`, `content_meta_id`, `order`, `created_at`, `updated_at`, `deleted_at`.
Expected `game_metas` indexes: primary key on `id`, the partial unique index `idx_game_metas_level_meta_unique ON game_metas (game_level_id, content_meta_id) WHERE deleted_at IS NULL`, and `idx_game_metas_level_order`, plus indexes on `game_id`, `content_meta_id`, `created_at`.

Expected `game_items` columns: same shape as `game_metas` but with `content_item_id` instead of `content_meta_id`.
Expected `game_items` indexes: analogous — partial unique on `(game_level_id, content_item_id) WHERE deleted_at IS NULL`, plus `idx_game_items_level_order`, `game_id`, `content_item_id`, `created_at`.

If any expected column or index is missing, something is wrong — stop and investigate before proceeding to Step 4.

- [ ] **Step 4: Restore the backup**

Run the backup restore using whichever mechanism you have locally. The backup was taken post-junction, so the dump's INSERT statements target the schema that now exists.

Expected: no column-mismatch errors. Row counts after restore should match whatever was in the backup.

If you see an error like `column "game_level_id" does not exist`, the backup is actually pre-junction (contradicting the brainstorming assumption), and the cleanup needs to be rolled back with `git revert` on Tasks 1–3 until the backup question is resolved.

- [ ] **Step 5: Smoke test**

Start the server:

```bash
cd dx-api && go run .
```

Expected: server starts on port 3001 without startup errors.

In a separate terminal, run the frontend:

```bash
cd dx-web && npm run dev
```

Open `http://localhost:3000` and:

1. Log in with an existing account (any user from the backup).
2. Navigate to the hall and pick a course with levels.
3. Start a single-play session. Expected: level loads, content items display with audio, you can advance through at least 5 items without errors.
4. Open the course editor (if you are logged in as an admin). Navigate to any course → any level. Expected: the level content loads, metas and items render in order.
5. Save a new AI-custom meta. Expected: save succeeds, the new meta appears in the editor list.

Watch the `go run .` server output while doing the above. Any 5xx error or stack trace means one of the runtime paths is still referencing something that was removed — stop and investigate.

- [ ] **Step 6: Final commit marker (optional)**

No new file changes, but if you made any tiny adjustments during verification, commit them now. Otherwise this step is a no-op.

```bash
git status
```

Expected: clean working tree. The four verification tasks should show up as the commits from Tasks 1, 2, 3.

---

## Self-Review Checklist (for the implementer, after all tasks)

- [ ] `bootstrap/migrations.go` has exactly the expected registrations: everything that was there before, minus `M20260407000001CreateGameJunctionTables`, `M20260414000002BackfillJunctionTables`, `M20260414000003DropLegacyColumnsFromContentTables`.
- [ ] Three files are gone from `dx-api/database/migrations/`: `20260407000001_*`, `20260414000002_*`, `20260414000003_*`.
- [ ] `20260322000036_create_content_metas_table.go` has no references to `game_level_id`, `order` column, or `Index("order")`.
- [ ] `20260322000037_create_content_items_table.go` has no references to `game_level_id`, `order` column, `is_active`, `Index("order")`, or `Index("is_active")`.
- [ ] `20260414000001_create_game_metas_and_game_items_tables.go` is **unchanged**.
- [ ] Models `content_meta.go` and `content_item.go` are **unchanged**.
- [ ] `go build ./...` and `go vet ./...` both pass in `dx-api`.
- [ ] Fresh `artisan migrate` succeeds end-to-end with no errors.
- [ ] Schema inspection matches the expected columns and indexes in Task 4 Step 3.
- [ ] Backup restored without column-mismatch errors.
- [ ] Smoke tests pass (single play + editor load + save AI meta).

If any item is unchecked, fix before declaring the task done.

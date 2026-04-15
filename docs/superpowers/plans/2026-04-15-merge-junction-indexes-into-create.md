# Merge Junction Indexes Into Create Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move all four junction-table indexes from their separate migration files into the `20260414000001` blueprint via Goravel's `table.Index(...).Name(...)`, then delete `20260414000002` and `20260415000001` entirely and unregister them from `bootstrap/migrations.go`. The user will reset the DB and restore from the post-relax data-only backup afterward.

**Architecture:** Pure migration-layer cleanup. No application code changes. 1 file edited, 2 files deleted, 1 bootstrap entry adjusted = 4 file touches in one atomic commit. On a fresh migrate, the squashed `20260414000001` creates both junction tables AND their four indexes through a single `facades.Schema().Create()` call — no raw SQL mixing required, which avoids the Goravel Schema/Orm connection-split issue that originally forced the index DDL into a separate migration.

**Tech Stack:** Go 1.22, Goravel framework 1.17.2, PostgreSQL 18.

---

## Context for the Implementer

### Background

The junction tables `game_metas` / `game_items` currently carry their four indexes across two separate migrations:

- `20260414000002_add_game_junction_partial_indexes.go` (edited on 2026-04-15 commit `6351322`) creates all four indexes via raw SQL: two non-unique 3-column `(game_level_id, content_*_id, deleted_at)` and two composite `(game_level_id, deleted_at, "order")`.
- `20260415000001_relax_game_junction_indexes.go` (new file on 2026-04-15 commit `6351322`) transitions live DBs from the pre-2026-04-15 partial unique indexes to the non-unique shape. It is a no-op on fresh migrate paths (`IF NOT EXISTS` / `IF EXISTS` guards cover everything).

The live DB has already been migrated to the target shape. The user has taken a **data-only** backup at `/Users/rainsen/Programs/Projects/douxue/db-backup/dx-data-only-post-relax-indexes-20260415-122012.sql` (1.3 GB, verified clean — 0 DDL statements, 51 COPY blocks including all four junction/content tables). After this squash, they'll drop `dxdb`, re-migrate fresh, and restore from that data-only dump.

### Why this squash is safe — the Goravel Schema/Orm connection split

In an earlier session we discovered that mixing `facades.Schema().Create(...)` with `facades.Orm().Query().Exec(raw SQL)` in the same migration fails on a fresh DB because `Schema()` and `Orm().Query()` use different DB connections — the raw SQL cannot see tables that `Schema()` just created. That bug is why the indexes were originally split out of the create migration.

This squash avoids the bug by keeping everything inside the `Schema().Create(...)` blueprint closure. Goravel's `Blueprint.Index(column ...string)` method accepts multiple columns and returns an `IndexDefinition` that exposes a `.Name(string)` method for explicit naming. Verified from the framework source at `/Users/rainsen/go/pkg/mod/github.com/goravel/framework@v1.17.2/database/schema/blueprint.go:328` and `contracts/database/schema/index.go:39`. All four indexes can be expressed this way without any raw SQL.

### What must NOT change

- Any non-migration file — the squash is migration-only.
- The model files `game_meta.go` / `game_item.go` — they don't reference indexes and don't need edits.
- The signature `20260414000001_create_game_metas_and_game_items_tables` on the existing migration struct (Goravel tracks migrations by signature; changing it would cause the runner to try to re-apply on any DB where it's already tracked).
- Live DB state — this change only affects code. The user handles their own reset + restore afterward.

### Confirmed Blueprint API (from Goravel 1.17.2 source)

- `func (r *Blueprint) Index(column ...string) schema.IndexDefinition` at `blueprint.go:328` — variadic, accepts any number of columns.
- `IndexDefinition.Name(name string) IndexDefinition` at `contracts/database/schema/index.go:39` — sets an explicit index name.
- Reserved word `order` is correctly quoted by Blueprint; precedent at `20260322000027_create_game_levels_table.go:30` and `20260322000006_create_game_presses_table.go:27`.
- Multi-column index precedent at `20260405000004_create_game_records_table.go:36`: `table.Index("user_id", "created_at")`.

### Working directory and branch

All absolute paths assume repo root `/Users/rainsen/Programs/Projects/douxue/dx-source`. Current branch `main`. User has explicitly consented to working on main for this cleanup.

### Verification approach

This task has no automated tests (the repo has no `game_metas` / `game_items` test coverage). Verification is: `go build ./...` + `go vet ./...` after the edits. The live-DB restore flow is user-executed afterward and not part of this plan's tasks.

---

## File Structure

Four file touches in one commit:

| File | Change type |
|---|---|
| `dx-api/database/migrations/20260414000001_create_game_metas_and_game_items_tables.go` | edit (add 4 index lines to the blueprint closures) |
| `dx-api/database/migrations/20260414000002_add_game_junction_partial_indexes.go` | delete |
| `dx-api/database/migrations/20260415000001_relax_game_junction_indexes.go` | delete |
| `dx-api/bootstrap/migrations.go` | edit (remove 2 struct registrations) |

---

## Task 1: Edit 20260414000001 — merge 4 indexes into the blueprint closures

**Rationale:** The blueprint already creates both tables with columns and single-column indexes. Adding the 4 index lines through `table.Index(...).Name(...)` keeps everything inside a single `Schema().Create()` call, avoiding the Schema/Orm split bug that forced the indexes into a separate migration originally.

**Files:**
- Modify: `dx-api/database/migrations/20260414000001_create_game_metas_and_game_items_tables.go`

- [ ] **Step 1: Replace the `Up()` body**

Find the current `Up()` function (approximately lines 14-54 of the current file):

```go
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
	}

	return nil
}
```

Replace with:

```go
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
			table.Index("game_level_id", "content_meta_id", "deleted_at").Name("idx_game_metas_level_meta")
			table.Index("game_level_id", "deleted_at", "order").Name("idx_game_metas_level_order")
		}); err != nil {
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
			table.Index("game_level_id", "content_item_id", "deleted_at").Name("idx_game_items_level_item")
			table.Index("game_level_id", "deleted_at", "order").Name("idx_game_items_level_order")
		}); err != nil {
			return err
		}
	}

	return nil
}
```

Changes vs. the original:
- `game_metas` blueprint gained two lines after `table.Index("created_at")`:
  - `table.Index("game_level_id", "content_meta_id", "deleted_at").Name("idx_game_metas_level_meta")`
  - `table.Index("game_level_id", "deleted_at", "order").Name("idx_game_metas_level_order")`
- `game_items` blueprint gained two analogous lines after its `table.Index("created_at")`.
- No other changes. The `Down()`, `Signature()`, struct declaration, and imports are untouched.

Each index uses explicit `.Name()` to match the current live-DB names exactly — this is important for two reasons: (1) docs and code elsewhere in the project reference these names, (2) if someone inspects the live DB's indexes after a fresh migrate, the names should match what they'd see on the currently-running migrated DB.

---

## Task 2: Delete the two now-redundant migration files

**Rationale:** After Task 1, the indexes are created by `20260414000001`. The separate index-creation migration (`20260414000002`) and the transition migration (`20260415000001`) are no longer needed. On a fresh DB they'd be no-ops anyway (one creates indexes that Task 1 just built with the same names; the other's `IF NOT EXISTS` / `IF EXISTS` guards skip everything).

**Files:**
- Delete: `dx-api/database/migrations/20260414000002_add_game_junction_partial_indexes.go`
- Delete: `dx-api/database/migrations/20260415000001_relax_game_junction_indexes.go`

- [ ] **Step 1: Delete both files**

Run from the repo root:

```bash
rm /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/database/migrations/20260414000002_add_game_junction_partial_indexes.go
rm /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/database/migrations/20260415000001_relax_game_junction_indexes.go
```

Expected: both `rm` commands succeed silently. `ls dx-api/database/migrations/ | grep 20260415` should return nothing, and `ls dx-api/database/migrations/ | grep 20260414` should return only `20260414000001_create_game_metas_and_game_items_tables.go`.

---

## Task 3: Edit bootstrap/migrations.go — unregister the deleted migrations

**Rationale:** Goravel's migration runner walks the `Migrations()` slice. Leaving references to deleted struct types in the slice causes a compile error. Removing them reaches the intended final state.

**Files:**
- Modify: `dx-api/bootstrap/migrations.go`

- [ ] **Step 1: Remove the two registrations**

Find these lines (they appear near the tail of the `Migrations()` slice):

```go
		&migrations.M20260414000001CreateGameMetasAndGameItemsTables{},
		&migrations.M20260414000002AddGameJunctionPartialIndexes{},
		&migrations.M20260415000001RelaxGameJunctionIndexes{},
	}
```

Replace with:

```go
		&migrations.M20260414000001CreateGameMetasAndGameItemsTables{},
	}
```

The `M20260414000001CreateGameMetasAndGameItemsTables{}` entry is kept. The two entries below it are removed. No reordering of any other entry.

---

## Task 4: Build check and commit

**Rationale:** Compilation catches any missing deletion or leftover reference. After a clean build, the squash is safe to commit as a single atomic change.

**Files:** all 4 from Tasks 1-3 (plus the two deleted files).

- [ ] **Step 1: Build check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: exits 0 with no output.

Typical failure modes if something is wrong:
- `undefined: migrations.M20260414000002AddGameJunctionPartialIndexes` → Task 3 Step 1 missed a line; re-check `bootstrap/migrations.go`.
- `undefined: migrations.M20260415000001RelaxGameJunctionIndexes` → same.

- [ ] **Step 2: Vet check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./...
```

Expected: exits 0 with no output.

- [ ] **Step 3: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && \
git add dx-api/database/migrations/20260414000001_create_game_metas_and_game_items_tables.go \
        dx-api/database/migrations/20260414000002_add_game_junction_partial_indexes.go \
        dx-api/database/migrations/20260415000001_relax_game_junction_indexes.go \
        dx-api/bootstrap/migrations.go && \
git commit -m "$(cat <<'EOF'
refactor(api): merge junction indexes into create migration

Move the four game_metas / game_items indexes from the separate
20260414000002 (raw-SQL index creation) and 20260415000001
(live-DB transition) migrations into the 20260414000001 blueprint
via table.Index(...).Name(...). Delete both now-redundant
migration files and unregister them from bootstrap/migrations.go.

Goravel Blueprint's Index(column ...string).Name(string) chain
supports the exact shape we need — multi-column 3-col keys,
explicit names matching the current live-DB indexes, and the
reserved "order" column handled by Blueprint's internal quoting.
This keeps everything inside a single Schema().Create() call,
avoiding the Schema/Orm connection-split bug that originally
forced the index DDL out of the create migration.

The live DB has already been migrated to the target shape. This
squash only affects fresh-migrate paths. User will drop + recreate
dxdb and restore from the post-relax data-only backup
(dx-data-only-post-relax-indexes-20260415-122012.sql).
EOF
)"
```

Expected: single commit created.

- [ ] **Step 4: Verify the commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git log -1 --stat
```

Expected: subject `refactor(api): merge junction indexes into create migration`, and a file list showing exactly 4 entries — one modify for `20260414000001_...go`, one delete for `20260414000002_...go`, one delete for `20260415000001_...go`, and one modify for `bootstrap/migrations.go`.

---

## Self-Review Checklist (for the implementer, after all tasks)

- [ ] `20260414000001_create_game_metas_and_game_items_tables.go` has the 4 new index lines (2 on `game_metas`, 2 on `game_items`) with explicit `.Name()` values matching the live DB.
- [ ] `20260414000002_add_game_junction_partial_indexes.go` no longer exists on disk.
- [ ] `20260415000001_relax_game_junction_indexes.go` no longer exists on disk.
- [ ] `bootstrap/migrations.go` has exactly one junction-table migration registration (`M20260414000001CreateGameMetasAndGameItemsTables{}`); the other two are gone.
- [ ] `cd dx-api && go build ./...` exits 0.
- [ ] `cd dx-api && go vet ./...` exits 0.
- [ ] `git log -1 --stat` shows 4 files changed: 1 modify + 2 deletes + 1 modify, with the expected commit subject.
- [ ] `git status` shows a clean working tree.

---

## User-Executed Follow-up (not for the implementer)

After the commit lands, the user will:

1. Stop any running `go run .` / `air` process.
2. Drop and recreate `dxdb`:
   ```bash
   docker compose -f deploy/docker-compose.dev.yml exec -T postgres \
     psql -U postgres -c "DROP DATABASE IF EXISTS dxdb;" -c "CREATE DATABASE dxdb;"
   ```
3. Run the squashed migrations:
   ```bash
   docker compose -f deploy/docker-compose.dev.yml exec dx-api go run . artisan migrate
   ```
4. Verify the junction-table indexes match expectations:
   ```bash
   docker compose -f deploy/docker-compose.dev.yml exec -T postgres \
     psql -U postgres -d dxdb -c "\d game_metas" -c "\d game_items"
   ```
   Expected: `idx_game_metas_level_meta` on `(game_level_id, content_meta_id, deleted_at)`, `idx_game_metas_level_order` on `(game_level_id, deleted_at, "order")`, plus their `game_items` counterparts. No `_unique` names.
5. Load the data-only backup:
   ```bash
   docker compose -f deploy/docker-compose.dev.yml exec -T postgres \
     psql -U postgres -d dxdb \
     < /Users/rainsen/Programs/Projects/douxue/db-backup/dx-data-only-post-relax-indexes-20260415-122012.sql
   ```
6. Confirm row counts: `SELECT COUNT(*) FROM game_items WHERE deleted_at IS NULL;` should match `1,220,803`.
7. Smoke test in the browser as in the previous task.

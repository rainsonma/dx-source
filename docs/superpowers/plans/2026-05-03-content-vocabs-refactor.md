# Content / Vocab Schema Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Drop `game_metas` / `game_items` junctions, denormalize `(game_id, game_level_id, order)` onto `content_metas` / `content_items` (word-sentence only), introduce a public-wiki vocab pool (`content_vocabs` canonical + `game_vocabs` placement) with anti-vandalism gating + audit, and thread two-FK polymorphism through 6 tracking tables. dx-mini stays untouched via server-side response-shape preservation.

**Architecture:** Goravel `Schema().Create()` for new tables; sibling `*_raw.go` migrations for partial unique indexes + CHECK constraints (Goravel API doesn't expose them); models follow plain GORM struct pattern; services in `app/services/api/` keep "controller-thin, service-thick" rule; tests use testify suite with per-test seed/teardown.

**Tech Stack:** Go 1.23 + Goravel 1.16 + GORM, PostgreSQL 16 + pg_partman, DeepSeek (sentence break + POS enrichment), Next.js 16 (dx-web frontend), shadcn/ui, TailwindCSS v4, Zod.

**Spec reference:** `docs/superpowers/specs/2026-05-03-content-vocabs-refactor-design.md`

---

## Pre-flight

Before starting any task, verify the working environment:

- [ ] **Confirm git state**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git status
```

Expected: clean tree on `main` (or on a fresh `refactor/content-vocabs` branch — recommended for the size of this change).

- [ ] **Confirm dx-api builds and tests pass before starting**

```bash
cd dx-api
go build ./... && go test -race ./...
```

Expected: build OK; tests pass.

- [ ] **Confirm dx-web builds**

```bash
cd ../dx-web
npx tsc --noEmit
```

Expected: 0 errors.

- [ ] **Confirm Postgres + Redis are up (dev compose)**

```bash
cd ../deploy
docker compose -f docker-compose.dev.yml ps postgres redis
```

Expected: both `Up` and healthy.

**Validation gates (run at end of each phase):**
- `cd dx-api && gofmt -l . && go vet ./... && go build ./... && go test -race ./...`
- `cd dx-web && npm run lint && npx tsc --noEmit && npm run build`

---

## Phase 1 — Schema migrations

The 8 affected create_table files plus 7 new sibling raw-SQL migrations + 2 new create migrations. Apply edits in prefix order so renumbered files don't collide. Run migrations only at the end of the phase.

### Task 1.1: Edit content_metas migration

**Files:**
- Modify: `dx-api/database/migrations/20260322000036_create_content_metas_table.go`

- [ ] **Step 1: Replace the file contents**

```go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000036CreateContentMetasTable struct{}

func (r *M20260322000036CreateContentMetasTable) Signature() string {
	return "20260322000036_create_content_metas_table"
}

func (r *M20260322000036CreateContentMetasTable) Up() error {
	if !facades.Schema().HasTable("content_metas") {
		return facades.Schema().Create("content_metas", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Text("source_from").Default("")
			table.Text("source_type").Default("")
			table.Text("source_data").Default("")
			table.Text("translation").Nullable()
			table.Text("speaker").Nullable()
			table.Boolean("is_break_done").Default(false)
			table.Double("order").Default(0)
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("source_from")
			table.Index("source_type")
			table.Index("created_at")
			table.Index("game_id")
			table.Index("game_level_id", "deleted_at", "order").Name("idx_content_metas_level_order")
		})
	}
	return nil
}

func (r *M20260322000036CreateContentMetasTable) Down() error {
	return facades.Schema().DropIfExists("content_metas")
}
```

- [ ] **Step 2: Verify formatting**

```bash
cd dx-api && gofmt -l database/migrations/20260322000036_create_content_metas_table.go
```

Expected: no output (file is gofmt-clean).

- [ ] **Step 3: Commit**

```bash
git add dx-api/database/migrations/20260322000036_create_content_metas_table.go
git commit -m "refactor(api): denormalize game_id/level_id/order onto content_metas; drop dedup index"
```

### Task 1.2: Edit content_items migration

**Files:**
- Modify: `dx-api/database/migrations/20260322000037_create_content_items_table.go`

- [ ] **Step 1: Replace the file contents**

```go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000037CreateContentItemsTable struct{}

func (r *M20260322000037CreateContentItemsTable) Signature() string {
	return "20260322000037_create_content_items_table"
}

func (r *M20260322000037CreateContentItemsTable) Up() error {
	if !facades.Schema().HasTable("content_items") {
		return facades.Schema().Create("content_items", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Uuid("content_meta_id").Nullable()
			table.Text("content").Default("")
			table.Text("content_type").Default("")
			table.Text("uk_audio_url").Nullable()
			table.Text("us_audio_url").Nullable()
			table.Text("definition").Nullable()
			table.Text("translation").Nullable()
			table.Text("explanation").Nullable()
			table.Text("speaker").Nullable()
			table.Json("items").Nullable()
			table.Json("structure").Nullable()
			table.Double("order").Default(0)
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("content_meta_id")
			table.Index("content_type")
			table.Index("created_at")
			table.Index("game_id")
			table.Index("game_level_id", "deleted_at", "order").Name("idx_content_items_level_order")
		})
	}
	return nil
}

func (r *M20260322000037CreateContentItemsTable) Down() error {
	return facades.Schema().DropIfExists("content_items")
}
```

- [ ] **Step 2: Verify formatting**

```bash
cd dx-api && gofmt -l database/migrations/20260322000037_create_content_items_table.go
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add dx-api/database/migrations/20260322000037_create_content_items_table.go
git commit -m "refactor(api): denormalize game_id/level_id/order onto content_items"
```

### Task 1.3: Edit game_reports migration

**Files:**
- Modify: `dx-api/database/migrations/20260322000043_create_game_reports_table.go`

- [ ] **Step 1: Replace the file contents** (drops existing `Unique`, makes `content_item_id` nullable, adds `content_vocab_id` + `SoftDeletesTz`)

```go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000043CreateGameReportsTable struct{}

func (r *M20260322000043CreateGameReportsTable) Signature() string {
	return "20260322000043_create_game_reports_table"
}

func (r *M20260322000043CreateGameReportsTable) Up() error {
	if !facades.Schema().HasTable("game_reports") {
		return facades.Schema().Create("game_reports", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Uuid("content_item_id").Nullable()
			table.Uuid("content_vocab_id").Nullable()
			table.Text("reason").Default("")
			table.Text("note").Nullable()
			table.Integer("count").Default(0)
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("user_id")
			table.Index("game_id")
		})
	}
	return nil
}

func (r *M20260322000043CreateGameReportsTable) Down() error {
	return facades.Schema().DropIfExists("game_reports")
}
```

- [ ] **Step 2: Verify formatting**

```bash
cd dx-api && gofmt -l database/migrations/20260322000043_create_game_reports_table.go
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add dx-api/database/migrations/20260322000043_create_game_reports_table.go
git commit -m "refactor(api): game_reports — nullable content_item_id, add content_vocab_id, soft-delete"
```

### Task 1.4: Add game_reports raw sibling

**Files:**
- Create: `dx-api/database/migrations/20260322000044_add_game_reports_raw.go`

- [ ] **Step 1: Create the file**

```go
package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260322000044AddGameReportsRaw struct{}

func (r *M20260322000044AddGameReportsRaw) Signature() string {
	return "20260322000044_add_game_reports_raw"
}

func (r *M20260322000044AddGameReportsRaw) Up() error {
	statements := []string{
		`CREATE INDEX idx_game_reports_content_item_id
           ON game_reports (content_item_id)
           WHERE content_item_id IS NOT NULL`,
		`CREATE INDEX idx_game_reports_content_vocab_id
           ON game_reports (content_vocab_id)
           WHERE content_vocab_id IS NOT NULL`,
		`CREATE UNIQUE INDEX idx_game_reports_user_item_reason_uq
           ON game_reports (user_id, content_item_id, reason)
           WHERE content_item_id IS NOT NULL AND deleted_at IS NULL`,
		`CREATE UNIQUE INDEX idx_game_reports_user_vocab_reason_uq
           ON game_reports (user_id, content_vocab_id, reason)
           WHERE content_vocab_id IS NOT NULL AND deleted_at IS NULL`,
		`ALTER TABLE game_reports
           ADD CONSTRAINT game_reports_content_xor
           CHECK ((content_item_id IS NULL) != (content_vocab_id IS NULL))`,
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260322000044AddGameReportsRaw) Down() error {
	statements := []string{
		"ALTER TABLE game_reports DROP CONSTRAINT IF EXISTS game_reports_content_xor",
		"DROP INDEX IF EXISTS idx_game_reports_user_vocab_reason_uq",
		"DROP INDEX IF EXISTS idx_game_reports_user_item_reason_uq",
		"DROP INDEX IF EXISTS idx_game_reports_content_vocab_id",
		"DROP INDEX IF EXISTS idx_game_reports_content_item_id",
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 2: Verify formatting**

```bash
cd dx-api && gofmt -l database/migrations/20260322000044_add_game_reports_raw.go
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add dx-api/database/migrations/20260322000044_add_game_reports_raw.go
git commit -m "refactor(api): add game_reports sibling raw — partial uniques + xor check"
```

### Task 1.5: Renumber + edit user_masters migration

**Files:**
- Delete: `dx-api/database/migrations/20260322000044_create_user_masters_table.go`
- Create: `dx-api/database/migrations/20260322000045_create_user_masters_table.go`

- [ ] **Step 1: Delete old file**

```bash
rm dx-api/database/migrations/20260322000044_create_user_masters_table.go
```

- [ ] **Step 2: Create renumbered file**

```go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000045CreateUserMastersTable struct{}

func (r *M20260322000045CreateUserMastersTable) Signature() string {
	return "20260322000045_create_user_masters_table"
}

func (r *M20260322000045CreateUserMastersTable) Up() error {
	if !facades.Schema().HasTable("user_masters") {
		return facades.Schema().Create("user_masters", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Uuid("content_item_id").Nullable()
			table.Uuid("content_vocab_id").Nullable()
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.TimestampTz("mastered_at").Nullable()
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("user_id")
			table.Index("game_id")
			table.Index("game_level_id")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000045CreateUserMastersTable) Down() error {
	return facades.Schema().DropIfExists("user_masters")
}
```

- [ ] **Step 3: Commit**

```bash
git add dx-api/database/migrations/20260322000044_create_user_masters_table.go dx-api/database/migrations/20260322000045_create_user_masters_table.go
git commit -m "refactor(api): renumber user_masters; nullable content_item_id, add content_vocab_id, soft-delete"
```

### Task 1.6: Add user_masters raw sibling

**Files:**
- Create: `dx-api/database/migrations/20260322000046_add_user_masters_raw.go`

- [ ] **Step 1: Create the file**

```go
package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260322000046AddUserMastersRaw struct{}

func (r *M20260322000046AddUserMastersRaw) Signature() string {
	return "20260322000046_add_user_masters_raw"
}

func (r *M20260322000046AddUserMastersRaw) Up() error {
	statements := []string{
		`CREATE INDEX idx_user_masters_content_item_id
           ON user_masters (content_item_id)
           WHERE content_item_id IS NOT NULL`,
		`CREATE INDEX idx_user_masters_content_vocab_id
           ON user_masters (content_vocab_id)
           WHERE content_vocab_id IS NOT NULL`,
		`CREATE UNIQUE INDEX idx_user_masters_user_item_uq
           ON user_masters (user_id, content_item_id)
           WHERE content_item_id IS NOT NULL AND deleted_at IS NULL`,
		`CREATE UNIQUE INDEX idx_user_masters_user_vocab_uq
           ON user_masters (user_id, content_vocab_id)
           WHERE content_vocab_id IS NOT NULL AND deleted_at IS NULL`,
		`ALTER TABLE user_masters
           ADD CONSTRAINT user_masters_content_xor
           CHECK ((content_item_id IS NULL) != (content_vocab_id IS NULL))`,
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260322000046AddUserMastersRaw) Down() error {
	statements := []string{
		"ALTER TABLE user_masters DROP CONSTRAINT IF EXISTS user_masters_content_xor",
		"DROP INDEX IF EXISTS idx_user_masters_user_vocab_uq",
		"DROP INDEX IF EXISTS idx_user_masters_user_item_uq",
		"DROP INDEX IF EXISTS idx_user_masters_content_vocab_id",
		"DROP INDEX IF EXISTS idx_user_masters_content_item_id",
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/database/migrations/20260322000046_add_user_masters_raw.go
git commit -m "refactor(api): add user_masters sibling raw — partial uniques + xor check"
```

### Task 1.7: Renumber + edit user_unknowns migration

**Files:**
- Delete: `dx-api/database/migrations/20260322000045_create_user_unknowns_table.go`
- Create: `dx-api/database/migrations/20260322000047_create_user_unknowns_table.go`

- [ ] **Step 1: Delete old file**

```bash
rm dx-api/database/migrations/20260322000045_create_user_unknowns_table.go
```

- [ ] **Step 2: Create renumbered file**

```go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000047CreateUserUnknownsTable struct{}

func (r *M20260322000047CreateUserUnknownsTable) Signature() string {
	return "20260322000047_create_user_unknowns_table"
}

func (r *M20260322000047CreateUserUnknownsTable) Up() error {
	if !facades.Schema().HasTable("user_unknowns") {
		return facades.Schema().Create("user_unknowns", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Uuid("content_item_id").Nullable()
			table.Uuid("content_vocab_id").Nullable()
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("user_id")
			table.Index("game_id")
			table.Index("game_level_id")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000047CreateUserUnknownsTable) Down() error {
	return facades.Schema().DropIfExists("user_unknowns")
}
```

- [ ] **Step 3: Commit**

```bash
git add dx-api/database/migrations/20260322000045_create_user_unknowns_table.go dx-api/database/migrations/20260322000047_create_user_unknowns_table.go
git commit -m "refactor(api): renumber user_unknowns; nullable content_item_id, add content_vocab_id, soft-delete"
```

### Task 1.8: Add user_unknowns raw sibling

**Files:**
- Create: `dx-api/database/migrations/20260322000048_add_user_unknowns_raw.go`

- [ ] **Step 1: Create the file**

```go
package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260322000048AddUserUnknownsRaw struct{}

func (r *M20260322000048AddUserUnknownsRaw) Signature() string {
	return "20260322000048_add_user_unknowns_raw"
}

func (r *M20260322000048AddUserUnknownsRaw) Up() error {
	statements := []string{
		`CREATE INDEX idx_user_unknowns_content_item_id
           ON user_unknowns (content_item_id)
           WHERE content_item_id IS NOT NULL`,
		`CREATE INDEX idx_user_unknowns_content_vocab_id
           ON user_unknowns (content_vocab_id)
           WHERE content_vocab_id IS NOT NULL`,
		`CREATE UNIQUE INDEX idx_user_unknowns_user_item_uq
           ON user_unknowns (user_id, content_item_id)
           WHERE content_item_id IS NOT NULL AND deleted_at IS NULL`,
		`CREATE UNIQUE INDEX idx_user_unknowns_user_vocab_uq
           ON user_unknowns (user_id, content_vocab_id)
           WHERE content_vocab_id IS NOT NULL AND deleted_at IS NULL`,
		`ALTER TABLE user_unknowns
           ADD CONSTRAINT user_unknowns_content_xor
           CHECK ((content_item_id IS NULL) != (content_vocab_id IS NULL))`,
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260322000048AddUserUnknownsRaw) Down() error {
	statements := []string{
		"ALTER TABLE user_unknowns DROP CONSTRAINT IF EXISTS user_unknowns_content_xor",
		"DROP INDEX IF EXISTS idx_user_unknowns_user_vocab_uq",
		"DROP INDEX IF EXISTS idx_user_unknowns_user_item_uq",
		"DROP INDEX IF EXISTS idx_user_unknowns_content_vocab_id",
		"DROP INDEX IF EXISTS idx_user_unknowns_content_item_id",
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/database/migrations/20260322000048_add_user_unknowns_raw.go
git commit -m "refactor(api): add user_unknowns sibling raw — partial uniques + xor check"
```

### Task 1.9: Renumber + edit user_reviews migration

**Files:**
- Delete: `dx-api/database/migrations/20260322000046_create_user_reviews_table.go`
- Create: `dx-api/database/migrations/20260322000049_create_user_reviews_table.go`

- [ ] **Step 1: Delete old file**

```bash
rm dx-api/database/migrations/20260322000046_create_user_reviews_table.go
```

- [ ] **Step 2: Create renumbered file**

```go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000049CreateUserReviewsTable struct{}

func (r *M20260322000049CreateUserReviewsTable) Signature() string {
	return "20260322000049_create_user_reviews_table"
}

func (r *M20260322000049CreateUserReviewsTable) Up() error {
	if !facades.Schema().HasTable("user_reviews") {
		return facades.Schema().Create("user_reviews", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Uuid("content_item_id").Nullable()
			table.Uuid("content_vocab_id").Nullable()
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.TimestampTz("last_review_at").Nullable()
			table.TimestampTz("next_review_at").Nullable()
			table.Integer("review_count").Default(0)
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("user_id")
			table.Index("game_id")
			table.Index("game_level_id")
			table.Index("next_review_at")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000049CreateUserReviewsTable) Down() error {
	return facades.Schema().DropIfExists("user_reviews")
}
```

- [ ] **Step 3: Commit**

```bash
git add dx-api/database/migrations/20260322000046_create_user_reviews_table.go dx-api/database/migrations/20260322000049_create_user_reviews_table.go
git commit -m "refactor(api): renumber user_reviews; nullable content_item_id, add content_vocab_id, soft-delete"
```

### Task 1.10: Add user_reviews raw sibling

**Files:**
- Create: `dx-api/database/migrations/20260322000050_add_user_reviews_raw.go`

- [ ] **Step 1: Create the file**

```go
package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260322000050AddUserReviewsRaw struct{}

func (r *M20260322000050AddUserReviewsRaw) Signature() string {
	return "20260322000050_add_user_reviews_raw"
}

func (r *M20260322000050AddUserReviewsRaw) Up() error {
	statements := []string{
		`CREATE INDEX idx_user_reviews_content_item_id
           ON user_reviews (content_item_id)
           WHERE content_item_id IS NOT NULL`,
		`CREATE INDEX idx_user_reviews_content_vocab_id
           ON user_reviews (content_vocab_id)
           WHERE content_vocab_id IS NOT NULL`,
		`CREATE UNIQUE INDEX idx_user_reviews_user_item_uq
           ON user_reviews (user_id, content_item_id)
           WHERE content_item_id IS NOT NULL AND deleted_at IS NULL`,
		`CREATE UNIQUE INDEX idx_user_reviews_user_vocab_uq
           ON user_reviews (user_id, content_vocab_id)
           WHERE content_vocab_id IS NOT NULL AND deleted_at IS NULL`,
		`ALTER TABLE user_reviews
           ADD CONSTRAINT user_reviews_content_xor
           CHECK ((content_item_id IS NULL) != (content_vocab_id IS NULL))`,
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260322000050AddUserReviewsRaw) Down() error {
	statements := []string{
		"ALTER TABLE user_reviews DROP CONSTRAINT IF EXISTS user_reviews_content_xor",
		"DROP INDEX IF EXISTS idx_user_reviews_user_vocab_uq",
		"DROP INDEX IF EXISTS idx_user_reviews_user_item_uq",
		"DROP INDEX IF EXISTS idx_user_reviews_content_vocab_id",
		"DROP INDEX IF EXISTS idx_user_reviews_content_item_id",
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/database/migrations/20260322000050_add_user_reviews_raw.go
git commit -m "refactor(api): add user_reviews sibling raw — partial uniques + xor check"
```

### Task 1.11: Edit game_sessions migration

**Files:**
- Modify: `dx-api/database/migrations/20260405000002_create_game_sessions_table.go`

- [ ] **Step 1: Replace the file contents**

```go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260405000002CreateGameSessionsTable struct{}

func (r *M20260405000002CreateGameSessionsTable) Signature() string {
	return "20260405000002_create_game_sessions_table"
}

func (r *M20260405000002CreateGameSessionsTable) Up() error {
	return facades.Schema().Create("game_sessions", func(table schema.Blueprint) {
		table.Uuid("id")
		table.Primary("id")
		table.Uuid("user_id")
		table.Uuid("game_id")
		table.Uuid("game_level_id")
		table.Text("degree").Default("")
		table.Text("pattern").Nullable()
		table.Uuid("current_content_item_id").Nullable()
		table.Uuid("current_content_vocab_id").Nullable()
		table.TimestampTz("started_at")
		table.TimestampTz("last_played_at")
		table.TimestampTz("ended_at").Nullable()
		table.Integer("score").Default(0)
		table.Integer("exp").Default(0)
		table.Integer("max_combo").Default(0)
		table.Integer("correct_count").Default(0)
		table.Integer("wrong_count").Default(0)
		table.Integer("skip_count").Default(0)
		table.Integer("play_time").Default(0)
		table.Integer("total_items_count").Default(0)
		table.Integer("played_items_count").Default(0)
		table.Uuid("game_group_id").Nullable()
		table.Uuid("game_subgroup_id").Nullable()
		table.Uuid("game_pk_id").Nullable()
		table.SoftDeletesTz()
		table.TimestampsTz()
	})
}

func (r *M20260405000002CreateGameSessionsTable) Down() error {
	return facades.Schema().DropIfExists("game_sessions")
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/database/migrations/20260405000002_create_game_sessions_table.go
git commit -m "refactor(api): game_sessions — add current_content_vocab_id, soft-delete"
```

### Task 1.12: Rename + edit add_game_session_indexes → add_game_sessions_raw

**Files:**
- Delete: `dx-api/database/migrations/20260405000003_add_game_session_indexes.go`
- Create: `dx-api/database/migrations/20260405000003_add_game_sessions_raw.go`

- [ ] **Step 1: Delete old file**

```bash
rm dx-api/database/migrations/20260405000003_add_game_session_indexes.go
```

- [ ] **Step 2: Create renamed file with appended index + at-most-one CHECK**

```go
package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260405000003AddGameSessionsRaw struct{}

func (r *M20260405000003AddGameSessionsRaw) Signature() string {
	return "20260405000003_add_game_sessions_raw"
}

func (r *M20260405000003AddGameSessionsRaw) Up() error {
	statements := []string{
		`CREATE UNIQUE INDEX idx_game_sessions_active_single ON game_sessions (user_id, game_level_id, degree, COALESCE(pattern, '')) WHERE ended_at IS NULL AND game_group_id IS NULL AND game_pk_id IS NULL`,
		`CREATE UNIQUE INDEX idx_game_sessions_active_group ON game_sessions (user_id, game_level_id, degree, COALESCE(pattern, ''), game_group_id) WHERE ended_at IS NULL AND game_group_id IS NOT NULL`,
		`CREATE UNIQUE INDEX idx_game_sessions_active_pk ON game_sessions (user_id, game_level_id, degree, COALESCE(pattern, ''), game_pk_id) WHERE ended_at IS NULL AND game_pk_id IS NOT NULL`,
		`CREATE INDEX idx_game_sessions_group ON game_sessions (game_group_id) WHERE game_group_id IS NOT NULL`,
		`CREATE INDEX idx_game_sessions_pk ON game_sessions (game_pk_id) WHERE game_pk_id IS NOT NULL`,
		`CREATE INDEX idx_game_sessions_leaderboard ON game_sessions (user_id, last_played_at)`,
		`CREATE INDEX idx_game_sessions_user_game ON game_sessions (user_id, game_id)`,
		`CREATE INDEX idx_game_sessions_current_content_vocab_id
           ON game_sessions (current_content_vocab_id)
           WHERE current_content_vocab_id IS NOT NULL`,
		`ALTER TABLE game_sessions
           ADD CONSTRAINT game_sessions_current_content_xor
           CHECK (NOT (current_content_item_id IS NOT NULL
                  AND current_content_vocab_id IS NOT NULL))`,
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260405000003AddGameSessionsRaw) Down() error {
	statements := []string{
		"ALTER TABLE game_sessions DROP CONSTRAINT IF EXISTS game_sessions_current_content_xor",
		"DROP INDEX IF EXISTS idx_game_sessions_current_content_vocab_id",
		"DROP INDEX IF EXISTS idx_game_sessions_active_single",
		"DROP INDEX IF EXISTS idx_game_sessions_active_group",
		"DROP INDEX IF EXISTS idx_game_sessions_active_pk",
		"DROP INDEX IF EXISTS idx_game_sessions_group",
		"DROP INDEX IF EXISTS idx_game_sessions_pk",
		"DROP INDEX IF EXISTS idx_game_sessions_leaderboard",
		"DROP INDEX IF EXISTS idx_game_sessions_user_game",
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 3: Commit**

```bash
git add dx-api/database/migrations/20260405000003_add_game_session_indexes.go dx-api/database/migrations/20260405000003_add_game_sessions_raw.go
git commit -m "refactor(api): rename add_game_session_indexes → add_game_sessions_raw; add at-most-one xor"
```

### Task 1.13: Edit game_records migration

**Files:**
- Modify: `dx-api/database/migrations/20260405000004_create_game_records_table.go`

- [ ] **Step 1: Replace the file contents**

```go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260405000004CreateGameRecordsTable struct{}

func (r *M20260405000004CreateGameRecordsTable) Signature() string {
	return "20260405000004_create_game_records_table"
}

func (r *M20260405000004CreateGameRecordsTable) Up() error {
	return facades.Schema().Create("game_records", func(table schema.Blueprint) {
		table.Uuid("id")
		table.Primary("id")
		table.Uuid("user_id")
		table.Uuid("game_session_id")
		table.Uuid("game_level_id")
		table.Uuid("content_item_id").Nullable()
		table.Uuid("content_vocab_id").Nullable()
		table.Boolean("is_correct").Default(false)
		table.Text("source_answer").Default("")
		table.Text("user_answer").Default("")
		table.Integer("base_score").Default(0)
		table.Integer("combo_score").Default(0)
		table.Integer("duration").Default(0)
		table.SoftDeletesTz()
		table.TimestampsTz()
		table.Index("user_id")
		table.Index("game_session_id")
		table.Index("game_level_id")
		table.Index("is_correct")
		table.Index("user_id", "created_at")
	})
}

func (r *M20260405000004CreateGameRecordsTable) Down() error {
	return facades.Schema().DropIfExists("game_records")
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/database/migrations/20260405000004_create_game_records_table.go
git commit -m "refactor(api): game_records — nullable content_item_id, add content_vocab_id, soft-delete"
```

### Task 1.14: Add game_records raw sibling

**Files:**
- Create: `dx-api/database/migrations/20260405000005_add_game_records_raw.go`

- [ ] **Step 1: Create the file**

```go
package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260405000005AddGameRecordsRaw struct{}

func (r *M20260405000005AddGameRecordsRaw) Signature() string {
	return "20260405000005_add_game_records_raw"
}

func (r *M20260405000005AddGameRecordsRaw) Up() error {
	statements := []string{
		`CREATE INDEX idx_game_records_content_item_id
           ON game_records (content_item_id)
           WHERE content_item_id IS NOT NULL`,
		`CREATE INDEX idx_game_records_content_vocab_id
           ON game_records (content_vocab_id)
           WHERE content_vocab_id IS NOT NULL`,
		`CREATE UNIQUE INDEX idx_game_records_session_item_uq
           ON game_records (game_session_id, content_item_id)
           WHERE content_item_id IS NOT NULL AND deleted_at IS NULL`,
		`CREATE UNIQUE INDEX idx_game_records_session_vocab_uq
           ON game_records (game_session_id, content_vocab_id)
           WHERE content_vocab_id IS NOT NULL AND deleted_at IS NULL`,
		`ALTER TABLE game_records
           ADD CONSTRAINT game_records_content_xor
           CHECK ((content_item_id IS NULL) != (content_vocab_id IS NULL))`,
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260405000005AddGameRecordsRaw) Down() error {
	statements := []string{
		"ALTER TABLE game_records DROP CONSTRAINT IF EXISTS game_records_content_xor",
		"DROP INDEX IF EXISTS idx_game_records_session_vocab_uq",
		"DROP INDEX IF EXISTS idx_game_records_session_item_uq",
		"DROP INDEX IF EXISTS idx_game_records_content_vocab_id",
		"DROP INDEX IF EXISTS idx_game_records_content_item_id",
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/database/migrations/20260405000005_add_game_records_raw.go
git commit -m "refactor(api): add game_records sibling raw — partial uniques + xor check"
```

### Task 1.15: Renumber game_pks migration

**Files:**
- Delete: `dx-api/database/migrations/20260405000005_create_game_pks_table.go`
- Create: `dx-api/database/migrations/20260405000006_create_game_pks_table.go`

- [ ] **Step 1: Read the existing file to preserve content**

```bash
cat dx-api/database/migrations/20260405000005_create_game_pks_table.go
```

- [ ] **Step 2: Create renumbered file**

Copy the content from the existing file verbatim, but update:
- Filename prefix `20260405000005` → `20260405000006`
- Struct name `M20260405000005CreateGamePksTable` → `M20260405000006CreateGamePksTable`
- Signature string `20260405000005_create_game_pks_table` → `20260405000006_create_game_pks_table`

(All three places must change; the table contents do NOT change.)

- [ ] **Step 3: Delete old file**

```bash
rm dx-api/database/migrations/20260405000005_create_game_pks_table.go
```

- [ ] **Step 4: Commit**

```bash
git add dx-api/database/migrations/20260405000005_create_game_pks_table.go dx-api/database/migrations/20260405000006_create_game_pks_table.go
git commit -m "refactor(api): renumber game_pks migration to make room for game_records sibling"
```

### Task 1.16: Rename + renumber add_game_pk_indexes → add_game_pks_raw

**Files:**
- Delete: `dx-api/database/migrations/20260405000006_add_game_pk_indexes.go`
- Create: `dx-api/database/migrations/20260405000007_add_game_pks_raw.go`

- [ ] **Step 1: Read the existing file to preserve content**

```bash
cat dx-api/database/migrations/20260405000006_add_game_pk_indexes.go
```

- [ ] **Step 2: Create renamed + renumbered file**

Copy the content verbatim, but update:
- Filename prefix `20260405000006` → `20260405000007`
- Struct name `M20260405000006AddGamePkIndexes` → `M20260405000007AddGamePksRaw`
- Signature string → `20260405000007_add_game_pks_raw`

(All three places must change; SQL statements unchanged.)

- [ ] **Step 3: Delete old file**

```bash
rm dx-api/database/migrations/20260405000006_add_game_pk_indexes.go
```

- [ ] **Step 4: Commit**

```bash
git add dx-api/database/migrations/20260405000006_add_game_pk_indexes.go dx-api/database/migrations/20260405000007_add_game_pks_raw.go
git commit -m "refactor(api): rename add_game_pk_indexes → add_game_pks_raw; renumber"
```

### Task 1.17: Delete game_metas_and_game_items migration

**Files:**
- Delete: `dx-api/database/migrations/20260414000001_create_game_metas_and_game_items_tables.go`

- [ ] **Step 1: Delete the file**

```bash
rm dx-api/database/migrations/20260414000001_create_game_metas_and_game_items_tables.go
```

- [ ] **Step 2: Commit**

```bash
git add -A dx-api/database/migrations/
git commit -m "refactor(api): delete game_metas_and_game_items migration — junctions removed"
```

### Task 1.18: Add new content_vocabs_and_game_vocabs migration

**Files:**
- Create: `dx-api/database/migrations/20260414000001_create_content_vocabs_and_game_vocabs_tables.go`

- [ ] **Step 1: Create the file**

```go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260414000001CreateContentVocabsAndGameVocabsTables struct{}

func (r *M20260414000001CreateContentVocabsAndGameVocabsTables) Signature() string {
	return "20260414000001_create_content_vocabs_and_game_vocabs_tables"
}

func (r *M20260414000001CreateContentVocabsAndGameVocabsTables) Up() error {
	if !facades.Schema().HasTable("content_vocabs") {
		if err := facades.Schema().Create("content_vocabs", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Text("content").Default("")
			table.Text("content_key").Default("")
			table.Text("uk_phonetic").Nullable()
			table.Text("us_phonetic").Nullable()
			table.Text("uk_audio_url").Nullable()
			table.Text("us_audio_url").Nullable()
			table.Json("definition").Nullable()
			table.Text("explanation").Nullable()
			table.Boolean("is_verified").Default(false)
			table.Uuid("created_by").Nullable()
			table.Uuid("last_edited_by").Nullable()
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("created_by")
			table.Index("created_at")
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("game_vocabs") {
		if err := facades.Schema().Create("game_vocabs", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("game_id")
			table.Uuid("game_level_id")
			table.Uuid("content_vocab_id")
			table.Double("order").Default(0)
			table.SoftDeletesTz()
			table.TimestampsTz()
			table.Index("game_id")
			table.Index("content_vocab_id")
			table.Index("created_at")
			table.Index("game_level_id", "deleted_at", "order").Name("idx_game_vocabs_level_order")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *M20260414000001CreateContentVocabsAndGameVocabsTables) Down() error {
	if err := facades.Schema().DropIfExists("game_vocabs"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("content_vocabs")
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/database/migrations/20260414000001_create_content_vocabs_and_game_vocabs_tables.go
git commit -m "feat(api): add content_vocabs (canonical wiki) and game_vocabs (placement) tables"
```

### Task 1.19: Add content_vocabs raw sibling

**Files:**
- Create: `dx-api/database/migrations/20260414000002_add_content_vocabs_raw.go`

- [ ] **Step 1: Create the file**

```go
package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260414000002AddContentVocabsRaw struct{}

func (r *M20260414000002AddContentVocabsRaw) Signature() string {
	return "20260414000002_add_content_vocabs_raw"
}

func (r *M20260414000002AddContentVocabsRaw) Up() error {
	statements := []string{
		`CREATE UNIQUE INDEX idx_content_vocabs_content_key_uq
           ON content_vocabs (content_key)
           WHERE deleted_at IS NULL`,
		`CREATE INDEX idx_content_vocabs_content_key
           ON content_vocabs (content_key, deleted_at)`,
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260414000002AddContentVocabsRaw) Down() error {
	statements := []string{
		"DROP INDEX IF EXISTS idx_content_vocabs_content_key",
		"DROP INDEX IF EXISTS idx_content_vocabs_content_key_uq",
	}
	for _, sql := range statements {
		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/database/migrations/20260414000002_add_content_vocabs_raw.go
git commit -m "refactor(api): add content_vocabs sibling raw — partial unique on content_key"
```

### Task 1.20: Add content_vocab_edits migration

**Files:**
- Create: `dx-api/database/migrations/20260414000003_create_content_vocab_edits_table.go`

- [ ] **Step 1: Create the file**

```go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260414000003CreateContentVocabEditsTable struct{}

func (r *M20260414000003CreateContentVocabEditsTable) Signature() string {
	return "20260414000003_create_content_vocab_edits_table"
}

func (r *M20260414000003CreateContentVocabEditsTable) Up() error {
	if facades.Schema().HasTable("content_vocab_edits") {
		return nil
	}
	return facades.Schema().Create("content_vocab_edits", func(table schema.Blueprint) {
		table.Uuid("id")
		table.Primary("id")
		table.Uuid("content_vocab_id")
		table.Uuid("editor_user_id").Nullable()
		table.Text("edit_type").Default("")
		table.Json("before").Nullable()
		table.Json("after").Nullable()
		table.SoftDeletesTz()
		table.TimestampsTz()
		table.Index("content_vocab_id")
		table.Index("editor_user_id")
		table.Index("created_at")
	})
}

func (r *M20260414000003CreateContentVocabEditsTable) Down() error {
	return facades.Schema().DropIfExists("content_vocab_edits")
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/database/migrations/20260414000003_create_content_vocab_edits_table.go
git commit -m "feat(api): add content_vocab_edits audit log table"
```

### Task 1.21: Update bootstrap/migrations.go

**Files:**
- Modify: `dx-api/bootstrap/migrations.go`

- [ ] **Step 1: Replace the file contents**

```go
package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dx-api/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		&migrations.M20260322000001CreateUsersTable{},
		&migrations.M20260322000002CreateAdmUsersTable{},
		&migrations.M20260322000003CreateAdmRolesTable{},
		&migrations.M20260322000004CreateAdmPermitsTable{},
		&migrations.M20260322000005CreateGameCategoriesTable{},
		&migrations.M20260322000006CreateGamePressesTable{},
		&migrations.M20260322000007CreateNoticesTable{},
		&migrations.M20260322000008CreateSettingsTable{},
		&migrations.M20260322000009CreateUserLoginsTable{},
		&migrations.M20260322000010CreateUserBeansTable{},
		&migrations.M20260322000011CreateUserFavoritesTable{},
		&migrations.M20260322000012CreateUserFollowsTable{},
		&migrations.M20260322000013CreateUserRedeemsTable{},
		&migrations.M20260322000014CreateUserReferralsTable{},
		&migrations.M20260322000015CreateUserSettingsTable{},
		&migrations.M20260322000016CreateGamesTable{},
		&migrations.M20260322000019CreateAdmUserRolesTable{},
		&migrations.M20260322000020CreateAdmUserPermitsTable{},
		&migrations.M20260322000021CreateAdmRolePermitsTable{},
		&migrations.M20260322000022CreateAdmMenusTable{},
		&migrations.M20260322000023CreateAdmLoginsTable{},
		&migrations.M20260322000024CreateAdmOperatesTable{},
		&migrations.M20260322000025CreateFeedbacksTable{},
		&migrations.M20260322000026CreateContentSeeksTable{},
		&migrations.M20260322000027CreateGameLevelsTable{},
		&migrations.M20260322000028CreateGameGroupsTable{},
		&migrations.M20260322000029CreateGameGroupMembersTable{},
		&migrations.M20260322000030CreateGameSubgroupsTable{},
		&migrations.M20260322000031CreateGameSubgroupMembersTable{},
		&migrations.M20260322000032CreatePostsTable{},
		&migrations.M20260322000033CreatePostCommentsTable{},
		&migrations.M20260322000034CreatePostLikesTable{},
		&migrations.M20260322000035CreatePostBookmarksTable{},
		&migrations.M20260322000036CreateContentMetasTable{},
		&migrations.M20260322000037CreateContentItemsTable{},
		&migrations.M20260322000043CreateGameReportsTable{},
		&migrations.M20260322000044AddGameReportsRaw{},
		&migrations.M20260322000045CreateUserMastersTable{},
		&migrations.M20260322000046AddUserMastersRaw{},
		&migrations.M20260322000047CreateUserUnknownsTable{},
		&migrations.M20260322000048AddUserUnknownsRaw{},
		&migrations.M20260322000049CreateUserReviewsTable{},
		&migrations.M20260322000050AddUserReviewsRaw{},
		&migrations.M20260324000002CreateGameGroupApplicationsTable{},
		&migrations.M20260403000001CreateOrdersTable{},
		&migrations.M20260405000002CreateGameSessionsTable{},
		&migrations.M20260405000003AddGameSessionsRaw{},
		&migrations.M20260405000004CreateGameRecordsTable{},
		&migrations.M20260405000005AddGameRecordsRaw{},
		&migrations.M20260405000006CreateGamePksTable{},
		&migrations.M20260405000007AddGamePksRaw{},
		&migrations.M20260414000001CreateContentVocabsAndGameVocabsTables{},
		&migrations.M20260414000002AddContentVocabsRaw{},
		&migrations.M20260414000003CreateContentVocabEditsTable{},
	}
}
```

- [ ] **Step 2: Verify build**

```bash
cd dx-api && go build ./bootstrap/
```

Expected: no errors. If you see "M20260414000001CreateGameMetasAndGameItemsTables undefined", that's expected — it's been deleted; just confirm it's not in the list above.

- [ ] **Step 3: Commit**

```bash
git add dx-api/bootstrap/migrations.go
git commit -m "refactor(api): re-register migrations after refactor renames + sibling additions"
```

### Task 1.22: Reset database and run migrations

**Files:** None (operational task)

- [ ] **Step 1: Drop and recreate the dev database**

```bash
cd dx-api
docker exec -it $(docker ps -qf "name=postgres") psql -U postgres -c "DROP DATABASE IF EXISTS douxue;"
docker exec -it $(docker ps -qf "name=postgres") psql -U postgres -c "CREATE DATABASE douxue;"
```

If your local Postgres isn't via docker, swap in your own admin shell.

Expected: `DROP DATABASE` (or `NOTICE: database does not exist`), `CREATE DATABASE`.

- [ ] **Step 2: Run migrations**

```bash
cd dx-api && go run . artisan migrate
```

Expected: All migrations run in order; ends with `Migration table created successfully` and a list of every migration name.

- [ ] **Step 3: Verify schema**

```bash
docker exec -it $(docker ps -qf "name=postgres") psql -U postgres -d douxue -c "\dt"
```

Expected: list contains `content_metas`, `content_items`, `content_vocabs`, `game_vocabs`, `content_vocab_edits`, `user_masters`, `user_unknowns`, `user_reviews`, `game_reports`, `game_records`, `game_sessions`. Does NOT contain `game_metas` or `game_items`.

- [ ] **Step 4: Verify a sample table structure**

```bash
docker exec -it $(docker ps -qf "name=postgres") psql -U postgres -d douxue -c "\d user_masters"
```

Expected output should show: `content_item_id` (uuid, NULLABLE), `content_vocab_id` (uuid, NULLABLE), `deleted_at`, partial unique indexes for both, and CHECK `user_masters_content_xor`.

- [ ] **Step 5: Verify content_vocabs partial unique works**

```bash
docker exec -it $(docker ps -qf "name=postgres") psql -U postgres -d douxue -c "\d content_vocabs"
```

Expected: shows `idx_content_vocabs_content_key_uq` as partial unique index `WHERE deleted_at IS NULL`.

- [ ] **Step 6: No code commit needed**

Phase 1 is complete. Database schema matches the spec.

### Phase 1 validation gate

- [ ] **Run all gates**

```bash
cd dx-api && gofmt -l . && go vet ./... && go build ./...
```

Expected: no output from gofmt; no errors from vet/build. Note: tests will fail in subsequent phases until services and models are updated — `go test` is deferred to Phase 4 onward.

---

## Phase 2 — Models

Edit existing models to match the schema and add 3 new model files. Delete `game_meta.go` and `game_item.go` last so we can update services that import them in Phase 4.

### Task 2.1: Edit content_meta.go

**Files:**
- Modify: `dx-api/app/models/content_meta.go`

- [ ] **Step 1: Replace the file contents**

```go
package models

import "github.com/goravel/framework/database/orm"

type ContentMeta struct {
	orm.Timestamps
	orm.SoftDeletes
	ID          string  `gorm:"column:id;primaryKey" json:"id"`
	GameID      string  `gorm:"column:game_id" json:"game_id"`
	GameLevelID string  `gorm:"column:game_level_id" json:"game_level_id"`
	SourceFrom  string  `gorm:"column:source_from" json:"source_from"`
	SourceType  string  `gorm:"column:source_type" json:"source_type"`
	SourceData  string  `gorm:"column:source_data" json:"source_data"`
	Translation *string `gorm:"column:translation" json:"translation"`
	Speaker     *string `gorm:"column:speaker" json:"speaker"`
	IsBreakDone bool    `gorm:"column:is_break_done" json:"is_break_done"`
	Order       float64 `gorm:"column:order" json:"order"`
}

// TableName returns the database table name.
func (c *ContentMeta) TableName() string {
	return "content_metas"
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/models/content_meta.go
git commit -m "refactor(api): content_meta model — add game_id/level_id/order, speaker"
```

### Task 2.2: Edit content_item.go (drop Tags, add fields)

**Files:**
- Modify: `dx-api/app/models/content_item.go`

- [ ] **Step 1: Replace the file contents** (drops `pq.StringArray` Tags field; the `lib/pq` import becomes unused — remove it)

```go
package models

import (
	"github.com/goravel/framework/database/orm"
)

type ContentItem struct {
	orm.Timestamps
	orm.SoftDeletes
	ID            string  `gorm:"column:id;primaryKey" json:"id"`
	GameID        string  `gorm:"column:game_id" json:"game_id"`
	GameLevelID   string  `gorm:"column:game_level_id" json:"game_level_id"`
	ContentMetaID *string `gorm:"column:content_meta_id" json:"content_meta_id"`
	Content       string  `gorm:"column:content" json:"content"`
	ContentType   string  `gorm:"column:content_type" json:"content_type"`
	UkAudioURL    *string `gorm:"column:uk_audio_url" json:"uk_audio_url"`
	UsAudioURL    *string `gorm:"column:us_audio_url" json:"us_audio_url"`
	Definition    *string `gorm:"column:definition" json:"definition"`
	Translation   *string `gorm:"column:translation" json:"translation"`
	Explanation   *string `gorm:"column:explanation" json:"explanation"`
	Speaker       *string `gorm:"column:speaker" json:"speaker"`
	Items         *string `gorm:"column:items;type:jsonb" json:"items"`
	Structure     *string `gorm:"column:structure;type:jsonb" json:"structure"`
	Order         float64 `gorm:"column:order" json:"order"`
}

// TableName returns the database table name.
func (c *ContentItem) TableName() string {
	return "content_items"
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/models/content_item.go
git commit -m "refactor(api): content_item model — add game_id/level_id/order, drop dead Tags field"
```

### Task 2.3: Add content_vocab.go

**Files:**
- Create: `dx-api/app/models/content_vocab.go`

- [ ] **Step 1: Create the file**

```go
package models

import "github.com/goravel/framework/database/orm"

type ContentVocab struct {
	orm.Timestamps
	orm.SoftDeletes
	ID            string  `gorm:"column:id;primaryKey" json:"id"`
	Content       string  `gorm:"column:content" json:"content"`
	ContentKey    string  `gorm:"column:content_key" json:"content_key"`
	UkPhonetic    *string `gorm:"column:uk_phonetic" json:"uk_phonetic"`
	UsPhonetic    *string `gorm:"column:us_phonetic" json:"us_phonetic"`
	UkAudioURL    *string `gorm:"column:uk_audio_url" json:"uk_audio_url"`
	UsAudioURL    *string `gorm:"column:us_audio_url" json:"us_audio_url"`
	Definition    *string `gorm:"column:definition;type:jsonb" json:"definition"`
	Explanation   *string `gorm:"column:explanation" json:"explanation"`
	IsVerified    bool    `gorm:"column:is_verified" json:"is_verified"`
	CreatedBy     *string `gorm:"column:created_by" json:"created_by"`
	LastEditedBy  *string `gorm:"column:last_edited_by" json:"last_edited_by"`
}

// TableName returns the database table name.
func (c *ContentVocab) TableName() string {
	return "content_vocabs"
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/models/content_vocab.go
git commit -m "feat(api): add ContentVocab model (canonical wiki entry)"
```

### Task 2.4: Add game_vocab.go

**Files:**
- Create: `dx-api/app/models/game_vocab.go`

- [ ] **Step 1: Create the file**

```go
package models

import "github.com/goravel/framework/database/orm"

type GameVocab struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string  `gorm:"column:id;primaryKey" json:"id"`
	GameID         string  `gorm:"column:game_id" json:"game_id"`
	GameLevelID    string  `gorm:"column:game_level_id" json:"game_level_id"`
	ContentVocabID string  `gorm:"column:content_vocab_id" json:"content_vocab_id"`
	Order          float64 `gorm:"column:order" json:"order"`
}

// TableName returns the database table name.
func (g *GameVocab) TableName() string {
	return "game_vocabs"
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/models/game_vocab.go
git commit -m "feat(api): add GameVocab model (placement junction)"
```

### Task 2.5: Add content_vocab_edit.go

**Files:**
- Create: `dx-api/app/models/content_vocab_edit.go`

- [ ] **Step 1: Create the file**

```go
package models

import "github.com/goravel/framework/database/orm"

type ContentVocabEdit struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string  `gorm:"column:id;primaryKey" json:"id"`
	ContentVocabID string  `gorm:"column:content_vocab_id" json:"content_vocab_id"`
	EditorUserID   *string `gorm:"column:editor_user_id" json:"editor_user_id"`
	EditType       string  `gorm:"column:edit_type" json:"edit_type"`
	Before         *string `gorm:"column:before;type:jsonb" json:"before"`
	After          *string `gorm:"column:after;type:jsonb" json:"after"`
}

// TableName returns the database table name.
func (c *ContentVocabEdit) TableName() string {
	return "content_vocab_edits"
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/models/content_vocab_edit.go
git commit -m "feat(api): add ContentVocabEdit model (audit log)"
```

### Task 2.6: Edit user_master.go (polymorphism + soft-delete)

**Files:**
- Modify: `dx-api/app/models/user_master.go`

- [ ] **Step 1: Replace the file contents**

```go
package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

type UserMaster struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string           `gorm:"column:id;primaryKey" json:"id"`
	UserID         string           `gorm:"column:user_id" json:"user_id"`
	ContentItemID  *string          `gorm:"column:content_item_id" json:"content_item_id"`
	ContentVocabID *string          `gorm:"column:content_vocab_id" json:"content_vocab_id"`
	GameID         string           `gorm:"column:game_id" json:"game_id"`
	GameLevelID    string           `gorm:"column:game_level_id" json:"game_level_id"`
	MasteredAt     *carbon.DateTime `gorm:"column:mastered_at" json:"mastered_at"`
}

// TableName returns the database table name.
func (u *UserMaster) TableName() string {
	return "user_masters"
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/models/user_master.go
git commit -m "refactor(api): user_master model — polymorphic content FK, soft-delete"
```

### Task 2.7: Edit user_unknown.go

**Files:**
- Modify: `dx-api/app/models/user_unknown.go`

- [ ] **Step 1: Replace the file contents**

```go
package models

import "github.com/goravel/framework/database/orm"

type UserUnknown struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string  `gorm:"column:id;primaryKey" json:"id"`
	UserID         string  `gorm:"column:user_id" json:"user_id"`
	ContentItemID  *string `gorm:"column:content_item_id" json:"content_item_id"`
	ContentVocabID *string `gorm:"column:content_vocab_id" json:"content_vocab_id"`
	GameID         string  `gorm:"column:game_id" json:"game_id"`
	GameLevelID    string  `gorm:"column:game_level_id" json:"game_level_id"`
}

// TableName returns the database table name.
func (u *UserUnknown) TableName() string {
	return "user_unknowns"
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/models/user_unknown.go
git commit -m "refactor(api): user_unknown model — polymorphic content FK, soft-delete"
```

### Task 2.8: Edit user_review.go

**Files:**
- Modify: `dx-api/app/models/user_review.go`

- [ ] **Step 1: Replace the file contents**

```go
package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

type UserReview struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string           `gorm:"column:id;primaryKey" json:"id"`
	UserID         string           `gorm:"column:user_id" json:"user_id"`
	ContentItemID  *string          `gorm:"column:content_item_id" json:"content_item_id"`
	ContentVocabID *string          `gorm:"column:content_vocab_id" json:"content_vocab_id"`
	GameID         string           `gorm:"column:game_id" json:"game_id"`
	GameLevelID    string           `gorm:"column:game_level_id" json:"game_level_id"`
	LastReviewAt   *carbon.DateTime `gorm:"column:last_review_at" json:"last_review_at"`
	NextReviewAt   *carbon.DateTime `gorm:"column:next_review_at" json:"next_review_at"`
	ReviewCount    int              `gorm:"column:review_count" json:"review_count"`
}

// TableName returns the database table name.
func (u *UserReview) TableName() string {
	return "user_reviews"
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/models/user_review.go
git commit -m "refactor(api): user_review model — polymorphic content FK, soft-delete"
```

### Task 2.9: Edit game_record.go

**Files:**
- Modify: `dx-api/app/models/game_record.go`

- [ ] **Step 1: Replace the file contents**

```go
package models

import "github.com/goravel/framework/database/orm"

type GameRecord struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string  `gorm:"column:id;primaryKey" json:"id"`
	UserID         string  `gorm:"column:user_id" json:"user_id"`
	GameSessionID  string  `gorm:"column:game_session_id" json:"game_session_id"`
	GameLevelID    string  `gorm:"column:game_level_id" json:"game_level_id"`
	ContentItemID  *string `gorm:"column:content_item_id" json:"content_item_id"`
	ContentVocabID *string `gorm:"column:content_vocab_id" json:"content_vocab_id"`
	IsCorrect      bool    `gorm:"column:is_correct" json:"is_correct"`
	SourceAnswer   string  `gorm:"column:source_answer" json:"source_answer"`
	UserAnswer     string  `gorm:"column:user_answer" json:"user_answer"`
	BaseScore      int     `gorm:"column:base_score" json:"base_score"`
	ComboScore     int     `gorm:"column:combo_score" json:"combo_score"`
	Duration       int     `gorm:"column:duration" json:"duration"`
}

func (g *GameRecord) TableName() string {
	return "game_records"
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/models/game_record.go
git commit -m "refactor(api): game_record model — polymorphic content FK, soft-delete"
```

### Task 2.10: Edit game_session.go

**Files:**
- Modify: `dx-api/app/models/game_session.go`

- [ ] **Step 1: Replace the file contents**

```go
package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

type GameSession struct {
	orm.Timestamps
	orm.SoftDeletes
	ID                    string     `gorm:"column:id;primaryKey" json:"id"`
	UserID                string     `gorm:"column:user_id" json:"user_id"`
	GameID                string     `gorm:"column:game_id" json:"game_id"`
	GameLevelID           string     `gorm:"column:game_level_id" json:"game_level_id"`
	Degree                string     `gorm:"column:degree" json:"degree"`
	Pattern               *string    `gorm:"column:pattern" json:"pattern"`
	CurrentContentItemID  *string    `gorm:"column:current_content_item_id" json:"current_content_item_id"`
	CurrentContentVocabID *string    `gorm:"column:current_content_vocab_id" json:"current_content_vocab_id"`
	StartedAt             time.Time  `gorm:"column:started_at" json:"started_at"`
	LastPlayedAt          time.Time  `gorm:"column:last_played_at" json:"last_played_at"`
	EndedAt               *time.Time `gorm:"column:ended_at" json:"ended_at"`
	Score                 int        `gorm:"column:score" json:"score"`
	Exp                   int        `gorm:"column:exp" json:"exp"`
	MaxCombo              int        `gorm:"column:max_combo" json:"max_combo"`
	CorrectCount          int        `gorm:"column:correct_count" json:"correct_count"`
	WrongCount            int        `gorm:"column:wrong_count" json:"wrong_count"`
	SkipCount             int        `gorm:"column:skip_count" json:"skip_count"`
	PlayTime              int        `gorm:"column:play_time" json:"play_time"`
	TotalItemsCount       int        `gorm:"column:total_items_count" json:"total_items_count"`
	PlayedItemsCount      int        `gorm:"column:played_items_count" json:"played_items_count"`
	GameGroupID           *string    `gorm:"column:game_group_id" json:"game_group_id"`
	GameSubgroupID        *string    `gorm:"column:game_subgroup_id" json:"game_subgroup_id"`
	GamePkID              *string    `gorm:"column:game_pk_id" json:"game_pk_id"`
}

func (g *GameSession) TableName() string {
	return "game_sessions"
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/models/game_session.go
git commit -m "refactor(api): game_session model — add CurrentContentVocabID, soft-delete"
```

### Task 2.11: Edit game_report.go

**Files:**
- Modify: `dx-api/app/models/game_report.go`

- [ ] **Step 1: Replace the file contents**

```go
package models

import "github.com/goravel/framework/database/orm"

type GameReport struct {
	orm.Timestamps
	orm.SoftDeletes
	ID             string  `gorm:"column:id;primaryKey" json:"id"`
	UserID         string  `gorm:"column:user_id" json:"user_id"`
	GameID         string  `gorm:"column:game_id" json:"game_id"`
	GameLevelID    string  `gorm:"column:game_level_id" json:"game_level_id"`
	ContentItemID  *string `gorm:"column:content_item_id" json:"content_item_id"`
	ContentVocabID *string `gorm:"column:content_vocab_id" json:"content_vocab_id"`
	Reason         string  `gorm:"column:reason" json:"reason"`
	Note           *string `gorm:"column:note" json:"note"`
	Count          int     `gorm:"column:count" json:"count"`
}

func (g *GameReport) TableName() string {
	return "game_reports"
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/models/game_report.go
git commit -m "refactor(api): game_report model — polymorphic content FK, soft-delete"
```

### Task 2.12: Delete game_meta.go and game_item.go models

**Files:**
- Delete: `dx-api/app/models/game_meta.go`
- Delete: `dx-api/app/models/game_item.go`

- [ ] **Step 1: Delete the files**

```bash
rm dx-api/app/models/game_meta.go dx-api/app/models/game_item.go
```

- [ ] **Step 2: Verify build fails (expected — services still reference them)**

```bash
cd dx-api && go build ./... 2>&1 | head -30
```

Expected: errors like `undefined: models.GameMeta` / `models.GameItem` from `app/services/api/*.go`. **This is expected** — Phases 4-7 will remove those references.

- [ ] **Step 3: Commit anyway (the deletes are correct; service edits come next)**

```bash
git add dx-api/app/models/game_meta.go dx-api/app/models/game_item.go
git commit -m "refactor(api): delete GameMeta and GameItem models — junctions removed"
```

### Phase 2 validation gate

- [ ] **Run gofmt + vet on models only**

```bash
cd dx-api && gofmt -l app/models/ && go vet ./app/models/...
```

Expected: no output. (Full `go build ./...` will fail until services are updated; that's planned for Phases 4-7.)

---

## Phase 3 — Consts (POS keys)

### Task 3.1: Add pos.go consts

**Files:**
- Create: `dx-api/app/consts/pos.go`

- [ ] **Step 1: Create the file**

```go
package consts

// Part-of-speech keys used in ContentVocab.Definition JSON entries.
//
// Each entry in the definition JSON is a single-key object whose key
// is one of these constants and whose value is the Chinese gloss, e.g.:
//
//   [{"adj": "快的"}, {"v": "斋戒"}]
//
// Validation: IsValidPos rejects unknown keys to keep the wiki canonical.
const (
	PosNoun        = "n"
	PosVerb        = "v"
	PosAdjective   = "adj"
	PosAdverb      = "adv"
	PosPreposition = "prep"
	PosConjunction = "conj"
	PosPronoun     = "pron"
	PosArticle     = "art"
	PosNumeral     = "num"
	PosInterject   = "int"
	PosAuxiliary   = "aux"
	PosDeterminer  = "det"
)

// AllPos lists every supported POS key.
var AllPos = []string{
	PosNoun, PosVerb, PosAdjective, PosAdverb,
	PosPreposition, PosConjunction, PosPronoun, PosArticle,
	PosNumeral, PosInterject, PosAuxiliary, PosDeterminer,
}

// PosLabels maps POS keys to their Chinese labels for UI rendering.
var PosLabels = map[string]string{
	PosNoun:        "名词",
	PosVerb:        "动词",
	PosAdjective:   "形容词",
	PosAdverb:      "副词",
	PosPreposition: "介词",
	PosConjunction: "连词",
	PosPronoun:     "代词",
	PosArticle:     "冠词",
	PosNumeral:     "数词",
	PosInterject:   "感叹词",
	PosAuxiliary:   "助动词",
	PosDeterminer:  "限定词",
}

var posSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(AllPos))
	for _, p := range AllPos {
		m[p] = struct{}{}
	}
	return m
}()

// IsValidPos returns true if s is one of the canonical POS keys.
func IsValidPos(s string) bool {
	_, ok := posSet[s]
	return ok
}
```

- [ ] **Step 2: Commit**

```bash
git add dx-api/app/consts/pos.go
git commit -m "feat(api): add POS consts (12-key set) for ContentVocab definitions"
```

### Task 3.2: Add pos_test.go

**Files:**
- Create: `dx-api/app/consts/pos_test.go`

- [ ] **Step 1: Create the test file**

```go
package consts

import "testing"

func TestIsValidPos(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"n", true},
		{"v", true},
		{"adj", true},
		{"adv", true},
		{"prep", true},
		{"conj", true},
		{"pron", true},
		{"art", true},
		{"num", true},
		{"int", true},
		{"aux", true},
		{"det", true},
		{"verb", false},
		{"adjective", false},
		{"", false},
		{"N", false},
		{"phr", false},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			if got := IsValidPos(tc.input); got != tc.want {
				t.Errorf("IsValidPos(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestAllPosCount(t *testing.T) {
	if len(AllPos) != 12 {
		t.Errorf("expected 12 POS keys, got %d", len(AllPos))
	}
	if len(PosLabels) != len(AllPos) {
		t.Errorf("PosLabels (%d) and AllPos (%d) must be the same size", len(PosLabels), len(AllPos))
	}
}
```

- [ ] **Step 2: Run the test**

```bash
cd dx-api && go test -race ./app/consts/... -run "TestIsValidPos|TestAllPosCount" -v
```

Expected: PASS for all subtests.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/consts/pos_test.go
git commit -m "test(api): cover IsValidPos and AllPos invariants"
```

### Phase 3 validation gate

- [ ] **Run gofmt + vet + test**

```bash
cd dx-api && gofmt -l app/consts/ && go vet ./app/consts/... && go test -race ./app/consts/...
```

Expected: no gofmt output, no vet errors, all tests pass.

---

## Phase 4 — Word-sentence backend rewrite

Rewrites three service files to drop junction joins and dedup logic. Use direct queries against `content_metas` / `content_items` with their new `(game_id, game_level_id, order)` columns.

### Task 4.1: Rewrite course_content_service.go

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go`

This is a large refactor. The new file is ~600 lines (down from 1058). Key changes:
- Remove `metaDedupKey`, `existingMetaRef`, `findExistingMetasForBatch`, `reuseItemsIntoLevel`, all `itemsByMetaCache` plumbing.
- `SaveMetadataBatch` becomes a plain insert loop.
- Read paths query `content_metas` / `content_items` directly with `WHERE game_level_id = ?`.
- Delete paths soft-delete the rows directly with no orphan-check.

- [ ] **Step 1: Replace the file contents**

```go
package api

import (
	"encoding/json"
	"fmt"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

// MetadataEntry represents a single entry in a batch metadata creation request.
type MetadataEntry struct {
	SourceData  string  `json:"sourceData"`
	Translation *string `json:"translation"`
	SourceType  string  `json:"sourceType"`
}

// CourseContentItemData represents a content item returned to the client.
type CourseContentItemData struct {
	ID            string          `json:"id"`
	ContentMetaID *string         `json:"contentMetaId"`
	Content       string          `json:"content"`
	ContentType   string          `json:"contentType"`
	Translation   *string         `json:"translation"`
	Items         json.RawMessage `json:"items"`
	Order         float64         `json:"order"`
}

// LevelContentData groups content items by their metadata.
type LevelContentData struct {
	Meta  ContentMetaData         `json:"meta"`
	Items []CourseContentItemData `json:"items"`
}

// ContentMetaData represents content metadata returned to the client.
type ContentMetaData struct {
	ID          string  `json:"id"`
	SourceFrom  string  `json:"sourceFrom"`
	SourceType  string  `json:"sourceType"`
	SourceData  string  `json:"sourceData"`
	Translation *string `json:"translation"`
	IsBreakDone bool    `json:"isBreakDone"`
	Order       float64 `json:"order"`
}

// SaveMetadataBatch creates content metadata entries in batch with capacity validation.
// No dedup: every entry becomes a fresh content_metas row. Reordering is by (game_level_id, order).
func SaveMetadataBatch(userID, gameID, gameLevelID string, entries []MetadataEntry, sourceFrom string) (int, error) {
	if err := requireVip(userID); err != nil {
		return 0, err
	}
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return 0, err
	}

	if game.Status == consts.GameStatusPublished {
		return 0, ErrGamePublished
	}

	// Verify level belongs to game
	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return 0, fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return 0, ErrLevelNotFound
	}

	// Capacity check + max order pulled from content_metas directly.
	type existingMetaRow struct {
		SourceType string  `gorm:"column:source_type"`
		MetaOrder  float64 `gorm:"column:meta_order"`
	}
	var existing []existingMetaRow
	if err := facades.Orm().Query().Raw(
		`SELECT source_type, "order" AS meta_order
		   FROM content_metas
		  WHERE game_level_id = ? AND deleted_at IS NULL`,
		gameLevelID,
	).Scan(&existing); err != nil {
		return 0, fmt.Errorf("failed to count metas: %w", err)
	}

	if consts.IsVocabMode(game.Mode) {
		// (NOTE: vocab modes use content_vocabs / game_vocabs in Phase 5; this
		// branch keeps the legacy capacity check for safety in case word-sentence
		// metadata is somehow still saved into a vocab game during migration.)
		if len(existing)+len(entries) > consts.MaxMetasPerLevel {
			return 0, ErrCapacityExceeded
		}
		batchSize := consts.VocabBatchSize(game.Mode)
		if batchSize > 0 && len(entries)%batchSize != 0 {
			return 0, ErrBatchSizeInvalid
		}
	} else {
		existingSentences := 0
		existingVocabs := 0
		for _, m := range existing {
			switch m.SourceType {
			case SourceTypeSentence:
				existingSentences++
			case SourceTypeVocab:
				existingVocabs++
			}
		}
		newSentences := 0
		newVocabs := 0
		for _, e := range entries {
			switch e.SourceType {
			case SourceTypeSentence:
				newSentences++
			case SourceTypeVocab:
				newVocabs++
			}
		}
		totalSentences := existingSentences + newSentences
		totalVocabs := existingVocabs + newVocabs
		if float64(totalSentences)/float64(MaxSentences)+float64(totalVocabs)/float64(MaxVocab) > 1 {
			return 0, ErrCapacityExceeded
		}
	}

	maxOrder := float64(0)
	for _, m := range existing {
		if m.MetaOrder > maxOrder {
			maxOrder = m.MetaOrder
		}
	}

	if err := facades.Orm().Transaction(func(tx orm.Query) error {
		for i, e := range entries {
			meta := models.ContentMeta{
				ID:          uuid.Must(uuid.NewV7()).String(),
				GameID:      gameID,
				GameLevelID: gameLevelID,
				SourceFrom:  sourceFrom,
				SourceType:  e.SourceType,
				SourceData:  e.SourceData,
				Translation: e.Translation,
				IsBreakDone: false,
				Order:       maxOrder + float64((i+1)*1000),
			}
			if err := tx.Create(&meta); err != nil {
				return fmt.Errorf("failed to create content meta: %w", err)
			}
		}
		return nil
	}); err != nil {
		return 0, err
	}

	return len(entries), nil
}

// ReorderMetadata updates the order of a content metadata entry within a level.
func ReorderMetadata(userID, gameID, gameLevelID, metaID string, newOrder float64) error {
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
	if err := verifyMetaBelongsToGame(metaID, gameID); err != nil {
		return err
	}
	if _, err := facades.Orm().Query().Exec(
		`UPDATE content_metas SET "order" = ?
		   WHERE id = ? AND game_level_id = ? AND deleted_at IS NULL`,
		newOrder, metaID, gameLevelID,
	); err != nil {
		return fmt.Errorf("failed to reorder metadata: %w", err)
	}
	return nil
}

// GetContentItemsByMeta returns content items grouped by their metadata for a given level.
func GetContentItemsByMeta(userID, gameID, gameLevelID string) ([]LevelContentData, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	if _, err := getCourseGameOwned(userID, gameID); err != nil {
		return nil, err
	}

	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return nil, fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return nil, ErrLevelNotFound
	}

	var contentMetas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Order(`"order" ASC`).
		Get(&contentMetas); err != nil {
		return nil, fmt.Errorf("failed to load content_metas: %w", err)
	}
	if len(contentMetas) == 0 {
		return []LevelContentData{}, nil
	}

	var items []models.ContentItem
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Order(`"order" ASC`).
		Get(&items); err != nil {
		return nil, fmt.Errorf("failed to load content_items: %w", err)
	}

	itemsByMeta := make(map[string][]CourseContentItemData)
	for _, it := range items {
		metaKey := ""
		if it.ContentMetaID != nil {
			metaKey = *it.ContentMetaID
		}
		raw := json.RawMessage("null")
		if it.Items != nil {
			raw = json.RawMessage(*it.Items)
		}
		itemsByMeta[metaKey] = append(itemsByMeta[metaKey], CourseContentItemData{
			ID:            it.ID,
			ContentMetaID: it.ContentMetaID,
			Content:       it.Content,
			ContentType:   it.ContentType,
			Translation:   it.Translation,
			Items:         raw,
			Order:         it.Order,
		})
	}

	result := make([]LevelContentData, 0, len(contentMetas))
	for _, cm := range contentMetas {
		result = append(result, LevelContentData{
			Meta: ContentMetaData{
				ID:          cm.ID,
				SourceFrom:  cm.SourceFrom,
				SourceType:  cm.SourceType,
				SourceData:  cm.SourceData,
				Translation: cm.Translation,
				IsBreakDone: cm.IsBreakDone,
				Order:       cm.Order,
			},
			Items: itemsByMeta[cm.ID],
		})
	}
	return result, nil
}

// InsertContentItem inserts a content item at a calculated position.
func InsertContentItem(userID, gameID, gameLevelID, contentMetaID string, content, contentType string, translation *string, referenceItemID, direction string) (*CourseContentItemData, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	game, err := getCourseGameOwned(userID, gameID)
	if err != nil {
		return nil, err
	}
	if game.Status == consts.GameStatusPublished {
		return nil, ErrGamePublished
	}

	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return nil, fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return nil, ErrLevelNotFound
	}

	if err := verifyMetaBelongsToGame(contentMetaID, gameID); err != nil {
		return nil, err
	}
	if referenceItemID != "" {
		if err := verifyItemBelongsToGame(referenceItemID, gameID); err != nil {
			return nil, err
		}
	}

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

	order, err := calculateInsertionOrder(gameLevelID, referenceItemID, direction)
	if err != nil {
		return nil, err
	}

	id := uuid.Must(uuid.NewV7()).String()
	item := models.ContentItem{
		ID:            id,
		GameID:        gameID,
		GameLevelID:   gameLevelID,
		ContentMetaID: &contentMetaID,
		Content:       content,
		ContentType:   contentType,
		Translation:   translation,
		Order:         order,
	}
	if err := facades.Orm().Query().Create(&item); err != nil {
		return nil, fmt.Errorf("failed to create content item: %w", err)
	}

	return &CourseContentItemData{
		ID:            id,
		ContentMetaID: &contentMetaID,
		Content:       content,
		ContentType:   contentType,
		Translation:   translation,
		Items:         nil,
		Order:         order,
	}, nil
}

// UpdateContentItemText updates the text and translation of a content item.
func UpdateContentItemText(userID, gameID, itemID, content string, translation *string) error {
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
	if _, err := facades.Orm().Query().Model(&models.ContentItem{}).Where("id", itemID).Update(map[string]any{
		"content":     content,
		"translation": translation,
	}); err != nil {
		return fmt.Errorf("failed to update content item: %w", err)
	}
	return nil
}

// ReorderContentItems updates the order of a content item within a level.
func ReorderContentItems(userID, gameID, gameLevelID, itemID string, newOrder float64) error {
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
	if _, err := facades.Orm().Query().Exec(
		`UPDATE content_items SET "order" = ?
		   WHERE id = ? AND game_level_id = ? AND deleted_at IS NULL`,
		newOrder, itemID, gameLevelID,
	); err != nil {
		return fmt.Errorf("failed to reorder content item: %w", err)
	}
	return nil
}

// DeleteContentItem soft-deletes a single item.
func DeleteContentItem(userID, gameID, gameLevelID, itemID string) error {
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

	var item models.ContentItem
	if err := facades.Orm().Query().Where("id", itemID).First(&item); err != nil {
		return fmt.Errorf("failed to load content item: %w", err)
	}
	if item.ID == "" {
		return ErrContentItemNotFound
	}

	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Exec(
			`UPDATE content_items SET deleted_at = NOW()
			  WHERE id = ? AND game_level_id = ? AND deleted_at IS NULL`,
			itemID, gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_item: %w", err)
		}

		// Reset is_break_done if the meta has no remaining live items in this level.
		if item.ContentMetaID != nil {
			if _, err := tx.Exec(
				`UPDATE content_metas SET is_break_done = false
				  WHERE id = ?
				    AND deleted_at IS NULL
				    AND NOT EXISTS (
				      SELECT 1 FROM content_items
				       WHERE content_meta_id = content_metas.id
				         AND game_level_id = ?
				         AND deleted_at IS NULL
				    )`,
				*item.ContentMetaID, gameLevelID,
			); err != nil {
				return fmt.Errorf("failed to reset meta break status: %w", err)
			}
		}
		return nil
	})
}

// DeleteAllLevelContent soft-deletes every content_meta and content_item in a level.
func DeleteAllLevelContent(userID, gameID, gameLevelID string) error {
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

	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return ErrLevelNotFound
	}

	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Exec(
			`UPDATE content_items SET deleted_at = NOW()
			  WHERE game_level_id = ? AND deleted_at IS NULL`,
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_items: %w", err)
		}
		if _, err := tx.Exec(
			`UPDATE content_metas SET deleted_at = NOW()
			  WHERE game_level_id = ? AND deleted_at IS NULL`,
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_metas: %w", err)
		}
		return nil
	})
}

// DeleteMetadata soft-deletes a meta plus all its items in this level.
func DeleteMetadata(userID, gameID, gameLevelID, metaID string) error {
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
	if err := verifyMetaBelongsToGame(metaID, gameID); err != nil {
		return err
	}

	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Exec(
			`UPDATE content_items SET deleted_at = NOW()
			  WHERE content_meta_id = ?
			    AND game_level_id = ?
			    AND deleted_at IS NULL`,
			metaID, gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_items: %w", err)
		}
		if _, err := tx.Exec(
			`UPDATE content_metas SET deleted_at = NOW()
			  WHERE id = ?
			    AND game_level_id = ?
			    AND deleted_at IS NULL`,
			metaID, gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_meta: %w", err)
		}
		return nil
	})
}

// verifyMetaBelongsToGame checks that a content meta belongs to a game.
func verifyMetaBelongsToGame(metaID, gameID string) error {
	n, err := facades.Orm().Query().Model(&models.ContentMeta{}).
		Where("id", metaID).
		Where("game_id", gameID).
		Count()
	if err != nil {
		return fmt.Errorf("failed to verify meta: %w", err)
	}
	if n == 0 {
		return ErrMetaNotFound
	}
	return nil
}

// verifyItemBelongsToGame checks that a content item belongs to a game.
func verifyItemBelongsToGame(itemID, gameID string) error {
	count, err := facades.Orm().Query().Model(&models.ContentItem{}).
		Where("id", itemID).
		Where("game_id", gameID).
		Count()
	if err != nil {
		return fmt.Errorf("failed to verify item: %w", err)
	}
	if count == 0 {
		return ErrContentItemNotFound
	}
	return nil
}

// calculateInsertionOrder computes the order for a new item relative to a reference item.
func calculateInsertionOrder(gameLevelID, referenceItemID, direction string) (float64, error) {
	if referenceItemID == "" {
		var lastItem models.ContentItem
		if err := facades.Orm().Query().
			Where("game_level_id", gameLevelID).
			Order(`"order" DESC`).
			First(&lastItem); err != nil || lastItem.ID == "" {
			return 1000, nil
		}
		return lastItem.Order + 1000, nil
	}

	var refItem models.ContentItem
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("id", referenceItemID).
		First(&refItem); err != nil {
		return 0, fmt.Errorf("failed to find reference item: %w", err)
	}
	if refItem.ID == "" {
		return 0, ErrContentItemNotFound
	}

	var items []models.ContentItem
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Order(`"order" ASC`).
		Get(&items); err != nil {
		return 0, fmt.Errorf("failed to load items: %w", err)
	}

	refIdx := -1
	for i, item := range items {
		if item.ID == referenceItemID {
			refIdx = i
			break
		}
	}
	if refIdx == -1 {
		return refItem.Order + 1000, nil
	}

	if direction == "above" || direction == "before" {
		if refIdx == 0 {
			return refItem.Order / 2, nil
		}
		prevOrder := items[refIdx-1].Order
		return (prevOrder + refItem.Order) / 2, nil
	}

	if refIdx == len(items)-1 {
		return refItem.Order + 1000, nil
	}
	nextOrder := items[refIdx+1].Order
	return (refItem.Order + nextOrder) / 2, nil
}
```

- [ ] **Step 2: Verify formatting**

```bash
cd dx-api && gofmt -l app/services/api/course_content_service.go
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/course_content_service.go
git commit -m "refactor(api): rewrite course_content_service to query content_metas/items directly"
```

### Task 4.2: Rewrite ai_custom_service.go BreakMetadata + GenerateContentItems

**Files:**
- Modify: `dx-api/app/services/api/ai_custom_service.go`

The bottom half of this file (functions `BreakMetadata`, `processBreakMeta`, `GenerateContentItems`, `processGenItems`, plus the helpers below) needs to drop game_metas/game_items joins and write directly. The top half (GenerateMetadata, FormatMetadata, parseFormattedLines, prompts, error helpers) is **unchanged**.

- [ ] **Step 1: Replace the file contents** (full file)

```go
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
)

// AI generation cost consts.
const aiGenerateCost = 5

// Concurrency limits for SSE batch operations.
const (
	breakConcurrencyLimit    = 20
	genItemsConcurrencyLimit = 50
)

// Difficulty level descriptions (CEFR).
var difficultyDescriptions = map[string]string{
	"a1-a2": "A1-A2 (beginner: simple vocabulary, short sentences, present tense, common everyday words)",
	"b1-b2": "B1-B2 (intermediate: varied vocabulary, compound sentences, multiple tenses, some idiomatic expressions)",
	"c1-c2": "C1-C2 (advanced: sophisticated vocabulary, complex sentence structures, nuanced expressions, academic/literary language)",
}

// Error sentinels for AI custom operations.
var (
	ErrAIServiceUnavailable = errors.New("AI service unavailable")
	ErrModerationWarning    = errors.New("moderation warning")
	ErrEmptyContent         = errors.New("content is empty")
	ErrFormatCountExceeded  = errors.New("format count exceeded")
)

// GenerateMetadataResult holds the response from story generation.
type GenerateMetadataResult struct {
	Generated  string `json:"generated,omitempty"`
	SourceType string `json:"sourceType,omitempty"`
	Warning    string `json:"warning,omitempty"`
}

// FormatMetadataResult holds the response from content formatting.
type FormatMetadataResult struct {
	Formatted   string   `json:"formatted,omitempty"`
	SourceTypes []string `json:"sourceTypes,omitempty"`
	Warning     string   `json:"warning,omitempty"`
}

// SSEProgressEvent is sent to the client during SSE streaming.
type SSEProgressEvent struct {
	Done      int    `json:"done"`
	Total     int    `json:"total"`
	Status    string `json:"status,omitempty"`
	Processed int    `json:"processed,omitempty"`
	Failed    int    `json:"failed,omitempty"`
	Complete  bool   `json:"complete,omitempty"`
}

// --- GenerateMetadata --- (unchanged from original)

func GenerateMetadata(userID string, difficulty string, keywords []string) (*GenerateMetadataResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	if err := ConsumeBeans(userID, aiGenerateCost, consts.BeanSlugAIGenerateConsume, consts.BeanReasonAIGenerateConsume); err != nil {
		return nil, err
	}

	levelDesc, ok := difficultyDescriptions[difficulty]
	if !ok {
		levelDesc = difficultyDescriptions["a1-a2"]
	}

	prompt := buildGeneratePrompt(levelDesc)
	userMsg := "Keywords: " + strings.Join(keywords, ", ")

	result, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
		Messages: []helpers.DeepSeekMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: userMsg},
		},
		Temperature: 0.7,
	})
	if err != nil {
		_ = RefundBeans(userID, aiGenerateCost, consts.BeanSlugAIGenerateRefund, consts.BeanReasonAIGenerateRefund)
		return nil, err
	}

	if rest, ok := strings.CutPrefix(result, "WARNING:"); ok {
		return &GenerateMetadataResult{Warning: strings.TrimSpace(rest)}, nil
	}

	return &GenerateMetadataResult{
		Generated:  result,
		SourceType: SourceTypeSentence,
	}, nil
}

func buildGeneratePrompt(levelDesc string) string {
	return `You are a story writer for an English learning app. Your job is to generate a short English story for language learners.

STEP 1 — CONTENT MODERATION (do this FIRST):
Check if the provided keywords contain any insulting, violent, sexually explicit, or otherwise inappropriate/sensitive material.
If they do, respond ONLY with: WARNING:包含不适当内容，请修改后重试
Do NOT generate any story. Stop here.

STEP 2 — GENERATE STORY:
Write a short, coherent English story that:
- Uses CEFR level ` + levelDesc + ` appropriate vocabulary and grammar
- Naturally incorporates ALL the provided keywords into the story
- Contains at most 20 sentences. It can be fewer than 20 but NEVER more than 20. This is a hard limit.
- Tells a complete, engaging narrative with a beginning, middle, and end
- Is suitable for English language learners

RULES:
- Output ONLY the story text. No title, no explanations, no headers, no numbering, no markdown.
- Each sentence must be on its own line.
- Each line must contain exactly ONE sentence.
- Do not include empty lines between sentences.
- Do not repeat sentences.
- Keep each sentence under 200 characters.`
}

// --- FormatMetadata --- (unchanged from original — preserves [S]/[V] mixed input)

func FormatMetadata(userID string, content string, formatType string) (*FormatMetadataResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	wordCount := helpers.CountWords(content)
	if wordCount == 0 {
		return nil, ErrEmptyContent
	}

	consumeSlug, consumeReason, refundSlug, refundReason := formatBeanSlugs(formatType)

	if err := ConsumeBeans(userID, wordCount, consumeSlug, consumeReason); err != nil {
		return nil, err
	}

	prompt := buildFormatPrompt(formatType)

	result, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
		Messages: []helpers.DeepSeekMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: content},
		},
		Temperature: 0.1,
	})
	if err != nil {
		_ = RefundBeans(userID, wordCount, refundSlug, refundReason)
		return nil, err
	}

	if rest, ok := strings.CutPrefix(result, "WARNING:"); ok {
		return &FormatMetadataResult{Warning: strings.TrimSpace(rest)}, nil
	}

	formatted, sourceTypes := parseFormattedLines(result)
	if formatted == "" {
		_ = RefundBeans(userID, wordCount, refundSlug, refundReason)
		return nil, helpers.ErrDeepSeekEmpty
	}

	if warning := validateFormatCounts(sourceTypes); warning != "" {
		return &FormatMetadataResult{Warning: warning}, nil
	}

	return &FormatMetadataResult{
		Formatted:   formatted,
		SourceTypes: sourceTypes,
	}, nil
}

func formatBeanSlugs(formatType string) (consumeSlug, consumeReason, refundSlug, refundReason string) {
	if formatType == SourceTypeSentence {
		return consts.BeanSlugAIFormatSentenceConsume, consts.BeanReasonAIFormatSentenceConsume,
			consts.BeanSlugAIFormatSentenceRefund, consts.BeanReasonAIFormatSentenceRefund
	}
	return consts.BeanSlugAIFormatVocabConsume, consts.BeanReasonAIFormatVocabConsume,
		consts.BeanSlugAIFormatVocabRefund, consts.BeanReasonAIFormatVocabRefund
}

func buildFormatPrompt(formatType string) string {
	formatLabel := "词汇"
	if formatType == SourceTypeSentence {
		formatLabel = "语句"
	}

	formatRule := `- If the content contains Chinese text: output alternating lines of English word/phrase followed by its Chinese translation.
- If the content contains NO Chinese text: output English words/phrases only, one per line.`
	mismatchRule := "If the content consists mostly of full sentences with punctuation, respond ONLY with: WARNING:内容看起来是语句而非词汇，请使用「语句格式化」按钮"

	if formatType == SourceTypeSentence {
		formatRule = `- If the content contains Chinese text: output alternating lines of English sentence followed by its Chinese translation.
- If the content contains NO Chinese text: output English sentences only, one per line.`
		mismatchRule = "If the content consists mostly of single words or short phrases without sentence structure, respond ONLY with: WARNING:内容看起来是词汇而非语句，请使用「词汇格式化」按钮"
	}

	return `You are a content formatter for an English learning app. Your job is to clean up and reformat messy user input into a strict line-by-line format for ` + formatLabel + `.

STEP 1 — CONTENT MODERATION (do this FIRST):
Check if the content contains any insulting, violent, sexually explicit, or otherwise inappropriate/sensitive material.
If it does, respond ONLY with: WARNING:内容包含不适当内容，请修改后重试
Do NOT format the content. Stop here.

STEP 2 — TYPE MISMATCH CHECK:
` + mismatchRule + `

STEP 3 — FORMAT WITH PER-LINE TYPE MARKERS:
` + formatRule + `

For EACH English line, prefix it with a type marker:
- [S] for complete sentences (has subject + verb, expresses a complete thought)
- [V] for vocabulary items (single words, short phrases, or expressions without sentence structure)

Chinese translation lines must NOT have any prefix marker.

Example output with Chinese translations:
[S] I like the food.
我喜欢这个食物。
[V] food
食物
[V] name
名字

Example output without Chinese translations:
[S] I like the food.
[V] food
[V] name

RULES:
- Every English line MUST start with [S] or [V] prefix.
- Chinese translation lines must NOT have any prefix.
- Output ONLY the formatted text with markers. No explanations, headers, numbering, or markdown.
- Remove duplicates.
- Fix obvious spelling errors in English.
- Preserve the original meaning. Do not add or remove content.
- CRITICAL: Each line must contain exactly ONE sentence (or one word/phrase or one translation). Sentences are often separated by punctuation like periods (.), question marks (?), or exclamation marks (!), but not always — some sentences have no ending punctuation. Use meaning and grammar to identify sentence boundaries. If the input has multiple sentences on one line, split them so each sentence is on its own line. Never combine two or more sentences on the same line.
- Remove any empty lines.`
}

func parseFormattedLines(result string) (string, []string) {
	lines := strings.Split(result, "\n")
	var cleanLines []string
	var sourceTypes []string

	for _, line := range lines {
		line = strings.TrimRight(line, " \t\r")
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[S] ") {
			sourceTypes = append(sourceTypes, SourceTypeSentence)
			cleanLines = append(cleanLines, line[4:])
		} else if strings.HasPrefix(line, "[V] ") {
			sourceTypes = append(sourceTypes, SourceTypeVocab)
			cleanLines = append(cleanLines, line[4:])
		} else {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n"), sourceTypes
}

func validateFormatCounts(sourceTypes []string) string {
	sentenceCount := 0
	vocabCount := 0
	for _, t := range sourceTypes {
		switch t {
		case SourceTypeSentence:
			sentenceCount++
		case SourceTypeVocab:
			vocabCount++
		}
	}

	if sentenceCount > MaxSentences {
		return fmt.Sprintf("格式化后有 %d 条语句，超过 %d 条上限。为保证最佳学习体验，请精简内容", sentenceCount, MaxSentences)
	}
	if vocabCount > MaxVocab {
		return fmt.Sprintf("格式化后有 %d 条词汇，超过 %d 条上限。请精简内容", vocabCount, MaxVocab)
	}
	return ""
}

// --- BreakMetadata --- (rewritten — direct queries on content_metas/content_items)

func BreakMetadata(userID, gameLevelID string, writer *helpers.NDJSONWriter) {
	if err := requireVip(userID); err != nil {
		writeSSEError(writer, err)
		return
	}
	game, level, err := verifyLevelOwnership(userID, gameLevelID)
	if err != nil {
		writeSSEError(writer, err)
		return
	}
	if game.Status == consts.GameStatusPublished {
		writeSSEError(writer, ErrGamePublished)
		return
	}
	gameID := level.GameID

	// Load unbroken metas in order — directly from content_metas
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("is_break_done", false).
		Order(`"order" ASC`).
		Get(&metas); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
		return
	}

	if len(metas) == 0 {
		_ = writer.Write(SSEProgressEvent{Done: 0, Total: 0, Processed: 0, Failed: 0, Complete: true})
		writer.Close()
		return
	}

	metaWordCounts := make([]int, len(metas))
	totalCost := 0
	for i, m := range metas {
		wc := helpers.CountWords(m.SourceData)
		metaWordCounts[i] = wc
		totalCost += wc
	}

	if totalCost == 0 {
		writeSSEError(writer, ErrEmptyContent)
		return
	}

	if err := ConsumeBeans(userID, totalCost, consts.BeanSlugAIBreakConsume, consts.BeanReasonAIBreakConsume); err != nil {
		writeSSEError(writer, err)
		return
	}

	var failedWords int64
	var processed int64
	var failed int64

	sem := make(chan struct{}, breakConcurrencyLimit)
	var wg sync.WaitGroup
	var done int64

	total := len(metas)

	for i, meta := range metas {
		wg.Add(1)
		sem <- struct{}{}

		go func(m models.ContentMeta, idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			success := processBreakMeta(m, gameID, gameLevelID)
			d := atomic.AddInt64(&done, 1)

			if success {
				atomic.AddInt64(&processed, 1)
				_ = writer.Write(SSEProgressEvent{Done: int(d), Total: total, Status: "ok"})
			} else {
				atomic.AddInt64(&failed, 1)
				atomic.AddInt64(&failedWords, int64(metaWordCounts[idx]))
				_ = writer.Write(SSEProgressEvent{Done: int(d), Total: total, Status: "failed"})
			}
		}(meta, i)
	}

	wg.Wait()

	fw := int(atomic.LoadInt64(&failedWords))
	if fw > 0 {
		_ = RefundBeans(userID, fw, consts.BeanSlugAIBreakRefund, consts.BeanReasonAIBreakRefund)
	}

	_ = writer.Write(SSEProgressEvent{
		Done:      total,
		Total:     total,
		Processed: int(atomic.LoadInt64(&processed)),
		Failed:    int(atomic.LoadInt64(&failed)),
		Complete:  true,
	})
	writer.Close()
}

// processBreakMeta calls DeepSeek to split a meta into content_items rows.
// Item orders fan out from the parent meta's order in increments of 10.
func processBreakMeta(meta models.ContentMeta, gameID, gameLevelID string) bool {
	userMsg := "English: " + meta.SourceData
	if meta.Translation != nil && *meta.Translation != "" {
		userMsg += "\nChinese translation: " + *meta.Translation
	}

	result, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
		Messages: []helpers.DeepSeekMessage{
			{Role: "system", Content: breakPrompt},
			{Role: "user", Content: userMsg},
		},
		Temperature: 0.1,
	})
	if err != nil {
		return false
	}

	items, err := helpers.ParseAIJSONArray(result)
	if err != nil || len(items) == 0 {
		return false
	}

	startOrder := meta.Order + 10

	for i, raw := range items {
		var unit struct {
			Content     string `json:"content"`
			ContentType string `json:"contentType"`
			Translation string `json:"translation"`
		}
		if err := json.Unmarshal(raw, &unit); err != nil {
			continue
		}

		id := uuid.Must(uuid.NewV7()).String()
		metaID := meta.ID
		var translation *string
		if unit.Translation != "" {
			translation = &unit.Translation
		}

		item := models.ContentItem{
			ID:            id,
			GameID:        gameID,
			GameLevelID:   gameLevelID,
			ContentMetaID: &metaID,
			Content:       unit.Content,
			ContentType:   unit.ContentType,
			Translation:   translation,
			Order:         startOrder + float64(i*10),
		}
		if err := facades.Orm().Query().Create(&item); err != nil {
			return false
		}
	}

	if _, err := facades.Orm().Query().Model(&models.ContentMeta{}).
		Where("id", meta.ID).
		Update("is_break_done", true); err != nil {
		return false
	}

	return true
}

var breakPrompt = `You are a language learning content processor. Your job is to analyze an English text and break it into structured learning units.

STEP 1 — DETERMINE TYPE:
- If the input is a complete sentence (has subject + verb, typically ends with punctuation like . ? !), treat it as a SENTENCE.
- Otherwise, treat it as a WORD or PHRASE.

STEP 2 — GENERATE LEARNING UNITS:

If SENTENCE (sequential left-to-right splitting):
Split the sentence from left to right, producing units in reading order. Use these types:
- "word": a single content word (noun, verb, adjective, adverb, pronoun)
- "block": a progressive combination building from the start of the sentence
- "phrase": a natural word grouping (collocation, prepositional phrase, noun phrase, etc.)
- "sentence": the full original sentence

IMPORTANT RULES:
- Articles (a, an, the) and prepositions (e.g. in, on, at, to, for, with, of, by, from, about, into, through, between, etc.) must NEVER be standalone "word" units. Always group them into the nearest phrase or block with the following content word(s).
- Linking verbs (am, is, are, was, were, be, been, being) must NEVER be standalone "word" units. Always group them with adjacent content — e.g. "is tall" (phrase), "He is" (block), "is reading" (phrase).
- Follow the natural reading order of the sentence. Do NOT group all words first, then all blocks, then all phrases. Instead, split sequentially left to right.
- Each segment of the sentence should appear in exactly one unit at its most granular level (word or phrase), then optionally in cumulative blocks and finally the full sentence.

Example for "I like the food.":
1. "I" (word)
2. "like" (word)
3. "I like" (block)
4. "the food" (phrase) — article grouped with its noun
5. "I like the food." (sentence)

Example for "She went to the park.":
1. "She" (word)
2. "went" (word)
3. "She went" (block)
4. "to the park" (phrase) — preposition + article grouped with noun
5. "She went to the park." (sentence)

Example for "He is reading a book in the library.":
1. "He" (word)
2. "is reading" (phrase) — auxiliary + verb grouped together
3. "He is reading" (block)
4. "a book" (phrase) — article grouped with noun
5. "He is reading a book" (block)
6. "in the library" (phrase) — preposition + article + noun grouped
7. "He is reading a book in the library." (sentence)

If NOT a sentence:
Generate a single unit with contentType "word" (single word) or "phrase" (multi-word expression) or "block" (neither a valid word nor a valid phrase).

Each unit needs:
- content: the text of this unit
- contentType: one of "word", "block", "phrase", "sentence"
- translation: Chinese translation of the entire unit

OUTPUT FORMAT:
Output ONLY a valid JSON array. No markdown code fences, no explanation, no extra text.

Example output:
[
  {"content": "I", "contentType": "word", "translation": "我"},
  {"content": "like", "contentType": "word", "translation": "喜欢"},
  {"content": "I like", "contentType": "block", "translation": "我喜欢"},
  {"content": "the food", "contentType": "phrase", "translation": "这食物"},
  {"content": "I like the food.", "contentType": "sentence", "translation": "我喜欢这食物。"}
]`

// --- GenerateContentItems --- (rewritten — direct queries)

func GenerateContentItems(userID, gameLevelID string, writer *helpers.NDJSONWriter) {
	if err := requireVip(userID); err != nil {
		writeSSEError(writer, err)
		return
	}
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

	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("is_break_done", true).
		Order(`"order" ASC`).
		Get(&metas); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
		return
	}
	if len(metas) == 0 {
		_ = writer.Write(SSEProgressEvent{Done: 0, Total: 0, Processed: 0, Failed: 0, Complete: true})
		writer.Close()
		return
	}

	metaIDs := make([]string, 0, len(metas))
	metaMap := make(map[string]models.ContentMeta)
	for _, m := range metas {
		metaIDs = append(metaIDs, m.ID)
		metaMap[m.ID] = m
	}

	var pendingItems []models.ContentItem
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("content_meta_id IN ?", metaIDs).
		Where("items IS NULL").
		Get(&pendingItems); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load pending items: %w", err))
		return
	}

	pendingByMeta := make(map[string][]models.ContentItem)
	metaItemWordCounts := make(map[string]int)
	for _, item := range pendingItems {
		if item.ContentMetaID == nil {
			continue
		}
		mid := *item.ContentMetaID
		pendingByMeta[mid] = append(pendingByMeta[mid], item)
		metaItemWordCounts[mid] += helpers.CountWords(item.Content)
	}

	var activeMetas []models.ContentMeta
	for _, m := range metas {
		if len(pendingByMeta[m.ID]) > 0 {
			activeMetas = append(activeMetas, m)
		}
	}
	if len(activeMetas) == 0 {
		_ = writer.Write(SSEProgressEvent{Done: 0, Total: 0, Processed: 0, Failed: 0, Complete: true})
		writer.Close()
		return
	}

	totalCost := 0
	for _, wc := range metaItemWordCounts {
		totalCost += wc
	}
	if totalCost == 0 {
		writeSSEError(writer, ErrEmptyContent)
		return
	}

	if err := ConsumeBeans(userID, totalCost, consts.BeanSlugAIGenItemsConsume, consts.BeanReasonAIGenItemsConsume); err != nil {
		writeSSEError(writer, err)
		return
	}

	var failedWords int64
	var processed int64
	var failed int64

	sem := make(chan struct{}, genItemsConcurrencyLimit)
	var wg sync.WaitGroup
	var done int64

	total := len(activeMetas)

	for _, meta := range activeMetas {
		wg.Add(1)
		sem <- struct{}{}

		go func(m models.ContentMeta) {
			defer wg.Done()
			defer func() { <-sem }()

			items := pendingByMeta[m.ID]
			success := processGenItems(m, items)
			d := atomic.AddInt64(&done, 1)

			if success {
				atomic.AddInt64(&processed, 1)
				_ = writer.Write(SSEProgressEvent{Done: int(d), Total: total, Status: "ok"})
			} else {
				atomic.AddInt64(&failed, 1)
				atomic.AddInt64(&failedWords, int64(metaItemWordCounts[m.ID]))
				_ = writer.Write(SSEProgressEvent{Done: int(d), Total: total, Status: "failed"})
			}
		}(meta)
	}

	wg.Wait()

	fw := int(atomic.LoadInt64(&failedWords))
	if fw > 0 {
		_ = RefundBeans(userID, fw, consts.BeanSlugAIGenItemsRefund, consts.BeanReasonAIGenItemsRefund)
	}

	_ = writer.Write(SSEProgressEvent{
		Done:      total,
		Total:     total,
		Processed: int(atomic.LoadInt64(&processed)),
		Failed:    int(atomic.LoadInt64(&failed)),
		Complete:  true,
	})
	writer.Close()
}

func processGenItems(meta models.ContentMeta, existingItems []models.ContentItem) bool {
	unitsInput := make([]map[string]string, len(existingItems))
	for i, item := range existingItems {
		unitsInput[i] = map[string]string{
			"content":     item.Content,
			"contentType": item.ContentType,
		}
	}

	unitsJSON, err := json.Marshal(unitsInput)
	if err != nil {
		return false
	}

	userMsg := "Source text: " + meta.SourceData + "\n\nUnits:\n" + string(unitsJSON)

	result, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
		Messages: []helpers.DeepSeekMessage{
			{Role: "system", Content: genItemsPrompt},
			{Role: "user", Content: userMsg},
		},
		Temperature: 0.1,
	})
	if err != nil {
		return false
	}

	aiUnits, err := helpers.ParseAIJSONArray(result)
	if err != nil || len(aiUnits) == 0 {
		return false
	}

	aiMap := make(map[string]json.RawMessage)
	for _, raw := range aiUnits {
		var unit struct {
			Content string          `json:"content"`
			Items   json.RawMessage `json:"items"`
		}
		if err := json.Unmarshal(raw, &unit); err != nil {
			continue
		}
		if unit.Content != "" {
			aiMap[unit.Content] = unit.Items
		}
	}

	for _, item := range existingItems {
		itemsJSON, ok := aiMap[item.Content]
		if !ok {
			continue
		}
		itemsStr := string(itemsJSON)
		if _, err := facades.Orm().Query().Model(&models.ContentItem{}).
			Where("id", item.ID).
			Update("items", itemsStr); err != nil {
			return false
		}
	}

	return true
}

var genItemsPrompt = `You are a language learning content processor. You will receive a list of learning units (each with content and contentType). Your job is to break each unit into individual word/punctuation items.

For each unit, produce an "items" array where each element represents one word or punctuation mark:
- position: 1-based index
- item: the word or punctuation character
- phonetic: {"uk": "IPA notation", "us": "IPA notation"} — set to null for punctuation marks
- pos: Chinese part-of-speech label like "名词", "动词", "形容词", "副词", "代词", "介词", "连词", "冠词", "感叹词", "助动词" etc — set to null for punctuation marks
- definition: Chinese definition of the word — set to empty string for punctuation marks
- answer: false for punctuation marks, proper names, place names, and abbreviations; true for all other words

Return a JSON array where each element has:
- content: the unit text (echo back exactly as given)
- items: the items array as described above

OUTPUT FORMAT:
Output ONLY a valid JSON array. No markdown code fences, no explanation, no extra text.

Example input units:
[{"content": "I", "contentType": "word"}, {"content": "I like", "contentType": "block"}]

Example output:
[
  {
    "content": "I",
    "items": [
      {"position": 1, "item": "I", "phonetic": {"uk": "/aɪ/", "us": "/aɪ/"}, "pos": "代词", "definition": "我", "answer": true}
    ]
  },
  {
    "content": "I like",
    "items": [
      {"position": 1, "item": "I", "phonetic": {"uk": "/aɪ/", "us": "/aɪ/"}, "pos": "代词", "definition": "我", "answer": true},
      {"position": 2, "item": "like", "phonetic": {"uk": "/laɪk/", "us": "/laɪk/"}, "pos": "动词", "definition": "喜欢", "answer": true}
    ]
  }
]`

// --- Helpers --- (verifyLevelOwnership unchanged)

func verifyLevelOwnership(userID, gameLevelID string) (*models.Game, *models.GameLevel, error) {
	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).First(&level); err != nil {
		return nil, nil, fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return nil, nil, ErrLevelNotFound
	}

	game, err := getCourseGameOwned(userID, level.GameID)
	if err != nil {
		return nil, nil, err
	}

	return game, &level, nil
}

func writeSSEError(writer *helpers.NDJSONWriter, err error) {
	msg := "服务异常"
	switch {
	case errors.Is(err, ErrVipRequired):
		msg = "升级会员解锁此功能"
	case errors.Is(err, ErrGamePublished):
		msg = "已发布的游戏不可编辑，请先撤回"
	case errors.Is(err, ErrInsufficientBeans):
		msg = "能量豆不足"
	case errors.Is(err, ErrEmptyContent):
		msg = "内容为空"
	case errors.Is(err, ErrGameNotFound):
		msg = "游戏不存在"
	case errors.Is(err, ErrLevelNotFound):
		msg = "关卡不存在"
	case errors.Is(err, ErrForbidden):
		msg = "无权操作"
	}
	_ = writer.WriteError(msg)
	writer.Close()
}
```

- [ ] **Step 2: Verify formatting**

```bash
cd dx-api && gofmt -l app/services/api/ai_custom_service.go
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/ai_custom_service.go
git commit -m "refactor(api): rewrite ai_custom_service to query content_metas/items directly"
```

### Task 4.3: Edit course_game_service.go DeleteGame, DeleteLevel, PublishGame

**Files:**
- Modify: `dx-api/app/services/api/course_game_service.go`

Three functions need rewrites: `DeleteGame`, `DeleteLevel`, `PublishGame`. Other functions (`ListUserGames`, `getCourseGameOwned`, `CreateGame`, `UpdateGame`, `WithdrawGame`, `CreateLevel`, `GetUserGameCounts`, `GetCourseGameDetail`, `isDuplicateKeyError`) are **unchanged**.

- [ ] **Step 1: Open the file and locate `DeleteGame` (around line 218). Replace it with:**

```go
// DeleteGame deletes a course game and cascades to levels and content. Rejects published games.
func DeleteGame(userID, gameID string) error {
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

	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Exec(
			`UPDATE content_items SET deleted_at = NOW() WHERE game_id = ? AND deleted_at IS NULL`, gameID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_items: %w", err)
		}
		if _, err := tx.Exec(
			`UPDATE content_metas SET deleted_at = NOW() WHERE game_id = ? AND deleted_at IS NULL`, gameID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_metas: %w", err)
		}
		if _, err := tx.Exec(
			`UPDATE game_vocabs SET deleted_at = NOW() WHERE game_id = ? AND deleted_at IS NULL`, gameID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete game_vocabs: %w", err)
		}
		if _, err := tx.Where("game_id", gameID).Delete(&models.GameLevel{}); err != nil {
			return fmt.Errorf("failed to delete levels: %w", err)
		}
		if _, err := tx.Where("id", gameID).Delete(&models.Game{}); err != nil {
			return fmt.Errorf("failed to delete game: %w", err)
		}
		return nil
	})
}
```

- [ ] **Step 2: Locate `DeleteLevel` (around line 459). Replace it with:**

```go
// DeleteLevel removes a level and its content from a course game.
func DeleteLevel(userID, gameID, levelID string) error {
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

	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", levelID).Where("game_id", gameID).First(&level); err != nil {
		return fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return ErrLevelNotFound
	}

	return facades.Orm().Transaction(func(tx orm.Query) error {
		if _, err := tx.Exec(
			`UPDATE content_items SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL`, levelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_items: %w", err)
		}
		if _, err := tx.Exec(
			`UPDATE content_metas SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL`, levelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete content_metas: %w", err)
		}
		if _, err := tx.Exec(
			`UPDATE game_vocabs SET deleted_at = NOW() WHERE game_level_id = ? AND deleted_at IS NULL`, levelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete game_vocabs: %w", err)
		}
		if _, err := tx.Where("id", levelID).Delete(&models.GameLevel{}); err != nil {
			return fmt.Errorf("failed to delete level: %w", err)
		}
		return nil
	})
}
```

- [ ] **Step 3: Locate `PublishGame` (around line 322). Replace the inner per-level loop body with mode-aware counting.** Replace the function from `// Check each level has content items` through the loop end with:

```go
	// Check each level has content
	var levels []models.GameLevel
	if err := facades.Orm().Query().Where("game_id", gameID).Where("is_active", true).Get(&levels); err != nil {
		return fmt.Errorf("failed to load levels: %w", err)
	}

	for _, l := range levels {
		var itemCount int64
		var ungeneratedCount int64
		var err error

		if consts.IsVocabMode(game.Mode) {
			itemCount, err = facades.Orm().Query().Model(&models.GameVocab{}).
				Where("game_level_id", l.ID).
				Count()
			if err != nil {
				return fmt.Errorf("failed to count vocab placements: %w", err)
			}
			// Vocab modes: enforce batch-size on game_vocabs count
			batchSize := consts.VocabBatchSize(game.Mode)
			if batchSize > 0 && itemCount%int64(batchSize) != 0 {
				return fmt.Errorf("关卡「%s」词汇数量必须是 %d 的倍数（当前 %d 条）", l.Name, batchSize, itemCount)
			}
		} else {
			itemCount, err = facades.Orm().Query().Model(&models.ContentItem{}).
				Where("game_level_id", l.ID).
				Count()
			if err != nil {
				return fmt.Errorf("failed to count items: %w", err)
			}
			ungeneratedCount, err = facades.Orm().Query().Model(&models.ContentItem{}).
				Where("game_level_id", l.ID).
				Where("items IS NULL").
				Count()
			if err != nil {
				return fmt.Errorf("failed to count ungenerated items: %w", err)
			}
			if ungeneratedCount > 0 {
				return fmt.Errorf("关卡「%s」有未生成的练习单元", l.Name)
			}
		}

		if itemCount == 0 {
			return fmt.Errorf("关卡「%s」没有练习内容", l.Name)
		}
	}
```

- [ ] **Step 4: Verify formatting**

```bash
cd dx-api && gofmt -l app/services/api/course_game_service.go
```

Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add dx-api/app/services/api/course_game_service.go
git commit -m "refactor(api): course_game_service — direct queries; mode-aware publish check"
```

### Phase 4 validation gate

- [ ] **Verify partial build progresses**

```bash
cd dx-api && go build ./app/services/api/course_content_service.go ./app/services/api/ai_custom_service.go ./app/services/api/course_game_service.go 2>&1 | head -20
```

Expected: compiler errors will reference `ai_custom_vocab_service.go` (still uses GameMeta/GameItem) — that's deleted in Phase 5. Build of just these three files won't work standalone (Go builds at package level), but `go vet ./app/services/api/...` should narrow which symbols still need updating.

```bash
cd dx-api && go vet ./app/services/api/... 2>&1 | head -30
```

Expected: errors only in files we'll handle in Phases 5-7 (`ai_custom_vocab_service.go`, `content_service.go`, `feedback_service.go`, `user_master_service.go`, `user_unknown_service.go`, `user_review_service.go`, `game_play_*_service.go`).

---




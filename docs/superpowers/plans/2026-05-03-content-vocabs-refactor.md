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



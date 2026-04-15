# Backfill Content Metas & Game Metas Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 1:1 backfill `content_metas` and `game_metas` for the 1,220,803 imported content items so that every item has an upstream meta marked with `source_from='import'`, without touching any existing workflow code.

**Architecture:** One new Goravel artisan command `app:backfill-metas`. Chunked transactional writes (5000 rows/transaction). Three SQL statements per chunk: bulk INSERT `content_metas`, bulk INSERT `game_metas`, bulk UPDATE `content_items.content_meta_id` via `UPDATE … FROM VALUES`. Idempotent via `WHERE content_meta_id IS NULL` filter. One small constants edit (`SourceFromImport = "import"`).

**Tech Stack:** Go 1.25+, Goravel framework (ORM + console command contracts), PostgreSQL, UUIDv7 via `github.com/google/uuid`.

**Spec:** `docs/superpowers/specs/2026-04-15-backfill-content-metas-design.md`

---

## File Structure

**New files:**
- `dx-api/app/console/commands/backfill_metas.go` — the `BackfillMetas` command (Signature/Description/Extend/Handle + helper functions `countBackfillCandidates`, `deriveSourceType`, `backfillChunk`, `bulkLinkItems`, and `backfillRow` struct)
- `dx-api/app/console/commands/backfill_metas_test.go` — unit test for the pure helper `deriveSourceType`

**Modified files:**
- `dx-api/app/consts/source_from.go` — add `SourceFromImport` constant and its `SourceFromLabels` entry
- `dx-api/bootstrap/app.go` — register the new `BackfillMetas` command in the `WithCommands` slice

No migrations, no model changes, no service changes, no route changes.

---

## Task 1: Add `SourceFromImport` constant

**Files:**
- Modify: `dx-api/app/consts/source_from.go`

- [ ] **Step 1: Edit the constants file**

Replace the whole file contents with:

```go
package consts

// Source origin values.
const (
	SourceFromManual = "manual"
	SourceFromAI     = "ai"
	SourceFromImport = "import"
)

// SourceFromLabels maps each source origin to its Chinese label.
var SourceFromLabels = map[string]string{
	SourceFromManual: "手动添加",
	SourceFromAI:     "AI 生成",
	SourceFromImport: "导入",
}
```

- [ ] **Step 2: Verify the package still builds**

Run: `cd dx-api && go build ./app/consts/...`
Expected: no output, exit code 0.

- [ ] **Step 3: Commit**

```bash
cd dx-api
git add app/consts/source_from.go
git commit -m "feat(api): add SourceFromImport source_from value"
```

---

## Task 2: TDD `deriveSourceType` helper

**Files:**
- Create: `dx-api/app/console/commands/backfill_metas_test.go`
- Create: `dx-api/app/console/commands/backfill_metas.go`

- [ ] **Step 1: Write the failing test**

Create `dx-api/app/console/commands/backfill_metas_test.go` with:

```go
package commands

import (
	"testing"

	"dx-api/app/consts"
)

func TestDeriveSourceType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        string
	}{
		{"word is vocab", consts.ContentTypeWord, consts.SourceTypeVocab},
		{"phrase is vocab", consts.ContentTypePhrase, consts.SourceTypeVocab},
		{"block is vocab", consts.ContentTypeBlock, consts.SourceTypeVocab},
		{"sentence is sentence", consts.ContentTypeSentence, consts.SourceTypeSentence},
		{"unknown defaults to vocab", "unknown", consts.SourceTypeVocab},
		{"empty defaults to vocab", "", consts.SourceTypeVocab},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveSourceType(tt.contentType)
			if got != tt.want {
				t.Errorf("deriveSourceType(%q) = %q, want %q", tt.contentType, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd dx-api && go test ./app/console/commands/ -run TestDeriveSourceType -v`
Expected: compile error — `undefined: deriveSourceType` (because `backfill_metas.go` does not exist yet).

- [ ] **Step 3: Create the minimal `backfill_metas.go` with just `deriveSourceType`**

Create `dx-api/app/console/commands/backfill_metas.go` with:

```go
package commands

import (
	"dx-api/app/consts"
)

// deriveSourceType maps a content_items.content_type to the corresponding
// content_metas.source_type per the backfill rule:
//   sentence → sentence (complete sentence)
//   word, phrase, block → vocab (all non-complete units)
func deriveSourceType(contentType string) string {
	if contentType == consts.ContentTypeSentence {
		return consts.SourceTypeSentence
	}
	return consts.SourceTypeVocab
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd dx-api && go test ./app/console/commands/ -run TestDeriveSourceType -v`
Expected: `PASS` with 6 subtests (`word is vocab`, `phrase is vocab`, `block is vocab`, `sentence is sentence`, `unknown defaults to vocab`, `empty defaults to vocab`).

- [ ] **Step 5: Commit**

```bash
cd dx-api
git add app/console/commands/backfill_metas.go app/console/commands/backfill_metas_test.go
git commit -m "feat(api): add deriveSourceType helper for backfill command"
```

---

## Task 3: Command skeleton (Signature / Description / Extend / empty Handle)

**Files:**
- Modify: `dx-api/app/console/commands/backfill_metas.go`

- [ ] **Step 1: Expand `backfill_metas.go` with the command struct and empty Handle**

Replace the file contents with:

```go
package commands

import (
	"dx-api/app/consts"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type BackfillMetas struct{}

func (c *BackfillMetas) Signature() string {
	return "app:backfill-metas"
}

func (c *BackfillMetas) Description() string {
	return "1:1 backfill content_metas and game_metas for imported content_items (source_from=import)"
}

func (c *BackfillMetas) Extend() command.Extend {
	return command.Extend{
		Flags: []command.Flag{
			&command.IntFlag{
				Name:  "batch-size",
				Value: 5000,
				Usage: "Rows per transaction",
			},
			&command.IntFlag{
				Name:  "limit",
				Value: 0,
				Usage: "Process at most N rows total (0 = no limit)",
			},
			&command.BoolFlag{
				Name:  "dry-run",
				Usage: "Count affected rows without writing",
			},
		},
	}
}

func (c *BackfillMetas) Handle(ctx console.Context) error {
	// Placeholder — filled in by Task 5.
	ctx.Info("backfill-metas: not implemented yet")
	return nil
}

// deriveSourceType maps a content_items.content_type to the corresponding
// content_metas.source_type per the backfill rule:
//   sentence → sentence (complete sentence)
//   word, phrase, block → vocab (all non-complete units)
func deriveSourceType(contentType string) string {
	if contentType == consts.ContentTypeSentence {
		return consts.SourceTypeSentence
	}
	return consts.SourceTypeVocab
}
```

- [ ] **Step 2: Run the unit test to make sure `deriveSourceType` still passes**

Run: `cd dx-api && go test ./app/console/commands/ -run TestDeriveSourceType -v`
Expected: `PASS` with all 6 subtests.

- [ ] **Step 3: Build the package**

Run: `cd dx-api && go build ./app/console/commands/...`
Expected: no output, exit code 0.

- [ ] **Step 4: Commit**

```bash
cd dx-api
git add app/console/commands/backfill_metas.go
git commit -m "feat(api): scaffold app:backfill-metas command"
```

---

## Task 4: Register the command in `bootstrap/app.go`

**Files:**
- Modify: `dx-api/bootstrap/app.go`

- [ ] **Step 1: Inspect the current `WithCommands` block**

Run: `cd dx-api && grep -n "WithCommands\|commands\." bootstrap/app.go`
Expected: shows the existing `WithCommands(func() []console.Command { ... })` block containing at least `UpdatePlayStreaks`, `ResetEnergyBeans`, `ImportCourses`, `ExpireStaleOrders`.

- [ ] **Step 2: Add `BackfillMetas` to the commands slice**

Inside `bootstrap/app.go`, in the `WithCommands` callback, add `&commands.BackfillMetas{},` as the last entry of the returned slice, right after `&commands.ExpireStaleOrders{},`.

The block should look like this after the edit:

```go
WithCommands(func() []console.Command {
    return []console.Command{
        &commands.UpdatePlayStreaks{},
        &commands.ResetEnergyBeans{},
        &commands.ImportCourses{},
        &commands.ExpireStaleOrders{},
        &commands.BackfillMetas{},
    }
}).
```

Do **not** add anything to the `WithSchedule` block — this command is one-shot, not scheduled.

- [ ] **Step 3: Build the whole module**

Run: `cd dx-api && go build ./...`
Expected: no output, exit code 0.

- [ ] **Step 4: Verify the command is discoverable**

Run: `cd dx-api && go run . artisan list 2>&1 | grep backfill`
Expected: one line containing `app:backfill-metas` followed by the description `1:1 backfill content_metas and game_metas ...`.

- [ ] **Step 5: Verify flags are wired up**

Run: `cd dx-api && go run . artisan app:backfill-metas --help 2>&1`
Expected: the help text lists `--batch-size`, `--limit`, `--dry-run` flags.

- [ ] **Step 6: Run the empty command to verify it executes end-to-end**

Run: `cd dx-api && go run . artisan app:backfill-metas`
Expected: prints `backfill-metas: not implemented yet` and exits 0.

- [ ] **Step 7: Commit**

```bash
cd dx-api
git add bootstrap/app.go
git commit -m "feat(api): register app:backfill-metas command"
```

---

## Task 5: Dry-run path — count candidates and short-circuit when requested

**Files:**
- Modify: `dx-api/app/console/commands/backfill_metas.go`

- [ ] **Step 1: Add `countBackfillCandidates` helper and wire up the dry-run branch in `Handle`**

Replace the current `backfill_metas.go` with:

```go
package commands

import (
	"dx-api/app/consts"
	"dx-api/app/models"
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
)

type BackfillMetas struct{}

func (c *BackfillMetas) Signature() string {
	return "app:backfill-metas"
}

func (c *BackfillMetas) Description() string {
	return "1:1 backfill content_metas and game_metas for imported content_items (source_from=import)"
}

func (c *BackfillMetas) Extend() command.Extend {
	return command.Extend{
		Flags: []command.Flag{
			&command.IntFlag{
				Name:  "batch-size",
				Value: 5000,
				Usage: "Rows per transaction",
			},
			&command.IntFlag{
				Name:  "limit",
				Value: 0,
				Usage: "Process at most N rows total (0 = no limit)",
			},
			&command.BoolFlag{
				Name:  "dry-run",
				Usage: "Count affected rows without writing",
			},
		},
	}
}

func (c *BackfillMetas) Handle(ctx console.Context) error {
	start := time.Now()
	batchSize := ctx.OptionInt("batch-size")
	limit := ctx.OptionInt("limit")
	dryRun := ctx.OptionBool("dry-run")

	if batchSize <= 0 {
		batchSize = 5000
	}

	total, err := countBackfillCandidates()
	if err != nil {
		return fmt.Errorf("failed to count candidates: %w", err)
	}
	if limit > 0 && int64(limit) < total {
		total = int64(limit)
	}
	ctx.Info(fmt.Sprintf("backfill candidates: %d", total))
	if total == 0 {
		ctx.Info("nothing to backfill")
		return nil
	}
	if dryRun {
		ctx.Info("dry-run — no writes")
		return nil
	}

	// Placeholder — filled in by Task 8.
	ctx.Info(fmt.Sprintf("batch-size=%d (not yet implemented; elapsed %s)", batchSize, time.Since(start)))
	return nil
}

// countBackfillCandidates returns the number of content_items still needing a meta.
func countBackfillCandidates() (int64, error) {
	return facades.Orm().Query().Model(&models.ContentItem{}).
		Where("content_meta_id IS NULL").
		Count()
}

// deriveSourceType maps a content_items.content_type to the corresponding
// content_metas.source_type per the backfill rule:
//   sentence → sentence (complete sentence)
//   word, phrase, block → vocab (all non-complete units)
func deriveSourceType(contentType string) string {
	if contentType == consts.ContentTypeSentence {
		return consts.SourceTypeSentence
	}
	return consts.SourceTypeVocab
}
```

- [ ] **Step 2: Run the unit test to confirm `deriveSourceType` still passes**

Run: `cd dx-api && go test ./app/console/commands/ -run TestDeriveSourceType -v`
Expected: `PASS` on all 6 subtests.

- [ ] **Step 3: Build**

Run: `cd dx-api && go build ./...`
Expected: no output, exit code 0.

- [ ] **Step 4: Run `--dry-run` against the dev DB to verify the count matches reality**

Run: `cd dx-api && go run . artisan app:backfill-metas --dry-run`
Expected output order:
```
backfill candidates: 1220803
dry-run — no writes
```

If the count differs (e.g., 0 or some other number), stop and investigate — the dev DB state may not match what the spec assumes.

- [ ] **Step 5: Run with `--dry-run --limit 100` to verify the limit clamp works**

Run: `cd dx-api && go run . artisan app:backfill-metas --dry-run --limit 100`
Expected output:
```
backfill candidates: 100
dry-run — no writes
```

- [ ] **Step 6: Commit**

```bash
cd dx-api
git add app/console/commands/backfill_metas.go
git commit -m "feat(api): implement dry-run and candidate counter for backfill-metas"
```

---

## Task 6: Row loader and `backfillRow` struct

**Files:**
- Modify: `dx-api/app/console/commands/backfill_metas.go`

- [ ] **Step 1: Add the `backfillRow` struct and `loadBackfillChunk` helper**

Append the following to `backfill_metas.go`, just after `deriveSourceType`:

```go
// backfillRow is a single (content_item, game_item) pair we need to process.
type backfillRow struct {
	CIID        string  `gorm:"column:ci_id"`
	Content     string  `gorm:"column:content"`
	ContentType string  `gorm:"column:content_type"`
	Translation *string `gorm:"column:translation"`
	GameID      string  `gorm:"column:game_id"`
	GameLevelID string  `gorm:"column:game_level_id"`
	GIOrder     float64 `gorm:"column:gi_order"`
}

// loadBackfillChunk selects up to `size` content_items that still need a meta,
// joined with their game_item so we know the target game/level/order.
// Rows are ordered by content_items.id (UUIDv7, time-sortable) so every run
// processes the oldest unlinked rows first.
func loadBackfillChunk(size int) ([]backfillRow, error) {
	var rows []backfillRow
	if err := facades.Orm().Query().Raw(`
		SELECT ci.id AS ci_id,
		       ci.content,
		       ci.content_type,
		       ci.translation,
		       gi.game_id,
		       gi.game_level_id,
		       gi."order" AS gi_order
		FROM content_items ci
		JOIN game_items gi
		  ON gi.content_item_id = ci.id
		 AND gi.deleted_at IS NULL
		WHERE ci.content_meta_id IS NULL
		  AND ci.deleted_at IS NULL
		ORDER BY ci.id
		LIMIT ?
	`, size).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to load chunk: %w", err)
	}
	return rows, nil
}
```

- [ ] **Step 2: Build**

Run: `cd dx-api && go build ./...`
Expected: no output, exit code 0.

- [ ] **Step 3: Commit**

```bash
cd dx-api
git add app/console/commands/backfill_metas.go
git commit -m "feat(api): add backfillRow loader for backfill-metas command"
```

---

## Task 7: Transactional chunk writer — `bulkLinkItems` and `backfillChunk` together

We add both helpers in a single task because `bulkLinkItems` takes `orm.Query` as its first parameter, which is only used once `backfillChunk` calls it — splitting them would leave the `orm` import unused in the intermediate state and break the build.

**Files:**
- Modify: `dx-api/app/console/commands/backfill_metas.go`

- [ ] **Step 1: Update the import block**

Replace the existing import block at the top of `backfill_metas.go` with:

```go
import (
	"dx-api/app/consts"
	"dx-api/app/models"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)
```

- [ ] **Step 2: Append `bulkLinkItems` at the bottom of the file (below `loadBackfillChunk`)**

```go
// bulkLinkItems issues a single UPDATE that sets content_meta_id for an entire
// chunk of rows via a VALUES list. Far faster than per-row UPDATEs.
//
// Goravel uses `?` placeholders (GORM-style) rather than PostgreSQL-native `$N`.
// String values bind as SQL text, so we cast both sides of the join back to
// uuid inside the query.
func bulkLinkItems(tx orm.Query, rows []backfillRow, metaIDs []string) error {
	if len(rows) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(`UPDATE content_items AS ci
	SET content_meta_id = v.meta_id::uuid
	FROM (VALUES `)

	args := make([]any, 0, len(rows)*2)
	for i, r := range rows {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(?, ?)")
		args = append(args, r.CIID, metaIDs[i])
	}
	sb.WriteString(`) AS v(ci_id, meta_id) WHERE ci.id = v.ci_id::uuid`)

	if _, err := tx.Exec(sb.String(), args...); err != nil {
		return fmt.Errorf("bulk link update: %w", err)
	}
	return nil
}
```

- [ ] **Step 3: Append `backfillChunk` at the very bottom of the file (below `bulkLinkItems`)**

```go
// backfillChunk loads up to `size` unprocessed rows, writes metas/game_metas,
// and links them back on content_items. All three writes run inside a single
// transaction so the state is consistent per chunk. Returns the number of
// rows processed (0 when nothing left to do).
func backfillChunk(size int) (int, error) {
	rows, err := loadBackfillChunk(size)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}

	// Pre-generate UUIDs outside the transaction so the work on the critical
	// path is purely I/O.
	metaIDs := make([]string, len(rows))
	gameMetaIDs := make([]string, len(rows))
	for i := range rows {
		metaIDs[i] = uuid.Must(uuid.NewV7()).String()
		gameMetaIDs[i] = uuid.Must(uuid.NewV7()).String()
	}

	err = facades.Orm().Transaction(func(tx orm.Query) error {
		// 1. Bulk insert content_metas.
		metas := make([]models.ContentMeta, len(rows))
		for i, r := range rows {
			metas[i] = models.ContentMeta{
				ID:          metaIDs[i],
				SourceFrom:  consts.SourceFromImport,
				SourceType:  deriveSourceType(r.ContentType),
				SourceData:  r.Content,
				Translation: r.Translation,
				IsBreakDone: true,
			}
		}
		if err := tx.Create(&metas); err != nil {
			return fmt.Errorf("insert content_metas: %w", err)
		}

		// 2. Bulk insert game_metas, each pointing at the matching content_meta
		//    with the same (game_id, game_level_id, order) as its game_item.
		gameMetas := make([]models.GameMeta, len(rows))
		for i, r := range rows {
			gameMetas[i] = models.GameMeta{
				ID:            gameMetaIDs[i],
				GameID:        r.GameID,
				GameLevelID:   r.GameLevelID,
				ContentMetaID: metaIDs[i],
				Order:         r.GIOrder,
			}
		}
		if err := tx.Create(&gameMetas); err != nil {
			return fmt.Errorf("insert game_metas: %w", err)
		}

		// 3. Bulk update content_items.content_meta_id in a single UPDATE.
		if err := bulkLinkItems(tx, rows, metaIDs); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}
```

- [ ] **Step 4: Build**

Run: `cd dx-api && go build ./...`
Expected: no output, exit code 0.

- [ ] **Step 5: Run the unit test**

Run: `cd dx-api && go test ./app/console/commands/ -run TestDeriveSourceType -v`
Expected: `PASS`.

- [ ] **Step 6: Commit**

```bash
cd dx-api
git add app/console/commands/backfill_metas.go
git commit -m "feat(api): implement transactional chunk writer for backfill-metas"
```

---

## Task 8: Wire the batching loop into `Handle`

**Files:**
- Modify: `dx-api/app/console/commands/backfill_metas.go`

- [ ] **Step 1: Replace the `Handle` body with the real loop**

Find this existing placeholder block:

```go
	// Placeholder — filled in by Task 8.
	ctx.Info(fmt.Sprintf("batch-size=%d (not yet implemented; elapsed %s)", batchSize, time.Since(start)))
	return nil
```

Replace it with:

```go
	var processed int64
	for processed < total {
		chunk := batchSize
		if remaining := int(total - processed); remaining < chunk {
			chunk = remaining
		}

		n, err := backfillChunk(chunk)
		if err != nil {
			return fmt.Errorf("chunk at offset %d: %w", processed, err)
		}
		if n == 0 {
			// No more rows match the filter — done even if the initial count
			// suggested otherwise (another process could have consumed some).
			break
		}
		processed += int64(n)
		ctx.Info(fmt.Sprintf("[%d/%d] done in %s", processed, total, time.Since(start)))
	}

	ctx.NewLine()
	ctx.Info(fmt.Sprintf("backfill complete: %d rows in %s", processed, time.Since(start)))
	return nil
```

- [ ] **Step 2: Build**

Run: `cd dx-api && go build ./...`
Expected: no output, exit code 0.

- [ ] **Step 3: Run the unit test**

Run: `cd dx-api && go test ./app/console/commands/ -run TestDeriveSourceType -v`
Expected: `PASS`.

- [ ] **Step 4: Run `--dry-run` to confirm Handle still short-circuits correctly**

Run: `cd dx-api && go run . artisan app:backfill-metas --dry-run`
Expected output:
```
backfill candidates: 1220803
dry-run — no writes
```

The write path must NOT be invoked on a dry-run — if you see "[N/1220803]" lines, stop and fix `Handle` to return before reaching the loop.

- [ ] **Step 5: Commit**

```bash
cd dx-api
git add app/console/commands/backfill_metas.go
git commit -m "feat(api): wire up batching loop in backfill-metas Handle"
```

---

## Task 9: Lint and vet

**Files:** (no edits — verification only)

- [ ] **Step 1: Run `go vet` over the whole module**

Run: `cd dx-api && go vet ./...`
Expected: no output, exit code 0.

- [ ] **Step 2: Run `go build` over the whole module**

Run: `cd dx-api && go build ./...`
Expected: no output, exit code 0.

- [ ] **Step 3: Run the full unit-test suite with `-race`**

Run: `cd dx-api && go test -race ./app/...`
Expected: all packages pass. If any unrelated test was failing before this branch, note it but do NOT attempt to fix it in this plan.

---

## Task 10: Small-batch smoke test against the dev DB

**Goal:** prove the write path works end-to-end on a small subset before touching all 1.22M rows.

**Files:** (no edits — execution only)

- [ ] **Step 1: Snapshot the pre-run counts**

Run:
```bash
docker exec deploy-postgres-1 psql -U postgres -d dxdb -c "
SELECT
  (SELECT COUNT(*) FROM content_items WHERE content_meta_id IS NULL AND deleted_at IS NULL) AS items_unlinked,
  (SELECT COUNT(*) FROM content_metas WHERE source_from = 'import' AND deleted_at IS NULL) AS import_metas,
  (SELECT COUNT(*) FROM game_metas WHERE deleted_at IS NULL) AS game_metas
;"
```

Expected (fresh DB, no prior backfill):
```
 items_unlinked | import_metas | game_metas
----------------+--------------+-----------
        1220803 |            0 |          0
```

- [ ] **Step 2: Run the small-batch smoke test**

Run: `cd dx-api && go run . artisan app:backfill-metas --limit 100 --batch-size 50`

Expected output ordering:
```
backfill candidates: 100
[50/100] done in <time>
[100/100] done in <time>

backfill complete: 100 rows in <time>
```

- [ ] **Step 3: Post-run counts — expect 100 rows processed**

Run:
```bash
docker exec deploy-postgres-1 psql -U postgres -d dxdb -c "
SELECT
  (SELECT COUNT(*) FROM content_items WHERE content_meta_id IS NULL AND deleted_at IS NULL) AS items_unlinked,
  (SELECT COUNT(*) FROM content_metas WHERE source_from = 'import' AND deleted_at IS NULL) AS import_metas,
  (SELECT COUNT(*) FROM game_metas WHERE deleted_at IS NULL) AS game_metas
;"
```

Expected:
```
 items_unlinked | import_metas | game_metas
----------------+--------------+-----------
        1220703 |          100 |        100
```

(`items_unlinked` drops by 100; `import_metas` and `game_metas` each rise by 100.)

- [ ] **Step 4: Verify each new meta is linked from exactly one content_item**

Run:
```bash
docker exec deploy-postgres-1 psql -U postgres -d dxdb -c "
SELECT COUNT(*) AS orphan_metas
FROM content_metas cm
WHERE cm.source_from = 'import'
  AND NOT EXISTS (
    SELECT 1 FROM content_items ci WHERE ci.content_meta_id = cm.id
  );"
```
Expected: `orphan_metas = 0`.

- [ ] **Step 5: Verify source_type assignment is correct for the 100 rows**

Run:
```bash
docker exec deploy-postgres-1 psql -U postgres -d dxdb -c "
SELECT cm.source_type, ci.content_type, COUNT(*)
FROM content_metas cm
JOIN content_items ci ON ci.content_meta_id = cm.id
WHERE cm.source_from = 'import'
GROUP BY cm.source_type, ci.content_type
ORDER BY cm.source_type, ci.content_type;"
```

Expected: every row where `ci.content_type = 'sentence'` has `cm.source_type = 'sentence'`; every other `content_type` (`word`/`phrase`/`block`) has `cm.source_type = 'vocab'`. No other combinations may appear.

- [ ] **Step 6: Verify `game_metas.order` lines up with its `game_items.order` at the same level**

Run:
```bash
docker exec deploy-postgres-1 psql -U postgres -d dxdb -c "
SELECT COUNT(*) AS misaligned
FROM game_metas gm
JOIN content_metas cm ON cm.id = gm.content_meta_id
WHERE cm.source_from = 'import'
  AND NOT EXISTS (
    SELECT 1 FROM game_items gi
    JOIN content_items ci ON ci.id = gi.content_item_id AND ci.content_meta_id = gm.content_meta_id
    WHERE gi.game_level_id = gm.game_level_id
      AND gi.\"order\" = gm.\"order\"
      AND gi.deleted_at IS NULL
  );"
```
Expected: `misaligned = 0`.

- [ ] **Step 7: Verify idempotency — a second run with the same args is a no-op on those 100 rows**

Run: `cd dx-api && go run . artisan app:backfill-metas --dry-run`
Expected:
```
backfill candidates: 1220703
dry-run — no writes
```

The `items_unlinked` should match the number from Step 3, confirming that the 100 already-processed rows are not re-counted.

---

## Task 11: Full backfill execution

**Files:** (no edits — execution only)

- [ ] **Step 1: Run the full backfill**

Run: `cd dx-api && time go run . artisan app:backfill-metas`

Expected:
- Stream of `[N/1220703] done in …` progress lines (245-ish chunks)
- Final `backfill complete: 1220703 rows in …` line (plus the 100 from Task 10 = total 1,220,803 backfilled across runs)
- Wall time: 2–6 minutes

If any `chunk at offset N: …` error is printed, the process exits non-zero. Capture the error, do NOT rerun blindly — investigate first. Idempotency means resuming is safe, but the underlying error needs diagnosis.

- [ ] **Step 2: Confirm the progress logging is monotonic**

Eyeball the output — the left-hand number in `[N/total]` must only ever increase. If you see it reset or go backwards, something is very wrong.

---

## Task 12: Post-backfill verification SQL

**Files:** (no edits — verification only)

- [ ] **Step 1: All content items are linked**

Run:
```bash
docker exec deploy-postgres-1 psql -U postgres -d dxdb -c "
SELECT COUNT(*) FROM content_items
WHERE content_meta_id IS NULL AND deleted_at IS NULL;"
```
Expected: `0`.

- [ ] **Step 2: Total imported metas matches total content items**

Run:
```bash
docker exec deploy-postgres-1 psql -U postgres -d dxdb -c "
SELECT
  (SELECT COUNT(*) FROM content_items WHERE deleted_at IS NULL) AS items,
  (SELECT COUNT(*) FROM content_metas WHERE source_from = 'import' AND deleted_at IS NULL AND is_break_done = true) AS import_metas,
  (SELECT COUNT(*) FROM game_metas WHERE deleted_at IS NULL) AS game_metas
;"
```
Expected: all three columns equal `1220803`.

- [ ] **Step 3: `source_type` distribution matches the spec's predictions**

Run:
```bash
docker exec deploy-postgres-1 psql -U postgres -d dxdb -c "
SELECT source_type, COUNT(*)
FROM content_metas
WHERE source_from = 'import' AND deleted_at IS NULL
GROUP BY source_type
ORDER BY source_type;"
```
Expected:
```
 source_type | count
-------------+--------
 sentence    | 569540
 vocab       | 651263
```

- [ ] **Step 4: No orphan imported metas**

Run:
```bash
docker exec deploy-postgres-1 psql -U postgres -d dxdb -c "
SELECT COUNT(*) AS orphan_metas
FROM content_metas cm
WHERE cm.source_from = 'import'
  AND cm.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM content_items ci WHERE ci.content_meta_id = cm.id AND ci.deleted_at IS NULL
  );"
```
Expected: `orphan_metas = 0`.

- [ ] **Step 5: All imported game_metas line up with a game_item at same (level, order)**

Run:
```bash
docker exec deploy-postgres-1 psql -U postgres -d dxdb -c "
SELECT COUNT(*) AS misaligned
FROM game_metas gm
WHERE gm.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_items gi
    WHERE gi.game_level_id = gm.game_level_id
      AND gi.\"order\" = gm.\"order\"
      AND gi.deleted_at IS NULL
  );"
```
Expected: `misaligned = 0`.

- [ ] **Step 6: Invariant — every imported meta has exactly 1 content_item linked**

Run:
```bash
docker exec deploy-postgres-1 psql -U postgres -d dxdb -c "
SELECT cnt, COUNT(*) AS metas
FROM (
  SELECT cm.id, COUNT(ci.id) AS cnt
  FROM content_metas cm
  LEFT JOIN content_items ci ON ci.content_meta_id = cm.id AND ci.deleted_at IS NULL
  WHERE cm.source_from = 'import' AND cm.deleted_at IS NULL
  GROUP BY cm.id
) sub
GROUP BY cnt
ORDER BY cnt;"
```
Expected: exactly one row — `cnt = 1, metas = 1220803`.

---

## Task 13: Workflow smoke tests

**Files:** (no edits — manual verification only)

- [ ] **Step 1: Verify an imported game still plays**

1. Start the dev stack if it's not running: `cd dx-source/deploy && docker compose -f docker-compose.dev.yml up -d`.
2. Open the web client in a browser at `http://localhost`.
3. Sign in with an existing dev user (or create one).
4. Navigate to the games hall and open any published game.
5. Start a level and play at least 10 items.

Expected:
- All items load in the same order as before the backfill.
- Audio, translations, and item JSON render normally (they come from `content_items`, which we updated only on the `content_meta_id` column — no content field was touched).
- No errors in the browser console.
- No 500 errors in the API logs (`docker logs deploy-dx-api-1 --tail 50`).

If any regression is observed, stop and diagnose before proceeding. Rollback is available via the SQL in the spec's "Blast radius of failure" section.

- [ ] **Step 2: Verify a fresh AI-custom flow still works on a brand new game**

1. Sign in as a VIP user.
2. Create a new game (mode: word-sentence).
3. On AI-custom, generate or format a short sentence metadata, save it to a level.
4. Run `Break` — should generate items.
5. Run `Generate items` — should fill the items JSONB.

Expected: the new metas have `source_from` = `manual` or `ai` (NOT `import`), and the flow completes without hitting any capacity, break, or generation errors. Imported metas must not interfere with the new flow.

Run this SQL to confirm the new metas are labeled correctly:
```bash
docker exec deploy-postgres-1 psql -U postgres -d dxdb -c "
SELECT source_from, COUNT(*)
FROM content_metas
WHERE created_at > NOW() - INTERVAL '10 minutes'
GROUP BY source_from
ORDER BY source_from;"
```
Expected: source_from for the newly created metas is `manual` or `ai`, never `import`.

---

## Task 14: Final commit cleanup and summary

**Files:** (no edits — verification only)

- [ ] **Step 1: Verify the git history for this feature is clean**

Run: `cd dx-source && git log --oneline main..HEAD`
Expected: 6 new commits in roughly this order:
```
... feat(api): wire up batching loop in backfill-metas Handle
... feat(api): implement transactional chunk writer for backfill-metas
... feat(api): add backfillRow loader for backfill-metas command
... feat(api): implement dry-run and candidate counter for backfill-metas
... feat(api): register app:backfill-metas command
... feat(api): scaffold app:backfill-metas command
... feat(api): add deriveSourceType helper for backfill command
... feat(api): add SourceFromImport source_from value
```

(Your exact commit list may differ slightly — what matters is that each commit is small, atomic, and the tree builds cleanly at every commit.)

- [ ] **Step 2: Confirm no stray edits**

Run: `cd dx-source && git status`
Expected: `nothing to commit, working tree clean`.

- [ ] **Step 3: Confirm the build and tests pass at HEAD**

Run: `cd dx-api && go build ./... && go vet ./... && go test -race ./app/console/commands/...`
Expected: all three succeed, exit code 0.

---

## Out of scope

The following are explicitly not part of this plan:

- Touching any service, controller, middleware, route, or frontend file
- Changing `MaxSentences`, `MaxVocab`, `MaxMetasPerLevel`, `MaxItemsPerMeta`, or any limit constant
- Changing the `SaveMetadataBatch` capacity check
- Grouping content items into shared metas (rejected during brainstorming in favor of 1:1)
- Any AI/DeepSeek runtime call (the backfill is fully deterministic)
- Fixing the pre-existing bug in `DeleteLevel` (`course_game_service.go:462-473`) that references a non-existent `game_level_id` column on `content_items`/`content_metas` — noted in the spec, to be addressed in a separate task
- Running the backfill against the production DB (this plan targets the dev DB in Docker; production rollout is a separate operational step)

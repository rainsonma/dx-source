---
title: 1:1 Backfill of Content Metas and Game Metas from Imported Content Items
date: 2026-04-15
status: approved
related:
  - dx-api/app/consts/source_from.go
  - dx-api/app/models/content_meta.go
  - dx-api/app/models/content_item.go
  - dx-api/app/models/game_meta.go
  - dx-api/app/models/game_item.go
  - dx-api/app/console/commands/import_courses.go
  - dx-api/app/services/api/ai_custom_service.go
  - dx-api/app/services/api/course_content_service.go
  - dx-api/bootstrap/app.go
---

# 1:1 Backfill of Content Metas and Game Metas from Imported Content Items

## Goal

Retroactively populate `content_metas` and `game_metas` for the 1,220,803 content items that were imported by `app:import-courses` without their upstream metas. After the backfill:

- Every `content_item` has a `content_meta_id` pointing to its own newly created meta.
- Every `game_item` has a matching `game_meta` row on the same `game_level_id` at the same `order`.
- Backfilled metas carry `source_from = 'import'` so they can be distinguished from manually added (`manual`) or AI-generated (`ai`) metas.
- The existing AI-custom, course-management, and game-play workflows continue to work unchanged — no code paths need to special-case imported data.

This restores the schema invariant that every content item is owned by a content meta, unblocks future data-reuse features that query `content_metas`, and makes the imported data visible in admin/meta listings.

## Background

### Current state (verified in production DB)

```
content_items:  1,220,803 rows, 100% with content_meta_id = NULL
game_items:     1,220,803 rows, 1:1 with content_items (no orphans either side)
content_metas:  0 rows
game_metas:     0 rows
games:          663 rows, all mode='word-sentence', all status='published'
game_levels:    24,248 rows
```

The data was imported via `app:import-courses` (`app/console/commands/import_courses.go`), which creates `content_items` and `game_items` in bulk from scraped JSON files but does **not** create the upstream `content_metas` (the original scrape lost that grouping information).

### Content type distribution

| content_type | items | % |
|---|---|---|
| sentence | 569,540 | 46.6% |
| word | 331,861 | 27.2% |
| phrase | 295,456 | 24.2% |
| block | 23,946 | 2.0% |

### How the AI-custom workflow uses metas

From `ai_custom_service.go` and `ai_custom_vocab_service.go`:

- A `content_meta` is the **source unit** a user submits (one sentence, or one vocab word/phrase).
- `source_type` has two values: `sentence` (full complete sentence) or `vocab` (single word or short phrase — `processVocabBreakMeta` uses this even for multi-word phrases).
- `BreakMetadata` splits a sentence meta into multiple `content_items` (word, word, block, phrase, sentence) and links them all to the parent meta via `content_meta_id`.
- `BreakVocabMetadata` creates exactly one content_item per vocab meta (content_type `word` for single-word, `phrase` for multi-word).
- `is_break_done = true` marks a meta whose items have already been created — `BreakMetadata` filters on `is_break_done = false`, so broken metas are never re-processed.
- `GenerateContentItems` fills the per-item `items` JSONB column; it filters on `items IS NULL`, so items whose JSONB is already populated are never re-processed.

### Key guarantees that make 1:1 backfill safe

1. **No orphans.** Every content_item has exactly one game_item and vice versa (verified by SQL). A 1:1 mapping is unambiguous.
2. **Imported items already have `items` populated.** `import_courses.go:400-403` runs `transformItems()` and stores the result. `GenerateContentItems` will skip them.
3. **Imported games are permanently published.** Per product decision, these 663 games are never withdrawn, so `SaveMetadataBatch`, `BreakMetadata`, `GenerateContentItems`, `InsertContentItem`, and all other edit-path operations never run against imported levels. No capacity-check or limit code needs to change.
4. **Game playback uses only the junction → content_items path.** `GetLevelContent` never reads `content_metas` or `game_metas`, so it is entirely unaffected by the backfill.

## Decisions

| Question | Decision |
|---|---|
| Cardinality | **1:1** — one `content_meta` per `content_item`, one `game_meta` per `game_item` |
| `source_from` value | New constant `SourceFromImport = "import"` |
| `source_data` | Copy verbatim from `content_item.content` |
| `translation` | Copy verbatim from `content_item.translation` |
| `source_type` rule | By `content_type`: only `sentence` → `'sentence'`; `word`, `phrase`, and `block` all → `'vocab'` |
| `is_break_done` | `true` for all backfilled metas (items already exist) |
| `game_meta.order` | Same value as `game_item.order` (same level, lined up for `ORDER BY "order" ASC` queries) |
| `game_meta.game_id` / `game_level_id` | Copied from `game_item` |
| ID generation | UUID v7 via `uuid.Must(uuid.NewV7())`, same as the rest of the codebase |
| Idempotency | Re-run safe via `WHERE content_items.content_meta_id IS NULL` filter |
| Batching | 5000 rows per transaction |
| Workflow code changes | **None** — no touching `SaveMetadataBatch`, `BreakMetadata`, capacity limits, or anywhere else |
| Game status changes | **None** — imported games stay `published`, editing is blocked by the existing guard |

### Why `block → vocab`, not `block → sentence`

The AI-custom `BreakMetadata` prompt (`ai_custom_service.go:530-548`) defines `block` as *"a progressive combination building from the start of the sentence"* — e.g., `"I like"` is a block of `"I like the food."` A block is always a **partial** unit, never a complete sentence. Mapping it to `source_type = 'vocab'` matches the AI-custom semantic that only full complete sentences are `source_type = 'sentence'`. Everything else that isn't a complete sentence is `vocab`, including multi-word blocks and phrases. `processVocabBreakMeta` already uses `source_type = 'vocab'` for multi-word vocab entries, so this is consistent with the existing convention.

### Why 1:1 and not grouped

Earlier iterations of this design considered grouping consecutive content items into source sentences (so multiple items would share one meta, matching the natural shape of `BreakMetadata` output). Three facts made 1:1 the right answer:

1. **The data is too inconsistent.** Only ~40% of items in mixed/article levels end with terminal punctuation, and punctuation patterns vary between source articles. A reliable grouping algorithm would need an LLM, and we explicitly want this migration to be fully deterministic with no runtime AI dependency.
2. **Imported games never enter the edit flow.** 1:1 produces 1,220,803 metas — more than the capacity limits would normally allow — but the capacity check is only invoked from `SaveMetadataBatch` on published-then-withdrawn games. Since imported games are never withdrawn, the limits never fire on imported data.
3. **1:1 is trivially correct.** Every content_item maps to its own meta with identical source_data. There is no information loss, no classification ambiguity beyond the `content_type → source_type` rule, and no risk of misgrouping items from different source sentences.

## Expected backfill output

After the command completes:

| Target table | Rows inserted | Notes |
|---|---|---|
| `content_metas` | 1,220,803 | `source_from='import'`, `is_break_done=true` |
| `game_metas` | 1,220,803 | 1:1 with `game_items`, same `order` |
| `content_items` (updated) | 1,220,803 | `content_meta_id` populated |

`source_type` distribution of the new metas:

| source_type | count | derived from |
|---|---|---|
| `sentence` | 569,540 | `content_type='sentence'` |
| `vocab` | 651,263 | `content_type IN ('word','phrase','block')` |

Approximate target table sizes after backfill (estimated from current `content_items`/`game_items` sizes):

- `content_metas`: ~800 MB (same columns as content_items minus audio/jsonb/tags columns)
- `game_metas`: ~350 MB (same columns as game_items)

## Architecture

### 1. New `SourceFromImport` constant

**File:** `dx-api/app/consts/source_from.go`

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

No other code references `SourceFromLabels` in a way that requires exhaustive matching (verified by grep), so adding the new entry is additive-only.

### 2. New artisan command `app:backfill-metas`

**File:** `dx-api/app/console/commands/backfill_metas.go` (new)

```go
package commands

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

	// 1. Count what we will process.
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

	// 2. Process in batches.
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
			// No more rows match the filter — done even if count disagreed
			// (shouldn't happen but defensive).
			break
		}
		processed += int64(n)
		ctx.Info(fmt.Sprintf("[%d/%d] done in %s", processed, total, time.Since(start)))
	}

	ctx.NewLine()
	ctx.Info(fmt.Sprintf("backfill complete: %d rows in %s", processed, time.Since(start)))
	return nil
}
```

### 3. Query and write helpers

Same file, below `Handle()`:

```go
// countBackfillCandidates returns the number of content_items needing backfill.
func countBackfillCandidates() (int64, error) {
	return facades.Orm().Query().Model(&models.ContentItem{}).
		Where("content_meta_id IS NULL").
		Count()
}

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

// backfillChunk loads up to `size` unprocessed rows, writes metas/game_metas,
// and links them back on content_items. All three writes run in a single
// transaction so the state is consistent per chunk.
func backfillChunk(size int) (int, error) {
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
		return 0, fmt.Errorf("failed to load chunk: %w", err)
	}
	if len(rows) == 0 {
		return 0, nil
	}

	// Pre-generate UUIDs for this chunk.
	metaIDs := make([]string, len(rows))
	gameMetaIDs := make([]string, len(rows))
	for i := range rows {
		metaIDs[i] = uuid.Must(uuid.NewV7()).String()
		gameMetaIDs[i] = uuid.Must(uuid.NewV7()).String()
	}

	err := facades.Orm().Transaction(func(tx orm.Query) error {
		// 2.1 Bulk insert content_metas.
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

		// 2.2 Bulk insert game_metas.
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

		// 2.3 Bulk update content_items.content_meta_id via UPDATE ... FROM VALUES.
		if err := bulkLinkItems(tx, rows, metaIDs); err != nil {
			return fmt.Errorf("link content_items: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

// deriveSourceType maps content_type to source_type per the backfill rule:
//   sentence -> sentence (complete sentence)
//   word, phrase, block -> vocab (all non-complete units)
func deriveSourceType(contentType string) string {
	if contentType == consts.ContentTypeSentence {
		return consts.SourceTypeSentence
	}
	return consts.SourceTypeVocab
}

// bulkLinkItems issues a single UPDATE that sets content_meta_id for an entire
// chunk of rows via a VALUES list. Much faster than per-row UPDATEs.
//
// Goravel uses `?` placeholders (GORM-style), not PostgreSQL-native `$N`.
// The Go string values bind as SQL text, so we cast both sides of the join
// back to uuid inside the query.
func bulkLinkItems(tx orm.Query, rows []backfillRow, metaIDs []string) error {
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
		return err
	}
	return nil
}
```

### 4. Register the command

**File:** `dx-api/bootstrap/app.go`

Add `&commands.BackfillMetas{}` to the `WithCommands` slice. Do **not** add a schedule — this is a one-shot command.

```go
WithCommands(func() []console.Command {
    return []console.Command{
        &commands.UpdatePlayStreaks{},
        &commands.ResetEnergyBeans{},
        &commands.ImportCourses{},
        &commands.ExpireStaleOrders{},
        &commands.BackfillMetas{},   // NEW
    }
}).
```

No schedule change.

### 5. No other code changes

- `app/services/api/course_content_service.go` — unchanged. `SaveMetadataBatch` capacity check is irrelevant because imported games never reach it.
- `app/services/api/ai_custom_service.go` — unchanged. `BreakMetadata` naturally skips backfilled metas via `is_break_done = false` filter.
- `app/services/api/ai_custom_vocab_service.go` — unchanged. Same reasoning.
- `app/services/api/course_game_service.go` — unchanged. `GetLevelContent` only reads `game_items`.
- Admin / public / game routes — unchanged.

## Data flow

```
┌──────────────────┐          ┌──────────────────┐
│  content_items   │  1,220,803 rows, all content_meta_id = NULL
│ (scraped data)   │
└────────┬─────────┘
         │ JOIN (1:1)
         ▼
┌──────────────────┐          ┌──────────────────┐
│   game_items     │  1,220,803 rows
└──────────────────┘

                  BACKFILL RUNS
                       │
                       ▼

┌──────────────────┐          ┌──────────────────┐
│  content_metas   │──────────│  content_items   │
│ (NEW, 1.22M)     │     FK   │ (UPDATED:        │
│ source=import    │          │  content_meta_id │
│ is_break_done=t  │          │  now populated)  │
└────────┬─────────┘          └──────────────────┘
         │
         │ FK (code-level only)
         ▼
┌──────────────────┐          ┌──────────────────┐
│   game_metas     │          │   game_items     │
│ (NEW, 1.22M)     │          │ (unchanged)      │
│ same order as gi │          │                  │
└──────────────────┘          └──────────────────┘

        ↓ queries use ORDER BY "order" ASC; gm.order == gi.order ↓
```

## Isolation & safety

### Transactions
Each chunk is one transaction. If a chunk fails mid-write, all three inserts/updates in that chunk roll back, leaving the DB in a consistent state (no half-written rows). Previously committed chunks remain. The next run resumes via the `content_meta_id IS NULL` filter.

### Idempotency
`WHERE content_meta_id IS NULL` in the selection query is the only idempotency guard we need. Once a row is linked to a meta, it's excluded from future chunk queries. A second run with the same args is a no-op.

### No concurrent writer assumption
The imported games are frozen and the edit-paths never run against imported levels, so we don't need to worry about a concurrent `SaveMetadataBatch` / `BreakMetadata` racing with the backfill on the same `content_items`. If this assumption later changes, we can add a `SELECT FOR UPDATE` on the content_items chunk, but it's not needed today.

### Blast radius of failure
- Partial completion is fine (idempotent resume).
- A bad `SourceFromImport` value or wrong `source_type` rule could be fixed by a follow-up SQL update keyed on `source_from='import'` — the import mark is the escape hatch.
- Accidental misclassification of a chunk can be undone by:
  ```sql
  UPDATE content_items SET content_meta_id = NULL
  WHERE content_meta_id IN (SELECT id FROM content_metas WHERE source_from = 'import');
  DELETE FROM content_metas WHERE source_from = 'import';
  DELETE FROM game_metas WHERE content_meta_id NOT IN (SELECT id FROM content_metas);
  ```
  The `source_from = 'import'` mark is what makes this reversal tractable.

## Performance

Estimated cost per chunk of 5000 rows:

- 1 SELECT with JOIN over `content_items` + `game_items` on indexed columns: ~50–200 ms
- 1 INSERT of 5000 `content_metas`: ~100–300 ms
- 1 INSERT of 5000 `game_metas`: ~100–300 ms
- 1 UPDATE with VALUES-based join over 5000 rows: ~200–500 ms
- Transaction overhead: ~20 ms

Expected per-chunk wall time: **~500 ms – 1.5 s**.

Total chunks: `ceil(1,220,803 / 5000) = 245`.

Total wall time: **~2 – 6 minutes** single-threaded. Acceptable for a one-shot migration, no parallelism needed.

## Testing

### Pre-backfill verification

```sql
-- Expect: 0
SELECT COUNT(*) FROM content_metas WHERE source_from = 'import';

-- Expect: 1220803
SELECT COUNT(*) FROM content_items WHERE content_meta_id IS NULL;

-- Expect: 0 orphans either way
SELECT COUNT(*) FROM content_items ci
  LEFT JOIN game_items gi ON gi.content_item_id = ci.id AND gi.deleted_at IS NULL
  WHERE ci.deleted_at IS NULL AND gi.id IS NULL;
SELECT COUNT(*) FROM game_items gi
  LEFT JOIN content_items ci ON ci.id = gi.content_item_id AND ci.deleted_at IS NULL
  WHERE gi.deleted_at IS NULL AND ci.id IS NULL;
```

### Dry-run first

```bash
cd dx-api
go run . artisan app:backfill-metas --dry-run
# Expected: "backfill candidates: 1220803" then "dry-run — no writes"
```

### Small-batch test

```bash
go run . artisan app:backfill-metas --limit 100 --batch-size 50
# Expected: 2 chunks, 100 total rows processed
```

### Post-backfill verification

```sql
-- Expect: 1220803
SELECT COUNT(*) FROM content_metas WHERE source_from = 'import' AND is_break_done = true;

-- Expect: 1220803
SELECT COUNT(*) FROM game_metas;

-- Expect: 0
SELECT COUNT(*) FROM content_items WHERE content_meta_id IS NULL AND deleted_at IS NULL;

-- Expect: 0 (no meta without its item link)
SELECT COUNT(*) FROM content_metas cm
  WHERE cm.source_from = 'import'
    AND NOT EXISTS (SELECT 1 FROM content_items ci WHERE ci.content_meta_id = cm.id);

-- Expect: 0 (no game_meta without matching game_item at same level)
SELECT COUNT(*) FROM game_metas gm
  WHERE NOT EXISTS (
    SELECT 1 FROM game_items gi
    WHERE gi.game_level_id = gm.game_level_id
      AND gi."order" = gm."order"
      AND gi.deleted_at IS NULL
  );

-- source_type distribution matches content_type distribution
SELECT cm.source_type, COUNT(*)
  FROM content_metas cm
  WHERE cm.source_from = 'import'
  GROUP BY cm.source_type;
-- Expect: sentence = 569540, vocab = 651263
```

### Workflow smoke tests (manual)

1. **Game play**: Open any imported game in the web UI, start a level, play through. All items load in order, no missing content, no errors. (Game play uses only `game_items` → `content_items`, so this should be unaffected.)
2. **Admin meta listing** (if available): Verify imported metas appear with `source_from=import` and `is_break_done=true`.
3. **AI custom sanity check**: Create a new user-owned game, run through SaveMetadataBatch → BreakMetadata → GenerateContentItems. Verify the new flow still works end-to-end — imported metas must not interfere.

### Lint / build

```bash
cd dx-api
go build ./...
go vet ./...
go test -race ./app/...
```

No new test file is added — the command is a one-shot data migration, not a reusable service. Correctness is verified by the post-backfill SQL assertions above.

## Rollout

This is a one-shot local operation against a single Postgres instance.

1. Review the spec and code.
2. Run `go build ./...` to confirm the new command compiles.
3. Run `--dry-run` to confirm candidate count is 1,220,803.
4. Run `--limit 100 --batch-size 50` to exercise the write path on a small subset.
5. Verify the post-backfill SQL on that subset.
6. Run the full `app:backfill-metas` with default args.
7. Run the post-backfill verification SQL.
8. Run the workflow smoke tests.

Rollback is via the SQL shown in **Blast radius of failure** above — the `source_from='import'` mark makes backfilled data fully reversible.

## Out of scope

The following are explicitly **not** part of this spec and will not be changed:

- Capacity limits in `SaveMetadataBatch` (irrelevant because imported games never enter the edit flow)
- `MaxSentences`, `MaxVocab`, `MaxMetasPerLevel`, `MaxItemsPerMeta` (unchanged)
- Any grouping of content items into shared metas
- AI-based reconstruction of source sentences
- Pre-existing bug in `DeleteLevel` (`course_game_service.go:462-473`) which references a non-existent `game_level_id` column on `content_items` / `content_metas` — noted for a separate fix
- Any frontend or UI changes
- Any changes to the `import_courses` command going forward (it continues to import without metas; this command backfills after the fact)

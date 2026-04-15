# Content Meta & Item Deduplication Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

## Resumption status (2026-04-15)

Two pieces of context for anyone picking this plan up mid-flight:

1. **The `game_metas` / `game_items` unique → non-unique index relax is already DONE.** It was split into its own PR and merged directly into the create migration (`20260414000001_create_game_metas_and_game_items_tables.go`). Both tables now use `table.Index(...)` with the original names `idx_game_metas_level_meta` and `idx_game_items_level_item`. **Task 2 below has been rewritten accordingly** — it now creates ONLY the `idx_content_metas_dedup_lookup` index. The historical relax steps, the create-new-then-drop-old safety dance, and the `CONCURRENTLY` discussion no longer apply to the resumed work.
2. **The `app:backfill-metas` command landed.** `content_metas` now contains ~1.22M `source_from='import'` rows owned by the 1,202 oldest real users. Under the spec's identity rule (which does NOT include `source_from`), these imported metas are legitimate dedup candidates for their owning users. No logic change — this is already the spec's intent. It does mean pre-migration `pg_dump` (Task 1) matters more now: the affected table is no longer trivially small.

Everything else in this plan — Tasks 1, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17 — is still pending and still valid as written. Execute in order.

## Local environment note (required reading for every task)

Two non-obvious facts about the local dev environment that the plan's literal shell commands don't reflect:

1. **The database name is `dxdb`, not `douxue`.** Everywhere below you see `douxue` in a `psql`/`pg_dump`/`pg_restore` command, change it to `dxdb`. (File system paths like `/Users/rainsen/Programs/Projects/douxue/...` are the path to the project folder and should NOT change.)
2. **No host-side `psql`/`pg_dump`/`pg_restore`.** Postgres runs inside the docker-compose `postgres` service defined in `deploy/docker-compose.dev.yml`. Every DB command must go through `docker compose -f deploy/docker-compose.dev.yml exec -T postgres ...` from the `dx-source/` repo root. Example template for the row-count verification command:
    ```
    docker compose -f deploy/docker-compose.dev.yml exec -T postgres \
      psql -U postgres -d dxdb -c "SELECT ..."
    ```
    For `pg_dump` you pipe stdout from `docker compose exec -T postgres pg_dump ...` to a host file. For `pg_restore --list`, copy the dump into the container first with `docker compose cp` and read it from there.

Baseline row counts captured by Task 1 before the Task 2 migration (used for post-migration verification — counts must match afterward):

| Table | Live row count |
|---|---|
| content_metas | 1,220,803 |
| content_items | 1,220,803 |
| game_metas | 1,220,803 |
| game_items | 1,220,803 |

**Goal:** When a user adds metadata to a level, reuse existing identical `content_metas` (and any associated broken-down `content_items`) by creating new junction rows instead of inserting duplicate underlying rows. Update delete paths to be reference-counted so reuse is safe.

**Architecture:** A single Goravel migration adds `idx_content_metas_dedup_lookup` on `content_metas (source_type, source_data) WHERE deleted_at IS NULL` — the junction non-unique indexes were already merged into their create migration and are out of scope here. `SaveMetadataBatch` gains a per-batch dedup map keyed on `(source_type, source_data, normalized_translation)` scoped to the current user's own games. Reused metas with `is_break_done = true` get parallel `game_items` rows in the new level pointing at existing `content_items`. Three delete service functions are rewritten to soft-delete junctions for the current scope and only soft-delete underlying rows when no live junctions remain anywhere. Two REST routes change shape to carry `levelId` in the path.

**Tech Stack:** Go 1.x + Goravel framework, GORM, PostgreSQL, Next.js 16 (frontend, minor edit), stretchr/testify suite for tests.

**Spec:** `docs/superpowers/specs/2026-04-14-content-meta-dedup-design.md`

---

## File Structure

| File | Action | Responsibility |
|---|---|---|
| `dx-api/database/migrations/20260415000001_add_content_metas_dedup_lookup_index.go` | Create | Add `idx_content_metas_dedup_lookup` on `content_metas (source_data, source_type) WHERE deleted_at IS NULL` |
| `dx-api/app/services/api/course_content_service.go` | Modify | `SaveMetadataBatch` rewrite (dedup loop), three delete functions rewritten with reference counting |
| `dx-api/app/http/controllers/api/course_game_controller.go` | Modify | `DeleteMetadata` and `DeleteContentItem` plumb `levelId` from new path params |
| `dx-api/routes/api.go` | Modify | Two DELETE routes get `/levels/{levelId}/` injected |
| `dx-api/tests/feature/course_content_dedup_test.go` | Create | Integration tests for save dedup, items reuse, and reference-counted deletes |
| `dx-web/src/features/web/ai-custom/actions/course-game.action.ts` | Modify | `deleteMetaAction` / `deleteContentItemAction` accept `levelId` and use new URL paths |
| `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx` | Modify | Two call sites pass `levelId` |

**Backup directory** (referenced in pre-migration step): `/Users/rainsen/Programs/Projects/douxue/db-backup/`

---

## Task 1: Pre-flight backup & verification

**Files:** none (operational)

- [ ] **Step 1: Verify the backup directory exists and is writable**

```bash
mkdir -p /Users/rainsen/Programs/Projects/douxue/db-backup
ls -la /Users/rainsen/Programs/Projects/douxue/db-backup
```

Expected: directory exists, writable.

- [ ] **Step 2: Capture a `pg_dump` snapshot of the current `douxue` database**

```bash
TS=$(date +%Y%m%d-%H%M%S)
pg_dump --host=localhost --port=5432 --username=postgres --no-owner --no-privileges \
  --format=custom douxue > /Users/rainsen/Programs/Projects/douxue/db-backup/dx-${TS}.dump
ls -lh /Users/rainsen/Programs/Projects/douxue/db-backup/dx-${TS}.dump
```

Expected: a non-empty `.dump` file. If credentials differ, read `dx-api/.env` for `DB_USERNAME`/`DB_PASSWORD` and adjust. Set `PGPASSWORD` env var if needed.

- [ ] **Step 3: Verify the dump is restorable (smoke test)**

```bash
pg_restore --list /Users/rainsen/Programs/Projects/douxue/db-backup/dx-${TS}.dump | head -20
```

Expected: a list of TOC entries (tables, indexes, sequences) — confirms the dump is intact.

- [ ] **Step 4: Note current row counts of affected tables**

```bash
psql -h localhost -U postgres -d douxue -c "SELECT 'content_metas' AS t, COUNT(*) FROM content_metas WHERE deleted_at IS NULL UNION ALL SELECT 'content_items', COUNT(*) FROM content_items WHERE deleted_at IS NULL UNION ALL SELECT 'game_metas', COUNT(*) FROM game_metas WHERE deleted_at IS NULL UNION ALL SELECT 'game_items', COUNT(*) FROM game_items WHERE deleted_at IS NULL;"
```

Expected: row counts printed. Save them to compare after migration. The counts must remain unchanged after the migration runs (the migration only touches indexes).

- [ ] **Step 5: Commit a marker file documenting the backup**

```bash
echo "Backup taken: dx-${TS}.dump (pre-content-meta-dedup migration)" >> /Users/rainsen/Programs/Projects/douxue/dx-source/.backup-log
git add /Users/rainsen/Programs/Projects/douxue/dx-source/.backup-log
git commit -m "chore: log pre-dedup migration backup"
```

Expected: commit succeeds.

---

## Task 2: Migration — add `idx_content_metas_dedup_lookup`

**Files:**
- Create: `dx-api/database/migrations/20260415000001_add_content_metas_dedup_lookup_index.go`

**Scope note.** The historical "relax junction unique indexes to non-unique" step that used to live in this task is already DONE — the non-unique indexes were merged directly into the junction create migration. This task now adds exactly one new index on `content_metas`: the dedup-lookup support index that the save-path SELECT in Task 6 depends on.

**Goravel migration pattern reminder (from repo memory).** `facades.Schema().Create(...)` and `facades.Orm().Query().Exec(...)` use separate database connections, so mixing `Schema.Create` and raw-SQL DDL in the same migration file is unsafe. This migration is 100% raw-SQL DDL, so it's fine as a single file.

- [ ] **Step 1: Confirm how migrations are registered**

```bash
ls /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/database/
```

Then check the registration pattern:

```bash
grep -rn "20260414000001\|migrations\." /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/database/ 2>/dev/null | head -20
```

Expected: locate the registration file (likely `dx-api/database/kernel.go`) and confirm the pattern. If migrations are auto-discovered from the `migrations/` directory, no registration is needed.

- [ ] **Step 2: Create the migration file**

Create `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/database/migrations/20260415000001_add_content_metas_dedup_lookup_index.go`:

```go
package migrations

import (
	"github.com/goravel/framework/facades"
)

type M20260415000001AddContentMetasDedupLookupIndex struct{}

func (r *M20260415000001AddContentMetasDedupLookupIndex) Signature() string {
	return "20260415000001_add_content_metas_dedup_lookup_index"
}

func (r *M20260415000001AddContentMetasDedupLookupIndex) Up() error {
	// Supports the per-user dedup SELECT in SaveMetadataBatch:
	//   WHERE cm.deleted_at IS NULL
	//     AND cm.source_data IN ?
	//     AND cm.source_type IN ?
	//   AND <join to games via user_id>
	//
	// Column order is (source_data, source_type) because source_data has
	// ~millions of distinct values on our dataset while source_type only has
	// 2 ('sentence' | 'vocab'). Leading with the more-selective column
	// narrows the B-tree scan aggressively before the second-column filter
	// runs — ~2-5x faster than the reverse order on the 1.22M-row table.
	_, err := facades.Orm().Query().Exec(
		`CREATE INDEX IF NOT EXISTS idx_content_metas_dedup_lookup
		   ON content_metas (source_data, source_type)
		   WHERE deleted_at IS NULL`,
	)
	return err
}

func (r *M20260415000001AddContentMetasDedupLookupIndex) Down() error {
	_, err := facades.Orm().Query().Exec(
		`DROP INDEX IF EXISTS idx_content_metas_dedup_lookup`,
	)
	return err
}
```

- [ ] **Step 3: Register the migration if the project requires it**

If Step 1 found an explicit registration file (e.g., `dx-api/database/kernel.go` with a slice of migrations), append the new migration there following the existing pattern. If migrations are auto-discovered, skip this step.

- [ ] **Step 4: Build**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: build succeeds with no errors.

- [ ] **Step 5: Run the migration**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go run . migrate
```

Expected: migration logs success. Should report the new migration applied.

**Locking note.** `CREATE INDEX` (without `CONCURRENTLY`) takes a `SHARE` lock on `content_metas` that blocks writes while the index builds. With ~1.22M rows the build can take a few seconds. For local dev this is fine. For a production Postgres with live writers, replace the body of `Up()` with `CREATE INDEX CONCURRENTLY`; Goravel's migration runner does NOT wrap migrations in a transaction, so `CONCURRENTLY` is permitted.

- [ ] **Step 6: Verify the new index in PostgreSQL**

```bash
psql -h localhost -U postgres -d douxue -c "SELECT indexname, indexdef FROM pg_indexes WHERE indexname = 'idx_content_metas_dedup_lookup';"
```

Expected: exactly one row showing the `CREATE INDEX ... ON public.content_metas USING btree (source_data, source_type) WHERE (deleted_at IS NULL)` definition.

Also confirm the pre-existing junction indexes are untouched:

```bash
psql -h localhost -U postgres -d douxue -c "SELECT indexname FROM pg_indexes WHERE indexname IN ('idx_game_metas_level_meta','idx_game_items_level_item') ORDER BY indexname;"
```

Expected: both names present (already relaxed to non-unique earlier).

- [ ] **Step 7: Verify row counts are unchanged**

```bash
psql -h localhost -U postgres -d douxue -c "SELECT 'content_metas' AS t, COUNT(*) FROM content_metas WHERE deleted_at IS NULL UNION ALL SELECT 'content_items', COUNT(*) FROM content_items WHERE deleted_at IS NULL UNION ALL SELECT 'game_metas', COUNT(*) FROM game_metas WHERE deleted_at IS NULL UNION ALL SELECT 'game_items', COUNT(*) FROM game_items WHERE deleted_at IS NULL;"
```

Expected: identical to the counts captured in Task 1 Step 4. Adding an index is non-destructive; if counts differ, STOP and investigate.

- [ ] **Step 8: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-api/database/migrations/20260415000001_add_content_metas_dedup_lookup_index.go && git commit -m "feat(api): add content_metas dedup-lookup index"
```

Expected: commit succeeds.

---

## Task 3: Test scaffolding — DB seed helpers

**Files:**
- Create: `dx-api/tests/feature/course_content_dedup_test.go` (test file with helpers; no test functions yet)

- [ ] **Step 1: Create the test file with imports and helpers**

Create `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/tests/feature/course_content_dedup_test.go`:

```go
package feature

import (
	"testing"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	"dx-api/app/consts"
	"dx-api/app/models"
	api "dx-api/app/services/api"
	"dx-api/tests"
)

type ContentDedupSuite struct {
	suite.Suite
	tests.TestCase
	userID string
}

func TestContentDedupSuite(t *testing.T) {
	suite.Run(t, new(ContentDedupSuite))
}

// SetupTest creates an isolated VIP user for each test and cleans up after.
func (s *ContentDedupSuite) SetupTest() {
	s.userID = s.seedVipUser()
}

func (s *ContentDedupSuite) TearDownTest() {
	if s.userID == "" {
		return
	}
	// Cascade soft-delete via raw SQL — content tests own all rows under this user.
	q := facades.Orm().Query()
	_, _ = q.Exec(`DELETE FROM game_items WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM game_metas WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM content_items WHERE content_meta_id IN (SELECT cm.id FROM content_metas cm JOIN game_metas gm ON gm.content_meta_id = cm.id JOIN games g ON g.id = gm.game_id WHERE g.user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM content_metas WHERE id IN (SELECT cm.id FROM content_metas cm LEFT JOIN game_metas gm ON gm.content_meta_id = cm.id LEFT JOIN games g ON g.id = gm.game_id WHERE g.user_id = ? OR g.user_id IS NULL AND cm.created_at > NOW() - INTERVAL '1 hour')`, s.userID)
	_, _ = q.Exec(`DELETE FROM game_levels WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, s.userID)
	_, _ = q.Exec(`DELETE FROM games WHERE user_id = ?`, s.userID)
	_, _ = q.Exec(`DELETE FROM users WHERE id = ?`, s.userID)
	s.userID = ""
}

// seedVipUser inserts a lifetime-grade user and returns its id.
func (s *ContentDedupSuite) seedVipUser() string {
	id := uuid.Must(uuid.NewV7()).String()
	user := models.User{
		ID:         id,
		Username:   "test_" + id[:8],
		Grade:      consts.UserGradeLifetime,
		IsActive:   true,
		InviteCode: id[:8],
		Password:   "x",
	}
	s.Require().NoError(facades.Orm().Query().Create(&user))
	return id
}

// seedGame inserts a draft course game owned by s.userID with the given mode.
func (s *ContentDedupSuite) seedGame(mode string) string {
	id := uuid.Must(uuid.NewV7()).String()
	g := models.Game{
		ID:       id,
		Name:     "test game " + id[:8],
		UserID:   &s.userID,
		Mode:     mode,
		IsActive: true,
		Status:   consts.GameStatusDraft,
	}
	s.Require().NoError(facades.Orm().Query().Create(&g))
	return id
}

// seedLevel inserts a level row for the given game and returns its id.
func (s *ContentDedupSuite) seedLevel(gameID string) string {
	id := uuid.Must(uuid.NewV7()).String()
	lv := models.GameLevel{
		ID:       id,
		GameID:   gameID,
		Name:     "L1",
		IsActive: true,
		Order:    1000,
	}
	s.Require().NoError(facades.Orm().Query().Create(&lv))
	return id
}

// strPtr returns a pointer to the given string.
func strPtr(v string) *string { return &v }

// countMetasOwnedByUser returns the number of live content_metas reachable
// from the given user via the junction chain.
func (s *ContentDedupSuite) countMetasOwnedByUser(userID string) int64 {
	var n int64
	row := struct{ N int64 }{}
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT COUNT(DISTINCT cm.id) AS n
		   FROM content_metas cm
		   JOIN game_metas gm ON gm.content_meta_id = cm.id AND gm.deleted_at IS NULL
		   JOIN game_levels gl ON gl.id = gm.game_level_id AND gl.deleted_at IS NULL
		   JOIN games g ON g.id = gl.game_id AND g.deleted_at IS NULL
		  WHERE cm.deleted_at IS NULL AND g.user_id = ?`,
		userID,
	).Scan(&row))
	n = row.N
	return n
}

// countGameMetasInLevel returns the number of live game_metas in the level.
func (s *ContentDedupSuite) countGameMetasInLevel(levelID string) int64 {
	n, err := facades.Orm().Query().Model(&models.GameMeta{}).
		Where("game_level_id", levelID).Count()
	s.Require().NoError(err)
	return n
}

// countGameItemsInLevel returns the number of live game_items in the level.
func (s *ContentDedupSuite) countGameItemsInLevel(levelID string) int64 {
	n, err := facades.Orm().Query().Model(&models.GameItem{}).
		Where("game_level_id", levelID).Count()
	s.Require().NoError(err)
	return n
}

// Smoke test — verifies the suite boots and the seed helpers work.
func (s *ContentDedupSuite) TestSetup_BootsAndSeeds() {
	s.NotEmpty(s.userID)
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)
	s.NotEmpty(gameID)
	s.NotEmpty(levelID)
}

// silence unused-import warning until later tasks add tests using the api package.
var _ = api.SaveMetadataBatch
```

- [ ] **Step 2: Verify game_levels has these fields and the consts exist**

```bash
grep -n "Name\|Order\|IsActive" /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/models/game_level.go
grep -rn "GameModeWordSentence\|GameStatusDraft" /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/consts/ | head
```

Expected: confirms the field names and consts exist. If a field name differs (e.g., `Title` instead of `Name`), update the helper.

If `consts.GameModeWordSentence` doesn't exist, find the actual constant name with:
```bash
grep -rn "word.sentence\|word-sentence\|WordSentence" /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/consts/
```

- [ ] **Step 3: Build and run only the smoke test**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestSetup_BootsAndSeeds" ./tests/feature/ -v
```

Expected: PASS. If the test fails because of missing fields/consts, fix the helpers and re-run.

- [ ] **Step 4: Commit the test scaffolding**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-api/tests/feature/course_content_dedup_test.go && git commit -m "test(api): scaffold dedup test suite with VIP user/game/level helpers"
```

Expected: commit succeeds.

---

## Task 4: Test — fresh save inserts and counts (TDD baseline)

**Files:**
- Modify: `dx-api/tests/feature/course_content_dedup_test.go`

- [ ] **Step 1: Add the baseline test**

Append to the file (before the `var _ = api.SaveMetadataBatch` line — and DELETE that placeholder line now since `api.SaveMetadataBatch` is used by real tests):

```go
// TestSave_FreshEntries_AllCreated verifies that submitting brand-new metadata
// creates one content_metas row and one game_metas row per entry.
func (s *ContentDedupSuite) TestSave_FreshEntries_AllCreated() {
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)

	entries := []api.MetadataEntry{
		{SourceData: "Hello world.", Translation: strPtr("你好世界。"), SourceType: "sentence"},
		{SourceData: "apple", Translation: strPtr("苹果"), SourceType: "vocab"},
	}

	count, err := api.SaveMetadataBatch(s.userID, gameID, levelID, entries, "manual")
	s.Require().NoError(err)
	s.Equal(2, count)

	s.Equal(int64(2), s.countMetasOwnedByUser(s.userID))
	s.Equal(int64(2), s.countGameMetasInLevel(levelID))
}
```

- [ ] **Step 2: Run the test — it should PASS today**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestSave_FreshEntries_AllCreated" ./tests/feature/ -v
```

Expected: PASS. The current code doesn't dedup, but for fresh entries the dedup loop's behavior should match the existing loop. If it fails, debug helpers/seed data before continuing.

- [ ] **Step 3: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-api/tests/feature/course_content_dedup_test.go && git commit -m "test(api): baseline test for fresh metadata save"
```

---

## Task 5: Test — cross-game dedup (RED, the new behavior)

**Files:**
- Modify: `dx-api/tests/feature/course_content_dedup_test.go`

- [ ] **Step 1: Add the failing test for cross-game reuse**

Append:

```go
// TestSave_DedupAcrossGames_ReusesContentMeta verifies that saving the same
// content into a second game owned by the same user reuses the existing
// content_metas row instead of creating a new one.
func (s *ContentDedupSuite) TestSave_DedupAcrossGames_ReusesContentMeta() {
	// Game A — original content
	gameA := s.seedGame(consts.GameModeWordSentence)
	levelA := s.seedLevel(gameA)
	entries := []api.MetadataEntry{
		{SourceData: "shared sentence", Translation: strPtr("共享句子"), SourceType: "sentence"},
	}
	_, err := api.SaveMetadataBatch(s.userID, gameA, levelA, entries, "manual")
	s.Require().NoError(err)
	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID))

	// Game B — same content, different game
	gameB := s.seedGame(consts.GameModeWordSentence)
	levelB := s.seedLevel(gameB)
	_, err = api.SaveMetadataBatch(s.userID, gameB, levelB, entries, "manual")
	s.Require().NoError(err)

	// content_metas row count must still be 1 (deduped), but both levels
	// have one game_metas junction row each.
	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID), "content_metas should be reused")
	s.Equal(int64(1), s.countGameMetasInLevel(levelA))
	s.Equal(int64(1), s.countGameMetasInLevel(levelB))
}
```

- [ ] **Step 2: Run — expected to FAIL**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestSave_DedupAcrossGames_ReusesContentMeta" ./tests/feature/ -v
```

Expected: FAIL on `s.Equal(int64(1), s.countMetasOwnedByUser(s.userID), "content_metas should be reused")` — the current code creates a second content_meta row, so the count is 2. **This is the RED step.**

- [ ] **Step 3: Commit the failing test**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-api/tests/feature/course_content_dedup_test.go && git commit -m "test(api): RED — cross-game dedup expects content_metas reuse"
```

---

## Task 6: Implement dedup loop in SaveMetadataBatch

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go` (specifically the `SaveMetadataBatch` function and add new private helpers in the same file)

- [ ] **Step 1: Read the current `SaveMetadataBatch` to confirm line ranges**

```bash
sed -n '50,170p' /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/services/api/course_content_service.go
```

Expected: prints `SaveMetadataBatch` from definition through the final `return`. Note the line numbers of the loop you will replace (currently lines 137-163 according to the spec; verify before editing).

- [ ] **Step 2: Add dedup helpers above `SaveMetadataBatch`**

Insert these private types and functions IMMEDIATELY ABOVE `func SaveMetadataBatch` (i.e., after the existing `ContentMetaData` struct, before `// SaveMetadataBatch creates content metadata...`):

```go
// metaDedupKey is the identity tuple used for content_metas reuse.
// Translation is normalized: a nil/empty translation collapses to "".
type metaDedupKey struct {
	SourceType  string
	SourceData  string
	Translation string
}

func makeMetaDedupKey(e MetadataEntry) metaDedupKey {
	t := ""
	if e.Translation != nil {
		t = *e.Translation
	}
	return metaDedupKey{e.SourceType, e.SourceData, t}
}

// existingMetaRef is a content_metas row already owned by the user that
// can be reused on save.
type existingMetaRef struct {
	ID          string `gorm:"column:id"`
	SourceType  string `gorm:"column:source_type"`
	SourceData  string `gorm:"column:source_data"`
	Translation string `gorm:"column:translation"` // COALESCE'd to ""
	IsBreakDone bool   `gorm:"column:is_break_done"`
}

// findExistingMetasForBatch loads, in a single query, all content_metas rows
// owned by userID that match any (source_type, source_data) pair in the
// batch. Returns a map keyed on metaDedupKey; first match wins per key.
func findExistingMetasForBatch(userID string, entries []MetadataEntry) (map[metaDedupKey]existingMetaRef, error) {
	if len(entries) == 0 {
		return map[metaDedupKey]existingMetaRef{}, nil
	}

	typeSet := map[string]struct{}{}
	dataSet := map[string]struct{}{}
	for _, e := range entries {
		typeSet[e.SourceType] = struct{}{}
		dataSet[e.SourceData] = struct{}{}
	}
	sourceTypes := make([]string, 0, len(typeSet))
	for t := range typeSet {
		sourceTypes = append(sourceTypes, t)
	}
	sourceData := make([]string, 0, len(dataSet))
	for d := range dataSet {
		sourceData = append(sourceData, d)
	}

	var rows []existingMetaRef
	if err := facades.Orm().Query().Raw(
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
	).Scan(&rows); err != nil {
		return nil, fmt.Errorf("failed to query existing metas for dedup: %w", err)
	}

	out := make(map[metaDedupKey]existingMetaRef, len(rows))
	for _, r := range rows {
		key := metaDedupKey{r.SourceType, r.SourceData, r.Translation}
		if _, exists := out[key]; !exists {
			out[key] = r
		}
	}
	return out, nil
}
```

- [ ] **Step 3: Replace the create-loop in `SaveMetadataBatch`**

Find the block from `// Create metas in batch` to the closing `}` of the loop (currently lines ~137-163). Replace it with:

```go
	// Create metas in batch — dedup against the user's existing content.
	existingByKey, err := findExistingMetasForBatch(userID, entries)
	if err != nil {
		return 0, err
	}

	// State carried across the entry loop for items reuse.
	itemsByMetaCache := map[string][]string{} // metaID -> ordered content_item IDs
	var maxItemOrderInLevel *float64
	itemsAddedSoFar := 0

	if err := facades.Orm().Transaction(func(tx orm.Query) error {
		for i, e := range entries {
			key := makeMetaDedupKey(e)

			var metaID string
			var isBreakDone bool
			if existing, ok := existingByKey[key]; ok {
				metaID = existing.ID
				isBreakDone = existing.IsBreakDone
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
				existingByKey[key] = existingMetaRef{
					ID:          metaID,
					SourceType:  e.SourceType,
					SourceData:  e.SourceData,
					Translation: key.Translation,
					IsBreakDone: false,
				}
				isBreakDone = false
			}

			// Always create a fresh game_metas junction row (allows in-level repetition).
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

			// If we are reusing a meta that has already been broken down,
			// also create game_items rows in this level pointing at the
			// existing content_items.
			if isBreakDone {
				if err := reuseItemsIntoLevel(
					tx, metaID, gameLevelID, level.GameID,
					itemsByMetaCache, &maxItemOrderInLevel, &itemsAddedSoFar,
				); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return 0, err
	}

	return len(entries), nil
}
```

- [ ] **Step 4: Add the `reuseItemsIntoLevel` helper at the bottom of the file**

Append after the existing `calculateInsertionOrder` function:

```go
// reuseItemsIntoLevel creates game_items rows in gameLevelID for every active
// content_item belonging to metaID. Item IDs are loaded once per metaID via
// itemsByMetaCache. The level's pre-save max game_items.order is loaded once
// via maxItemOrderInLevel. itemsAddedSoFar is incremented for every new
// game_items row to keep ordering monotonically increasing across multiple
// reused metas in the same batch.
func reuseItemsIntoLevel(
	tx orm.Query,
	metaID, gameLevelID, gameID string,
	itemsByMetaCache map[string][]string,
	maxItemOrderInLevel **float64,
	itemsAddedSoFar *int,
) error {
	itemIDs, ok := itemsByMetaCache[metaID]
	if !ok {
		var rows []struct {
			ID string `gorm:"column:id"`
		}
		if err := tx.Raw(
			`SELECT id FROM content_items
			  WHERE content_meta_id = ? AND deleted_at IS NULL
			  ORDER BY id`,
			metaID,
		).Scan(&rows); err != nil {
			return fmt.Errorf("failed to load content_items for reuse: %w", err)
		}
		itemIDs = make([]string, 0, len(rows))
		for _, r := range rows {
			itemIDs = append(itemIDs, r.ID)
		}
		itemsByMetaCache[metaID] = itemIDs
	}
	if len(itemIDs) == 0 {
		return nil
	}

	if *maxItemOrderInLevel == nil {
		var row struct {
			MaxOrder float64 `gorm:"column:max_order"`
		}
		if err := tx.Raw(
			`SELECT COALESCE(MAX("order"), 0) AS max_order
			   FROM game_items
			  WHERE game_level_id = ? AND deleted_at IS NULL`,
			gameLevelID,
		).Scan(&row); err != nil {
			return fmt.Errorf("failed to load max game_items order: %w", err)
		}
		v := row.MaxOrder
		*maxItemOrderInLevel = &v
	}

	for j, contentItemID := range itemIDs {
		gi := models.GameItem{
			ID:            uuid.Must(uuid.NewV7()).String(),
			GameID:        gameID,
			GameLevelID:   gameLevelID,
			ContentItemID: contentItemID,
			Order:         **maxItemOrderInLevel + float64((*itemsAddedSoFar+j+1)*1000),
		}
		if err := tx.Create(&gi); err != nil {
			return fmt.Errorf("failed to create game item: %w", err)
		}
	}
	*itemsAddedSoFar += len(itemIDs)
	return nil
}
```

- [ ] **Step 5: Build and check for compile errors**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: build succeeds. Likely fix needed: ensure `orm` is imported in `course_content_service.go`. Check at top of file — if `"github.com/goravel/framework/contracts/database/orm"` is missing, add it. Run `goimports -w app/services/api/course_content_service.go` if available.

- [ ] **Step 6: Run the cross-game dedup test — should now PASS**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestSave_DedupAcrossGames_ReusesContentMeta" ./tests/feature/ -v
```

Expected: PASS. Also re-run the baseline test to confirm we didn't regress:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestSave_FreshEntries_AllCreated|TestContentDedupSuite/TestSave_DedupAcrossGames_ReusesContentMeta" ./tests/feature/ -v
```

Expected: both PASS.

- [ ] **Step 7: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-api/app/services/api/course_content_service.go && git commit -m "feat(api): dedup content_metas on save, scoped to current user"
```

---

## Task 7: Test — within-batch repetition allowed

**Files:**
- Modify: `dx-api/tests/feature/course_content_dedup_test.go`

- [ ] **Step 1: Add the test**

Append:

```go
// TestSave_WithinBatchRepetition_OneMetaTwoJunctions verifies that submitting
// the same entry twice in one batch creates ONE content_meta row but TWO
// game_meta junction rows.
func (s *ContentDedupSuite) TestSave_WithinBatchRepetition_OneMetaTwoJunctions() {
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)

	entries := []api.MetadataEntry{
		{SourceData: "repeat me", Translation: strPtr("重复"), SourceType: "vocab"},
		{SourceData: "repeat me", Translation: strPtr("重复"), SourceType: "vocab"},
	}

	count, err := api.SaveMetadataBatch(s.userID, gameID, levelID, entries, "manual")
	s.Require().NoError(err)
	s.Equal(2, count)

	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID), "content_metas deduped within batch")
	s.Equal(int64(2), s.countGameMetasInLevel(levelID), "two junction rows created")
}
```

- [ ] **Step 2: Run the test**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestSave_WithinBatchRepetition_OneMetaTwoJunctions" ./tests/feature/ -v
```

Expected: PASS. (The Task 6 implementation already supports this — the test verifies it.)

If FAIL, confirm the junction index on `game_metas` is non-unique (it should already be, from the earlier merge into the create migration):
```bash
psql -h localhost -U postgres -d douxue -c "\d game_metas"
```
Expect `idx_game_metas_level_meta` present and NOT unique. If it's still unique, something is out of sync with the current state of `20260414000001_create_game_metas_and_game_items_tables.go` — stop and investigate before continuing.

- [ ] **Step 3: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-api/tests/feature/course_content_dedup_test.go && git commit -m "test(api): in-batch repetition creates one meta and two junctions"
```

---

## Task 8: Test — translation matching rules

**Files:**
- Modify: `dx-api/tests/feature/course_content_dedup_test.go`

- [ ] **Step 1: Add three translation-matching tests**

Append:

```go
// TestSave_NullEqualsEmpty verifies that NULL and "" translations are treated
// as equivalent for dedup purposes.
func (s *ContentDedupSuite) TestSave_NullEqualsEmpty() {
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)

	// Save with NULL translation
	_, err := api.SaveMetadataBatch(s.userID, gameID, levelID,
		[]api.MetadataEntry{{SourceData: "fish", Translation: nil, SourceType: "vocab"}},
		"manual")
	s.Require().NoError(err)
	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID))

	// Save with empty-string translation in another level
	level2 := s.seedLevel(gameID)
	emptyStr := ""
	_, err = api.SaveMetadataBatch(s.userID, gameID, level2,
		[]api.MetadataEntry{{SourceData: "fish", Translation: &emptyStr, SourceType: "vocab"}},
		"manual")
	s.Require().NoError(err)

	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID), "NULL and empty translation must dedup")
	s.Equal(int64(1), s.countGameMetasInLevel(levelID))
	s.Equal(int64(1), s.countGameMetasInLevel(level2))
}

// TestSave_DifferentTranslations_DoNotDedup verifies that the same source_data
// with different translations creates two separate content_metas rows.
func (s *ContentDedupSuite) TestSave_DifferentTranslations_DoNotDedup() {
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)

	_, err := api.SaveMetadataBatch(s.userID, gameID, levelID,
		[]api.MetadataEntry{{SourceData: "bank", Translation: strPtr("银行"), SourceType: "vocab"}},
		"manual")
	s.Require().NoError(err)

	level2 := s.seedLevel(gameID)
	_, err = api.SaveMetadataBatch(s.userID, gameID, level2,
		[]api.MetadataEntry{{SourceData: "bank", Translation: strPtr("河岸"), SourceType: "vocab"}},
		"manual")
	s.Require().NoError(err)

	s.Equal(int64(2), s.countMetasOwnedByUser(s.userID), "different translations are not deduped")
}

// TestSave_DifferentSourceTypes_DoNotDedup verifies that the same source_data
// with different source_types creates two separate content_metas rows.
func (s *ContentDedupSuite) TestSave_DifferentSourceTypes_DoNotDedup() {
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)

	_, err := api.SaveMetadataBatch(s.userID, gameID, levelID,
		[]api.MetadataEntry{
			{SourceData: "apple", Translation: strPtr("苹果"), SourceType: "vocab"},
			{SourceData: "apple", Translation: strPtr("苹果"), SourceType: "sentence"},
		},
		"manual")
	s.Require().NoError(err)

	s.Equal(int64(2), s.countMetasOwnedByUser(s.userID), "different source_types are not deduped")
	s.Equal(int64(2), s.countGameMetasInLevel(levelID))
}
```

- [ ] **Step 2: Run the three tests**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestSave_NullEqualsEmpty|TestContentDedupSuite/TestSave_DifferentTranslations_DoNotDedup|TestContentDedupSuite/TestSave_DifferentSourceTypes_DoNotDedup" ./tests/feature/ -v
```

Expected: all three PASS.

- [ ] **Step 3: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-api/tests/feature/course_content_dedup_test.go && git commit -m "test(api): translation NULL≡empty, different translation/source_type bypass dedup"
```

---

## Task 9: Test — cross-user isolation

**Files:**
- Modify: `dx-api/tests/feature/course_content_dedup_test.go`

- [ ] **Step 1: Add the test**

Append:

```go
// TestSave_CrossUserIsolation verifies that User B's identical content does NOT
// reuse User A's content_meta. Each user has a private dedup pool.
func (s *ContentDedupSuite) TestSave_CrossUserIsolation() {
	// User A saves "secret"
	gameA := s.seedGame(consts.GameModeWordSentence)
	levelA := s.seedLevel(gameA)
	_, err := api.SaveMetadataBatch(s.userID, gameA, levelA,
		[]api.MetadataEntry{{SourceData: "secret", Translation: strPtr("秘密"), SourceType: "vocab"}},
		"manual")
	s.Require().NoError(err)

	// User B saves the same thing
	userB := s.seedVipUser()
	defer func() {
		_, _ = facades.Orm().Query().Exec(`DELETE FROM game_items WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, userB)
		_, _ = facades.Orm().Query().Exec(`DELETE FROM game_metas WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, userB)
		_, _ = facades.Orm().Query().Exec(`DELETE FROM content_metas WHERE id IN (SELECT cm.id FROM content_metas cm JOIN game_metas gm ON gm.content_meta_id = cm.id JOIN games g ON g.id = gm.game_id WHERE g.user_id = ?)`, userB)
		_, _ = facades.Orm().Query().Exec(`DELETE FROM game_levels WHERE game_id IN (SELECT id FROM games WHERE user_id = ?)`, userB)
		_, _ = facades.Orm().Query().Exec(`DELETE FROM games WHERE user_id = ?`, userB)
		_, _ = facades.Orm().Query().Exec(`DELETE FROM users WHERE id = ?`, userB)
	}()

	gameB := uuid.Must(uuid.NewV7()).String()
	gB := models.Game{ID: gameB, Name: "B", UserID: &userB, Mode: consts.GameModeWordSentence, IsActive: true, Status: consts.GameStatusDraft}
	s.Require().NoError(facades.Orm().Query().Create(&gB))
	levelB := uuid.Must(uuid.NewV7()).String()
	lvB := models.GameLevel{ID: levelB, GameID: gameB, Name: "L1", IsActive: true, Order: 1000}
	s.Require().NoError(facades.Orm().Query().Create(&lvB))

	_, err = api.SaveMetadataBatch(userB, gameB, levelB,
		[]api.MetadataEntry{{SourceData: "secret", Translation: strPtr("秘密"), SourceType: "vocab"}},
		"manual")
	s.Require().NoError(err)

	// Each user should own exactly 1 content_meta — no leakage.
	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID))
	s.Equal(int64(1), s.countMetasOwnedByUser(userB))
}
```

- [ ] **Step 2: Run**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestSave_CrossUserIsolation" ./tests/feature/ -v
```

Expected: PASS. The dedup query joins through `games.user_id = ?`, so User B's query won't see User A's row.

- [ ] **Step 3: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-api/tests/feature/course_content_dedup_test.go && git commit -m "test(api): cross-user dedup isolation"
```

---

## Task 10: Test — items reuse on broken-down meta

**Files:**
- Modify: `dx-api/tests/feature/course_content_dedup_test.go`

- [ ] **Step 1: Add the test**

Append:

```go
// TestSave_ReuseBrokenDownMeta_CopiesItemsViaJunction verifies that reusing a
// meta with is_break_done=true creates parallel game_items rows in the new
// level pointing at the existing content_items (no row duplication).
func (s *ContentDedupSuite) TestSave_ReuseBrokenDownMeta_CopiesItemsViaJunction() {
	// Set up Game A with a broken-down meta
	gameA := s.seedGame(consts.GameModeWordSentence)
	levelA := s.seedLevel(gameA)
	_, err := api.SaveMetadataBatch(s.userID, gameA, levelA,
		[]api.MetadataEntry{{SourceData: "broken sentence", Translation: strPtr("已拆解"), SourceType: "sentence"}},
		"manual")
	s.Require().NoError(err)

	// Find the meta and mark it broken-down with two manual content_items.
	var metaID string
	row := struct{ ID string }{}
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT cm.id FROM content_metas cm
		   JOIN game_metas gm ON gm.content_meta_id = cm.id
		   JOIN game_levels gl ON gl.id = gm.game_level_id
		   JOIN games g ON g.id = gl.game_id
		  WHERE g.user_id = ? AND cm.source_data = 'broken sentence'`,
		s.userID,
	).Scan(&row))
	metaID = row.ID
	s.Require().NotEmpty(metaID)

	item1 := models.ContentItem{ID: uuid.Must(uuid.NewV7()).String(), ContentMetaID: &metaID, Content: "broken", ContentType: "word"}
	item2 := models.ContentItem{ID: uuid.Must(uuid.NewV7()).String(), ContentMetaID: &metaID, Content: "sentence", ContentType: "word"}
	s.Require().NoError(facades.Orm().Query().Create(&item1))
	s.Require().NoError(facades.Orm().Query().Create(&item2))

	// Link items to level A via game_items junction
	gi1 := models.GameItem{ID: uuid.Must(uuid.NewV7()).String(), GameID: gameA, GameLevelID: levelA, ContentItemID: item1.ID, Order: 1000}
	gi2 := models.GameItem{ID: uuid.Must(uuid.NewV7()).String(), GameID: gameA, GameLevelID: levelA, ContentItemID: item2.ID, Order: 2000}
	s.Require().NoError(facades.Orm().Query().Create(&gi1))
	s.Require().NoError(facades.Orm().Query().Create(&gi2))

	// Mark meta as broken-down
	_, err = facades.Orm().Query().Exec(`UPDATE content_metas SET is_break_done = true WHERE id = ?`, metaID)
	s.Require().NoError(err)

	// Now reuse the meta in a new level via dedup
	gameB := s.seedGame(consts.GameModeWordSentence)
	levelB := s.seedLevel(gameB)
	_, err = api.SaveMetadataBatch(s.userID, gameB, levelB,
		[]api.MetadataEntry{{SourceData: "broken sentence", Translation: strPtr("已拆解"), SourceType: "sentence"}},
		"manual")
	s.Require().NoError(err)

	// content_items table should still have only 2 rows for this meta (no duplication)
	var itemCount int64
	itemCount, err = facades.Orm().Query().Model(&models.ContentItem{}).Where("content_meta_id", metaID).Count()
	s.Require().NoError(err)
	s.Equal(int64(2), itemCount, "content_items should not be duplicated")

	// Level B should now have 2 game_items pointing at the existing content_items
	s.Equal(int64(2), s.countGameItemsInLevel(levelB))

	// Both should reference the original IDs
	var itemIDs []string
	itemIDsRows := []struct{ ContentItemID string }{}
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT content_item_id FROM game_items WHERE game_level_id = ? ORDER BY "order"`,
		levelB,
	).Scan(&itemIDsRows))
	for _, r := range itemIDsRows {
		itemIDs = append(itemIDs, r.ContentItemID)
	}
	s.ElementsMatch([]string{item1.ID, item2.ID}, itemIDs)
}
```

- [ ] **Step 2: Run**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestSave_ReuseBrokenDownMeta_CopiesItemsViaJunction" ./tests/feature/ -v
```

Expected: PASS. The Task 6 implementation already calls `reuseItemsIntoLevel` when the reused meta has `IsBreakDone = true`.

- [ ] **Step 3: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-api/tests/feature/course_content_dedup_test.go && git commit -m "test(api): reuse broken-down meta links existing items via game_items"
```

---

## Task 11: Test — capacity check still enforced with dedup

**Files:**
- Modify: `dx-api/tests/feature/course_content_dedup_test.go`

- [ ] **Step 1: Add the test**

Append:

```go
// TestSave_CapacityCountsDedupedEntries verifies that deduped entries STILL
// count toward the level's capacity limit (capacity reflects displayed items,
// not unique underlying rows).
func (s *ContentDedupSuite) TestSave_CapacityCountsDedupedEntries() {
	gameID := s.seedGame(consts.GameModeVocabEn)
	levelID := s.seedLevel(gameID)

	// Pre-fill the level to one short of MaxMetasPerLevel
	preEntries := make([]api.MetadataEntry, api.MaxMetasPerLevel-1)
	for i := range preEntries {
		preEntries[i] = api.MetadataEntry{
			SourceData:  uuid.Must(uuid.NewV7()).String(),
			Translation: strPtr("t"),
			SourceType:  "vocab",
		}
	}
	_, err := api.SaveMetadataBatch(s.userID, gameID, levelID, preEntries, "manual")
	s.Require().NoError(err)

	// Now try to add 2 entries that would dedup to 0 new content_metas — but
	// they're 2 displayable items, which pushes the level over capacity.
	entries := []api.MetadataEntry{
		preEntries[0], // dedups
		preEntries[1], // dedups
	}
	_, err = api.SaveMetadataBatch(s.userID, gameID, levelID, entries, "manual")
	s.Error(err, "capacity check must consider total entries, not just net new metas")
}
```

- [ ] **Step 2: Verify the constant name and game mode constant**

```bash
grep -n "MaxMetasPerLevel\|GameModeVocab" /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/services/api/*.go /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/consts/*.go
```

Expected: confirms `api.MaxMetasPerLevel` and `consts.GameModeVocabEn` (or similar). Adjust the test if the actual names differ. If `IsVocabMode` requires a specific mode like `vocab-en`, use that.

- [ ] **Step 3: Run**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestSave_CapacityCountsDedupedEntries" ./tests/feature/ -v
```

Expected: PASS. The capacity check uses `len(existing) + len(entries)`, which is unchanged by dedup.

- [ ] **Step 4: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-api/tests/feature/course_content_dedup_test.go && git commit -m "test(api): capacity check counts deduped entries toward limit"
```

---

## Task 12: Update DeleteMetadata service signature & logic (RED → GREEN)

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go`
- Modify: `dx-api/tests/feature/course_content_dedup_test.go`

- [ ] **Step 1: Add a failing test for reference-counted DeleteMetadata**

Append to `course_content_dedup_test.go`:

```go
// TestDeleteMetadata_PreservesSharedMeta verifies that deleting a meta from
// one level leaves a shared meta intact in another level.
func (s *ContentDedupSuite) TestDeleteMetadata_PreservesSharedMeta() {
	gameA := s.seedGame(consts.GameModeWordSentence)
	levelA := s.seedLevel(gameA)
	gameB := s.seedGame(consts.GameModeWordSentence)
	levelB := s.seedLevel(gameB)

	entries := []api.MetadataEntry{
		{SourceData: "shared", Translation: strPtr("共享"), SourceType: "vocab"},
	}
	_, err := api.SaveMetadataBatch(s.userID, gameA, levelA, entries, "manual")
	s.Require().NoError(err)
	_, err = api.SaveMetadataBatch(s.userID, gameB, levelB, entries, "manual")
	s.Require().NoError(err)

	// Find the shared meta ID
	var metaID string
	row := struct{ ID string }{}
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT id FROM content_metas WHERE source_data = 'shared' AND deleted_at IS NULL LIMIT 1`,
	).Scan(&row))
	metaID = row.ID

	// Delete it from level A
	s.Require().NoError(api.DeleteMetadata(s.userID, gameA, levelA, metaID))

	// content_metas row should still exist (level B references it)
	s.Equal(int64(1), s.countMetasOwnedByUser(s.userID))
	// Level A junction is gone
	s.Equal(int64(0), s.countGameMetasInLevel(levelA))
	// Level B junction still present
	s.Equal(int64(1), s.countGameMetasInLevel(levelB))
}

// TestDeleteMetadata_LastReferenceSoftDeletesUnderlying verifies that deleting
// the only reference to a meta DOES soft-delete the underlying content_metas.
func (s *ContentDedupSuite) TestDeleteMetadata_LastReferenceSoftDeletesUnderlying() {
	gameID := s.seedGame(consts.GameModeWordSentence)
	levelID := s.seedLevel(gameID)
	entries := []api.MetadataEntry{
		{SourceData: "lonely", Translation: strPtr("孤独"), SourceType: "vocab"},
	}
	_, err := api.SaveMetadataBatch(s.userID, gameID, levelID, entries, "manual")
	s.Require().NoError(err)

	var metaID string
	row := struct{ ID string }{}
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT id FROM content_metas WHERE source_data = 'lonely' AND deleted_at IS NULL LIMIT 1`,
	).Scan(&row))
	metaID = row.ID

	s.Require().NoError(api.DeleteMetadata(s.userID, gameID, levelID, metaID))

	s.Equal(int64(0), s.countMetasOwnedByUser(s.userID), "underlying should be soft-deleted")

	// Verify it's soft-deleted not hard-deleted
	var dt *string
	dtRow := struct {
		DeletedAt *string `gorm:"column:deleted_at"`
	}{}
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT deleted_at FROM content_metas WHERE id = ?`, metaID,
	).Scan(&dtRow))
	dt = dtRow.DeletedAt
	s.NotNil(dt, "soft delete sets deleted_at")
}
```

- [ ] **Step 2: Run — both tests should FAIL to compile**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestDeleteMetadata_PreservesSharedMeta" ./tests/feature/ -v 2>&1 | head -20
```

Expected: build error — `api.DeleteMetadata` is called with 4 args but current signature is 3. **This is the RED step.**

- [ ] **Step 3: Update `DeleteMetadata` signature and logic**

In `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/services/api/course_content_service.go`, find the `DeleteMetadata` function (around line 611). Replace its signature and entire body with:

```go
// DeleteMetadata removes a metadata entry from one level. With reuse enabled,
// only the level's junction row(s) are soft-deleted; the underlying
// content_metas / content_items rows are soft-deleted only when no other
// junction references them.
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
		// 1. Collect content_item_ids referenced by this level for this meta.
		var itemRows []struct {
			ContentItemID string `gorm:"column:content_item_id"`
		}
		if err := tx.Raw(
			`SELECT gi.content_item_id
			   FROM game_items gi
			   JOIN content_items ci ON ci.id = gi.content_item_id AND ci.deleted_at IS NULL
			  WHERE ci.content_meta_id = ?
			    AND gi.game_level_id = ?
			    AND gi.deleted_at IS NULL`,
			metaID, gameLevelID,
		).Scan(&itemRows); err != nil {
			return fmt.Errorf("failed to collect items for delete: %w", err)
		}

		// 2. Soft-delete the level-scoped game_items rows (all repetitions).
		if _, err := tx.Exec(
			`UPDATE game_items SET deleted_at = NOW()
			  WHERE game_level_id = ?
			    AND content_item_id IN (
			      SELECT id FROM content_items WHERE content_meta_id = ?
			    )
			    AND deleted_at IS NULL`,
			gameLevelID, metaID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete game_items: %w", err)
		}

		// 3. Soft-delete the level-scoped game_metas rows (all repetitions).
		if _, err := tx.Exec(
			`UPDATE game_metas SET deleted_at = NOW()
			  WHERE content_meta_id = ?
			    AND game_level_id = ?
			    AND deleted_at IS NULL`,
			metaID, gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete game_metas: %w", err)
		}

		// 4. For each orphaned content_item_id, count remaining live junctions
		//    across ALL levels; if 0, soft-delete the content_item.
		seen := map[string]struct{}{}
		for _, r := range itemRows {
			if _, ok := seen[r.ContentItemID]; ok {
				continue
			}
			seen[r.ContentItemID] = struct{}{}
			n, err := tx.Table("game_items").
				Where("content_item_id", r.ContentItemID).
				Where("deleted_at IS NULL").
				Count()
			if err != nil {
				return fmt.Errorf("failed to count game_items: %w", err)
			}
			if n == 0 {
				if _, err := tx.Exec(
					`UPDATE content_items SET deleted_at = NOW()
					  WHERE id = ? AND deleted_at IS NULL`,
					r.ContentItemID,
				); err != nil {
					return fmt.Errorf("failed to soft-delete content_item: %w", err)
				}
			}
		}

		// 5. Count remaining live game_metas for this content_meta across ALL levels.
		n, err := tx.Table("game_metas").
			Where("content_meta_id", metaID).
			Where("deleted_at IS NULL").
			Count()
		if err != nil {
			return fmt.Errorf("failed to count game_metas: %w", err)
		}
		if n == 0 {
			if _, err := tx.Exec(
				`UPDATE content_metas SET deleted_at = NOW()
				  WHERE id = ? AND deleted_at IS NULL`,
				metaID,
			); err != nil {
				return fmt.Errorf("failed to soft-delete content_meta: %w", err)
			}
		}
		return nil
	})
}
```

- [ ] **Step 4: Build to find broken callers**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./... 2>&1
```

Expected: compile errors at the controller call site (`DeleteMetadata` is now 4 args). We fix the controller in Task 13. For now, comment out the controller call site so the rest builds — OR just continue and fix it in Task 13. For TDD discipline, jump to Task 13 NOW, then come back to verify the test.

- [ ] **Step 5: Defer test verification until controller is updated**

(See Task 13.)

---

## Task 13: Update DeleteMetadata controller and route

**Files:**
- Modify: `dx-api/app/http/controllers/api/course_game_controller.go`
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Read the current controller**

```bash
sed -n '301,320p' /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/http/controllers/api/course_game_controller.go
```

Expected: shows `DeleteMetadata` controller method.

- [ ] **Step 2: Update the controller method**

In `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/http/controllers/api/course_game_controller.go`, replace the `DeleteMetadata` controller method body. The user fetch line and error response stay the same — just plumb a new `levelID` from the route param and pass it as the new third argument:

```go
// DeleteMetadata removes a single metadata entry and its content items.
func (c *CourseGameController) DeleteMetadata(ctx contractshttp.Context) contractshttp.Response {
	userID := getUserIDFromCtx(ctx)
	if userID == "" {
		return helpers.Unauthorized(ctx, "未登录")
	}
	gameID := ctx.Request().Route("id")
	levelID := ctx.Request().Route("levelId")
	metaID := ctx.Request().Route("metaId")

	if err := services.DeleteMetadata(userID, gameID, levelID, metaID); err != nil {
		return helpers.Error(ctx, err.Error())
	}
	return helpers.Ok(ctx, nil)
}
```

(If your project uses different helpers like `helpers.Unauthorized`, `helpers.Ok`, `helpers.Error`, match the pattern of the existing function. Read 5-10 lines of context above and below the function to confirm.)

- [ ] **Step 3: Update the route**

In `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/routes/api.go` find the line:

```go
cg.Delete("/{id}/metadata/{metaId}", courseGameController.DeleteMetadata)
```

Change it to:

```go
cg.Delete("/{id}/levels/{levelId}/metadata/{metaId}", courseGameController.DeleteMetadata)
```

- [ ] **Step 4: Build**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```

Expected: build succeeds.

- [ ] **Step 5: Run the DeleteMetadata tests added in Task 12**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestDeleteMetadata_PreservesSharedMeta|TestContentDedupSuite/TestDeleteMetadata_LastReferenceSoftDeletesUnderlying" ./tests/feature/ -v
```

Expected: both PASS.

- [ ] **Step 6: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-api/app/services/api/course_content_service.go dx-api/app/http/controllers/api/course_game_controller.go dx-api/routes/api.go dx-api/tests/feature/course_content_dedup_test.go && git commit -m "feat(api): reference-counted DeleteMetadata with levelId in path"
```

---

## Task 14: Update DeleteContentItem service + controller + route

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go`
- Modify: `dx-api/app/http/controllers/api/course_game_controller.go`
- Modify: `dx-api/routes/api.go`
- Modify: `dx-api/tests/feature/course_content_dedup_test.go`

- [ ] **Step 1: Add a failing test for reference-counted DeleteContentItem**

Append to `course_content_dedup_test.go`:

```go
// TestDeleteContentItem_PreservesSharedItem verifies that deleting an item
// from one level leaves the underlying content_item alive when another level
// still references it.
func (s *ContentDedupSuite) TestDeleteContentItem_PreservesSharedItem() {
	// Set up a meta with one item, then reuse it across two levels
	gameA := s.seedGame(consts.GameModeWordSentence)
	levelA := s.seedLevel(gameA)
	_, err := api.SaveMetadataBatch(s.userID, gameA, levelA,
		[]api.MetadataEntry{{SourceData: "shareit", Translation: strPtr("共享"), SourceType: "sentence"}},
		"manual")
	s.Require().NoError(err)

	var metaID string
	row := struct{ ID string }{}
	s.Require().NoError(facades.Orm().Query().Raw(
		`SELECT id FROM content_metas WHERE source_data = 'shareit'`,
	).Scan(&row))
	metaID = row.ID

	// Add an item to the meta in level A
	itemID := uuid.Must(uuid.NewV7()).String()
	item := models.ContentItem{ID: itemID, ContentMetaID: &metaID, Content: "share", ContentType: "word"}
	s.Require().NoError(facades.Orm().Query().Create(&item))
	gi := models.GameItem{ID: uuid.Must(uuid.NewV7()).String(), GameID: gameA, GameLevelID: levelA, ContentItemID: itemID, Order: 1000}
	s.Require().NoError(facades.Orm().Query().Create(&gi))
	_, err = facades.Orm().Query().Exec(`UPDATE content_metas SET is_break_done = true WHERE id = ?`, metaID)
	s.Require().NoError(err)

	// Reuse the meta into level B — items get linked too
	gameB := s.seedGame(consts.GameModeWordSentence)
	levelB := s.seedLevel(gameB)
	_, err = api.SaveMetadataBatch(s.userID, gameB, levelB,
		[]api.MetadataEntry{{SourceData: "shareit", Translation: strPtr("共享"), SourceType: "sentence"}},
		"manual")
	s.Require().NoError(err)

	// Delete the item from level B
	s.Require().NoError(api.DeleteContentItem(s.userID, gameB, levelB, itemID))

	// content_items row should still exist (level A still references it)
	var n int64
	n, err = facades.Orm().Query().Model(&models.ContentItem{}).Where("id", itemID).Count()
	s.Require().NoError(err)
	s.Equal(int64(1), n, "shared content_item should not be deleted")

	// Level B's junction is gone
	s.Equal(int64(0), s.countGameItemsInLevel(levelB))
	// Level A's junction is intact
	s.Equal(int64(1), s.countGameItemsInLevel(levelA))
}
```

- [ ] **Step 2: Run — should FAIL to compile**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestDeleteContentItem_PreservesSharedItem" ./tests/feature/ -v 2>&1 | head -10
```

Expected: build error — `api.DeleteContentItem` called with 4 args. **RED step.**

- [ ] **Step 3: Update `DeleteContentItem` in the service**

In `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/services/api/course_content_service.go`, replace the entire `DeleteContentItem` function (around line 467) with:

```go
// DeleteContentItem removes a content item from one level. With reuse enabled,
// only the level's game_items junction row is soft-deleted; the underlying
// content_item is soft-deleted only when no other junction references it.
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

	// Load the underlying content_item up front so we know its content_meta_id
	// for the is_break_done reset below.
	var item models.ContentItem
	if err := facades.Orm().Query().Where("id", itemID).First(&item); err != nil {
		return fmt.Errorf("failed to load content item: %w", err)
	}
	if item.ID == "" {
		return ErrContentItemNotFound
	}

	return facades.Orm().Transaction(func(tx orm.Query) error {
		// 1. Soft-delete this level's game_items rows for this item (all repetitions).
		if _, err := tx.Exec(
			`UPDATE game_items SET deleted_at = NOW()
			  WHERE content_item_id = ?
			    AND game_level_id = ?
			    AND deleted_at IS NULL`,
			itemID, gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete game_item: %w", err)
		}

		// 2. Count live game_items across all levels for this content_item.
		n, err := tx.Table("game_items").
			Where("content_item_id", itemID).
			Where("deleted_at IS NULL").
			Count()
		if err != nil {
			return fmt.Errorf("failed to count game_items: %w", err)
		}
		if n == 0 {
			if _, err := tx.Exec(
				`UPDATE content_items SET deleted_at = NOW()
				  WHERE id = ? AND deleted_at IS NULL`,
				itemID,
			); err != nil {
				return fmt.Errorf("failed to soft-delete content_item: %w", err)
			}
		}

		// 3. Reset is_break_done if this LEVEL has no remaining game_items
		//    for the meta. (Existing per-level logic, unchanged.)
		if item.ContentMetaID != nil {
			if _, err := tx.Exec(
				`UPDATE content_metas SET is_break_done = false
				  WHERE id = ?
				    AND deleted_at IS NULL
				    AND NOT EXISTS (
				      SELECT 1 FROM game_items gi
				      JOIN content_items ci ON ci.id = gi.content_item_id AND ci.deleted_at IS NULL
				      WHERE ci.content_meta_id = content_metas.id
				        AND gi.game_level_id = ?
				        AND gi.deleted_at IS NULL
				    )`,
				*item.ContentMetaID, gameLevelID,
			); err != nil {
				return fmt.Errorf("failed to reset meta break status: %w", err)
			}
		}
		return nil
	})
}
```

- [ ] **Step 4: Update the controller**

In `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/http/controllers/api/course_game_controller.go`, find `DeleteContentItem` controller method (around line 418). Update its body to fetch `levelId` from the route and pass it through:

```go
// DeleteContentItem removes a single content item.
func (c *CourseGameController) DeleteContentItem(ctx contractshttp.Context) contractshttp.Response {
	userID := getUserIDFromCtx(ctx)
	if userID == "" {
		return helpers.Unauthorized(ctx, "未登录")
	}
	gameID := ctx.Request().Route("id")
	levelID := ctx.Request().Route("levelId")
	itemID := ctx.Request().Route("itemId")

	if err := services.DeleteContentItem(userID, gameID, levelID, itemID); err != nil {
		return helpers.Error(ctx, err.Error())
	}
	return helpers.Ok(ctx, nil)
}
```

(Match the actual helper function names from your codebase.)

- [ ] **Step 5: Update the route**

In `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/routes/api.go`, find:

```go
cg.Delete("/{id}/content-items/{itemId}", courseGameController.DeleteContentItem)
```

Change to:

```go
cg.Delete("/{id}/levels/{levelId}/content-items/{itemId}", courseGameController.DeleteContentItem)
```

- [ ] **Step 6: Build and run the test**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./... && go test -race -run "TestContentDedupSuite/TestDeleteContentItem_PreservesSharedItem" ./tests/feature/ -v
```

Expected: build succeeds; test PASSES.

- [ ] **Step 7: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-api/app/services/api/course_content_service.go dx-api/app/http/controllers/api/course_game_controller.go dx-api/routes/api.go dx-api/tests/feature/course_content_dedup_test.go && git commit -m "feat(api): reference-counted DeleteContentItem with levelId in path"
```

---

## Task 15: Update DeleteAllLevelContent (signature unchanged, logic only)

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go`
- Modify: `dx-api/tests/feature/course_content_dedup_test.go`

- [ ] **Step 1: Add a failing test**

Append:

```go
// TestDeleteAllLevelContent_PreservesSharedContent verifies that bulk
// deletion of one level's content does not delete content shared with another level.
func (s *ContentDedupSuite) TestDeleteAllLevelContent_PreservesSharedContent() {
	gameA := s.seedGame(consts.GameModeWordSentence)
	levelA := s.seedLevel(gameA)
	gameB := s.seedGame(consts.GameModeWordSentence)
	levelB := s.seedLevel(gameB)

	entries := []api.MetadataEntry{
		{SourceData: "first", Translation: strPtr("一"), SourceType: "vocab"},
		{SourceData: "second", Translation: strPtr("二"), SourceType: "vocab"},
	}
	_, err := api.SaveMetadataBatch(s.userID, gameA, levelA, entries, "manual")
	s.Require().NoError(err)
	_, err = api.SaveMetadataBatch(s.userID, gameB, levelB, entries, "manual")
	s.Require().NoError(err)

	// Two metas exist, each referenced by both levels
	s.Equal(int64(2), s.countMetasOwnedByUser(s.userID))

	// Bulk delete everything in level A
	s.Require().NoError(api.DeleteAllLevelContent(s.userID, gameA, levelA))

	// content_metas still exist (level B references them)
	s.Equal(int64(2), s.countMetasOwnedByUser(s.userID), "shared content survives bulk delete")
	s.Equal(int64(0), s.countGameMetasInLevel(levelA))
	s.Equal(int64(2), s.countGameMetasInLevel(levelB))
}
```

- [ ] **Step 2: Run — expected to FAIL**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite/TestDeleteAllLevelContent_PreservesSharedContent" ./tests/feature/ -v
```

Expected: FAIL — current code cascade-deletes the shared content_metas, so the count drops to 0. **RED step.**

- [ ] **Step 3: Replace `DeleteAllLevelContent` body**

In `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/services/api/course_content_service.go`, replace the entire `DeleteAllLevelContent` function (around line 546) with:

```go
// DeleteAllLevelContent removes all content from a level. With reuse enabled,
// the level's junction rows are soft-deleted unconditionally; underlying
// content_metas / content_items are soft-deleted only when no other junction
// references them.
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

	// Verify level belongs to game
	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).Where("game_id", gameID).First(&level); err != nil {
		return fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return ErrLevelNotFound
	}

	return facades.Orm().Transaction(func(tx orm.Query) error {
		// 1. Collect distinct content_meta_ids and content_item_ids referenced
		//    by this level BEFORE we soft-delete the junctions.
		var metaRows []struct {
			ID string `gorm:"column:content_meta_id"`
		}
		if err := tx.Raw(
			`SELECT DISTINCT content_meta_id FROM game_metas
			  WHERE game_level_id = ? AND deleted_at IS NULL`,
			gameLevelID,
		).Scan(&metaRows); err != nil {
			return fmt.Errorf("failed to collect metas: %w", err)
		}

		var itemRows []struct {
			ID string `gorm:"column:content_item_id"`
		}
		if err := tx.Raw(
			`SELECT DISTINCT content_item_id FROM game_items
			  WHERE game_level_id = ? AND deleted_at IS NULL`,
			gameLevelID,
		).Scan(&itemRows); err != nil {
			return fmt.Errorf("failed to collect items: %w", err)
		}

		// 2. Soft-delete junctions for this level.
		if _, err := tx.Exec(
			`UPDATE game_items SET deleted_at = NOW()
			  WHERE game_level_id = ? AND deleted_at IS NULL`,
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete game_items: %w", err)
		}
		if _, err := tx.Exec(
			`UPDATE game_metas SET deleted_at = NOW()
			  WHERE game_level_id = ? AND deleted_at IS NULL`,
			gameLevelID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete game_metas: %w", err)
		}

		// 3. For each collected content_item_id, count remaining live game_items;
		//    if 0, soft-delete the content_item.
		for _, r := range itemRows {
			n, err := tx.Table("game_items").
				Where("content_item_id", r.ID).
				Where("deleted_at IS NULL").
				Count()
			if err != nil {
				return fmt.Errorf("failed to count game_items: %w", err)
			}
			if n == 0 {
				if _, err := tx.Exec(
					`UPDATE content_items SET deleted_at = NOW()
					  WHERE id = ? AND deleted_at IS NULL`,
					r.ID,
				); err != nil {
					return fmt.Errorf("failed to soft-delete content_item: %w", err)
				}
			}
		}

		// 4. For each collected content_meta_id, count remaining live game_metas;
		//    if 0, soft-delete the content_meta.
		for _, r := range metaRows {
			n, err := tx.Table("game_metas").
				Where("content_meta_id", r.ID).
				Where("deleted_at IS NULL").
				Count()
			if err != nil {
				return fmt.Errorf("failed to count game_metas: %w", err)
			}
			if n == 0 {
				if _, err := tx.Exec(
					`UPDATE content_metas SET deleted_at = NOW()
					  WHERE id = ? AND deleted_at IS NULL`,
					r.ID,
				); err != nil {
					return fmt.Errorf("failed to soft-delete content_meta: %w", err)
				}
			}
		}
		return nil
	})
}
```

- [ ] **Step 4: Run all dedup-related tests**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race -run "TestContentDedupSuite" ./tests/feature/ -v
```

Expected: ALL tests in the suite PASS. The new test from Step 1 should now pass; existing tests should still pass.

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-api/app/services/api/course_content_service.go dx-api/tests/feature/course_content_dedup_test.go && git commit -m "feat(api): reference-counted DeleteAllLevelContent"
```

---

## Task 16: Frontend — update delete actions and call sites

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/actions/course-game.action.ts`
- Modify: `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx`

- [ ] **Step 1: Update `deleteMetaAction` signature and URL**

In `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-web/src/features/web/ai-custom/actions/course-game.action.ts`, find the `deleteMetaAction` function (around line 318) and replace it with:

```typescript
/** Delete a single metadata entry via Go API. */
export async function deleteMetaAction(
  gameId: string,
  levelId: string,
  metaId: string
): Promise<SimpleActionResult> {
  try {
    const res = await apiClient.delete<null>(
      `/api/course-games/${gameId}/levels/${levelId}/metadata/${metaId}`
    );
    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "删除元数据失败" };
  }
}
```

- [ ] **Step 2: Update `deleteContentItemAction` signature and URL**

In the same file, find `deleteContentItemAction` (around line 334) and replace with:

```typescript
/** Delete a content item via Go API. */
export async function deleteContentItemAction(
  gameId: string,
  levelId: string,
  itemId: string
): Promise<SimpleActionResult> {
  try {
    const res = await apiClient.delete<null>(
      `/api/course-games/${gameId}/levels/${levelId}/content-items/${itemId}`
    );
    if (res.code !== 0) return { error: res.message };
    return { success: true };
  } catch {
    return { error: "删除失败" };
  }
}
```

- [ ] **Step 3: Update the call sites in `level-units-panel.tsx`**

In `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-web/src/features/web/ai-custom/components/level-units-panel.tsx`:

Find this line (around line 494):

```typescript
    const result = await deleteMetaAction(gameId, pendingDeleteMetaId);
```

Replace with:

```typescript
    const result = await deleteMetaAction(gameId, levelId, pendingDeleteMetaId);
```

Find this line (around line 532):

```typescript
    const result = await deleteContentItemAction(gameId, pendingDeleteItemId);
```

Replace with:

```typescript
    const result = await deleteContentItemAction(gameId, levelId, pendingDeleteItemId);
```

- [ ] **Step 4: Update the dependency arrays of the surrounding `useCallback`**

In the same file, find the `handleConfirmDeleteMeta` callback's dependency array (the line ending with `, [pendingDeleteMetaId, metas, selectedId, gameId]);`). Add `levelId` so the callback re-binds when the level changes:

```typescript
  }, [pendingDeleteMetaId, metas, selectedId, gameId, levelId]);
```

Find the `handleConfirmDeleteItem` callback's dependency array (`, [pendingDeleteItemId, contentItems, metas, selectedId, gameId]);`) and update similarly:

```typescript
  }, [pendingDeleteItemId, contentItems, metas, selectedId, gameId, levelId]);
```

- [ ] **Step 5: Verify `levelId` is in scope as a prop**

```bash
grep -n "levelId" /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web/src/features/web/ai-custom/components/level-units-panel.tsx | head -10
```

Expected: shows `levelId` in the component's props and other usages — confirming it's already in scope where we use it.

- [ ] **Step 6: Run frontend lint and type check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run lint
```

Expected: no lint errors. If there's a `npm run typecheck` script, run it too:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run typecheck 2>/dev/null || npx tsc --noEmit
```

Expected: no type errors.

- [ ] **Step 7: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git add dx-web/src/features/web/ai-custom/actions/course-game.action.ts dx-web/src/features/web/ai-custom/components/level-units-panel.tsx && git commit -m "feat(web): delete actions accept levelId for new path-based delete routes"
```

---

## Task 17: Final verification — full test suite + manual smoke

**Files:** none (verification only)

- [ ] **Step 1: Run the full Go test suite with race detector**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./...
```

Expected: ALL tests PASS, including all `TestContentDedupSuite` tests and the existing example test. No race warnings.

- [ ] **Step 2: Run `go vet`**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./...
```

Expected: no warnings.

- [ ] **Step 3: Run staticcheck (if installed)**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && staticcheck ./... 2>/dev/null || echo "staticcheck not installed, skipping"
```

Expected: no warnings (or skipped if not installed).

- [ ] **Step 4: Run frontend lint**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run lint
```

Expected: no errors.

- [ ] **Step 5: Manual smoke test — start both servers**

In one terminal:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go run .
```
In another:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run dev
```

- [ ] **Step 6: Manual smoke test — exercise the dedup flow**

In a browser:
1. Sign in as a VIP user
2. Open an existing AI-custom game level (or create a new one)
3. Click "添加" → 添加词汇 panel
4. Add some metadata entries
5. Save successfully — verify entries appear
6. Create a SECOND game, open a level, add the SAME metadata entries
7. Save — should succeed without error
8. In a database client, verify `content_metas` row count grew by 0 for the duplicated entries:
   ```sql
   SELECT COUNT(*) FROM content_metas WHERE source_data = 'your-test-string' AND deleted_at IS NULL;
   ```
   Expected: 1 (not 2).
9. Delete the meta from the second game's level
10. Verify the FIRST game's level still shows the meta (data not deleted from under it)

- [ ] **Step 7: Update task tracking and report**

If any step in this task fails: STOP, debug, fix, and re-run. Do NOT mark the plan complete until every test passes and the smoke test confirms end-to-end behavior.

If everything passes, the implementation is complete. Notify the user with a summary of what changed and the verification results.

---

## Self-review notes

**Spec coverage check:**

| Spec section | Implemented in |
|---|---|
| Schema: relax junction unique indexes to non-unique | **ALREADY DONE** (merged into the create migration — out of scope for this resumption) |
| Schema: add `idx_content_metas_dedup_lookup` on `content_metas` | Task 2 (narrowed) |
| Save dedup logic, helpers, transaction | Tasks 4-11 (tests) + Task 6 (impl) |
| Within-batch dedup | Task 7 |
| Items reuse on broken-down meta | Task 10 (test) + Task 6 (impl, via `reuseItemsIntoLevel`) |
| Cross-user isolation | Task 9 |
| `DeleteMetadata` reference counting + `levelId` in path | Tasks 12-13 |
| `DeleteContentItem` reference counting + `levelId` in path | Task 14 |
| `DeleteAllLevelContent` reference counting | Task 15 |
| Frontend: delete actions accept `levelId` | Task 16 |
| Pre-migration backup | Task 1 |
| Final verification (lint, race, smoke) | Task 17 |

**Function name consistency:**
- `findExistingMetasForBatch` — defined in Task 6, no other reference
- `reuseItemsIntoLevel` — defined in Task 6, called from `SaveMetadataBatch`
- `metaDedupKey` / `existingMetaRef` — types defined and used consistently in Task 6
- `DeleteMetadata(userID, gameID, gameLevelID, metaID)` — new 4-arg signature defined in Task 12, called from Task 13 controller, tested in Task 12
- `DeleteContentItem(userID, gameID, gameLevelID, itemID)` — new 4-arg signature defined in Task 14, called from Task 14 controller, tested in Task 14

**Placeholder scan:** none found. Each step contains the literal code or command needed.

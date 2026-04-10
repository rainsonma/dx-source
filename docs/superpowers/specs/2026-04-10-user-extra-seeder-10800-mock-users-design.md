# User Extra Seeder — Append 10,800 Mock Users

**Date:** 2026-04-10
**Status:** Design approved, pending implementation
**Scope:** dx-api (backend only; dx-web and deploy untouched)
**Lifecycle:** One-shot throwaway seeder — deleted after successful run

## Problem

The `users` table currently holds 1,202 rows: 2 named users (`rainson`, `june`) and 1,200 mock users seeded by `dx-api/database/seeders/user_seeder.go`. We need to grow the mock user population to 12,000 (append 10,800 new rows) to better exercise list pagination, leaderboards, PK matching, and partition-ready code paths — without touching any of the existing 1,202 rows.

Constraints:

- **Zero changes** to the existing 1,202 users. No field writes, no ID reassignment, no re-randomization.
- **Same username / nickname strategy** as `user_seeder.go`: English username from the `firstNames` pool (single or paired), English nickname from `nickFirsts`, and Chinese nicknames built from `cnBases` × 7 `cnPatterns`.
- **5,000 of the 10,800 new users** must have Chinese nicknames, **randomly interleaved** with the 5,800 English-nickname users (not grouped in a contiguous block).
- Must produce no lint issues (`gofmt`, `go vet`, `go build`).
- Must not break any existing function. In particular:
  - `bootstrap/seeders.go` remains compilable and preserves `DatabaseSeeder` registration.
  - `user_seeder.go`, `user_bean_seeder.go`, and all other seeders stay byte-identical.
  - `app/console/commands/import_courses.go`'s `loadUserIDs(1202)` still resolves the original 1,202 users (UUIDv7 preserves insertion order via `ORDER BY created_at ASC`).
- The project uses **code-level FK constraints only** (no DB-level foreign keys) to support PostgreSQL partitioning via `pg_partman`. New user inserts therefore face no cascading referential checks — only the `UNIQUE` constraints on `username`, `email`, `phone`, and `invite_code`. Since mock users leave `email` and `phone` NULL (and Postgres `UNIQUE NULLS DISTINCT` permits multiple NULLs), only `username` and `invite_code` need collision handling.

## Non-Goals

- No changes to `user_seeder.go`, `database_seeder.go`, migrations, the `User` model, or any other seeder.
- No dx-web or deploy/docker-compose changes.
- No test file for the new seeder. The codebase has zero tests for any existing seeder; adding one here contradicts the established convention and the "no bloat" rule in `CLAUDE.md`. Verification is done manually via SQL queries after the run.
- No shared helper extraction between `user_seeder.go` and the new file. The new file duplicates the name pools on purpose because it is throwaway — coupling it to `user_seeder.go` would leave dead code behind after removal.
- No support for configurable counts or a `--count` flag. The numbers (10,800 new users, 5,000 Chinese nicknames, 500-row chunk size) are hard-coded constants.
- No determinism. The seeder inherits Go 1.26's auto-seeded global `math/rand`, same as `user_seeder.go`. Each run produces a different valid set; the seeder is designed to be run exactly once.

## Architecture

### File changes

| File | Change | Lines |
|---|---|---|
| `dx-api/database/seeders/user_extra_seeder.go` | **Added** (new file) | ~300 |
| `dx-api/bootstrap/seeders.go` | **Modified** (one added line) | +1 |

All other files are untouched.

### Seeder file layout — `user_extra_seeder.go`

```
package seeders

imports
    fmt, log, math/rand
    dx-api/app/helpers
    dx-api/app/models
    github.com/google/uuid
    github.com/goravel/framework/contracts/database/orm
    github.com/goravel/framework/facades

types
    UserExtraSeeder      struct{}
    extraMockUser        struct{ Username, Nickname string }
        // distinct name from user_seeder.go's private `mockUser` type
        // to avoid a package-level type collision

methods
    Signature() → "UserExtraSeeder"
    Run() error
        Phase 1 — hash password once
        Phase 2 — load existing usernames into `seen` map
        Phase 3 — buildExtraMockUsers(seen) → 10,800 users
        Phase 4 — assemble []models.User structs
        Phase 5 — chunked batch insert inside Transaction

private helpers
    buildExtraMockUsers(seen map[string]bool) []extraMockUser

private data
    firstNames   []string   (copied verbatim from user_seeder.go)
    nickFirsts   []string   (copied verbatim)
    cnBases      []string   (copied verbatim)
    cnPrefixes   []string   (copied verbatim)
    cnSuffixes   []string   (copied verbatim)
    cnNums       []string   (copied verbatim)
    cnPatterns   []func(string) string   (copied verbatim — 7 patterns)
    grades       []string   = {"month","season","year","lifetime"}
```

### Bootstrap registration — `bootstrap/seeders.go`

```go
func Seeders() []seeder.Seeder {
    return []seeder.Seeder{
        &seeders.DatabaseSeeder{},
        &seeders.UserExtraSeeder{},   // ← added
    }
}
```

This registration makes `--seeder=UserExtraSeeder` resolvable via Goravel's `SeederFacade.GetSeeder(name)` (which only scans the top-level registered list). `DatabaseSeeder` remains untouched.

**Side effect acknowledged**: plain `go run . artisan db:seed` (no flag) will now run **both** `DatabaseSeeder` and `UserExtraSeeder`. This is acceptable because:

1. The user's workflow is `db:seed --seeder=UserExtraSeeder` — never plain `db:seed`.
2. The new seeder is removed after the single run, reverting `bootstrap/seeders.go` to its original state.

## Run() Data Flow

```
┌───────────────────────────────────────────────────────────────────┐
│  Phase 1 — Hash password once                                     │
│    mockPw, err := helpers.HashPassword("Mock!@#Pass")             │
│    if err != nil: return wrapped error                            │
│    cost: ~100ms (bcrypt cost 12, same as user_seeder.go)          │
├───────────────────────────────────────────────────────────────────┤
│  Phase 2 — Load existing usernames into seen map                  │
│    rows := []struct{ Username string }                            │
│    facades.Orm().Query().Model(&models.User{}).                   │
│        Select("username").Get(&rows)                              │
│    seen := make(map[string]bool, len(rows)+2)                     │
│    for _, r := range rows: seen[r.Username] = true                │
│    seen["rainson"] = true   // defensive                          │
│    seen["june"]    = true   // defensive                          │
│    cost: ~1,202-row single-column SELECT, ~10ms                   │
├───────────────────────────────────────────────────────────────────┤
│  Phase 3 — Generate 10,800 unique extraMockUsers                  │
│    generated := buildExtraMockUsers(seen)                         │
│    // Each user has a unique Username absent from `seen` and      │
│    // mutates `seen` as it adds entries. Returns slice of length  │
│    // exactly 10,800 — 5,000 with Chinese nicknames, 5,800 with   │
│    // English nicknames, randomly interleaved via rand.Perm.      │
│    cost: pure in-memory, <100ms                                   │
├───────────────────────────────────────────────────────────────────┤
│  Phase 4 — Build 10,800 models.User structs                       │
│    built := make([]models.User, 0, 10800)                         │
│    for _, m := range generated:                                   │
│        grade := grades[rand.Intn(4)]                              │
│        nickname := m.Nickname  // copy for &ref                   │
│        built = append(built, models.User{                         │
│            ID:         uuid.Must(uuid.NewV7()).String(),          │
│            Username:   m.Username,                                │
│            Nickname:   &nickname,                                 │
│            Grade:      grade,                                     │
│            Password:   mockPw,                                    │
│            InviteCode: helpers.GenerateInviteCode(8),             │
│            IsActive:   true,                                      │
│            IsMock:     true,                                      │
│        })                                                         │
│    cost: in-memory + 10,800 crypto/rand invite_code calls, ~500ms │
├───────────────────────────────────────────────────────────────────┤
│  Phase 5 — Chunked batch insert inside Transaction                │
│    const chunkSize = 500                                          │
│    err := facades.Orm().Transaction(func(tx orm.Query) error {    │
│        for i := 0; i < len(built); i += chunkSize {               │
│            end := min(i+chunkSize, len(built))                    │
│            chunk := built[i:end]                                  │
│            if err := tx.Create(&chunk); err != nil {              │
│                return fmt.Errorf(                                 │
│                    "failed to insert chunk %d (rows %d-%d): %w",  │
│                    i/chunkSize, i, end-1, err)                    │
│            }                                                      │
│        }                                                          │
│        return nil                                                 │
│    })                                                             │
│    cost: 22 INSERT statements (21 × 500 + 1 × 300), ~1-3s total   │
│    Parameter budget: 500 × ~23 fields ≈ 11,500 params per INSERT, │
│    well under PostgreSQL's 65,535-parameter hard limit.           │
│    (models.User has 21 explicit fields + 2 from orm.Timestamps.)  │
├───────────────────────────────────────────────────────────────────┤
│  Phase 6 — Log success                                            │
│    log.Printf("Seeded %d extra mock users "+                      │
│        "(%d Chinese + %d English)", 10800, 5000, 5800)            │
└───────────────────────────────────────────────────────────────────┘
```

## `buildExtraMockUsers` Generation Logic

```go
func buildExtraMockUsers(seen map[string]bool) []extraMockUser {
    const (
        targetCount  = 10800
        chineseCount = 5000
    )

    users := make([]extraMockUser, 0, targetCount)

    // Phase A — generate 10,800 unique usernames with English nicknames
    for len(users) < targetCount {
        username := firstNames[rand.Intn(len(firstNames))]
        if rand.Intn(2) == 0 {
            username += firstNames[rand.Intn(len(firstNames))]
        }
        if seen[username] {
            continue
        }
        seen[username] = true

        nickname := nickFirsts[rand.Intn(len(nickFirsts))]
        users = append(users, extraMockUser{
            Username: username,
            Nickname: nickname,
        })
    }

    // Phase B — overwrite 5,000 random indices with Chinese nicknames
    for _, idx := range rand.Perm(len(users))[:chineseCount] {
        base := cnBases[rand.Intn(len(cnBases))]
        users[idx].Nickname = cnPatterns[rand.Intn(len(cnPatterns))](base)
    }

    return users
}
```

### Differences from `user_seeder.go`'s `buildMockUsers()`

1. **`seen` is an input parameter**, not a local variable, so the caller can pre-populate it with existing DB usernames. Mutation of the caller's map is safe because the map is scratch space owned solely by `Run()`.
2. **`targetCount = 10800`** (was 1,200) and **`chineseCount = 5000`** (was 200).
3. **No explicit `"rainson"`/`"june"` guard** — unnecessary because those usernames are already in `seen` (both from Phase 2's DB SELECT and the defensive fallbacks).

### Name pool capacity check

- `firstNames` contains ~240 entries (copied from `user_seeder.go`).
- Possible username patterns: **single** (240) + **paired** (240²) = **~57,840 unique combinations**.
- Minus ~1,202 already in `seen` = **~56,640 free slots**.
- Target draw: 10,800 → a sampling rate of ~19%. The rejection loop (`if seen[username] { continue }`) will reject ~20-30% of draws toward the tail due to birthday-collision accumulation, but the loop remains bounded and terminates in well under a second. No risk of infinite loop.

## Error Handling

Per `/Users/rainsen/.claude/rules/coding-style.md` ("Always wrap errors with context"):

| Failure point | Wrapping | Rollback |
|---|---|---|
| `helpers.HashPassword(...)` | `fmt.Errorf("failed to hash mock password: %w", err)` | N/A — pre-Tx |
| `SELECT username FROM users` | `fmt.Errorf("failed to load existing usernames: %w", err)` | N/A — pre-Tx |
| `tx.Create(&chunk)` inside Tx | `fmt.Errorf("failed to insert chunk %d (rows %d-%d): %w", ...)` | Automatic: entire Tx rolls back |
| `facades.Orm().Transaction(...)` outer | `fmt.Errorf("failed to seed extra mock users: %w", err)` | Already rolled back |

The transaction wrapper guarantees **all-or-nothing**: if any chunk fails mid-flight (a freak `invite_code` collision, a network blip, a constraint violation), the previously inserted chunks in the same transaction roll back. The seeder either inserts exactly 10,800 rows or zero.

## Collision Risk Analysis

| Constraint | Source | Collision probability | Handling |
|---|---|---|---|
| `UNIQUE(username)` | User-generated from ~57,840 pool | ~0 after `seen` dedup | Pre-populated `seen` map from DB + in-batch `seen` mutation |
| `UNIQUE(invite_code)` | `crypto/rand` over 62⁸ space | ~2.7×10⁻⁷ for 10,800 draws | None — statistically negligible; Tx rollback as safety net if it ever hits |
| `UNIQUE(email)` | All NULL | 0 — `NULLS DISTINCT` | None |
| `UNIQUE(phone)` | All NULL | 0 — `NULLS DISTINCT` | None |

## Verification Plan

### Pre-run (before invoking the seeder)

```bash
cd dx-api
gofmt -l database/seeders/user_extra_seeder.go bootstrap/seeders.go
# Expected: empty output

go vet ./database/seeders/... ./bootstrap/...
# Expected: empty output

go build ./...
# Expected: successful build
```

### Invocation

```bash
cd dx-api
go run . artisan db:seed --seeder=UserExtraSeeder
# Expected log line:
#   Seeded 10800 extra mock users (5000 Chinese + 5800 English)
```

### Post-run (SQL verification against PostgreSQL)

```sql
-- 1. Total mock user count should be exactly 12,000
SELECT COUNT(*) FROM users WHERE is_mock = true;
-- Expected: 12000

-- 2. Total user count should be exactly 12,002 (2 named + 12,000 mock)
SELECT COUNT(*) FROM users;
-- Expected: 12002

-- 3. The 2 named users should be untouched
SELECT id, username, nickname, grade FROM users
 WHERE username IN ('rainson', 'june')
 ORDER BY username;
-- Expected: same 2 rows as before the run, IDs unchanged

-- 4. Chinese-nickname count across the whole mock population
--    (existing exactly 200 + new exactly 5,000 = exactly 5,200).
--    Detection uses octet_length != char_length, which catches any
--    multi-byte UTF-8 char. All cnPatterns produce at least one CJK
--    char, and English nicknames are pure ASCII, so this is precise.
SELECT COUNT(*) FROM users
 WHERE is_mock = true
   AND octet_length(nickname) <> char_length(nickname);
-- Expected: exactly 5,200

-- 5. No duplicate usernames
SELECT username, COUNT(*) FROM users
 GROUP BY username HAVING COUNT(*) > 1;
-- Expected: empty

-- 6. No duplicate invite codes
SELECT invite_code, COUNT(*) FROM users
 GROUP BY invite_code HAVING COUNT(*) > 1;
-- Expected: empty

-- 7. Chinese-nickname interleaving sanity check — sample 20 rows
--    ordered by created_at DESC (most recent 20 new users) and
--    confirm they're a mix of CJK and ASCII nicknames
SELECT username, nickname FROM users
 WHERE is_mock = true
 ORDER BY created_at DESC
 LIMIT 20;
-- Expected: a mix of Chinese and English nicknames, not all one kind
```

### Do-not-break verification

- `go build ./...` proves nothing else in the app is broken by the bootstrap change.
- `git diff --stat` should show **exactly two files touched**: the new `user_extra_seeder.go` and the one-line change in `bootstrap/seeders.go`.
- `import_courses.go:77`'s `loadUserIDs(1202)` uses `ORDER BY created_at ASC LIMIT 1202`. Since UUIDv7 embeds timestamp prefixes and the original 1,202 users were inserted before the new 10,800, the ORDER BY continues to resolve to the same original 1,202 rows. No code change required.

## Removal Plan

After successful verification, the seeder is removed with a two-step revert:

1. Delete `dx-api/database/seeders/user_extra_seeder.go`.
2. Revert `dx-api/bootstrap/seeders.go` — remove the `&seeders.UserExtraSeeder{},` line.

No migration, no data cleanup, no other touch-points. The 10,800 new rows persist in the database; only the code that created them is removed.

## Out of Scope — Risks Deferred to Manual Ops

- **Re-run safety**: if the seeder is re-invoked after a successful first run, it will **successfully add another 10,800 users** — not fail. The `seen` map (now containing all 12,002 usernames) just pushes the generator toward the free slots in the ~57k-combination name pool (~45,838 remaining), which is more than enough for another 10,800 draws. So re-running is **not** idempotent and doubles the mock-user count. The primary safeguard is operational: "don't re-run it." The seeder is documented as one-shot, you delete it after the first successful run, and re-run protection is explicitly out of scope.
- **Production safety**: `db:seed` in production requires `--force` per Goravel's convention. This seeder inherits that protection via the framework's `ConfirmToProceed` guard. No additional prod-specific logic added.
- **Metrics / analytics impact**: adding 10,800 mock users may shift dashboard counts, leaderboards, and average-statistic computations in dev. This is expected and desired.

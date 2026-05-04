# Content / Vocab Schema Refactor — Design

Date: 2026-05-03
Status: **Partially superseded by `2026-05-04-vocab-pool-pivot-design.md`** — see notice below

> ## ⚠️ Partially Superseded
>
> The vocab-related portions of this spec describe a public-wiki design (canonical pool shared across users, anti-vandalism gating via `complement` / gated `replace` / admin `verify`, audit log via `content_vocab_edits`). That design shipped, then was reconsidered and replaced by a per-user private pool model.
>
> **The replacement spec is `docs/superpowers/specs/2026-05-04-vocab-pool-pivot-design.md`.** Read it for the current vocab design.
>
> **Sections of THIS spec that the pivot supersedes:**
> - Section 1 — `content_vocabs` table (now has `user_id NOT NULL`; dropped `is_verified` / `created_by` / `last_edited_by`)
> - Section 1 — `content_vocabs` partial unique index (now per-user `(user_id, content_key)`)
> - Section 1 — `content_vocab_edits` table — **deleted entirely**
> - Section 2 — `content_vocab_service` wiki ops (`ComplementContentVocab`, `ReplaceContentVocab`, `VerifyContentVocab`, `GetContentVocabByContent`, `GenerateContentVocabFields` SSE) — **all replaced** by per-user CRUD (`ListUserVocabs`, `CreateUserVocab`, `CreateUserVocabsBatch`, `UpdateUserVocab`, `DeleteUserVocab`, `GetUserVocabByContent`)
> - Section 2 — `AddVocabsToLevel` signature changed: now takes `vocabIDs []string` (from the user's pool) instead of `entries []string` (raw text); enforces ownership
> - Section 2 — Routes: dropped wiki endpoints (`GET /api/content-vocabs?content=`, `/complement`, `/verify`, `/generate-content-vocab-fields`); added user-pool routes (`/mine`, `/`, `/batch`, `DELETE`, `/generate-vocabs-from-keywords`)
> - Section 2 — `IsAdmin`, `CanReplaceVocab`, `MergeDefinition`, `SnapshotVocab`, `WriteVocabEdit` helpers — all dropped
> - Section 3 — `LevelVocabsPanel` no longer has inline add or AI补全 / Complement / Edit / Verify buttons; replaced by a `SelectVocabsDialog` picker that pulls from the user's pool. Wiki dialogs (`AddVocabsDialog`, `ComplementVocabDialog`, `EditVocabDialog`) deleted; replaced by a new "AI 词汇库" page with sibling components (`vocab-list`, `add-vocab-from-ai-dialog`, `add-vocab-manual-dialog`, `vocab-edit-dialog`)
> - Section 3 — New sidebar menu entry "AI 词汇库" above "AI 随心学" at route `/hall/ai-vocabs/`
> - Section 5 — `content_vocab_wiki_test.go` — deleted; replaced by `user_vocab_crud_test.go`
>
> **Sections of THIS spec that REMAIN ACCURATE:**
> - Section 1 — `content_metas`, `content_items`, `game_vocabs` schemas
> - Section 1 — All 6 tracking tables (polymorphism via two FK columns, soft-delete additions, partial unique indexes, XOR CHECK constraints)
> - Section 1 — POS consts (12-key set)
> - Section 2 — Word-sentence pipeline rewrite (`SaveMetadataBatch`, `BreakMetadata`, `GenerateContentItems`, etc.)
> - Section 2 — Mode-aware `PublishGame`, `countLevelItems`, `content_service.GetLevelContent` envelope synthesis
> - Section 4 — dx-mini compatibility (the synthesized envelope shape is preserved by the pivot — mini still works without changes)
> - Section 5 — `level_content_branching_test`, `tracking_polymorphic_test`, `game_vocab_placement_test` (the last had its `entries → vocabIDs` signature update folded in)
> - Section 6 — Migration & rollout principles
> - Section 7 — Phases 1–7 are accurate as shipped; Phase 8/9 partially altered by the pivot

---

## Background

The dx-api content layer today threads four tables:

- `content_metas` — raw metadata text (sentence/vocab) plus translation
- `content_items` — broken-down learning units (a sentence split into word/block/phrase/sentence units, or a single vocab unit)
- `game_metas` — junction `(game_id, game_level_id, content_meta_id, order)` allowing in-level repetition + per-user dedup
- `game_items` — junction `(game_id, game_level_id, content_item_id, order)` similarly

The same content tables back two very different game-mode families:

- `word-sentence` — sentences are AI-broken into many items; vocab metas optionally added too
- `vocab-battle` / `vocab-match` / `vocab-elimination` — vocab-only, batched (20 / 5 / 8 per batch); each meta becomes 1 item with no AI break, then DeepSeek populates the per-token `items` JSON

The junctions support per-user dedup (reusing the same content_meta across game-levels) and in-level repetition. This is over-engineered for the actual usage pattern — most placements are single-game, single-level, and the dedup logic adds significant code surface (`findExistingMetasForBatch`, `reuseItemsIntoLevel`, etc.).

For vocab modes specifically, there is a separate desire: the vocab pool should behave like a public wiki — once one user defines `fast → 快的`, every other user should benefit, and anyone can complement (add missing POS keys, audio, etc.) without being able to maliciously erase correct data.

This spec refactors the schema, the ai-custom flow, the play services, the dx-web editor, and the cross-cutting tracking tables to reflect that distinction.

## Goals

1. Drop the `game_metas` and `game_items` junctions; denormalize `(game_id, game_level_id, order)` directly onto `content_metas` and `content_items`.
2. Make `content_metas` / `content_items` exclusive to `word-sentence` mode.
3. Introduce a public-wiki vocab pool for the three vocab game modes, split across `content_vocabs` (canonical) and `game_vocabs` (placement junction).
4. Add an audit/revert log for the wiki (`content_vocab_edits`).
5. Update tracking tables to support vocab-mode answers via two-FK polymorphism.
6. Keep dx-mini's response shapes unchanged; absorb all branching server-side.
7. Maintain code-level FK constraints only (PostgreSQL native partitioning compat).
8. No lint/test regressions; no breaking changes to user-visible game play.

## Non-goals

- A wiki browse / search page in dx-web (deferred).
- TTS audio generation pipeline.
- Cross-language vocab support.
- Backfill of existing data (this is a fresh project — edit migrations in place).
- Rich edit-history UI on top of `content_vocab_edits` (table exists for revert; UI is admin-only).

## Naming conventions used here

- "WS" = `word-sentence` game mode.
- "Vocab modes" = `vocab-battle`, `vocab-match`, `vocab-elimination`.

## Section 1 — Schema

All tables follow the existing pattern: code-level FKs only, soft-delete + timestamps via `table.SoftDeletesTz()` then `table.TimestampsTz()` (so trailing column order is `deleted_at, created_at, updated_at`).

### `content_metas` (WS only)

```
id              uuid PK
game_id         uuid                ◄── NEW (code-level FK → games)
game_level_id   uuid                ◄── NEW (code-level FK → game_levels)
source_from     text default ''
source_type     text default ''     ◄── 'sentence' | 'vocab'
source_data     text default ''
translation     text null
speaker         text null
is_break_done   bool default false
order           double              ◄── NEW (lifted from game_metas)
deleted_at, created_at, updated_at
```

Indexes:
- `source_from`, `source_type`, `created_at`
- `(game_level_id, deleted_at, "order")` (replaces `idx_game_metas_level_order`)

Dropped:
- `idx_content_metas_dedup_lookup` — dedup feature is removed.

### `content_items` (WS only)

```
id              uuid PK
game_id         uuid                ◄── NEW
game_level_id   uuid                ◄── NEW
content_meta_id uuid null
content         text default ''
content_type    text default ''     ◄── 'word' | 'block' | 'phrase' | 'sentence'
uk_audio_url    text null
us_audio_url    text null
definition      text null
translation     text null
explanation     text null
speaker         text null
items           jsonb null
structure       jsonb null
order           double              ◄── NEW (lifted from game_items)
deleted_at, created_at, updated_at
```

Indexes:
- `content_meta_id`, `content_type`, `created_at`
- `(game_level_id, deleted_at, "order")` (replaces `idx_game_items_level_order`)

Model cleanup:
- Drop `Tags pq.StringArray` from `models/content_item.go` — the column was never created in any migration; this removes a dead model field.

### `content_vocabs` (NEW — canonical wiki source)

```
id              uuid PK
content         text default ''
content_key     text                ◄── lower(trim(content)) — dedup key
uk_phonetic     text null           ◄── AI-generated (e.g. "/fæst/")
us_phonetic     text null
uk_audio_url    text null
us_audio_url    text null
definition      jsonb null          ◄── [{"adj":"快的"},{"v":"斋戒"}]
explanation     text null
is_verified     bool default false  ◄── admin lock; locked rows admin-only edit
created_by      uuid null           ◄── code-level FK → users (first contributor)
last_edited_by  uuid null           ◄── code-level FK → users (most recent editor)
deleted_at, created_at, updated_at
```

Indexes:
- `UNIQUE(content_key) WHERE deleted_at IS NULL`
- `(content_key, deleted_at)`

### `game_vocabs` (NEW — placement junction, vocab modes only)

```
id                 uuid PK
game_id            uuid             ◄── code-level FK → games
game_level_id      uuid             ◄── code-level FK → game_levels
content_vocab_id   uuid             ◄── code-level FK → content_vocabs
order              double
deleted_at, created_at, updated_at
```

Indexes:
- `game_id`, `content_vocab_id`, `created_at`
- `(game_level_id, deleted_at, "order")`

In-level repetition is **allowed** (no UNIQUE on `(game_level_id, content_vocab_id)`), consistent with today's `game_items` pattern and with "duplication acceptable."

### `content_vocab_edits` (NEW — append-style audit/revert log)

```
id                 uuid PK
content_vocab_id   uuid             ◄── code-level FK → content_vocabs
editor_user_id     uuid null        ◄── code-level FK → users
edit_type          text             ◄── 'create' | 'complement' | 'replace' | 'verify' | 'delete'
before             jsonb null       ◄── full row snapshot pre-edit (null on 'create')
after              jsonb null       ◄── full row snapshot post-edit (null on 'delete')
deleted_at, created_at, updated_at
```

Indexes: `content_vocab_id`, `editor_user_id`, `created_at`.

The table follows the same `SoftDeletesTz + TimestampsTz` pattern as everything else; in practice rows are append-only.

### Cross-cutting tracking schema (two-FK polymorphism)

Six tables get a parallel `content_vocab_id` FK alongside their existing `content_item_id`:

| Table | Existing | Add | Constraint |
|---|---|---|---|
| `game_records` | `content_item_id uuid` | `content_vocab_id uuid null` | `CHECK ((content_item_id IS NULL) != (content_vocab_id IS NULL))` |
| `user_masters` | `content_item_id uuid` | `content_vocab_id uuid null` | exactly-one CHECK |
| `user_unknowns` | `content_item_id uuid` | `content_vocab_id uuid null` | exactly-one CHECK |
| `user_reviews` | `content_item_id uuid` | `content_vocab_id uuid null` | exactly-one CHECK |
| `game_reports` | `content_item_id uuid` | `content_vocab_id uuid null` | exactly-one CHECK |
| `game_sessions` | `current_content_item_id uuid null` | `current_content_vocab_id uuid null` | at-most-one CHECK (both null = no current item) |

**All 6 tables also gain `SoftDeletesTz()`** — they don't have it today (audit confirmed: only `content_metas` and `content_items` have soft-delete among the 8 affected tables). This is a project-pattern fix more than a behavioral one; soft-delete fits the convention used elsewhere.

`content_item_id` becomes nullable on the 5 non-session tables (it was `NOT NULL` before).

The existing simple `table.Unique(...)` calls in each create_table file are **removed**. The new partial uniques (which include `WHERE deleted_at IS NULL`) live in the per-table `*_raw.go` sibling migration:

```
-- per-table sibling raw SQL (per affected table)
CREATE UNIQUE INDEX idx_<table>_user_item_uq
  ON <table> (user_id, content_item_id)
  WHERE content_item_id IS NOT NULL AND deleted_at IS NULL

CREATE UNIQUE INDEX idx_<table>_user_vocab_uq
  ON <table> (user_id, content_vocab_id)
  WHERE content_vocab_id IS NOT NULL AND deleted_at IS NULL
```

Variants:
- `game_reports` uses `(user_id, content_*_id, reason)` (3-col)
- `game_records` uses `(game_session_id, content_*_id)`
- `game_sessions` has no unique pair, only the at-most-one CHECK

Both regular indexes (on `content_item_id` / `content_vocab_id`) also live in the sibling as partial indexes (`WHERE col IS NOT NULL`), since the columns are nullable.

Rationale (two-FK + CHECK over discriminator + single-ID):
- Single source of truth (kind is derivable from which FK is non-null).
- Standard Postgres pattern when the kinds are finite and known at design time (we have exactly 2).
- `CHECK` enforces the invariant at the DB level — no app-side drift, negligible write cost (~tens of ns per row, no read-side impact, no partition impact).
- Discriminator columns are mostly an ORM-driven convenience (Rails STI, Django GenericFK) for open-ended type sets; we don't need that flexibility.

Behavioral consequence of the partial uniques (`WHERE deleted_at IS NULL`):
- Soft-deleted rows do **not** occupy the unique slot, so re-marking after unmark is a plain `INSERT` of a new row — no UPSERT/reactivate needed.
- Each mark/unmark cycle leaves a soft-deleted history row; one always-active row exists at most.
- Service code: `MarkMastered` / `MarkUnknown` / `MarkReview` are simple `INSERT`s (after a "is there a live row already?" check for idempotency); `Unmark*` is `UPDATE ... SET deleted_at = NOW()`.

## Section 2 — ai-custom rewrite

### Word-sentence pipeline (`ai_custom_service.go`) — kept, simplified

The whole "play vocabs the word-sentence way" use case stays. Mixed `[S]/[V]` source types in WS mode are preserved.

Changes:

| Function | Before | After |
|---|---|---|
| `GenerateMetadata` | story from keywords | unchanged |
| `FormatMetadata` | clean text, emit `[S]/[V]` markers | unchanged |
| `SaveMetadataBatch` | per-user dedup via 4-table JOIN; reuse items if `is_break_done`; create game_metas row | **just inserts content_metas rows** with `(game_id, game_level_id, order)`; no dedup, no junction, no item reuse. Capacity check (sentences vs vocab ratio for WS) preserved by querying content_metas directly. |
| `BreakMetadata` | reads via `JOIN game_metas`; writes content_items + game_items | reads `content_metas WHERE game_level_id = ? AND is_break_done = false`; writes content_items directly with `(game_id, game_level_id, order = parent_meta_order + 10*(i+1))` (matches today's `baseOrder + 10*(i+1)` cadence) |
| `GenerateContentItems` | reads via `JOIN game_items`; updates `items` JSON | reads `content_items WHERE game_level_id = ? AND items IS NULL`; updates `items` |
| `ReorderMetadata` / `ReorderContentItems` | UPDATE on junction `"order"` | UPDATE on `content_metas."order"` / `content_items."order"` directly |
| `DeleteMetadata` / `DeleteContentItem` / `DeleteAllLevelContent` / `DeleteGame` / `DeleteLevel` | soft-delete junction rows + orphan-check canonical | soft-delete `content_metas` / `content_items` rows directly; no junctions, no orphan check |
| `InsertContentItem` | manual insert + game_items row | manual insert directly |
| `verifyMetaBelongsToGame` / `verifyItemBelongsToGame` | count rows in junction by game_id | check `content_metas.game_id` / `content_items.game_id` directly |
| `calculateInsertionOrder` | reference walks via `game_items` | reference walks via `content_items` ordered by `"order"` |

Helpers/types removed:
- `metaDedupKey`, `existingMetaRef`, `findExistingMetasForBatch`, `reuseItemsIntoLevel`, the `itemsByMetaCache` flow, the per-batch `maxItemOrderInLevel` plumbing.

### Vocab-game-mode pipeline — entirely new

`ai_custom_vocab_service.go` is **deleted**. Its endpoints
`/api/ai-custom/break-vocab-metadata` and `/api/ai-custom/generate-vocab-content-items` are removed.

`/api/ai-custom/generate-vocab` and `/api/ai-custom/format-vocab` are **kept** as raw text helpers (the user pastes the result into the vocab input box; they no longer create metas).

New AI enrichment endpoint:

```
POST /api/ai-custom/generate-content-vocab-fields  (SSE)
  body: { gameLevelId }
  effect: enriches every content_vocabs row referenced by this level's
          game_vocabs that has uk_phonetic IS NULL — fills uk_phonetic,
          us_phonetic, definition, explanation via DeepSeek
  cost:  per-word; failure refund per failed word
  audit: each updated canonical row gets a content_vocab_edits row
         (edit_type = 'complement', editor_user_id = current user)
```

Audio URL generation is out of scope.

New service files:

- `app/services/api/content_vocab_service.go` — wiki ops + AI enrichment
- `app/services/api/game_vocab_service.go` — placement ops
- `app/services/api/content_vocab_helpers.go` — definition merge, content_key normalization, edit-log writer, gating rules (created_by | admin | unverified-and-<24h)

#### Wiki operations (`content_vocab_service`)

```
GetByContent(content) → *ContentVocab
  -- normalized lookup; returns nil on miss

ComplementVocab(userID, vocabID, patch) → updated row
  -- additive merge:
  --   definition: append POS keys NOT already present (existing keys win)
  --   uk_phonetic / us_phonetic: set only if currently null
  --   uk_audio_url / us_audio_url: set only if currently null
  --   explanation: set only if currently null/empty
  -- NEVER overwrites existing values; vandalism floor.
  -- Writes content_vocab_edits('complement', before, after).
  -- Stamps last_edited_by = userID.

ReplaceVocab(userID, vocabID, patch) → updated row
  -- full overwrite of any field except is_verified, created_by, content_key.
  -- Allowed iff:
  --     userID == row.created_by
  --  OR userID is admin
  --  OR (row.is_verified == false AND now - row.created_at < 24h)
  -- Writes content_vocab_edits('replace', before, after).

VerifyVocab(adminUserID, vocabID, verified bool)
  -- admin only (per CLAUDE.md: username == "rainson")
  -- Sets is_verified, writes content_vocab_edits('verify', before, after).
```

#### Placement operations (`game_vocab_service`)

```
AddVocabsToLevel(userID, gameID, levelID, entries []string) → []AddedVocab
  -- For each non-empty, lowercase-trimmed entry:
  --   1. Validate as word/phrase: no punctuation other than ' or -.
  --   2. Look up canonical by content_key.
  --   3. Hit  → create game_vocabs row pointing at existing canonical_id.
  --      Miss → INSERT canonical (content, content_key, created_by = userID),
  --              write content_vocab_edits('create', null, snapshot),
  --              create game_vocabs row pointing at new canonical_id.
  -- Vocab batch-size enforced via consts.VocabBatchSize(game.Mode) on TOTAL
  -- count of game_vocabs in level after the batch is applied:
  --   vocab-match       → multiple of 5
  --   vocab-elimination → multiple of 8
  --   vocab-battle      → 0 (no batch constraint)
  -- A flat MaxMetasPerLevel (20) cap also applies, mirroring today's behavior.
  -- Returns array of {gameVocabId, contentVocabId, content, wasReused: bool}
  -- so the UI can render "用了已有词条" vs "新建词条" badges.

GetLevelVocabs(userID, gameID, levelID) → []LevelVocabData
  -- joins game_vocabs → content_vocabs ordered by gv.order

ReorderGameVocab(userID, gameID, gameVocabID, newOrder)
  -- UPDATE game_vocabs SET "order" = ? WHERE id = ?

DeleteGameVocab(userID, gameID, gameVocabID)
  -- soft-delete the placement only; canonical content_vocabs row stays
```

### Routes (`routes/api.go`)

Removed:
```
POST /api/ai-custom/break-vocab-metadata
POST /api/ai-custom/generate-vocab-content-items
```

Added:
```
POST   /api/ai-custom/generate-content-vocab-fields                   (SSE)

GET    /api/content-vocabs?content=<key>                              (lookup-by-content)
POST   /api/content-vocabs/{id}/complement
PUT    /api/content-vocabs/{id}                                       (replace; gated)
POST   /api/content-vocabs/{id}/verify                                (admin)

POST   /api/course-games/{id}/levels/{levelId}/game-vocabs            (batch add)
GET    /api/course-games/{id}/levels/{levelId}/game-vocabs            (list)
PUT    /api/course-games/{id}/game-vocabs/{gvId}/reorder
DELETE /api/course-games/{id}/game-vocabs/{gvId}

PUT    /api/play-single/{id}/content-vocab
PUT    /api/play-pk/{id}/content-vocab
PUT    /api/play-group/{id}/content-vocab
```

### Game play services (single / pk / group)

`countLevelItems` and the play-set loaders branch on `game.mode`:

```sql
-- WS mode (unchanged shape, queries direct table now)
SELECT * FROM content_items
WHERE game_level_id = ? AND deleted_at IS NULL
  [AND content_type IN (...degree filter...)]
ORDER BY "order" ASC

-- Vocab modes (new branch)
SELECT cv.*, gv.id AS gv_id, gv."order" AS gv_order
FROM content_vocabs cv
JOIN game_vocabs gv
  ON gv.content_vocab_id = cv.id AND gv.deleted_at IS NULL
WHERE gv.game_level_id = ? AND cv.deleted_at IS NULL
ORDER BY gv."order" ASC
```

Insert sites that record answers (`game_records`, `user_masters`, `user_unknowns`, `user_reviews`, `game_reports`) populate `content_vocab_id` for vocab modes and `content_item_id` for WS — exactly one of the two on every row. The lookup chain for vocab modes when the client sends back the answered item id (which is `game_vocab_id` — see Section 4): play service loads `game_vocabs` row → reads `content_vocab_id` → writes that into the record.

The PK robot loop in `game_play_pk_service.go` gets the same branching for both content load and record insertion.

## Section 3 — dx-web changes

### Word-sentence editor — minimal change
- `LevelUnitsPanel` keeps its two-panel layout (left: metadata, right: items).
- `AddMetadataDialog` keeps the `[S]/[V]` mixed input.
- "Duplicate skipped" toast → "Added" (no more dedup; same text twice creates two rows).
- All server actions (`saveMetadataAction`, `breakMetadata`, `generateContentItems`, `reorderMetaAction`, etc.) keep the same signatures; only their underlying queries change.

### Vocab-mode editor — replaced

`LevelVocabsPanel` (new) replaces `LevelUnitsPanel` entirely when `game.mode` is a vocab mode.

Single-list UI; each row shows:
- `content` (word/phrase)
- `definition` rendered as POS-keyed pills (e.g., `adj 快的` / `v 斋戒`)
- Phonetic chips (`UK /fæst/` / `US /fæst/`)
- Audio play buttons (when URLs present)
- Per-row actions: Complement, Edit (gated, see below), Delete

Add flow:
1. User types vocab list (or pastes via existing `format-vocab` AI helper which still returns text)
2. Submit → `POST /api/course-games/{id}/levels/{levelId}/game-vocabs` with the array
3. Response shows per-row badges: "用了已有词条" (reused) vs "新建词条" (created)
4. User can immediately click "AI 补全" → SSE call enriches all newly-created canonical rows; reused rows skipped (already enriched)

Complement flow (`ComplementVocabDialog`):
- Inline edit form showing current values for definition (POS keys), phonetic, audio, explanation
- Submit calls `/api/content-vocabs/{id}/complement`
- Toast feedback: "added 2 POS entries; phonetic/audio unchanged (already set)"

Replace flow:
- "Edit" button visible only when the user is `created_by` OR an admin OR the row is `is_verified=false` AND `now - created_at < 24h`. Disabled with tooltip otherwise.

Verify toggle:
- Visible only to admins. Toggling sets `is_verified` and writes a `content_vocab_edits('verify')` row.

### Action files

- `src/features/web/course/actions/content-vocab.action.ts` (NEW) — getByContent, complement, replace, verify
- `src/features/web/course/actions/game-vocab.action.ts` (NEW) — list, addBatch, reorder, delete
- `src/features/web/course/actions/course-game.action.ts` (EDIT) — drop dedup-related types

### API client (`src/lib/api-client.ts`)
- Add `contentVocabApi` (canonical wiki ops)
- Add `gameVocabApi` (placement ops)

### Tracking pages (mastered/unknown/review)
- Existing pages are unaffected; the API still returns `{content, contentType, translation}` shapes per the polymorphic-loading branch in the backend.
- No new "vocab" pill in v1; deferred.

### Browse page
- Wiki search/browse UI deferred; not in v1 scope.

## Section 4 — dx-mini compatibility

**dx-mini code: zero changes required.**

### `/api/games/{id}/levels/{lid}/content` shape preservation

For WS mode: query `content_items` directly, return existing `ContentItemData`.

For vocab modes: backend joins `game_vocabs` → `content_vocabs` and synthesizes the same envelope shape:

```jsonc
{
  "id": "<game_vocab_id>",      // placement id; mini treats this as the level "item"
  "content": "fast",
  "contentType": "vocab",        // new value, mini just passes through
  "translation": null,
  "definition": "快的; 快速地; 斋戒",   // joined from canonical definition JSON
  "items": [
    { "position": 1, "item": "fast",
      "phonetic": {"uk": "/fæst/", "us": "/fæst/"},
      "pos": "adj/v",
      "definition": "快的",
      "answer": true }
  ]
}
```

Mini's `buildChoices()` (`miniprogram/pages/games/play/play.ts`) reads `items[0]` and renders without modification.

### Tracking endpoints — shape preservation

The polymorphic loader branches on which FK is set:
- `content_item_id` non-null → `SELECT FROM content_items WHERE id IN ?`
- `content_vocab_id` non-null → `SELECT FROM content_vocabs WHERE id IN ?`

Both map to the same response shape `{content, translation, contentType}`:
- Vocab-row mapping: `translation` = first definition value (or null), `contentType = 'vocab'`.

### WebSocket
Untouched — no content-table joins.

## Section 5 — Migration & rollout, affected files, validation

### Migration file changes (no production data — edit in place)

**Naming convention:** sibling raw-SQL migrations are named `<prefix>_<table>_raw.go` and **must come immediately after their create_table file in prefix order**. Existing `*_indexes.go` siblings are renamed to `*_raw.go`. Some create_table files get **renumbered** to make room for their new sibling.

| Action | File | Notes |
|---|---|---|
| EDIT | `20260322000036_create_content_metas_table.go` | Add (game_id, game_level_id, order); column order per Section 1; drop dedup index. No sibling needed. |
| EDIT | `20260322000037_create_content_items_table.go` | Add (game_id, game_level_id, order); column order per Section 1. No sibling needed. |
| EDIT | `20260322000043_create_game_reports_table.go` | Add SoftDeletesTz, add nullable `content_vocab_id`, make `content_item_id` nullable, drop existing simple `Unique(...)` (moves to sibling). |
| ADD | `20260322000044_add_game_reports_raw.go` | NEW sibling: regular indexes + partial uniques + XOR CHECK (raw SQL). |
| RENUMBER + EDIT | `20260322000044_create_user_masters_table.go` → `20260322000045_create_user_masters_table.go` | Same edit pattern as game_reports. |
| ADD | `20260322000046_add_user_masters_raw.go` | NEW sibling. |
| RENUMBER + EDIT | `20260322000045_create_user_unknowns_table.go` → `20260322000047_create_user_unknowns_table.go` | Same. |
| ADD | `20260322000048_add_user_unknowns_raw.go` | NEW sibling. |
| RENUMBER + EDIT | `20260322000046_create_user_reviews_table.go` → `20260322000049_create_user_reviews_table.go` | Same. |
| ADD | `20260322000050_add_user_reviews_raw.go` | NEW sibling. |
| EDIT | `20260405000002_create_game_sessions_table.go` | Add SoftDeletesTz, add nullable `current_content_vocab_id`. |
| RENAME + EDIT | `20260405000003_add_game_session_indexes.go` → `20260405000003_add_game_sessions_raw.go` | Same prefix; pluralized name to match table. **Append** index on `current_content_vocab_id` + at-most-one CHECK. Existing index statements untouched. |
| EDIT | `20260405000004_create_game_records_table.go` | Same pattern as game_reports/user_*. |
| ADD | `20260405000005_add_game_records_raw.go` | NEW sibling. |
| RENUMBER | `20260405000005_create_game_pks_table.go` → `20260405000006_create_game_pks_table.go` | Bumped to make room for game_records sibling. No content change. |
| RENAME + RENUMBER | `20260405000006_add_game_pk_indexes.go` → `20260405000007_add_game_pks_raw.go` | Renamed and bumped; content unchanged (was already a raw-SQL sibling). |
| DELETE | `20260414000001_create_game_metas_and_game_items_tables.go` | Junctions are gone with the refactor. |
| ADD (at freed prefix) | `20260414000001_create_content_vocabs_and_game_vocabs_tables.go` | Two tables in one file (mirrors the old "create both junctions in one file" pattern). |
| ADD | `20260414000002_add_content_vocabs_raw.go` | NEW sibling for `content_vocabs`: partial unique on `content_key` (`WHERE deleted_at IS NULL`). `game_vocabs` needs no sibling. |
| ADD | `20260414000003_create_content_vocab_edits_table.go` | NEW table; no sibling needed. |
| EDIT | `bootstrap/migrations.go` | Drop `M20260414000001CreateGameMetasAndGameItemsTables`. Register all new + renumbered files in correct order. Update struct names for renumbered create files. |

### Models — delete

```
DELETE: app/models/game_meta.go
DELETE: app/models/game_item.go
```

These models have no callers after the service refactor. Remove the imports from any file that still references them.

### `*_raw.go` sibling structure (template)

All `*_raw.go` siblings follow the existing `add_game_session_indexes.go` shape — a `Signature()` returning the file's prefix, an `Up()` running an ordered slice of raw SQL `Exec(...)` calls, a `Down()` running the mirrored `DROP` slice. Example for `add_game_reports_raw.go`:

```go
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
    // ... loop + Exec + return
}
```

`add_user_*_raw.go` use the 2-col `(user_id, content_*_id)` unique (no `reason`). `add_game_records_raw.go` uses `(game_session_id, content_*_id)`. `add_game_sessions_raw.go` (renamed file) appends only the index on `current_content_vocab_id` and the at-most-one CHECK to its existing statement slice.

### Affected Go files

**Models** (`app/models/`):
- DELETE: `game_meta.go`, `game_item.go` (the junctions are gone)
- ADD: `content_vocab.go`, `game_vocab.go`, `content_vocab_edit.go`
- EDIT: `content_meta.go`, `content_item.go` (add fields, drop `Tags` from content_item)
- EDIT: `game_record.go`, `game_session.go`, `user_master.go`, `user_unknown.go`, `user_review.go`, `game_report.go` (add `ContentVocabID *string`; make existing `ContentItemID` a pointer; add `orm.SoftDeletes` to the 6 tracking models since the underlying tables now have `deleted_at`)

**Consts** (`app/consts/`):
- ADD: `pos.go` — 12-key set: n, v, adj, adv, prep, conj, pron, art, num, int, aux, det; `AllPos []string`; `IsValidPos(s string) bool`

**Services** (`app/services/api/`):
- HEAVY EDIT: `course_content_service.go`
- HEAVY EDIT: `ai_custom_service.go`
- DELETE: `ai_custom_vocab_service.go`
- EDIT: `course_game_service.go` (DeleteGame, DeleteLevel, PublishGame)
- EDIT: `content_service.go` (`GetLevelContent` mode-branching + envelope synthesis)
- EDIT: `game_play_single_service.go`, `game_play_pk_service.go`, `game_play_group_service.go`
- EDIT: `user_master_service.go`, `user_unknown_service.go`, `user_review_service.go`, `feedback_service.go`
- ADD: `content_vocab_service.go`, `game_vocab_service.go`, `content_vocab_helpers.go`

**Controllers** (`app/http/controllers/api/`):
- DELETE: `ai_custom_vocab_controller.go`
- EDIT: `ai_custom_controller.go` (drop break-vocab + gen-vocab-items handlers; add `GenerateContentVocabFields` SSE handler)
- ADD: `content_vocab_controller.go`, `game_vocab_controller.go`
- EDIT: `course_game_controller.go` (re-aim content-items endpoints; signatures stable)
- EDIT: `game_play_single_controller.go`, `game_play_pk_controller.go`, `game_play_group_controller.go` (add `UpdateContentVocab` siblings)
- EDIT: `user_master_controller.go`, `user_unknown_controller.go`, `user_review_controller.go`, `game_report_controller.go` (accept `contentVocabId`)

**Requests** (`app/http/requests/api/`):
- EDIT: `user_master_request.go`, `user_unknown_request.go`, `user_review_request.go`, `game_report_request.go`, `session_request.go` — accept optional `contentVocabId`; validate "exactly one of contentItemId / contentVocabId"
- ADD: `content_vocab_request.go`, `game_vocab_request.go`

**Routes** (`routes/api.go`):
- DELETE: `/api/ai-custom/break-vocab-metadata`, `/api/ai-custom/generate-vocab-content-items`
- ADD: see Section 2 routes list

**Console commands** (`app/console/commands/`):
- DELETE: `backfill_metas.go` (linked content_items to a synthetic content_meta via game_metas — obsolete with the new schema and no production data)
- EDIT: `import_courses.go` if it references game_metas/game_items; otherwise leave untouched. Decision deferred to implementation pass after reading the file

### Affected dx-web files

- `src/features/web/course/components/LevelUnitsPanel.tsx` (and friends) — mode-branch render: WS unchanged path, vocab modes route to `LevelVocabsPanel`
- `src/features/web/course/components/LevelVocabsPanel.tsx` (NEW)
- `src/features/web/course/components/AddVocabDialog.tsx` (REWRITE)
- `src/features/web/course/components/ComplementVocabDialog.tsx` (NEW)
- `src/features/web/course/actions/content-vocab.action.ts` (NEW)
- `src/features/web/course/actions/game-vocab.action.ts` (NEW)
- `src/features/web/course/actions/course-game.action.ts` (EDIT — drop dedup types)
- `src/lib/api-client.ts` (EDIT — add contentVocabApi, gameVocabApi)
- Shared types updated for two-FK polymorphism on tracking

### Tests

- DELETE: `tests/feature/course_content_dedup_test.go` (the feature it tests is being removed)
- ADD: `tests/feature/content_vocab_wiki_test.go` — create / lookup / complement (additive merge) / replace gating / verify lock / edits log
- ADD: `tests/feature/game_vocab_placement_test.go` — add to level / in-level repetition allowed / batch-size enforcement / reorder / delete
- ADD: `tests/feature/level_content_branching_test.go` — `/api/games/{id}/levels/{lid}/content` returns the right shape for both WS and vocab modes
- ADD: `tests/feature/tracking_polymorphic_test.go` — mark mastered as item (WS) and as vocab; list returns both with consistent shape
- KEEP and UPDATE existing tests that reference the dropped junctions

### Validation gates (every commit boundary)

- `cd dx-api && gofmt -l . && go vet ./... && go build ./... && go test -race ./...`
- `cd dx-web && npm run lint && npx tsc --noEmit && npm run build`
- `staticcheck ./...` (per the project's hooks)
- After backend edits, smoke-curl every changed route with response shapes

### Implementation phases

This stays as one design spec; implementation breaks into ordered phases (each phase ends with the validation gates above passing):

1. **Phase 1 — Schema** — migration edits + model edits, run migrations.
2. **Phase 2 — Word-sentence backend** — rewrite `course_content_service`, `ai_custom_service`, `course_game_service` for direct queries; verify play services keep working.
3. **Phase 3 — Vocab wiki backend** — `content_vocab_service` + `game_vocab_service` + helpers + controllers + routes + AI enrichment.
4. **Phase 4 — Tracking polymorphism** — the 6 tables + their services + controllers + requests.
5. **Phase 5 — dx-web** — vocab editor split, new dialogs, action files, API client.
6. **Phase 6 — Tests + smoke + cleanup** — delete obsolete code, drop dead consts, drop `Tags` from model, smoke-test golden paths.

## Open questions

None at sign-off. All resolved during brainstorming:
- Wiki shape: two-table (content_vocabs + game_vocabs).
- Anti-vandalism: complement = additive merge (anyone); replace = gated (creator/admin/<24h-unverified); admin verify-lock; full audit via content_vocab_edits.
- POS keys: standard 12-key set.
- WS mode keeps mixed sentence + vocab content (the "play vocabs the WS way" use case).
- Polymorphism: two FK columns + CHECK (no discriminator); CHECK is partition-safe (row-local, no coordination).
- Soft-delete added to all 6 tracking tables (was missing); aligns with project pattern.
- Mark/unmark behavior: insert-new on each mark, soft-delete on unmark (partial unique with `WHERE deleted_at IS NULL` allows it).
- Sibling raw-SQL migrations: per-table `*_raw.go`, named to follow create_table file in prefix order; `add_game_session_indexes.go` and `add_game_pk_indexes.go` renamed to `*_raw.go` form.
- `game_metas` / `game_items` migration file + model files deleted entirely.
- Wiki browse page: deferred.
- Audio TTS: out of scope.

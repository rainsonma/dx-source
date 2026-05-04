# Vocab Pool Pivot — Design (delta)

Date: 2026-05-04
Status: Approved (pending implementation)
Supersedes: parts of `2026-05-03-content-vocabs-refactor-design.md` covering the public-wiki design

## Background

The 2026-05-03 refactor shipped vocab as a public wiki: one canonical `content_vocabs` row per `content_key` shared across all users, with anti-vandalism gating (additive complement, gated replace, admin verify, audit log via `content_vocab_edits`).

Re-evaluation: vocab is actually per-user content, not a shared knowledge base. The public-wiki model added complexity (gating, audit, admin lock, `Complement` vs `Edit` UX split) that earned no real value because the access pattern is single-user-per-vocab.

This spec replaces the wiki model with a per-user private pool, decouples vocab management from game authoring (its own page + sidebar menu), and lets vocab-mode level editors *select* from the user's existing pool instead of re-typing words.

## Goals

1. Each user owns a private vocab pool (`content_vocabs.user_id` added).
2. Drop the public-wiki machinery: `content_vocab_edits` table, the `Complement`/`Replace`/`Verify` endpoints, the gating helpers, the audit log writes, the AI enrichment SSE on canonical rows.
3. Add a dedicated "AI 词汇库" page where users browse / add (AI-generated or manual paste) / edit / delete their own vocabs.
4. Add a sidebar menu entry "AI 词汇库" above "AI 随心学".
5. Vocab-mode level editors get a "select from my vocabs" dialog (checkbox picker + search + batch-size enforcement). No more inline vocab creation in the level editor.
6. Preserve the live link: editing a vocab on the user's page propagates to every game-level placement (since `game_vocabs` references the canonical row).

## Non-goals

- Cross-user sharing or vocab discovery. (User pools are private.)
- Edit history / revert. (User edits their own data freely; no audit needed.)
- Admin moderation. (Nothing to moderate — no shared content.)
- TTS audio generation.
- Migrating from the previous wiki model with data preservation. (Fresh DB; edit migrations in place.)

## Section 1 — Schema delta

### `content_vocabs` (modified)

```
id              uuid PK
user_id         uuid             ◄── NEW (code-level FK → users)
content         text default ''
content_key     text             ◄── lower(trim(content)) — per-user dedup key
uk_phonetic     text null
us_phonetic     text null
uk_audio_url    text null
us_audio_url    text null
definition      jsonb null
explanation     text null
deleted_at, created_at, updated_at

DROPPED:
  is_verified
  created_by
  last_edited_by
```

Indexes (sibling raw-SQL migration):
- `UNIQUE(user_id, content_key) WHERE deleted_at IS NULL` (was `UNIQUE(content_key) WHERE deleted_at IS NULL`)
- `INDEX(user_id)` for user-scoped list queries
- `INDEX(content_key, deleted_at)` (kept for completeness; cross-user search not exposed)

### `content_vocab_edits` (deleted)

Table + model + migration deleted entirely. No audit log.

### `game_vocabs` (unchanged)

Still `(id, game_id, game_level_id, content_vocab_id, order)` with soft-delete. Still allows in-level repetition. The fact that the canonical row now belongs to a user is internal — `game_vocabs` doesn't change shape.

### Other tables — unchanged

`content_metas`, `content_items`, the 6 tracking tables — all unchanged from the prior refactor.

## Section 2 — Backend delta

### `content_vocab_helpers.go`

**Keep:**
- `NormalizeVocabContent`, `ValidateVocabContent`
- `ValidateDefinition`, `ValidatePosEntries`
- Error sentinels: `ErrVocabContentEmpty`, `ErrVocabContentInvalid`, `ErrVocabNotFound`, `ErrInvalidPosKey`

**Drop:**
- `IsAdmin`, `CanReplaceVocab`, `unverifiedEditWindow`, `adminUsername`
- `MergeDefinition` (no more wiki additive merge)
- `SnapshotVocab`, `WriteVocabEdit`
- Error sentinels: `ErrVocabNotEditable`, `ErrVocabAdminOnly`

### `content_vocab_service.go` — replace with user-scoped CRUD

**Drop:**
- `GetContentVocabByContent` (replaced by user-scoped lookup)
- `ComplementContentVocab`, `ReplaceContentVocab`, `VerifyContentVocab`
- `GenerateContentVocabFields`, `enrichContentVocab`, `vocabFieldsPrompt`
- `VocabComplementPatch`, `VocabReplacePatch` types

**Add:**

```go
// Public response shape — note: no createdBy / lastEditedBy / isVerified anymore.
type ContentVocabData struct {
    ID          string  `json:"id"`
    Content     string  `json:"content"`
    UkPhonetic  *string `json:"ukPhonetic"`
    UsPhonetic  *string `json:"usPhonetic"`
    UkAudioURL  *string `json:"ukAudioUrl"`
    UsAudioURL  *string `json:"usAudioUrl"`
    Definition  *string `json:"definition"`     // JSON string [{pos: gloss}, ...]
    Explanation *string `json:"explanation"`
    CreatedAt   any     `json:"createdAt"`
    UpdatedAt   any     `json:"updatedAt"`
}

type VocabInput struct {
    Content     string              `json:"content"`
    Definition  []map[string]string `json:"definition"`
    UkPhonetic  *string             `json:"ukPhonetic"`
    UsPhonetic  *string             `json:"usPhonetic"`
    UkAudioURL  *string             `json:"ukAudioUrl"`
    UsAudioURL  *string             `json:"usAudioUrl"`
    Explanation *string             `json:"explanation"`
}

// ListUserVocabs returns the user's vocabs paginated by ID cursor (UUIDv7
// time-sortable). search filters by content_key contains lower(trim(query)).
func ListUserVocabs(userID, cursor, search string, limit int) (items []ContentVocabData, nextCursor string, hasMore bool, err error)

// GetUserVocabByContent looks up by (user_id, content_key); returns nil on miss.
func GetUserVocabByContent(userID, content string) (*ContentVocabData, error)

// CreateUserVocab inserts a new vocab. ErrVocabContentEmpty / Invalid on bad
// content; if content_key already exists for this user, returns that existing
// row (idempotent — caller can decide whether to surface the dup).
func CreateUserVocab(userID string, in VocabInput) (*ContentVocabData, error)

// CreateUserVocabsBatch — used by the AI-generated flow + manual paste flow.
// Each input row is normalize-deduped against the user's pool (existing rows
// are returned unchanged; new rows are inserted). Returns one ContentVocabData
// per input in order, plus a per-row wasReused flag.
func CreateUserVocabsBatch(userID string, inputs []VocabInput) ([]CreateVocabResult, error)

type CreateVocabResult struct {
    Vocab     *ContentVocabData `json:"vocab"`
    WasReused bool              `json:"wasReused"`
}

// UpdateUserVocab is a full overwrite of the user's own row. Returns
// ErrVocabNotFound if the row doesn't exist OR doesn't belong to userID.
// content_key recomputed from new content; if the new content_key collides
// with another of the user's vocabs, returns ErrDuplicateVocab.
func UpdateUserVocab(userID, vocabID string, in VocabInput) (*ContentVocabData, error)

// DeleteUserVocab soft-deletes; verifies ownership. Game-level placements
// (game_vocabs rows) referencing this vocab are NOT cascaded — the placement
// loader filters via cv.deleted_at IS NULL, so soft-deleted vocabs disappear
// from game play automatically. (Trade-off acknowledged: deleting a vocab
// silently removes it from any game-level using it. Document in the UI.)
func DeleteUserVocab(userID, vocabID string) error
```

### AI generation endpoint (new)

```go
// GenerateVocabsFromKeywords: 5 keywords → 20 vocab JSON entries with phonetic
// + definition + explanation. Returns text the user reviews before "使用".
// Bean cost = aiGenerateCost (5 beans). Refunds on AI failure.
//
// Output JSON shape (one element per generated vocab):
// [{"content":"fast","ukPhonetic":"/fɑːst/","usPhonetic":"/fæst/",
//   "definition":[{"adj":"快的"},{"v":"斋戒"}],"explanation":"形容速度快..."}, ...]
func GenerateVocabsFromKeywords(userID string, keywords []string) (string, error)
```

The existing `GenerateVocab` and `FormatVocab` text helpers stay (used by the manual paste flow's "检查并格式化" button — they output text the user pastes; we then call `CreateUserVocabsBatch` with the parsed entries).

### `game_vocab_service.go` — `AddVocabsToLevel` signature change

**Was** (wiki design):
```go
func AddVocabsToLevel(userID, gameID, levelID string, entries []string) ([]AddedGameVocab, error)
// entries are raw vocab strings; service creates canonical rows on the fly
```

**Now**:
```go
func AddVocabsToLevel(userID, gameID, levelID string, vocabIDs []string) ([]AddedGameVocab, error)
// vocabIDs are existing content_vocabs.id values from the user's pool.
// Service:
//   1. Loads each vocab; verifies content_vocabs.user_id == userID for all.
//      Mismatch → ErrForbidden.
//   2. Validates batch-size against game.Mode (5 / 8 / 20).
//   3. Capacity check against MaxMetasPerLevel.
//   4. Creates game_vocabs placement rows, one per vocabID in order.
// In-level repetition still allowed (same vocabID can appear twice).
```

`AddedGameVocab` no longer has `wasReused` — that flag was about wiki canonical reuse; here we're always selecting from the user's pool, every entry IS a "reuse." Drop the field. Returns `[]AddedGameVocab{ GameVocabID, ContentVocabID, Content }`.

`GetLevelVocabs`, `ReorderGameVocab`, `DeleteGameVocab` — unchanged (the read shape is preserved; level play and the level editor see the same data).

### Routes delta (`routes/api.go`)

**Drop:**
- `GET /api/content-vocabs?content=...` (was `getByContent` for wiki lookup)
- `POST /api/content-vocabs/{id}/complement`
- `POST /api/content-vocabs/{id}/verify`
- `POST /api/ai-custom/generate-content-vocab-fields` (the canonical-row enrichment SSE)

**Keep, edit semantics:**
- `PUT /api/content-vocabs/{id}` — now: full update of *own* vocab, no gating beyond ownership
- `POST /api/course-games/{id}/levels/{levelId}/game-vocabs` — body now `{ vocabIds: string[] }` (was `{ entries: string[] }`)
- `GET /api/course-games/{id}/levels/{levelId}/game-vocabs` — unchanged
- `PUT /api/course-games/{id}/game-vocabs/{gvId}/reorder` — unchanged
- `DELETE /api/course-games/{id}/game-vocabs/{gvId}` — unchanged

**Add:**
- `GET /api/content-vocabs/mine?cursor=&search=&limit=` — paginated list of user's vocabs
- `POST /api/content-vocabs` — create one user vocab (used by manual single-add, not common)
- `POST /api/content-vocabs/batch` — create many (used by AI flow + paste flow)
- `DELETE /api/content-vocabs/{id}` — soft-delete own vocab
- `POST /api/ai-custom/generate-vocabs-from-keywords` — AI 5 keywords → 20 vocabs

### `ContentVocabController` + `GameVocabController` updates

Mirror the service changes. Add `mapVocabError` cases for new errors (e.g., `ErrDuplicateVocab` if the user tries to update content to collide with another of their existing vocabs).

## Section 3 — dx-web delta

### New page + sidebar menu

**Sidebar menu** (find the sidebar config — likely `consts/hall_menu.ts` or in the layout component):
- Add entry "AI 词汇库" with route `/hall/ai-vocabs`, position immediately ABOVE the existing "AI 随心学" entry.

**New page**: `app/(web)/hall/(main)/ai-vocabs/page.tsx` — sibling of `ai-custom/`.

Layout:
- Top bar: "AI 词汇库" title; right-aligned: "AI 生成" button + "手动添加" button + search input (filters by `content_key contains query`).
- List: paginated table-like list of vocabs. Each row shows: content, definition POS pills, phonetic chips, audio play buttons (when URLs present), Edit / Delete actions.
- Empty state: helpful prompt ("还没有词汇 — 使用 AI 生成 或 手动添加").

### New components (under `features/web/ai-vocabs/components/`)

- `vocab-list.tsx` — paginated list with search; reads via `contentVocabApi.listMine`.
- `add-vocab-from-ai-dialog.tsx` — modal: 5 keyword inputs (or comma-separated textarea); "AI 生成" button calls `/api/ai-custom/generate-vocabs-from-keywords`; shows the 20 results in a preview table; "使用" button checks all (user can uncheck unwanted ones); "保存" calls `contentVocabApi.createBatch` with the selected entries; toast on success.
- `add-vocab-manual-dialog.tsx` — modal: textarea for paste; "检查并格式化" button calls `/api/ai-custom/format-vocab` (existing endpoint), replaces textarea with formatted output; "保存" parses textarea into VocabInput[] and calls `contentVocabApi.createBatch`.
- `vocab-edit-dialog.tsx` — modal: editable form for one vocab (content, definition POS rows, phonetic, audio, explanation). "保存" calls `contentVocabApi.update`. Reuse the structure from the existing wiki `EditVocabDialog` but drop all gating logic (it's the user's own data).

### Picker dialog (used in level editor)

`select-vocabs-dialog.tsx` — under `features/web/ai-custom/components/` (same dir as `level-vocabs-panel`).

Props: `{ gameMode, levelId, alreadyPlacedVocabIds: string[], batchSize: number, onClose, onSelected: (vocabIds: string[]) => void }`.

UI:
- Search input filtering the user's pool
- Checkbox list of vocabs (same row layout as the new vocab manager)
- Already-placed vocabs are checked + disabled (prevents accidental re-add unless intentional — actually, the spec says in-level repetition is allowed, so they should be re-selectable; design choice — keep them re-selectable but visually marked)
- Footer: shows "已选 N / 应选 M" where M is batch-size; disable submit until N % batch-size === 0 (or N is exactly the right count for that mode)
- Submit calls `gameVocabApi.add(gameId, levelId, selectedVocabIds)`; closes dialog, parent refreshes its list.

### Edit `level-vocabs-panel.tsx` (vocab-mode level editor)

- Replace "添加" button → "选择词汇" — opens `select-vocabs-dialog.tsx`
- DROP "AI 补全" button (no more SSE enrichment of canonical rows from this page)
- DROP per-row "Complement", "Edit", "Verify" buttons (vocab editing happens on the AI 词汇库 page; this page is just for *placement* management — list + reorder + delete placement)
- Keep per-row "Delete" — soft-deletes the placement only (not the canonical vocab)
- Keep per-row content/definition/phonetic/audio rendering for clarity

### Drop these components

- `add-vocabs-dialog.tsx` (the wiki add dialog)
- `complement-vocab-dialog.tsx`
- `edit-vocab-dialog.tsx` (the wiki version) — replaced by `vocab-edit-dialog.tsx` on the new page
- `level-vocabs-panel.tsx`'s wiki helpers (e.g., `useNow`/`canEdit` 24h-gate logic)

### `api-client.ts` updates

`contentVocabApi`:
- DROP: `getByContent`, `complement`, `verify`
- KEEP: `replace` → rename to `update` for clarity
- ADD: `listMine(cursor, search, limit)`, `create(input)`, `createBatch(inputs)`, `delete(id)`

`gameVocabApi`:
- `add(gameId, levelId, vocabIds: string[])` — body changes from `{ entries }` to `{ vocabIds }`
- Other methods unchanged

Types:
- DROP: `ContentVocabComplementPatch`, `ContentVocabReplacePatch`, `is_verified` / `createdBy` / `lastEditedBy` from `ContentVocabData`
- ADD: `VocabInput`, `CreateVocabResult`
- `AddedGameVocab` drops `wasReused`

### Action files

`content-vocab.action.ts`:
- DROP: `getVocabByContent`, `complementVocabAction`, `verifyVocabAction`
- KEEP, EDIT: `replaceVocabAction` → `updateVocabAction`
- ADD: `listMyVocabsAction`, `createVocabAction`, `createVocabsBatchAction`, `deleteVocabAction`, `generateVocabsFromKeywordsAction`

`game-vocab.action.ts`:
- `addGameVocabsAction(gameId, levelId, vocabIds)` — signature update

## Section 4 — dx-mini compatibility

Unchanged. Vocab play still reads `game_vocabs JOIN content_vocabs` and synthesizes the same `ContentItemData` envelope. The fact that the canonical row has a `user_id` is invisible to mini.

## Section 5 — Test delta

**Drop:**
- `content_vocab_wiki_test.go` — wiki feature gone

**Update:**
- `game_vocab_placement_test.go` — `AddVocabsToLevel` now takes vocab IDs from the user's pool. Add a test for `ErrForbidden` when a user tries to attach another user's vocab.
- `level_content_branching_test.go` — fixture setup must seed the canonical vocab with a `user_id` (the game owner). Read shape unchanged; assertions unchanged.

**Add:**
- `user_vocab_crud_test.go` — covers:
  - `TestCreate_NewVocab_StoredWithUser`: insert; verify `user_id` set.
  - `TestCreate_DuplicateContentKey_ReturnsExisting`: same content_key for same user → returns existing row, doesn't create duplicate.
  - `TestCreate_SameContentKey_AcrossUsers_TwoRows`: User A and User B can each have "fast" — distinct rows.
  - `TestList_ReturnsOnlyOwnerVocabs`: User A's list query never returns User B's vocabs.
  - `TestUpdate_OwnVocab_Succeeds`: editing your own row works.
  - `TestUpdate_OthersVocab_ErrVocabNotFound`: editing someone else's row by ID returns `ErrVocabNotFound` (not `ErrForbidden` — we treat it as not-found from this user's perspective).
  - `TestDelete_SoftDelete`: deletes set `deleted_at`; row still in DB but excluded from list.
  - `TestDelete_PlacementsAlsoExcludedFromPlay`: after vocab delete, the synthesized envelope from `GetLevelContent` no longer includes that vocab in any game-level it was placed in.

**Keep unchanged:**
- `tracking_polymorphic_test.go`

## Section 6 — Migration & rollout

We're still on the fresh-DB / no-production-data assumption from the original refactor. **Edit migrations in place** rather than adding ALTERs:

- `20260414000001_create_content_vocabs_table.go` — add `Uuid("user_id")` (NOT nullable), drop `Boolean("is_verified")` / `Uuid("created_by")` / `Uuid("last_edited_by")`. Add `Index("user_id")`.
- `20260414000002_add_content_vocabs_raw.go` — change the partial unique from `(content_key) WHERE deleted_at IS NULL` to `(user_id, content_key) WHERE deleted_at IS NULL`. Keep the `(content_key, deleted_at)` index.
- `20260414000004_create_content_vocab_edits_table.go` — DELETE entirely. Update `bootstrap/migrations.go`.

Re-reset dev DB and run migrations (same procedure as Phase 1 Group F of the original plan).

## Section 7 — Implementation phases

Three phases, on the same `refactor/content-vocabs` branch. Each ends with build + lint + test gates passing.

### Phase X1 — Schema + model + bootstrap diff

- Edit `create_content_vocabs_table.go` (add user_id, drop is_verified/created_by/last_edited_by, add user_id index)
- Edit `add_content_vocabs_raw.go` (per-user partial unique)
- DELETE `create_content_vocab_edits_table.go` migration + `content_vocab_edit.go` model
- Update `bootstrap/migrations.go`
- Update `content_vocab.go` model (add `UserID`, drop dropped fields)
- Re-reset dev DB + run migrations
- Gate: `gofmt -l`, `go vet`, `go build`, schema verified via `\d content_vocabs`

### Phase X2 — Backend rewrite

- Edit `content_vocab_helpers.go` (drop wiki helpers; keep validators + sentinels)
- Replace `content_vocab_service.go` (drop wiki ops; add user-CRUD + AI keyword endpoint helper)
- Edit `game_vocab_service.go` (`AddVocabsToLevel` takes vocabIds; ownership check)
- Edit `ai_custom_service.go` + controller (add `GenerateVocabsFromKeywords` + handler)
- Edit `content_vocab_controller.go` + `game_vocab_controller.go` (drop / add handlers per route table)
- Edit `content_vocab_request.go` + `game_vocab_request.go` (drop / add request types)
- Edit `routes/api.go` (drop / add routes)
- Gate: `gofmt`, `go vet`, `go build`, `go test -race ./...`. Update `game_vocab_placement_test.go`. Add `user_vocab_crud_test.go`. Delete `content_vocab_wiki_test.go`.

### Phase X3 — dx-web rewrite

- Edit `api-client.ts`: drop wiki API methods + types; add user-vocab + AI keyword methods + types
- Edit action files: drop wiki actions; add user-vocab + AI keyword actions
- Add new page `app/(web)/hall/(main)/ai-vocabs/page.tsx`
- Add new components under `features/web/ai-vocabs/components/`: `vocab-list`, `add-vocab-from-ai-dialog`, `add-vocab-manual-dialog`, `vocab-edit-dialog`
- Add `select-vocabs-dialog.tsx` under `features/web/ai-custom/components/`
- Edit `level-vocabs-panel.tsx` (replace Add → 选择词汇 button; drop wiki actions; drop AI 补全 button; drop 24h-gate code)
- Delete obsolete components: `add-vocabs-dialog`, `complement-vocab-dialog`, `edit-vocab-dialog` (wiki version)
- Edit sidebar menu config: insert "AI 词汇库" entry above "AI 随心学"
- Gate: `npm run lint`, `npx tsc --noEmit`, `npm run build`. Manual smoke test (you).

## Open questions

None at sign-off. All resolved during brainstorming:
- `content_vocabs.user_id` per-user, partial unique on `(user_id, content_key)`.
- `content_key` retained — explicit dedup column over functional index.
- "AI 词汇库" page route: `/hall/ai-vocabs/`.
- Level-editor "select from my vocabs" is a checkbox dialog with search + batch-size enforcement.
- Editing a vocab on the user's page propagates to every placement (live link via `game_vocabs.content_vocab_id`).
- Deleting a vocab silently removes it from any game-level it was placed in (acknowledged trade-off; UI message warns user).
- No edit history / no audit log (it's the user's own data).
- `IsAdmin` helper deleted (no admin gating left in the vocab area).
- dx-mini unaffected.

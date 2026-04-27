# Rename items[*].translation → items[*].definition Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rename the per-token gloss key inside the `content_items.items` JSONB array from `translation` to `definition` across the full stack (Go writers, AI prompt, Go test, dx-web type), without breaking any existing behavior.

**Architecture:** Two atomic commits. Commit 1 changes dx-api (Go struct + JSON tag + three struct-literal assignments + Go test + DeepSeek `genItemsPrompt`) — these MUST move together so the parser, the prompt, and the persisted shape stay in sync. Commit 2 changes the dx-web `SpellingItem` type (no current readers — type-only enforcement for future correctness). No DB migration (D1: pre-launch).

**Tech Stack:** Go 1.22+ with Goravel/GORM, TypeScript 5 with Next.js 16, `go test`, ESLint, `tsc`.

---

## Spec

`docs/superpowers/specs/2026-04-27-rename-items-translation-to-definition-design.md`

---

### Task 1: dx-api atomic rename — struct, JSON tag, assignments, test, AI prompt

**Files:**
- Modify: `dx-api/app/console/commands/import_courses_transform.go:91, 204, 223, 234`
- Modify: `dx-api/app/console/commands/import_courses_transform_test.go:157-158, 189-190`
- Modify: `dx-api/app/services/api/ai_custom_service.go:806, 824, 830, 831`

- [ ] **Step 1: Update test expectations to use `Definition` (RED phase)**

Edit `dx-api/app/console/commands/import_courses_transform_test.go`.

Replace L157–158:

```go
				if items[3].Translation != "句号" {
					t.Errorf("item[3].Translation = %q", items[3].Translation)
				}
```

with:

```go
				if items[3].Definition != "句号" {
					t.Errorf("item[3].Definition = %q", items[3].Definition)
				}
```

Replace L189–190:

```go
				if items[1].Translation != "逗号" {
					t.Errorf("item[1].Translation = %q", items[1].Translation)
				}
```

with:

```go
				if items[1].Definition != "逗号" {
					t.Errorf("item[1].Definition = %q", items[1].Definition)
				}
```

- [ ] **Step 2: Run the test, confirm it fails to compile (RED)**

```bash
cd dx-api && go test ./app/console/commands/... -run TestTransformItems
```

Expected: build error along the lines of `items[3].Definition undefined (type ItemEntry has no field or method Definition)`. This proves the test now demands the renamed field.

- [ ] **Step 3: Rename `ItemEntry.Translation` to `ItemEntry.Definition` (struct + JSON tag)**

Edit `dx-api/app/console/commands/import_courses_transform.go` L86–93. Replace:

```go
type ItemEntry struct {
	Position    int            `json:"position"`
	Item        string         `json:"item"`
	Phonetic    *PhoneticEntry `json:"phonetic"`
	Pos         *string        `json:"pos"`
	Translation string         `json:"translation"`
	Answer      bool           `json:"answer"`
}
```

with:

```go
type ItemEntry struct {
	Position   int            `json:"position"`
	Item       string         `json:"item"`
	Phonetic   *PhoneticEntry `json:"phonetic"`
	Pos        *string        `json:"pos"`
	Definition string         `json:"definition"`
	Answer     bool           `json:"answer"`
}
```

(Field alignment recomputed by `gofmt`/`goimports` — the longest name is now `Definition` at 10 chars.)

- [ ] **Step 4: Update the three struct-literal assignments inside `transformItems`**

In the same file `dx-api/app/console/commands/import_courses_transform.go`:

L201–207 (leading-punct branch). Replace:

```go
			items = append(items, ItemEntry{
				Position:    pos,
				Item:        string(r),
				Translation: punctTranslations[r],
				Answer:      false,
			})
```

with:

```go
			items = append(items, ItemEntry{
				Position:   pos,
				Item:       string(r),
				Definition: punctTranslations[r],
				Answer:     false,
			})
```

L215–225 (word branch). Replace:

```go
			items = append(items, ItemEntry{
				Position: pos,
				Item:     word,
				Phonetic: &PhoneticEntry{
					UK: wrapPhonetic(d.Phonetic.UK),
					US: wrapPhonetic(d.Phonetic.US),
				},
				Pos:         d.Pos,
				Translation: d.Definition,
				Answer:      true,
			})
```

with:

```go
			items = append(items, ItemEntry{
				Position: pos,
				Item:     word,
				Phonetic: &PhoneticEntry{
					UK: wrapPhonetic(d.Phonetic.UK),
					US: wrapPhonetic(d.Phonetic.US),
				},
				Pos:        d.Pos,
				Definition: d.Definition,
				Answer:     true,
			})
```

(The right-hand `d.Definition` is `WordDetail.Definition` from a different struct and stays as-is — only the field name on the left changes.)

L231–237 (trailing-punct branch). Replace:

```go
			items = append(items, ItemEntry{
				Position:    pos,
				Item:        string(r),
				Translation: punctTranslations[r],
				Answer:      false,
			})
```

with:

```go
			items = append(items, ItemEntry{
				Position:   pos,
				Item:       string(r),
				Definition: punctTranslations[r],
				Answer:     false,
			})
```

- [ ] **Step 5: Update the DeepSeek `genItemsPrompt` schema-spec line and example output**

Edit `dx-api/app/services/api/ai_custom_service.go`.

L806 — change the schema-spec bullet. Replace:

```
- translation: Chinese translation of the word — set to empty string for punctuation marks
```

with:

```
- definition: Chinese definition of the word — set to empty string for punctuation marks
```

L824 — change the first example output line. Replace:

```
      {"position": 1, "item": "I", "phonetic": {"uk": "/aɪ/", "us": "/aɪ/"}, "pos": "代词", "translation": "我", "answer": true}
```

with:

```
      {"position": 1, "item": "I", "phonetic": {"uk": "/aɪ/", "us": "/aɪ/"}, "pos": "代词", "definition": "我", "answer": true}
```

L830 — same replacement for the second example's first item:

```
      {"position": 1, "item": "I", "phonetic": {"uk": "/aɪ/", "us": "/aɪ/"}, "pos": "代词", "translation": "我", "answer": true},
```

becomes:

```
      {"position": 1, "item": "I", "phonetic": {"uk": "/aɪ/", "us": "/aɪ/"}, "pos": "代词", "definition": "我", "answer": true},
```

L831 — same replacement for the second example's second item:

```
      {"position": 2, "item": "like", "phonetic": {"uk": "/laɪk/", "us": "/laɪk/"}, "pos": "动词", "translation": "喜欢", "answer": true}
```

becomes:

```
      {"position": 2, "item": "like", "phonetic": {"uk": "/laɪk/", "us": "/laɪk/"}, "pos": "动词", "definition": "喜欢", "answer": true}
```

- [ ] **Step 6: Run formatting, vet, build, and tests — confirm GREEN**

```bash
cd dx-api && gofmt -w app/console/commands/import_courses_transform.go app/console/commands/import_courses_transform_test.go app/services/api/ai_custom_service.go
cd dx-api && go vet ./...
cd dx-api && go build ./...
cd dx-api && go test -race ./app/console/commands/...
```

Expected for each: clean exit (no diff from `gofmt -l`, no `vet` warnings, build succeeds, tests pass — including the renamed `TestTransformItems` cases and any race detection).

If `go vet ./...` flags anything in unrelated packages, address it; the rule is "no new lint issues", and the user has explicitly required no lint issues anywhere.

- [ ] **Step 7: Sanity-check no stragglers**

```bash
cd dx-api && grep -rn '"translation"' app/console/commands/import_courses_transform.go app/services/api/ai_custom_service.go
cd dx-api && grep -rn '\.Translation\b' app/console/commands/import_courses_transform.go app/console/commands/import_courses_transform_test.go
```

Expected output from both: empty (no remaining matches in these specific files). Other files mentioning `Translation` / `"translation"` are out-of-scope columns (per spec) and must be left alone.

- [ ] **Step 8: Commit dx-api atomic change**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/console/commands/import_courses_transform.go \
        dx-api/app/console/commands/import_courses_transform_test.go \
        dx-api/app/services/api/ai_custom_service.go
git commit -m "refactor(api): rename items[*].translation to definition

Per-token gloss inside content_items.items is renamed from translation
to definition to disambiguate from the parent-row content_items.translation
column. Atomic change covering struct + JSON tag + three transformItems
literals + transform test + genItemsPrompt schema and example output.
The shared genItemsPrompt is referenced by both ai_custom_service and
ai_custom_vocab_service, so both sentence and vocab paths are updated
in one stroke. No DB backfill (pre-launch DB)."
```

---

### Task 2: dx-web — rename `SpellingItem.translation` to `definition`

**Files:**
- Modify: `dx-web/src/features/web/play-core/types/spelling.ts:6`

- [ ] **Step 1: Rename the type field**

Edit `dx-web/src/features/web/play-core/types/spelling.ts`. Replace:

```ts
export type SpellingItem = {
  item: string;
  answer: boolean;
  pos: string | null;
  position: number;
  translation: string;
  phonetic: { uk: string; us: string } | null;
};
```

with:

```ts
export type SpellingItem = {
  item: string;
  answer: boolean;
  pos: string | null;
  position: number;
  definition: string;
  phonetic: { uk: string; us: string } | null;
};
```

- [ ] **Step 2: Verify no consumer reads the renamed field**

```bash
cd dx-web && grep -rn "SpellingItem" src/
cd dx-web && grep -rn "\.translation" src/features/web/play-core/components src/features/web/play-core/hooks
```

The `SpellingItem` references should compile; the `.translation` greps under play-core inner-items contexts should already be empty (every `.translation` access in play-core is on `currentItem` / parent-row content items, not on `SpellingItem`). If any new reader of `si.translation` exists, rename it to `si.definition` in this same step.

- [ ] **Step 3: Run lint and build**

```bash
cd dx-web && npm run lint
cd dx-web && npm run build
```

Expected: `npm run lint` exits 0 with no errors. `npm run build` produces a successful production build with no TS errors. If the build surfaces lint issues elsewhere not introduced by this change, fix them — the user requires no lint issues anywhere.

- [ ] **Step 4: Commit dx-web change**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/features/web/play-core/types/spelling.ts
git commit -m "refactor(web): rename SpellingItem.translation to definition

Aligns the inner-items type with the dx-api JSON shape (items[*].definition).
No component currently reads SpellingItem.translation; this is a type-only
change to enforce correctness for future readers."
```

---

### Task 3: Manual smoke test (post-implementation verification)

**Files:** none (runtime check only)

This task validates the full path: AI generation produces the new key, persistence keeps it, and the play UI still renders. Since dx-mini is intentionally out of scope and dx-web has no current reader, the play UI should be visually unchanged — we are looking for an *absence* of regression and the *presence* of the new key in the DB.

- [ ] **Step 1: Boot dx-api locally**

```bash
cd dx-api && air
```

Wait for `Listening on :3001` (or equivalent). If `air` is not installed, fall back to `cd dx-api && go run .`.

- [ ] **Step 2: Boot dx-web locally**

In a second terminal:

```bash
cd dx-web && npm run dev
```

Wait for `Ready` at `http://localhost:3000`.

- [ ] **Step 3: Trigger AI gen-items end-to-end**

In a browser:
1. Sign in as user `rainson` (the only admin per `middleware.AdminGuard()`).
2. Open the AI Custom admin panel (`/hall/ai-custom`).
3. Pick or create a course with at least one level whose content_metas have `is_break_done = true` and at least one content_item with `items IS NULL`. (If none exist, run the format/break flow first to seed one.)
4. Click "Generate items" (or the equivalent button that triggers the SSE endpoint backed by `processGenItems` in `ai_custom_service.go`).
5. Wait for the SSE stream to report `complete: true`.

- [ ] **Step 4: Inspect a generated row in PostgreSQL**

```bash
psql -d douxue -c "SELECT id, content, items FROM content_items WHERE items IS NOT NULL ORDER BY created_at DESC LIMIT 1;"
```

Expected: the `items` JSON contains element objects whose keys are exactly `position`, `item`, `phonetic`, `pos`, `definition`, `answer`. There must be **no** `translation` key inside any element. If there is, DeepSeek ignored the prompt update — re-check Task 1 Step 5 was committed.

- [ ] **Step 5: Verify play UI renders without regression**

Open a game level whose content has the newly-generated items. Walk through one or two questions in the spelling/vocab game:
- The Chinese gloss above the prompt area renders unchanged (it comes from the parent-row `content_item.translation` column, which we did not touch).
- Phonetics, POS pills, and word tokens all render correctly.

If anything regresses, it indicates an unanticipated consumer of `SpellingItem.translation` not caught by Task 2 Step 2 — file an issue and revert.

- [ ] **Step 6: Stop the dev servers**

`Ctrl+C` both terminals.

---

## Self-Review

**Spec coverage:**
- In-scope items 1–5 (Go struct, prompt, literals, test, TS type) → all covered by Task 1 + Task 2.
- D1 (skip migration) → respected; no Goravel migration created.
- D2 (rename Go ident) → Task 1 Step 3.
- D3 (defer dx-mini) → Task 2 Step 2 confirms no dx-mini change; spec already documents the deferred mismatch.
- Verification plan → Task 1 Step 6, Task 2 Step 3, Task 3 mirror the spec verification commands.

**Placeholder scan:** none.

**Type/identifier consistency:** `Definition` (Go field), `definition` (JSON tag, prompt text, TS field) used uniformly. `d.Definition` (RHS in transformItems) is the unrelated `WordDetail.Definition` field and is intentionally left as-is — called out in Task 1 Step 4 to avoid confusion.

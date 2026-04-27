# Rename `items[*].translation` → `items[*].definition`

**Date:** 2026-04-27
**Status:** Approved (design phase)

## Goal

Rename the per-token gloss key inside the `content_items.items` JSONB array from
`translation` to `definition`. Disambiguates from the parent-row
`content_items.translation` column (per-unit Chinese translation) and aligns with
the existing parent-row `content_items.definition` column.

## In scope

- Go: `ItemEntry.Translation` struct field + JSON tag → `ItemEntry.Definition` /
  `json:"definition"`.
- Go: every struct-literal assignment that builds an `ItemEntry`.
- Go: the `genItemsPrompt` text (schema-spec line + example output blocks).
  This prompt is shared by both sentence and vocab AI generation paths.
- Go: transform-test field accesses.
- TypeScript: `SpellingItem.translation` → `SpellingItem.definition`.

## Out of scope

| Concept | Storage | Why unchanged |
|---|---|---|
| `content_metas.translation` | DB column | Whole-source Chinese; unrelated layer. |
| `content_items.translation` | DB column | Per-broken-unit Chinese; unrelated layer. |
| dx-mini `play.ts` `{text, correct}` shape | TS code | Pre-existing unrelated mismatch — D3, deferred. |
| Production data backfill | DB rows | D1 = skip. DB is pre-launch; create-migrations have been edited in place (e.g. recent `speaker` column). |

## Files

### dx-api

- `app/console/commands/import_courses_transform.go`
  - L86–93: `ItemEntry.Translation string \`json:"translation"\`` →
    `ItemEntry.Definition string \`json:"definition"\``.
  - L204, L223, L234: three struct-literal assignments inside `transformItems`
    (`Translation: …` → `Definition: …`).
- `app/console/commands/import_courses_transform_test.go`
  - L157–158, L189–190: field accesses (`items[i].Translation` →
    `items[i].Definition`); error messages updated to match.
- `app/services/api/ai_custom_service.go`
  - L806: prompt schema-spec line `- translation: ...` → `- definition: ...`.
  - L824, L830, L831: example output JSON keys (`"translation"` →
    `"definition"`).
  - The same prompt is referenced by `ai_custom_vocab_service.go:529`; no
    additional vocab change needed.

### dx-web

- `src/features/web/play-core/types/spelling.ts`
  - L6: `translation: string` → `definition: string`.

No component currently reads `SpellingItem.translation` (`game-word-sentence.tsx`
reads `si.item / si.position / si.phonetic / si.pos / si.answer` only;
`use-vocab-battle.ts` reads `items[0]?.phonetic` only). The type change is
type-only — its purpose is to enforce correctness for future readers.

### dx-mini

No change. `pages/games/play/play.ts:93` parses `item.items` with a
`{text, correct}` shape and never accesses `.translation`. Verified.

## Decisions

- **D1 — backfill:** Skip migration. Pre-launch DB; developers re-run AI
  gen-items or wipe content tables to refresh local data.
- **D2 — Go identifier:** Rename both the field and the JSON tag. Avoids the
  visual mismatch `Translation string \`json:"definition"\``.
- **D3 — dx-mini play.ts shape:** Leave. Tracked separately.

## Risks & mitigations

- **AI prompt drift.** Prompt + struct + JSON tag must change atomically in one
  commit. If only the prompt changes, the Go parser silently drops the field.
  If only the struct changes, DeepSeek emits `translation` and writes empty
  `Definition`. Mitigation: single-commit change, verified by gen-items smoke
  test post-merge.
- **Stale dev DB rows.** Existing local rows still carry `translation` keys.
  Acceptable per D1; resolve by re-running gen-items or truncating
  `content_items` / `content_metas` in dev.

## Verification

1. `cd dx-api && go build ./... && go vet ./... && go test -race ./app/console/commands/...`
2. `cd dx-web && npm run lint && npm run build`
3. Manual smoke test: trigger AI gen-items on a fresh `content_meta` from the
   dx-web AI Custom admin UI; inspect the resulting row with
   `psql -c "SELECT items FROM content_items WHERE id = '…'"` and confirm each
   element carries `definition`, not `translation`. Confirm play UI renders
   unchanged.

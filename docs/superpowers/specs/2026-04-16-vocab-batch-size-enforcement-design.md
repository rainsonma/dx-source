# Vocab Batch-Size Enforcement

**Date:** 2026-04-16
**Status:** Approved

## Problem

vocab-match and vocab-elimination games batch content items during play (5 and 8 per batch respectively). Currently, nothing prevents a user from saving a non-multiple count (e.g. 7 items for vocab-match), which produces incomplete batches with broken layouts at play time:

- vocab-match: two-column layout expects 5 rows per batch
- vocab-elimination: 4x4 grid expects 8 items (16 tiles) per batch

vocab-battle is unaffected — one word per round, no batching.

## Design

### Approach: Enforce per-addition multiples + publish gate

Each addition must contain a complete batch (a multiple of the mode's batch size). A publish gate catches edge cases from deletion.

### Batch sizes

| Mode | Batch size | Valid totals (within MaxMetasPerLevel=20) |
|------|-----------|------------------------------------------|
| vocab-match | 5 | 5, 10, 15, 20 |
| vocab-elimination | 8 | 8, 16 |
| vocab-battle | 0 (none) | 1–20 |

### Backend changes

**`dx-api/app/consts/ai_custom_vocab.go`**

Add `VocabBatchSize(mode) int` — returns 5 for match, 8 for elimination, 0 for battle.

**`dx-api/app/services/api/course_content_service.go`** — `SaveMetadataBatch()`

After the existing `IsVocabMode` capacity check, add:

```
batchSize := consts.VocabBatchSize(game.Mode)
if batchSize > 0 && len(entries) % batchSize != 0 {
    return 0, ErrBatchSizeInvalid
}
```

New sentinel: `ErrBatchSizeInvalid` in `errors.go`, mapped to 400 in the controller.

**Publish validation** — In `PublishGame()` at `course_game_service.go`, inside the existing `for _, l := range levels` loop, after the `itemCount == 0` check. Reuses the already-queried `itemCount`:

```go
batchSize := consts.VocabBatchSize(game.Mode)
if batchSize > 0 && itemCount % int64(batchSize) != 0 {
    return fmt.Errorf("关卡「%s」词汇数量必须是 %d 的倍数（当前 %d 条）", l.Name, batchSize, itemCount)
}
```

### Frontend ai-custom changes

**`dx-web/src/features/web/ai-custom/helpers/vocab-format-metadata.ts`**

Add `vocabBatchSize(mode: GameMode): number` — returns 5/8/0 matching backend.

Update `parseVocabText()` signature to accept `batchSize: number`. Add check:

```ts
if (batchSize > 0 && pairs.length % batchSize !== 0) {
  return { ok: false, error: `词汇数量必须是 ${batchSize} 的倍数（当前 ${pairs.length} 条）` };
}
```

**`dx-web/src/features/web/ai-custom/components/add-vocab-dialog.tsx`**

- Compute `batchSize = vocabBatchSize(gameMode)` and pass to `parseVocabText()`
- Update hint text in manual tab to mention batch constraint

**`dx-web/src/features/web/ai-custom/components/level-units-panel.tsx`**

When `batchSize > 0 && metas.length % batchSize !== 0`, show an amber warning near the count stats: "数量需为 N 的倍数". Catches invalid state after deletion.

**AI generation path** — No changes. `VocabGenerateCount` already returns exact batch sizes (5/8/20).

### Frontend play-core changes — "下一组" button

**Hooks** (`use-vocab-match.ts` / `use-vocab-elimination.ts`):

- Expose `isLastBatch: batchEnd >= totalItems`
- Expose `isBatchComplete: matchedIndices.size === batchItems.length` (match) / `eliminatedIndices.size === batchItems.length` (elimination)
- Expose `nextBatch()` that clears the pending auto-advance timer and calls `advanceBatch()`
- Keep existing 600ms auto-advance behavior unchanged

**Components** (`game-vocab-match.tsx` / `game-vocab-elimination.tsx`):

Add a "下一组" button:
- **Visible** when `!isLastBatch` (hidden on last batch)
- **Disabled** when `!isBatchComplete` (enabled once all items matched/eliminated)
- Clicking calls `nextBatch()` (cancels pending timer, advances immediately)
- If the 600ms timer fires first, batch advances normally — button resets to disabled for the new batch
- Placement: bottom of card, right-aligned, small teal button with ArrowRight icon + "下一组" text

**No changes to `game-vocab-battle.tsx`.**

## Files affected

### Backend
- `dx-api/app/consts/ai_custom_vocab.go` — add `VocabBatchSize()`
- `dx-api/app/services/api/errors.go` — add `ErrBatchSizeInvalid`
- `dx-api/app/services/api/course_content_service.go` — batch-size validation in `SaveMetadataBatch()`
- `dx-api/app/services/api/course_game_service.go` — publish validation
- `dx-api/app/http/controllers/api/course_game_controller.go` — add `ErrBatchSizeInvalid` case in `mapCourseGameError()` → 400

### Frontend ai-custom
- `dx-web/src/features/web/ai-custom/helpers/vocab-format-metadata.ts` — add `vocabBatchSize()`, update `parseVocabText()`
- `dx-web/src/features/web/ai-custom/components/add-vocab-dialog.tsx` — pass batchSize, update hints
- `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx` — warning when count breaks multiple

### Frontend play-core
- `dx-web/src/features/web/play-core/hooks/use-vocab-match.ts` — expose `isLastBatch`, `isBatchComplete`, `nextBatch()`
- `dx-web/src/features/web/play-core/hooks/use-vocab-elimination.ts` — same
- `dx-web/src/features/web/play-core/components/game-vocab-match.tsx` — add "下一组" button
- `dx-web/src/features/web/play-core/components/game-vocab-elimination.tsx` — same

## Out of scope

- Changing AI generation counts (already correct)
- Changing vocab-battle behavior (no batching)
- Changing `MaxMetasPerLevel` constant
- Changing the 600ms auto-advance timing

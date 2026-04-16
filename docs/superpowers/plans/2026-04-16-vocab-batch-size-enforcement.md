# Vocab Batch-Size Enforcement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enforce that vocab-match levels contain multiples of 5 items and vocab-elimination levels contain multiples of 8 items, and add a "下一组" button to both game play pages.

**Architecture:** Backend validates batch-size multiples on save and publish. Frontend validates in the add dialog parser and shows warnings on the level editor. Play hooks expose batch state; play components render a next-batch button.

**Tech Stack:** Go/Goravel (backend), Next.js/React/TypeScript (frontend), Zustand (game store)

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `dx-api/app/consts/ai_custom_vocab.go` | Modify | Add `VocabBatchSize()` |
| `dx-api/app/services/api/errors.go` | Modify | Add `ErrBatchSizeInvalid` |
| `dx-api/app/services/api/course_content_service.go` | Modify | Batch-size check in `SaveMetadataBatch()` |
| `dx-api/app/services/api/course_game_service.go` | Modify | Batch-size check in `PublishGame()` |
| `dx-api/app/http/controllers/api/course_game_controller.go` | Modify | Map `ErrBatchSizeInvalid` in `mapCourseGameError()` |
| `dx-web/src/features/web/ai-custom/helpers/vocab-format-metadata.ts` | Modify | Add `vocabBatchSize()`, update `parseVocabText()` |
| `dx-web/src/features/web/ai-custom/components/add-vocab-dialog.tsx` | Modify | Pass `batchSize` to parser |
| `dx-web/src/features/web/ai-custom/components/vocab-manual-tab.tsx` | Modify | Show batch-size hint |
| `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx` | Modify | Amber warning when count breaks multiple |
| `dx-web/src/features/web/play-core/hooks/use-vocab-match.ts` | Modify | Expose `isLastBatch`, `isBatchComplete`, `nextBatch()` |
| `dx-web/src/features/web/play-core/hooks/use-vocab-elimination.ts` | Modify | Same |
| `dx-web/src/features/web/play-core/components/game-vocab-match.tsx` | Modify | Render "下一组" button |
| `dx-web/src/features/web/play-core/components/game-vocab-elimination.tsx` | Modify | Same |

---

### Task 1: Backend — Add `VocabBatchSize()` and `ErrBatchSizeInvalid`

**Files:**
- Modify: `dx-api/app/consts/ai_custom_vocab.go:25-29`
- Modify: `dx-api/app/services/api/errors.go:69`

- [ ] **Step 1: Add `VocabBatchSize()` to consts**

In `dx-api/app/consts/ai_custom_vocab.go`, add after `IsVocabMode()` (after line 29):

```go
// VocabBatchSize returns the required batch size for the given vocab mode.
// Returns 0 for modes with no batch constraint (vocab-battle).
func VocabBatchSize(mode string) int {
	switch mode {
	case GameModeVocabMatch:
		return VocabMatchCount
	case GameModeVocabElimination:
		return VocabEliminationCount
	default:
		return 0
	}
}
```

- [ ] **Step 2: Add `ErrBatchSizeInvalid` sentinel**

In `dx-api/app/services/api/errors.go`, add after `ErrGameNameTaken` (line 69):

```go
	ErrBatchSizeInvalid = errors.New("词汇数量不是批次大小的倍数")
```

- [ ] **Step 3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/consts/ai_custom_vocab.go dx-api/app/services/api/errors.go
git commit -m "feat(api): add VocabBatchSize() and ErrBatchSizeInvalid sentinel"
```

---

### Task 2: Backend — Enforce batch-size on save and publish

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go:168-171`
- Modify: `dx-api/app/services/api/course_game_service.go:359-369`
- Modify: `dx-api/app/http/controllers/api/course_game_controller.go:480-483`

- [ ] **Step 1: Add batch-size validation in `SaveMetadataBatch()`**

In `dx-api/app/services/api/course_content_service.go`, inside the `if consts.IsVocabMode(game.Mode)` block (line 168), add the batch-size check **after** the capacity check (after line 171):

```go
	if consts.IsVocabMode(game.Mode) {
		// Vocab modes: flat limit of MaxMetasPerLevel
		if len(existing)+len(entries) > consts.MaxMetasPerLevel {
			return 0, ErrCapacityExceeded
		}
		// Vocab modes: entries must be a multiple of the batch size
		batchSize := consts.VocabBatchSize(game.Mode)
		if batchSize > 0 && len(entries)%batchSize != 0 {
			return 0, ErrBatchSizeInvalid
		}
	} else {
```

This replaces the existing lines 168-171 plus the `} else {` on line 173.

- [ ] **Step 2: Add batch-size validation in `PublishGame()`**

In `dx-api/app/services/api/course_game_service.go`, inside the `for _, l := range levels` loop, after the `ungeneratedCount > 0` check (after line 381), add:

```go
		// Vocab modes: item count must be a multiple of the batch size
		batchSize := consts.VocabBatchSize(game.Mode)
		if batchSize > 0 && itemCount%int64(batchSize) != 0 {
			return fmt.Errorf("关卡「%s」词汇数量必须是 %d 的倍数（当前 %d 条）", l.Name, batchSize, itemCount)
		}
```

- [ ] **Step 3: Map `ErrBatchSizeInvalid` in controller**

In `dx-api/app/http/controllers/api/course_game_controller.go`, in `mapCourseGameError()`, add a new case after the `ErrCapacityExceeded` case (after line 481):

```go
	case errors.Is(err, services.ErrBatchSizeInvalid):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "词汇数量必须是批次大小的倍数")
```

- [ ] **Step 4: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add dx-api/app/services/api/course_content_service.go dx-api/app/services/api/course_game_service.go dx-api/app/http/controllers/api/course_game_controller.go
git commit -m "feat(api): enforce vocab batch-size multiples on save and publish"
```

---

### Task 3: Frontend — Update parser and add dialog

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/helpers/vocab-format-metadata.ts`
- Modify: `dx-web/src/features/web/ai-custom/components/add-vocab-dialog.tsx`
- Modify: `dx-web/src/features/web/ai-custom/components/vocab-manual-tab.tsx`

- [ ] **Step 1: Add `vocabBatchSize()` and update `parseVocabText()`**

In `dx-web/src/features/web/ai-custom/helpers/vocab-format-metadata.ts`:

Add `vocabBatchSize()` after `maxPairsForMode()` (after line 15):

```ts
export function vocabBatchSize(mode: GameMode): number {
  switch (mode) {
    case "vocab-match": return 5;
    case "vocab-elimination": return 8;
    default: return 0;
  }
}
```

Update the `parseVocabText` signature on line 30 to accept `batchSize`:

```ts
export function parseVocabText(raw: string, maxPairs: number, batchSize: number): ParseVocabResult {
```

Add the batch-size check after the existing `pairs.length > maxPairs` check (after line 68), before the final return:

```ts
  if (batchSize > 0 && pairs.length % batchSize !== 0) {
    return { ok: false, error: `词汇数量必须是 ${batchSize} 的倍数（当前 ${pairs.length} 条）` };
  }
```

- [ ] **Step 2: Pass `batchSize` in `add-vocab-dialog.tsx`**

In `dx-web/src/features/web/ai-custom/components/add-vocab-dialog.tsx`:

Update the import on line 28 to also import `vocabBatchSize`:

```ts
import { parseVocabText, maxPairsForMode, vocabBatchSize, MAX_METAS_PER_LEVEL } from "@/features/web/ai-custom/helpers/vocab-format-metadata";
```

Add after line 69 (`const maxPairs = maxPairsForMode(gameMode);`):

```ts
  const batchSize = vocabBatchSize(gameMode);
```

Update `handleSave()` on line 102 — change the `parseVocabText` call:

```ts
    const result = parseVocabText(manualText, maxPairs, batchSize);
```

- [ ] **Step 3: Add batch-size hint in `vocab-manual-tab.tsx`**

In `dx-web/src/features/web/ai-custom/components/vocab-manual-tab.tsx`:

Update the props type (line 10-15) to accept `batchSize`:

```ts
type VocabManualTabProps = {
  value: string;
  onChange: (value: string) => void;
  error?: string;
  maxPairs: number;
  batchSize: number;
};
```

Update the function signature (line 17):

```ts
export function VocabManualTab({ value, onChange, error, maxPairs, batchSize }: VocabManualTabProps) {
```

Update the hint text on line 24 to mention the batch constraint:

```ts
            请输入英文-中文词汇对，英文一行、中文释义下一行，依次交替。每次最多添加 {maxPairs} 对词汇{batchSize > 0 ? `，数量须为 ${batchSize} 的倍数` : ""}。
```

Then in `add-vocab-dialog.tsx`, update the `VocabManualTab` usage (around line 293-298) to pass `batchSize`:

```tsx
            <VocabManualTab
              value={manualText}
              onChange={(v) => { setManualText(v); setErrorMessage(""); setIsFromAi(false); }}
              error={errorMessage}
              maxPairs={maxPairs}
              batchSize={batchSize}
            />
```

- [ ] **Step 4: Verify lint**

Run: `cd dx-web && npx next lint`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add dx-web/src/features/web/ai-custom/helpers/vocab-format-metadata.ts dx-web/src/features/web/ai-custom/components/add-vocab-dialog.tsx dx-web/src/features/web/ai-custom/components/vocab-manual-tab.tsx
git commit -m "feat(web): enforce vocab batch-size multiples in add dialog"
```

---

### Task 4: Frontend — Level editor warning for invalid count

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx`

- [ ] **Step 1: Import `vocabBatchSize` and compute warning**

In `dx-web/src/features/web/ai-custom/components/level-units-panel.tsx`:

Update the import from `vocab-format-metadata` on line 80:

```ts
import { MAX_METAS_PER_LEVEL, vocabBatchSize } from "@/features/web/ai-custom/helpers/vocab-format-metadata";
```

Import the `TriangleAlert` icon — add to the lucide-react import on lines 24-37:

```ts
  TriangleAlert,
```

After `totalItemCount` (line 576), add:

```ts
  const levelBatchSize = vocabBatchSize(gameMode);
  const isBatchInvalid = isVocabMode && levelBatchSize > 0 && metas.length > 0 && metas.length % levelBatchSize !== 0;
```

- [ ] **Step 2: Add warning in stats bar**

In the stats bar `<div>` (line 662), after the closing `</span>` of the "练习单元总数" stat (line 670), add:

```tsx
              {isBatchInvalid && (
                <span className="flex items-center gap-1 text-amber-600">
                  <TriangleAlert className="h-3 w-3" />
                  数量需为 {levelBatchSize} 的倍数
                </span>
              )}
```

- [ ] **Step 3: Verify lint**

Run: `cd dx-web && npx next lint`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/ai-custom/components/level-units-panel.tsx
git commit -m "feat(web): show warning when vocab count breaks batch-size multiple"
```

---

### Task 5: Play hooks — Expose batch state

**Files:**
- Modify: `dx-web/src/features/web/play-core/hooks/use-vocab-match.ts`
- Modify: `dx-web/src/features/web/play-core/hooks/use-vocab-elimination.ts`

- [ ] **Step 1: Update `use-vocab-match.ts`**

In `dx-web/src/features/web/play-core/hooks/use-vocab-match.ts`:

Add state and derived values. After line 40 (`const wrongTimerRef = ...`), add:

```ts
  const advanceCalledRef = useRef(false);
```

Replace the `advanceBatch` callback (currently around lines 136-143) with:

```ts
  const advanceBatch = useCallback(() => {
    if (advanceCalledRef.current) return;
    advanceCalledRef.current = true;
    const nextStart = batchEnd;
    if (nextStart >= totalItems) {
      setPhase("result");
    } else {
      useGameStore.setState({ currentIndex: nextStart });
    }
  }, [batchEnd, totalItems, setPhase]);
```

Reset the ref in the existing `useEffect` that fires on `batchStart` change (lines 70-75). Add inside the effect body:

```ts
    advanceCalledRef.current = false;
```

Derive batch state values before the return statement (before line 189):

```ts
  const isBatchComplete = matchedIndices.size === batchItems.length && batchItems.length > 0;
  const isLastBatch = batchEnd >= totalItems;

  const nextBatch = useCallback(() => {
    if (advanceTimerRef.current) {
      clearTimeout(advanceTimerRef.current);
      advanceTimerRef.current = null;
    }
    advanceBatch();
  }, [advanceBatch]);
```

Update the return object (lines 189-199) to include the new values:

```ts
  return {
    batchItems,
    shuffledDefs,
    selectedWordIndex,
    matchedIndices,
    wrongPair,
    progress,
    combo,
    selectWord,
    selectDef,
    isBatchComplete,
    isLastBatch,
    nextBatch,
  };
```

- [ ] **Step 2: Update `use-vocab-elimination.ts`**

In `dx-web/src/features/web/play-core/hooks/use-vocab-elimination.ts`:

Add state. After line 49 (`const wrongTimerRef = ...`), add:

```ts
  const advanceCalledRef = useRef(false);
```

Replace the `advanceBatch` callback (currently around lines 152-159) with:

```ts
  const advanceBatch = useCallback(() => {
    if (advanceCalledRef.current) return;
    advanceCalledRef.current = true;
    const nextStart = batchEnd;
    if (nextStart >= totalItems) {
      setPhase("result");
    } else {
      useGameStore.setState({ currentIndex: nextStart });
    }
  }, [batchEnd, totalItems, setPhase]);
```

Reset the ref in the existing `useEffect` on `batchStart` change (lines 86-91). Add inside the effect body:

```ts
    advanceCalledRef.current = false;
```

Derive batch state values before the return statement (before line 215):

```ts
  const isBatchComplete = eliminatedIndices.size === batchItems.length && batchItems.length > 0;
  const isLastBatch = batchEnd >= totalItems;

  const nextBatch = useCallback(() => {
    if (advanceTimerRef.current) {
      clearTimeout(advanceTimerRef.current);
      advanceTimerRef.current = null;
    }
    advanceBatch();
  }, [advanceBatch]);
```

Update the return object (lines 215-225) to include the new values:

```ts
  return {
    gridRows,
    tiles,
    selectedTileId,
    eliminatedIndices,
    wrongPair,
    progress,
    combo,
    selectTile,
    isBatchComplete,
    isLastBatch,
    nextBatch,
  };
```

- [ ] **Step 3: Verify lint**

Run: `cd dx-web && npx next lint`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/play-core/hooks/use-vocab-match.ts dx-web/src/features/web/play-core/hooks/use-vocab-elimination.ts
git commit -m "feat(web): expose batch state in vocab play hooks"
```

---

### Task 6: Play components — Render "下一组" button

**Files:**
- Modify: `dx-web/src/features/web/play-core/components/game-vocab-match.tsx`
- Modify: `dx-web/src/features/web/play-core/components/game-vocab-elimination.tsx`

- [ ] **Step 1: Add button to `game-vocab-match.tsx`**

In `dx-web/src/features/web/play-core/components/game-vocab-match.tsx`:

Add `ArrowRight` to the lucide-react import (line 3):

```ts
import { Zap, Circle, CheckCircle2, ArrowRight } from "lucide-react";
```

Destructure the new values from `useVocabMatch()` (after line 16, add to the destructure):

```ts
    isBatchComplete,
    isLastBatch,
    nextBatch,
```

Replace the hint `<p>` at the bottom of the component (line 135-137) with:

```tsx
      {/* Footer */}
      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          点击左侧单词，再点击右侧匹配的释义
        </p>
        {!isLastBatch && (
          <button
            type="button"
            disabled={!isBatchComplete}
            onClick={nextBatch}
            className="flex items-center gap-1.5 rounded-lg bg-teal-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-teal-700 disabled:opacity-40"
          >
            下一组
            <ArrowRight className="h-4 w-4" />
          </button>
        )}
      </div>
```

- [ ] **Step 2: Add button to `game-vocab-elimination.tsx`**

In `dx-web/src/features/web/play-core/components/game-vocab-elimination.tsx`:

Add `ArrowRight` to the lucide-react import (line 3):

```ts
import { Sparkles, ArrowRight } from "lucide-react";
```

Destructure the new values from `useVocabElimination()` (after line 14, add to the destructure):

```ts
    isBatchComplete,
    isLastBatch,
    nextBatch,
```

Replace the hint `<p>` at the bottom of the component (line 90-92) with:

```tsx
      {/* Footer */}
      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          点击两个匹配的方块进行消除
        </p>
        {!isLastBatch && (
          <button
            type="button"
            disabled={!isBatchComplete}
            onClick={nextBatch}
            className="flex items-center gap-1.5 rounded-lg bg-teal-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-teal-700 disabled:opacity-40"
          >
            下一组
            <ArrowRight className="h-4 w-4" />
          </button>
        )}
      </div>
```

- [ ] **Step 3: Verify lint**

Run: `cd dx-web && npx next lint`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/play-core/components/game-vocab-match.tsx dx-web/src/features/web/play-core/components/game-vocab-elimination.tsx
git commit -m "feat(web): add next-batch button to vocab-match and vocab-elimination"
```

---

### Task 7: Verify end-to-end

- [ ] **Step 1: Verify backend compiles**

Run: `cd dx-api && go build ./...`
Expected: no errors

- [ ] **Step 2: Verify frontend lints**

Run: `cd dx-web && npx next lint`
Expected: no errors

- [ ] **Step 3: Verify frontend builds**

Run: `cd dx-web && npm run build`
Expected: no errors

- [ ] **Step 4: Final commit if any fixups needed**

If any lint or build issues were found and fixed in previous steps, commit the fixes:

```bash
git add -A
git commit -m "fix: lint and build fixes for vocab batch-size enforcement"
```

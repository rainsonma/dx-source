# Vocab Game Modes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire the three vocab game modes (vocab-match, vocab-elimination, vocab-battle) into the existing play system so they work across single, PK, and group shells.

**Architecture:** Each game mode gets a custom React hook (game logic) and a rewritten component (UI). Hooks follow the same contract as the existing `useWordSentence` — reading from `useGameStore` and calling `useGamePlayActions()`. No backend, store, or shell changes needed.

**Tech Stack:** React 19, TypeScript, Zustand, TailwindCSS v4, Lucide React icons

**Spec:** `docs/superpowers/specs/2026-04-08-vocab-game-modes-design.md`

---

## File Map

### New Files (3 hooks)

| File | Responsibility |
|------|----------------|
| `dx-web/src/features/web/play-core/hooks/use-vocab-match.ts` | Batch-mode hook: manages pair selection, match validation, shuffled definitions, batch advancement |
| `dx-web/src/features/web/play-core/hooks/use-vocab-elimination.ts` | Batch-mode hook: manages tile grid generation, pair elimination, batch advancement |
| `dx-web/src/features/web/play-core/hooks/use-vocab-battle.ts` | Per-item hook: manages letter keyboard, letter selection, shield state, opponent muting |

### Rewritten Files (3 components)

| File | Change |
|------|--------|
| `dx-web/src/features/web/play-core/components/game-vocab-match.tsx` | Replace static mockup with data-driven component consuming `useVocabMatch` |
| `dx-web/src/features/web/play-core/components/game-vocab-elimination.tsx` | Replace static mockup with data-driven component consuming `useVocabElimination` |
| `dx-web/src/features/web/play-core/components/game-vocab-battle.tsx` | Replace static mockup with data-driven component consuming `useVocabBattle` |

### Untouched Files

All shells, store, context, backend, word-sentence files — zero changes.

---

## Task 1: vocab-match hook

**Files:**
- Create: `dx-web/src/features/web/play-core/hooks/use-vocab-match.ts`

- [ ] **Step 1: Create the `useVocabMatch` hook**

```typescript
// dx-web/src/features/web/play-core/hooks/use-vocab-match.ts
"use client";

import { useState, useCallback, useEffect, useRef, useMemo } from "react";
import { useGameStore, type ContentItem } from "@/features/web/play-core/hooks/use-game-store";
import { useGamePlayActions } from "@/features/web/play-core/context/game-play-context";
import { getElapsedSeconds } from "@/features/web/play-core/hooks/use-game-timer";
import { SCORING } from "@/consts/scoring";

const BATCH_SIZE = 5;

/** Deterministic shuffle seeded by batch start index (Fisher-Yates) */
function shuffleArray<T>(arr: T[], seed: number): T[] {
  const result = [...arr];
  let s = seed;
  for (let i = result.length - 1; i > 0; i--) {
    s = (s * 16807 + 0) % 2147483647;
    const j = s % (i + 1);
    [result[i], result[j]] = [result[j], result[i]];
  }
  return result;
}

export function useVocabMatch() {
  const contentItems = useGameStore((s) => s.contentItems);
  const currentIndex = useGameStore((s) => s.currentIndex);
  const sessionId = useGameStore((s) => s.sessionId);
  const levelId = useGameStore((s) => s.levelId);
  const gameId = useGameStore((s) => s.gameId);
  const recordResult = useGameStore((s) => s.recordResult);
  const setPhase = useGameStore((s) => s.setPhase);
  const combo = useGameStore((s) => s.combo);

  const {
    recordAnswer: recordAnswerAction,
    markAsReview: markAsReviewAction,
  } = useGamePlayActions();

  const [selectedWordIndex, setSelectedWordIndex] = useState<number | null>(null);
  const [matchedIndices, setMatchedIndices] = useState<Set<number>>(new Set());
  const [wrongPair, setWrongPair] = useState<{ word: number; def: number } | null>(null);

  const wrongTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const advanceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const itemStartTimeRef = useRef<number>(Date.now());
  const reviewedIdsRef = useRef(new Set<string>());

  const totalItems = contentItems?.length ?? 0;

  // Current batch
  const batchStart = currentIndex;
  const batchEnd = Math.min(batchStart + BATCH_SIZE, totalItems);
  const batchItems: ContentItem[] = useMemo(
    () => contentItems?.slice(batchStart, batchEnd) ?? [],
    [contentItems, batchStart, batchEnd]
  );

  // Shuffled definitions for right column
  const shuffledDefs = useMemo(() => {
    const defs = batchItems.map((item, i) => ({
      batchIndex: i,
      translation: (item.translation as string) ?? "",
    }));
    return shuffleArray(defs, batchStart + 1);
  }, [batchItems, batchStart]);

  // Overall progress (items matched across all batches)
  const totalMatched = currentIndex + matchedIndices.size;

  const progress = {
    current: totalMatched,
    total: totalItems,
  };

  // Reset local state when batch changes
  useEffect(() => {
    setSelectedWordIndex(null);
    setMatchedIndices(new Set());
    setWrongPair(null);
    itemStartTimeRef.current = Date.now();
  }, [batchStart]);

  // Cleanup timers
  useEffect(() => {
    return () => {
      if (wrongTimerRef.current) clearTimeout(wrongTimerRef.current);
      if (advanceTimerRef.current) clearTimeout(advanceTimerRef.current);
    };
  }, []);

  /** Record an answer to the server (fire-and-forget) */
  const fireServerRecord = useCallback(
    (item: ContentItem, isCorrect: boolean, batchIdx: number) => {
      if (!sessionId || !levelId) return;

      const prevScore = useGameStore.getState().score;
      recordResult(isCorrect);

      const latestState = useGameStore.getState();
      const pointsEarned = latestState.score - prevScore;
      const baseScore = isCorrect ? SCORING.CORRECT_ANSWER : 0;
      const comboScore = pointsEarned - baseScore;

      // Next content item = next unmatched in batch or first of next batch
      const nextItemId =
        contentItems?.[batchStart + batchIdx + 1]?.id ??
        contentItems?.[batchEnd]?.id ??
        null;

      const duration = Math.round(
        (Date.now() - itemStartTimeRef.current) / 1000
      );

      recordAnswerAction({
        gameSessionId: sessionId,
        gameLevelId: levelId,
        contentItemId: item.id,
        isCorrect,
        userAnswer: isCorrect ? item.content : "",
        sourceAnswer: item.content,
        baseScore,
        comboScore,
        score: latestState.score,
        maxCombo: latestState.combo.maxCombo,
        playTime: getElapsedSeconds(),
        nextContentItemId: nextItemId,
        duration,
      });

      // Mark incorrect for review (deduped)
      if (!isCorrect && gameId && !reviewedIdsRef.current.has(item.id)) {
        reviewedIdsRef.current.add(item.id);
        markAsReviewAction({
          contentItemId: item.id,
          gameId,
          gameLevelId: levelId,
        });
      }

      itemStartTimeRef.current = Date.now();
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [sessionId, levelId, gameId, contentItems, batchStart, batchEnd, recordResult]
  );

  /** Advance to next batch or complete */
  const advanceBatch = useCallback(() => {
    const nextStart = batchEnd;
    if (nextStart >= totalItems) {
      setPhase("result");
    } else {
      useGameStore.setState({ currentIndex: nextStart });
    }
  }, [batchEnd, totalItems, setPhase]);

  /** Select an English word (left column) */
  const selectWord = useCallback(
    (batchIndex: number) => {
      if (matchedIndices.has(batchIndex)) return;
      if (wrongPair) return; // Ignore during wrong flash
      setSelectedWordIndex(batchIndex);
    },
    [matchedIndices, wrongPair]
  );

  /** Select a definition (right column) — check match */
  const selectDef = useCallback(
    (defBatchIndex: number) => {
      if (selectedWordIndex === null) return;
      if (matchedIndices.has(defBatchIndex)) return;
      if (wrongPair) return; // Ignore during wrong flash

      const item = batchItems[selectedWordIndex];
      if (!item) return;

      const isCorrect = defBatchIndex === selectedWordIndex;

      if (isCorrect) {
        const newMatched = new Set(matchedIndices);
        newMatched.add(selectedWordIndex);
        setMatchedIndices(newMatched);
        setSelectedWordIndex(null);

        fireServerRecord(item, true, selectedWordIndex);

        // Check if batch complete
        if (newMatched.size === batchItems.length) {
          advanceTimerRef.current = setTimeout(advanceBatch, 600);
        }
      } else {
        // Wrong match — flash for 500ms
        setWrongPair({ word: selectedWordIndex, def: defBatchIndex });
        fireServerRecord(item, false, selectedWordIndex);

        wrongTimerRef.current = setTimeout(() => {
          setWrongPair(null);
          setSelectedWordIndex(null);
        }, 500);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [selectedWordIndex, matchedIndices, batchItems, wrongPair, fireServerRecord, advanceBatch]
  );

  return {
    // Batch data
    batchItems,
    shuffledDefs,

    // Selection state
    selectedWordIndex,
    matchedIndices,
    wrongPair,

    // Progress
    progress,
    combo,

    // Actions
    selectWord,
    selectDef,
  };
}
```

- [ ] **Step 2: Verify lint passes**

Run: `cd dx-web && npx eslint src/features/web/play-core/hooks/use-vocab-match.ts --no-error-on-unmatched-pattern`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/play-core/hooks/use-vocab-match.ts
git commit -m "feat: add useVocabMatch hook for batch pair-matching logic"
```

---

## Task 2: vocab-match component

**Files:**
- Rewrite: `dx-web/src/features/web/play-core/components/game-vocab-match.tsx`

- [ ] **Step 1: Rewrite `game-vocab-match.tsx` with data-driven UI**

```tsx
// dx-web/src/features/web/play-core/components/game-vocab-match.tsx
"use client";

import { Zap, Circle, CheckCircle2 } from "lucide-react";
import { useVocabMatch } from "@/features/web/play-core/hooks/use-vocab-match";

export function GameVocabMatch() {
  const {
    batchItems,
    shuffledDefs,
    selectedWordIndex,
    matchedIndices,
    wrongPair,
    progress,
    combo,
    selectWord,
    selectDef,
  } = useVocabMatch();

  if (batchItems.length === 0) return null;

  return (
    <div className="flex w-full max-w-3xl flex-col gap-7 rounded-[20px] border border-border bg-card p-6 shadow-sm md:p-8">
      {/* Progress */}
      <div className="flex flex-col gap-2.5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-foreground">
              进度 {progress.current}/{progress.total}
            </span>
          </div>
          {combo.streak >= 3 && (
            <div className="flex items-center gap-1.5 rounded-lg bg-teal-600/10 px-3 py-1">
              <Zap className="h-3.5 w-3.5 text-teal-600" />
              <span className="text-xs font-bold text-teal-600">
                连击 &times;{combo.streak}
              </span>
            </div>
          )}
        </div>
        <div className="h-1.5 w-full rounded-full bg-border">
          <div
            className="h-1.5 rounded-full bg-gradient-to-r from-blue-500 to-teal-500 transition-all duration-300"
            style={{
              width: `${(progress.current / Math.max(progress.total, 1)) * 100}%`,
            }}
          />
        </div>
      </div>

      {/* Match area */}
      <div className="flex flex-col gap-4 sm:flex-row sm:gap-6">
        {/* English words */}
        <div className="flex flex-1 flex-col gap-2.5">
          <span className="text-xs font-semibold text-muted-foreground">
            英文单词
          </span>
          {batchItems.map((item, i) => {
            const isMatched = matchedIndices.has(i);
            const isSelected = selectedWordIndex === i;
            const isWrong = wrongPair?.word === i;
            return (
              <button
                key={item.id}
                type="button"
                disabled={isMatched}
                onClick={() => selectWord(i)}
                className={`flex items-center gap-2.5 rounded-xl border px-4 py-3 transition-colors ${
                  isMatched
                    ? "border-emerald-300 bg-emerald-50"
                    : isWrong
                      ? "animate-[shake_0.3s_ease-in-out] border-red-400 bg-red-50"
                      : isSelected
                        ? "border-blue-400 bg-blue-50"
                        : "border-border bg-card hover:bg-muted/50"
                }`}
              >
                {isMatched ? (
                  <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                ) : (
                  <Circle
                    className={`h-4 w-4 ${isSelected ? "text-blue-400" : "text-slate-300"}`}
                  />
                )}
                <span
                  className={`text-sm font-medium ${
                    isMatched
                      ? "text-emerald-600"
                      : isSelected
                        ? "text-blue-600"
                        : "text-foreground"
                  }`}
                >
                  {item.content}
                </span>
              </button>
            );
          })}
        </div>

        {/* Chinese definitions */}
        <div className="flex flex-1 flex-col gap-2.5">
          <span className="text-xs font-semibold text-muted-foreground">
            中文释义
          </span>
          {shuffledDefs.map((def) => {
            const isMatched = matchedIndices.has(def.batchIndex);
            const isWrong = wrongPair?.def === def.batchIndex;
            return (
              <button
                key={`def-${def.batchIndex}`}
                type="button"
                disabled={isMatched}
                onClick={() => selectDef(def.batchIndex)}
                className={`flex items-center justify-center rounded-xl border px-4 py-3 transition-colors ${
                  isMatched
                    ? "border-emerald-300 bg-emerald-50"
                    : isWrong
                      ? "animate-[shake_0.3s_ease-in-out] border-red-400 bg-red-50"
                      : "border-border bg-card hover:bg-muted/50"
                }`}
              >
                <span
                  className={`text-sm font-medium ${
                    isMatched ? "text-emerald-600" : "text-foreground"
                  }`}
                >
                  {def.translation}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      <p className="text-center text-xs text-muted-foreground">
        点击左侧单词，再点击右侧匹配的释义
      </p>
    </div>
  );
}
```

- [ ] **Step 2: Verify lint passes**

Run: `cd dx-web && npx eslint src/features/web/play-core/components/game-vocab-match.tsx --no-error-on-unmatched-pattern`
Expected: No errors

- [ ] **Step 3: Verify build passes**

Run: `cd dx-web && npx next build 2>&1 | tail -20`
Expected: Build succeeds with no type errors

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/play-core/components/game-vocab-match.tsx
git commit -m "feat: implement vocab-match game component with pair-matching UI"
```

---

## Task 3: vocab-elimination hook

**Files:**
- Create: `dx-web/src/features/web/play-core/hooks/use-vocab-elimination.ts`

- [ ] **Step 1: Create the `useVocabElimination` hook**

```typescript
// dx-web/src/features/web/play-core/hooks/use-vocab-elimination.ts
"use client";

import { useState, useCallback, useEffect, useRef, useMemo } from "react";
import { useGameStore, type ContentItem } from "@/features/web/play-core/hooks/use-game-store";
import { useGamePlayActions } from "@/features/web/play-core/context/game-play-context";
import { getElapsedSeconds } from "@/features/web/play-core/hooks/use-game-timer";
import { SCORING } from "@/consts/scoring";

const BATCH_SIZE = 8;
const COLUMNS = 4;

export type Tile = {
  id: string;
  type: "en" | "zh";
  text: string;
  /** Index into batchItems — tiles with same itemIndex are a pair */
  itemIndex: number;
};

/** Deterministic shuffle seeded by batch start index */
function shuffleTiles(tiles: Tile[], seed: number): Tile[] {
  const result = [...tiles];
  let s = seed;
  for (let i = result.length - 1; i > 0; i--) {
    s = (s * 16807 + 0) % 2147483647;
    const j = s % (i + 1);
    [result[i], result[j]] = [result[j], result[i]];
  }
  return result;
}

export function useVocabElimination() {
  const contentItems = useGameStore((s) => s.contentItems);
  const currentIndex = useGameStore((s) => s.currentIndex);
  const sessionId = useGameStore((s) => s.sessionId);
  const levelId = useGameStore((s) => s.levelId);
  const gameId = useGameStore((s) => s.gameId);
  const recordResult = useGameStore((s) => s.recordResult);
  const setPhase = useGameStore((s) => s.setPhase);
  const combo = useGameStore((s) => s.combo);

  const {
    recordAnswer: recordAnswerAction,
    markAsReview: markAsReviewAction,
  } = useGamePlayActions();

  const [selectedTileId, setSelectedTileId] = useState<string | null>(null);
  const [eliminatedIndices, setEliminatedIndices] = useState<Set<number>>(new Set());
  const [wrongPair, setWrongPair] = useState<{ t1: string; t2: string } | null>(null);

  const wrongTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const advanceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const itemStartTimeRef = useRef<number>(Date.now());
  const reviewedIdsRef = useRef(new Set<string>());

  const totalItems = contentItems?.length ?? 0;

  // Current batch
  const batchStart = currentIndex;
  const batchEnd = Math.min(batchStart + BATCH_SIZE, totalItems);
  const batchItems: ContentItem[] = useMemo(
    () => contentItems?.slice(batchStart, batchEnd) ?? [],
    [contentItems, batchStart, batchEnd]
  );

  // Generate shuffled tile grid
  const tiles = useMemo(() => {
    const raw: Tile[] = [];
    batchItems.forEach((item, i) => {
      raw.push({ id: `en-${i}`, type: "en", text: item.content, itemIndex: i });
      raw.push({ id: `zh-${i}`, type: "zh", text: (item.translation as string) ?? "", itemIndex: i });
    });
    return shuffleTiles(raw, batchStart + 1);
  }, [batchItems, batchStart]);

  // Grid rows for rendering
  const gridRows = useMemo(() => {
    const rows: Tile[][] = [];
    for (let i = 0; i < tiles.length; i += COLUMNS) {
      rows.push(tiles.slice(i, i + COLUMNS));
    }
    return rows;
  }, [tiles]);

  // Progress
  const totalEliminated = currentIndex + eliminatedIndices.size;
  const progress = {
    current: totalEliminated,
    total: totalItems,
  };

  // Reset local state when batch changes
  useEffect(() => {
    setSelectedTileId(null);
    setEliminatedIndices(new Set());
    setWrongPair(null);
    itemStartTimeRef.current = Date.now();
  }, [batchStart]);

  // Cleanup timers
  useEffect(() => {
    return () => {
      if (wrongTimerRef.current) clearTimeout(wrongTimerRef.current);
      if (advanceTimerRef.current) clearTimeout(advanceTimerRef.current);
    };
  }, []);

  /** Record answer to server (fire-and-forget) */
  const fireServerRecord = useCallback(
    (item: ContentItem, isCorrect: boolean, batchIdx: number) => {
      if (!sessionId || !levelId) return;

      const prevScore = useGameStore.getState().score;
      recordResult(isCorrect);

      const latestState = useGameStore.getState();
      const pointsEarned = latestState.score - prevScore;
      const baseScore = isCorrect ? SCORING.CORRECT_ANSWER : 0;
      const comboScore = pointsEarned - baseScore;

      const nextItemId =
        contentItems?.[batchStart + batchIdx + 1]?.id ??
        contentItems?.[batchEnd]?.id ??
        null;

      const duration = Math.round(
        (Date.now() - itemStartTimeRef.current) / 1000
      );

      recordAnswerAction({
        gameSessionId: sessionId,
        gameLevelId: levelId,
        contentItemId: item.id,
        isCorrect,
        userAnswer: isCorrect ? item.content : "",
        sourceAnswer: item.content,
        baseScore,
        comboScore,
        score: latestState.score,
        maxCombo: latestState.combo.maxCombo,
        playTime: getElapsedSeconds(),
        nextContentItemId: nextItemId,
        duration,
      });

      if (!isCorrect && gameId && !reviewedIdsRef.current.has(item.id)) {
        reviewedIdsRef.current.add(item.id);
        markAsReviewAction({
          contentItemId: item.id,
          gameId,
          gameLevelId: levelId,
        });
      }

      itemStartTimeRef.current = Date.now();
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [sessionId, levelId, gameId, contentItems, batchStart, batchEnd, recordResult]
  );

  /** Advance to next batch or complete */
  const advanceBatch = useCallback(() => {
    const nextStart = batchEnd;
    if (nextStart >= totalItems) {
      setPhase("result");
    } else {
      useGameStore.setState({ currentIndex: nextStart });
    }
  }, [batchEnd, totalItems, setPhase]);

  /** Handle tile click */
  const selectTile = useCallback(
    (tileId: string) => {
      if (wrongPair) return; // Ignore during wrong flash

      const tile = tiles.find((t) => t.id === tileId);
      if (!tile) return;
      if (eliminatedIndices.has(tile.itemIndex)) return;

      // Deselect if clicking the same tile
      if (selectedTileId === tileId) {
        setSelectedTileId(null);
        return;
      }

      // First selection
      if (selectedTileId === null) {
        setSelectedTileId(tileId);
        return;
      }

      // Second selection — check match
      const firstTile = tiles.find((t) => t.id === selectedTileId);
      if (!firstTile) {
        setSelectedTileId(tileId);
        return;
      }

      const item = batchItems[firstTile.itemIndex];
      if (!item) return;

      const isMatch =
        firstTile.itemIndex === tile.itemIndex && firstTile.type !== tile.type;

      if (isMatch) {
        const newEliminated = new Set(eliminatedIndices);
        newEliminated.add(firstTile.itemIndex);
        setEliminatedIndices(newEliminated);
        setSelectedTileId(null);

        fireServerRecord(item, true, firstTile.itemIndex);

        // Check if batch complete
        if (newEliminated.size === batchItems.length) {
          advanceTimerRef.current = setTimeout(advanceBatch, 600);
        }
      } else {
        // Wrong pair — flash for 500ms
        setWrongPair({ t1: selectedTileId, t2: tileId });
        fireServerRecord(item, false, firstTile.itemIndex);

        wrongTimerRef.current = setTimeout(() => {
          setWrongPair(null);
          setSelectedTileId(null);
        }, 500);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [selectedTileId, eliminatedIndices, tiles, batchItems, wrongPair, fireServerRecord, advanceBatch]
  );

  return {
    // Grid
    gridRows,
    tiles,

    // Selection
    selectedTileId,
    eliminatedIndices,
    wrongPair,

    // Progress
    progress,
    combo,

    // Actions
    selectTile,
  };
}
```

- [ ] **Step 2: Verify lint passes**

Run: `cd dx-web && npx eslint src/features/web/play-core/hooks/use-vocab-elimination.ts --no-error-on-unmatched-pattern`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/play-core/hooks/use-vocab-elimination.ts
git commit -m "feat: add useVocabElimination hook for batch grid elimination logic"
```

---

## Task 4: vocab-elimination component

**Files:**
- Rewrite: `dx-web/src/features/web/play-core/components/game-vocab-elimination.tsx`

- [ ] **Step 1: Rewrite `game-vocab-elimination.tsx` with data-driven UI**

```tsx
// dx-web/src/features/web/play-core/components/game-vocab-elimination.tsx
"use client";

import { Sparkles } from "lucide-react";
import { useVocabElimination } from "@/features/web/play-core/hooks/use-vocab-elimination";

export function GameVocabElimination() {
  const {
    gridRows,
    selectedTileId,
    eliminatedIndices,
    wrongPair,
    progress,
    combo,
    selectTile,
  } = useVocabElimination();

  if (gridRows.length === 0) return null;

  return (
    <div className="flex w-full max-w-[700px] flex-col items-center gap-5">
      {/* Status row */}
      <div className="flex w-full flex-col items-center gap-3 sm:flex-row sm:justify-between">
        <span className="text-sm font-semibold text-foreground">
          已消除 {progress.current}/{progress.total} 对
        </span>
        <div className="h-1.5 w-full max-w-[300px] rounded-full bg-border">
          <div
            className="h-1.5 rounded-full bg-gradient-to-r from-pink-500 to-teal-500 transition-all duration-300"
            style={{
              width: `${(progress.current / Math.max(progress.total, 1)) * 100}%`,
            }}
          />
        </div>
        {combo.streak >= 2 && (
          <div className="flex items-center gap-1.5 rounded-lg border border-pink-500/20 bg-pink-50 px-3 py-1">
            <Sparkles className="h-3.5 w-3.5 text-pink-500" />
            <span className="text-xs font-bold text-pink-500">
              连击 &times;{combo.streak}
            </span>
          </div>
        )}
      </div>

      {/* Grid */}
      <div className="flex w-full flex-col gap-2.5 rounded-[20px] border border-border bg-card p-4 shadow-sm md:p-6">
        {gridRows.map((row, ri) => (
          <div key={ri} className="flex gap-2.5">
            {row.map((tile) => {
              const isEliminated = eliminatedIndices.has(tile.itemIndex);
              const isSelected = selectedTileId === tile.id;
              const isWrong =
                wrongPair?.t1 === tile.id || wrongPair?.t2 === tile.id;

              return (
                <button
                  key={tile.id}
                  type="button"
                  disabled={isEliminated}
                  onClick={() => selectTile(tile.id)}
                  className={`flex h-14 flex-1 items-center justify-center rounded-xl border transition-all md:h-16 ${
                    isEliminated
                      ? "border-border/20 bg-muted opacity-40"
                      : isWrong
                        ? "animate-[shake_0.3s_ease-in-out] border-2 border-red-400 bg-red-50"
                        : isSelected
                          ? "border-2 border-pink-500 bg-pink-50"
                          : "border-[1.5px] border-border bg-card hover:bg-muted/50"
                  }`}
                >
                  <span
                    className={`text-sm font-medium ${
                      isEliminated
                        ? "text-muted-foreground line-through"
                        : isWrong
                          ? "text-red-600"
                          : isSelected
                            ? "text-pink-600"
                            : "text-foreground"
                    }`}
                  >
                    {tile.text}
                  </span>
                </button>
              );
            })}
          </div>
        ))}
      </div>

      <p className="text-xs text-muted-foreground">
        点击两个匹配的方块进行消除
      </p>
    </div>
  );
}
```

- [ ] **Step 2: Verify lint passes**

Run: `cd dx-web && npx eslint src/features/web/play-core/components/game-vocab-elimination.tsx --no-error-on-unmatched-pattern`
Expected: No errors

- [ ] **Step 3: Verify build passes**

Run: `cd dx-web && npx next build 2>&1 | tail -20`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/play-core/components/game-vocab-elimination.tsx
git commit -m "feat: implement vocab-elimination game component with grid UI"
```

---

## Task 5: vocab-battle hook

**Files:**
- Create: `dx-web/src/features/web/play-core/hooks/use-vocab-battle.ts`

- [ ] **Step 1: Create the `useVocabBattle` hook**

```typescript
// dx-web/src/features/web/play-core/hooks/use-vocab-battle.ts
"use client";

import { useState, useCallback, useEffect, useRef, useMemo } from "react";
import { useGameStore } from "@/features/web/play-core/hooks/use-game-store";
import { useGamePlayActions } from "@/features/web/play-core/context/game-play-context";
import { getElapsedSeconds } from "@/features/web/play-core/hooks/use-game-timer";
import { SCORING } from "@/consts/scoring";

const SHIELD_COUNT = 5;
const MIN_KEYBOARD_SIZE = 6;

/** Shuffle an array with a seed */
function seededShuffle<T>(arr: T[], seed: number): T[] {
  const result = [...arr];
  let s = seed;
  for (let i = result.length - 1; i > 0; i--) {
    s = (s * 16807 + 0) % 2147483647;
    const j = s % (i + 1);
    [result[i], result[j]] = [result[j], result[i]];
  }
  return result;
}

/** Generate keyboard letters: word letters + random distractors, shuffled */
function buildKeyboard(word: string, seed: number): string[] {
  const letters = word.toUpperCase().split("");
  // Add distractors for short words
  const distractors = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
  let s = seed;
  while (letters.length < MIN_KEYBOARD_SIZE) {
    s = (s * 16807 + 0) % 2147483647;
    const ch = distractors[s % distractors.length];
    // Avoid adding a letter already present (makes it too easy)
    if (!letters.includes(ch)) {
      letters.push(ch);
    }
  }
  return seededShuffle(letters, seed + 7);
}

export function useVocabBattle() {
  const contentItems = useGameStore((s) => s.contentItems);
  const currentIndex = useGameStore((s) => s.currentIndex);
  const sessionId = useGameStore((s) => s.sessionId);
  const levelId = useGameStore((s) => s.levelId);
  const gameId = useGameStore((s) => s.gameId);
  const recordResult = useGameStore((s) => s.recordResult);
  const recordSkipStore = useGameStore((s) => s.recordSkip);
  const nextItem = useGameStore((s) => s.nextItem);
  const setPhase = useGameStore((s) => s.setPhase);
  const combo = useGameStore((s) => s.combo);

  const {
    recordAnswer: recordAnswerAction,
    recordSkip: recordSkipAction,
    markAsReview: markAsReviewAction,
    competitive,
  } = useGamePlayActions();

  const [filledLetters, setFilledLetters] = useState<string[]>([]);
  const [usedKeyIndices, setUsedKeyIndices] = useState<Set<number>>(new Set());
  const [hasError, setHasError] = useState(false);
  const [isRevealed, setIsRevealed] = useState(false);
  const [playerShields, setPlayerShields] = useState<boolean[]>([]);
  const [opponentShields, setOpponentShields] = useState<boolean[]>([]);
  const [opponentFilledCount, setOpponentFilledCount] = useState(0);

  const hadWrongAttemptRef = useRef(false);
  const errorTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const isProcessingRef = useRef(false);
  const itemStartTimeRef = useRef<number>(Date.now());
  const reviewedIdsRef = useRef(new Set<string>());

  const totalItems = contentItems?.length ?? 0;
  const currentItem = contentItems?.[currentIndex] ?? null;
  const targetWord = (currentItem?.content ?? "").toUpperCase();
  const translation = (currentItem?.translation as string) ?? "";

  // Extract phonetic from items array
  const phonetic = useMemo(() => {
    const raw = currentItem?.items;
    const items = Array.isArray(raw)
      ? raw
      : typeof raw === "string"
        ? (() => { try { return JSON.parse(raw); } catch { return []; } })()
        : [];
    return items[0]?.phonetic ?? null;
  }, [currentItem]);

  // Keyboard letters for current word
  const keyboardLetters = useMemo(
    () => (targetWord ? buildKeyboard(targetWord, currentIndex + 1) : []),
    [targetWord, currentIndex]
  );

  // Letter slots for display
  const letterSlots = useMemo(() => {
    return targetWord.split("").map((letter, i) => ({
      letter,
      filled: i < filledLetters.length,
      filledLetter: filledLetters[i] ?? null,
    }));
  }, [targetWord, filledLetters]);

  // Opponent letter slots (for competitive display)
  const opponentSlots = useMemo(() => {
    return targetWord.split("").map((letter, i) => ({
      letter,
      filled: i < opponentFilledCount,
    }));
  }, [targetWord, opponentFilledCount]);

  const progress = {
    current: currentIndex + 1,
    total: totalItems,
  };

  // Reset state on item change
  useEffect(() => {
    setFilledLetters([]);
    setUsedKeyIndices(new Set());
    setHasError(false);
    setIsRevealed(false);
    setPlayerShields(Array(SHIELD_COUNT).fill(true));
    setOpponentShields(Array(SHIELD_COUNT).fill(true));
    setOpponentFilledCount(0);
    hadWrongAttemptRef.current = false;
    isProcessingRef.current = false;
    itemStartTimeRef.current = Date.now();
  }, [currentIndex]);

  // Cleanup timers
  useEffect(() => {
    return () => {
      if (errorTimerRef.current) clearTimeout(errorTimerRef.current);
    };
  }, []);

  /** Fire answer record to server */
  const fireServerRecord = useCallback(
    (isCorrect: boolean) => {
      if (!sessionId || !levelId || !currentItem) return;

      const prevScore = useGameStore.getState().score;
      recordResult(isCorrect);

      const latestState = useGameStore.getState();
      const pointsEarned = latestState.score - prevScore;
      const baseScore = isCorrect ? SCORING.CORRECT_ANSWER : 0;
      const comboScore = pointsEarned - baseScore;

      const nextItemId = contentItems?.[currentIndex + 1]?.id ?? null;
      const duration = Math.round(
        (Date.now() - itemStartTimeRef.current) / 1000
      );

      recordAnswerAction({
        gameSessionId: sessionId,
        gameLevelId: levelId,
        contentItemId: currentItem.id,
        isCorrect,
        userAnswer: filledLetters.join(""),
        sourceAnswer: currentItem.content,
        baseScore,
        comboScore,
        score: latestState.score,
        maxCombo: latestState.combo.maxCombo,
        playTime: getElapsedSeconds(),
        nextContentItemId: nextItemId,
        duration,
      });

      if (!isCorrect && gameId && !reviewedIdsRef.current.has(currentItem.id)) {
        reviewedIdsRef.current.add(currentItem.id);
        markAsReviewAction({
          contentItemId: currentItem.id,
          gameId,
          gameLevelId: levelId,
        });
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [sessionId, levelId, gameId, currentItem, contentItems, currentIndex, filledLetters, recordResult]
  );

  /** Handle keyboard letter click */
  const pressLetter = useCallback(
    (keyIndex: number) => {
      if (isRevealed) return;
      if (usedKeyIndices.has(keyIndex)) return;

      const letter = keyboardLetters[keyIndex];
      if (!letter) return;

      const nextPos = filledLetters.length;
      const expectedLetter = targetWord[nextPos];

      if (letter === expectedLetter) {
        // Correct letter
        const newFilled = [...filledLetters, letter];
        setFilledLetters(newFilled);
        setUsedKeyIndices(new Set([...usedKeyIndices, keyIndex]));

        // Remove one opponent shield (visual feedback)
        setOpponentShields((prev) => {
          const idx = prev.lastIndexOf(true);
          if (idx === -1) return prev;
          const next = [...prev];
          next[idx] = false;
          return next;
        });

        // Clear error state
        if (hasError) setHasError(false);

        // Check if word complete
        if (newFilled.length === targetWord.length) {
          const isItemCorrect = !hadWrongAttemptRef.current;
          setIsRevealed(true);
          fireServerRecord(isItemCorrect);
        }
      } else {
        // Wrong letter
        hadWrongAttemptRef.current = true;
        setHasError(true);
        if (errorTimerRef.current) clearTimeout(errorTimerRef.current);
        errorTimerRef.current = setTimeout(() => setHasError(false), 400);

        // Remove one player shield (visual feedback)
        setPlayerShields((prev) => {
          const idx = prev.lastIndexOf(true);
          if (idx === -1) return prev;
          const next = [...prev];
          next[idx] = false;
          return next;
        });
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [isRevealed, usedKeyIndices, keyboardLetters, filledLetters, targetWord, hasError, fireServerRecord]
  );

  /** Advance to next word after reveal */
  const advanceAfterReveal = useCallback(() => {
    if (isProcessingRef.current) return;
    isProcessingRef.current = true;

    if (currentIndex + 1 >= totalItems) {
      setPhase("result");
    } else {
      nextItem();
    }
  }, [currentIndex, totalItems, setPhase, nextItem]);

  /** Skip current word */
  const skipItem = useCallback(() => {
    if (!currentItem) return;

    if (isRevealed) {
      advanceAfterReveal();
      return;
    }

    if (isProcessingRef.current) return;
    isProcessingRef.current = true;

    recordSkipStore();

    if (sessionId && levelId) {
      const nextItemId = contentItems?.[currentIndex + 1]?.id ?? null;
      recordSkipAction({
        gameSessionId: sessionId,
        gameLevelId: levelId,
        playTime: getElapsedSeconds(),
        nextContentItemId: nextItemId,
      });
    }

    setIsRevealed(true);
    isProcessingRef.current = false;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentItem, isRevealed, recordSkipStore, sessionId, levelId, contentItems, currentIndex, advanceAfterReveal]);

  // Keyboard shortcut: Enter/Space to advance after reveal
  useEffect(() => {
    if (!isRevealed) return;
    const handler = (e: KeyboardEvent) => {
      if (e.repeat) return;
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        advanceAfterReveal();
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [isRevealed, advanceAfterReveal]);

  return {
    // Current item
    targetWord,
    translation,
    phonetic,

    // Player state
    letterSlots,
    keyboardLetters,
    usedKeyIndices,
    filledLetters,
    hasError,
    isRevealed,
    playerShields,
    opponentShields,

    // Opponent (competitive)
    opponentSlots,
    opponentFilledCount,
    competitive: competitive ?? false,

    // Progress
    progress,
    combo,

    // Actions
    pressLetter,
    advanceAfterReveal,
    skipItem,
  };
}
```

- [ ] **Step 2: Verify lint passes**

Run: `cd dx-web && npx eslint src/features/web/play-core/hooks/use-vocab-battle.ts --no-error-on-unmatched-pattern`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/play-core/hooks/use-vocab-battle.ts
git commit -m "feat: add useVocabBattle hook for per-item letter spelling logic"
```

---

## Task 6: vocab-battle component

**Files:**
- Rewrite: `dx-web/src/features/web/play-core/components/game-vocab-battle.tsx`

- [ ] **Step 1: Rewrite `game-vocab-battle.tsx` with data-driven UI**

```tsx
// dx-web/src/features/web/play-core/components/game-vocab-battle.tsx
"use client";

import { Zap, SkipForward, Check } from "lucide-react";
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/components/ui/hover-card";
import { useVocabBattle } from "@/features/web/play-core/hooks/use-vocab-battle";

export function GameVocabBattle() {
  const {
    targetWord,
    translation,
    letterSlots,
    keyboardLetters,
    usedKeyIndices,
    hasError,
    isRevealed,
    playerShields,
    opponentShields,
    opponentSlots,
    competitive,
    progress,
    combo,
    pressLetter,
    advanceAfterReveal,
    skipItem,
  } = useVocabBattle();

  if (!targetWord) return null;

  return (
    <div className="flex w-full max-w-[760px] flex-col rounded-[20px] border border-border bg-card shadow-sm">
      {/* Opponent zone */}
      <div
        className={`flex flex-col items-center gap-4 px-6 py-7 md:px-8 ${
          !competitive ? "pointer-events-none opacity-40" : ""
        }`}
      >
        <div className="flex items-center gap-2.5">
          <span className="text-xs text-muted-foreground">🤖 对手</span>
        </div>
        <div className="flex items-center justify-center gap-2">
          {opponentShields.map((active, i) => (
            <div
              key={i}
              className={`h-6 w-6 rounded-full border-2 transition-colors ${
                active
                  ? "border-red-400 bg-red-400"
                  : "border-border bg-muted"
              }`}
            />
          ))}
        </div>
        <div className="flex items-center justify-center gap-2.5">
          {opponentSlots.map((slot, i) => (
            <div
              key={i}
              className="flex h-10 w-10 items-center justify-center rounded-lg border border-border bg-muted"
            >
              <span className="text-sm font-medium text-slate-300">
                {slot.filled ? slot.letter : "?"}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Translation zone */}
      <div className="flex flex-col items-center gap-2.5 bg-gradient-to-b from-red-50/0 via-red-50 to-red-50/0 px-6 py-4 md:px-8">
        <p className="text-center text-2xl font-extrabold tracking-wider text-foreground md:text-[32px]">
          {translation}
        </p>
        <div className="h-0.5 w-full rounded-full bg-gradient-to-r from-red-500/0 via-red-500/30 via-30% via-teal-500/30 via-70% to-teal-500/0" />
      </div>

      {/* Player zone */}
      <div className="flex flex-col items-center gap-4 px-6 py-5 md:px-8">
        <div
          className={`flex items-center justify-center gap-2.5 ${
            hasError ? "animate-[shake_0.3s_ease-in-out]" : ""
          }`}
        >
          {letterSlots.map((slot, i) => (
            <div
              key={i}
              className={`flex h-10 w-10 items-center justify-center rounded-lg border transition-colors ${
                slot.filled
                  ? "border-teal-300 bg-teal-50"
                  : "border-border bg-muted"
              }`}
            >
              <span
                className={`text-sm font-semibold ${
                  slot.filled ? "text-teal-600" : "text-slate-300"
                }`}
              >
                {slot.filled ? slot.filledLetter : "_"}
              </span>
            </div>
          ))}
        </div>
        <div className="flex items-center justify-center gap-2">
          {playerShields.map((active, i) => (
            <div
              key={i}
              className={`h-6 w-6 rounded-full border-2 transition-colors ${
                active
                  ? "border-teal-400 bg-teal-400"
                  : "border-border bg-muted"
              }`}
            />
          ))}
        </div>
        <div className="flex items-center gap-2.5">
          <span className="text-xs text-muted-foreground">🎯 我</span>
        </div>
      </div>

      <div className="h-px w-full bg-muted" />

      {/* Combo row */}
      {combo.streak >= 3 && (
        <div className="flex items-center justify-center gap-3 px-6 py-2 md:px-8">
          <span className="text-[13px] font-medium text-muted-foreground">连击</span>
          <div className="flex items-center gap-1.5 rounded-lg bg-red-500 px-3 py-1">
            <Zap className="h-3 w-3 text-white" />
            <span className="text-xs font-bold text-white">
              &times;{combo.streak}
            </span>
          </div>
        </div>
      )}

      {/* Hint + action row */}
      <div className="flex flex-col items-center gap-3 px-6 pb-6 pt-3 md:px-8">
        <span className="text-xs font-medium text-muted-foreground">
          {competitive ? "点击字母发射炮弹击碎对手护盾" : "拼写单词"}
        </span>

        {/* Letter keyboard */}
        {!isRevealed && (
          <div className="flex flex-wrap items-center justify-center gap-2.5">
            {keyboardLetters.map((letter, i) => {
              const isUsed = usedKeyIndices.has(i);
              return (
                <button
                  key={i}
                  type="button"
                  disabled={isUsed}
                  onClick={() => pressLetter(i)}
                  className={`flex h-12 w-12 items-center justify-center rounded-xl shadow-md transition-opacity md:h-14 md:w-14 ${
                    isUsed
                      ? "bg-slate-400 opacity-40"
                      : "bg-slate-800 hover:bg-slate-700"
                  }`}
                >
                  <span className="text-lg font-bold text-white">{letter}</span>
                </button>
              );
            })}
          </div>
        )}

        {/* Revealed: show full word + advance */}
        {isRevealed && (
          <div className="flex flex-col items-center gap-3">
            <span className="text-lg font-bold text-teal-600">{targetWord}</span>
            <button
              type="button"
              onClick={advanceAfterReveal}
              className="flex items-center gap-2 rounded-xl bg-teal-600 px-9 py-3"
            >
              <Check className="h-4 w-4 text-white" />
              <span className="text-xs font-semibold text-white">
                {progress.current >= progress.total ? "查看结果" : "下一题"}
              </span>
            </button>
          </div>
        )}

        {/* Skip button (non-competitive only) */}
        {!isRevealed && (
          <div className="flex items-center gap-3">
            {competitive ? (
              <HoverCard openDelay={200}>
                <HoverCardTrigger asChild>
                  <button
                    type="button"
                    disabled
                    className="flex items-center gap-2 rounded-xl border border-border bg-muted px-5 py-3 opacity-40 cursor-not-allowed"
                  >
                    <SkipForward className="h-4 w-4 text-muted-foreground" />
                    <span className="text-xs font-medium text-muted-foreground">跳过</span>
                  </button>
                </HoverCardTrigger>
                <HoverCardContent className="w-auto px-3 py-1.5 text-sm" side="top">
                  竞技模式禁用
                </HoverCardContent>
              </HoverCard>
            ) : (
              <button
                type="button"
                onClick={skipItem}
                className="flex items-center gap-2 rounded-xl border border-border bg-muted px-5 py-3"
              >
                <SkipForward className="h-4 w-4 text-muted-foreground" />
                <span className="text-xs font-medium text-muted-foreground">跳过</span>
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify lint passes**

Run: `cd dx-web && npx eslint src/features/web/play-core/components/game-vocab-battle.tsx --no-error-on-unmatched-pattern`
Expected: No errors

- [ ] **Step 3: Verify build passes**

Run: `cd dx-web && npx next build 2>&1 | tail -20`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/play-core/components/game-vocab-battle.tsx
git commit -m "feat: implement vocab-battle game component with letter keyboard and opponent zone"
```

---

## Task 7: Full verification

**Files:** None (verification only)

- [ ] **Step 1: Run full ESLint on all changed files**

Run:
```bash
cd dx-web && npx eslint \
  src/features/web/play-core/hooks/use-vocab-match.ts \
  src/features/web/play-core/hooks/use-vocab-elimination.ts \
  src/features/web/play-core/hooks/use-vocab-battle.ts \
  src/features/web/play-core/components/game-vocab-match.tsx \
  src/features/web/play-core/components/game-vocab-elimination.tsx \
  src/features/web/play-core/components/game-vocab-battle.tsx \
  --no-error-on-unmatched-pattern
```
Expected: No errors

- [ ] **Step 2: Run full production build**

Run: `cd dx-web && npx next build 2>&1 | tail -30`
Expected: Build succeeds with zero type errors

- [ ] **Step 3: Verify existing components are untouched**

Run:
```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source && git diff --name-only HEAD~6..HEAD
```
Expected: Only these files appear:
- `docs/superpowers/specs/2026-04-08-vocab-game-modes-design.md`
- `dx-web/src/features/web/play-core/hooks/use-vocab-match.ts` (new)
- `dx-web/src/features/web/play-core/hooks/use-vocab-elimination.ts` (new)
- `dx-web/src/features/web/play-core/hooks/use-vocab-battle.ts` (new)
- `dx-web/src/features/web/play-core/components/game-vocab-match.tsx` (modified)
- `dx-web/src/features/web/play-core/components/game-vocab-elimination.tsx` (modified)
- `dx-web/src/features/web/play-core/components/game-vocab-battle.tsx` (modified)

No other files changed. word-sentence, shells, store, context, backend — all untouched.

- [ ] **Step 4: Spot-check that shells still register all modes**

Run:
```bash
grep -n "modeComponents" \
  dx-web/src/features/web/play-single/components/game-play-shell.tsx \
  dx-web/src/features/web/play-pk/components/pk-play-shell.tsx \
  dx-web/src/features/web/play-group/components/group-play-shell.tsx
```
Expected: Each file shows all 4 modes (WORD_SENTENCE, VOCAB_MATCH, VOCAB_ELIMINATION, VOCAB_BATTLE)

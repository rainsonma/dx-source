# Spelling Input Dot Indicators and Answer-Count Fix

**Date:** 2026-04-17
**Scope:** `dx-web` only — word-sentence game play page
**Type:** UX refinement + counter bug fix

## Background

The game play page (single-player word-sentence mode) shows a spelling input row where the user types the English words of a sentence one at a time. Items in the sentence include both answer words (`answer: true`) and non-answer glue tokens such as punctuation (`answer: false`). Non-answer tokens auto-fill into the typed-words strip so the user only has to type the answer words.

Below the input row, a counter renders `( {wordProgress.current} / {wordProgress.total} )`. It currently uses:

```ts
// dx-web/src/features/web/play-core/hooks/use-word-sentence.ts:52
const wordProgress = {
  current: typedWords.length,
  total: items.length,
};
```

Both sides include non-answer items, so the denominator and numerator are inflated by punctuation. The displayed fraction does not match what the user perceives as "words to type."

Separately, the user has no per-word visual indicator above the input row — only a single progress bar under the input showing prefix correctness for the currently-typed word. There is no at-a-glance view of how many answer words remain and which ones have been attempted.

## Goals

1. Fix `wordProgress` to count only `answer === true` items on both sides of the fraction.
2. Add a right-aligned row of small circular dot indicators directly above the spelling input bar, one dot per answer word, that communicate per-word status in real time:
   - **Gray** — default / not yet reached.
   - **Red** — active wrong attempt on the current word.
   - **Green (teal)** — word typed correctly.

## Non-Goals

- No backend changes. Pure `dx-web` refactor.
- No score, session, tracking, or review logic changes.
- No accessibility announcement for dot color changes (the counter remains the accessible progress signal; dots are `aria-hidden="true"`).
- No animation beyond a simple CSS color transition.
- No persistence of per-word wrongness across items or sessions.

## Design

### Architecture overview

All changes live inside the existing `play-core` feature folder. No new files, no new types, no new context.

```
features/web/play-core/
├── hooks/
│   └── use-word-sentence.ts          # fix wordProgress filter
└── components/
    ├── game-word-sentence.tsx        # pass wordProgress prop through
    └── spelling-input-row.tsx        # add dot row above existing bar
```

Data flow:

```
useWordSentence (hook)
  └─> wordProgress { current, total }  // both filtered to answer=true
        └─> GameWordSentence (page-level component)
              ├─> count badge (existing)
              └─> SpellingInputRow (dots + bar)
                     └─> derives per-dot color from wordProgress + hasError + isWrongInput
```

### Part 1 — Counter fix

**File:** `dx-web/src/features/web/play-core/hooks/use-word-sentence.ts`

Replace the `wordProgress` assignment at line 52:

```ts
const wordProgress = {
  current: typedWords.filter((w) => w.isAnswer).length,
  total: items.filter((it) => it.answer).length,
};
```

Rationale:
- `typedWords: TypedWord[]` has `isAnswer: boolean` (see `types/spelling.ts`). Non-answer glue tokens are pushed into the strip with `isAnswer: false` by `skipNonAnswers`. Filtering by `isAnswer` yields the true count of user-typed words so far.
- `items: SpellingItem[]` has `answer: boolean`. Filtering by `answer` yields the target count of words the user must type for the item.

Grep confirms `wordProgress` is consumed only by `GameWordSentence` (render) and returned from the hook — no other call site to coordinate with.

### Part 2 — Dot indicators

**File:** `dx-web/src/features/web/play-core/components/spelling-input-row.tsx`

#### New prop

```ts
interface SpellingInputRowProps {
  typedWords: TypedWord[];
  inputValue: string;
  hasError: boolean;
  isRevealed: boolean;
  currentWord: SpellingItem | null;
  showAnswer: boolean;
  wordProgress: { current: number; total: number }; // NEW
  onInputChange: (value: string) => void;
  onKeyDown: (e: React.KeyboardEvent<HTMLInputElement>) => void;
}
```

`GameWordSentence` already destructures `wordProgress` from `useWordSentence` — it passes it straight through.

#### Dot state

For each dot at index `i` in `[0, wordProgress.total)`:

| Condition                                               | Color                 |
| ------------------------------------------------------- | --------------------- |
| `i < wordProgress.current`                              | `bg-teal-500` (done)  |
| `i === wordProgress.current && (hasError \|\| isWrongInput)` | `bg-red-500` (active wrong) |
| `i === wordProgress.current`                            | `bg-slate-300` (active) |
| `i > wordProgress.current`                              | `bg-slate-300` (pending) |

`isWrongInput` (already computed on line 121 of the current file) captures real-time prefix mismatches while typing, matching the semantics of the existing red progress bar under the input. `hasError` captures post-submit wrongness. Using both OR'd means the dot mirrors the existing bar's red state exactly.

The palette uses teal (not literal green) for "done" to stay consistent with the rest of the game UI — progress ring, correct-prefix text color, correct progress bar segment all use teal.

#### Layout

The component's root changes from a single `div` to a `flex-col` wrapper with two children:

```tsx
return (
  <div className="flex flex-col gap-2">
    {/* Dot row — right-aligned, aligned with the bar's right padding */}
    <div
      className={`flex h-2 items-center justify-end gap-1.5 px-4 md:px-6 ${
        isRevealed ? "invisible" : ""
      }`}
      aria-hidden="true"
    >
      {Array.from({ length: wordProgress.total }, (_, i) => {
        const color =
          i < wordProgress.current
            ? "bg-teal-500"
            : i === wordProgress.current && (hasError || isWrongInput)
              ? "bg-red-500"
              : "bg-slate-300";
        return (
          <span
            key={i}
            className={`h-1.5 w-1.5 rounded-full transition-colors ${color}`}
          />
        );
      })}
    </div>

    {/* Existing input bar — unchanged structure */}
    <div
      ref={containerRef}
      onMouseDown={onMouseDown}
      onMouseMove={onMouseMove}
      onMouseUp={handleDragEnd}
      onMouseLeave={handleDragEnd}
      onTouchStart={onTouchStart}
      onTouchMove={onTouchMove}
      onTouchEnd={handleDragEnd}
      className="flex items-center gap-2 overflow-x-hidden rounded-[14px] border border-border bg-muted px-4 py-3.5 md:gap-3 md:px-6"
      style={{
        maskImage: "linear-gradient(to right, transparent, black 40px, black)",
        WebkitMaskImage: "linear-gradient(to right, transparent, black 40px, black)",
      }}
    >
      {/* all existing content: spacer, typedWords, input wrapper */}
    </div>
  </div>
);
```

Drag-to-scroll, overflow mask, focus logic, `containerRef`, auto-scroll-to-rightmost, and the shake animation all remain on the inner bar `div` — nothing about the bar's behavior is touched.

#### Visual specs

- Dot size: `h-1.5 w-1.5` (6px × 6px)
- Gap between dots: `gap-1.5` (6px)
- Row vertical slot: `h-2` (8px, reserves height even when `wordProgress.total === 0`)
- Row horizontal padding: `px-4 md:px-6` — matches the bar's horizontal padding so dots align flush with the bar's right edge
- Transition: `transition-colors` on each dot

### Part 3 — `GameWordSentence` wiring

**File:** `dx-web/src/features/web/play-core/components/game-word-sentence.tsx`

Pass the already-destructured `wordProgress` to `<SpellingInputRow />`:

```tsx
<SpellingInputRow
  typedWords={typedWords}
  inputValue={inputValue}
  hasError={hasError}
  isRevealed={isRevealed}
  currentWord={currentWord}
  showAnswer={showAnswer}
  wordProgress={wordProgress}  {/* NEW */}
  onInputChange={setInputValue}
  onKeyDown={handleKeyDown}
/>
```

The count badge below the bar already reads `wordProgress.current` and `wordProgress.total` (line 190 of the current file) — once the hook fix lands, it will show the corrected values automatically.

## Behavior Matrix

| Action                                    | Counter                    | Dots                                                    |
| ----------------------------------------- | -------------------------- | ------------------------------------------------------- |
| New item loads (N answer words)           | `(0/N)`                    | N gray dots                                             |
| Auto-skip a leading punctuation token     | `(0/N)` (punctuation ignored) | no change (no dot consumed)                           |
| User types matching prefix                | `(0/N)`                    | dot 0 stays gray                                        |
| User types mismatching prefix             | `(0/N)`                    | dot 0 red                                               |
| User clears input                         | `(0/N)`                    | dot 0 back to gray                                      |
| User submits wrong word                   | `(0/N)`                    | dot 0 red                                               |
| User submits correct word                 | `(1/N)`                    | dot 0 green; dot 1 now "active" (gray)                  |
| User presses Tab (skip)                   | freezes at current         | completed dots stay green; remaining dots stay gray     |
| Item reveals after last correct word      | `(N/N)`                    | all N green, then fade to `invisible` with reveal       |
| Next item loads                           | recomputes                 | dots reset to next item's answer count                  |

## Testing

### Lint / static checks

```bash
cd dx-web
npm run lint
```

The new prop is typed; Next.js will surface any TypeScript errors on build.

### Manual verification

1. `cd dx-web && npm run dev`, sign in, open `/hall`, enter any single-player game/level with sentence items.
2. Pick an item whose sentence contains punctuation (period, comma, exclamation, etc.).
3. Verify:
   - Counter below the bar shows answer-only progress: e.g., `(0/3)` for a 3-word sentence with punctuation, not `(0/4)` or `(1/4)`.
   - Exactly 3 gray dots render right-aligned above the input bar.
   - Typing a wrong character → the first dot turns red immediately.
   - Clearing the input → the first dot returns to gray.
   - Typing the correct word and pressing Space/Enter → first dot turns green, counter → `(1/3)`, second dot becomes the active (gray) one.
   - Repeat for remaining words; all dots should end green and the bar reveals.
4. Try a Tab-skip mid-item — completed dots stay green, remaining dots stay gray (no red carpet).
5. Resize the viewport to mobile — dots stay right-aligned, no layout jump when items change.
6. Try an item that starts with a punctuation token — it auto-fills, the dot row is unaffected.

### Regression checks

- Drag-to-scroll the typed-words strip still works.
- Shake animation on wrong submit still plays on the input wrapper.
- Answer hint (`showAnswer`) toggle still shows the ghost text inside the input.
- Reveal mode still hides the input field and flips the dots to invisible without layout jump.
- Score, session recording, review marking, mastered/unknown marking — all untouched.

## Edge Cases

- **Item with zero answer words** (theoretically impossible but defended): counter shows `(0/0)`, dot row renders empty but reserves `h-2` height so the layout doesn't collapse.
- **Very long sentences** (10+ answer words): dots at 6px with 6px gap = ~120px total, fits comfortably in the 3xl container. No wrapping needed.
- **Mobile breakpoints**: padding matches the bar (`px-4 md:px-6`), so right alignment holds across viewports.
- **Rapid item transitions**: dots derive from `wordProgress` (derived state), so they reset atomically with item change.
- **User focuses away from input**: no effect on dot state; `hasError` and `isWrongInput` are independent of focus.

## Risks

- **Risk:** Misreading `isAnswer` / `answer` fields in legacy content items where the flag might be missing.
  - **Mitigation:** `.filter((it) => it.answer)` treats `undefined` / `null` / `false` uniformly as non-answer. Same for `typedWords.filter((w) => w.isAnswer)`. No runtime explosion.

- **Risk:** A downstream consumer (not found by grep) depends on the old `wordProgress` semantics.
  - **Mitigation:** Grep confirmed only two callsites — the hook's own return and the one render site. The change is safe.

- **Risk:** Dot row alignment differs by 1px from the bar's right edge on certain zoom levels.
  - **Mitigation:** Both dot row and bar use the same `px-4 md:px-6` padding and no extra borders. Verified in the behavior matrix.

## Rollout

Single PR, small diff (~30 lines added, 2 lines changed across 3 files). No migrations, no feature flag. Ship on merge to `main`.

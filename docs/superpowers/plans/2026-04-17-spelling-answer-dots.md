# Spelling Answer-Count Fix and Dot Indicators Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the word-progress fraction to count only answer words, and add a right-aligned row of status dots above the spelling input bar that turn gray → red (on wrong attempt) → teal (when correct).

**Architecture:** Pure frontend change in `dx-web`. One hook (`use-word-sentence.ts`) gets a 2-line filter fix. One component (`spelling-input-row.tsx`) gains a new `wordProgress` prop and renders a small dot row above its existing rounded input bar. One call site (`game-word-sentence.tsx`) passes the already-destructured `wordProgress` through.

**Tech Stack:** Next.js 16 (App Router), React 19, TypeScript, TailwindCSS v4, Zustand (via existing `use-game-store` — untouched).

**Spec:** `docs/superpowers/specs/2026-04-17-spelling-answer-dots-design.md`

**Verification model:** There is no component test runner in `dx-web` (only `lint` and `build` scripts). Verification per task is `npm run lint` + TypeScript compile via `npm run build` where appropriate, and end-of-plan manual browser verification against the spec's test plan.

---

## File Structure

Files touched (no new files):

| File | Role |
| ---- | ---- |
| `dx-web/src/features/web/play-core/hooks/use-word-sentence.ts` | Fix `wordProgress` to filter by answer-only. |
| `dx-web/src/features/web/play-core/components/spelling-input-row.tsx` | Accept `wordProgress` prop; render dot row above bar. |
| `dx-web/src/features/web/play-core/components/game-word-sentence.tsx` | Pass `wordProgress` to `<SpellingInputRow />`. |

No backend, no API types, no store, no new files.

---

## Task 1: Fix `wordProgress` to count only answer words

**Files:**
- Modify: `dx-web/src/features/web/play-core/hooks/use-word-sentence.ts:52-55`

- [ ] **Step 1: Open the file and locate the current `wordProgress` block (lines 52–55)**

Current code to replace:

```ts
const wordProgress = {
  current: typedWords.length,
  total: items.length,
};
```

- [ ] **Step 2: Replace with answer-filtered counts**

```ts
const wordProgress = {
  current: typedWords.filter((w) => w.isAnswer).length,
  total: items.filter((it) => it.answer).length,
};
```

Rationale:
- `typedWords: TypedWord[]` — each element has `isAnswer: boolean`. Auto-skipped punctuation is pushed with `isAnswer: false`.
- `items: SpellingItem[]` — each element has `answer: boolean`. Non-answer items are punctuation/glue.
- Filtering both sides yields the fraction of answer words the user has typed out of total required.

- [ ] **Step 3: Run lint to confirm no syntax/style issues**

```bash
cd dx-web
npm run lint
```

Expected: no errors, no warnings introduced by this change.

- [ ] **Step 4: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/features/web/play-core/hooks/use-word-sentence.ts
git commit -m "fix(web): count only answer words in spelling word progress"
```

---

## Task 2: Add `wordProgress` prop to `SpellingInputRow` and render dot row

**Files:**
- Modify: `dx-web/src/features/web/play-core/components/spelling-input-row.tsx`

- [ ] **Step 1: Extend the `SpellingInputRowProps` interface with a `wordProgress` field**

Locate the interface at lines 20–29 in `spelling-input-row.tsx`. Replace it with:

```ts
interface SpellingInputRowProps {
  typedWords: TypedWord[];
  inputValue: string;
  hasError: boolean;
  isRevealed: boolean;
  currentWord: SpellingItem | null;
  showAnswer: boolean;
  wordProgress: { current: number; total: number };
  onInputChange: (value: string) => void;
  onKeyDown: (e: React.KeyboardEvent<HTMLInputElement>) => void;
}
```

- [ ] **Step 2: Destructure `wordProgress` in the component signature**

Locate the function signature at lines 31–40. Replace it with:

```ts
export function SpellingInputRow({
  typedWords,
  inputValue,
  hasError,
  isRevealed,
  currentWord,
  showAnswer,
  wordProgress,
  onInputChange,
  onKeyDown,
}: SpellingInputRowProps) {
```

- [ ] **Step 3: Wrap the existing root `div` in a `flex-col` wrapper and insert the dot row above it**

The current `return` begins at line 138 with a single outer `div` that carries drag handlers and the rounded-bar styling. Replace the whole `return (...)` block (lines 138–231) with:

```tsx
  return (
    <div className="flex flex-col gap-2">
      {/* Dot indicators — one per answer word, right-aligned above the bar */}
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

      {/* Existing input bar — drag/scroll handlers and styling unchanged */}
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
          maskImage:
            "linear-gradient(to right, transparent, black 40px, black)",
          WebkitMaskImage:
            "linear-gradient(to right, transparent, black 40px, black)",
        }}
      >
        <div className="min-w-0 flex-1" />

        {typedWords.map((word, i) => (
          <span
            key={`${i}-${word.text}`}
            className={`shrink-0 text-base ${
              word.isAnswer
                ? "font-medium text-foreground"
                : "font-medium text-muted-foreground"
            }`}
          >
            {word.text}
          </span>
        ))}

        <div
          className={`relative shrink-0 ${isRevealed ? "invisible" : ""} ${hasError ? "animate-[shake_0.4s_ease-in-out]" : ""}`}
          style={{ width: `${inputWidthCh}ch` }}
        >
          {/* Ghost text — answer hint */}
          {showAnswer && currentWord && (
            <span
              className="pointer-events-none absolute inset-0 flex items-center justify-center text-base font-bold text-slate-300"
              aria-hidden="true"
            >
              {currentWord.item}
            </span>
          )}
          <input
            ref={inputRef}
            type="text"
            value={inputValue}
            onChange={(e) => onInputChange(e.target.value)}
            onKeyDown={handleKeyDownWithSound}
            onFocus={() => setIsFocused(true)}
            onBlur={() => {
              if (!isRevealed && !overlay) {
                requestAnimationFrame(() => inputRef.current?.focus());
              } else {
                setIsFocused(false);
              }
            }}
            aria-label="输入单词"
            autoComplete="off"
            autoCapitalize="off"
            spellCheck={false}
            className={`relative z-10 w-full px-1 text-center text-base font-bold outline-none ${
              showAnswer ? "bg-transparent" : "bg-border"
            } ${
              hasError || isWrongInput
                ? "text-red-600"
                : isCorrectPrefix
                  ? "text-teal-600"
                  : "text-foreground"
            }`}
          />
          <div className="flex h-[3px] overflow-hidden rounded-full">
            {isFocused && !(hasError || isWrongInput) && tealPercent > 0 && (
              <div
                className="bg-teal-600 transition-all duration-150"
                style={{ width: `${tealPercent}%` }}
              />
            )}
            <div
              className={`flex-1 transition-colors ${
                hasError || isWrongInput
                  ? "bg-red-500"
                  : isFocused
                    ? "bg-slate-900"
                    : "bg-slate-400"
              }`}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
```

Key correctness points:
- The inner bar `<div ref={containerRef} ...>` is the **same** element it was before — its `className`, `style`, drag/scroll handlers, and all children are byte-for-byte identical to the previous top-level `div`. We simply put it inside a `flex-col` wrapper with a sibling above it.
- `aria-hidden="true"` on the dot row — the counter text below the bar (e.g., `( 1 / 3 )`) remains the accessible progress signal. This prevents screen readers from announcing every dot color change.
- The dot row uses `isRevealed` → `invisible` so it reserves height and fades in sync with the input field when an item reveals. This avoids layout jumps on reveal transitions.
- `transition-colors` gives a smooth color change when a dot flips red (wrong keystroke) or teal (correct submit).
- `h-2` on the row reserves 8px of height even when `wordProgress.total === 0` so empty-dots items don't collapse the layout.
- Horizontal padding `px-4 md:px-6` matches the bar's own padding, so the last dot aligns flush with the bar's right edge at every breakpoint.

- [ ] **Step 4: Run lint and TypeScript build to confirm the new prop and restructured JSX compile cleanly**

```bash
cd dx-web
npm run lint
```

Expected: no new errors or warnings.

Then:

```bash
npm run build
```

Expected: build succeeds. Watch for any TypeScript error about the `wordProgress` prop — if the call site in `game-word-sentence.tsx` hasn't been updated yet, this step **will** report a type error for the missing prop at `<SpellingInputRow />`. That is expected and will be fixed in Task 3. If you prefer, skip `npm run build` here and run it only after Task 3.

- [ ] **Step 5: Do not commit yet — the next task must land together to keep the tree compiling**

The `wordProgress` prop is required. Until Task 3 passes it from `game-word-sentence.tsx`, the tree fails type-checking. Hold the commit.

---

## Task 3: Pass `wordProgress` from `GameWordSentence` to `SpellingInputRow`

**Files:**
- Modify: `dx-web/src/features/web/play-core/components/game-word-sentence.tsx:174-183`

- [ ] **Step 1: Add `wordProgress={wordProgress}` to the `<SpellingInputRow />` render**

The `<SpellingInputRow />` invocation currently reads (lines 174–183 of `game-word-sentence.tsx`):

```tsx
<SpellingInputRow
  typedWords={typedWords}
  inputValue={inputValue}
  hasError={hasError}
  isRevealed={isRevealed}
  currentWord={currentWord}
  showAnswer={showAnswer}
  onInputChange={setInputValue}
  onKeyDown={handleKeyDown}
/>
```

Replace with:

```tsx
<SpellingInputRow
  typedWords={typedWords}
  inputValue={inputValue}
  hasError={hasError}
  isRevealed={isRevealed}
  currentWord={currentWord}
  showAnswer={showAnswer}
  wordProgress={wordProgress}
  onInputChange={setInputValue}
  onKeyDown={handleKeyDown}
/>
```

`wordProgress` is already destructured from `useWordSentence()` at line 36 of this file and is already used by the counter badge at line 190 — no new import, no other wiring needed.

- [ ] **Step 2: Run lint**

```bash
cd dx-web
npm run lint
```

Expected: no errors.

- [ ] **Step 3: Run TypeScript build to confirm the whole tree compiles**

```bash
cd dx-web
npm run build
```

Expected: build succeeds with no TypeScript errors. The `wordProgress` prop is now satisfied.

- [ ] **Step 4: Commit Tasks 2 and 3 together**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/src/features/web/play-core/components/spelling-input-row.tsx dx-web/src/features/web/play-core/components/game-word-sentence.tsx
git commit -m "feat(web): add answer-word dot indicators above spelling input"
```

---

## Task 4: Manual browser verification

**Files:** none — this is a verification task.

- [ ] **Step 1: Start the dev server**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run dev
```

Wait for the server to report it is listening on `http://localhost:3000`.

- [ ] **Step 2: Also start the API if it isn't already running**

In another terminal:

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api
go run .
```

The API must be reachable at `http://localhost:3001` for dashboard data to load.

- [ ] **Step 3: Sign in and open a single-player game/level with sentence items**

- Navigate to `http://localhost:3000`, sign in.
- Go to `/hall`, pick a game whose level has sentence-type content items (any sentence mode).
- Enter a level and reach the word-sentence play screen.

- [ ] **Step 4: Verify the counter shows answer-only progress**

- Pick (or wait for) an item whose sentence contains punctuation, e.g., `"Hello, world."` — 2 answer words, 2 punctuation tokens.
- Confirm the counter under the input bar shows `( 0 / 2 )`, not `( 0 / 4 )` or `( 2 / 4 )`.

Expected (for a 2-answer-word item): `( 0 / 2 )` initially; advances to `( 1 / 2 )` after the first correct word, `( 2 / 2 )` after the second.

- [ ] **Step 5: Verify dot count and default color**

- Exactly N dots render right-aligned above the input bar, where N = the number of answer words in the item.
- All dots are gray (`bg-slate-300`) before the user types anything.

- [ ] **Step 6: Verify red on wrong keystroke**

- Type a single character that does NOT match the first letter of the current answer word.
- The first (leftmost remaining) dot turns red (`bg-red-500`) immediately.
- Clear the input — the dot returns to gray.

- [ ] **Step 7: Verify red on wrong submit**

- Type a full word that is wrong, press Enter or Space.
- The first dot turns red; the input shakes; the counter stays at `( 0 / N )`.
- Clear the input — the dot stays red until you start typing a correct prefix (or stays red per existing `hasError` semantics until the input changes); typing a correct character clears `hasError`.

- [ ] **Step 8: Verify teal on correct submit**

- Type the correct word, press Enter or Space.
- The first dot turns teal (`bg-teal-500`), the counter advances to `( 1 / N )`, and the second dot becomes "active" (gray, ready to turn red on a mistake).
- Repeat for the remaining answer words.

- [ ] **Step 9: Verify reveal behavior**

- Continue until all answer words are typed correctly. The item reveals (the hidden mask flips to the revealed sentence).
- The dot row becomes invisible (opacity zero / CSS `invisible`) — no layout jump.
- Press Enter or Space to advance to the next item; the dots reset to the new item's answer count.

- [ ] **Step 10: Verify skip (Tab) behavior**

- On a new item (ideally one with 3+ answer words), type the first word correctly — the first dot turns teal.
- Press Tab to skip.
- Completed dots stay teal; remaining dots stay gray. No dots flip red on skip.
- After pressing Tab/Enter/Space to advance, the next item's dots render fresh.

- [ ] **Step 11: Verify mobile breakpoints**

- Open Chrome DevTools, switch to a mobile viewport (e.g., iPhone 14 Pro, 390px wide).
- Confirm dots are still right-aligned and visible above the bar.
- Confirm no layout jump when items change, when revealing, or when typing.

- [ ] **Step 12: Verify regressions in existing behaviors**

Spot-check each of the following to ensure nothing else broke:

- Drag-to-scroll the typed-words strip still works (click-and-drag on the bar away from the input field).
- The red progress bar under the input still turns red on wrong input and shows teal fill on correct prefix.
- The shake animation still plays on wrong submit.
- The `答案` hint toggle still shows the ghost text inside the input.
- The `生词`, `掌握`, `跳过`, `确认` buttons still work.
- In competitive modes (PK or group), the `答案` and `跳过` buttons are still disabled with their tooltips.
- Score, combo, and streak animations still play as before.

- [ ] **Step 13: Stop dev server**

Stop the `npm run dev` process (Ctrl-C) once verification is complete.

- [ ] **Step 14: No commit needed — verification task**

Task 4 produces no file changes; nothing to commit.

---

## Task 5: Final checks and merge

**Files:** none — repo hygiene.

- [ ] **Step 1: Confirm the repo is clean and on `main`**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git status
git log --oneline -3
```

Expected:
- Working tree clean.
- Two new commits on `main`: the `fix(web): count only answer words...` commit from Task 1 and the `feat(web): add answer-word dot indicators...` commit from Tasks 2–3.

- [ ] **Step 2: Run a final full lint to sanity-check the whole web project**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run lint
```

Expected: no errors, no warnings introduced.

- [ ] **Step 3: Run a full build to confirm the production bundle compiles**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web
npm run build
```

Expected: build succeeds.

- [ ] **Step 4: Report completion**

Summarize to the user:
- What shipped: answer-only word-progress counter + per-answer-word dot indicators above the spelling input bar.
- Where: `main` branch, two commits.
- Verified: lint, build, manual browser walkthrough including edge cases (punctuation items, skip, reveal, mobile, regression spots).

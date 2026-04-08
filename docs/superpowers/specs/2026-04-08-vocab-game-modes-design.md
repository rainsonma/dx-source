# Vocab Game Modes Implementation Design

**Date:** 2026-04-08
**Status:** Approved
**Scope:** Wire vocab-battle, vocab-match, vocab-elimination into play system

---

## 1. Context

The word-sentence (连词成句) game mode is fully implemented across single, PK, and group play shells. Three additional vocab modes exist as static UI mockups but have no game logic:

- **vocab-battle** (词汇对轰) — letter-by-letter spelling duel
- **vocab-match** (词汇配对) — two-column word-definition matching
- **vocab-elimination** (词汇消消乐) — grid-based pair elimination

All three shells (`GamePlayShell`, `PkPlayShell`, `GroupPlayShell`) already import and register these components in their `modeComponents` map. The backend content pipeline already generates vocab ContentItems with `items` JSON (SpellingItem arrays containing phonetic, POS, translation data). No backend or shell changes are needed.

---

## 2. Content Data Model

Each vocab ContentItem has:

| Field | Example | Usage |
|-------|---------|-------|
| `content` | `"apple"` | English word/phrase (the answer) |
| `translation` | `"苹果"` | Chinese definition (the prompt) |
| `contentType` | `"word"` or `"phrase"` | Single word vs multi-word |
| `items` | `[{ item: "apple", phonetic: { uk: "/ˈæpl/", us: "/ˈæpəl/" }, pos: "名词", translation: "苹果", answer: true, position: 1 }]` | SpellingItem array with linguistic metadata |

Level sizes by mode (set during content authoring):
- vocab-match: ~5 pairs per meta group, multiple groups per level
- vocab-elimination: ~8 pairs per meta group, multiple groups per level
- vocab-battle: ~20 items per level

---

## 3. Game Mode Specifications

### 3.1 vocab-battle (词汇对轰)

**Progression:** Per-item (same as word-sentence). Uses `currentIndex` to track current word.

**Gameplay:**
1. Display Chinese translation in center divider zone
2. Show player's letter slots (bottom, teal) — blanks to fill
3. Show scrambled letter keyboard buttons at bottom
4. Player clicks letter buttons to spell the word one letter at a time
5. Correct letter → fills next blank in player's letter row
6. All letters filled correctly → word complete, reveal, advance
7. Wrong letter → visual error feedback (shake), mark `hadWrongAttempt`
8. After reveal: `recordResult(isCorrect)` + `recordAnswerAction()` fire-and-forget
9. Press Enter/Space/click confirm to advance to next word

**Shields (lives) system:**
- Each player starts with 5 shields per word
- Correct letter by player → removes one opponent shield
- Opponent correct letter → removes one player shield (in competitive mode)
- Shields are purely visual feedback — do not affect scoring

**Single-player exception:**
- Detected via `!competitive` from `useGamePlayActions()`
- Opponent zone rendered at reduced opacity (~40%), shields and letters frozen
- No opponent progress animation
- Hint text changes from "击碎对手护盾" to "拼写单词"

**Competitive mode (PK/group):**
- Opponent zone fully active
- Opponent progress updated via SSE `onPlayerAction` events (already handled by PK/group shells)
- Skip and answer buttons disabled (existing `competitive` pattern)

**Letter keyboard generation:**
- Extract letters from `content` field
- Shuffle them randomly
- For short words (< 5 letters): pad with 1-2 random distractor letters

**Hook:** `useVocabBattle` in `play-core/hooks/use-vocab-battle.ts`

**State:**
```typescript
{
  // Derived from contentItems[currentIndex]
  targetWord: string;           // e.g., "apple"
  translation: string;          // e.g., "苹果"
  phonetic: { uk: string; us: string } | null;

  // Player state
  filledLetters: string[];      // letters placed so far
  keyboardLetters: string[];    // shuffled available letters
  usedKeyIndices: Set<number>;  // which keyboard buttons are used
  hasError: boolean;            // shake animation trigger
  hadWrongAttempt: boolean;     // track for isCorrect
  isRevealed: boolean;          // word complete, showing result

  // Shields (visual only)
  playerShields: boolean[];     // [true, true, true, false, false]
  opponentShields: boolean[];

  // Opponent (competitive only)
  opponentFilledCount: number;  // how many letters opponent has filled
}
```

### 3.2 vocab-match (词汇配对)

**Progression:** Batch mode. Multiple items displayed simultaneously per round.

**Batch logic:**
- `batchSize` = number of items to show per round (capped at 5 for UX)
- `batchStart` = `currentIndex` from store
- `batchItems` = `contentItems.slice(batchStart, batchStart + batchSize)`
- When batch complete: `useGameStore.setState({ currentIndex: batchStart + batchSize })`
- If next batch start >= total items: `setPhase("result")`

**Gameplay:**
1. Display progress bar (overall items matched / total)
2. Show two columns: English words (left) and Chinese definitions (right)
3. Chinese definitions are shuffled (English column keeps original order)
4. Player clicks an English word → it highlights blue (selected)
5. Player clicks a Chinese definition:
   - Correct match → both highlight green with checkmark, `recordResult(true)` + `recordAnswerAction()`
   - Wrong match → brief shake on both, `recordResult(false)` + `recordAnswerAction()`, selection cleared
6. When all pairs in batch matched → short delay (600ms) → advance to next batch
7. Combo display shows current streak

**Competitive mode:**
- Skip button disabled (existing pattern)
- Answer/hint button disabled

**Hook:** `useVocabMatch` in `play-core/hooks/use-vocab-match.ts`

**State:**
```typescript
{
  // Batch
  batchItems: ContentItem[];
  shuffledDefs: { index: number; translation: string }[];

  // Selection
  selectedWordIndex: number | null;    // which English word is selected
  selectedDefIndex: number | null;     // which Chinese def is selected (for wrong match flash)

  // Tracking
  matchedIndices: Set<number>;         // indices within batch that are matched
  wrongPairFlash: { word: number; def: number } | null;  // brief error highlight

  // Progress
  totalMatched: number;                // across all batches
  totalItems: number;
}
```

### 3.3 vocab-elimination (词汇消消乐)

**Progression:** Batch mode. Grid of tiles displayed per round.

**Batch logic:**
- `batchSize` = number of pairs per round (capped at 8 for 4×4 grid)
- Same batch advancement as vocab-match
- Grid = `batchSize * 2` tiles (each pair = 1 English + 1 Chinese tile)
- Tiles shuffled into a flat array, arranged in 4-column rows

**Gameplay:**
1. Display status row: eliminated count, progress bar, combo badge
2. Show grid of tiles (4 columns, rows = ceil(batchSize * 2 / 4))
3. Each tile shows either an English word or Chinese definition
4. Player clicks first tile → it highlights (selected, pink border)
5. Player clicks second tile:
   - If they form a correct English-Chinese pair → both fade out (eliminated), `recordResult(true)` + `recordAnswerAction()`
   - If not a match → brief shake, selection cleared, `recordResult(false)` + `recordAnswerAction()`
   - If same tile clicked again → deselect
6. When all pairs eliminated → short delay (600ms) → advance to next batch
7. Combo display with pink accent

**Tile generation:**
- For each batch item: create 2 tiles — `{ type: "en", text: content, itemIndex: i }` and `{ type: "zh", text: translation, itemIndex: i }`
- Shuffle all tiles randomly
- Correct pair = tiles sharing the same `itemIndex` but different `type`

**Hook:** `useVocabElimination` in `play-core/hooks/use-vocab-elimination.ts`

**State:**
```typescript
{
  // Grid
  tiles: { id: string; type: "en" | "zh"; text: string; itemIndex: number }[];
  columns: number;  // always 4

  // Selection
  selectedTileId: string | null;
  wrongPairFlash: { tile1: string; tile2: string } | null;

  // Tracking
  eliminatedItemIndices: Set<number>;
  totalEliminated: number;     // across all batches
  totalPairs: number;
}
```

---

## 4. Shared Patterns Across All Three Hooks

Each hook follows the `useWordSentence` contract:

1. **Read from `useGameStore`:** `contentItems`, `currentIndex`, `sessionId`, `levelId`, `gameId`, `recordResult`, `recordSkip`, `nextItem`, `setPhase`
2. **Read from `useGamePlayActions`:** `recordAnswer`, `recordSkip`, `markAsReview`, `competitive`
3. **Fire-and-forget server sync:** `recordAnswerAction()` with full payload (sessionId, levelId, contentItemId, isCorrect, userAnswer, sourceAnswer, baseScore, comboScore, score, maxCombo, playTime, nextContentItemId, duration)
4. **Score derivation:** Capture `prevScore` before `recordResult()`, compute `baseScore` and `comboScore` from delta (same pattern as word-sentence)
5. **Duration tracking:** `itemStartTimeRef` for per-item timing
6. **Deduped tracking:** `markAsReviewAction()` for incorrect items via ref-based dedup
7. **Cleanup:** Clear timers on unmount, reset state on `currentIndex` change

---

## 5. Files to Create

| File | Purpose |
|------|---------|
| `play-core/hooks/use-vocab-battle.ts` | Per-item hook: letter selection, shield tracking, keyboard generation |
| `play-core/hooks/use-vocab-match.ts` | Batch hook: pair selection, match validation, batch advancement |
| `play-core/hooks/use-vocab-elimination.ts` | Batch hook: tile grid, pair elimination, batch advancement |

## 6. Files to Rewrite

| File | Change |
|------|--------|
| `play-core/components/game-vocab-battle.tsx` | Replace static mockup with data-driven component using `useVocabBattle` |
| `play-core/components/game-vocab-match.tsx` | Replace static mockup with data-driven component using `useVocabMatch` |
| `play-core/components/game-vocab-elimination.tsx` | Replace static mockup with data-driven component using `useVocabElimination` |

## 7. Files NOT Modified

- `play-core/hooks/use-game-store.ts` — no store changes needed
- `play-core/context/game-play-context.tsx` — no context changes
- `play-single/components/game-play-shell.tsx` — already wired
- `play-pk/components/pk-play-shell.tsx` — already wired
- `play-group/components/group-play-shell.tsx` — already wired
- All backend Go files — no changes needed
- `game-word-sentence.tsx` — untouched
- All existing hooks — untouched

---

## 8. Component UI Specifications

### 8.1 vocab-battle Component Structure

```
Card (rounded-[20px], border, bg-card)
├── Opponent zone (px-6 py-7) [opacity-40 when !competitive]
│   ├── Label: "🤖 对手"
│   ├── Shield row: 5 circles (red-400 active, muted inactive)
│   └── Letter slots: squares showing opponent progress
├── Translation zone (gradient red→teal divider)
│   ├── Chinese translation (text-2xl font-extrabold)
│   └── Gradient divider line
├── Player zone (px-6 py-5)
│   ├── Letter slots: squares filling as player types (teal-50 filled, muted empty)
│   ├── Shield row: 5 circles (teal-400 active, muted inactive)
│   └── Label: "🎯 我"
├── Combo row (conditional, when streak >= 3)
└── Letter keyboard (dark buttons, grid)
```

### 8.2 vocab-match Component Structure

```
Card (rounded-[20px], border, bg-card, p-6)
├── Progress section
│   ├── "进度 N/M" label
│   ├── Combo badge (teal)
│   └── Progress bar (gradient blue→teal)
├── Match area (flex-row on sm+, flex-col on mobile)
│   ├── English column
│   │   ├── "英文单词" header
│   │   └── Word buttons (emerald=matched, blue=selected, default)
│   └── Chinese column
│       ├── "中文释义" header
│       └── Definition buttons (emerald=matched, default)
└── Hint text: "点击左侧单词，再点击右侧匹配的释义"
```

### 8.3 vocab-elimination Component Structure

```
Wrapper (flex-col, gap-5)
├── Status row
│   ├── "已消除 N/M 对" label
│   ├── Progress bar (gradient pink→teal)
│   └── Combo badge (pink)
├── Grid card (rounded-[20px], border, bg-card, p-4)
│   └── 4-column grid of tile buttons
│       - default: border-border bg-card
│       - selected: border-2 border-pink-500 bg-pink-50
│       - eliminated: opacity-40 bg-muted line-through
└── Hint text: "点击两个匹配的方块进行消除"
```

---

## 9. Constraints

- **No lint errors** — all components must pass ESLint with existing config
- **No breaking changes** — word-sentence and all existing functionality untouched
- **No store modifications** — use Zustand `setState` directly for batch advancement
- **No backend changes** — frontend-only implementation
- **Responsive** — mobile-first with sm/md breakpoints (same as existing components)
- **Competitive mode** — respect `competitive` flag consistently across all modes

# AI Custom Vocab Mode Support

## Overview

Add vocabulary mode support to the ai-custom workflow for the three vocab game modes: vocab-battle (词汇对轰), vocab-match (词汇配对), and vocab-elimination (词汇消消乐). These modes only accept English-Chinese vocabulary pairs — no sentences.

Full split from the existing word-sentence (连词成句) ai-custom feature: separate backend service, separate frontend feature directory, separate routes.

## Game Mode Specifications

| | vocab-match | vocab-elimination | vocab-battle |
|---|---|---|---|
| AI generates per click | 5 pairs | 8 pairs | 20 pairs |
| 1 content meta = | 1 English-Chinese pair | 1 English-Chinese pair | 1 English-Chinese pair |
| Max metas per level | 20 | 20 | 20 |
| Max levels per game | 20 | 20 | 20 |
| Format button | 词汇检查并格式化 only | 词汇检查并格式化 only | 词汇检查并格式化 only |
| Sentences allowed | No | No | No |

Word-sentence mode keeps its existing limits (MaxSentences=20, MaxVocab=200 ratio) unchanged.

---

## Backend Design

### New Files

| File | Purpose |
|------|---------|
| `dx-api/app/services/api/ai_custom_vocab_service.go` | All vocab AI logic |
| `dx-api/app/http/controllers/api/ai_custom_vocab_controller.go` | Vocab AI endpoints |
| `dx-api/app/http/requests/api/ai_custom_vocab_request.go` | Request validation |
| `dx-api/app/consts/ai_custom_vocab.go` | Vocab-specific constants |

### Service Functions (`ai_custom_vocab_service.go`)

#### `GenerateVocab(userID, difficulty, keywords, gameMode) -> (*GenerateVocabResult, error)`
- AI generates vocabulary pairs from keywords using DeepSeek
- Count determined by game mode: match=5, elimination=8, battle=20
- Cost: 5 beans (fixed, same as story generation)
- Prompt: `buildVocabGeneratePrompt(levelDesc, count)` — generates `count` English-Chinese pairs
- Output format: alternating lines (English\nChinese\nEnglish\nChinese...)
- Includes content moderation check
- Returns `sourceType: "vocab"` always
- Refunds on AI failure

#### `FormatVocab(userID, content) -> (*FormatVocabResult, error)`
- AI checks and formats raw vocab text
- Cost: word count of input
- Prompt: `buildVocabFormatPrompt()` — only `[V]` markers, rejects sentences
- If content looks like sentences, returns warning: "内容看起来是语句而非词汇"
- No `formatType` parameter needed (always vocab)
- Parsing: `parseVocabFormattedLines()` — strips `[V]` prefixes, rejects any `[S]` lines
- Refunds on AI failure

#### `BreakVocabMetadata(userID, gameLevelID, writer *SSEWriter)`
- Simpler than sentence break — each vocab meta produces 1 content item
- Single word: `contentType: "word"` / Multi-word: `contentType: "phrase"`
- Translation passed through from meta (AI only called if translation missing)
- Creates GameItem junction record for each content item
- Marks meta as `is_break_done = true`
- SSE streaming with progress events
- Cost: word count of all unbroken metas
- Concurrency: same semaphore pattern as word-sentence

#### `GenerateVocabContentItems(userID, gameLevelID, writer *SSEWriter)`
- Adds phonetics/POS/translation items JSON to content items
- Same logic as existing `GenerateContentItems` — word-level breakdown is identical
- Uses same `genItemsPrompt` (duplicated into vocab service for independence)
- SSE streaming with progress events
- Cost: word count of all items needing generation
- Concurrency: same semaphore pattern

#### Helper Functions (duplicated for independence)
- `verifyVocabLevelOwnership(userID, gameLevelID)` — validates game mode is one of 3 vocab modes
- `writeVocabSSEError(writer, err)` — maps errors to Chinese SSE messages
- `vocabGenerateCount(gameMode)` — returns 5/8/20 based on mode

### AI Prompts

#### `buildVocabGeneratePrompt(levelDesc, count)`
- System prompt for generating `count` English-Chinese vocabulary pairs
- CEFR level-appropriate vocabulary
- Moderation check (same as story generation)
- Output: alternating lines, English then Chinese
- No sentences, no stories — only single words or short phrases

#### `buildVocabFormatPrompt()`
- System prompt for formatting messy vocab input
- Only `[V]` markers — no `[S]` allowed
- Type mismatch check: warns if content looks like sentences
- Fixes spelling, removes duplicates, ensures paired format

#### `vocabBreakPrompt`
- AI determines contentType: "word" (single word) or "phrase" (multi-word)
- Returns 1 item per vocab entry with translation
- Much simpler than sentence breakPrompt (no block/sentence types)

### Routes

```
POST /api/ai-custom-vocab/generate-vocab
POST /api/ai-custom-vocab/format-vocab
POST /api/ai-custom-vocab/break-metadata
POST /api/ai-custom-vocab/generate-content-items
```

Added to `routes/api.go` under a new route group with user JWT auth middleware.

### Controller (`ai_custom_vocab_controller.go`)

| Method | Route | Request Fields |
|--------|-------|----------------|
| `GenerateVocab` | POST /generate-vocab | `difficulty`, `keywords[]`, `gameMode` |
| `FormatVocab` | POST /format-vocab | `content` |
| `BreakMetadata` | POST /break-metadata | `gameLevelId` |
| `GenerateContentItems` | POST /generate-content-items | `gameLevelId` |

### Request Validation (`ai_custom_vocab_request.go`)

- `GenerateVocabRequest`: keywords (1-5), difficulty (a1-a2/b1-b2/c1-c2), gameMode (must be vocab-battle/match/elimination)
- `FormatVocabRequest`: content (non-empty)
- `BreakVocabMetadataRequest`: gameLevelId (UUID)
- `GenerateVocabContentItemsRequest`: gameLevelId (UUID)

### Constants (`ai_custom_vocab.go`)

```go
const (
    MaxMetasPerLevel      = 20
    MaxLevelsPerGame      = 20
    VocabMatchCount       = 5
    VocabEliminationCount = 8
    VocabBattleCount      = 20
)
```

### Bean Slugs (new entries in `bean_slug.go` / `bean_reason.go`)

```
ai-vocab-generate-consume / ai-vocab-generate-refund
ai-vocab-format-consume / ai-vocab-format-refund
ai-vocab-break-consume / ai-vocab-break-refund
ai-vocab-gen-items-consume / ai-vocab-gen-items-refund
```

### `SaveMetadataBatch` Update

Add mode-aware capacity check in `course_content_service.go`:

```go
if consts.IsVocabMode(game.Mode) {
    // Per-level limit: max 20 metas total
    if existingCount + len(entries) > consts.MaxMetasPerLevel {
        return error
    }
    // Per-submission limit based on mode
    maxPerSubmission := consts.VocabGenerateCount(game.Mode)
    if len(entries) > maxPerSubmission {
        return error
    }
} else {
    // existing sentence/vocab ratio check (unchanged)
}
```

`IsVocabMode(mode string) bool` — new helper in `consts/game_mode.go`, returns true for vocab-battle/match/elimination.

---

## Frontend Design

### New Feature Directory

```
dx-web/src/features/web/ai-custom-vocab/
├── actions/
│   └── course-game.action.ts
├── components/
│   ├── ai-custom-vocab-grid.tsx
│   ├── add-level-dialog.tsx
│   ├── add-vocab-dialog.tsx
│   ├── vocab-ai-tab.tsx
│   ├── vocab-manual-tab.tsx
│   ├── course-detail-content.tsx
│   ├── create-course-form.tsx
│   ├── edit-game-dialog.tsx
│   ├── game-card-item.tsx
│   ├── game-hero-card.tsx
│   ├── game-info-card.tsx
│   ├── game-levels-card.tsx
│   ├── level-units-panel.tsx
│   ├── processing-overlay.tsx
│   ├── sortable-content-item.tsx
│   └── sortable-meta-item.tsx
├── helpers/
│   ├── count-words.ts
│   ├── format-api.ts
│   ├── format-metadata.ts
│   ├── generate-api.ts
│   ├── generate-items-api.ts
│   └── stream-progress.ts
├── hooks/
│   ├── use-create-course-game.ts
│   ├── use-create-game-level.ts
│   ├── use-game-actions.ts
│   ├── use-infinite-games.ts
│   └── use-update-course-game.ts
└── schemas/
    └── course-game.schema.ts
```

### App Routes (new)

```
dx-web/src/app/(web)/hall/(main)/ai-custom-vocab/page.tsx
dx-web/src/app/(web)/hall/(main)/ai-custom-vocab/[id]/page.tsx
dx-web/src/app/(web)/hall/(main)/ai-custom-vocab/[id]/[levelId]/page.tsx
```

### `add-vocab-dialog.tsx`

- Two tabs: 手动添加 / AI 生成
- Footer: `[ 词汇检查并格式化 ] [ 保存 ]` — no sentence format button
- Receives `gameMode` prop to determine max pairs per submission
- Save validates paired input and mode-specific max count
- `handleBeanError` same as existing

### `vocab-manual-tab.tsx`

Validation rules:
- Every English line must have a Chinese translation on the next line (paired input required)
- Max entries per submission based on mode: match=5, elimination=8, battle=20
- Capacity check: `existingMetaCount + newPairs <= 20`
- Only vocab format example HoverCard (no sentence example)
- Textarea placeholder: "输入词汇，每行一条，英文一行、中文下一行..."

Format example (HoverCard):
```
apple
苹果
banana
香蕉
polar bear
北极熊
```

### `vocab-ai-tab.tsx`

- Difficulty selector (same A1-A2, B1-B2, C1-C2)
- Keywords input (max 5 words, space-separated)
- Preview shows generated English-Chinese pairs
- Stats: 词汇对数 (pair count, not sentence/word count)
- API: `generateVocab(difficulty, keywords, gameMode)` → `/api/ai-custom-vocab/generate-vocab`
- Generated format: alternating English\nChinese lines
- "使用" imports to vocab manual tab

### `parseVocabText(text, maxPairs)` — New parsing helper

```typescript
type VocabPair = { english: string; chinese: string };
type ParseResult = { pairs: VocabPair[]; error?: string };
```

- Parses alternating English/Chinese lines into pairs
- Error if odd number of non-empty lines (unpaired entry)
- Error if pairs exceed `maxPairs` (5/8/20 based on mode)
- Returns structured pairs for save action

### `format-metadata.ts` — Vocab version

- `MAX_METAS_PER_LEVEL = 20` (replaces MaxSentences/MaxVocab)
- `parseVocabText()` as described above
- No sentence-related constants

### `format-api.ts` — Vocab version

- Calls `/api/ai-custom-vocab/format-vocab`
- Only sends `content`, no `formatType` parameter

### `generate-api.ts` — Vocab version

- `generateVocab(difficulty, keywords, gameMode)`
- Calls `/api/ai-custom-vocab/generate-vocab`
- Returns `{ ok: true, generated: string, sourceType: "vocab" }`

### `generate-items-api.ts` — Vocab version

- `breakVocabMetadata(levelId, signal, onProgress)` → `/api/ai-custom-vocab/break-metadata`
- `generateVocabContentItems(levelId, signal, onProgress)` → `/api/ai-custom-vocab/generate-content-items`

### `level-units-panel.tsx` — Vocab version

- Stats bar: 共计 + 练习单元总数 (no sentence/vocab split)
- Capacity: `metaCount >= 20` (flat limit)
- Add button tooltip: `N/20 元数据`
- Renders `<AddVocabDialog gameMode={gameMode}>` instead of `<AddMetadataDialog>`
- All other functionality identical (drag-and-drop, break, generate, delete)

### `create-course-form.tsx` — Vocab version

- Mode selector shows only: 词汇对轰, 词汇配对, 词汇消消乐
- Word-sentence not selectable

### Existing `ai-custom` Changes

Only change: `create-course-form.tsx` filters mode selector to show only word-sentence. Vocab modes removed from its options. All other files untouched.

---

## What Stays the Same

- Database schema: no new tables, no migrations
- Models: `ContentMeta`, `ContentItem`, `Game`, `GameLevel`, `GameMeta`, `GameItem` — unchanged
- Game CRUD endpoints: `/api/course-games/*` — shared by both features
- Shared UI components: `@/components/ui/*`, `@/components/in/*` — untouched
- Constants: `GAME_MODES`, `SOURCE_TYPES`, `DIFFICULTY_OPTIONS` — unchanged
- Docker compose: no changes

## What Must Not Break

- Existing word-sentence ai-custom workflow (zero changes to `ai_custom_service.go`)
- Existing game CRUD (course_game_controller, course_game_service)
- Published game validation and publish/withdraw flow
- Bean consumption and refund logic for word-sentence
- SSE streaming for word-sentence break/generate
- All existing routes and API contracts

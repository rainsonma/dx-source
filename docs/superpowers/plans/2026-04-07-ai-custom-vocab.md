# AI Custom Vocab Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add full vocab mode support (vocab-battle, vocab-match, vocab-elimination) to the ai-custom workflow with completely separate backend service and frontend feature directory.

**Architecture:** Full split — new `ai_custom_vocab_service.go` in backend, new `dx-web/src/features/web/ai-custom-vocab/` in frontend, new API routes under `/api/ai-custom-vocab/`, new app routes under `/hall/ai-custom-vocab/`. Shared game CRUD endpoints (`/api/course-games/*`) remain unchanged. No database migrations needed.

**Tech Stack:** Go/Goravel, Next.js 16, PostgreSQL, DeepSeek AI, SSE streaming, SWR, dnd-kit, shadcn/ui, Zod

**Spec:** `docs/superpowers/specs/2026-04-07-ai-custom-vocab-design.md`

---

## File Map

### Backend — New Files
| File | Purpose |
|------|---------|
| `dx-api/app/consts/ai_custom_vocab.go` | Vocab constants (MaxMetasPerLevel, counts, IsVocabMode) |
| `dx-api/app/services/api/ai_custom_vocab_service.go` | All vocab AI service logic |
| `dx-api/app/http/controllers/api/ai_custom_vocab_controller.go` | Vocab AI controller |

### Backend — Modified Files
| File | Change |
|------|--------|
| `dx-api/app/consts/bean_slug.go` | Add 8 vocab bean slug constants |
| `dx-api/app/consts/bean_reason.go` | Add 8 vocab bean reason constants |
| `dx-api/routes/api.go` | Add `/api/ai-custom-vocab/*` route group |
| `dx-api/app/services/api/course_content_service.go` | Mode-aware capacity check in SaveMetadataBatch |

### Frontend — New Feature Directory
| File | Source (copy from ai-custom, then modify) |
|------|-------------------------------------------|
| `dx-web/src/features/web/ai-custom-vocab/helpers/count-words.ts` | Direct copy |
| `dx-web/src/features/web/ai-custom-vocab/helpers/stream-progress.ts` | Direct copy |
| `dx-web/src/features/web/ai-custom-vocab/helpers/format-metadata.ts` | New (vocab-specific parsing) |
| `dx-web/src/features/web/ai-custom-vocab/helpers/generate-api.ts` | New (calls vocab endpoint) |
| `dx-web/src/features/web/ai-custom-vocab/helpers/format-api.ts` | New (calls vocab endpoint) |
| `dx-web/src/features/web/ai-custom-vocab/helpers/generate-items-api.ts` | New (calls vocab endpoints) |
| `dx-web/src/features/web/ai-custom-vocab/schemas/course-game.schema.ts` | Copy + modify validation |
| `dx-web/src/features/web/ai-custom-vocab/actions/course-game.action.ts` | Copy + adjust imports |
| `dx-web/src/features/web/ai-custom-vocab/hooks/use-create-course-game.ts` | Copy + adjust imports |
| `dx-web/src/features/web/ai-custom-vocab/hooks/use-create-game-level.ts` | Copy |
| `dx-web/src/features/web/ai-custom-vocab/hooks/use-game-actions.ts` | Copy + adjust route |
| `dx-web/src/features/web/ai-custom-vocab/hooks/use-infinite-games.ts` | Copy + adjust filter |
| `dx-web/src/features/web/ai-custom-vocab/hooks/use-update-course-game.ts` | Copy + adjust imports |
| `dx-web/src/features/web/ai-custom-vocab/components/vocab-manual-tab.tsx` | New (vocab-only UI) |
| `dx-web/src/features/web/ai-custom-vocab/components/vocab-ai-tab.tsx` | New (vocab generation) |
| `dx-web/src/features/web/ai-custom-vocab/components/add-vocab-dialog.tsx` | New (vocab dialog) |
| `dx-web/src/features/web/ai-custom-vocab/components/level-units-panel.tsx` | Copy + modify stats/dialog |
| `dx-web/src/features/web/ai-custom-vocab/components/ai-custom-vocab-grid.tsx` | Copy from ai-custom-grid + modify |
| `dx-web/src/features/web/ai-custom-vocab/components/create-course-form.tsx` | Copy + vocab modes only |
| `dx-web/src/features/web/ai-custom-vocab/components/course-detail-content.tsx` | Copy + adjust imports/route |
| `dx-web/src/features/web/ai-custom-vocab/components/edit-game-dialog.tsx` | Copy + vocab modes only |
| `dx-web/src/features/web/ai-custom-vocab/components/game-card-item.tsx` | Copy + adjust route |
| `dx-web/src/features/web/ai-custom-vocab/components/game-hero-card.tsx` | Copy + adjust route |
| `dx-web/src/features/web/ai-custom-vocab/components/game-info-card.tsx` | Copy |
| `dx-web/src/features/web/ai-custom-vocab/components/game-levels-card.tsx` | Copy + adjust route |
| `dx-web/src/features/web/ai-custom-vocab/components/add-level-dialog.tsx` | Copy |
| `dx-web/src/features/web/ai-custom-vocab/components/sortable-meta-item.tsx` | Copy |
| `dx-web/src/features/web/ai-custom-vocab/components/sortable-content-item.tsx` | Copy |
| `dx-web/src/features/web/ai-custom-vocab/components/processing-overlay.tsx` | Copy |

### Frontend — New App Routes
| File | Purpose |
|------|---------|
| `dx-web/src/app/(web)/hall/(main)/ai-custom-vocab/page.tsx` | Landing page |
| `dx-web/src/app/(web)/hall/(main)/ai-custom-vocab/[id]/page.tsx` | Course detail |
| `dx-web/src/app/(web)/hall/(main)/ai-custom-vocab/[id]/[levelId]/page.tsx` | Level editor |

### Frontend — Modified Files
| File | Change |
|------|--------|
| `dx-web/src/features/web/ai-custom/components/create-course-form.tsx` | Filter to word-sentence only |

---

## Task 1: Backend Constants

**Files:**
- Create: `dx-api/app/consts/ai_custom_vocab.go`
- Modify: `dx-api/app/consts/bean_slug.go`
- Modify: `dx-api/app/consts/bean_reason.go`

- [ ] **Step 1: Create vocab constants file**

```go
// dx-api/app/consts/ai_custom_vocab.go
package consts

// Vocab mode limits.
const (
	MaxMetasPerLevel      = 20
	MaxLevelsPerGame      = 20
	VocabMatchCount       = 5
	VocabEliminationCount = 8
	VocabBattleCount      = 20
)

// VocabGenerateCount returns how many pairs to generate based on game mode.
func VocabGenerateCount(mode string) int {
	switch mode {
	case GameModeVocabMatch:
		return VocabMatchCount
	case GameModeVocabElimination:
		return VocabEliminationCount
	case GameModeVocabBattle:
		return VocabBattleCount
	default:
		return VocabMatchCount
	}
}

// IsVocabMode returns true if the mode is one of the three vocab game modes.
func IsVocabMode(mode string) bool {
	return mode == GameModeVocabBattle || mode == GameModeVocabMatch || mode == GameModeVocabElimination
}
```

- [ ] **Step 2: Add vocab bean slugs**

Add to `dx-api/app/consts/bean_slug.go` after `BeanSlugAIGenItemsRefund`:

```go
	BeanSlugAIVocabGenerateConsume = "ai-vocab-generate-consume"
	BeanSlugAIVocabGenerateRefund  = "ai-vocab-generate-refund"
	BeanSlugAIVocabFormatConsume   = "ai-vocab-format-consume"
	BeanSlugAIVocabFormatRefund    = "ai-vocab-format-refund"
	BeanSlugAIVocabBreakConsume    = "ai-vocab-break-consume"
	BeanSlugAIVocabBreakRefund     = "ai-vocab-break-refund"
	BeanSlugAIVocabGenItemsConsume = "ai-vocab-gen-items-consume"
	BeanSlugAIVocabGenItemsRefund  = "ai-vocab-gen-items-refund"
```

Add to `BeanSlugLabels` map:

```go
	BeanSlugAIVocabGenerateConsume: "词汇 AI 生成消耗",
	BeanSlugAIVocabGenerateRefund:  "词汇 AI 生成失败退还",
	BeanSlugAIVocabFormatConsume:   "词汇格式化消耗",
	BeanSlugAIVocabFormatRefund:    "词汇格式化失败退还",
	BeanSlugAIVocabBreakConsume:    "词汇分解消耗",
	BeanSlugAIVocabBreakRefund:     "词汇分解失败退还",
	BeanSlugAIVocabGenItemsConsume: "词汇生成消耗",
	BeanSlugAIVocabGenItemsRefund:  "词汇生成失败退还",
```

- [ ] **Step 3: Add vocab bean reasons**

Add to `dx-api/app/consts/bean_reason.go` after `BeanReasonAIGenItemsRefund`:

```go
	BeanReasonAIVocabGenerateConsume = "词汇 AI 生成消耗"
	BeanReasonAIVocabGenerateRefund  = "词汇 AI 生成失败退还"
	BeanReasonAIVocabFormatConsume   = "词汇格式化消耗"
	BeanReasonAIVocabFormatRefund    = "词汇格式化失败退还"
	BeanReasonAIVocabBreakConsume    = "词汇分解消耗"
	BeanReasonAIVocabBreakRefund     = "词汇分解失败退还"
	BeanReasonAIVocabGenItemsConsume = "词汇生成消耗"
	BeanReasonAIVocabGenItemsRefund  = "词汇生成失败退还"
```

- [ ] **Step 4: Verify compilation**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: No errors

---

## Task 2: Backend Vocab Service — GenerateVocab + FormatVocab

**Files:**
- Create: `dx-api/app/services/api/ai_custom_vocab_service.go`

- [ ] **Step 1: Create vocab service with GenerateVocab and FormatVocab**

```go
// dx-api/app/services/api/ai_custom_vocab_service.go
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
)

// Concurrency limits for vocab SSE batch operations.
const (
	vocabBreakConcurrencyLimit    = 20
	vocabGenItemsConcurrencyLimit = 50
)

// GenerateVocabResult holds the response from vocab generation.
type GenerateVocabResult struct {
	Generated  string `json:"generated,omitempty"`
	SourceType string `json:"sourceType,omitempty"`
	Warning    string `json:"warning,omitempty"`
}

// FormatVocabResult holds the response from vocab formatting.
type FormatVocabResult struct {
	Formatted string `json:"formatted,omitempty"`
	Warning   string `json:"warning,omitempty"`
}

// --- GenerateVocab ---

// GenerateVocab generates English-Chinese vocabulary pairs from keywords using DeepSeek AI.
// Count is determined by game mode. Consumes 5 beans. Refunds on AI failure.
func GenerateVocab(userID, difficulty string, keywords []string, gameMode string) (*GenerateVocabResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	if err := ConsumeBeans(userID, aiGenerateCost, consts.BeanSlugAIVocabGenerateConsume, consts.BeanReasonAIVocabGenerateConsume); err != nil {
		return nil, err
	}

	levelDesc, ok := difficultyDescriptions[difficulty]
	if !ok {
		levelDesc = difficultyDescriptions["a1-a2"]
	}

	count := consts.VocabGenerateCount(gameMode)
	prompt := buildVocabGeneratePrompt(levelDesc, count)
	userMsg := "Keywords: " + strings.Join(keywords, ", ")

	result, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
		Messages: []helpers.DeepSeekMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: userMsg},
		},
		Temperature: 0.7,
	})
	if err != nil {
		_ = RefundBeans(userID, aiGenerateCost, consts.BeanSlugAIVocabGenerateRefund, consts.BeanReasonAIVocabGenerateRefund)
		return nil, err
	}

	if strings.HasPrefix(result, "WARNING:") {
		warning := strings.TrimSpace(strings.TrimPrefix(result, "WARNING:"))
		return &GenerateVocabResult{Warning: warning}, nil
	}

	return &GenerateVocabResult{
		Generated:  result,
		SourceType: SourceTypeVocab,
	}, nil
}

func buildVocabGeneratePrompt(levelDesc string, count int) string {
	return fmt.Sprintf(`You are a vocabulary generator for an English learning app. Your job is to generate English-Chinese vocabulary pairs for language learners.

STEP 1 — CONTENT MODERATION (do this FIRST):
Check if the provided keywords contain any insulting, violent, sexually explicit, or otherwise inappropriate/sensitive material.
If they do, respond ONLY with: WARNING:包含不适当内容，请修改后重试
Do NOT generate any vocabulary. Stop here.

STEP 2 — GENERATE VOCABULARY:
Generate exactly %d English-Chinese vocabulary pairs that:
- Use CEFR level %s appropriate vocabulary
- Are thematically related to the provided keywords
- Include single words or short phrases (NOT full sentences)
- Have accurate Chinese translations

OUTPUT FORMAT:
Output alternating lines: English word/phrase on one line, Chinese translation on the next line.
No numbering, no prefixes, no markdown, no explanations, no empty lines.

Example output:
apple
苹果
banana
香蕉
polar bear
北极熊

RULES:
- Output ONLY the vocabulary pairs in alternating lines.
- Each English entry must be a single word or short phrase (2-3 words max).
- Do not repeat vocabulary.
- Do not include full sentences.
- Exactly %d pairs, no more, no fewer.`, count, levelDesc, count)
}

// --- FormatVocab ---

// FormatVocab formats raw user text into structured vocabulary content.
// Bean cost = word count of input.
func FormatVocab(userID, content string) (*FormatVocabResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	wordCount := helpers.CountWords(content)
	if wordCount == 0 {
		return nil, ErrEmptyContent
	}

	if err := ConsumeBeans(userID, wordCount, consts.BeanSlugAIVocabFormatConsume, consts.BeanReasonAIVocabFormatConsume); err != nil {
		return nil, err
	}

	result, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
		Messages: []helpers.DeepSeekMessage{
			{Role: "system", Content: vocabFormatPrompt},
			{Role: "user", Content: content},
		},
		Temperature: 0.1,
	})
	if err != nil {
		_ = RefundBeans(userID, wordCount, consts.BeanSlugAIVocabFormatRefund, consts.BeanReasonAIVocabFormatRefund)
		return nil, err
	}

	if strings.HasPrefix(result, "WARNING:") {
		warning := strings.TrimSpace(strings.TrimPrefix(result, "WARNING:"))
		return &FormatVocabResult{Warning: warning}, nil
	}

	formatted := cleanVocabFormatted(result)
	if formatted == "" {
		_ = RefundBeans(userID, wordCount, consts.BeanSlugAIVocabFormatRefund, consts.BeanReasonAIVocabFormatRefund)
		return nil, helpers.ErrDeepSeekEmpty
	}

	return &FormatVocabResult{Formatted: formatted}, nil
}

var vocabFormatPrompt = `You are a content formatter for an English learning app. Your job is to clean up and reformat messy user input into strict vocabulary pairs.

STEP 1 — CONTENT MODERATION (do this FIRST):
Check if the content contains any insulting, violent, sexually explicit, or otherwise inappropriate/sensitive material.
If it does, respond ONLY with: WARNING:内容包含不适当内容，请修改后重试
Do NOT format the content. Stop here.

STEP 2 — TYPE MISMATCH CHECK:
If the content consists mostly of full sentences with punctuation (subject + verb + object), respond ONLY with: WARNING:内容看起来是语句而非词汇，请前往「连词成句」模式添加语句内容
Do NOT format the content. Stop here.

STEP 3 — FORMAT VOCABULARY:
Clean up and reformat the input into strict alternating lines:
- English word/phrase on one line
- Chinese translation on the next line
- Repeat for each pair

RULES:
- Output ONLY alternating English/Chinese lines. No prefixes, no markers, no numbering, no markdown, no explanations.
- Each English entry must be a single word or short phrase (NOT a full sentence).
- Fix obvious spelling errors in English.
- Remove duplicates.
- Remove empty lines.
- If the input has English without Chinese translations, add accurate Chinese translations.
- If the input has Chinese without English, add accurate English translations.
- Preserve the original meaning. Do not add or remove vocabulary items beyond fixing errors.`

func cleanVocabFormatted(result string) string {
	lines := strings.Split(result, "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimRight(line, " \t\r")
		// Strip any [V] or [S] prefixes the AI might add despite instructions
		if strings.HasPrefix(line, "[V] ") {
			line = line[4:]
		} else if strings.HasPrefix(line, "[S] ") {
			line = line[4:]
		}
		if line == "" {
			continue
		}
		cleanLines = append(cleanLines, line)
	}

	return strings.Join(cleanLines, "\n")
}

// --- BreakVocabMetadata ---

// BreakVocabMetadata processes vocab content metas: creates 1 content item per meta via SSE.
func BreakVocabMetadata(userID, gameLevelID string, writer *helpers.SSEWriter) {
	if err := requireVip(userID); err != nil {
		writeVocabSSEError(writer, err)
		return
	}
	game, level, err := verifyVocabLevelOwnership(userID, gameLevelID)
	if err != nil {
		writeVocabSSEError(writer, err)
		return
	}
	if game.Status == consts.GameStatusPublished {
		writeVocabSSEError(writer, ErrGamePublished)
		return
	}
	_ = level
	gameID := game.ID

	// Fetch unbroken metas
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
		Where("gm.game_level_id", gameLevelID).
		Where("content_metas.is_break_done", false).
		Order("content_metas.\"order\" ASC").
		Get(&metas); err != nil {
		writeVocabSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
		return
	}

	if len(metas) == 0 {
		_ = writer.Write(SSEProgressEvent{Done: 0, Total: 0, Processed: 0, Failed: 0, Complete: true})
		writer.Close()
		return
	}

	// Calculate bean cost per meta (word count)
	metaWordCounts := make([]int, len(metas))
	totalCost := 0
	for i, m := range metas {
		wc := helpers.CountWords(m.SourceData)
		metaWordCounts[i] = wc
		totalCost += wc
	}

	if totalCost == 0 {
		writeVocabSSEError(writer, ErrEmptyContent)
		return
	}

	if err := ConsumeBeans(userID, totalCost, consts.BeanSlugAIVocabBreakConsume, consts.BeanReasonAIVocabBreakConsume); err != nil {
		writeVocabSSEError(writer, err)
		return
	}

	var failedWords int64
	var processed int64
	var failed int64

	sem := make(chan struct{}, vocabBreakConcurrencyLimit)
	var wg sync.WaitGroup
	var done int64

	total := len(metas)

	for i, meta := range metas {
		wg.Add(1)
		sem <- struct{}{}

		go func(m models.ContentMeta, idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			success := processVocabBreakMeta(m, gameID, gameLevelID)
			d := atomic.AddInt64(&done, 1)

			if success {
				atomic.AddInt64(&processed, 1)
				_ = writer.Write(SSEProgressEvent{Done: int(d), Total: total, Status: "ok"})
			} else {
				atomic.AddInt64(&failed, 1)
				atomic.AddInt64(&failedWords, int64(metaWordCounts[idx]))
				_ = writer.Write(SSEProgressEvent{Done: int(d), Total: total, Status: "failed"})
			}
		}(meta, i)
	}

	wg.Wait()

	fw := int(atomic.LoadInt64(&failedWords))
	if fw > 0 {
		_ = RefundBeans(userID, fw, consts.BeanSlugAIVocabBreakRefund, consts.BeanReasonAIVocabBreakRefund)
	}

	_ = writer.Write(SSEProgressEvent{
		Done:      total,
		Total:     total,
		Processed: int(atomic.LoadInt64(&processed)),
		Failed:    int(atomic.LoadInt64(&failed)),
		Complete:  true,
	})
	writer.Close()
}

func processVocabBreakMeta(meta models.ContentMeta, gameID, gameLevelID string) bool {
	// For vocab, each meta becomes exactly 1 content item.
	// Determine contentType: single word or phrase.
	contentType := consts.ContentTypeWord
	if strings.Contains(strings.TrimSpace(meta.SourceData), " ") {
		contentType = consts.ContentTypePhrase
	}

	id := uuid.Must(uuid.NewV7()).String()
	metaID := meta.ID

	item := models.ContentItem{
		ID:            id,
		ContentMetaID: &metaID,
		Content:       meta.SourceData,
		ContentType:   contentType,
		Translation:   meta.Translation,
		Order:         meta.Order,
		IsActive:      true,
	}
	if err := facades.Orm().Query().Create(&item); err != nil {
		return false
	}

	gi := models.GameItem{
		ID:            uuid.Must(uuid.NewV7()).String(),
		GameID:        gameID,
		GameLevelID:   gameLevelID,
		ContentItemID: id,
	}
	if err := facades.Orm().Query().Create(&gi); err != nil {
		return false
	}

	// Mark meta as broken
	if _, err := facades.Orm().Query().Model(&models.ContentMeta{}).
		Where("id", meta.ID).
		Update("is_break_done", true); err != nil {
		return false
	}

	return true
}

// --- GenerateVocabContentItems ---

// GenerateVocabContentItems generates word-level phonetics/POS/translations for vocab content items via SSE.
func GenerateVocabContentItems(userID, gameLevelID string, writer *helpers.SSEWriter) {
	if err := requireVip(userID); err != nil {
		writeVocabSSEError(writer, err)
		return
	}
	game, level, err := verifyVocabLevelOwnership(userID, gameLevelID)
	if err != nil {
		writeVocabSSEError(writer, err)
		return
	}
	if game.Status == consts.GameStatusPublished {
		writeVocabSSEError(writer, ErrGamePublished)
		return
	}
	_ = level

	// Fetch broken metas (ready for item generation)
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
		Where("gm.game_level_id", gameLevelID).
		Where("content_metas.is_break_done", true).
		Order("content_metas.\"order\" ASC").
		Get(&metas); err != nil {
		writeVocabSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
		return
	}

	if len(metas) == 0 {
		_ = writer.Write(SSEProgressEvent{Done: 0, Total: 0, Processed: 0, Failed: 0, Complete: true})
		writer.Close()
		return
	}

	// Filter to metas that have items needing generation (items column is null)
	metaIDs := make([]string, 0, len(metas))
	metaMap := make(map[string]models.ContentMeta)
	for _, m := range metas {
		metaIDs = append(metaIDs, m.ID)
		metaMap[m.ID] = m
	}

	var pendingItems []models.ContentItem
	if err := facades.Orm().Query().
		Where("content_meta_id IN ?", metaIDs).
		Where("is_active", true).
		Where("items IS NULL").
		Get(&pendingItems); err != nil {
		writeVocabSSEError(writer, fmt.Errorf("failed to load pending items: %w", err))
		return
	}

	// Group pending items by meta and compute word counts
	pendingByMeta := make(map[string][]models.ContentItem)
	metaItemWordCounts := make(map[string]int)
	for _, item := range pendingItems {
		if item.ContentMetaID == nil {
			continue
		}
		mid := *item.ContentMetaID
		pendingByMeta[mid] = append(pendingByMeta[mid], item)
		metaItemWordCounts[mid] += helpers.CountWords(item.Content)
	}

	// Only process metas that have pending items
	var activeMetas []models.ContentMeta
	for _, m := range metas {
		if len(pendingByMeta[m.ID]) > 0 {
			activeMetas = append(activeMetas, m)
		}
	}

	if len(activeMetas) == 0 {
		_ = writer.Write(SSEProgressEvent{Done: 0, Total: 0, Processed: 0, Failed: 0, Complete: true})
		writer.Close()
		return
	}

	totalCost := 0
	for _, wc := range metaItemWordCounts {
		totalCost += wc
	}

	if totalCost == 0 {
		writeVocabSSEError(writer, ErrEmptyContent)
		return
	}

	if err := ConsumeBeans(userID, totalCost, consts.BeanSlugAIVocabGenItemsConsume, consts.BeanReasonAIVocabGenItemsConsume); err != nil {
		writeVocabSSEError(writer, err)
		return
	}

	var failedWords int64
	var processed int64
	var failed int64

	sem := make(chan struct{}, vocabGenItemsConcurrencyLimit)
	var wg sync.WaitGroup
	var done int64

	total := len(activeMetas)

	for _, meta := range activeMetas {
		wg.Add(1)
		sem <- struct{}{}

		go func(m models.ContentMeta) {
			defer wg.Done()
			defer func() { <-sem }()

			items := pendingByMeta[m.ID]
			success := processVocabGenItems(m, items)
			d := atomic.AddInt64(&done, 1)

			if success {
				atomic.AddInt64(&processed, 1)
				_ = writer.Write(SSEProgressEvent{Done: int(d), Total: total, Status: "ok"})
			} else {
				atomic.AddInt64(&failed, 1)
				atomic.AddInt64(&failedWords, int64(metaItemWordCounts[m.ID]))
				_ = writer.Write(SSEProgressEvent{Done: int(d), Total: total, Status: "failed"})
			}
		}(meta)
	}

	wg.Wait()

	fw := int(atomic.LoadInt64(&failedWords))
	if fw > 0 {
		_ = RefundBeans(userID, fw, consts.BeanSlugAIVocabGenItemsRefund, consts.BeanReasonAIVocabGenItemsRefund)
	}

	_ = writer.Write(SSEProgressEvent{
		Done:      total,
		Total:     total,
		Processed: int(atomic.LoadInt64(&processed)),
		Failed:    int(atomic.LoadInt64(&failed)),
		Complete:  true,
	})
	writer.Close()
}

func processVocabGenItems(meta models.ContentMeta, existingItems []models.ContentItem) bool {
	unitsInput := make([]map[string]string, len(existingItems))
	for i, item := range existingItems {
		unitsInput[i] = map[string]string{
			"content":     item.Content,
			"contentType": item.ContentType,
		}
	}

	unitsJSON, err := json.Marshal(unitsInput)
	if err != nil {
		return false
	}

	userMsg := "Source text: " + meta.SourceData + "\n\nUnits:\n" + string(unitsJSON)

	result, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
		Messages: []helpers.DeepSeekMessage{
			{Role: "system", Content: vocabGenItemsPrompt},
			{Role: "user", Content: userMsg},
		},
		Temperature: 0.1,
	})
	if err != nil {
		return false
	}

	aiUnits, err := helpers.ParseAIJSONArray(result)
	if err != nil || len(aiUnits) == 0 {
		return false
	}

	aiMap := make(map[string]json.RawMessage)
	for _, raw := range aiUnits {
		var unit struct {
			Content string          `json:"content"`
			Items   json.RawMessage `json:"items"`
		}
		if err := json.Unmarshal(raw, &unit); err != nil {
			continue
		}
		if unit.Content != "" {
			aiMap[unit.Content] = unit.Items
		}
	}

	for _, item := range existingItems {
		itemsJSON, ok := aiMap[item.Content]
		if !ok {
			continue
		}
		itemsStr := string(itemsJSON)
		if _, err := facades.Orm().Query().Model(&models.ContentItem{}).
			Where("id", item.ID).
			Update("items", itemsStr); err != nil {
			return false
		}
	}

	return true
}

// Same prompt as word-sentence — word-level breakdown is identical for vocab.
var vocabGenItemsPrompt = `You are a language learning content processor. You will receive a list of learning units (each with content and contentType). Your job is to break each unit into individual word/punctuation items.

For each unit, produce an "items" array where each element represents one word or punctuation mark:
- position: 1-based index
- item: the word or punctuation character
- phonetic: {"uk": "IPA notation", "us": "IPA notation"} — set to null for punctuation marks
- pos: Chinese part-of-speech label like "名词", "动词", "形容词", "副词", "代词", "介词", "连词", "冠词", "感叹词", "助动词" etc — set to null for punctuation marks
- translation: Chinese translation of the word — set to empty string for punctuation marks
- answer: false for punctuation marks, proper names, place names, and abbreviations; true for all other words

Return a JSON array where each element has:
- content: the unit text (echo back exactly as given)
- items: the items array as described above

OUTPUT FORMAT:
Output ONLY a valid JSON array. No markdown code fences, no explanation, no extra text.

Example input units:
[{"content": "apple", "contentType": "word"}]

Example output:
[
  {
    "content": "apple",
    "items": [
      {"position": 1, "item": "apple", "phonetic": {"uk": "/ˈæp.əl/", "us": "/ˈæp.əl/"}, "pos": "名词", "translation": "苹果", "answer": true}
    ]
  }
]`

// --- Helpers ---

// verifyVocabLevelOwnership checks user owns the game and it is a vocab mode.
func verifyVocabLevelOwnership(userID, gameLevelID string) (*models.Game, *models.GameLevel, error) {
	var level models.GameLevel
	if err := facades.Orm().Query().Where("id", gameLevelID).First(&level); err != nil {
		return nil, nil, fmt.Errorf("failed to find level: %w", err)
	}
	if level.ID == "" {
		return nil, nil, ErrLevelNotFound
	}

	game, err := getCourseGameOwned(userID, level.GameID)
	if err != nil {
		return nil, nil, err
	}

	if !consts.IsVocabMode(game.Mode) {
		return nil, nil, ErrForbidden
	}

	return game, &level, nil
}

func writeVocabSSEError(writer *helpers.SSEWriter, err error) {
	msg := "服务异常"
	switch {
	case errors.Is(err, ErrVipRequired):
		msg = "升级会员解锁此功能"
	case errors.Is(err, ErrGamePublished):
		msg = "已发布的游戏不可编辑，请先撤回"
	case errors.Is(err, ErrInsufficientBeans):
		msg = "能量豆不足"
	case errors.Is(err, ErrEmptyContent):
		msg = "内容为空"
	case errors.Is(err, ErrGameNotFound):
		msg = "游戏不存在"
	case errors.Is(err, ErrLevelNotFound):
		msg = "关卡不存在"
	case errors.Is(err, ErrForbidden):
		msg = "无权操作"
	}
	_ = writer.WriteError(msg)
	writer.Close()
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: No errors

---

## Task 3: Backend Vocab Controller + Routes

**Files:**
- Create: `dx-api/app/http/controllers/api/ai_custom_vocab_controller.go`
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Create vocab controller**

```go
// dx-api/app/http/controllers/api/ai_custom_vocab_controller.go
package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type AiCustomVocabController struct{}

func NewAiCustomVocabController() *AiCustomVocabController {
	return &AiCustomVocabController{}
}

// GenerateVocab generates vocabulary pairs from keywords using AI.
func (c *AiCustomVocabController) GenerateVocab(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req struct {
		Difficulty string   `json:"difficulty"`
		Keywords   []string `json:"keywords"`
		GameMode   string   `json:"gameMode"`
	}
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的请求")
	}

	if req.Difficulty == "" {
		req.Difficulty = "a1-a2"
	}
	if len(req.Keywords) == 0 || len(req.Keywords) > 5 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "请提供1-5个关键词")
	}
	if !consts.IsVocabMode(req.GameMode) {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的游戏模式")
	}

	result, err := services.GenerateVocab(userID, req.Difficulty, req.Keywords, req.GameMode)
	if err != nil {
		return mapVocabServiceError(ctx, err, "词汇 AI 服务")
	}

	if result.Warning != "" {
		return helpers.Success(ctx, map[string]any{"warning": result.Warning})
	}

	return helpers.Success(ctx, map[string]any{
		"generated":  result.Generated,
		"sourceType": result.SourceType,
	})
}

// FormatVocab formats raw text into structured vocabulary content using AI.
func (c *AiCustomVocabController) FormatVocab(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的请求")
	}

	if req.Content == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "请输入内容")
	}

	result, err := services.FormatVocab(userID, req.Content)
	if err != nil {
		return mapVocabServiceError(ctx, err, "词汇格式化服务")
	}

	if result.Warning != "" {
		return helpers.Success(ctx, map[string]any{"warning": result.Warning})
	}

	return helpers.Success(ctx, map[string]any{
		"formatted": result.Formatted,
	})
}

// BreakMetadata breaks vocab content metas into content items via SSE.
func (c *AiCustomVocabController) BreakMetadata(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req struct {
		GameLevelID string `json:"gameLevelId"`
	}
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的请求")
	}

	if req.GameLevelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "gameLevelId is required")
	}

	w := ctx.Response().Writer()
	writer := helpers.NewSSEWriter(w)

	services.BreakVocabMetadata(userID, req.GameLevelID, writer)

	return nil
}

// GenerateContentItems generates word-level phonetics and translations via SSE.
func (c *AiCustomVocabController) GenerateContentItems(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req struct {
		GameLevelID string `json:"gameLevelId"`
	}
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的请求")
	}

	if req.GameLevelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "gameLevelId is required")
	}

	w := ctx.Response().Writer()
	writer := helpers.NewSSEWriter(w)

	services.GenerateVocabContentItems(userID, req.GameLevelID, writer)

	return nil
}

func mapVocabServiceError(ctx contractshttp.Context, err error, serviceLabel string) contractshttp.Response {
	switch {
	case errors.Is(err, services.ErrVipRequired):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
	case errors.Is(err, services.ErrInsufficientBeans):
		return helpers.Error(ctx, http.StatusPaymentRequired, consts.CodeInsufficientBeans, "能量豆不足")
	case errors.Is(err, services.ErrEmptyContent):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "内容为空")
	case errors.Is(err, helpers.ErrDeepSeekEmpty),
		errors.Is(err, helpers.ErrDeepSeekAuth),
		errors.Is(err, helpers.ErrDeepSeekQuota),
		errors.Is(err, helpers.ErrDeepSeekRateLimit),
		errors.Is(err, helpers.ErrDeepSeekNotConfigured),
		errors.Is(err, helpers.ErrDeepSeekUnavail):
		msg, status := helpers.MapDeepSeekError(err, serviceLabel)
		return helpers.Error(ctx, status, consts.CodeAIServiceError, msg)
	default:
		msg, status := helpers.MapDeepSeekError(err, serviceLabel)
		return helpers.Error(ctx, status, consts.CodeAIServiceError, msg)
	}
}
```

- [ ] **Step 2: Add vocab routes to api.go**

Add after the existing `/ai-custom` route group (after line 217 in `dx-api/routes/api.go`), inside the protected middleware group:

```go
			// AI custom vocab content routes
			aiCustomVocabController := apicontrollers.NewAiCustomVocabController()
			protected.Prefix("/ai-custom-vocab").Group(func(aiv route.Router) {
				aiv.Post("/generate-vocab", aiCustomVocabController.GenerateVocab)
				aiv.Post("/format-vocab", aiCustomVocabController.FormatVocab)
				aiv.Post("/break-metadata", aiCustomVocabController.BreakMetadata)
				aiv.Post("/generate-content-items", aiCustomVocabController.GenerateContentItems)
			})
```

- [ ] **Step 3: Verify compilation**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: No errors

---

## Task 4: Backend SaveMetadataBatch Capacity Update

**Files:**
- Modify: `dx-api/app/services/api/course_content_service.go:52-111`

- [ ] **Step 1: Add mode-aware capacity check**

In `SaveMetadataBatch`, after verifying the level (line 70), add a game mode parameter and modify the capacity check. The function signature needs the game object, which is already fetched.

Replace the capacity check block (lines 74-111) with:

```go
	// Check existing capacity
	var existingMetas []models.ContentMeta
	if err := facades.Orm().Query().
		Join("JOIN game_metas gm ON gm.content_meta_id = content_metas.id AND gm.deleted_at IS NULL").
		Where("gm.game_level_id", gameLevelID).
		Get(&existingMetas); err != nil {
		return 0, fmt.Errorf("failed to count metas: %w", err)
	}

	if consts.IsVocabMode(game.Mode) {
		// Vocab modes: flat limit of MaxMetasPerLevel
		if len(existingMetas)+len(entries) > consts.MaxMetasPerLevel {
			return 0, ErrCapacityExceeded
		}
	} else {
		// Word-sentence mode: existing ratio formula
		existingSentences := 0
		existingVocabs := 0
		for _, m := range existingMetas {
			switch m.SourceType {
			case SourceTypeSentence:
				existingSentences++
			case SourceTypeVocab:
				existingVocabs++
			}
		}

		newSentences := 0
		newVocabs := 0
		for _, e := range entries {
			switch e.SourceType {
			case SourceTypeSentence:
				newSentences++
			case SourceTypeVocab:
				newVocabs++
			}
		}

		totalSentences := existingSentences + newSentences
		totalVocabs := existingVocabs + newVocabs

		if float64(totalSentences)/float64(MaxSentences)+float64(totalVocabs)/float64(MaxVocab) > 1 {
			return 0, ErrCapacityExceeded
		}
	}
```

- [ ] **Step 2: Verify compilation**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...`
Expected: No errors

- [ ] **Step 3: Run go vet**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./...`
Expected: No issues

---

## Task 5: Frontend Helpers

**Files:**
- Create: `dx-web/src/features/web/ai-custom-vocab/helpers/count-words.ts`
- Create: `dx-web/src/features/web/ai-custom-vocab/helpers/stream-progress.ts`
- Create: `dx-web/src/features/web/ai-custom-vocab/helpers/format-metadata.ts`
- Create: `dx-web/src/features/web/ai-custom-vocab/helpers/generate-api.ts`
- Create: `dx-web/src/features/web/ai-custom-vocab/helpers/format-api.ts`
- Create: `dx-web/src/features/web/ai-custom-vocab/helpers/generate-items-api.ts`

- [ ] **Step 1: Copy unchanged helpers**

Copy these files directly from `ai-custom`:
- `dx-web/src/features/web/ai-custom/helpers/count-words.ts` → `dx-web/src/features/web/ai-custom-vocab/helpers/count-words.ts`
- `dx-web/src/features/web/ai-custom/helpers/stream-progress.ts` → `dx-web/src/features/web/ai-custom-vocab/helpers/stream-progress.ts`

- [ ] **Step 2: Create vocab format-metadata.ts**

```typescript
// dx-web/src/features/web/ai-custom-vocab/helpers/format-metadata.ts
import type { GameMode } from "@/consts/game-mode";

const CJK_REGEX = /[\u4e00-\u9fff]/;

export const MAX_METAS_PER_LEVEL = 20;
export const MAX_CONTENT_LENGTH = 600;

/** Max pairs per submission based on game mode */
export function maxPairsForMode(mode: GameMode): number {
  switch (mode) {
    case "vocab-match": return 5;
    case "vocab-elimination": return 8;
    case "vocab-battle": return 20;
    default: return 5;
  }
}

export type VocabPair = {
  sourceData: string;
  translation: string;
};

function isChinese(line: string): boolean {
  return CJK_REGEX.test(line);
}

export type ParseVocabResult =
  | { ok: true; pairs: VocabPair[] }
  | { ok: false; error: string };

/**
 * Parse alternating English/Chinese lines into vocab pairs.
 * Every English line MUST have a Chinese translation on the next line.
 */
export function parseVocabText(raw: string, maxPairs: number): ParseVocabResult {
  const lines = raw
    .split("\n")
    .map((l) => l.trim())
    .filter((l) => l.length > 0);

  if (lines.length === 0) {
    return { ok: false, error: "未解析到有效内容，请检查输入" };
  }

  const pairs: VocabPair[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    if (isChinese(line)) {
      // Orphan Chinese line without preceding English — error
      return { ok: false, error: `第 ${i + 1} 行是中文但缺少对应的英文词汇` };
    }

    // English line — must have Chinese next
    if (i + 1 >= lines.length || !isChinese(lines[i + 1])) {
      return { ok: false, error: `词汇「${line}」缺少中文释义，请确保每个英文词汇下方都有对应的中文释义` };
    }

    const oversized = line.length > MAX_CONTENT_LENGTH || lines[i + 1].length > MAX_CONTENT_LENGTH;
    if (oversized) {
      return { ok: false, error: `单条内容或翻译超过 ${MAX_CONTENT_LENGTH} 字符限制` };
    }

    pairs.push({
      sourceData: line,
      translation: lines[i + 1],
    });
    i += 2;
  }

  if (pairs.length > maxPairs) {
    return { ok: false, error: `词汇数量（${pairs.length}）超过当前模式上限 ${maxPairs} 对，请精简后重试` };
  }

  return { ok: true, pairs };
}
```

- [ ] **Step 3: Create vocab generate-api.ts**

```typescript
// dx-web/src/features/web/ai-custom-vocab/helpers/generate-api.ts
import type { GameMode } from "@/consts/game-mode";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

type GenerateResult =
  | { ok: true; generated: string; sourceType: "vocab" }
  | { ok: false; message: string; code?: string; required?: number; available?: number };

/** Call Go API to generate vocabulary pairs from keywords using AI */
export async function generateVocab(
  difficulty: string,
  keywords: string[],
  gameMode: GameMode
): Promise<GenerateResult> {
  try {
    const res = await fetch(`${API_URL}/api/ai-custom-vocab/generate-vocab`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ difficulty, keywords, gameMode }),
    });

    const json = await res.json();

    if (!res.ok || json.code !== 0) {
      return {
        ok: false,
        message: json.message ?? "生成失败",
        code: json.code === 40007 ? "INSUFFICIENT_BEANS" : undefined,
      };
    }

    const data = json.data;

    if (data.warning) {
      return { ok: false, message: data.warning };
    }

    return { ok: true, generated: data.generated, sourceType: "vocab" };
  } catch {
    return { ok: false, message: "网络错误，请稍后重试" };
  }
}
```

- [ ] **Step 4: Create vocab format-api.ts**

```typescript
// dx-web/src/features/web/ai-custom-vocab/helpers/format-api.ts
const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001";

type FormatResult =
  | { ok: true; formatted: string }
  | { ok: false; message: string; code?: string; required?: number; available?: number };

/** Call Go API to format raw text into structured vocabulary content */
export async function formatVocab(content: string): Promise<FormatResult> {
  try {
    const res = await fetch(`${API_URL}/api/ai-custom-vocab/format-vocab`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ content }),
    });

    const json = await res.json();

    if (!res.ok || json.code !== 0) {
      return {
        ok: false,
        message: json.message ?? "格式化失败",
        code: json.code === 40007 ? "INSUFFICIENT_BEANS" : undefined,
      };
    }

    const data = json.data;

    if (data.warning) {
      return { ok: false, message: data.warning };
    }

    return { ok: true, formatted: data.formatted };
  } catch {
    return { ok: false, message: "网络错误，请稍后重试" };
  }
}
```

- [ ] **Step 5: Create vocab generate-items-api.ts**

```typescript
// dx-web/src/features/web/ai-custom-vocab/helpers/generate-items-api.ts
import { fetchWithProgress, type ProgressEvent } from "@/features/web/ai-custom-vocab/helpers/stream-progress";

type BatchResult =
  | { ok: true; processed: number; failed: number }
  | { ok: false; message: string; code?: string; required?: number; available?: number };

type OnProgress = (event: ProgressEvent) => void;

/** Call Go API to break vocab content metas into content items via SSE */
export async function breakVocabMetadata(
  gameLevelId: string,
  signal?: AbortSignal,
  onProgress?: OnProgress
): Promise<BatchResult> {
  return fetchWithProgress(
    "/api/ai-custom-vocab/break-metadata",
    { gameLevelId },
    signal,
    onProgress ?? (() => {})
  );
}

/** Call Go API to generate word-level details for vocab items via SSE */
export async function generateVocabContentItems(
  gameLevelId: string,
  signal?: AbortSignal,
  onProgress?: OnProgress
): Promise<BatchResult> {
  return fetchWithProgress(
    "/api/ai-custom-vocab/generate-content-items",
    { gameLevelId },
    signal,
    onProgress ?? (() => {})
  );
}
```

- [ ] **Step 6: Verify no TypeScript errors**

Run: `cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npx tsc --noEmit --pretty 2>&1 | head -30`
Expected: No errors in the new files (existing errors may be present)

---

## Task 6: Frontend Schemas + Actions

**Files:**
- Create: `dx-web/src/features/web/ai-custom-vocab/schemas/course-game.schema.ts`
- Create: `dx-web/src/features/web/ai-custom-vocab/actions/course-game.action.ts`

- [ ] **Step 1: Copy and modify schemas**

Copy `dx-web/src/features/web/ai-custom/schemas/course-game.schema.ts` to `dx-web/src/features/web/ai-custom-vocab/schemas/course-game.schema.ts`.

No modifications needed — the Zod schemas validate game CRUD which is shared.

- [ ] **Step 2: Copy and modify actions**

Copy `dx-web/src/features/web/ai-custom/actions/course-game.action.ts` to `dx-web/src/features/web/ai-custom-vocab/actions/course-game.action.ts`.

Modify all imports to point to the vocab feature directory:
- `@/features/web/ai-custom/schemas/course-game.schema` → `@/features/web/ai-custom-vocab/schemas/course-game.schema`

---

## Task 7: Frontend Hooks

**Files:**
- Create: all 5 hooks under `dx-web/src/features/web/ai-custom-vocab/hooks/`

- [ ] **Step 1: Copy all hooks from ai-custom**

Copy all files from `dx-web/src/features/web/ai-custom/hooks/` to `dx-web/src/features/web/ai-custom-vocab/hooks/`.

Modify all internal imports to point to vocab feature:
- `@/features/web/ai-custom/` → `@/features/web/ai-custom-vocab/`

In `use-infinite-games.ts`, change the SWR key filter to only return vocab mode games. The filter is applied server-side via query params — add `&modes=vocab-battle,vocab-match,vocab-elimination` to the SWR key if the API supports it. If not, filter client-side in the hook.

In `use-game-actions.ts`, change navigation after delete from `/hall/ai-custom` to `/hall/ai-custom-vocab`.

---

## Task 8: Frontend Vocab Manual Tab + AI Tab

**Files:**
- Create: `dx-web/src/features/web/ai-custom-vocab/components/vocab-manual-tab.tsx`
- Create: `dx-web/src/features/web/ai-custom-vocab/components/vocab-ai-tab.tsx`

- [ ] **Step 1: Create vocab-manual-tab.tsx**

Similar to `manual-add-tab.tsx` but:
- Only vocab format example HoverCard (remove sentence example)
- Textarea placeholder: "输入词汇，英文一行、中文下一行..."
- No changes to Copy/Clear buttons

```tsx
// dx-web/src/features/web/ai-custom-vocab/components/vocab-manual-tab.tsx
import { BookOpen, CircleAlert, Copy, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Textarea } from "@/components/ui/textarea";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@/components/ui/hover-card";

type VocabManualTabProps = {
  value: string;
  onChange: (value: string) => void;
  error?: string;
  maxPairs: number;
};

export function VocabManualTab({ value, onChange, error, maxPairs }: VocabManualTabProps) {
  return (
    <div className="flex flex-col gap-3 px-6 py-3">
      <div className="flex flex-col gap-2">
        <div className="flex items-start gap-2 rounded-xl bg-amber-50 px-3.5 py-3 text-xs leading-relaxed">
          <span className="shrink-0 font-semibold text-amber-600">输入说明：</span>
          <span className="text-amber-500">
            请输入英文-中文词汇对，每个英文词汇占一行，紧接着下一行填写对应的中文释义。每次最多添加 {maxPairs} 对词汇。
          </span>
        </div>
        <div className="flex items-start gap-2 rounded-xl bg-muted px-3.5 py-3 text-xs leading-relaxed">
          <span className="shrink-0 font-semibold text-muted-foreground">格式说明：</span>
          <span className="text-muted-foreground">
            严格按照英文一行、中文释义下一行的格式输入，每对词汇之间不需要空行。所有词汇必须成对出现。
          </span>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <HoverCard openDelay={200}>
          <HoverCardTrigger asChild>
            <button
              type="button"
              className="flex items-center gap-1.5 rounded-lg bg-muted px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent"
            >
              <BookOpen className="h-3.5 w-3.5" />
              词汇输入格式示例
            </button>
          </HoverCardTrigger>
          <HoverCardContent align="start" className="w-80">
            <div className="flex flex-col gap-2">
              <p className="text-xs font-semibold text-foreground">词汇输入格式示例</p>
              <div className="rounded-lg bg-muted p-3 text-xs leading-[1.8] text-muted-foreground">
                <p>apple</p>
                <p className="text-muted-foreground">苹果</p>
                <p>banana</p>
                <p className="text-muted-foreground">香蕉</p>
                <p>polar bear</p>
                <p className="text-muted-foreground">北极熊</p>
              </div>
            </div>
          </HoverCardContent>
        </HoverCard>
        <div className="ml-auto flex items-center gap-2">
          <button
            type="button"
            disabled={!value}
            onClick={async () => {
              await navigator.clipboard.writeText(value);
              toast.success("已复制到剪贴板");
            }}
            className="flex items-center gap-1.5 rounded-lg bg-muted px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent disabled:opacity-50"
          >
            <Copy className="h-3.5 w-3.5" />
            复制
          </button>
          <button
            type="button"
            disabled={!value}
            onClick={() => onChange("")}
            className="flex items-center gap-1.5 rounded-lg bg-muted px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent disabled:opacity-50"
          >
            <Trash2 className="h-3.5 w-3.5" />
            清空
          </button>
        </div>
      </div>
      <Textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="输入词汇，英文一行、中文下一行..."
        className={`h-[280px] resize-none overflow-y-auto rounded-xl bg-muted px-4 py-3.5 text-[13px] leading-[1.8] text-foreground shadow-none focus-visible:ring-1 ${error ? "border-red-400 focus-visible:ring-red-400" : "border-border focus-visible:ring-teal-500"}`}
      />
      {error && (
        <p className="flex items-center gap-1.5 text-xs text-red-500">
          <CircleAlert className="h-3.5 w-3.5 shrink-0" />
          {error}
        </p>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Create vocab-ai-tab.tsx**

Similar to `ai-generate-tab.tsx` but preview stats show pair count:

```tsx
// dx-web/src/features/web/ai-custom-vocab/components/vocab-ai-tab.tsx
"use client";

import { Gauge, TextCursorInput, Eye, Hash, CircleAlert } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DIFFICULTY_OPTIONS } from "@/consts/difficulty";

type VocabAiTabProps = {
  difficulty: string;
  onDifficultyChange: (value: string) => void;
  keywords: string;
  onKeywordsChange: (value: string) => void;
  preview: string;
  error: string;
};

const MAX_KEYWORDS = 5;
const MAX_WORD_LENGTH = 30;

export function getKeywordsWarning(keywords: string): string {
  const trimmed = keywords.trim();
  if (!trimmed) return "";
  const words = trimmed.split(/\s+/).filter(Boolean);
  if (words.length === 1 && /[,，、;；/|]/.test(trimmed)) {
    return "请用空格分隔关键词";
  }
  if (words.length > MAX_KEYWORDS) {
    return `最多输入 ${MAX_KEYWORDS} 个关键词，当前 ${words.length} 个`;
  }
  const long = words.find((w) => w.length > MAX_WORD_LENGTH);
  if (long) {
    return `单个关键词不能超过 ${MAX_WORD_LENGTH} 个字符`;
  }
  return "";
}

function countVocabPairs(preview: string): number {
  if (!preview) return 0;
  const lines = preview.split("\n").filter((l) => l.trim().length > 0);
  return Math.floor(lines.length / 2);
}

export function VocabAiTab({
  difficulty,
  onDifficultyChange,
  keywords,
  onKeywordsChange,
  preview,
  error,
}: VocabAiTabProps) {
  const keywordsWarning = getKeywordsWarning(keywords);

  return (
    <div className="flex flex-col gap-5 px-6 py-3">
      {/* Difficulty */}
      <div className="flex flex-col gap-2">
        <div className="flex items-center gap-1.5">
          <Gauge className="h-3.5 w-3.5 text-teal-600" />
          <span className="text-[13px] font-semibold text-foreground">难度</span>
        </div>
        <Select value={difficulty} onValueChange={onDifficultyChange}>
          <SelectTrigger className="h-11 rounded-xl border-border bg-muted px-4 text-[13px] shadow-none focus:ring-1 focus:ring-teal-500">
            <SelectValue placeholder="选择难度" />
          </SelectTrigger>
          <SelectContent>
            {DIFFICULTY_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Keywords */}
      <div className="flex flex-col gap-2">
        <div className="flex items-center gap-2">
          <TextCursorInput className="h-3.5 w-3.5 text-teal-600" />
          <span className="text-[13px] font-semibold text-foreground">
            关键词
          </span>
          <span className="text-xs text-muted-foreground">
            最多 5 个单词，用空格分开
          </span>
        </div>
        <input
          value={keywords}
          onChange={(e) => onKeywordsChange(e.target.value)}
          placeholder="示例: fruit animal color"
          className={`h-11 rounded-xl border bg-muted px-4 text-[13px] text-foreground outline-none focus:ring-1 ${keywordsWarning ? "border-red-400 focus:ring-red-400" : "border-border focus:ring-teal-500"}`}
        />
        {keywordsWarning && (
          <p className="flex items-center gap-1.5 text-xs text-red-500">
            <CircleAlert className="h-3.5 w-3.5 shrink-0" />
            {keywordsWarning}
          </p>
        )}
      </div>

      {/* Preview */}
      <div className="flex flex-col gap-2">
        <div className="flex items-center gap-1.5">
          <Eye className="h-3.5 w-3.5 text-teal-600" />
          <span className="text-[13px] font-semibold text-foreground">生成预览</span>
        </div>
        <div className="min-h-[180px] max-h-[280px] overflow-y-auto rounded-xl border border-border bg-muted p-4">
          {preview ? (
            <p className="whitespace-pre-line text-[13px] leading-[1.8] text-foreground">
              {preview}
            </p>
          ) : (
            <p className="text-xs text-muted-foreground">
              生成后将在此处显示预览内容...
            </p>
          )}
        </div>
        {error && (
          <p className="flex items-center gap-1.5 text-xs text-red-500">
            <CircleAlert className="h-3.5 w-3.5 shrink-0" />
            {error}
          </p>
        )}
        <div className="flex items-center gap-1.5">
          <Hash className="h-3 w-3 text-muted-foreground" />
          <span className="text-xs text-muted-foreground">
            词汇对数：{countVocabPairs(preview)}
          </span>
        </div>
      </div>
    </div>
  );
}
```

---

## Task 9: Frontend Add Vocab Dialog

**Files:**
- Create: `dx-web/src/features/web/ai-custom-vocab/components/add-vocab-dialog.tsx`

- [ ] **Step 1: Create add-vocab-dialog.tsx**

This is the core dialog. Key differences from `AddMetadataDialog`:
- Only "词汇检查并格式化" button (no sentence format)
- Uses `parseVocabText()` instead of `parseMetadataText()`
- All entries get `sourceType: "vocab"`
- Validates paired input and mode-specific max count
- Receives `gameMode` prop
- Uses vocab API endpoints

Create the file by copying `dx-web/src/features/web/ai-custom/components/add-metadata-dialog.tsx` and applying these changes:
- Replace imports to use vocab feature paths
- Replace `ManualAddTab` with `VocabManualTab`
- Replace `AiGenerateTab` with `VocabAiTab`
- Replace `parseMetadataText` with `parseVocabText` from vocab format-metadata
- Replace `formatMetadata` with `formatVocab` from vocab format-api
- Replace `generateStory` with `generateVocab` from vocab generate-api
- Replace `MAX_ENTRIES`, `MAX_SENTENCES`, `MAX_VOCAB` with `MAX_METAS_PER_LEVEL`, `maxPairsForMode`
- Remove `formattingType` state for sentence (only vocab format)
- Remove `sourceTypes` state (always vocab)
- In `handleSave()`: use `parseVocabText(manualText, maxPairs)` — if `!result.ok`, show error. Map pairs to entries with `sourceType: "vocab"` and `translation` from pair
- In `handleFormat()`: call `formatVocab(text)` — no formatType param
- In `handleGenerate()`: call `generateVocab(difficulty, words, gameMode)`
- Footer: single "词汇检查并格式化" button + "保存" button (no sentence format button)
- Pass `maxPairs` to `VocabManualTab`
- Props: add `gameMode: GameMode`
- Capacity check: `existingMetaCount + pairs.length > MAX_METAS_PER_LEVEL`

---

## Task 10: Frontend Core Components

**Files:**
- Create: all remaining components under `dx-web/src/features/web/ai-custom-vocab/components/`

- [ ] **Step 1: Copy unchanged components**

Copy these files directly from `ai-custom/components/` to `ai-custom-vocab/components/`:
- `processing-overlay.tsx`
- `sortable-content-item.tsx`
- `sortable-meta-item.tsx`
- `add-level-dialog.tsx`
- `game-info-card.tsx`

No modifications needed for these files.

- [ ] **Step 2: Copy and modify route-dependent components**

Copy these files and change all route references from `/hall/ai-custom` to `/hall/ai-custom-vocab` and imports from `@/features/web/ai-custom/` to `@/features/web/ai-custom-vocab/`:

- `game-card-item.tsx` — change link href
- `game-hero-card.tsx` — change breadcrumb/back link
- `game-levels-card.tsx` — change level link hrefs
- `course-detail-content.tsx` — change imports, breadcrumb hrefs
- `edit-game-dialog.tsx` — change imports, filter to vocab modes only

- [ ] **Step 3: Create vocab create-course-form.tsx**

Copy from `ai-custom/components/create-course-form.tsx` and:
- Change imports to vocab feature paths
- Filter `GAME_MODE_OPTIONS` to only vocab modes:
  ```tsx
  const VOCAB_MODE_OPTIONS = GAME_MODE_OPTIONS.filter(
    (opt) => opt.value !== "word-sentence"
  );
  ```
- Use `VOCAB_MODE_OPTIONS` in the mode selector
- Change success redirect from `/hall/ai-custom/${id}` to `/hall/ai-custom-vocab/${id}`

- [ ] **Step 4: Create vocab ai-custom-vocab-grid.tsx**

Copy from `ai-custom/components/ai-custom-grid.tsx` and:
- Change imports to vocab feature paths
- Change page title/subtitle to reflect vocab games
- Filter games to only show vocab modes (client-side filter or API param)
- Use `CreateCourseForm` from vocab feature
- Change card links to `/hall/ai-custom-vocab/`

---

## Task 11: Frontend Level Units Panel (Vocab Version)

**Files:**
- Create: `dx-web/src/features/web/ai-custom-vocab/components/level-units-panel.tsx`

- [ ] **Step 1: Create vocab level-units-panel.tsx**

Copy from `ai-custom/components/level-units-panel.tsx` and apply these changes:

1. Change imports to use vocab feature paths:
   - `@/features/web/ai-custom-vocab/components/add-vocab-dialog` (instead of add-metadata-dialog)
   - `@/features/web/ai-custom-vocab/helpers/generate-items-api` (breakVocabMetadata, generateVocabContentItems)
   - `@/features/web/ai-custom-vocab/helpers/format-metadata` (MAX_METAS_PER_LEVEL)
   - Remove `SOURCE_TYPES` import (not needed for stats)

2. Props: add `gameMode: GameMode`, remove `sentenceItemCount` and `vocabItemCount`

3. Stats bar: simplify to 2 stats:
   ```tsx
   <span>共计：<span>{metas.length}</span></span>
   <span>练习单元总数：<span>{totalItemCount}</span></span>
   ```

4. Capacity check: replace ratio formula with:
   ```tsx
   const isAtCapacity = metas.length >= MAX_METAS_PER_LEVEL;
   ```

5. Add button tooltip: `已达上限（${metas.length}/${MAX_METAS_PER_LEVEL}）`

6. Replace `<AddMetadataDialog>` with:
   ```tsx
   <AddVocabDialog
     gameId={gameId}
     levelId={levelId}
     gameMode={gameMode}
     open={metadataDialogOpen}
     onOpenChange={setMetadataDialogOpen}
     existingMetaCount={metas.length}
   />
   ```

7. Replace `breakMetadata` / `generateContentItems` calls with `breakVocabMetadata` / `generateVocabContentItems`

---

## Task 12: Frontend App Route Pages

**Files:**
- Create: `dx-web/src/app/(web)/hall/(main)/ai-custom-vocab/page.tsx`
- Create: `dx-web/src/app/(web)/hall/(main)/ai-custom-vocab/[id]/page.tsx`
- Create: `dx-web/src/app/(web)/hall/(main)/ai-custom-vocab/[id]/[levelId]/page.tsx`

- [ ] **Step 1: Create landing page**

```tsx
// dx-web/src/app/(web)/hall/(main)/ai-custom-vocab/page.tsx
import { PageTopBar } from "@/features/web/hall/components/page-top-bar"
import { AiCustomVocabGrid } from "@/features/web/ai-custom-vocab/components/ai-custom-vocab-grid"

export default function AiCustomVocabPage() {
  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <PageTopBar
        title="AI 词汇工坊"
        subtitle="AI 驱动的个性化词汇练习游戏"
      />
      <AiCustomVocabGrid />
    </div>
  )
}
```

- [ ] **Step 2: Create course detail page**

```tsx
// dx-web/src/app/(web)/hall/(main)/ai-custom-vocab/[id]/page.tsx
import { CourseDetailContent } from "@/features/web/ai-custom-vocab/components/course-detail-content";

export default async function CourseGameDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <CourseDetailContent id={id} />
    </div>
  );
}
```

- [ ] **Step 3: Create level editor page**

```tsx
// dx-web/src/app/(web)/hall/(main)/ai-custom-vocab/[id]/[levelId]/page.tsx
"use client"

import { use } from "react"
import useSWR from "swr"
import { BreadcrumbTopBar } from "@/features/web/hall/components/breadcrumb-top-bar"
import { LevelUnitsPanel } from "@/features/web/ai-custom-vocab/components/level-units-panel"
import { GAME_STATUSES } from "@/consts/game-status"
import { PageSpinner } from "@/components/in/page-spinner"
import type { GameMode } from "@/consts/game-mode"

export default function CourseGameLevelPage({
  params,
}: {
  params: Promise<{ id: string; levelId: string }>
}) {
  const { id, levelId } = use(params)

  const { data: game, isLoading: gameLoading } = useSWR(`/api/course-games/${id}`)
  type ContentGroupItem = { items: unknown[] | null; contentType: string };
  type ContentGroup = {
    meta: {
      id: string;
      sourceData: string;
      translation: string | null;
      sourceFrom: string;
      sourceType: string;
      isBreakDone: boolean;
      order: number;
    };
    items?: ContentGroupItem[];
  };

  const { data: contentGroups, isLoading: contentLoading } = useSWR<ContentGroup[]>(
    `/api/course-games/${id}/levels/${levelId}/content-items`
  )

  if (gameLoading || contentLoading) return <PageSpinner size="lg" />

  const metas = (contentGroups ?? []).map((group) => ({
    id: group.meta.id,
    sourceData: group.meta.sourceData,
    translation: group.meta.translation ?? null,
    sourceFrom: group.meta.sourceFrom,
    sourceType: group.meta.sourceType,
    isBreakDone: group.meta.isBreakDone,
    isItemDone: group.meta.isBreakDone && (group.items?.length ?? 0) > 0
      && group.items!.every((item) => item.items !== null),
    order: group.meta.order,
    itemCount: group.items?.length ?? 0,
  }))

  const level = game?.levels?.find((l: { id: string }) => l.id === levelId)
  const isPublished = game?.status === GAME_STATUSES.PUBLISHED

  return (
    <div className="flex min-h-full flex-col gap-5 px-4 pt-5 pb-12 lg:gap-6 lg:px-8 lg:pt-7 lg:pb-16">
      <BreadcrumbTopBar
        backHref={`/hall/ai-custom-vocab/${id}`}
        items={[
          { label: "我创建的词汇游戏", href: "/hall/ai-custom-vocab", maxChars: 10 },
          { label: game?.name ?? id, href: `/hall/ai-custom-vocab/${id}`, maxChars: 5 },
          { label: level?.name ?? levelId, maxChars: 5 },
        ]}
      />

      <LevelUnitsPanel
        gameId={id}
        levelId={levelId}
        gameMode={(game?.mode ?? "vocab-match") as GameMode}
        initialMetas={metas}
        readOnly={isPublished}
      />
    </div>
  )
}
```

---

## Task 13: Update Existing AI Custom Create Form

**Files:**
- Modify: `dx-web/src/features/web/ai-custom/components/create-course-form.tsx`

- [ ] **Step 1: Filter mode selector to word-sentence only**

In the create form, find where `GAME_MODE_OPTIONS` is used in the mode select dropdown and filter:

```tsx
const SENTENCE_MODE_OPTIONS = GAME_MODE_OPTIONS.filter(
  (opt) => opt.value === "word-sentence"
);
```

Use `SENTENCE_MODE_OPTIONS` in the Select component instead of `GAME_MODE_OPTIONS`.

Apply the same change to `dx-web/src/features/web/ai-custom/components/edit-game-dialog.tsx` if it also has a mode selector.

---

## Task 14: Lint + Build Verification

- [ ] **Step 1: Run Go build and vet**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./... && go vet ./...
```
Expected: No errors

- [ ] **Step 2: Run frontend lint**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run lint
```
Expected: No lint errors in new files

- [ ] **Step 3: Run frontend type check**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npx tsc --noEmit
```
Expected: No type errors in new files

- [ ] **Step 4: Verify existing ai-custom still works**

Verify that `dx-web/src/features/web/ai-custom/` imports are unchanged (except create-course-form mode filter). No other existing files should have been modified in the ai-custom feature.

- [ ] **Step 5: Manual smoke test checklist**

1. Navigate to `/hall/ai-custom` — verify only word-sentence mode available in create form
2. Navigate to `/hall/ai-custom-vocab` — verify all 3 vocab modes available
3. Create a vocab-match game → create level → open level editor
4. Click "添加" → verify only "词汇检查并格式化" button visible
5. Try manual input with unpaired text → verify error shown
6. Try manual input with paired text → verify save works
7. Try AI generation → verify vocab pairs generated
8. Click "分解" → verify items created
9. Click "生成" → verify phonetics/POS added
10. Verify existing word-sentence games still work end-to-end

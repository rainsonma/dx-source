package api

import (
	"context"
	"encoding/json"
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

// GenerateVocabResult holds the response from vocab generation.
type GenerateVocabResult struct {
	Generated string `json:"generated,omitempty"`
	Warning   string `json:"warning,omitempty"`
}

// FormatVocabResult holds the response from vocab formatting.
type FormatVocabResult struct {
	Formatted string `json:"formatted,omitempty"`
	Warning   string `json:"warning,omitempty"`
}

// --- GenerateVocab ---

// GenerateVocab generates English-Chinese vocab pairs from keywords using DeepSeek AI.
// Consumes 5 beans. Count based on game mode (5/8/20).
func GenerateVocab(userID string, difficulty string, keywords []string, gameMode string) (*GenerateVocabResult, error) {
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

	if warning, ok := strings.CutPrefix(result, "WARNING:"); ok {
		return &GenerateVocabResult{Warning: strings.TrimSpace(warning)}, nil
	}

	return &GenerateVocabResult{Generated: result}, nil
}

func buildVocabGeneratePrompt(levelDesc string, count int) string {
	return fmt.Sprintf(`You are a vocabulary generator for an English learning app. Your job is to generate English-Chinese vocabulary pairs.

STEP 1 — CONTENT MODERATION (do this FIRST):
Check if the provided keywords contain any insulting, violent, sexually explicit, or otherwise inappropriate/sensitive material.
If they do, respond ONLY with: WARNING:包含不适当内容，请修改后重试
Do NOT generate any vocabulary. Stop here.

STEP 2 — GENERATE VOCABULARY:
Generate exactly %d English-Chinese vocabulary pairs that:
- Are related to ALL the provided keywords
- Use CEFR level %s appropriate vocabulary
- Include a mix of single words and short phrases
- Are suitable for English language learners

OUTPUT FORMAT:
- Each pair on two lines: English word/phrase, then Chinese translation
- No numbering, no bullets, no markdown, no explanations
- No empty lines between pairs

Example output:
apple
苹果
banana
香蕉
red apple
红苹果`, count, levelDesc)
}

// --- FormatVocab ---

// FormatVocab formats raw vocab text via DeepSeek.
// Only vocab — rejects sentences. Cost = word count.
func FormatVocab(userID string, content string) (*FormatVocabResult, error) {
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

	if warning, ok := strings.CutPrefix(result, "WARNING:"); ok {
		return &FormatVocabResult{Warning: strings.TrimSpace(warning)}, nil
	}

	formatted := cleanVocabFormatted(result)
	if formatted == "" {
		_ = RefundBeans(userID, wordCount, consts.BeanSlugAIVocabFormatRefund, consts.BeanReasonAIVocabFormatRefund)
		return nil, helpers.ErrDeepSeekEmpty
	}

	return &FormatVocabResult{Formatted: formatted}, nil
}

var vocabFormatPrompt = `You are a content formatter for an English learning app. Your job is to clean up and reformat messy user input into a strict line-by-line format for vocabulary.

STEP 1 — CONTENT MODERATION (do this FIRST):
Check if the content contains any insulting, violent, sexually explicit, or otherwise inappropriate/sensitive material.
If it does, respond ONLY with: WARNING:内容包含不适当内容，请修改后重试
Do NOT format the content. Stop here.

STEP 2 — TYPE MISMATCH CHECK:
If the content consists mostly of full sentences with punctuation, respond ONLY with: WARNING:内容看起来是语句而非词汇，请使用「连词成句」模式

STEP 3 — FORMAT:
- If the content contains Chinese text: output alternating lines of English word/phrase followed by its Chinese translation.
- If the content contains NO Chinese text: output English words/phrases only, one per line.

RULES:
- Output ONLY the formatted text. No explanations, headers, numbering, or markdown.
- Remove duplicates.
- Fix obvious spelling errors in English.
- Preserve the original meaning. Do not add or remove content.
- Each line must contain exactly ONE word or phrase (or one translation).
- Remove any empty lines.`

// cleanVocabFormatted removes empty lines and trims whitespace.
func cleanVocabFormatted(result string) string {
	lines := strings.Split(result, "\n")
	var clean []string
	for _, line := range lines {
		line = strings.TrimRight(line, " \t\r")
		if line != "" {
			clean = append(clean, line)
		}
	}
	return strings.Join(clean, "\n")
}

// --- BreakVocabMetadata ---

// BreakVocabMetadata processes vocab metas for a game level via SSE.
// Each vocab meta becomes exactly 1 content item. NO AI call needed.
func BreakVocabMetadata(userID, gameLevelID string, writer *helpers.NDJSONWriter) {
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

	// Fetch unbroken metas
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("is_break_done", false).
		Order("\"order\" ASC").
		Get(&metas); err != nil {
		writeVocabSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
		return
	}

	if len(metas) == 0 {
		_ = writer.Write(SSEProgressEvent{Done: 0, Total: 0, Processed: 0, Failed: 0, Complete: true})
		writer.Close()
		return
	}

	// Bean cost = number of metas (1 per meta, since no AI call)
	totalCost := len(metas)
	if err := ConsumeBeans(userID, totalCost, consts.BeanSlugAIVocabBreakConsume, consts.BeanReasonAIVocabBreakConsume); err != nil {
		writeVocabSSEError(writer, err)
		return
	}

	var processed int64
	var failed int64

	sem := make(chan struct{}, breakConcurrencyLimit)
	var wg sync.WaitGroup
	var done int64

	total := len(metas)

	for _, meta := range metas {
		wg.Add(1)
		sem <- struct{}{}

		go func(m models.ContentMeta) {
			defer wg.Done()
			defer func() { <-sem }()

			success := processVocabBreakMeta(m, gameLevelID)
			d := atomic.AddInt64(&done, 1)

			if success {
				atomic.AddInt64(&processed, 1)
				_ = writer.Write(SSEProgressEvent{Done: int(d), Total: total, Status: "ok"})
			} else {
				atomic.AddInt64(&failed, 1)
				_ = writer.Write(SSEProgressEvent{Done: int(d), Total: total, Status: "failed"})
			}
		}(meta)
	}

	wg.Wait()

	// Refund failed metas
	fw := int(atomic.LoadInt64(&failed))
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

// processVocabBreakMeta creates exactly 1 content item per meta.
// Single word -> contentType "word", multi-word -> "phrase". No AI call.
func processVocabBreakMeta(meta models.ContentMeta, gameLevelID string) bool {
	contentType := "word"
	if strings.Contains(strings.TrimSpace(meta.SourceData), " ") {
		contentType = "phrase"
	}

	id := uuid.Must(uuid.NewV7()).String()
	metaID := meta.ID

	item := models.ContentItem{
		ID:            id,
		GameLevelID:   gameLevelID,
		ContentMetaID: &metaID,
		Content:       meta.SourceData,
		ContentType:   contentType,
		Translation:   meta.Translation,
		Order:         meta.Order + 10,
		IsActive:      true,
	}
	if err := facades.Orm().Query().Create(&item); err != nil {
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
func GenerateVocabContentItems(ctx context.Context, userID, gameLevelID string, writer *helpers.NDJSONWriter) {
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

	// Fetch broken metas
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("is_break_done", true).
		Order("\"order\" ASC").
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

	sem := make(chan struct{}, genItemsConcurrencyLimit)
	var wg sync.WaitGroup
	var done int64

	total := len(activeMetas)

	for _, meta := range activeMetas {
		// Stop dispatching if client disconnected
		if ctx.Err() != nil {
			break
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(m models.ContentMeta) {
			defer wg.Done()
			defer func() { <-sem }()

			// Skip if client already gone
			if ctx.Err() != nil {
				atomic.AddInt64(&done, 1)
				return
			}

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

// processVocabGenItems generates phonetics/POS items JSON for vocab content items.
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
			{Role: "system", Content: genItemsPrompt},
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

// --- Helpers ---

// verifyVocabLevelOwnership checks ownership and that the game is a vocab mode.
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

// writeVocabSSEError writes an error event to the SSE stream and closes it.
func writeVocabSSEError(writer *helpers.NDJSONWriter, err error) {
	writeSSEError(writer, err)
}

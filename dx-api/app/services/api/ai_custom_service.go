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

// AI generation cost consts.
const aiGenerateCost = 5

// Concurrency limits for SSE batch operations.
const (
	breakConcurrencyLimit    = 20
	genItemsConcurrencyLimit = 50
)

// Difficulty level descriptions (CEFR).
var difficultyDescriptions = map[string]string{
	"a1-a2": "A1-A2 (beginner: simple vocabulary, short sentences, present tense, common everyday words)",
	"b1-b2": "B1-B2 (intermediate: varied vocabulary, compound sentences, multiple tenses, some idiomatic expressions)",
	"c1-c2": "C1-C2 (advanced: sophisticated vocabulary, complex sentence structures, nuanced expressions, academic/literary language)",
}

// Error sentinels for AI custom operations.
var (
	ErrAIServiceUnavailable = errors.New("AI service unavailable")
	ErrModerationWarning    = errors.New("moderation warning")
	ErrEmptyContent         = errors.New("content is empty")
	ErrFormatCountExceeded  = errors.New("format count exceeded")
)

// GenerateMetadataResult holds the response from story generation.
type GenerateMetadataResult struct {
	Generated  string `json:"generated,omitempty"`
	SourceType string `json:"sourceType,omitempty"`
	Warning    string `json:"warning,omitempty"`
}

// FormatMetadataResult holds the response from content formatting.
type FormatMetadataResult struct {
	Formatted   string   `json:"formatted,omitempty"`
	SourceTypes []string `json:"sourceTypes,omitempty"`
	Warning     string   `json:"warning,omitempty"`
}

// SSEProgressEvent is sent to the client during SSE streaming.
type SSEProgressEvent struct {
	Done      int    `json:"done"`
	Total     int    `json:"total"`
	Status    string `json:"status,omitempty"`
	Processed int    `json:"processed,omitempty"`
	Failed    int    `json:"failed,omitempty"`
	Complete  bool   `json:"complete,omitempty"`
}

// --- GenerateMetadata --- (unchanged from original)

func GenerateMetadata(userID string, difficulty string, keywords []string) (*GenerateMetadataResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	if err := ConsumeBeans(userID, aiGenerateCost, consts.BeanSlugAIGenerateConsume, consts.BeanReasonAIGenerateConsume); err != nil {
		return nil, err
	}

	levelDesc, ok := difficultyDescriptions[difficulty]
	if !ok {
		levelDesc = difficultyDescriptions["a1-a2"]
	}

	prompt := buildGeneratePrompt(levelDesc)
	userMsg := "Keywords: " + strings.Join(keywords, ", ")

	result, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
		Messages: []helpers.DeepSeekMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: userMsg},
		},
		Temperature: 0.7,
	})
	if err != nil {
		_ = RefundBeans(userID, aiGenerateCost, consts.BeanSlugAIGenerateRefund, consts.BeanReasonAIGenerateRefund)
		return nil, err
	}

	if rest, ok := strings.CutPrefix(result, "WARNING:"); ok {
		return &GenerateMetadataResult{Warning: strings.TrimSpace(rest)}, nil
	}

	return &GenerateMetadataResult{
		Generated:  result,
		SourceType: SourceTypeSentence,
	}, nil
}

func buildGeneratePrompt(levelDesc string) string {
	return `You are a story writer for an English learning app. Your job is to generate a short English story for language learners.

STEP 1 — CONTENT MODERATION (do this FIRST):
Check if the provided keywords contain any insulting, violent, sexually explicit, or otherwise inappropriate/sensitive material.
If they do, respond ONLY with: WARNING:包含不适当内容，请修改后重试
Do NOT generate any story. Stop here.

STEP 2 — GENERATE STORY:
Write a short, coherent English story that:
- Uses CEFR level ` + levelDesc + ` appropriate vocabulary and grammar
- Naturally incorporates ALL the provided keywords into the story
- Contains at most 20 sentences. It can be fewer than 20 but NEVER more than 20. This is a hard limit.
- Tells a complete, engaging narrative with a beginning, middle, and end
- Is suitable for English language learners

RULES:
- Output ONLY the story text. No title, no explanations, no headers, no numbering, no markdown.
- Each sentence must be on its own line.
- Each line must contain exactly ONE sentence.
- Do not include empty lines between sentences.
- Do not repeat sentences.
- Keep each sentence under 200 characters.`
}

// --- FormatMetadata --- (unchanged from original — preserves [S]/[V] mixed input)

func FormatMetadata(userID string, content string, formatType string) (*FormatMetadataResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}
	wordCount := helpers.CountWords(content)
	if wordCount == 0 {
		return nil, ErrEmptyContent
	}

	consumeSlug, consumeReason, refundSlug, refundReason := formatBeanSlugs(formatType)

	if err := ConsumeBeans(userID, wordCount, consumeSlug, consumeReason); err != nil {
		return nil, err
	}

	prompt := buildFormatPrompt(formatType)

	result, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
		Messages: []helpers.DeepSeekMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: content},
		},
		Temperature: 0.1,
	})
	if err != nil {
		_ = RefundBeans(userID, wordCount, refundSlug, refundReason)
		return nil, err
	}

	if rest, ok := strings.CutPrefix(result, "WARNING:"); ok {
		return &FormatMetadataResult{Warning: strings.TrimSpace(rest)}, nil
	}

	formatted, sourceTypes := parseFormattedLines(result)
	if formatted == "" {
		_ = RefundBeans(userID, wordCount, refundSlug, refundReason)
		return nil, helpers.ErrDeepSeekEmpty
	}

	if warning := validateFormatCounts(sourceTypes); warning != "" {
		return &FormatMetadataResult{Warning: warning}, nil
	}

	return &FormatMetadataResult{
		Formatted:   formatted,
		SourceTypes: sourceTypes,
	}, nil
}

func formatBeanSlugs(formatType string) (consumeSlug, consumeReason, refundSlug, refundReason string) {
	if formatType == SourceTypeSentence {
		return consts.BeanSlugAIFormatSentenceConsume, consts.BeanReasonAIFormatSentenceConsume,
			consts.BeanSlugAIFormatSentenceRefund, consts.BeanReasonAIFormatSentenceRefund
	}
	return consts.BeanSlugAIFormatVocabConsume, consts.BeanReasonAIFormatVocabConsume,
		consts.BeanSlugAIFormatVocabRefund, consts.BeanReasonAIFormatVocabRefund
}

func buildFormatPrompt(formatType string) string {
	formatLabel := "词汇"
	if formatType == SourceTypeSentence {
		formatLabel = "语句"
	}

	formatRule := `- If the content contains Chinese text: output alternating lines of English word/phrase followed by its Chinese translation.
- If the content contains NO Chinese text: output English words/phrases only, one per line.`
	mismatchRule := "If the content consists mostly of full sentences with punctuation, respond ONLY with: WARNING:内容看起来是语句而非词汇，请使用「语句格式化」按钮"

	if formatType == SourceTypeSentence {
		formatRule = `- If the content contains Chinese text: output alternating lines of English sentence followed by its Chinese translation.
- If the content contains NO Chinese text: output English sentences only, one per line.`
		mismatchRule = "If the content consists mostly of single words or short phrases without sentence structure, respond ONLY with: WARNING:内容看起来是词汇而非语句，请使用「词汇格式化」按钮"
	}

	return `You are a content formatter for an English learning app. Your job is to clean up and reformat messy user input into a strict line-by-line format for ` + formatLabel + `.

STEP 1 — CONTENT MODERATION (do this FIRST):
Check if the content contains any insulting, violent, sexually explicit, or otherwise inappropriate/sensitive material.
If it does, respond ONLY with: WARNING:内容包含不适当内容，请修改后重试
Do NOT format the content. Stop here.

STEP 2 — TYPE MISMATCH CHECK:
` + mismatchRule + `

STEP 3 — FORMAT WITH PER-LINE TYPE MARKERS:
` + formatRule + `

For EACH English line, prefix it with a type marker:
- [S] for complete sentences (has subject + verb, expresses a complete thought)
- [V] for vocabulary items (single words, short phrases, or expressions without sentence structure)

Chinese translation lines must NOT have any prefix marker.

Example output with Chinese translations:
[S] I like the food.
我喜欢这个食物。
[V] food
食物
[V] name
名字

Example output without Chinese translations:
[S] I like the food.
[V] food
[V] name

RULES:
- Every English line MUST start with [S] or [V] prefix.
- Chinese translation lines must NOT have any prefix.
- Output ONLY the formatted text with markers. No explanations, headers, numbering, or markdown.
- Remove duplicates.
- Fix obvious spelling errors in English.
- Preserve the original meaning. Do not add or remove content.
- CRITICAL: Each line must contain exactly ONE sentence (or one word/phrase or one translation). Sentences are often separated by punctuation like periods (.), question marks (?), or exclamation marks (!), but not always — some sentences have no ending punctuation. Use meaning and grammar to identify sentence boundaries. If the input has multiple sentences on one line, split them so each sentence is on its own line. Never combine two or more sentences on the same line.
- Remove any empty lines.`
}

func parseFormattedLines(result string) (string, []string) {
	lines := strings.Split(result, "\n")
	var cleanLines []string
	var sourceTypes []string

	for _, line := range lines {
		line = strings.TrimRight(line, " \t\r")
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[S] ") {
			sourceTypes = append(sourceTypes, SourceTypeSentence)
			cleanLines = append(cleanLines, line[4:])
		} else if strings.HasPrefix(line, "[V] ") {
			sourceTypes = append(sourceTypes, SourceTypeVocab)
			cleanLines = append(cleanLines, line[4:])
		} else {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n"), sourceTypes
}

func validateFormatCounts(sourceTypes []string) string {
	sentenceCount := 0
	vocabCount := 0
	for _, t := range sourceTypes {
		switch t {
		case SourceTypeSentence:
			sentenceCount++
		case SourceTypeVocab:
			vocabCount++
		}
	}

	if sentenceCount > MaxSentences {
		return fmt.Sprintf("格式化后有 %d 条语句，超过 %d 条上限。为保证最佳学习体验，请精简内容", sentenceCount, MaxSentences)
	}
	if vocabCount > MaxVocab {
		return fmt.Sprintf("格式化后有 %d 条词汇，超过 %d 条上限。请精简内容", vocabCount, MaxVocab)
	}
	return ""
}

// --- BreakMetadata --- (rewritten — direct queries on content_metas/content_items)

func BreakMetadata(userID, gameLevelID string, writer *helpers.NDJSONWriter) {
	if err := requireVip(userID); err != nil {
		writeSSEError(writer, err)
		return
	}
	game, level, err := verifyLevelOwnership(userID, gameLevelID)
	if err != nil {
		writeSSEError(writer, err)
		return
	}
	if game.Status == consts.GameStatusPublished {
		writeSSEError(writer, ErrGamePublished)
		return
	}
	gameID := level.GameID

	// Load unbroken metas in order — directly from content_metas
	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("is_break_done", false).
		Order(`"order" ASC`).
		Get(&metas); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
		return
	}

	if len(metas) == 0 {
		_ = writer.Write(SSEProgressEvent{Done: 0, Total: 0, Processed: 0, Failed: 0, Complete: true})
		writer.Close()
		return
	}

	metaWordCounts := make([]int, len(metas))
	totalCost := 0
	for i, m := range metas {
		wc := helpers.CountWords(m.SourceData)
		metaWordCounts[i] = wc
		totalCost += wc
	}

	if totalCost == 0 {
		writeSSEError(writer, ErrEmptyContent)
		return
	}

	if err := ConsumeBeans(userID, totalCost, consts.BeanSlugAIBreakConsume, consts.BeanReasonAIBreakConsume); err != nil {
		writeSSEError(writer, err)
		return
	}

	var failedWords int64
	var processed int64
	var failed int64

	sem := make(chan struct{}, breakConcurrencyLimit)
	var wg sync.WaitGroup
	var done int64

	total := len(metas)

	for i, meta := range metas {
		wg.Add(1)
		sem <- struct{}{}

		go func(m models.ContentMeta, idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			success := processBreakMeta(m, gameID, gameLevelID)
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
		_ = RefundBeans(userID, fw, consts.BeanSlugAIBreakRefund, consts.BeanReasonAIBreakRefund)
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

// processBreakMeta calls DeepSeek to split a meta into content_items rows.
// Item orders fan out from the parent meta's order in increments of 10.
func processBreakMeta(meta models.ContentMeta, gameID, gameLevelID string) bool {
	userMsg := "English: " + meta.SourceData
	if meta.Translation != nil && *meta.Translation != "" {
		userMsg += "\nChinese translation: " + *meta.Translation
	}

	result, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
		Messages: []helpers.DeepSeekMessage{
			{Role: "system", Content: breakPrompt},
			{Role: "user", Content: userMsg},
		},
		Temperature: 0.1,
	})
	if err != nil {
		return false
	}

	items, err := helpers.ParseAIJSONArray(result)
	if err != nil || len(items) == 0 {
		return false
	}

	startOrder := meta.Order + 10

	for i, raw := range items {
		var unit struct {
			Content     string `json:"content"`
			ContentType string `json:"contentType"`
			Translation string `json:"translation"`
		}
		if err := json.Unmarshal(raw, &unit); err != nil {
			continue
		}

		id := uuid.Must(uuid.NewV7()).String()
		metaID := meta.ID
		var translation *string
		if unit.Translation != "" {
			translation = &unit.Translation
		}

		item := models.ContentItem{
			ID:            id,
			GameID:        gameID,
			GameLevelID:   gameLevelID,
			ContentMetaID: &metaID,
			Content:       unit.Content,
			ContentType:   unit.ContentType,
			Translation:   translation,
			Order:         startOrder + float64(i*10),
		}
		if err := facades.Orm().Query().Create(&item); err != nil {
			return false
		}
	}

	if _, err := facades.Orm().Query().Model(&models.ContentMeta{}).
		Where("id", meta.ID).
		Update("is_break_done", true); err != nil {
		return false
	}

	return true
}

var breakPrompt = `You are a language learning content processor. Your job is to analyze an English text and break it into structured learning units.

STEP 1 — DETERMINE TYPE:
- If the input is a complete sentence (has subject + verb, typically ends with punctuation like . ? !), treat it as a SENTENCE.
- Otherwise, treat it as a WORD or PHRASE.

STEP 2 — GENERATE LEARNING UNITS:

If SENTENCE (sequential left-to-right splitting):
Split the sentence from left to right, producing units in reading order. Use these types:
- "word": a single content word (noun, verb, adjective, adverb, pronoun)
- "block": a progressive combination building from the start of the sentence
- "phrase": a natural word grouping (collocation, prepositional phrase, noun phrase, etc.)
- "sentence": the full original sentence

IMPORTANT RULES:
- Articles (a, an, the) and prepositions (e.g. in, on, at, to, for, with, of, by, from, about, into, through, between, etc.) must NEVER be standalone "word" units. Always group them into the nearest phrase or block with the following content word(s).
- Linking verbs (am, is, are, was, were, be, been, being) must NEVER be standalone "word" units. Always group them with adjacent content — e.g. "is tall" (phrase), "He is" (block), "is reading" (phrase).
- Follow the natural reading order of the sentence. Do NOT group all words first, then all blocks, then all phrases. Instead, split sequentially left to right.
- Each segment of the sentence should appear in exactly one unit at its most granular level (word or phrase), then optionally in cumulative blocks and finally the full sentence.

Example for "I like the food.":
1. "I" (word)
2. "like" (word)
3. "I like" (block)
4. "the food" (phrase) — article grouped with its noun
5. "I like the food." (sentence)

Example for "She went to the park.":
1. "She" (word)
2. "went" (word)
3. "She went" (block)
4. "to the park" (phrase) — preposition + article grouped with noun
5. "She went to the park." (sentence)

Example for "He is reading a book in the library.":
1. "He" (word)
2. "is reading" (phrase) — auxiliary + verb grouped together
3. "He is reading" (block)
4. "a book" (phrase) — article grouped with noun
5. "He is reading a book" (block)
6. "in the library" (phrase) — preposition + article + noun grouped
7. "He is reading a book in the library." (sentence)

If NOT a sentence:
Generate a single unit with contentType "word" (single word) or "phrase" (multi-word expression) or "block" (neither a valid word nor a valid phrase).

Each unit needs:
- content: the text of this unit
- contentType: one of "word", "block", "phrase", "sentence"
- translation: Chinese translation of the entire unit

OUTPUT FORMAT:
Output ONLY a valid JSON array. No markdown code fences, no explanation, no extra text.

Example output:
[
  {"content": "I", "contentType": "word", "translation": "我"},
  {"content": "like", "contentType": "word", "translation": "喜欢"},
  {"content": "I like", "contentType": "block", "translation": "我喜欢"},
  {"content": "the food", "contentType": "phrase", "translation": "这食物"},
  {"content": "I like the food.", "contentType": "sentence", "translation": "我喜欢这食物。"}
]`

// --- GenerateContentItems --- (rewritten — direct queries)

func GenerateContentItems(userID, gameLevelID string, writer *helpers.NDJSONWriter) {
	if err := requireVip(userID); err != nil {
		writeSSEError(writer, err)
		return
	}
	game, level, err := verifyLevelOwnership(userID, gameLevelID)
	if err != nil {
		writeSSEError(writer, err)
		return
	}
	if game.Status == consts.GameStatusPublished {
		writeSSEError(writer, ErrGamePublished)
		return
	}
	_ = level

	var metas []models.ContentMeta
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("is_break_done", true).
		Order(`"order" ASC`).
		Get(&metas); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load metas: %w", err))
		return
	}
	if len(metas) == 0 {
		_ = writer.Write(SSEProgressEvent{Done: 0, Total: 0, Processed: 0, Failed: 0, Complete: true})
		writer.Close()
		return
	}

	metaIDs := make([]string, 0, len(metas))
	metaMap := make(map[string]models.ContentMeta)
	for _, m := range metas {
		metaIDs = append(metaIDs, m.ID)
		metaMap[m.ID] = m
	}

	var pendingItems []models.ContentItem
	if err := facades.Orm().Query().
		Where("game_level_id", gameLevelID).
		Where("content_meta_id IN ?", metaIDs).
		Where("items IS NULL").
		Get(&pendingItems); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load pending items: %w", err))
		return
	}

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
		writeSSEError(writer, ErrEmptyContent)
		return
	}

	if err := ConsumeBeans(userID, totalCost, consts.BeanSlugAIGenItemsConsume, consts.BeanReasonAIGenItemsConsume); err != nil {
		writeSSEError(writer, err)
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
		wg.Add(1)
		sem <- struct{}{}

		go func(m models.ContentMeta) {
			defer wg.Done()
			defer func() { <-sem }()

			items := pendingByMeta[m.ID]
			success := processGenItems(m, items)
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
		_ = RefundBeans(userID, fw, consts.BeanSlugAIGenItemsRefund, consts.BeanReasonAIGenItemsRefund)
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

func processGenItems(meta models.ContentMeta, existingItems []models.ContentItem) bool {
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

var genItemsPrompt = `You are a language learning content processor. You will receive a list of learning units (each with content and contentType). Your job is to break each unit into individual word/punctuation items.

For each unit, produce an "items" array where each element represents one word or punctuation mark:
- position: 1-based index
- item: the word or punctuation character
- phonetic: {"uk": "IPA notation", "us": "IPA notation"} — set to null for punctuation marks
- pos: Chinese part-of-speech label like "名词", "动词", "形容词", "副词", "代词", "介词", "连词", "冠词", "感叹词", "助动词" etc — set to null for punctuation marks
- definition: Chinese definition of the word — set to empty string for punctuation marks
- answer: false for punctuation marks, proper names, place names, and abbreviations; true for all other words

Return a JSON array where each element has:
- content: the unit text (echo back exactly as given)
- items: the items array as described above

OUTPUT FORMAT:
Output ONLY a valid JSON array. No markdown code fences, no explanation, no extra text.

Example input units:
[{"content": "I", "contentType": "word"}, {"content": "I like", "contentType": "block"}]

Example output:
[
  {
    "content": "I",
    "items": [
      {"position": 1, "item": "I", "phonetic": {"uk": "/aɪ/", "us": "/aɪ/"}, "pos": "代词", "definition": "我", "answer": true}
    ]
  },
  {
    "content": "I like",
    "items": [
      {"position": 1, "item": "I", "phonetic": {"uk": "/aɪ/", "us": "/aɪ/"}, "pos": "代词", "definition": "我", "answer": true},
      {"position": 2, "item": "like", "phonetic": {"uk": "/laɪk/", "us": "/laɪk/"}, "pos": "动词", "definition": "喜欢", "answer": true}
    ]
  }
]`

// --- Helpers --- (verifyLevelOwnership unchanged)

func verifyLevelOwnership(userID, gameLevelID string) (*models.Game, *models.GameLevel, error) {
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

	return game, &level, nil
}

func writeSSEError(writer *helpers.NDJSONWriter, err error) {
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

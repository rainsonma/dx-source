package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"
)

// ContentVocabData is the public response shape — no createdBy / lastEditedBy / isVerified.
type ContentVocabData struct {
	ID          string  `json:"id"`
	Content     string  `json:"content"`
	UkPhonetic  *string `json:"ukPhonetic"`
	UsPhonetic  *string `json:"usPhonetic"`
	UkAudioURL  *string `json:"ukAudioUrl"`
	UsAudioURL  *string `json:"usAudioUrl"`
	Definition  *string `json:"definition"`
	Explanation *string `json:"explanation"`
	CreatedAt   any     `json:"createdAt"`
	UpdatedAt   any     `json:"updatedAt"`
}

func vocabToData(v *models.ContentVocab) *ContentVocabData {
	return &ContentVocabData{
		ID:          v.ID,
		Content:     v.Content,
		UkPhonetic:  v.UkPhonetic,
		UsPhonetic:  v.UsPhonetic,
		UkAudioURL:  v.UkAudioURL,
		UsAudioURL:  v.UsAudioURL,
		Definition:  v.Definition,
		Explanation: v.Explanation,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

// VocabInput is the create/update payload.
type VocabInput struct {
	Content     string              `json:"content"`
	Definition  []map[string]string `json:"definition"`
	UkPhonetic  *string             `json:"ukPhonetic"`
	UsPhonetic  *string             `json:"usPhonetic"`
	UkAudioURL  *string             `json:"ukAudioUrl"`
	UsAudioURL  *string             `json:"usAudioUrl"`
	Explanation *string             `json:"explanation"`
}

// CreateVocabResult — one entry in the batch-create response.
type CreateVocabResult struct {
	Vocab     *ContentVocabData `json:"vocab"`
	WasReused bool              `json:"wasReused"`
}

// ErrDuplicateVocab is returned when an UPDATE would collide with another of the user's vocabs.
var ErrDuplicateVocab = errors.New("duplicate vocab content for this user")

// ListUserVocabs returns a paginated list of the user's vocab pool.
func ListUserVocabs(userID, cursor, search string, limit int) ([]ContentVocabData, string, bool, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	q := facades.Orm().Query().Model(&models.ContentVocab{}).Where("user_id", userID)
	if search != "" {
		q = q.Where("content_key LIKE ?", "%"+NormalizeVocabContent(search)+"%")
	}
	if cursor != "" {
		var cursorRow models.ContentVocab
		if err := facades.Orm().Query().Where("id", cursor).First(&cursorRow); err == nil && cursorRow.ID != "" {
			q = q.Where("(created_at < ? OR (created_at = ? AND id < ?))", cursorRow.CreatedAt, cursorRow.CreatedAt, cursor)
		}
	}

	var rows []models.ContentVocab
	if err := q.Order("created_at DESC").Order("id DESC").Limit(limit + 1).Get(&rows); err != nil {
		return nil, "", false, fmt.Errorf("failed to list user vocabs: %w", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	nextCursor := ""
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].ID
	}
	out := make([]ContentVocabData, 0, len(rows))
	for i := range rows {
		out = append(out, *vocabToData(&rows[i]))
	}
	return out, nextCursor, hasMore, nil
}

// GetUserVocabByContent returns the user's vocab matching content (case-insensitive).
// Returns nil, nil if not found.
func GetUserVocabByContent(userID, content string) (*ContentVocabData, error) {
	key := NormalizeVocabContent(content)
	if key == "" {
		return nil, nil
	}
	var v models.ContentVocab
	if err := facades.Orm().Query().Where("user_id", userID).Where("content_key", key).First(&v); err != nil || v.ID == "" {
		return nil, nil
	}
	return vocabToData(&v), nil
}

// CreateUserVocab creates a vocab in the user's pool. Idempotent by content_key.
func CreateUserVocab(userID string, in VocabInput) (*CreateVocabResult, error) {
	if err := ValidateVocabContent(in.Content); err != nil {
		return nil, err
	}
	if err := ValidatePosEntries(in.Definition); err != nil {
		return nil, err
	}
	key := NormalizeVocabContent(in.Content)

	// Idempotent — return existing if user already has this content_key.
	var existing models.ContentVocab
	if err := facades.Orm().Query().Where("user_id", userID).Where("content_key", key).First(&existing); err == nil && existing.ID != "" {
		return &CreateVocabResult{Vocab: vocabToData(&existing), WasReused: true}, nil
	}

	defJSON, err := json.Marshal(in.Definition)
	if err != nil {
		return nil, fmt.Errorf("definition marshal: %w", err)
	}
	defStr := string(defJSON)

	v := models.ContentVocab{
		ID:          uuid.Must(uuid.NewV7()).String(),
		UserID:      userID,
		Content:     in.Content,
		ContentKey:  key,
		UkPhonetic:  NormalizePhonetic(in.UkPhonetic),
		UsPhonetic:  NormalizePhonetic(in.UsPhonetic),
		UkAudioURL:  in.UkAudioURL,
		UsAudioURL:  in.UsAudioURL,
		Definition:  &defStr,
		Explanation: in.Explanation,
	}
	if err := facades.Orm().Query().Create(&v); err != nil {
		return nil, fmt.Errorf("failed to create user vocab: %w", err)
	}
	return &CreateVocabResult{Vocab: vocabToData(&v), WasReused: false}, nil
}

// CreateUserVocabsBatch creates multiple vocabs in the user's pool sequentially.
func CreateUserVocabsBatch(userID string, inputs []VocabInput) ([]CreateVocabResult, error) {
	out := make([]CreateVocabResult, 0, len(inputs))
	for _, in := range inputs {
		res, err := CreateUserVocab(userID, in)
		if err != nil {
			return nil, fmt.Errorf("entry %q: %w", in.Content, err)
		}
		out = append(out, *res)
	}
	return out, nil
}

// UpdateUserVocab fully overwrites a user's vocab row.
func UpdateUserVocab(userID, vocabID string, in VocabInput) (*ContentVocabData, error) {
	if err := ValidateVocabContent(in.Content); err != nil {
		return nil, err
	}
	if err := ValidatePosEntries(in.Definition); err != nil {
		return nil, err
	}

	var v models.ContentVocab
	if err := facades.Orm().Query().Where("id", vocabID).Where("user_id", userID).First(&v); err != nil || v.ID == "" {
		return nil, ErrVocabNotFound
	}

	newKey := NormalizeVocabContent(in.Content)
	if newKey != v.ContentKey {
		// Collision check: another vocab of same user with that content_key
		var collision models.ContentVocab
		if err := facades.Orm().Query().Where("user_id", userID).Where("content_key", newKey).Where("id != ?", vocabID).First(&collision); err == nil && collision.ID != "" {
			return nil, ErrDuplicateVocab
		}
	}

	defJSON, err := json.Marshal(in.Definition)
	if err != nil {
		return nil, fmt.Errorf("definition marshal: %w", err)
	}

	updates := map[string]any{
		"content":      in.Content,
		"content_key":  newKey,
		"definition":   string(defJSON),
		"uk_phonetic":  NormalizePhonetic(in.UkPhonetic),
		"us_phonetic":  NormalizePhonetic(in.UsPhonetic),
		"uk_audio_url": in.UkAudioURL,
		"us_audio_url": in.UsAudioURL,
		"explanation":  in.Explanation,
	}
	if _, err := facades.Orm().Query().Model(&models.ContentVocab{}).Where("id", vocabID).Update(updates); err != nil {
		return nil, fmt.Errorf("failed to update user vocab: %w", err)
	}

	var updated models.ContentVocab
	if err := facades.Orm().Query().Where("id", vocabID).First(&updated); err != nil {
		return nil, fmt.Errorf("failed to reload user vocab: %w", err)
	}
	return vocabToData(&updated), nil
}

// DeleteUserVocab soft-deletes a user's vocab row.
func DeleteUserVocab(userID, vocabID string) error {
	var v models.ContentVocab
	if err := facades.Orm().Query().Where("id", vocabID).Where("user_id", userID).First(&v); err != nil || v.ID == "" {
		return ErrVocabNotFound
	}
	if _, err := facades.Orm().Query().Exec(
		`UPDATE content_vocabs SET deleted_at = NOW() WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		vocabID, userID,
	); err != nil {
		return fmt.Errorf("failed to delete user vocab: %w", err)
	}
	return nil
}

// GenerateVocabWords generates 15-25 English words/phrases from keywords at the given CEFR difficulty.
// Bean cost = aiGenerateCost (5). Refunds on AI failure.
func GenerateVocabWords(userID string, keywords []string, difficulty string) ([]string, error) {
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

	prompt := buildVocabWordsPrompt(levelDesc)
	userMsg := "Keywords: " + strings.Join(keywords, ", ")

	raw, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
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

	// Check moderation warning before parsing
	if _, ok := strings.CutPrefix(raw, "WARNING:"); ok {
		_ = RefundBeans(userID, aiGenerateCost, consts.BeanSlugAIVocabGenerateRefund, consts.BeanReasonAIVocabGenerateRefund)
		return nil, ErrModerationWarning
	}

	items, err := helpers.ParseAIJSONArray(raw)
	if err != nil {
		_ = RefundBeans(userID, aiGenerateCost, consts.BeanSlugAIVocabGenerateRefund, consts.BeanReasonAIVocabGenerateRefund)
		return nil, helpers.ErrDeepSeekEmpty
	}

	words := make([]string, 0, len(items))
	for _, item := range items {
		var s string
		if err := json.Unmarshal(item, &s); err == nil && s != "" {
			words = append(words, s)
		}
	}

	if len(words) == 0 {
		_ = RefundBeans(userID, aiGenerateCost, consts.BeanSlugAIVocabGenerateRefund, consts.BeanReasonAIVocabGenerateRefund)
		return nil, helpers.ErrDeepSeekEmpty
	}

	return words, nil
}

func buildVocabWordsPrompt(levelDesc string) string {
	return `You are a vocabulary generator for an English learning app. Given keywords, produce a JSON array of 15-25 English words or short phrases related to those keywords.

STEP 1 — CONTENT MODERATION (do this FIRST):
Check if the provided keywords contain any insulting, violent, sexually explicit, or otherwise inappropriate/sensitive material.
If they do, respond ONLY with: WARNING:包含不适当内容，请修改后重试
Do NOT generate any vocabulary. Stop here.

STEP 2 — GENERATE VOCABULARY:
Generate 15-25 English words or short phrases that:
- Are related to the provided keywords
- Use CEFR level ` + levelDesc + ` appropriate vocabulary
- Include a mix of single words and short phrases (2-3 words max)
- Are suitable for English language learners

OUTPUT FORMAT:
Output ONLY a JSON array of strings. Each string is one English word or phrase.
No markdown code fences, no explanation, no extra text.

Example output:
["fast","quick","sprint","run fast","speed up","rapid","swift","acceleration","pace","velocity","dash","rush","hurry","brisk","agile"]`
}

// CreateVocabsFromWords: for each word, dedup against the user's pool (reuse existing),
// then run DeepSeek in parallel to enrich new words with phonetics/definition/explanation.
// Bean cost = 1 per NEW word. Refunds failed AI enrichments.
const vocabFromWordsConcurrencyLimit = 20

func CreateVocabsFromWords(userID string, words []string) ([]CreateVocabResult, error) {
	if err := requireVip(userID); err != nil {
		return nil, err
	}

	// Deduplicate input words (preserve order, case-insensitive)
	seen := make(map[string]bool)
	unique := make([]string, 0, len(words))
	for _, w := range words {
		key := NormalizeVocabContent(w)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, w)
	}

	if len(unique) == 0 {
		return []CreateVocabResult{}, nil
	}

	// Check which words already exist in the user's pool
	type wordCheck struct {
		word     string
		existing *models.ContentVocab
	}
	checks := make([]wordCheck, len(unique))
	for i, w := range unique {
		checks[i].word = w
		key := NormalizeVocabContent(w)
		var existing models.ContentVocab
		if err := facades.Orm().Query().Where("user_id", userID).Where("content_key", key).First(&existing); err == nil && existing.ID != "" {
			existing := existing
			checks[i].existing = &existing
		}
	}

	// Identify new words and charge beans
	newIndices := make([]int, 0)
	for i, c := range checks {
		if c.existing == nil {
			newIndices = append(newIndices, i)
		}
	}

	newCount := len(newIndices)
	if newCount > 0 {
		if err := ConsumeBeans(userID, newCount, consts.BeanSlugAIVocabGenerateConsume, consts.BeanReasonAIVocabGenerateConsume); err != nil {
			return nil, err
		}
	}

	// For new words: run DeepSeek enrichment in parallel
	type enrichResult struct {
		index int
		input VocabInput
		err   error
	}
	enrichResults := make([]enrichResult, len(newIndices))

	sem := make(chan struct{}, vocabFromWordsConcurrencyLimit)
	var wg sync.WaitGroup

	for i, idx := range newIndices {
		wg.Add(1)
		sem <- struct{}{}
		go func(resultIdx, checkIdx int) {
			defer wg.Done()
			defer func() { <-sem }()

			word := checks[checkIdx].word
			inp, err := enrichVocabWord(word)
			enrichResults[resultIdx] = enrichResult{index: checkIdx, input: inp, err: err}
		}(i, idx)
	}
	wg.Wait()

	// Count failed enrichments and refund
	failedCount := 0
	for i := range enrichResults {
		if enrichResults[i].err != nil {
			failedCount++
			// Fall back to empty fields — still create the row
			enrichResults[i].input = VocabInput{
				Content:    checks[enrichResults[i].index].word,
				Definition: []map[string]string{},
			}
		}
	}
	if failedCount > 0 {
		_ = RefundBeans(userID, failedCount, consts.BeanSlugAIVocabGenerateRefund, consts.BeanReasonAIVocabGenerateRefund)
	}

	// Build enrich map by check index
	enrichMap := make(map[int]VocabInput)
	for _, er := range enrichResults {
		enrichMap[er.index] = er.input
	}

	// Create results in original order
	out := make([]CreateVocabResult, 0, len(checks))
	for i, c := range checks {
		if c.existing != nil {
			out = append(out, CreateVocabResult{Vocab: vocabToData(c.existing), WasReused: true})
			continue
		}
		inp := enrichMap[i]
		res, err := CreateUserVocab(userID, inp)
		if err != nil {
			// Idempotent: if another concurrent request raced us, return existing
			existing, getErr := GetUserVocabByContent(userID, c.word)
			if getErr == nil && existing != nil {
				out = append(out, CreateVocabResult{Vocab: existing, WasReused: true})
			}
			// Otherwise skip this word silently
			continue
		}
		out = append(out, *res)
	}

	return out, nil
}

// enrichVocabWord calls DeepSeek to get phonetics, definition, and explanation for a single word.
func enrichVocabWord(word string) (VocabInput, error) {
	userMsg := "Word: " + word

	raw, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
		Messages: []helpers.DeepSeekMessage{
			{Role: "system", Content: vocabEnrichPrompt},
			{Role: "user", Content: userMsg},
		},
		Temperature: 0.1,
	})
	if err != nil {
		return VocabInput{}, err
	}

	items, err := helpers.ParseAIJSONArray(raw)
	if err != nil || len(items) == 0 {
		return VocabInput{}, fmt.Errorf("enrich: invalid JSON for %q", word)
	}

	var obj struct {
		UkPhonetic  string              `json:"ukPhonetic"`
		UsPhonetic  string              `json:"usPhonetic"`
		Definition  []map[string]string `json:"definition"`
		Explanation string              `json:"explanation"`
	}
	if err := json.Unmarshal(items[0], &obj); err != nil {
		return VocabInput{}, fmt.Errorf("enrich: unmarshal failed for %q: %w", word, err)
	}

	if err := ValidatePosEntries(obj.Definition); err != nil {
		obj.Definition = []map[string]string{}
	}

	inp := VocabInput{
		Content:    word,
		Definition: obj.Definition,
	}
	if obj.UkPhonetic != "" {
		inp.UkPhonetic = NormalizePhonetic(&obj.UkPhonetic)
	}
	if obj.UsPhonetic != "" {
		inp.UsPhonetic = NormalizePhonetic(&obj.UsPhonetic)
	}
	if obj.Explanation != "" {
		inp.Explanation = &obj.Explanation
	}
	return inp, nil
}

var vocabEnrichPrompt = `You are a vocabulary enrichment service for an English learning app. Given a single English word or phrase, return a JSON array containing ONE object with phonetics, part-of-speech definition, and a brief Chinese explanation.

The object must have:
- ukPhonetic: IPA UK pronunciation, e.g. "/fɑːst/"
- usPhonetic: IPA US pronunciation, e.g. "/fæst/"
- definition: array of single-key objects mapping POS to Chinese gloss; allowed POS keys: n, v, adj, adv, prep, conj, pron, art, num, int, aux, det
- explanation: 1-2 sentence Chinese explanation with example usage

Output ONLY a valid JSON array containing exactly one object. No markdown code fences, no extra text.

Example for "fast":
[{"ukPhonetic":"/fɑːst/","usPhonetic":"/fæst/","definition":[{"adj":"快的"},{"v":"斋戒"}],"explanation":"形容速度快；也可作动词表示禁食。例：He runs fast."}]`

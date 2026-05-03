package api

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"
)

// ContentVocabData is the public response shape for canonical wiki entries.
type ContentVocabData struct {
	ID           string  `json:"id"`
	Content      string  `json:"content"`
	UkPhonetic   *string `json:"ukPhonetic"`
	UsPhonetic   *string `json:"usPhonetic"`
	UkAudioURL   *string `json:"ukAudioUrl"`
	UsAudioURL   *string `json:"usAudioUrl"`
	Definition   *string `json:"definition"`
	Explanation  *string `json:"explanation"`
	IsVerified   bool    `json:"isVerified"`
	CreatedBy    *string `json:"createdBy"`
	LastEditedBy *string `json:"lastEditedBy"`
}

func vocabToData(v *models.ContentVocab) *ContentVocabData {
	return &ContentVocabData{
		ID:           v.ID,
		Content:      v.Content,
		UkPhonetic:   v.UkPhonetic,
		UsPhonetic:   v.UsPhonetic,
		UkAudioURL:   v.UkAudioURL,
		UsAudioURL:   v.UsAudioURL,
		Definition:   v.Definition,
		Explanation:  v.Explanation,
		IsVerified:   v.IsVerified,
		CreatedBy:    v.CreatedBy,
		LastEditedBy: v.LastEditedBy,
	}
}

// GetContentVocabByContent returns the canonical wiki row matching content (case-insensitive).
// Returns nil, nil if not found.
func GetContentVocabByContent(content string) (*ContentVocabData, error) {
	key := NormalizeVocabContent(content)
	if key == "" {
		return nil, nil
	}
	var v models.ContentVocab
	if err := facades.Orm().Query().Where("content_key", key).First(&v); err != nil {
		return nil, nil
	}
	if v.ID == "" {
		return nil, nil
	}
	return vocabToData(&v), nil
}

// VocabComplementPatch is the request body for ComplementContentVocab.
// All fields optional; only non-nil fields are applied.
// Definition is the new POS entries to merge additively.
type VocabComplementPatch struct {
	Definition  []map[string]string `json:"definition,omitempty"`
	UkPhonetic  *string             `json:"ukPhonetic,omitempty"`
	UsPhonetic  *string             `json:"usPhonetic,omitempty"`
	UkAudioURL  *string             `json:"ukAudioUrl,omitempty"`
	UsAudioURL  *string             `json:"usAudioUrl,omitempty"`
	Explanation *string             `json:"explanation,omitempty"`
}

// ComplementContentVocab applies an additive merge: definition appends only
// new POS keys; phonetic/audio/explanation set only if currently null.
// Anyone may complement.
func ComplementContentVocab(userID, vocabID string, patch VocabComplementPatch) (*ContentVocabData, error) {
	var v models.ContentVocab
	if err := facades.Orm().Query().Where("id", vocabID).First(&v); err != nil || v.ID == "" {
		return nil, ErrVocabNotFound
	}

	beforeSnapshot, err := SnapshotVocab(&v)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}
	if patch.Definition != nil {
		if err := ValidatePosEntries(patch.Definition); err != nil {
			return nil, err
		}
		existing := ""
		if v.Definition != nil {
			existing = *v.Definition
		}
		merged, err := MergeDefinition(existing, patch.Definition)
		if err != nil {
			return nil, err
		}
		updates["definition"] = merged
	}
	if patch.UkPhonetic != nil && (v.UkPhonetic == nil || *v.UkPhonetic == "") {
		updates["uk_phonetic"] = *patch.UkPhonetic
	}
	if patch.UsPhonetic != nil && (v.UsPhonetic == nil || *v.UsPhonetic == "") {
		updates["us_phonetic"] = *patch.UsPhonetic
	}
	if patch.UkAudioURL != nil && (v.UkAudioURL == nil || *v.UkAudioURL == "") {
		updates["uk_audio_url"] = *patch.UkAudioURL
	}
	if patch.UsAudioURL != nil && (v.UsAudioURL == nil || *v.UsAudioURL == "") {
		updates["us_audio_url"] = *patch.UsAudioURL
	}
	if patch.Explanation != nil && (v.Explanation == nil || *v.Explanation == "") {
		updates["explanation"] = *patch.Explanation
	}

	if len(updates) == 0 {
		// Nothing to merge — return current state without writing an edit log.
		return vocabToData(&v), nil
	}
	updates["last_edited_by"] = userID

	if _, err := facades.Orm().Query().Model(&models.ContentVocab{}).
		Where("id", vocabID).Update(updates); err != nil {
		return nil, fmt.Errorf("failed to update content_vocab: %w", err)
	}

	// Reload + write audit
	var updated models.ContentVocab
	if err := facades.Orm().Query().Where("id", vocabID).First(&updated); err != nil {
		return nil, fmt.Errorf("failed to reload content_vocab: %w", err)
	}
	afterSnapshot, _ := SnapshotVocab(&updated)
	_ = WriteVocabEdit(nil, vocabID, userID, "complement", beforeSnapshot, afterSnapshot)

	return vocabToData(&updated), nil
}

// VocabReplacePatch is the request body for ReplaceContentVocab — full overwrite.
type VocabReplacePatch struct {
	Content     string              `json:"content"`
	Definition  []map[string]string `json:"definition"`
	UkPhonetic  *string             `json:"ukPhonetic"`
	UsPhonetic  *string             `json:"usPhonetic"`
	UkAudioURL  *string             `json:"ukAudioUrl"`
	UsAudioURL  *string             `json:"usAudioUrl"`
	Explanation *string             `json:"explanation"`
}

// ReplaceContentVocab fully overwrites the row, gated by CanReplaceVocab.
func ReplaceContentVocab(userID, vocabID string, patch VocabReplacePatch) (*ContentVocabData, error) {
	var v models.ContentVocab
	if err := facades.Orm().Query().Where("id", vocabID).First(&v); err != nil || v.ID == "" {
		return nil, ErrVocabNotFound
	}
	if !CanReplaceVocab(userID, &v) {
		return nil, ErrVocabNotEditable
	}
	if err := ValidateVocabContent(patch.Content); err != nil {
		return nil, err
	}
	if err := ValidatePosEntries(patch.Definition); err != nil {
		return nil, err
	}

	beforeSnapshot, _ := SnapshotVocab(&v)

	defJSON, err := json.Marshal(patch.Definition)
	if err != nil {
		return nil, fmt.Errorf("definition marshal: %w", err)
	}

	updates := map[string]any{
		"content":        patch.Content,
		"content_key":    NormalizeVocabContent(patch.Content),
		"definition":     string(defJSON),
		"uk_phonetic":    patch.UkPhonetic,
		"us_phonetic":    patch.UsPhonetic,
		"uk_audio_url":   patch.UkAudioURL,
		"us_audio_url":   patch.UsAudioURL,
		"explanation":    patch.Explanation,
		"last_edited_by": userID,
	}
	if _, err := facades.Orm().Query().Model(&models.ContentVocab{}).
		Where("id", vocabID).Update(updates); err != nil {
		return nil, fmt.Errorf("failed to replace content_vocab: %w", err)
	}

	var updated models.ContentVocab
	if err := facades.Orm().Query().Where("id", vocabID).First(&updated); err != nil {
		return nil, err
	}
	afterSnapshot, _ := SnapshotVocab(&updated)
	_ = WriteVocabEdit(nil, vocabID, userID, "replace", beforeSnapshot, afterSnapshot)

	return vocabToData(&updated), nil
}

// VerifyContentVocab toggles is_verified. Admin only.
func VerifyContentVocab(adminUserID, vocabID string, verified bool) (*ContentVocabData, error) {
	if !IsAdmin(adminUserID) {
		return nil, ErrVocabAdminOnly
	}
	var v models.ContentVocab
	if err := facades.Orm().Query().Where("id", vocabID).First(&v); err != nil || v.ID == "" {
		return nil, ErrVocabNotFound
	}
	beforeSnapshot, _ := SnapshotVocab(&v)

	if _, err := facades.Orm().Query().Model(&models.ContentVocab{}).
		Where("id", vocabID).Update(map[string]any{
		"is_verified":    verified,
		"last_edited_by": adminUserID,
	}); err != nil {
		return nil, fmt.Errorf("failed to verify content_vocab: %w", err)
	}

	var updated models.ContentVocab
	if err := facades.Orm().Query().Where("id", vocabID).First(&updated); err != nil {
		return nil, err
	}
	afterSnapshot, _ := SnapshotVocab(&updated)
	_ = WriteVocabEdit(nil, vocabID, adminUserID, "verify", beforeSnapshot, afterSnapshot)

	return vocabToData(&updated), nil
}

// --- AI enrichment SSE: GenerateContentVocabFields ---

// GenerateContentVocabFields enriches every content_vocabs row referenced by
// this level's game_vocabs that has uk_phonetic IS NULL.
func GenerateContentVocabFields(userID, gameLevelID string, writer *helpers.NDJSONWriter) {
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
	if !consts.IsVocabMode(game.Mode) {
		writeSSEError(writer, ErrForbidden)
		return
	}
	_ = level

	// Find canonical rows via game_vocabs
	var vocabs []models.ContentVocab
	if err := facades.Orm().Query().Model(&models.ContentVocab{}).
		Select("DISTINCT content_vocabs.*").
		Join("JOIN game_vocabs gv ON gv.content_vocab_id = content_vocabs.id AND gv.deleted_at IS NULL").
		Where("gv.game_level_id", gameLevelID).
		Where("content_vocabs.uk_phonetic IS NULL").
		Get(&vocabs); err != nil {
		writeSSEError(writer, fmt.Errorf("failed to load vocabs needing enrichment: %w", err))
		return
	}
	if len(vocabs) == 0 {
		_ = writer.Write(SSEProgressEvent{Done: 0, Total: 0, Processed: 0, Failed: 0, Complete: true})
		writer.Close()
		return
	}

	// Bean cost = total word count across vocabs
	totalCost := 0
	for _, v := range vocabs {
		totalCost += helpers.CountWords(v.Content)
	}
	if totalCost == 0 {
		writeSSEError(writer, ErrEmptyContent)
		return
	}
	if err := ConsumeBeans(userID, totalCost, consts.BeanSlugAIVocabGenItemsConsume, consts.BeanReasonAIVocabGenItemsConsume); err != nil {
		writeSSEError(writer, err)
		return
	}

	var failedWords int64
	var processed int64
	var failed int64
	sem := make(chan struct{}, genItemsConcurrencyLimit)
	var wg sync.WaitGroup
	var done int64
	total := len(vocabs)

	for _, v := range vocabs {
		wg.Add(1)
		sem <- struct{}{}

		go func(vocab models.ContentVocab) {
			defer wg.Done()
			defer func() { <-sem }()

			ok := enrichContentVocab(userID, vocab)
			d := atomic.AddInt64(&done, 1)
			if ok {
				atomic.AddInt64(&processed, 1)
				_ = writer.Write(SSEProgressEvent{Done: int(d), Total: total, Status: "ok"})
			} else {
				atomic.AddInt64(&failed, 1)
				atomic.AddInt64(&failedWords, int64(helpers.CountWords(vocab.Content)))
				_ = writer.Write(SSEProgressEvent{Done: int(d), Total: total, Status: "failed"})
			}
		}(v)
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

func enrichContentVocab(userID string, v models.ContentVocab) bool {
	userMsg := "Word: " + v.Content
	result, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
		Messages: []helpers.DeepSeekMessage{
			{Role: "system", Content: vocabFieldsPrompt},
			{Role: "user", Content: userMsg},
		},
		Temperature: 0.1,
	})
	if err != nil {
		return false
	}

	var ai struct {
		UkPhonetic  string              `json:"ukPhonetic"`
		UsPhonetic  string              `json:"usPhonetic"`
		Definition  []map[string]string `json:"definition"`
		Explanation string              `json:"explanation"`
	}
	if err := json.Unmarshal([]byte(result), &ai); err != nil {
		return false
	}
	if err := ValidatePosEntries(ai.Definition); err != nil {
		return false
	}

	beforeSnapshot, _ := SnapshotVocab(&v)

	// Additive merge — never overwrite existing curated data.
	// Phonetic / explanation set only when currently null/empty.
	// Definition merges new POS keys (existing keys win on conflict).
	updates := map[string]any{"last_edited_by": userID}
	if v.UkPhonetic == nil || *v.UkPhonetic == "" {
		updates["uk_phonetic"] = ai.UkPhonetic
	}
	if v.UsPhonetic == nil || *v.UsPhonetic == "" {
		updates["us_phonetic"] = ai.UsPhonetic
	}
	if v.Explanation == nil || *v.Explanation == "" {
		updates["explanation"] = ai.Explanation
	}
	existingDef := ""
	if v.Definition != nil {
		existingDef = *v.Definition
	}
	mergedDef, err := MergeDefinition(existingDef, ai.Definition)
	if err != nil {
		return false
	}
	if mergedDef != existingDef {
		updates["definition"] = mergedDef
	}

	if len(updates) == 1 {
		// Only "last_edited_by" — nothing to merge in. Skip update + edit log.
		return true
	}

	if _, err := facades.Orm().Query().Model(&models.ContentVocab{}).
		Where("id", v.ID).Update(updates); err != nil {
		return false
	}
	var updated models.ContentVocab
	if err := facades.Orm().Query().Where("id", v.ID).First(&updated); err == nil {
		afterSnapshot, _ := SnapshotVocab(&updated)
		_ = WriteVocabEdit(nil, v.ID, userID, "complement", beforeSnapshot, afterSnapshot)
	}
	return true
}

var vocabFieldsPrompt = `You are an English dictionary writer. Given a single English word or phrase, produce JSON with phonetic, definition (POS entries), and a short explanation.

OUTPUT FORMAT:
A JSON object with these keys:
- ukPhonetic: IPA pronunciation in UK style, e.g. "/fæst/"
- usPhonetic: IPA pronunciation in US style, e.g. "/fæst/"
- definition: array of single-key objects mapping POS to Chinese gloss; allowed POS keys are n, v, adj, adv, prep, conj, pron, art, num, int, aux, det
- explanation: short Chinese explanation/example (1-2 sentences)

Output ONLY the JSON. No markdown, no extra text.

Example for "fast":
{"ukPhonetic":"/fɑːst/","usPhonetic":"/fæst/","definition":[{"adj":"快的"},{"adv":"快速地"},{"v":"斋戒"}],"explanation":"形容速度快或动作迅速；动词义为禁食。"}`

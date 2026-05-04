package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
		UkPhonetic:  in.UkPhonetic,
		UsPhonetic:  in.UsPhonetic,
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
		"uk_phonetic":  in.UkPhonetic,
		"us_phonetic":  in.UsPhonetic,
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

// GenerateVocabsFromKeywords: keywords → JSON array of 20 vocab entries.
// Bean cost = aiGenerateCost (5). Refunds on AI failure.
func GenerateVocabsFromKeywords(userID string, keywords []string) (string, error) {
	if err := requireVip(userID); err != nil {
		return "", err
	}
	if err := ConsumeBeans(userID, aiGenerateCost, consts.BeanSlugAIVocabGenerateConsume, consts.BeanReasonAIVocabGenerateConsume); err != nil {
		return "", err
	}

	userMsg := "Keywords: " + strings.Join(keywords, ", ")

	result, err := helpers.CallDeepSeek(helpers.DeepSeekRequest{
		Messages: []helpers.DeepSeekMessage{
			{Role: "system", Content: vocabsFromKeywordsPrompt},
			{Role: "user", Content: userMsg},
		},
		Temperature: 0.7,
	})
	if err != nil {
		_ = RefundBeans(userID, aiGenerateCost, consts.BeanSlugAIVocabGenerateRefund, consts.BeanReasonAIVocabGenerateRefund)
		return "", err
	}

	if rest, ok := strings.CutPrefix(result, "WARNING:"); ok {
		return "WARNING:" + strings.TrimSpace(rest), nil
	}

	return result, nil
}

var vocabsFromKeywordsPrompt = `You are a vocabulary generator. Given keywords, produce a JSON array of EXACTLY 20 English vocabulary entries related to those keywords.

OUTPUT FORMAT:
A JSON array. Each element has:
- content: English word or short phrase
- ukPhonetic: IPA pronunciation in UK style, e.g. "/fæst/"
- usPhonetic: IPA pronunciation in US style, e.g. "/fæst/"
- definition: array of single-key objects mapping POS to Chinese gloss; allowed POS keys are n, v, adj, adv, prep, conj, pron, art, num, int, aux, det
- explanation: short Chinese explanation/example (1-2 sentences)

Output ONLY the JSON array. No markdown code fences, no explanation, no extra text.

CONTENT MODERATION: if keywords contain any insulting, violent, sexually explicit, or otherwise inappropriate material, respond ONLY with: WARNING:包含不适当内容，请修改后重试

Example for keyword "speed":
[
  {"content":"fast","ukPhonetic":"/fɑːst/","usPhonetic":"/fæst/","definition":[{"adj":"快的"},{"v":"斋戒"}],"explanation":"形容速度快或动作迅速；动词义为禁食。"},
  ...
]`

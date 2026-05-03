package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
)

// Edit gating: rows are editable freely up to this age if not yet verified.
const unverifiedEditWindow = 24 * time.Hour

// Admin username (per CLAUDE.md). Wiki replace + verify operations require this.
const adminUsername = "rainson"

// Vocab content validator: words/phrases only — letters, digits, spaces,
// apostrophe, and hyphen. Rejects punctuation otherwise.
var vocabContentRe = regexp.MustCompile(`^[A-Za-z0-9' \-]+$`)

var (
	ErrVocabContentEmpty   = errors.New("vocab content is empty")
	ErrVocabContentInvalid = errors.New("vocab content contains disallowed characters")
	ErrVocabNotFound       = errors.New("content vocab not found")
	ErrVocabNotEditable    = errors.New("content vocab is not editable by this user")
	ErrVocabAdminOnly      = errors.New("operation requires admin")
	ErrInvalidPosKey       = errors.New("definition contains invalid POS key")
)

// NormalizeVocabContent trims and lowercases the content for use as content_key.
// Multiple internal whitespace runs are collapsed to a single space.
func NormalizeVocabContent(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	if s == "" {
		return ""
	}
	// Collapse internal whitespace
	parts := strings.Fields(s)
	return strings.Join(parts, " ")
}

// ValidateVocabContent ensures a vocab string is non-empty and contains only
// allowed characters (letters, digits, spaces, apostrophe, hyphen).
func ValidateVocabContent(s string) error {
	t := strings.TrimSpace(s)
	if t == "" {
		return ErrVocabContentEmpty
	}
	if !vocabContentRe.MatchString(t) {
		return ErrVocabContentInvalid
	}
	return nil
}

// ValidateDefinition ensures every entry is a single-key object with a known POS.
// definition is the JSON text from a request body.
func ValidateDefinition(definitionJSON string) error {
	if definitionJSON == "" {
		return nil
	}
	var entries []map[string]string
	if err := json.Unmarshal([]byte(definitionJSON), &entries); err != nil {
		return fmt.Errorf("invalid definition JSON: %w", err)
	}
	for _, entry := range entries {
		if len(entry) != 1 {
			return fmt.Errorf("each definition entry must be a single-key object")
		}
		for k := range entry {
			if !consts.IsValidPos(k) {
				return ErrInvalidPosKey
			}
		}
	}
	return nil
}

// MergeDefinition merges newEntries into existingJSON additively: only POS keys
// not already present in existing get appended. Returns the merged JSON text.
func MergeDefinition(existingJSON string, newEntries []map[string]string) (string, error) {
	var existing []map[string]string
	if existingJSON != "" {
		if err := json.Unmarshal([]byte(existingJSON), &existing); err != nil {
			return "", fmt.Errorf("invalid existing definition JSON: %w", err)
		}
	}
	seen := make(map[string]struct{})
	for _, entry := range existing {
		for k := range entry {
			seen[k] = struct{}{}
		}
	}
	merged := append([]map[string]string{}, existing...)
	for _, entry := range newEntries {
		for k := range entry {
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			merged = append(merged, entry)
		}
	}
	out, err := json.Marshal(merged)
	if err != nil {
		return "", fmt.Errorf("failed to marshal merged definition: %w", err)
	}
	return string(out), nil
}

// IsAdmin returns true if the user has the admin username.
func IsAdmin(userID string) bool {
	var user models.User
	if err := orm0().Where("id", userID).Select("username").First(&user); err != nil {
		return false
	}
	return user.Username == adminUsername
}

// CanReplaceVocab returns true if userID can perform a destructive replace on vocab.
func CanReplaceVocab(userID string, vocab *models.ContentVocab) bool {
	if vocab.CreatedBy != nil && *vocab.CreatedBy == userID {
		return true
	}
	if IsAdmin(userID) {
		return true
	}
	if !vocab.IsVerified && vocab.CreatedAt != nil && time.Since(vocab.CreatedAt.StdTime()) < unverifiedEditWindow {
		return true
	}
	return false
}

// SnapshotVocab serializes a ContentVocab into a JSON map for audit log.
func SnapshotVocab(v *models.ContentVocab) (string, error) {
	b, err := json.Marshal(map[string]any{
		"id":             v.ID,
		"content":        v.Content,
		"content_key":    v.ContentKey,
		"uk_phonetic":    v.UkPhonetic,
		"us_phonetic":    v.UsPhonetic,
		"uk_audio_url":   v.UkAudioURL,
		"us_audio_url":   v.UsAudioURL,
		"definition":     v.Definition,
		"explanation":    v.Explanation,
		"is_verified":    v.IsVerified,
		"created_by":     v.CreatedBy,
		"last_edited_by": v.LastEditedBy,
	})
	if err != nil {
		return "", fmt.Errorf("snapshot marshal failed: %w", err)
	}
	return string(b), nil
}

// WriteVocabEdit appends an audit row. tx is optional — pass nil to use facades.Orm().
func WriteVocabEdit(tx orm.Query, vocabID, editorUserID, editType, beforeJSON, afterJSON string) error {
	q := tx
	if q == nil {
		q = orm0()
	}
	edit := models.ContentVocabEdit{
		ID:             uuid.Must(uuid.NewV7()).String(),
		ContentVocabID: vocabID,
		EditType:       editType,
	}
	if editorUserID != "" {
		v := editorUserID
		edit.EditorUserID = &v
	}
	if beforeJSON != "" {
		v := beforeJSON
		edit.Before = &v
	}
	if afterJSON != "" {
		v := afterJSON
		edit.After = &v
	}
	if err := q.Create(&edit); err != nil {
		return fmt.Errorf("failed to create content_vocab_edit: %w", err)
	}
	return nil
}

// orm0 returns the default ORM query handle.
func orm0() orm.Query {
	return facades.Orm().Query()
}

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"dx-api/app/consts"
)

// Vocab content validator: words/phrases only — letters, digits, spaces,
// apostrophe, and hyphen. Rejects punctuation otherwise.
var vocabContentRe = regexp.MustCompile(`^[A-Za-z0-9' \-]+$`)

var (
	ErrVocabContentEmpty   = errors.New("vocab content is empty")
	ErrVocabContentInvalid = errors.New("vocab content contains disallowed characters")
	ErrVocabNotFound       = errors.New("content vocab not found")
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
	return ValidatePosEntries(entries)
}

// ValidatePosEntries enforces single-key objects with known POS keys on an
// already-parsed definition slice.
func ValidatePosEntries(entries []map[string]string) error {
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

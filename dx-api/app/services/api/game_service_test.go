package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Pins the expected two-argument signature at compile time so downstream
// callers (controllers) cannot drift from it silently.
func TestGetGameDetailFunctionExists(t *testing.T) {
	assert.NotNil(t, GetGameDetail)
	var _ func(string, string) (*GameDetailData, error) = GetGameDetail
}

// Pins the expected zero-arg signature at compile time so downstream
// callers (controllers) cannot drift from it silently.
func TestGetSearchSuggestionsFunctionExists(t *testing.T) {
	assert.NotNil(t, GetSearchSuggestions)
	var _ func() ([]string, error) = GetSearchSuggestions
}

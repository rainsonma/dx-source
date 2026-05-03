package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

// ComplementVocabRequest — POST /api/content-vocabs/{id}/complement
type ComplementVocabRequest struct {
	Definition  []map[string]string `form:"definition" json:"definition"`
	UkPhonetic  *string             `form:"ukPhonetic" json:"ukPhonetic"`
	UsPhonetic  *string             `form:"usPhonetic" json:"usPhonetic"`
	UkAudioURL  *string             `form:"ukAudioUrl" json:"ukAudioUrl"`
	UsAudioURL  *string             `form:"usAudioUrl" json:"usAudioUrl"`
	Explanation *string             `form:"explanation" json:"explanation"`
}

func (r *ComplementVocabRequest) Authorize(http.Context) error { return nil }
func (r *ComplementVocabRequest) Rules(http.Context) map[string]string {
	return map[string]string{}
}
func (r *ComplementVocabRequest) Filters(http.Context) map[string]string    { return nil }
func (r *ComplementVocabRequest) Messages(http.Context) map[string]string   { return nil }
func (r *ComplementVocabRequest) Attributes(http.Context) map[string]string { return nil }
func (r *ComplementVocabRequest) PrepareForValidation(_ http.Context, _ validation.Data) error {
	return nil
}

// ReplaceVocabRequest — PUT /api/content-vocabs/{id}
type ReplaceVocabRequest struct {
	Content     string              `form:"content" json:"content"`
	Definition  []map[string]string `form:"definition" json:"definition"`
	UkPhonetic  *string             `form:"ukPhonetic" json:"ukPhonetic"`
	UsPhonetic  *string             `form:"usPhonetic" json:"usPhonetic"`
	UkAudioURL  *string             `form:"ukAudioUrl" json:"ukAudioUrl"`
	UsAudioURL  *string             `form:"usAudioUrl" json:"usAudioUrl"`
	Explanation *string             `form:"explanation" json:"explanation"`
}

func (r *ReplaceVocabRequest) Authorize(http.Context) error { return nil }
func (r *ReplaceVocabRequest) Rules(http.Context) map[string]string {
	return map[string]string{
		"content": "required|min_len:1|max_len:200",
	}
}
func (r *ReplaceVocabRequest) Filters(http.Context) map[string]string    { return nil }
func (r *ReplaceVocabRequest) Messages(http.Context) map[string]string   { return nil }
func (r *ReplaceVocabRequest) Attributes(http.Context) map[string]string { return nil }
func (r *ReplaceVocabRequest) PrepareForValidation(_ http.Context, _ validation.Data) error {
	return nil
}

// VerifyVocabRequest — POST /api/content-vocabs/{id}/verify
type VerifyVocabRequest struct {
	Verified bool `form:"verified" json:"verified"`
}

func (r *VerifyVocabRequest) Authorize(http.Context) error { return nil }
func (r *VerifyVocabRequest) Rules(http.Context) map[string]string {
	return map[string]string{}
}
func (r *VerifyVocabRequest) Filters(http.Context) map[string]string    { return nil }
func (r *VerifyVocabRequest) Messages(http.Context) map[string]string   { return nil }
func (r *VerifyVocabRequest) Attributes(http.Context) map[string]string { return nil }
func (r *VerifyVocabRequest) PrepareForValidation(_ http.Context, _ validation.Data) error {
	return nil
}

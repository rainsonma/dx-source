package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

// UpdateVocabRequest — PUT /api/content-vocabs/{id}
type UpdateVocabRequest struct {
	Content     string              `form:"content" json:"content"`
	Definition  []map[string]string `form:"definition" json:"definition"`
	UkPhonetic  *string             `form:"ukPhonetic" json:"ukPhonetic"`
	UsPhonetic  *string             `form:"usPhonetic" json:"usPhonetic"`
	UkAudioURL  *string             `form:"ukAudioUrl" json:"ukAudioUrl"`
	UsAudioURL  *string             `form:"usAudioUrl" json:"usAudioUrl"`
	Explanation *string             `form:"explanation" json:"explanation"`
}

func (r *UpdateVocabRequest) Authorize(http.Context) error { return nil }
func (r *UpdateVocabRequest) Rules(http.Context) map[string]string {
	return map[string]string{
		"content": "required|min_len:1|max_len:200",
	}
}
func (r *UpdateVocabRequest) Filters(http.Context) map[string]string    { return nil }
func (r *UpdateVocabRequest) Messages(http.Context) map[string]string   { return nil }
func (r *UpdateVocabRequest) Attributes(http.Context) map[string]string { return nil }
func (r *UpdateVocabRequest) PrepareForValidation(_ http.Context, _ validation.Data) error {
	return nil
}

// CreateVocabRequest — POST /api/content-vocabs
type CreateVocabRequest struct {
	Content     string              `form:"content" json:"content"`
	Definition  []map[string]string `form:"definition" json:"definition"`
	UkPhonetic  *string             `form:"ukPhonetic" json:"ukPhonetic"`
	UsPhonetic  *string             `form:"usPhonetic" json:"usPhonetic"`
	UkAudioURL  *string             `form:"ukAudioUrl" json:"ukAudioUrl"`
	UsAudioURL  *string             `form:"usAudioUrl" json:"usAudioUrl"`
	Explanation *string             `form:"explanation" json:"explanation"`
}

func (r *CreateVocabRequest) Authorize(http.Context) error { return nil }
func (r *CreateVocabRequest) Rules(http.Context) map[string]string {
	return map[string]string{
		"content": "required|min_len:1|max_len:200",
	}
}
func (r *CreateVocabRequest) Filters(http.Context) map[string]string    { return nil }
func (r *CreateVocabRequest) Messages(http.Context) map[string]string   { return nil }
func (r *CreateVocabRequest) Attributes(http.Context) map[string]string { return nil }
func (r *CreateVocabRequest) PrepareForValidation(_ http.Context, _ validation.Data) error {
	return nil
}

// VocabInputRequest mirrors VocabInput for batch creation.
type VocabInputRequest struct {
	Content     string              `form:"content" json:"content"`
	Definition  []map[string]string `form:"definition" json:"definition"`
	UkPhonetic  *string             `form:"ukPhonetic" json:"ukPhonetic"`
	UsPhonetic  *string             `form:"usPhonetic" json:"usPhonetic"`
	UkAudioURL  *string             `form:"ukAudioUrl" json:"ukAudioUrl"`
	UsAudioURL  *string             `form:"usAudioUrl" json:"usAudioUrl"`
	Explanation *string             `form:"explanation" json:"explanation"`
}

// CreateVocabsBatchRequest — POST /api/content-vocabs/batch
type CreateVocabsBatchRequest struct {
	Inputs []VocabInputRequest `form:"inputs" json:"inputs"`
}

func (r *CreateVocabsBatchRequest) Authorize(http.Context) error { return nil }
func (r *CreateVocabsBatchRequest) Rules(http.Context) map[string]string {
	return map[string]string{
		"inputs": "required",
	}
}
func (r *CreateVocabsBatchRequest) Filters(http.Context) map[string]string    { return nil }
func (r *CreateVocabsBatchRequest) Messages(http.Context) map[string]string   { return nil }
func (r *CreateVocabsBatchRequest) Attributes(http.Context) map[string]string { return nil }
func (r *CreateVocabsBatchRequest) PrepareForValidation(_ http.Context, _ validation.Data) error {
	return nil
}

// GenerateVocabsFromKeywordsRequest — POST /api/ai-custom/generate-vocabs-from-keywords
type GenerateVocabsFromKeywordsRequest struct {
	Keywords []string `form:"keywords" json:"keywords"`
}

func (r *GenerateVocabsFromKeywordsRequest) Authorize(http.Context) error { return nil }
func (r *GenerateVocabsFromKeywordsRequest) Rules(http.Context) map[string]string {
	return map[string]string{
		"keywords": "required",
	}
}
func (r *GenerateVocabsFromKeywordsRequest) Filters(http.Context) map[string]string    { return nil }
func (r *GenerateVocabsFromKeywordsRequest) Messages(http.Context) map[string]string   { return nil }
func (r *GenerateVocabsFromKeywordsRequest) Attributes(http.Context) map[string]string { return nil }
func (r *GenerateVocabsFromKeywordsRequest) PrepareForValidation(_ http.Context, _ validation.Data) error {
	return nil
}

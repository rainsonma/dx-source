package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

// AddGameVocabsRequest — POST /api/course-games/{id}/levels/{levelId}/game-vocabs
type AddGameVocabsRequest struct {
	Entries []string `form:"entries" json:"entries"`
}

func (r *AddGameVocabsRequest) Authorize(http.Context) error { return nil }
func (r *AddGameVocabsRequest) Rules(http.Context) map[string]string {
	return map[string]string{
		"entries": "required",
	}
}
func (r *AddGameVocabsRequest) Filters(http.Context) map[string]string    { return nil }
func (r *AddGameVocabsRequest) Messages(http.Context) map[string]string   { return nil }
func (r *AddGameVocabsRequest) Attributes(http.Context) map[string]string { return nil }
func (r *AddGameVocabsRequest) PrepareForValidation(_ http.Context, _ validation.Data) error {
	return nil
}

// ReorderGameVocabRequest — PUT /api/course-games/{id}/game-vocabs/{gvId}/reorder
type ReorderGameVocabRequest struct {
	NewOrder float64 `form:"newOrder" json:"newOrder"`
}

func (r *ReorderGameVocabRequest) Authorize(http.Context) error { return nil }
func (r *ReorderGameVocabRequest) Rules(http.Context) map[string]string {
	return map[string]string{
		"newOrder": "required",
	}
}
func (r *ReorderGameVocabRequest) Filters(http.Context) map[string]string    { return nil }
func (r *ReorderGameVocabRequest) Messages(http.Context) map[string]string   { return nil }
func (r *ReorderGameVocabRequest) Attributes(http.Context) map[string]string { return nil }
func (r *ReorderGameVocabRequest) PrepareForValidation(_ http.Context, _ validation.Data) error {
	return nil
}

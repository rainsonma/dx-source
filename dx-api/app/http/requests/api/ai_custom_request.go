package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

// AiCustomLevelRequest — shared request for AI ops that take a game level ID.
type AiCustomLevelRequest struct {
	GameLevelID string `form:"gameLevelId" json:"gameLevelId"`
}

func (r *AiCustomLevelRequest) Authorize(http.Context) error { return nil }
func (r *AiCustomLevelRequest) Rules(http.Context) map[string]string {
	return map[string]string{
		"gameLevelId": "required",
	}
}
func (r *AiCustomLevelRequest) Filters(http.Context) map[string]string    { return nil }
func (r *AiCustomLevelRequest) Messages(http.Context) map[string]string   { return nil }
func (r *AiCustomLevelRequest) Attributes(http.Context) map[string]string { return nil }
func (r *AiCustomLevelRequest) PrepareForValidation(_ http.Context, _ validation.Data) error {
	return nil
}

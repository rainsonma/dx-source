package api

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"

	"dx-api/app/consts"
	"dx-api/app/helpers"
)

// ListGamesRequest holds query parameters for listing published games.
type ListGamesRequest struct {
	Cursor      string   `form:"cursor" json:"cursor"`
	Limit       int      `form:"limit" json:"limit"`
	CategoryIDs []string `form:"categoryIds" json:"categoryIds"`
	PressID     string   `form:"pressId" json:"pressId"`
	Mode        string   `form:"mode" json:"mode"`
}

func (r *ListGamesRequest) Authorize(ctx http.Context) error { return nil }
func (r *ListGamesRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"pressId": "uuid",
		"mode":    helpers.InEnum("mode"),
	}
}
func (r *ListGamesRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"cursor":  "trim",
		"pressId": "trim",
	}
}
func (r *ListGamesRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"limit.min":    "每页数量不能小于1",
		"limit.max":    "每页数量不能超过50",
		"pressId.uuid": "无效的出版社ID",
		"mode.in":      "无效的游戏模式",
	}
}

// SearchGamesRequest holds query parameters for searching games.
type SearchGamesRequest struct {
	Query string `form:"q" json:"q"`
	Limit int    `form:"limit" json:"limit"`
}

func (r *SearchGamesRequest) Authorize(ctx http.Context) error { return nil }
func (r *SearchGamesRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"q": "required|min_len:1|max_len:50",
	}
}
func (r *SearchGamesRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"q": "trim",
	}
}
func (r *SearchGamesRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"q.required": "请输入搜索关键词",
		"q.max_len":  "搜索关键词不能超过50个字符",
		"limit.min":  "每页数量不能小于1",
		"limit.max":  "每页数量不能超过50",
	}
}

// LevelContentRequest holds query parameters for fetching level content.
type LevelContentRequest struct {
	Degree string `form:"degree" json:"degree"`
}

func (r *LevelContentRequest) Authorize(ctx http.Context) error { return nil }
func (r *LevelContentRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"degree": helpers.InEnum("degree"),
	}
}
func (r *LevelContentRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"degree.in": "无效的难度级别",
	}
}
func (r *LevelContentRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	degree, _ := data.Get("degree")
	if degree == nil || degree == "" {
		data.Set("degree", consts.GameDegreeBeginner)
	}
	return nil
}

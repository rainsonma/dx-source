package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type AiCustomVocabController struct{}

func NewAiCustomVocabController() *AiCustomVocabController {
	return &AiCustomVocabController{}
}

// GenerateVocab generates vocab pairs from keywords using AI.
func (c *AiCustomVocabController) GenerateVocab(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req struct {
		Difficulty string   `json:"difficulty"`
		Keywords   []string `json:"keywords"`
		GameMode   string   `json:"gameMode"`
	}
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的请求")
	}

	if req.Difficulty == "" {
		req.Difficulty = "a1-a2"
	}
	if len(req.Keywords) == 0 || len(req.Keywords) > 5 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "请提供1-5个关键词")
	}
	if !consts.IsVocabMode(req.GameMode) {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的游戏模式")
	}

	result, err := services.GenerateVocab(userID, req.Difficulty, req.Keywords, req.GameMode)
	if err != nil {
		return mapVocabServiceError(ctx, err, "AI 词汇服务")
	}

	if result.Warning != "" {
		return helpers.Success(ctx, map[string]any{"warning": result.Warning})
	}

	return helpers.Success(ctx, map[string]any{
		"generated": result.Generated,
	})
}

// FormatVocab formats raw vocab text using AI.
func (c *AiCustomVocabController) FormatVocab(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的请求")
	}

	if req.Content == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "请输入内容")
	}

	result, err := services.FormatVocab(userID, req.Content)
	if err != nil {
		return mapVocabServiceError(ctx, err, "格式化服务")
	}

	if result.Warning != "" {
		return helpers.Success(ctx, map[string]any{"warning": result.Warning})
	}

	return helpers.Success(ctx, map[string]any{
		"formatted": result.Formatted,
	})
}

// BreakMetadata breaks vocab metas into content items via SSE.
func (c *AiCustomVocabController) BreakMetadata(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req struct {
		GameLevelID string `json:"gameLevelId"`
	}
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的请求")
	}

	if req.GameLevelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "gameLevelId is required")
	}

	w := ctx.Response().Writer()
	writer := helpers.NewSSEWriter(w)

	services.BreakVocabMetadata(userID, req.GameLevelID, writer)

	return nil
}

// GenerateContentItems generates word-level phonetics and translations for vocab items via SSE.
func (c *AiCustomVocabController) GenerateContentItems(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req struct {
		GameLevelID string `json:"gameLevelId"`
	}
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的请求")
	}

	if req.GameLevelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "gameLevelId is required")
	}

	w := ctx.Response().Writer()
	writer := helpers.NewSSEWriter(w)

	services.GenerateVocabContentItems(userID, req.GameLevelID, writer)

	return nil
}

// mapVocabServiceError maps service errors to HTTP responses.
func mapVocabServiceError(ctx contractshttp.Context, err error, serviceLabel string) contractshttp.Response {
	switch {
	case errors.Is(err, services.ErrVipRequired):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
	case errors.Is(err, services.ErrInsufficientBeans):
		return helpers.Error(ctx, http.StatusPaymentRequired, consts.CodeInsufficientBeans, "能量豆不足")
	case errors.Is(err, services.ErrEmptyContent):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "内容为空")
	case errors.Is(err, helpers.ErrDeepSeekEmpty),
		errors.Is(err, helpers.ErrDeepSeekAuth),
		errors.Is(err, helpers.ErrDeepSeekQuota),
		errors.Is(err, helpers.ErrDeepSeekRateLimit),
		errors.Is(err, helpers.ErrDeepSeekNotConfigured),
		errors.Is(err, helpers.ErrDeepSeekUnavail):
		msg, status := helpers.MapDeepSeekError(err, serviceLabel)
		return helpers.Error(ctx, status, consts.CodeAIServiceError, msg)
	default:
		msg, status := helpers.MapDeepSeekError(err, serviceLabel)
		return helpers.Error(ctx, status, consts.CodeAIServiceError, msg)
	}
}

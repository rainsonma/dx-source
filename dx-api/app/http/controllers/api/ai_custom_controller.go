package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	apiReq "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type AiCustomController struct{}

func NewAiCustomController() *AiCustomController {
	return &AiCustomController{}
}

// GenerateMetadata generates a story from keywords using AI.
func (c *AiCustomController) GenerateMetadata(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req struct {
		Difficulty string   `json:"difficulty"`
		Keywords   []string `json:"keywords"`
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

	result, err := services.GenerateMetadata(userID, req.Difficulty, req.Keywords)
	if err != nil {
		return mapAIServiceError(ctx, err, "AI 服务")
	}

	if result.Warning != "" {
		return helpers.Success(ctx, map[string]any{"warning": result.Warning})
	}

	return helpers.Success(ctx, map[string]any{
		"generated":  result.Generated,
		"sourceType": result.SourceType,
	})
}

// FormatMetadata formats raw text into structured learning content using AI.
func (c *AiCustomController) FormatMetadata(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req struct {
		Content    string `json:"content"`
		FormatType string `json:"formatType"`
	}
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的请求")
	}

	if req.Content == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "请输入内容")
	}
	if req.FormatType != "sentence" && req.FormatType != "vocab" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "格式类型必须是句子或词汇")
	}

	result, err := services.FormatMetadata(userID, req.Content, req.FormatType)
	if err != nil {
		return mapAIServiceError(ctx, err, "格式化服务")
	}

	if result.Warning != "" {
		return helpers.Success(ctx, map[string]any{"warning": result.Warning})
	}

	return helpers.Success(ctx, map[string]any{
		"formatted":   result.Formatted,
		"sourceTypes": result.SourceTypes,
	})
}

// BreakMetadata breaks content metas into learning units via SSE.
func (c *AiCustomController) BreakMetadata(ctx contractshttp.Context) contractshttp.Response {
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
	writer := helpers.NewNDJSONWriter(w)

	services.BreakMetadata(userID, req.GameLevelID, writer)

	return nil
}

// GenerateContentItems generates word-level phonetics and translations via SSE.
func (c *AiCustomController) GenerateContentItems(ctx contractshttp.Context) contractshttp.Response {
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
	writer := helpers.NewNDJSONWriter(w)

	services.GenerateContentItems(userID, req.GameLevelID, writer)

	return nil
}

// GenerateVocab generates vocab pairs from keywords using AI.
func (c *AiCustomController) GenerateVocab(ctx contractshttp.Context) contractshttp.Response {
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
		return mapAIServiceError(ctx, err, "AI 词汇服务")
	}

	if result.Warning != "" {
		return helpers.Success(ctx, map[string]any{"warning": result.Warning})
	}

	return helpers.Success(ctx, map[string]any{
		"generated": result.Generated,
	})
}

// FormatVocab formats raw vocab text using AI.
func (c *AiCustomController) FormatVocab(ctx contractshttp.Context) contractshttp.Response {
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
		return mapAIServiceError(ctx, err, "格式化服务")
	}

	if result.Warning != "" {
		return helpers.Success(ctx, map[string]any{"warning": result.Warning})
	}

	return helpers.Success(ctx, map[string]any{
		"formatted": result.Formatted,
	})
}

// GenerateVocabWords generates 15-25 English word strings from keywords (Phase 1).
func (c *AiCustomController) GenerateVocabWords(ctx contractshttp.Context) contractshttp.Response {
	userID, authErr := facades.Auth(ctx).Guard("user").ID()
	if authErr != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}
	var req apiReq.GenerateVocabWordsRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}
	if len(req.Keywords) == 0 || len(req.Keywords) > 10 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "请提供1-10个关键词")
	}
	for _, kw := range req.Keywords {
		if len(kw) > 30 {
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "单个关键词不能超过30个字符")
		}
	}
	if req.Difficulty == "" {
		req.Difficulty = "a1-a2"
	}

	words, err := services.GenerateVocabWords(userID, req.Keywords, req.Difficulty)
	if err != nil {
		if errors.Is(err, services.ErrModerationWarning) {
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "包含不适当内容，请修改后重试")
		}
		return mapAIServiceError(ctx, err, "AI 词汇生成")
	}
	return helpers.Success(ctx, map[string]any{"words": words})
}

// mapAIServiceError maps service errors to HTTP responses.
func mapAIServiceError(ctx contractshttp.Context, err error, serviceLabel string) contractshttp.Response {
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

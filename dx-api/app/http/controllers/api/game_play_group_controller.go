package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type GamePlayGroupController struct{}

func NewGamePlayGroupController() *GamePlayGroupController {
	return &GamePlayGroupController{}
}

// Start starts or resumes a group game session.
func (c *GamePlayGroupController) Start(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.GroupPlayStartSessionRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.GroupPlayStartSession(userID, req.GameID, req.GameLevelID, req.Degree, req.Pattern, req.GameGroupID)
	if err != nil {
		if errors.Is(err, services.ErrNoGameLevels) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "游戏没有关卡")
		}
		if errors.Is(err, services.ErrGroupNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeGroupNotFound, "group not found")
		}
		if errors.Is(err, services.ErrNotInSubgroup) {
			return helpers.Error(ctx, http.StatusForbidden, consts.CodeForbidden, "member not in any subgroup")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to start session")
	}

	return helpers.Success(ctx, result)
}

// CompleteLevel completes a level within a group game session.
func (c *GamePlayGroupController) CompleteLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")
	gameLevelID := ctx.Request().Route("levelId")

	var req requests.CompleteLevelRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.GroupPlayCompleteLevel(userID, sessionID, gameLevelID, req.Score, req.MaxCombo, req.TotalItems)
	if err != nil {
		if errors.Is(err, services.ErrSessionLevelNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "关卡会话不存在")
		}
		if errors.Is(err, services.ErrSessionNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeSessionNotFound, "会话不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to complete level")
	}

	return helpers.Success(ctx, result)
}

// RecordAnswer records a single answer in a group game session.
func (c *GamePlayGroupController) RecordAnswer(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.RecordAnswerRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	err = services.GroupPlayRecordAnswer(userID, services.RecordAnswerInput{
		GameSessionID:      req.GameSessionId,
		GameLevelID:        req.GameLevelID,
		ContentItemID:      req.ContentItemID,
		ContentVocabID:     req.ContentVocabID,
		IsCorrect:          req.IsCorrect,
		UserAnswer:         req.UserAnswer,
		SourceAnswer:       req.SourceAnswer,
		BaseScore:          req.BaseScore,
		ComboScore:         req.ComboScore,
		Score:              req.Score,
		MaxCombo:           req.MaxCombo,
		PlayTime:           req.PlayTime,
		NextContentItemID:  req.NextContentItemID,
		NextContentVocabID: req.NextContentVocabID,
		Duration:           req.Duration,
	})
	if err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "操作过于频繁，请稍后再试")
		}
		if errors.Is(err, services.ErrSessionLevelNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "关卡会话不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to record answer")
	}

	return helpers.Success(ctx, nil)
}

// SyncPlayTime syncs playtime for a group game session.
func (c *GamePlayGroupController) SyncPlayTime(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.SyncPlayTimeRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.GroupPlaySyncPlayTime(userID, sessionID, req.PlayTime); err != nil {
		if errors.Is(err, services.ErrInvalidPlayTime) {
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "游玩时长必须在0到86400秒之间")
		}
		return mapGroupPlayError(ctx, err)
	}

	return helpers.Success(ctx, map[string]bool{"ok": true})
}

// Restore returns accumulated stats for restoring client state in a group game session.
func (c *GamePlayGroupController) Restore(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	result, err := services.GroupPlayRestoreSessionData(userID, sessionID)
	if err != nil {
		return mapGroupPlayError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

// UpdateContentItem updates the group game session's current content item.
func (c *GamePlayGroupController) UpdateContentItem(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.UpdateContentItemRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.GroupPlayUpdateContentItem(userID, sessionID, req.ContentItemID, req.ContentVocabID); err != nil {
		return mapGroupPlayError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// mapGroupPlayError maps common group play errors to HTTP responses.
func mapGroupPlayError(ctx contractshttp.Context, err error) contractshttp.Response {
	if errors.Is(err, services.ErrSessionNotFound) {
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeSessionNotFound, "会话不存在")
	}
	if errors.Is(err, services.ErrForbidden) {
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeForbidden, "forbidden")
	}
	if errors.Is(err, services.ErrRateLimited) {
		return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "操作过于频繁，请稍后再试")
	}
	if errors.Is(err, services.ErrVipRequired) {
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
	}
	return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "internal server error")
}

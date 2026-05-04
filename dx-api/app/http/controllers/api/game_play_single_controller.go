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

type GamePlaySingleController struct{}

func NewGamePlaySingleController() *GamePlaySingleController {
	return &GamePlaySingleController{}
}

// Start starts or resumes a game session.
func (c *GamePlaySingleController) Start(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.StartSessionRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.StartSession(userID, req.GameID, req.GameLevelID, req.Degree, req.Pattern)
	if err != nil {
		if errors.Is(err, services.ErrNoGameLevels) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "游戏没有关卡")
		}
		if errors.Is(err, services.ErrVipRequired) {
			return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to start session")
	}

	return helpers.Success(ctx, result)
}

// End ends a game session and updates stats.
func (c *GamePlaySingleController) End(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.EndSessionRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	err = services.EndSession(userID, sessionID, services.EndSessionInput{
		Score:        req.Score,
		Exp:          req.Exp,
		MaxCombo:     req.MaxCombo,
		CorrectCount: req.CorrectCount,
		WrongCount:   req.WrongCount,
		SkipCount:    req.SkipCount,
	})
	if err != nil {
		return mapSessionError(ctx, err)
	}

	return helpers.Success(ctx, map[string]bool{"completed": true})
}

// ForceComplete force-completes a session.
func (c *GamePlaySingleController) ForceComplete(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	if err := services.ForceCompleteSession(userID, sessionID); err != nil {
		return mapSessionError(ctx, err)
	}

	return helpers.Success(ctx, map[string]bool{"completed": true})
}

// CompleteLevel completes a level within a session.
func (c *GamePlaySingleController) CompleteLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.CompleteLevelRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.CompleteLevel(userID, sessionID, req.Score, req.MaxCombo, req.TotalItems)
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

// RestartLevel restarts a level within a session.
func (c *GamePlaySingleController) RestartLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	if err := services.RestartLevel(userID, sessionID); err != nil {
		return mapSessionError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// RecordAnswer records a single answer.
func (c *GamePlaySingleController) RecordAnswer(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.RecordAnswerRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	err = services.RecordAnswer(userID, services.RecordAnswerInput{
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

// RecordSkip records a skip.
func (c *GamePlaySingleController) RecordSkip(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.RecordSkipRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	err = services.RecordSkip(userID, services.RecordSkipInput{
		GameSessionID:      req.GameSessionId,
		GameLevelID:        req.GameLevelID,
		PlayTime:           req.PlayTime,
		NextContentItemID:  req.NextContentItemID,
		NextContentVocabID: req.NextContentVocabID,
	})
	if err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "操作过于频繁，请稍后再试")
		}
		if errors.Is(err, services.ErrSessionLevelNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "关卡会话不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to record skip")
	}

	return helpers.Success(ctx, nil)
}

// SyncPlayTime syncs playtime for a session.
func (c *GamePlaySingleController) SyncPlayTime(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.SyncPlayTimeRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.SyncPlayTime(userID, sessionID, req.PlayTime); err != nil {
		if errors.Is(err, services.ErrInvalidPlayTime) {
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "游玩时长必须在0到86400秒之间")
		}
		return mapSessionError(ctx, err)
	}

	return helpers.Success(ctx, map[string]bool{"ok": true})
}

// CheckActive checks for an active session by degree+pattern.
func (c *GamePlaySingleController) CheckActive(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CheckActiveSessionRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.CheckActiveSession(userID, req.GameLevelID, req.Degree, req.Pattern)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to check active session")
	}

	return helpers.Success(ctx, result)
}

// CheckAnyActive checks for any active session for a game.
func (c *GamePlaySingleController) CheckAnyActive(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Query("game_id", "")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game_id is required")
	}

	result, err := services.CheckAnyActiveSession(userID, gameID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to check active session")
	}

	return helpers.Success(ctx, result)
}

// Restore returns accumulated stats for restoring client state.
func (c *GamePlaySingleController) Restore(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	result, err := services.RestoreSessionData(userID, sessionID)
	if err != nil {
		return mapSessionError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

// UpdateContentItem updates the session's current content item.
func (c *GamePlaySingleController) UpdateContentItem(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.UpdateContentItemRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.UpdateCurrentContentItem(userID, sessionID, req.ContentItemID, req.ContentVocabID); err != nil {
		return mapSessionError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// mapSessionError maps common session errors to HTTP responses.
func mapSessionError(ctx contractshttp.Context, err error) contractshttp.Response {
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

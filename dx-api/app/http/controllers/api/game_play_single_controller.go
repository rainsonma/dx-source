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

	result, err := services.StartSession(userID, req.GameID, req.Degree, req.Pattern, req.LevelID)
	if err != nil {
		if errors.Is(err, services.ErrNoGameLevels) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "游戏没有关卡")
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
		GameID:             req.GameID,
		Score:              req.Score,
		Exp:                req.Exp,
		MaxCombo:           req.MaxCombo,
		CorrectCount:       req.CorrectCount,
		WrongCount:         req.WrongCount,
		SkipCount:          req.SkipCount,
		AllLevelsCompleted: req.AllLevelsCompleted,
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

// StartLevel starts a level within a session.
func (c *GamePlaySingleController) StartLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.StartLevelRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.StartLevel(userID, sessionID, req.GameLevelID, req.Degree, req.Pattern)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to start level")
	}

	return helpers.Success(ctx, result)
}

// CompleteLevel completes a level within a session.
func (c *GamePlaySingleController) CompleteLevel(ctx contractshttp.Context) contractshttp.Response {
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

	result, err := services.CompleteLevel(userID, sessionID, gameLevelID, req.Score, req.MaxCombo, req.TotalItems)
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

// AdvanceLevel advances to the next level.
func (c *GamePlaySingleController) AdvanceLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")
	gameLevelID := ctx.Request().Route("levelId")

	var req requests.AdvanceLevelRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	// Fallback to route param if body field is empty
	nextLevelID := req.NextLevelID
	if nextLevelID == "" {
		nextLevelID = gameLevelID
	}
	if nextLevelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "next_level_id is required")
	}

	if err := services.AdvanceLevel(userID, sessionID, nextLevelID); err != nil {
		return mapSessionError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// RestartLevel restarts a level within a session.
func (c *GamePlaySingleController) RestartLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")
	gameLevelID := ctx.Request().Route("levelId")

	if err := services.RestartLevel(userID, sessionID, gameLevelID); err != nil {
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

	sessionID := ctx.Request().Route("id")

	var req requests.RecordAnswerRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	err = services.RecordAnswer(userID, services.RecordAnswerInput{
		GameSessionTotalID: sessionID,
		GameSessionLevelID: req.GameSessionLevelID,
		GameLevelID:        req.GameLevelID,
		ContentItemID:      req.ContentItemID,
		IsCorrect:          req.IsCorrect,
		UserAnswer:         req.UserAnswer,
		SourceAnswer:       req.SourceAnswer,
		BaseScore:          req.BaseScore,
		ComboScore:         req.ComboScore,
		Score:              req.Score,
		MaxCombo:           req.MaxCombo,
		PlayTime:           req.PlayTime,
		NextContentItemID:  req.NextContentItemID,
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

	sessionID := ctx.Request().Route("id")

	var req requests.RecordSkipRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	err = services.RecordSkip(userID, services.RecordSkipInput{
		GameSessionTotalID: sessionID,
		GameLevelID:        req.GameLevelID,
		PlayTime:           req.PlayTime,
		NextContentItemID:  req.NextContentItemID,
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

	if err := services.SyncPlayTime(userID, sessionID, req.GameLevelID, req.PlayTime); err != nil {
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

	result, err := services.CheckActiveSession(userID, req.GameID, req.Degree, req.Pattern)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to check active session")
	}

	return helpers.Success(ctx, result)
}

// CheckActiveLevel checks for an active level session.
func (c *GamePlaySingleController) CheckActiveLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CheckActiveLevelSessionRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.CheckActiveLevelSession(userID, req.GameID, req.Degree, req.Pattern, req.GameLevelID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to check active level session")
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

	var req requests.RestoreSessionRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.RestoreSessionData(userID, sessionID, req.GameLevelID)
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

	if err := services.UpdateCurrentContentItem(userID, sessionID, req.ContentItemID); err != nil {
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
	return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "internal server error")
}

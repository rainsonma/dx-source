package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type GamePlayPkController struct{}

func NewGamePlayPkController() *GamePlayPkController {
	return &GamePlayPkController{}
}

// Start starts a new PK match.
func (c *GamePlayPkController) Start(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.PkStartRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.StartPk(userID, req.GameID, req.Degree, req.Pattern, req.LevelID, req.Difficulty)
	if err != nil {
		return mapPkError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

// StartLevel starts a level within a PK session.
func (c *GamePlayPkController) StartLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.StartLevelRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.StartPkLevel(userID, sessionID, req.GameLevelID, req.Degree, req.Pattern)
	if err != nil {
		return mapPkError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

// CompleteLevel completes a level within a PK session.
func (c *GamePlayPkController) CompleteLevel(ctx contractshttp.Context) contractshttp.Response {
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

	result, err := services.CompletePkLevel(userID, sessionID, gameLevelID, req.Score, req.MaxCombo, req.TotalItems)
	if err != nil {
		return mapPkError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

// RecordAnswer records a single answer in a PK session.
func (c *GamePlayPkController) RecordAnswer(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.RecordAnswerRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	err = services.PkRecordAnswer(userID, services.RecordAnswerInput{
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
		return mapPkError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// RecordSkip records a skip in a PK session.
func (c *GamePlayPkController) RecordSkip(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.RecordSkipRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	err = services.PkRecordSkip(userID, services.RecordSkipInput{
		GameSessionTotalID: sessionID,
		GameLevelID:        req.GameLevelID,
		PlayTime:           req.PlayTime,
		NextContentItemID:  req.NextContentItemID,
	})
	if err != nil {
		return mapPkError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// SyncPlayTime syncs playtime for a PK session.
func (c *GamePlayPkController) SyncPlayTime(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.SyncPlayTimeRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.PkSyncPlayTime(userID, sessionID, req.GameLevelID, req.PlayTime); err != nil {
		if errors.Is(err, services.ErrInvalidPlayTime) {
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "游玩时长必须在0到86400秒之间")
		}
		return mapPkError(ctx, err)
	}

	return helpers.Success(ctx, map[string]bool{"ok": true})
}

// Restore returns accumulated stats for restoring client state in a PK session.
func (c *GamePlayPkController) Restore(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.RestoreSessionRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.PkRestoreSessionData(userID, sessionID, req.GameLevelID)
	if err != nil {
		return mapPkError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

// UpdateContentItem updates the PK session's current content item.
func (c *GamePlayPkController) UpdateContentItem(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	sessionID := ctx.Request().Route("id")

	var req requests.UpdateContentItemRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.PkUpdateContentItem(userID, sessionID, req.ContentItemID); err != nil {
		return mapPkError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// End forcefully ends a PK match.
func (c *GamePlayPkController) End(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	if err := services.EndPk(userID, pkID); err != nil {
		return mapPkError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// NextLevel advances the PK match to the next level.
func (c *GamePlayPkController) NextLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	var req requests.PkNextLevelRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.NextPkLevel(userID, pkID, req.CurrentLevelID); err != nil {
		return mapPkError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Pause pauses the robot goroutine in a PK match.
func (c *GamePlayPkController) Pause(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	if err := services.PausePkRobot(userID, pkID); err != nil {
		return mapPkError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Resume resumes the robot goroutine in a PK match.
func (c *GamePlayPkController) Resume(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	if err := services.ResumePkRobot(userID, pkID); err != nil {
		return mapPkError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Events establishes a persistent SSE connection for PK events.
func (c *GamePlayPkController) Events(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, 0, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	w := ctx.Response().Writer()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	conn := helpers.PkHub.Register(pkID, userID, w)
	defer func() {
		helpers.PkHub.Unregister(pkID, userID, conn)
		// Do NOT call OnPkDisconnect here — SSE reconnects frequently
		// (e.g. nginx proxy timeout). Sessions are only ended by explicit
		// user actions (EndPk, exit button) or robot timeout.
	}()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	clientGone := ctx.Request().Origin().Context().Done()

	for {
		select {
		case <-clientGone:
			return nil
		case <-conn.Done():
			return nil
		case <-ticker.C:
			if err := conn.SendHeartbeat(); err != nil {
				return nil
			}
		}
	}
}

// mapPkError maps PK service errors to HTTP responses.
func mapPkError(ctx contractshttp.Context, err error) contractshttp.Response {
	switch {
	case errors.Is(err, services.ErrPkNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodePkNotFound, "PK对战不存在")
	case errors.Is(err, services.ErrPkNotPlaying):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodePkNotPlaying, "没有进行中的PK对战")
	case errors.Is(err, services.ErrNoMockUserAvail):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeNoMockUser, "没有可用的对手，请稍后再试")
	case errors.Is(err, services.ErrGameNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeGameNotFound, "游戏不存在")
	case errors.Is(err, services.ErrGameNotPublished):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "游戏未发布")
	case errors.Is(err, services.ErrNoGameLevels):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "游戏没有关卡")
	case errors.Is(err, services.ErrLevelNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "关卡不存在")
	case errors.Is(err, services.ErrSessionNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeSessionNotFound, "会话不存在")
	case errors.Is(err, services.ErrSessionLevelNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeSessionNotFound, "关卡会话不存在")
	case errors.Is(err, services.ErrVipRequired):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
	case errors.Is(err, services.ErrForbidden):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeForbidden, "forbidden")
	case errors.Is(err, services.ErrRateLimited):
		return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "操作过于频繁，请稍后再试")
	case errors.Is(err, services.ErrLastLevel):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, err.Error())
	default:
		fmt.Printf("[PK] unhandled error: %v\n", err)
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "internal server error")
	}
}

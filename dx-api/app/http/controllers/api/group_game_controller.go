package api

import (
	"errors"
	nethttp "net/http"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	"dx-api/app/models"
	services "dx-api/app/services/api"
)

type GroupGameController struct{}

func NewGroupGameController() *GroupGameController {
	return &GroupGameController{}
}

// SearchGames searches published games for group game selection.
func (c *GroupGameController) SearchGames(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	if err := services.VerifyGroupOwnership(userID, id); err != nil {
		return mapGroupGameError(ctx, err)
	}

	q := ctx.Request().Query("q", "")
	limit := ctx.Request().QueryInt("limit", 20)
	items, err := services.SearchGamesForGroup(userID, q, limit)
	if err != nil {
		return mapGroupGameError(ctx, err)
	}

	return helpers.Success(ctx, items)
}

// SetGame sets the current game for a group.
func (c *GroupGameController) SetGame(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	var req requests.SetGroupGameRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.SetGroupGame(userID, id, req.GameID, req.GameMode, req.LevelTimeLimit, req.StartGameLevelID); err != nil {
		return mapGroupGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// ClearGame clears the current game for a group.
func (c *GroupGameController) ClearGame(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	if err := services.ClearGroupGame(userID, id); err != nil {
		return mapGroupGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// StartGame initiates a group game round.
func (c *GroupGameController) StartGame(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	groupID := ctx.Request().Route("id")
	if groupID == "" {
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	var req requests.StartGroupGameRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.StartGroupGame(userID, groupID, req.Degree, req.Pattern); err != nil {
		return mapGroupGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// ForceEnd force-ends the current group game round.
func (c *GroupGameController) ForceEnd(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	groupID := ctx.Request().Route("id")
	if groupID == "" {
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	results, err := services.ForceEndGroupGame(userID, groupID)
	if err != nil {
		return mapGroupGameError(ctx, err)
	}

	return helpers.Success(ctx, map[string]any{
		"results": results,
	})
}

// NextLevel triggers the next level for all group members.
func (c *GroupGameController) NextLevel(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	groupID := ctx.Request().Route("id")
	if groupID == "" {
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	var req requests.NextLevelRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.NextGroupLevel(userID, groupID, req.CurrentLevelID); err != nil {
		return mapGroupGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Events establishes a persistent SSE connection for group events.
func (c *GroupGameController) Events(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, 0, "unauthorized")
	}

	groupID := ctx.Request().Route("id")

	var member models.GameGroupMember
	if err := facades.Orm().Query().Where("game_group_id", groupID).
		Where("user_id", userID).First(&member); err != nil || member.ID == "" {
		return helpers.Error(ctx, nethttp.StatusForbidden, 0, "not a group member")
	}

	w := ctx.Response().Writer()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	if f, ok := w.(nethttp.Flusher); ok {
		f.Flush()
	}

	conn := helpers.GroupSSEHub.Register(groupID, userID, w)
	defer func() {
		helpers.GroupSSEHub.Unregister(groupID, userID, conn)
		// Re-check winner for in-progress levels after disconnect.
		// Safe and non-destructive — no sessions are ended. Since
		// CheckAndDetermineWinner only counts connected players, a
		// disconnect may unblock waiting players.
		services.RecheckGroupWinners(groupID)
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

// RoomMembers returns the list of users currently connected to the group SSE (in the game room).
func (c *GroupGameController) RoomMembers(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	groupID := ctx.Request().Route("id")
	if groupID == "" {
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	connectedIDs := helpers.GroupSSEHub.ConnectedUserIDs(groupID)

	type roomMember struct {
		UserID   string `json:"user_id"`
		UserName string `json:"user_name"`
	}

	members := make([]roomMember, 0, len(connectedIDs))
	for _, uid := range connectedIDs {
		var user models.User
		if err := facades.Orm().Query().Select("id", "username", "nickname").Where("id", uid).First(&user); err == nil && user.ID != "" {
			name := user.Username
			if user.Nickname != nil && *user.Nickname != "" {
				name = *user.Nickname
			}
			members = append(members, roomMember{UserID: uid, UserName: name})
		}
	}

	return helpers.Success(ctx, members)
}

// mapGroupGameError maps service errors to HTTP responses.
func mapGroupGameError(ctx contractshttp.Context, err error) contractshttp.Response {
	switch {
	case errors.Is(err, services.ErrGroupNotFound):
		return helpers.Error(ctx, nethttp.StatusNotFound, consts.CodeGroupNotFound, "学习群不存在")
	case errors.Is(err, services.ErrNotGroupOwner):
		return helpers.Error(ctx, nethttp.StatusForbidden, consts.CodeGroupForbidden, "无权操作此学习群")
	case errors.Is(err, services.ErrGameNotFound):
		return helpers.Error(ctx, nethttp.StatusNotFound, consts.CodeNotFound, "游戏不存在")
	case errors.Is(err, services.ErrGameNotPublished):
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "游戏未发布")
	case errors.Is(err, services.ErrGroupIsPlaying):
		return helpers.Error(ctx, nethttp.StatusConflict, consts.CodeValidationError, "游戏正在进行中")
	case errors.Is(err, services.ErrGroupNotPlaying):
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "没有正在进行的游戏")
	case errors.Is(err, services.ErrNoGameSet):
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "未设置当前游戏")
	case errors.Is(err, services.ErrNoGameModeSet):
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, "未设置游戏模式")
	case errors.Is(err, services.ErrNotEnoughMembers):
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, err.Error())
	case errors.Is(err, services.ErrNotEnoughSubgroups):
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, err.Error())
	case errors.Is(err, services.ErrUnequalSubgroups):
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, err.Error())
	case errors.Is(err, services.ErrLevelNotFound):
		return helpers.Error(ctx, nethttp.StatusNotFound, consts.CodeNotFound, "关卡不存在")
	case errors.Is(err, services.ErrLastLevel):
		return helpers.Error(ctx, nethttp.StatusBadRequest, consts.CodeValidationError, err.Error())
	case errors.Is(err, services.ErrNotGroupMemberForAction):
		return helpers.Error(ctx, nethttp.StatusForbidden, consts.CodeGroupForbidden, "非群组成员")
	case errors.Is(err, services.ErrVipRequired):
		return helpers.Error(ctx, nethttp.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
	default:
		return helpers.Error(ctx, nethttp.StatusInternalServerError, consts.CodeInternalError, "操作失败")
	}
}

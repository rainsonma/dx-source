package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
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
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	if err := services.VerifyGroupOwnership(userID, id); err != nil {
		return mapGroupGameError(ctx, err)
	}

	q := ctx.Request().Query("q", "")
	items, err := services.SearchGamesForGroup(q, 20)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "操作失败")
	}

	return helpers.Success(ctx, items)
}

// SetGame sets the current game for a group.
func (c *GroupGameController) SetGame(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	var req requests.SetGroupGameRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.SetGroupGame(userID, id, req.GameID, req.GameMode); err != nil {
		return mapGroupGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// ClearGame clears the current game for a group.
func (c *GroupGameController) ClearGame(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	if err := services.ClearGroupGame(userID, id); err != nil {
		return mapGroupGameError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// mapGroupGameError maps service errors to HTTP responses.
func mapGroupGameError(ctx contractshttp.Context, err error) contractshttp.Response {
	switch {
	case errors.Is(err, services.ErrGroupNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeGroupNotFound, "学习群不存在")
	case errors.Is(err, services.ErrNotGroupOwner):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeGroupForbidden, "无权操作此学习群")
	case errors.Is(err, services.ErrGameNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeNotFound, "游戏不存在")
	case errors.Is(err, services.ErrGameNotPublished):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "游戏未发布")
	default:
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "操作失败")
	}
}

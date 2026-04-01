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

type GameController struct{}

// List returns published games with cursor pagination and optional filters.
func (c *GameController) List(ctx contractshttp.Context) contractshttp.Response {
	var req requests.ListGamesRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	limit := req.Limit
	if limit <= 0 {
		limit = helpers.DefaultCursorLimit
	} else if limit > 50 {
		limit = 50
	}

	games, nextCursor, hasMore, err := services.ListPublishedGames(req.Cursor, limit, req.ParseCategoryIDs(), req.PressID, req.Mode)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list games")
	}

	return helpers.Paginated(ctx, games, nextCursor, hasMore)
}

// Search returns published games matching a name query.
func (c *GameController) Search(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SearchGamesRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	} else if limit > 50 {
		limit = 50
	}

	games, err := services.SearchGames(req.Query, limit)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to search games")
	}

	return helpers.Success(ctx, games)
}

// Played returns all games the authenticated user has played.
func (c *GameController) Played(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	games, err := services.GetPlayedGames(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get played games")
	}

	return helpers.Success(ctx, games)
}

// Detail returns full game detail with levels.
func (c *GameController) Detail(ctx contractshttp.Context) contractshttp.Response {
	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id is required")
	}

	detail, err := services.GetGameDetail(gameID)
	if err != nil {
		if errors.Is(err, services.ErrGameNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeGameNotFound, "游戏不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get game detail")
	}

	return helpers.Success(ctx, detail)
}

package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type UserFavoriteController struct{}

func NewUserFavoriteController() *UserFavoriteController {
	return &UserFavoriteController{}
}

// ToggleFavorite toggles a game favorite.
func (c *UserFavoriteController) ToggleFavorite(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.ToggleFavoriteRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.ToggleFavorite(userID, req.GameID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to toggle favorite")
	}

	return helpers.Success(ctx, result)
}

// ListFavorites returns the user's favorite games.
func (c *UserFavoriteController) ListFavorites(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	items, err := services.ListFavorites(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list favorites")
	}

	return helpers.Success(ctx, items)
}

// Favorited checks whether the user has favorited a specific game.
func (c *UserFavoriteController) Favorited(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id is required")
	}

	favorited, err := services.IsGameFavorited(userID, gameID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to check favorite status")
	}

	return helpers.Success(ctx, map[string]bool{"favorited": favorited})
}

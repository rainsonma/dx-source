package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type GameStatsController struct{}

func NewGameStatsController() *GameStatsController {
	return &GameStatsController{}
}

// Stats returns the user's stats for a specific game.
func (c *GameStatsController) Stats(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	gameID := ctx.Request().Route("id")
	if gameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game id is required")
	}

	stats, err := services.GetGameStats(userID, gameID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get game stats")
	}

	return helpers.Success(ctx, stats)
}

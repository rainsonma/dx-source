package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type LeaderboardController struct{}

func NewLeaderboardController() *LeaderboardController {
	return &LeaderboardController{}
}

// GetLeaderboard returns the leaderboard by type and period.
func (c *LeaderboardController) GetLeaderboard(ctx contractshttp.Context) contractshttp.Response {
	userID, _ := facades.Auth(ctx).Guard("user").ID()

	lbType := ctx.Request().Query("type", "exp")
	period := ctx.Request().Query("period", "all")

	if lbType != "exp" && lbType != "playtime" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "类型必须是经验值或游玩时长")
	}
	if period != "all" && period != "day" && period != "week" && period != "month" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "时间范围必须是全部、日、周或月")
	}

	result, err := services.GetLeaderboard(lbType, period, userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get leaderboard")
	}

	return helpers.Success(ctx, result)
}

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

type GameReportController struct{}

func NewGameReportController() *GameReportController {
	return &GameReportController{}
}

// SubmitReport creates a game content report.
func (c *GameReportController) SubmitReport(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.SubmitReportRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.SubmitReport(userID, req.GameID, req.GameLevelID, req.ContentItemID, req.ContentVocabID, req.Reason, req.Note)
	if err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "举报过于频繁，请稍后再试")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to submit report")
	}

	return helpers.Success(ctx, result)
}

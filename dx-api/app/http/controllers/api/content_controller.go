package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type ContentController struct{}

// LevelContent returns content items for a game level, filtered by degree.
func (c *ContentController) LevelContent(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	levelID := ctx.Request().Route("levelId")
	if levelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "level id is required")
	}

	degree := ctx.Request().Query("degree", consts.GameDegreePractice)

	items, err := services.GetLevelContent(levelID, degree)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get level content")
	}

	return helpers.Success(ctx, items)
}

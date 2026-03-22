package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/constants"
	"github.com/goravel/framework/facades"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type ContentController struct{}

// LevelContent returns content items for a game level, filtered by degree.
func (c *ContentController) LevelContent(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	levelID := ctx.Request().Route("levelId")
	if levelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "level id is required")
	}

	degree := ctx.Request().Query("degree", constants.GameDegreePractice)

	items, err := services.GetLevelContent(levelID, degree)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to get level content")
	}

	return helpers.Success(ctx, items)
}

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

	var req requests.LevelContentRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	items, err := services.GetLevelContent(userID, levelID, req.Degree)
	if err != nil {
		if errors.Is(err, services.ErrVipRequired) {
			return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
		}
		if errors.Is(err, services.ErrLevelNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "关卡不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get level content")
	}

	return helpers.Success(ctx, items)
}

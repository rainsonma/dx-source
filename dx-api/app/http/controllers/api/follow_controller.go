package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type FollowController struct{}

func NewFollowController() *FollowController {
	return &FollowController{}
}

// ToggleFollow follows or unfollows a user.
func (c *FollowController) ToggleFollow(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	targetUserID := ctx.Request().Route("id")
	result, err := services.ToggleFollow(userID, targetUserID)
	if err != nil {
		if errors.Is(err, services.ErrSelfFollow) {
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "不能关注自己")
		}
		if errors.Is(err, services.ErrUserNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "用户不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to toggle follow")
	}

	return helpers.Success(ctx, result)
}

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

type UserVerifyController struct{}

func NewUserVerifyController() *UserVerifyController {
	return &UserVerifyController{}
}

// VerifyOnline checks if a user exists, is online, and has active VIP.
func (c *UserVerifyController) VerifyOnline(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.VerifyOnlineRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.VerifyUserOnline(userID, req.Username)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "用户不存在")
		case errors.Is(err, services.ErrCannotChallengeSelf):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeCannotChallengeSelf, "不能挑战自己")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "internal server error")
		}
	}

	return helpers.Success(ctx, result)
}

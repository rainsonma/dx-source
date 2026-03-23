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

type UserRedeemController struct{}

func NewUserRedeemController() *UserRedeemController {
	return &UserRedeemController{}
}

// GetRedeems returns the user's redemption records.
func (c *UserRedeemController) GetRedeems(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	page, pageSize, _ := helpers.ParseOffsetParams(ctx, 15)

	items, total, err := services.GetRedeems(userID, page, pageSize)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get redeems")
	}

	return helpers.PaginatedOffset(ctx, items, total, page, pageSize)
}

// RedeemCode processes a redemption code.
func (c *UserRedeemController) RedeemCode(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.RedeemCodeRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.RedeemCode(userID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRedeemNotFound):
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeNotFound, "兑换码不存在")
		case errors.Is(err, services.ErrRedeemAlreadyUsed):
			return helpers.Error(ctx, http.StatusConflict, consts.CodeValidationError, "兑换码已使用")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to redeem code")
		}
	}

	return helpers.Success(ctx, result)
}

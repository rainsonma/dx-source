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

type UserReferralController struct{}

func NewUserReferralController() *UserReferralController {
	return &UserReferralController{}
}

// GetInviteData returns invite code, stats, and first page of referrals.
func (c *UserReferralController) GetInviteData(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	data, err := services.GetInviteData(userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "用户不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get invite data")
	}

	return helpers.Success(ctx, data)
}

// GetReferrals returns paginated referral records.
func (c *UserReferralController) GetReferrals(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	page, pageSize, _ := helpers.ParseOffsetParams(ctx, 15)

	referrals, err := services.GetReferrals(userID, page, pageSize)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get referrals")
	}

	total, err := services.CountReferrals(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to count referrals")
	}

	return helpers.PaginatedOffset(ctx, referrals, total, page, pageSize)
}

// ValidateCode is a public endpoint that reports whether an invite_code is valid.
func (c *UserReferralController) ValidateCode(ctx contractshttp.Context) contractshttp.Response {
	code := ctx.Request().Query("code", "")
	ok, err := services.ValidateInviteCode(code)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to validate invite")
	}
	return helpers.Success(ctx, map[string]any{"valid": ok})
}

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
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "user not found")
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

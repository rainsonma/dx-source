package adm

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/adm"
	services "dx-api/app/services/adm"
)

type RedeemController struct{}

func NewRedeemController() *RedeemController {
	return &RedeemController{}
}

// GenerateCodes generates a batch of redeem codes.
func (c *RedeemController) GenerateCodes(ctx contractshttp.Context) contractshttp.Response {
	var req requests.GenerateCodesRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid request")
	}

	validGrades := map[string]bool{
		consts.UserGradeMonth:    true,
		consts.UserGradeSeason:   true,
		consts.UserGradeYear:     true,
		consts.UserGradeLifetime: true,
	}
	if !validGrades[req.Grade] {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid grade")
	}

	validCounts := map[int]bool{10: true, 50: true, 100: true, 500: true}
	if !validCounts[req.Count] {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "count must be 10, 50, 100, or 500")
	}

	count, err := services.GenerateCodes(req.Grade, req.Count)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to generate codes")
	}

	return helpers.Success(ctx, map[string]int{"count": count})
}

// GetAllRedeems returns all redeem codes (admin view, paginated).
func (c *RedeemController) GetAllRedeems(ctx contractshttp.Context) contractshttp.Response {
	page, pageSize, _ := helpers.ParseOffsetParams(ctx, 15)

	items, total, err := services.GetAllRedeems(page, pageSize)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get redeems")
	}

	return helpers.PaginatedOffset(ctx, items, total, page, pageSize)
}

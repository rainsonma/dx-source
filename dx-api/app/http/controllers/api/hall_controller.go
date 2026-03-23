package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type HallController struct{}

func NewHallController() *HallController {
	return &HallController{}
}

// GetDashboard returns aggregated dashboard data.
func (c *HallController) GetDashboard(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	data, err := services.GetDashboard(userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "user not found")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get dashboard")
	}

	return helpers.Success(ctx, data)
}

// GetHeatmap returns daily activity counts for a year.
func (c *HallController) GetHeatmap(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	yearStr := ctx.Request().Query("year", "")
	year := 0
	if yearStr != "" {
		parsed, err := strconv.Atoi(yearStr)
		if err != nil || parsed < 2000 || parsed > 2100 {
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid year")
		}
		year = parsed
	}
	if year == 0 {
		year = currentYear()
	}

	data, err := services.GetHeatmap(userID, year)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "user not found")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get heatmap")
	}

	return helpers.Success(ctx, data)
}

// currentYear returns the current year.
func currentYear() int {
	return time.Now().Year()
}

package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/adm"
	services "dx-api/app/services/adm"
)

type AdminCommunityController struct{}

func NewAdminCommunityController() *AdminCommunityController {
	return &AdminCommunityController{}
}

// CreateNotice creates a new system notice.
func (c *AdminCommunityController) CreateNotice(ctx contractshttp.Context) contractshttp.Response {
	var req requests.CreateNoticeRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid request")
	}

	if req.Title == "" || len(req.Title) > 200 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "title must be 1-200 characters")
	}

	notice, err := services.CreateNotice(req.Title, req.Content, req.Icon)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to create notice")
	}

	return helpers.Success(ctx, notice)
}

// UpdateNotice updates an existing notice.
func (c *AdminCommunityController) UpdateNotice(ctx contractshttp.Context) contractshttp.Response {
	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "notice id is required")
	}

	var req requests.UpdateNoticeRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid request")
	}

	if req.Title == "" || len(req.Title) > 200 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "title must be 1-200 characters")
	}

	notice, err := services.UpdateNotice(id, req.Title, req.Content, req.Icon)
	if err != nil {
		if errors.Is(err, services.ErrNoticeNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeNotFound, "notice not found")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to update notice")
	}

	return helpers.Success(ctx, notice)
}

// DeleteNotice soft-deletes a notice.
func (c *AdminCommunityController) DeleteNotice(ctx contractshttp.Context) contractshttp.Response {
	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "notice id is required")
	}

	if err := services.DeleteNotice(id); err != nil {
		if errors.Is(err, services.ErrNoticeNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeNotFound, "notice not found")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to delete notice")
	}

	return helpers.Success(ctx, nil)
}

// GenerateCodes generates a batch of redeem codes.
func (c *AdminCommunityController) GenerateCodes(ctx contractshttp.Context) contractshttp.Response {
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
func (c *AdminCommunityController) GetAllRedeems(ctx contractshttp.Context) contractshttp.Response {
	page, pageSize, _ := helpers.ParseOffsetParams(ctx, 15)

	items, total, err := services.GetAllRedeems(page, pageSize)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get redeems")
	}

	return helpers.PaginatedOffset(ctx, items, total, page, pageSize)
}

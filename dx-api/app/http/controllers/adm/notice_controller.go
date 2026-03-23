package adm

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/adm"
	services "dx-api/app/services/adm"
)

type NoticeController struct{}

func NewNoticeController() *NoticeController {
	return &NoticeController{}
}

// CreateNotice creates a new system notice.
func (c *NoticeController) CreateNotice(ctx contractshttp.Context) contractshttp.Response {
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
func (c *NoticeController) UpdateNotice(ctx contractshttp.Context) contractshttp.Response {
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
func (c *NoticeController) DeleteNotice(ctx contractshttp.Context) contractshttp.Response {
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

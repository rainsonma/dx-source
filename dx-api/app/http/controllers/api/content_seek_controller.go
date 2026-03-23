package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type ContentSeekController struct{}

func NewContentSeekController() *ContentSeekController {
	return &ContentSeekController{}
}

// GetContentSeeks returns the user's content seek records.
func (c *ContentSeekController) GetContentSeeks(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	items, err := services.GetContentSeeks(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get content seeks")
	}

	return helpers.Success(ctx, items)
}

// SubmitContentSeek creates or updates a content seek record.
func (c *ContentSeekController) SubmitContentSeek(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.SubmitContentSeekRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid request")
	}

	if req.CourseName == "" || len(req.CourseName) > 30 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "course_name must be 1-30 characters")
	}
	if req.Description == "" || len(req.Description) > 30 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "description must be 1-30 characters")
	}
	if req.DiskUrl == "" || len(req.DiskUrl) > 30 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "disk_url must be 1-30 characters")
	}

	result, err := services.SubmitContentSeek(userID, req.CourseName, req.Description, req.DiskUrl)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to submit content seek")
	}

	return helpers.Success(ctx, result)
}

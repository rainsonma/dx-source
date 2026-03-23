package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type UserUnknownController struct{}

func NewUserUnknownController() *UserUnknownController {
	return &UserUnknownController{}
}

// MarkUnknown marks a content item as unknown.
func (c *UserUnknownController) MarkUnknown(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.MarkUnknownRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid request")
	}
	if req.ContentItemID == "" || req.GameID == "" || req.GameLevelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "content_item_id, game_id, and game_level_id are required")
	}

	if err := services.MarkAsUnknown(userID, req.ContentItemID, req.GameID, req.GameLevelID); err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "操作过于频繁，请稍后再试")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to mark as unknown")
	}

	return helpers.Success(ctx, nil)
}

// ListUnknown returns paginated unknown items.
func (c *UserUnknownController) ListUnknown(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	cursor, limit := helpers.ParseCursorParams(ctx, 20)
	items, nextCursor, hasMore, err := services.ListUnknown(userID, cursor, limit)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list unknown")
	}

	return helpers.Paginated(ctx, items, nextCursor, hasMore)
}

// UnknownStats returns unknown word statistics.
func (c *UserUnknownController) UnknownStats(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	stats, err := services.GetUnknownStats(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get stats")
	}

	return helpers.Success(ctx, stats)
}

// DeleteUnknown removes a single unknown entry.
func (c *UserUnknownController) DeleteUnknown(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if err := services.DeleteUnknown(userID, id); err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to delete")
	}

	return helpers.Success(ctx, nil)
}

// BulkDeleteUnknown removes multiple unknown entries.
func (c *UserUnknownController) BulkDeleteUnknown(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.BulkDeleteRequest
	if err := ctx.Request().Bind(&req); err != nil || len(req.IDs) == 0 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "ids are required")
	}

	if err := services.BulkDeleteUnknown(userID, req.IDs); err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to bulk delete")
	}

	return helpers.Success(ctx, nil)
}

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

type UserMasterController struct{}

func NewUserMasterController() *UserMasterController {
	return &UserMasterController{}
}

// MarkMastered marks a content item as mastered.
func (c *UserMasterController) MarkMastered(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.MarkMasteredRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.MarkAsMastered(userID, req.ContentItemID, req.GameID, req.GameLevelID); err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "操作过于频繁，请稍后再试")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to mark as mastered")
	}

	return helpers.Success(ctx, nil)
}

// ListMastered returns paginated mastered items.
func (c *UserMasterController) ListMastered(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	cursor, limit := helpers.ParseCursorParams(ctx, 20)
	items, nextCursor, hasMore, err := services.ListMastered(userID, cursor, limit)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list mastered")
	}

	return helpers.Paginated(ctx, items, nextCursor, hasMore)
}

// MasterStats returns mastered word statistics.
func (c *UserMasterController) MasterStats(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	stats, err := services.GetMasterStats(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get stats")
	}

	return helpers.Success(ctx, stats)
}

// DeleteMastered removes a single mastered entry.
func (c *UserMasterController) DeleteMastered(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if err := services.DeleteMastered(userID, id); err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to delete")
	}

	return helpers.Success(ctx, nil)
}

// BulkDeleteMastered removes multiple mastered entries.
func (c *UserMasterController) BulkDeleteMastered(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.BulkDeleteRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.BulkDeleteMastered(userID, req.IDs); err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to bulk delete")
	}

	return helpers.Success(ctx, nil)
}

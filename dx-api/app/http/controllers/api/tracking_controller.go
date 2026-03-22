package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"github.com/goravel/framework/facades"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type TrackingController struct{}

func NewTrackingController() *TrackingController {
	return &TrackingController{}
}

// --- Mastered ---

// MarkMastered marks a content item as mastered.
func (c *TrackingController) MarkMastered(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.MarkTrackingRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid request")
	}
	if req.ContentItemID == "" || req.GameID == "" || req.GameLevelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "content_item_id, game_id, and game_level_id are required")
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
func (c *TrackingController) ListMastered(ctx contractshttp.Context) contractshttp.Response {
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
func (c *TrackingController) MasterStats(ctx contractshttp.Context) contractshttp.Response {
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
func (c *TrackingController) DeleteMastered(ctx contractshttp.Context) contractshttp.Response {
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
func (c *TrackingController) BulkDeleteMastered(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.BulkDeleteRequest
	if err := ctx.Request().Bind(&req); err != nil || len(req.IDs) == 0 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "ids are required")
	}

	if err := services.BulkDeleteMastered(userID, req.IDs); err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to bulk delete")
	}

	return helpers.Success(ctx, nil)
}

// --- Unknown ---

// MarkUnknown marks a content item as unknown.
func (c *TrackingController) MarkUnknown(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.MarkTrackingRequest
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
func (c *TrackingController) ListUnknown(ctx contractshttp.Context) contractshttp.Response {
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
func (c *TrackingController) UnknownStats(ctx contractshttp.Context) contractshttp.Response {
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
func (c *TrackingController) DeleteUnknown(ctx contractshttp.Context) contractshttp.Response {
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
func (c *TrackingController) BulkDeleteUnknown(ctx contractshttp.Context) contractshttp.Response {
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

// --- Review ---

// MarkReview marks a content item for review.
func (c *TrackingController) MarkReview(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.MarkTrackingRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid request")
	}
	if req.ContentItemID == "" || req.GameID == "" || req.GameLevelID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "content_item_id, game_id, and game_level_id are required")
	}

	if err := services.MarkAsReview(userID, req.ContentItemID, req.GameID, req.GameLevelID); err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "操作过于频繁，请稍后再试")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to mark for review")
	}

	return helpers.Success(ctx, nil)
}

// ListReviews returns paginated review items.
func (c *TrackingController) ListReviews(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	cursor, limit := helpers.ParseCursorParams(ctx, 20)
	items, nextCursor, hasMore, err := services.ListReviews(userID, cursor, limit)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list reviews")
	}

	return helpers.Paginated(ctx, items, nextCursor, hasMore)
}

// ReviewStats returns review statistics.
func (c *TrackingController) ReviewStats(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	stats, err := services.GetReviewStats(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get stats")
	}

	return helpers.Success(ctx, stats)
}

// DeleteReview removes a single review entry.
func (c *TrackingController) DeleteReview(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if err := services.DeleteReview(userID, id); err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to delete")
	}

	return helpers.Success(ctx, nil)
}

// BulkDeleteReviews removes multiple review entries.
func (c *TrackingController) BulkDeleteReviews(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.BulkDeleteRequest
	if err := ctx.Request().Bind(&req); err != nil || len(req.IDs) == 0 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "ids are required")
	}

	if err := services.BulkDeleteReviews(userID, req.IDs); err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to bulk delete")
	}

	return helpers.Success(ctx, nil)
}

// --- Favorites ---

// ToggleFavorite toggles a game favorite.
func (c *TrackingController) ToggleFavorite(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.ToggleFavoriteRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid request")
	}
	if req.GameID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "game_id is required")
	}

	result, err := services.ToggleFavorite(userID, req.GameID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to toggle favorite")
	}

	return helpers.Success(ctx, result)
}

// ListFavorites returns the user's favorite games.
func (c *TrackingController) ListFavorites(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	items, err := services.ListFavorites(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list favorites")
	}

	return helpers.Success(ctx, items)
}

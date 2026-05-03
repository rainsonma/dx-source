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

type UserReviewController struct{}

func NewUserReviewController() *UserReviewController {
	return &UserReviewController{}
}

// MarkReview marks a content item for review.
func (c *UserReviewController) MarkReview(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.MarkReviewRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.MarkAsReview(userID, req.ContentItemID, req.ContentVocabID, req.GameID, req.GameLevelID); err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "操作过于频繁，请稍后再试")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to mark for review")
	}

	return helpers.Success(ctx, nil)
}

// ListReviews returns paginated review items.
func (c *UserReviewController) ListReviews(ctx contractshttp.Context) contractshttp.Response {
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
func (c *UserReviewController) ReviewStats(ctx contractshttp.Context) contractshttp.Response {
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
func (c *UserReviewController) DeleteReview(ctx contractshttp.Context) contractshttp.Response {
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
func (c *UserReviewController) BulkDeleteReviews(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.BulkDeleteRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.BulkDeleteReviews(userID, req.IDs); err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to bulk delete")
	}

	return helpers.Success(ctx, nil)
}

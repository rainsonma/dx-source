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

type PostInteractController struct{}

func NewPostInteractController() *PostInteractController {
	return &PostInteractController{}
}

// ToggleLike likes or unlikes a post.
func (c *PostInteractController) ToggleLike(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	postID := ctx.Request().Route("id")
	result, err := services.ToggleLike(userID, postID)
	if err != nil {
		if errors.Is(err, services.ErrPostNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodePostNotFound, "帖子不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to toggle like")
	}

	return helpers.Success(ctx, result)
}

// ToggleBookmark bookmarks or unbookmarks a post.
func (c *PostInteractController) ToggleBookmark(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	postID := ctx.Request().Route("id")
	result, err := services.ToggleBookmark(userID, postID)
	if err != nil {
		if errors.Is(err, services.ErrPostNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodePostNotFound, "帖子不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to toggle bookmark")
	}

	return helpers.Success(ctx, result)
}

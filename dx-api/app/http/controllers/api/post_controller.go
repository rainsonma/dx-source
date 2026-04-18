package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type PostController struct{}

func NewPostController() *PostController {
	return &PostController{}
}

// Create creates a new post.
func (c *PostController) Create(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CreatePostRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if req.ImageURL != nil && *req.ImageURL != "" {
		if !helpers.IsUploadedImageURL(*req.ImageURL) {
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的图片URL")
		}
	}

	result, err := services.CreatePost(userID, req.Content, req.ImageURL, req.Tags)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to create post")
	}

	return helpers.Success(ctx, result)
}

// List returns paginated posts for the given tab.
func (c *PostController) List(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	tab := ctx.Request().Query("tab", "latest")
	cursor, limit := helpers.ParseCursorParams(ctx, helpers.DefaultCursorLimit)

	items, nextCursor, hasMore, err := services.ListPosts(userID, tab, cursor, limit)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list posts")
	}

	return helpers.Paginated(ctx, items, nextCursor, hasMore)
}

// Show returns a single post.
func (c *PostController) Show(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	postID := ctx.Request().Route("id")
	result, err := services.GetPost(userID, postID)
	if err != nil {
		return c.mapPostError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

// Update updates an owned post.
func (c *PostController) Update(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.UpdatePostRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if req.ImageURL != nil && *req.ImageURL != "" {
		if !helpers.IsUploadedImageURL(*req.ImageURL) {
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的图片URL")
		}
	}

	postID := ctx.Request().Route("id")
	err = services.UpdatePost(userID, postID, req.Content, req.ImageURL, req.Tags)
	if err != nil {
		return c.mapPostError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Delete soft-deletes an owned post.
func (c *PostController) Delete(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	postID := ctx.Request().Route("id")
	err = services.DeletePost(userID, postID)
	if err != nil {
		return c.mapPostError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

func (c *PostController) mapPostError(ctx contractshttp.Context, err error) contractshttp.Response {
	if errors.Is(err, services.ErrPostNotFound) {
		return helpers.Error(ctx, http.StatusNotFound, consts.CodePostNotFound, "帖子不存在")
	}
	if errors.Is(err, services.ErrPostNotOwner) {
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeForbidden, "无权操作此帖子")
	}
	return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "internal error")
}

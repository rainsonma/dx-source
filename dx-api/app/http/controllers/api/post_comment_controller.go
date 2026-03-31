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

type PostCommentController struct{}

func NewPostCommentController() *PostCommentController {
	return &PostCommentController{}
}

// Create creates a comment (or reply) on a post.
func (c *PostCommentController) Create(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CreateCommentRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	postID := ctx.Request().Route("id")
	result, err := services.CreateComment(userID, postID, req.ParentID, req.Content)
	if err != nil {
		return c.mapCommentError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

// List returns paginated comments with replies for a post.
func (c *PostCommentController) List(ctx contractshttp.Context) contractshttp.Response {
	postID := ctx.Request().Route("id")
	cursor, limit := helpers.ParseCursorParams(ctx, helpers.DefaultCursorLimit)

	items, nextCursor, hasMore, err := services.ListComments(postID, cursor, limit)
	if err != nil {
		return c.mapCommentError(ctx, err)
	}

	return helpers.Paginated(ctx, items, nextCursor, hasMore)
}

// Update updates an owned comment.
func (c *PostCommentController) Update(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.UpdateCommentRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	postID := ctx.Request().Route("id")
	commentID := ctx.Request().Route("commentId")
	err = services.UpdateComment(userID, postID, commentID, req.Content)
	if err != nil {
		return c.mapCommentError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Delete deletes an owned comment and its replies.
func (c *PostCommentController) Delete(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	postID := ctx.Request().Route("id")
	commentID := ctx.Request().Route("commentId")
	err = services.DeleteComment(userID, postID, commentID)
	if err != nil {
		return c.mapCommentError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

func (c *PostCommentController) mapCommentError(ctx contractshttp.Context, err error) contractshttp.Response {
	if errors.Is(err, services.ErrPostNotFound) {
		return helpers.Error(ctx, http.StatusNotFound, consts.CodePostNotFound, "帖子不存在")
	}
	if errors.Is(err, services.ErrCommentNotFound) {
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeCommentNotFound, "评论不存在")
	}
	if errors.Is(err, services.ErrCommentNotOwner) {
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeForbidden, "无权操作此评论")
	}
	if errors.Is(err, services.ErrNestedReply) {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "不能回复评论的回复")
	}
	return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "internal error")
}

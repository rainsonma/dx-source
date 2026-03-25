package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type GroupMemberController struct{}

func NewGroupMemberController() *GroupMemberController {
	return &GroupMemberController{}
}

// List returns a paginated list of members for a group.
func (c *GroupMemberController) List(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	cursor, limit := helpers.ParseCursorParams(ctx, helpers.DefaultCursorLimit)

	members, nextCursor, hasMore, err := services.ListGroupMembers(userID, id, cursor, limit)
	if err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Paginated(ctx, members, nextCursor, hasMore)
}

// Kick removes a member from the group (owner only).
func (c *GroupMemberController) Kick(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	targetUserID := ctx.Request().Route("userId")
	if id == "" || targetUserID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id and user id are required")
	}

	if err := services.KickMember(userID, id, targetUserID); err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Leave removes the current user from a group.
func (c *GroupMemberController) Leave(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	if err := services.LeaveGroup(userID, id); err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// GetInviteInfo returns public group info for a given invite code (no auth required).
func (c *GroupMemberController) GetInviteInfo(ctx contractshttp.Context) contractshttp.Response {
	code := ctx.Request().Route("code")
	if code == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invite code is required")
	}
	info, err := services.GetGroupByInviteCode(code)
	if err != nil {
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeGroupNotFound, "邀请链接无效或群组已关闭")
	}
	return helpers.Success(ctx, info)
}

// JoinByCode submits a join application for a group via invite code.
func (c *GroupMemberController) JoinByCode(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	code := ctx.Request().Route("code")
	if code == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invite code is required")
	}

	groupID, err := services.JoinByCode(userID, code)
	if err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, map[string]string{"group_id": groupID})
}

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

type GroupController struct{}

func NewGroupController() *GroupController {
	return &GroupController{}
}

// List returns a paginated list of groups the user belongs to.
func (c *GroupController) List(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	tab := ctx.Request().Query("tab", "")
	cursor, limit := helpers.ParseCursorParams(ctx, helpers.DefaultCursorLimit)

	groups, nextCursor, hasMore, err := services.ListGroups(userID, tab, cursor, limit)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "操作失败")
	}

	return helpers.Paginated(ctx, groups, nextCursor, hasMore)
}

// Create creates a new group.
func (c *GroupController) Create(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CreateGroupRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.CreateGroup(userID, req.Name, req.Description)
	if err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

// Detail returns full detail of a group.
func (c *GroupController) Detail(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	detail, err := services.GetGroupDetail(userID, id)
	if err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, detail)
}

// Update updates a group's name and description.
func (c *GroupController) Update(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	var req requests.UpdateGroupRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.UpdateGroup(userID, id, req.Name, req.Description, req.LevelTimeLimit); err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Dismiss soft-dismisses a group by setting dismissed_at.
func (c *GroupController) Dismiss(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	if err := services.DismissGroup(userID, id); err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Apply submits an application to join a group.
func (c *GroupController) Apply(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	if _, err := services.ApplyToGroup(userID, id); err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// CancelApply cancels a pending application.
func (c *GroupController) CancelApply(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	if err := services.CancelApplication(userID, id); err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// ListApplications returns pending applications for a group (owner only).
func (c *GroupController) ListApplications(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	cursor, limit := helpers.ParseCursorParams(ctx, helpers.DefaultCursorLimit)

	apps, nextCursor, hasMore, err := services.ListApplications(userID, id, cursor, limit)
	if err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Paginated(ctx, apps, nextCursor, hasMore)
}

// HandleApplication accepts or rejects an application.
func (c *GroupController) HandleApplication(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	appID := ctx.Request().Route("appId")
	if id == "" || appID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id and application id are required")
	}

	var req requests.HandleApplicationRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.HandleApplication(userID, id, appID, req.Action); err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// mapGroupError maps service errors to HTTP responses.
func mapGroupError(ctx contractshttp.Context, err error) contractshttp.Response {
	switch {
	case errors.Is(err, services.ErrGroupNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeGroupNotFound, "学习群不存在")
	case errors.Is(err, services.ErrNotGroupOwner):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeGroupForbidden, "无权操作此学习群")
	case errors.Is(err, services.ErrNotGroupMember):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeGroupForbidden, "您不是该群成员")
	case errors.Is(err, services.ErrAlreadyMember):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeAlreadyMember, "您已是该群成员")
	case errors.Is(err, services.ErrAlreadyApplied):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeAlreadyApplied, "已提交过申请")
	case errors.Is(err, services.ErrApplicationNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeApplicationNotFound, "申请不存在")
	case errors.Is(err, services.ErrCannotLeaveOwned):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "群主不能退出自己的群")
	case errors.Is(err, services.ErrSubgroupNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeNotFound, "小组不存在")
	case errors.Is(err, services.ErrGroupMembersFull):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeGroupMembersFull, "当前群组已满员")
	case errors.Is(err, services.ErrGroupSubgroupsFull):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeGroupSubgroupsFull, "每群最多 10 个小组")
	default:
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "操作失败")
	}
}

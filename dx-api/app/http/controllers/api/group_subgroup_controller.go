package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type GroupSubgroupController struct{}

func NewGroupSubgroupController() *GroupSubgroupController {
	return &GroupSubgroupController{}
}

// List returns all subgroups for a group.
func (c *GroupSubgroupController) List(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	subgroups, err := services.ListSubgroups(userID, id)
	if err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, subgroups)
}

// Create creates a new subgroup.
func (c *GroupSubgroupController) Create(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	if id == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id is required")
	}

	var req requests.CreateSubgroupRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	subgroupID, err := services.CreateSubgroup(userID, id, req.Name)
	if err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, map[string]string{"id": subgroupID})
}

// Update updates a subgroup's name.
func (c *GroupSubgroupController) Update(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	sid := ctx.Request().Route("sid")
	if id == "" || sid == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id and subgroup id are required")
	}

	var req requests.UpdateSubgroupRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.UpdateSubgroup(userID, id, sid, req.Name); err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Delete removes a subgroup.
func (c *GroupSubgroupController) Delete(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	sid := ctx.Request().Route("sid")
	if id == "" || sid == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id and subgroup id are required")
	}

	if err := services.DeleteSubgroup(userID, id, sid); err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// ListMembers returns all members of a subgroup.
func (c *GroupSubgroupController) ListMembers(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	sid := ctx.Request().Route("sid")
	if id == "" || sid == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id and subgroup id are required")
	}

	members, err := services.ListSubgroupMembers(userID, id, sid)
	if err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, members)
}

// Assign adds group members to a subgroup.
func (c *GroupSubgroupController) Assign(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	sid := ctx.Request().Route("sid")
	if id == "" || sid == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id and subgroup id are required")
	}

	var req requests.AssignSubgroupMembersRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.AssignSubgroupMembers(userID, id, sid, req.UserIDs); err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// RemoveMember removes a member from a subgroup.
func (c *GroupSubgroupController) RemoveMember(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	id := ctx.Request().Route("id")
	sid := ctx.Request().Route("sid")
	targetUserID := ctx.Request().Route("userId")
	if id == "" || sid == "" || targetUserID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "group id, subgroup id, and user id are required")
	}

	if err := services.RemoveSubgroupMember(userID, id, sid, targetUserID); err != nil {
		return mapGroupError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

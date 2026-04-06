package api

import (
	"errors"
	"fmt"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type PkInviteController struct{}

func NewPkInviteController() *PkInviteController {
	return &PkInviteController{}
}

// Invite creates a specified PK and sends an invitation.
func (c *PkInviteController) Invite(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.PkInviteRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.InvitePk(userID, req.GameID, req.GameLevelID, req.Degree, req.Pattern, req.OpponentID)
	if err != nil {
		return mapInviteError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

// Accept accepts a PK invitation.
func (c *PkInviteController) Accept(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	result, err := services.AcceptPkInvite(userID, pkID)
	if err != nil {
		return mapInviteError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

// Decline declines a PK invitation.
func (c *PkInviteController) Decline(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	if err := services.DeclinePkInvite(userID, pkID); err != nil {
		return mapInviteError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

// Details returns PK info for the room page.
func (c *PkInviteController) Details(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	pkID := ctx.Request().Route("id")

	result, err := services.GetPkDetails(userID, pkID)
	if err != nil {
		return mapInviteError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

func mapInviteError(ctx contractshttp.Context, err error) contractshttp.Response {
	switch {
	case errors.Is(err, services.ErrVipRequired):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeVipRequired, "升级会员解锁此功能")
	case errors.Is(err, services.ErrOpponentOffline):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeOpponentOffline, "对方不在线")
	case errors.Is(err, services.ErrOpponentNotVip):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeOpponentNotVip, "对方会员已过期")
	case errors.Is(err, services.ErrCannotChallengeSelf):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeCannotChallengeSelf, "不能挑战自己")
	case errors.Is(err, services.ErrPkNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodePkNotFound, "PK对战不存在")
	case errors.Is(err, services.ErrForbidden):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeForbidden, "forbidden")
	case errors.Is(err, services.ErrInvitationNotPending):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvitationNotPending, "邀请状态已变更")
	case errors.Is(err, services.ErrGameNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeGameNotFound, "游戏不存在")
	case errors.Is(err, services.ErrGameNotPublished):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "游戏未发布")
	case errors.Is(err, services.ErrLevelNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeLevelNotFound, "关卡不存在")
	default:
		fmt.Printf("[PK Invite] unhandled error: %v\n", err)
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "internal server error")
	}
}

package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type WechatAuthController struct{}

func NewWechatAuthController() *WechatAuthController {
	return &WechatAuthController{}
}

func (c *WechatAuthController) MiniSignIn(ctx contractshttp.Context) contractshttp.Response {
	var req requests.WechatMiniAuthRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	token, user, err := services.WechatMiniSignIn(ctx, req.Code)
	if err != nil {
		facades.Log().Errorf("wechat mini sign in failed: %v", err)
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "微信登录失败")
	}

	ip := ctx.Request().Ip()
	userAgent := ctx.Request().Header("User-Agent", "")
	go services.RecordLogin(user.ID, ip, userAgent, consts.PlatformMini)

	return helpers.Success(ctx, map[string]any{"token": token, "user": user})
}

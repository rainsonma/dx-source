package middleware

import (
	"fmt"

	"dx-api/app/consts"
	"dx-api/app/helpers"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

func JwtAuth() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		token := ctx.Request().Header("Authorization", "")
		if token == "" {
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    consts.CodeUnauthorized,
				Message: "unauthorized",
			}).Abort()
			return
		}

		payload, err := facades.Auth(ctx).Guard("user").Parse(token)
		if err != nil || payload == nil {
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    consts.CodeUnauthorized,
				Message: "unauthorized",
			}).Abort()
			return
		}

		// Single-device check: verify auth_id matches current session
		userID, _ := facades.Auth(ctx).Guard("user").ID()
		authID := helpers.ExtractAuthID(token)
		currentAuthID, redisErr := helpers.RedisGet(fmt.Sprintf("user_auth:%s:user", userID))
		if redisErr != nil || authID == "" || currentAuthID != authID {
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    consts.CodeSessionReplaced,
				Message: "您的账号已在其他设备登录",
			}).Abort()
			return
		}

		ctx.Request().Next()
	}
}

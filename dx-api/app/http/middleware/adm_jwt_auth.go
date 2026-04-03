package middleware

import (
	"errors"
	"strconv"

	"github.com/goravel/framework/auth"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
)

func AdmJwtAuth() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		token := ctx.Request().Cookie("dx_adm_token")
		if token == "" {
			abortUnauthorized(ctx)
			return
		}

		payload, err := facades.Auth(ctx).Guard("admin").Parse(token)
		if err != nil {
			if errors.Is(err, auth.ErrorTokenExpired) {
				newToken, refreshErr := facades.Auth(ctx).Guard("admin").Refresh()
				if refreshErr != nil {
					clearTokenCookie(ctx, "dx_adm_token")
					abortUnauthorized(ctx)
					return
				}
				setTokenCookie(ctx, "dx_adm_token", newToken)
				payload, _ = facades.Auth(ctx).Guard("admin").Parse(newToken)
			} else {
				clearTokenCookie(ctx, "dx_adm_token")
				abortUnauthorized(ctx)
				return
			}
		}

		if payload == nil {
			clearTokenCookie(ctx, "dx_adm_token")
			abortUnauthorized(ctx)
			return
		}

		userID, _ := facades.Auth(ctx).Guard("admin").ID()
		loginTsStr, redisErr := helpers.RedisGet("user_auth:" + userID + ":admin")
		if redisErr != nil {
			clearTokenCookie(ctx, "dx_adm_token")
			abortUnauthorized(ctx)
			return
		}

		loginTs, _ := strconv.ParseInt(loginTsStr, 10, 64)
		if payload.IssuedAt.Unix() < loginTs {
			clearTokenCookie(ctx, "dx_adm_token")
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    consts.CodeSessionReplaced,
				Message: "您的账号已在其他设备登录",
			}).Abort()
			return
		}

		ctx.Request().Next()
	}
}

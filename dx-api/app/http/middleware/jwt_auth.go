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

func JwtAuth() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		token := ctx.Request().Cookie("dx_token")
		if token == "" {
			abortUnauthorized(ctx)
			return
		}

		payload, err := facades.Auth(ctx).Guard("user").Parse(token)
		if err != nil {
			if errors.Is(err, auth.ErrorTokenExpired) {
				newToken, refreshErr := facades.Auth(ctx).Guard("user").Refresh()
				if refreshErr != nil {
					clearTokenCookie(ctx, "dx_token")
					abortUnauthorized(ctx)
					return
				}
				setTokenCookie(ctx, "dx_token", newToken)
				payload, _ = facades.Auth(ctx).Guard("user").Parse(newToken)
			} else {
				clearTokenCookie(ctx, "dx_token")
				abortUnauthorized(ctx)
				return
			}
		}

		if payload == nil {
			clearTokenCookie(ctx, "dx_token")
			abortUnauthorized(ctx)
			return
		}

		// Single-device check: token iat must be >= login timestamp
		userID, _ := facades.Auth(ctx).Guard("user").ID()
		loginTsStr, redisErr := helpers.RedisGet("user_auth:" + userID + ":user")
		if redisErr != nil {
			clearTokenCookie(ctx, "dx_token")
			abortUnauthorized(ctx)
			return
		}

		loginTs, _ := strconv.ParseInt(loginTsStr, 10, 64)
		if payload.IssuedAt.Unix() < loginTs {
			clearTokenCookie(ctx, "dx_token")
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    consts.CodeSessionReplaced,
				Message: "您的账号已在其他设备登录",
			}).Abort()
			return
		}

		ctx.Request().Next()
	}
}

func abortUnauthorized(ctx contractshttp.Context) {
	_ = ctx.Response().Json(401, helpers.Response{
		Code:    consts.CodeUnauthorized,
		Message: "unauthorized",
	}).Abort()
}

func setTokenCookie(ctx contractshttp.Context, name, token string) {
	secure := facades.Config().GetBool("jwt_cookie.secure", true)
	ttl := facades.Config().GetInt("jwt.refresh_ttl", 20160)
	ctx.Response().Cookie(contractshttp.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		MaxAge:   ttl * 60,
		Secure:   secure,
		HttpOnly: true,
		SameSite: "Lax",
	})
}

func clearTokenCookie(ctx contractshttp.Context, name string) {
	ctx.Response().Cookie(contractshttp.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   facades.Config().GetBool("jwt_cookie.secure", true),
		HttpOnly: true,
		SameSite: "Lax",
	})
}

func SetTokenCookie(ctx contractshttp.Context, name, token string) {
	setTokenCookie(ctx, name, token)
}

func ClearTokenCookie(ctx contractshttp.Context, name string) {
	clearTokenCookie(ctx, name)
}

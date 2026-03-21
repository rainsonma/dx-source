package middleware

import (
	"dx-api/app/helpers"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// JwtAuth verifies user JWT token for client API routes (/api/*).
// It extracts the Bearer token from the Authorization header and validates
// it using the "user" guard.
func JwtAuth() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		token := ctx.Request().Header("Authorization", "")
		if token == "" {
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    40100,
				Message: "unauthorized",
			}).Abort()
			return
		}

		// Parse validates the JWT token; Goravel strips the "Bearer " prefix internally.
		payload, err := facades.Auth(ctx).Guard("user").Parse(token)
		if err != nil || payload == nil {
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    40100,
				Message: "unauthorized",
			}).Abort()
			return
		}

		ctx.Request().Next()
	}
}

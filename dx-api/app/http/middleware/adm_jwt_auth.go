package middleware

import (
	"dx-api/app/helpers"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// AdmJwtAuth verifies admin JWT token for admin API routes (/adm/*).
// It extracts the Bearer token from the Authorization header and validates
// it using the "admin" guard.
func AdmJwtAuth() contractshttp.Middleware {
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
		payload, err := facades.Auth(ctx).Guard("admin").Parse(token)
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

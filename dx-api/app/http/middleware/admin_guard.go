package middleware

import (
	"dx-api/app/helpers"
	"dx-api/app/models"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// AdminGuard checks that the authenticated user is an admin.
// It must be placed after JwtAuth so that the user JWT is already parsed.
// The check is a simple username comparison against the hardcoded admin name.
func AdminGuard() contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		userID, err := facades.Auth(ctx).Guard("user").ID()
		if err != nil || userID == "" {
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    40100,
				Message: "unauthorized",
			}).Abort()
			return
		}

		var user models.User
		if err := facades.Orm().Query().Where("id", userID).First(&user); err != nil || user.ID == "" {
			_ = ctx.Response().Json(401, helpers.Response{
				Code:    40100,
				Message: "unauthorized",
			}).Abort()
			return
		}

		if user.Username != "rainson" {
			_ = ctx.Response().Json(403, helpers.Response{
				Code:    40300,
				Message: "forbidden",
			}).Abort()
			return
		}

		ctx.Request().Next()
	}
}

package adm

import (
	"errors"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	admservice "dx-api/app/services/adm"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

func (c *AuthController) Login(ctx contractshttp.Context) contractshttp.Response {
	username := ctx.Request().Input("username")
	password := ctx.Request().Input("password")

	if username == "" || password == "" {
		return helpers.Error(ctx, 400, consts.CodeValidationError, "username and password are required")
	}

	token, admUser, err := admservice.AdminSignIn(ctx, username, password)
	if err != nil {
		if errors.Is(err, admservice.ErrAdminNotFound) || errors.Is(err, admservice.ErrInvalidPassword) {
			return helpers.Error(ctx, 401, consts.CodeUnauthorized, "invalid username or password")
		}
		if errors.Is(err, admservice.ErrAdminInactive) {
			return helpers.Error(ctx, 403, consts.CodeForbidden, "admin account is inactive")
		}
		return helpers.Error(ctx, 500, consts.CodeInternalError, "internal server error")
	}

	return helpers.Success(ctx, map[string]any{
		"token": token,
		"user":  admUser,
	})
}

func (c *AuthController) Me(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("admin").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, 401, consts.CodeUnauthorized, "unauthorized")
	}

	admUser, err := admservice.GetAdminUser(userID)
	if err != nil {
		if errors.Is(err, admservice.ErrAdminNotFound) {
			return helpers.Error(ctx, 404, consts.CodeUserNotFound, "admin user not found")
		}
		return helpers.Error(ctx, 500, consts.CodeInternalError, "internal server error")
	}

	return helpers.Success(ctx, admUser)
}

func (c *AuthController) Logout(ctx contractshttp.Context) contractshttp.Response {
	if err := facades.Auth(ctx).Guard("admin").Logout(); err != nil {
		return helpers.Error(ctx, 500, consts.CodeInternalError, "failed to logout")
	}

	return helpers.Success(ctx, nil)
}

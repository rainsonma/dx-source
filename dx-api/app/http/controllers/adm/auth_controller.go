package adm

import (
	"errors"
	"fmt"

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

func setAdmRefreshCookie(ctx contractshttp.Context, token string) {
	secure := facades.Config().GetBool("refresh_token.cookie_secure", true)
	ttl := facades.Config().GetInt("refresh_token.ttl", 10080)
	ctx.Response().Cookie(contractshttp.Cookie{
		Name:     "dx_adm_refresh",
		Value:    token,
		Path:     "/adm/auth",
		MaxAge:   ttl * 60,
		Secure:   secure,
		HttpOnly: true,
		SameSite: "Lax",
	})
}

func clearAdmRefreshCookie(ctx contractshttp.Context) {
	ctx.Response().Cookie(contractshttp.Cookie{
		Name:     "dx_adm_refresh",
		Value:    "",
		Path:     "/adm/auth",
		MaxAge:   -1,
		Secure:   facades.Config().GetBool("refresh_token.cookie_secure", true),
		HttpOnly: true,
		SameSite: "Lax",
	})
}

func getAdmRefreshToken(ctx contractshttp.Context) string {
	if token := ctx.Request().Cookie("dx_adm_refresh"); token != "" {
		return token
	}
	return ctx.Request().Input("refresh_token")
}

func (c *AuthController) Login(ctx contractshttp.Context) contractshttp.Response {
	username := ctx.Request().Input("username")
	password := ctx.Request().Input("password")

	if username == "" || password == "" {
		return helpers.Error(ctx, 400, consts.CodeValidationError, "username and password are required")
	}

	result, admUser, err := admservice.AdminSignIn(ctx, username, password)
	if err != nil {
		if errors.Is(err, admservice.ErrAdminNotFound) || errors.Is(err, admservice.ErrInvalidPassword) {
			return helpers.Error(ctx, 401, consts.CodeUnauthorized, "invalid username or password")
		}
		if errors.Is(err, admservice.ErrAdminInactive) {
			return helpers.Error(ctx, 403, consts.CodeForbidden, "admin account is inactive")
		}
		return helpers.Error(ctx, 500, consts.CodeInternalError, "internal server error")
	}

	setAdmRefreshCookie(ctx, result.RefreshToken)
	return helpers.Success(ctx, map[string]any{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          admUser,
	})
}

func (c *AuthController) Refresh(ctx contractshttp.Context) contractshttp.Response {
	ip := ctx.Request().Ip()
	allowed, err := helpers.CheckRateLimit(fmt.Sprintf("rate:adm_refresh:%s", ip), 10, 60)
	if err != nil || !allowed {
		return helpers.Error(ctx, 429, consts.CodeRateLimited, "too many refresh requests")
	}

	oldToken := getAdmRefreshToken(ctx)
	if oldToken == "" {
		return helpers.Error(ctx, 401, consts.CodeInvalidRefreshToken, "refresh token required")
	}

	result, err := admservice.RefreshToken(ctx, oldToken)
	if err != nil {
		if errors.Is(err, admservice.ErrInvalidRefreshToken) {
			clearAdmRefreshCookie(ctx)
			return helpers.Error(ctx, 401, consts.CodeInvalidRefreshToken, "invalid or expired refresh token")
		}
		return helpers.Error(ctx, 500, consts.CodeInternalError, "failed to refresh token")
	}

	setAdmRefreshCookie(ctx, result.RefreshToken)
	return helpers.Success(ctx, map[string]any{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
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
	token := getAdmRefreshToken(ctx)
	if token != "" {
		_ = admservice.Logout(token)
	}

	clearAdmRefreshCookie(ctx)
	return helpers.Success(ctx, nil)
}

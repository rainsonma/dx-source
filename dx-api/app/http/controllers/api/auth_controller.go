package api

import (
	"errors"
	"fmt"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	"dx-api/app/models"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

// SignUp registers a new user.
func (c *AuthController) SignUp(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SignUpRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, user, err := services.SignUp(ctx, req.Email, req.Code, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCode):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidCode, "验证码无效或已过期")
		case errors.Is(err, services.ErrDuplicateEmail):
			return helpers.Error(ctx, http.StatusConflict, consts.CodeDuplicateEmail, "该邮箱已注册")
		case errors.Is(err, services.ErrDuplicateUsername):
			return helpers.Error(ctx, http.StatusConflict, consts.CodeDuplicateUsername, "用户名已被使用")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to sign up")
		}
	}

	setRefreshCookie(ctx, result.RefreshToken)
	return helpers.Success(ctx, map[string]any{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          user,
	})
}

// SignIn authenticates a user via email+code or account+password.
func (c *AuthController) SignIn(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SignInRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的请求")
	}

	var (
		result *services.AuthResult
		user   *models.User
		err    error
	)

	if req.Email != "" {
		result, user, err = services.SignInByEmail(ctx, req.Email, req.Code)
	} else if req.Account != "" {
		result, user, err = services.SignInByAccount(ctx, req.Account, req.Password)
	} else {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "请输入邮箱或账号")
	}

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCode):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidCode, "验证码无效或已过期")
		case errors.Is(err, services.ErrUserNotFound):
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "用户不存在")
		case errors.Is(err, services.ErrInvalidPassword):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidPassword, "密码错误")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to sign in")
		}
	}

	// Record login asynchronously
	ip := ctx.Request().Ip()
	userAgent := ctx.Request().Header("User-Agent", "")
	go services.RecordLogin(user.ID, ip, userAgent)

	setRefreshCookie(ctx, result.RefreshToken)
	return helpers.Success(ctx, map[string]any{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          user,
	})
}

// Refresh validates an opaque refresh token, issues a new JWT, and rotates the refresh token.
func (c *AuthController) Refresh(ctx contractshttp.Context) contractshttp.Response {
	ip := ctx.Request().Ip()
	allowed, err := helpers.CheckRateLimit(fmt.Sprintf("rate:refresh:%s", ip), 10, 60)
	if err != nil || !allowed {
		return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "刷新请求过于频繁")
	}

	oldToken := getRefreshToken(ctx)
	if oldToken == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeInvalidRefreshToken, "refresh token required")
	}

	result, err := services.RefreshToken(ctx, oldToken)
	if err != nil {
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			clearRefreshCookie(ctx)
			return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeInvalidRefreshToken, "invalid or expired refresh token")
		}
		if errors.Is(err, services.ErrSessionReplaced) {
			clearRefreshCookie(ctx)
			return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeSessionReplaced, "您的账号已在其他设备登录")
		}
		if errors.Is(err, services.ErrUserNotFound) {
			clearRefreshCookie(ctx)
			return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeInvalidRefreshToken, "invalid or expired refresh token")
		}
		clearRefreshCookie(ctx)
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to refresh token")
	}

	setRefreshCookie(ctx, result.RefreshToken)
	return helpers.Success(ctx, map[string]any{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

// Me returns the current authenticated user's profile.
func (c *AuthController) Me(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	user, err := services.GetCurrentUser(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "用户不存在")
	}

	return helpers.Success(ctx, user)
}

// Logout deletes the refresh token from Redis and clears the cookie.
func (c *AuthController) Logout(ctx contractshttp.Context) contractshttp.Response {
	token := getRefreshToken(ctx)
	if token != "" {
		_ = services.Logout(token)
	}

	clearRefreshCookie(ctx)
	return helpers.Success(ctx, nil)
}

// setRefreshCookie sets the refresh token as an httpOnly cookie.
func setRefreshCookie(ctx contractshttp.Context, token string) {
	secure := facades.Config().GetBool("refresh_token.cookie_secure", true)
	ttl := facades.Config().GetInt("refresh_token.ttl", 10080)
	ctx.Response().Cookie(contractshttp.Cookie{
		Name:     "dx_refresh",
		Value:    token,
		Path:     "/",
		MaxAge:   ttl * 60,
		Secure:   secure,
		HttpOnly: true,
		SameSite: "Lax",
	})
}

// clearRefreshCookie clears the refresh token cookie.
func clearRefreshCookie(ctx contractshttp.Context) {
	ctx.Response().Cookie(contractshttp.Cookie{
		Name:     "dx_refresh",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   facades.Config().GetBool("refresh_token.cookie_secure", true),
		HttpOnly: true,
		SameSite: "Lax",
	})
}

// getRefreshToken reads the refresh token from cookie first, then request body.
func getRefreshToken(ctx contractshttp.Context) string {
	if token := ctx.Request().Cookie("dx_refresh"); token != "" {
		return token
	}
	return ctx.Request().Input("refresh_token")
}

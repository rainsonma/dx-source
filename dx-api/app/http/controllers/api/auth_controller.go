package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/http/middleware"
	requests "dx-api/app/http/requests/api"
	"dx-api/app/models"
	services "dx-api/app/services/api"
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

	token, user, err := services.SignUp(ctx, req.Email, req.Code, req.Username, req.Password)
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

	middleware.SetTokenCookie(ctx, "dx_token", token)
	return helpers.Success(ctx, map[string]any{"user": user})
}

// SignIn authenticates a user via email+code or account+password.
func (c *AuthController) SignIn(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SignInRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的请求")
	}

	var (
		token string
		user  *models.User
		err   error
	)

	if req.Email != "" {
		token, user, err = services.SignInByEmail(ctx, req.Email, req.Code)
	} else if req.Account != "" {
		token, user, err = services.SignInByAccount(ctx, req.Account, req.Password)
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
	go services.RecordLogin(user.ID, ip, userAgent, consts.PlatformWebsite)

	middleware.SetTokenCookie(ctx, "dx_token", token)
	return helpers.Success(ctx, map[string]any{"user": user})
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

// Logout blacklists the JWT and clears the cookie.
func (c *AuthController) Logout(ctx contractshttp.Context) contractshttp.Response {
	// Parse token from cookie so Goravel can blacklist it
	token := ctx.Request().Cookie("dx_token")
	if token != "" {
		_, _ = facades.Auth(ctx).Guard("user").Parse(token)
		services.Logout(ctx)
	}

	middleware.ClearTokenCookie(ctx, "dx_token")
	return helpers.Success(ctx, nil)
}

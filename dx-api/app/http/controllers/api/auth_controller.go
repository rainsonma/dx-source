package api

import (
	"errors"
	"fmt"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"github.com/goravel/framework/facades"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	"dx-api/app/models"
	services "dx-api/app/services/api"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

// SendSignUpCode sends a verification code for signup.
func (c *AuthController) SendSignUpCode(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SendCodeRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid request")
	}

	if req.Email == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidEmail, "email is required")
	}

	if err := services.SendSignUpCode(req.Email); err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "please wait before requesting another code")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeEmailSendError, "failed to send verification code")
	}

	return helpers.Success(ctx, nil)
}

// SignUp registers a new user.
func (c *AuthController) SignUp(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SignUpRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid request")
	}

	if req.Email == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidEmail, "email is required")
	}
	if req.Code == "" || len(req.Code) != 6 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidCode, "a 6-digit verification code is required")
	}

	result, user, err := services.SignUp(ctx, req.Email, req.Code, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCode):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidCode, "invalid or expired verification code")
		case errors.Is(err, services.ErrDuplicateEmail):
			return helpers.Error(ctx, http.StatusConflict, consts.CodeDuplicateEmail, "email already registered")
		case errors.Is(err, services.ErrDuplicateUsername):
			return helpers.Error(ctx, http.StatusConflict, consts.CodeDuplicateUsername, "username already taken")
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

// SendSignInCode sends a verification code for signin.
func (c *AuthController) SendSignInCode(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SendCodeRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid request")
	}

	if req.Email == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidEmail, "email is required")
	}

	if err := services.SendSignInCode(req.Email); err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "please wait before requesting another code")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeEmailSendError, "failed to send verification code")
	}

	return helpers.Success(ctx, nil)
}

// SignIn authenticates a user via email+code or account+password.
func (c *AuthController) SignIn(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SignInRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid request")
	}

	var (
		result *services.AuthResult
		user   *models.User
		err    error
	)

	if req.Email != "" && req.Code != "" {
		// Email + code flow
		result, user, err = services.SignInByEmail(ctx, req.Email, req.Code)
	} else if req.Account != "" && req.Password != "" {
		// Account + password flow
		result, user, err = services.SignInByAccount(ctx, req.Account, req.Password)
	} else {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "provide email+code or account+password")
	}

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCode):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidCode, "invalid or expired verification code")
		case errors.Is(err, services.ErrUserNotFound):
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "user not found")
		case errors.Is(err, services.ErrInvalidPassword):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidPassword, "invalid password")
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
		return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "too many refresh requests")
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
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "user not found")
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

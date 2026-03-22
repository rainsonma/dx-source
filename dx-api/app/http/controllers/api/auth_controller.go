package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/constants"
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
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid request")
	}

	if req.Email == "" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeInvalidEmail, "email is required")
	}

	if err := services.SendSignUpCode(req.Email); err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, constants.CodeRateLimited, "please wait before requesting another code")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeEmailSendError, "failed to send verification code")
	}

	return helpers.Success(ctx, nil)
}

// SignUp registers a new user.
func (c *AuthController) SignUp(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SignUpRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid request")
	}

	if req.Email == "" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeInvalidEmail, "email is required")
	}
	if req.Code == "" || len(req.Code) != 6 {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeInvalidCode, "a 6-digit verification code is required")
	}

	token, user, err := services.SignUp(ctx, req.Email, req.Code, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCode):
			return helpers.Error(ctx, http.StatusBadRequest, constants.CodeInvalidCode, "invalid or expired verification code")
		case errors.Is(err, services.ErrDuplicateEmail):
			return helpers.Error(ctx, http.StatusConflict, constants.CodeDuplicateEmail, "email already registered")
		case errors.Is(err, services.ErrDuplicateUsername):
			return helpers.Error(ctx, http.StatusConflict, constants.CodeDuplicateUsername, "username already taken")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to sign up")
		}
	}

	return helpers.Success(ctx, map[string]any{
		"token": token,
		"user":  user,
	})
}

// SendSignInCode sends a verification code for signin.
func (c *AuthController) SendSignInCode(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SendCodeRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid request")
	}

	if req.Email == "" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeInvalidEmail, "email is required")
	}

	if err := services.SendSignInCode(req.Email); err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, constants.CodeRateLimited, "please wait before requesting another code")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeEmailSendError, "failed to send verification code")
	}

	return helpers.Success(ctx, nil)
}

// SignIn authenticates a user via email+code or account+password.
func (c *AuthController) SignIn(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SignInRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid request")
	}

	var (
		token string
		user  *models.User
		err   error
	)

	if req.Email != "" && req.Code != "" {
		// Email + code flow
		token, user, err = services.SignInByEmail(ctx, req.Email, req.Code)
	} else if req.Account != "" && req.Password != "" {
		// Account + password flow
		token, user, err = services.SignInByAccount(ctx, req.Account, req.Password)
	} else {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "provide email+code or account+password")
	}

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCode):
			return helpers.Error(ctx, http.StatusBadRequest, constants.CodeInvalidCode, "invalid or expired verification code")
		case errors.Is(err, services.ErrUserNotFound):
			return helpers.Error(ctx, http.StatusNotFound, constants.CodeUserNotFound, "user not found")
		case errors.Is(err, services.ErrInvalidPassword):
			return helpers.Error(ctx, http.StatusBadRequest, constants.CodeInvalidPassword, "invalid password")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to sign in")
		}
	}

	// Record login asynchronously
	ip := ctx.Request().Ip()
	userAgent := ctx.Request().Header("User-Agent", "")
	go services.RecordLogin(user.ID, ip, userAgent)

	return helpers.Success(ctx, map[string]any{
		"token": token,
		"user":  user,
	})
}

// Refresh refreshes the JWT token for the authenticated user.
func (c *AuthController) Refresh(ctx contractshttp.Context) contractshttp.Response {
	token, err := services.RefreshToken(ctx)
	if err != nil {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "failed to refresh token")
	}

	return helpers.Success(ctx, map[string]any{
		"token": token,
	})
}

// Me returns the current authenticated user's profile.
func (c *AuthController) Me(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	user, err := services.GetCurrentUser(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusNotFound, constants.CodeUserNotFound, "user not found")
	}

	return helpers.Success(ctx, user)
}

// Logout logs the current user out by invalidating the JWT token.
func (c *AuthController) Logout(ctx contractshttp.Context) contractshttp.Response {
	if err := facades.Auth(ctx).Guard("user").Logout(); err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to logout")
	}

	return helpers.Success(ctx, nil)
}

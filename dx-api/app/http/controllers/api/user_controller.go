package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/constants"
	"github.com/goravel/framework/facades"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

// GetProfile returns the authenticated user's full profile.
func (c *UserController) GetProfile(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	profile, err := services.GetProfile(userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, constants.CodeUserNotFound, "user not found")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to get profile")
	}

	return helpers.Success(ctx, profile)
}

// UpdateProfile updates the authenticated user's nickname, city, and introduction.
func (c *UserController) UpdateProfile(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	var req requests.UpdateProfileRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid request")
	}

	if req.Nickname != "" && len(req.Nickname) > 20 {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "nickname must be at most 20 characters")
	}
	if req.City != "" && len(req.City) > 50 {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "city must be at most 50 characters")
	}
	if req.Introduction != "" && len(req.Introduction) > 200 {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "introduction must be at most 200 characters")
	}

	if err := services.UpdateProfile(userID, req.Nickname, req.City, req.Introduction); err != nil {
		if errors.Is(err, services.ErrNicknameTaken) {
			return helpers.Error(ctx, http.StatusConflict, constants.CodeNicknameTaken, "nickname already taken")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to update profile")
	}

	return helpers.Success(ctx, nil)
}

// UpdateAvatar sets the authenticated user's avatar from an image ID.
func (c *UserController) UpdateAvatar(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	var req requests.UpdateAvatarRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid request")
	}

	if req.ImageID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "image_id is required")
	}

	if err := services.UpdateAvatar(userID, req.ImageID); err != nil {
		switch {
		case errors.Is(err, services.ErrImageNotFound):
			return helpers.Error(ctx, http.StatusNotFound, constants.CodeImageNotFound, "image not found")
		case errors.Is(err, services.ErrImageNotOwned):
			return helpers.Error(ctx, http.StatusForbidden, constants.CodeForbidden, "image does not belong to you")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to update avatar")
		}
	}

	return helpers.Success(ctx, nil)
}

// SendEmailCode sends a verification code for changing email.
func (c *UserController) SendEmailCode(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	var req requests.SendEmailCodeRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid request")
	}

	if req.Email == "" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeInvalidEmail, "email is required")
	}

	if err := services.SendChangeEmailCode(userID, req.Email); err != nil {
		switch {
		case errors.Is(err, services.ErrRateLimited):
			return helpers.Error(ctx, http.StatusTooManyRequests, constants.CodeRateLimited, "please wait before requesting another code")
		case errors.Is(err, services.ErrDuplicateEmail):
			return helpers.Error(ctx, http.StatusConflict, constants.CodeDuplicateEmail, "email already registered")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeEmailSendError, "failed to send verification code")
		}
	}

	return helpers.Success(ctx, nil)
}

// ChangeEmail changes the authenticated user's email with a verification code.
func (c *UserController) ChangeEmail(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	var req requests.ChangeEmailRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid request")
	}

	if req.Email == "" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeInvalidEmail, "email is required")
	}
	if req.Code == "" || len(req.Code) != 6 {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeInvalidCode, "a 6-digit verification code is required")
	}

	if err := services.ChangeEmail(userID, req.Email, req.Code); err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCode):
			return helpers.Error(ctx, http.StatusBadRequest, constants.CodeInvalidCode, "invalid or expired verification code")
		case errors.Is(err, services.ErrDuplicateEmail):
			return helpers.Error(ctx, http.StatusConflict, constants.CodeDuplicateEmail, "email already registered")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to change email")
		}
	}

	return helpers.Success(ctx, nil)
}

// ChangePassword changes the authenticated user's password.
func (c *UserController) ChangePassword(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthorized")
	}

	var req requests.ChangePasswordRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "invalid request")
	}

	if req.CurrentPassword == "" {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "current_password is required")
	}
	if req.NewPassword == "" || len(req.NewPassword) < 8 {
		return helpers.Error(ctx, http.StatusBadRequest, constants.CodeValidationError, "new_password must be at least 8 characters")
	}

	if err := services.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, services.ErrInvalidPassword) {
			return helpers.Error(ctx, http.StatusBadRequest, constants.CodeInvalidPassword, "current password is incorrect")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, constants.CodeInternalError, "failed to change password")
	}

	return helpers.Success(ctx, nil)
}

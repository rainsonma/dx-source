package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
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
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	profile, err := services.GetProfile(userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeUserNotFound, "用户不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get profile")
	}

	return helpers.Success(ctx, profile)
}

// UpdateProfile updates the authenticated user's nickname, city, and introduction.
func (c *UserController) UpdateProfile(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.UpdateProfileRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.UpdateProfile(userID, req.Nickname, req.City, req.Introduction); err != nil {
		if errors.Is(err, services.ErrNicknameTaken) {
			return helpers.Error(ctx, http.StatusConflict, consts.CodeNicknameTaken, "昵称已被使用")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to update profile")
	}

	return helpers.Success(ctx, nil)
}

// UpdateAvatar sets the authenticated user's avatar URL.
func (c *UserController) UpdateAvatar(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.UpdateAvatarRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if !helpers.IsUploadedImageURL(req.AvatarURL) {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的头像URL")
	}

	if err := services.UpdateAvatar(userID, req.AvatarURL); err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to update avatar")
	}

	return helpers.Success(ctx, nil)
}

// ChangeEmail changes the authenticated user's email with a verification code.
func (c *UserController) ChangeEmail(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.ChangeEmailRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.ChangeEmail(userID, req.Email, req.Code); err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCode):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidCode, "验证码无效或已过期")
		case errors.Is(err, services.ErrDuplicateEmail):
			return helpers.Error(ctx, http.StatusConflict, consts.CodeDuplicateEmail, "该邮箱已注册")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to change email")
		}
	}

	return helpers.Success(ctx, nil)
}

// ChangePassword changes the authenticated user's password.
func (c *UserController) ChangePassword(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.ChangePasswordRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, services.ErrInvalidPassword) {
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidPassword, "当前密码错误")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to change password")
	}

	return helpers.Success(ctx, nil)
}

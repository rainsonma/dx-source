package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type EmailController struct{}

func NewEmailController() *EmailController {
	return &EmailController{}
}

// SendSignUpCode sends a verification code for signup.
func (c *EmailController) SendSignUpCode(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SendCodeRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.SendSignUpCode(req.Email); err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "请稍后再请求验证码")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeEmailSendError, "failed to send verification code")
	}

	return helpers.Success(ctx, nil)
}

// SendSignInCode sends a verification code for signin.
func (c *EmailController) SendSignInCode(ctx contractshttp.Context) contractshttp.Response {
	var req requests.SendCodeRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.SendSignInCode(req.Email); err != nil {
		if errors.Is(err, services.ErrRateLimited) {
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "请稍后再请求验证码")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeEmailSendError, "failed to send verification code")
	}

	return helpers.Success(ctx, nil)
}

// SendChangeCode sends a verification code for changing email.
func (c *EmailController) SendChangeCode(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.SendCodeRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	if err := services.SendChangeEmailCode(userID, req.Email); err != nil {
		switch {
		case errors.Is(err, services.ErrRateLimited):
			return helpers.Error(ctx, http.StatusTooManyRequests, consts.CodeRateLimited, "请稍后再请求验证码")
		case errors.Is(err, services.ErrDuplicateEmail):
			return helpers.Error(ctx, http.StatusConflict, consts.CodeDuplicateEmail, "该邮箱已注册")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeEmailSendError, "failed to send verification code")
		}
	}

	return helpers.Success(ctx, nil)
}

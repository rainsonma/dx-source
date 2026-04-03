package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type PaymentController struct{}

func NewPaymentController() *PaymentController {
	return &PaymentController{}
}

func (c *PaymentController) GetOrder(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	orderID := ctx.Request().Route("id")
	if orderID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "missing order id")
	}

	order, err := services.GetOrder(orderID, userID)
	if err != nil {
		if errors.Is(err, services.ErrOrderNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeOrderNotFound, "订单不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get order")
	}

	return helpers.Success(ctx, order)
}

func (c *PaymentController) WechatCallback(ctx contractshttp.Context) contractshttp.Response {
	return helpers.Success(ctx, nil)
}

func (c *PaymentController) AlipayCallback(ctx contractshttp.Context) contractshttp.Response {
	return helpers.Success(ctx, nil)
}

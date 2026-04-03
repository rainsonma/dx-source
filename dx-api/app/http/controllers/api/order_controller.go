package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type OrderController struct{}

func NewOrderController() *OrderController {
	return &OrderController{}
}

func (c *OrderController) CreateMembershipOrder(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CreateMembershipOrderRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	price, ok := consts.UserGradePrices[req.Grade]
	if !ok || price == 0 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidProduct, "无效的会员等级")
	}

	amountFen := price * 100

	order, err := services.CreateOrder(userID, consts.OrderTypeMembership, req.Grade, amountFen, req.PaymentMethod)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to create order")
	}

	return helpers.Success(ctx, order)
}

func (c *OrderController) CreateBeansOrder(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CreateBeansOrderRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	pkg, ok := consts.BeanPackages[req.Package]
	if !ok {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidProduct, "无效的能量豆套餐")
	}

	order, err := services.CreateOrder(userID, consts.OrderTypeBeans, req.Package, pkg.Price, req.PaymentMethod)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to create order")
	}

	return helpers.Success(ctx, order)
}

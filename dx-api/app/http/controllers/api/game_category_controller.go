package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type GameCategoryController struct{}

func NewGameCategoryController() *GameCategoryController {
	return &GameCategoryController{}
}

// Categories returns all enabled categories in hierarchical order.
func (c *GameCategoryController) Categories(ctx contractshttp.Context) contractshttp.Response {
	categories, err := services.ListCategories()
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list categories")
	}

	return helpers.Success(ctx, categories)
}

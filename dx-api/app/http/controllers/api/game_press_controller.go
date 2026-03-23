package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type GamePressController struct{}

func NewGamePressController() *GamePressController {
	return &GamePressController{}
}

// Presses returns all game publishers.
func (c *GamePressController) Presses(ctx contractshttp.Context) contractshttp.Response {
	presses, err := services.ListPresses()
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list presses")
	}

	return helpers.Success(ctx, presses)
}

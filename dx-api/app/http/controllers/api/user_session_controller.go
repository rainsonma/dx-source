package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type UserSessionController struct{}

func NewUserSessionController() *UserSessionController {
	return &UserSessionController{}
}

// ListSessions returns the user's recent solo-mode game session progress (up to 20, ordered by last played).
func (c *UserSessionController) ListSessions(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	rows, err := services.ListSessionProgress(userID)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list sessions")
	}
	return helpers.Success(ctx, rows)
}

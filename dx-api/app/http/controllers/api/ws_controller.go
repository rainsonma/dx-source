package api

import (
	"context"
	"net/http"
	"strings"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	gorillaWs "github.com/gorilla/websocket"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/realtime"
)

// upgrader handles HTTP→WebSocket upgrades via gorilla/websocket, which is
// the officially supported WebSocket library for Goravel (see goravel/example).
var upgrader = gorillaWs.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		originsRaw, ok := facades.Config().Env("CORS_ALLOWED_ORIGINS", "http://localhost:3000").(string)
		if !ok {
			return false
		}
		for _, allowed := range strings.Split(originsRaw, ",") {
			if strings.TrimSpace(allowed) == origin {
				return true
			}
		}
		return false
	},
}

type WSController struct{}

func NewWSController() *WSController {
	return &WSController{}
}

func (c *WSController) Handle(ctx contractshttp.Context) contractshttp.Response {
	hub := realtime.DefaultHub()
	if hub != nil && hub.IsShuttingDown() {
		return helpers.Error(ctx, http.StatusServiceUnavailable, consts.CodeInternalError, "server shutting down")
	}

	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	if hub == nil {
		return helpers.Error(ctx, http.StatusServiceUnavailable, consts.CodeInternalError, "realtime hub not initialized")
	}

	w := ctx.Response().Writer()
	r := ctx.Request().Origin()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}
	defer conn.Close()

	wsCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = hub.Attach(wsCtx, userID, conn)
	return nil
}

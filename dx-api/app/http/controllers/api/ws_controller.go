package api

import (
	"context"
	"net/http"
	"strings"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"github.com/coder/websocket"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/realtime"
)

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

	originsRaw := facades.Config().Env("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	originPatterns := strings.Split(originsRaw.(string), ",")
	for i := range originPatterns {
		originPatterns[i] = strings.TrimSpace(originPatterns[i])
	}

	w := ctx.Response().Writer()
	r := ctx.Request().Origin()

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns:  originPatterns,
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		return nil
	}
	defer conn.Close(websocket.StatusInternalError, "server error")

	// CRITICAL: detached context. Goravel's global Timeout middleware
	// cancels the request context after http.request_timeout (30s default).
	// Using that context would kill the WebSocket after 30s. A background
	// context is immune; shutdown is driven by Hub.Shutdown.
	wsCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = hub.Attach(wsCtx, userID, conn)
	return nil
}

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

	// Block in Hub.Attach for the lifetime of the WebSocket connection.
	// This handler runs inside the Timeout middleware's goroutine, which
	// has a context.WithTimeout of http.request_timeout (set to 24h in
	// config/http.go). We use a detached context so the WS read/write
	// loops are immune to that timeout's context cancellation. The 24h
	// value ensures the middleware never fires Abort(408) which would
	// write HTTP bytes to the hijacked connection and corrupt the WS
	// frame stream.
	wsCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = hub.Attach(wsCtx, userID, conn)
	return nil
}

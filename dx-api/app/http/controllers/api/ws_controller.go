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

	// CRITICAL: spawn Hub.Attach in a goroutine so the HTTP handler returns
	// immediately. This is necessary because Goravel's global Timeout
	// middleware (goravel/gin middleware_timeout.go) wraps every handler in a
	// goroutine and races it against http.request_timeout (30s). If our
	// handler blocks in Attach, after 30s the middleware calls Abort(408)
	// which writes HTTP bytes to the hijacked WebSocket connection —
	// corrupting the frame stream and causing "Invalid frame header" errors.
	//
	// By returning nil immediately, the middleware sees the handler as
	// "done" and never fires the timeout. The WebSocket is fully owned by
	// the goroutine below via a detached context (immune to the middleware's
	// context cancellation). Shutdown is driven by Hub.Shutdown / the
	// RealtimeRunner.
	wsCtx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		defer conn.Close(websocket.StatusInternalError, "server error")
		_ = hub.Attach(wsCtx, userID, conn)
	}()

	return nil
}

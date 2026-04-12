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

// upgrader is the gorilla/websocket upgrader. Using gorilla instead of
// coder/websocket for the HTTP→WS upgrade because gorilla calls Hijack()
// BEFORE writing the 101 response — bypassing Gin's ResponseWriter which
// rejects Hijack after any WriteHeader call. Once upgraded, the raw
// net.Conn is handed to coder/websocket for frame-level I/O.
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

	// gorilla/websocket.Upgrader.Upgrade calls Hijack() FIRST, then writes
	// the 101 Switching Protocols response directly to the raw net.Conn.
	// This bypasses Gin's ResponseWriter entirely, avoiding the
	// "response already written" error that coder/websocket's Accept hits.
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

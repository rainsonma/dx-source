package api

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"github.com/coder/websocket"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/realtime"
)

// ginHijackBypass wraps Gin's ResponseWriter to bypass its "response already
// written" check on Hijack(). coder/websocket's Accept calls WriteHeader(101)
// then Hijack(), but Gin's responseWriter rejects Hijack after any WriteHeader
// call. This wrapper unwraps all ResponseWriter layers (Goravel → Gin →
// net/http) and delegates Hijack to the innermost writer which doesn't have
// that restriction.
type ginHijackBypass struct {
	http.ResponseWriter
}

func (w *ginHijackBypass) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	rw := w.ResponseWriter
	for {
		u, ok := rw.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			break
		}
		rw = u.Unwrap()
	}
	hj, ok := rw.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("innermost ResponseWriter (%T) does not implement http.Hijacker", rw)
	}
	return hj.Hijack()
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

	originsRaw := facades.Config().Env("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	originPatterns := strings.Split(originsRaw.(string), ",")
	for i := range originPatterns {
		originPatterns[i] = strings.TrimSpace(originPatterns[i])
	}

	r := ctx.Request().Origin()
	w := &ginHijackBypass{ctx.Response().Writer()}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns:  originPatterns,
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		return nil
	}
	defer conn.Close(websocket.StatusInternalError, "server error")

	wsCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = hub.Attach(wsCtx, userID, conn)
	return nil
}

package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

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
		// Allow loopback on any port — covers local dev and the WeChat DevTools
		// proxy, whose port rotates per session (e.g. http://127.0.0.1:21618).
		if origin == "http://localhost" || strings.HasPrefix(origin, "http://localhost:") ||
			origin == "http://127.0.0.1" || strings.HasPrefix(origin, "http://127.0.0.1:") {
			return true
		}
		// Allow WeChat Mini Program real-device origins.
		if origin == "https://servicewechat.com" || strings.HasPrefix(origin, "https://servicewechat.com/") {
			return true
		}
		originsRaw, ok := facades.Config().Env("CORS_ALLOWED_ORIGINS", "http://localhost:3000").(string)
		if !ok {
			return false
		}
		for allowed := range strings.SplitSeq(originsRaw, ",") {
			if strings.TrimSpace(allowed) == origin {
				return true
			}
		}
		return false
	},
}

// wsAuthDeadline bounds how long an unauthenticated connection may live
// before it must produce a valid auth envelope as its first frame.
const wsAuthDeadline = 10 * time.Second

type WSController struct{}

func NewWSController() *WSController {
	return &WSController{}
}

// Handle accepts the WebSocket upgrade and runs an auth handshake on the
// first frame. The route is registered publicly (no JWT middleware) because
// WeChat Mini Program's wx.connectSocket cannot forward custom headers or
// cookies on the upgrade request; the token therefore arrives inside the
// first post-upgrade message as `{"type":"auth","token":"..."}`.
//
// Flow:
//  1. Upgrade HTTP → WebSocket.
//  2. Read first frame with a 10s deadline.
//  3. Parse the JWT + run the same Redis single-device check the HTTP
//     middleware uses. Emit auth_failed / session_replaced and close on
//     failure.
//  4. Emit auth_success and hand off to Hub.Attach for the normal session
//     lifecycle.
func (c *WSController) Handle(ctx contractshttp.Context) contractshttp.Response {
	hub := realtime.DefaultHub()
	if hub == nil || hub.IsShuttingDown() {
		return helpers.Error(ctx, http.StatusServiceUnavailable, consts.CodeInternalError, "realtime hub not available")
	}

	w := ctx.Response().Writer()
	r := ctx.Request().Origin()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}

	userID, authErr := authenticateWebSocket(ctx, conn)
	if authErr != nil {
		_ = conn.Close()
		return nil
	}

	wsCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer conn.Close()

	_ = hub.Attach(wsCtx, userID, conn)
	return nil
}

// authEnvelope is the shape of the first frame a newly-connected client
// MUST send. The server closes the connection if any other frame shape
// arrives first.
type authEnvelope struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// authenticateWebSocket reads exactly one frame within wsAuthDeadline,
// validates it as an auth envelope, runs the JWT + Redis checks, and
// writes an auth_success / auth_failed / session_replaced event back to
// the client. Returns the authenticated user ID on success.
//
// On any failure, the caller must close the connection. This function
// does NOT close it directly so the caller can control the defer chain.
func authenticateWebSocket(ctx contractshttp.Context, conn *gorillaWs.Conn) (string, error) {
	if err := conn.SetReadDeadline(time.Now().Add(wsAuthDeadline)); err != nil {
		return "", err
	}
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }()

	_, raw, err := conn.ReadMessage()
	if err != nil {
		return "", errors.New("auth read failed: " + err.Error())
	}

	var msg authEnvelope
	if err := json.Unmarshal(raw, &msg); err != nil || msg.Type != "auth" || msg.Token == "" {
		_ = conn.WriteJSON(map[string]string{"event": "auth_failed", "message": "first message must be auth"})
		return "", errors.New("bad auth envelope")
	}

	payload, err := facades.Auth(ctx).Guard("user").Parse(msg.Token)
	if err != nil || payload == nil || payload.Key == "" {
		_ = conn.WriteJSON(map[string]string{"event": "auth_failed", "message": "invalid token"})
		return "", errors.New("token parse failed")
	}
	userID := payload.Key

	loginTsStr, redisErr := helpers.RedisGet("user_auth:" + userID + ":user")
	if redisErr != nil {
		_ = conn.WriteJSON(map[string]string{"event": "auth_failed", "message": "session lookup failed"})
		return "", redisErr
	}
	loginTs, _ := strconv.ParseInt(loginTsStr, 10, 64)
	if payload.IssuedAt.Unix() < loginTs {
		_ = conn.WriteJSON(map[string]string{"event": "session_replaced", "code": strconv.Itoa(consts.CodeSessionReplaced)})
		return "", errors.New("session replaced")
	}

	if err := conn.WriteJSON(map[string]string{"event": "auth_success"}); err != nil {
		return "", err
	}
	return userID, nil
}

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

// Close codes for WebSocket auth outcomes. dx-web's provider already knows
// about these (see websocket-provider.tsx onclose handler):
//
//	4001 session replaced
//	4401 auth expired / invalid
const (
	wsCloseSessionReplaced = 4001
	wsCloseAuthFailed      = 4401
)

// wsAuthDeadline bounds how long an unauthenticated connection may live
// before it must produce a valid auth envelope as its first frame.
const wsAuthDeadline = 10 * time.Second

type WSController struct{}

func NewWSController() *WSController {
	return &WSController{}
}

// Handle accepts the WebSocket upgrade and authenticates the session via
// one of two paths:
//
//  1. HTTP-level auth (dx-web): the dx_token cookie or Authorization: Bearer
//     header is present on the upgrade request. Validated before calling
//     Hub.Attach; on failure, the connection is closed with close code 4001
//     or 4401 matching dx-web's existing onclose handler.
//
//  2. First-frame auth (WeChat Mini Program): no cookie, no header. The
//     controller waits up to 10 s for the first frame, which MUST be
//     {"op":"auth","token":"..."}. Server emits {"event":"auth_success"} on
//     success or {"event":"auth_failed"} then closes on failure.
//
// The route is registered publicly (no JWT middleware) because WeChat
// Mini Program's wx.connectSocket cannot forward custom headers or cookies
// on the upgrade request.
func (c *WSController) Handle(ctx contractshttp.Context) contractshttp.Response {
	hub := realtime.DefaultHub()
	if hub == nil || hub.IsShuttingDown() {
		return helpers.Error(ctx, http.StatusServiceUnavailable, consts.CodeInternalError, "realtime hub not available")
	}

	// Path 1: check for HTTP-level credentials (dx-web's cookie).
	httpToken := ctx.Request().Cookie("dx_token")
	if httpToken == "" {
		if bearer := ctx.Request().Header("Authorization", ""); len(bearer) > 7 && bearer[:7] == "Bearer " {
			httpToken = bearer[7:]
		}
	}

	w := ctx.Response().Writer()
	r := ctx.Request().Origin()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil
	}

	var userID string
	if httpToken != "" {
		// Cookie / header present — validate now (before taking the connection over).
		uid, closeCode, authErr := validateWSToken(ctx, httpToken)
		if authErr != nil {
			_ = conn.WriteMessage(gorillaWs.CloseMessage, gorillaWs.FormatCloseMessage(closeCode, authErr.Error()))
			_ = conn.Close()
			return nil
		}
		userID = uid
	} else {
		// No HTTP credentials — fall back to first-frame auth.
		uid, authErr := authenticateWebSocketByFirstFrame(ctx, conn)
		if authErr != nil {
			_ = conn.Close()
			return nil
		}
		userID = uid
	}

	wsCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer conn.Close()

	_ = hub.Attach(wsCtx, userID, conn)
	return nil
}

// validateWSToken runs the JWT parse + Redis single-device check used by the
// HTTP JwtAuth middleware, but returns values instead of writing to ctx. The
// second return value is the WebSocket close code the caller should use if
// auth fails (4001 session_replaced, 4401 otherwise, 0 on success).
func validateWSToken(ctx contractshttp.Context, token string) (string, int, error) {
	payload, err := facades.Auth(ctx).Guard("user").Parse(token)
	if err != nil || payload == nil || payload.Key == "" {
		return "", wsCloseAuthFailed, errors.New("invalid token")
	}
	userID := payload.Key

	loginTsStr, redisErr := helpers.RedisGet("user_auth:" + userID + ":user")
	if redisErr != nil {
		return "", wsCloseAuthFailed, errors.New("session lookup failed")
	}
	loginTs, _ := strconv.ParseInt(loginTsStr, 10, 64)
	if payload.IssuedAt.Unix() < loginTs {
		return "", wsCloseSessionReplaced, errors.New("session replaced")
	}
	return userID, 0, nil
}

// authEnvelope is the shape a newly-connected client MUST send as its first
// frame when no HTTP-level credentials are present. Matches the main realtime
// Envelope's `op` field name for protocol consistency.
type authEnvelope struct {
	Op    string `json:"op"`
	Token string `json:"token"`
}

// authenticateWebSocketByFirstFrame reads exactly one frame within
// wsAuthDeadline, validates it as {"op":"auth","token":"..."}, runs the JWT
// + Redis check, and writes an auth_success / auth_failed / session_replaced
// event back to the client. On failure the caller must close the connection.
func authenticateWebSocketByFirstFrame(ctx contractshttp.Context, conn *gorillaWs.Conn) (string, error) {
	if err := conn.SetReadDeadline(time.Now().Add(wsAuthDeadline)); err != nil {
		return "", err
	}
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }()

	_, raw, err := conn.ReadMessage()
	if err != nil {
		return "", errors.New("auth read failed: " + err.Error())
	}

	var msg authEnvelope
	if err := json.Unmarshal(raw, &msg); err != nil || msg.Op != "auth" || msg.Token == "" {
		_ = conn.WriteJSON(map[string]any{"op": "event", "type": "auth_failed", "data": map[string]string{"message": "first message must be auth"}})
		return "", errors.New("bad auth envelope")
	}

	userID, closeCode, authErr := validateWSToken(ctx, msg.Token)
	if authErr != nil {
		if closeCode == wsCloseSessionReplaced {
			_ = conn.WriteJSON(map[string]any{"op": "event", "type": "session_replaced"})
		} else {
			_ = conn.WriteJSON(map[string]any{"op": "event", "type": "auth_failed", "data": map[string]string{"message": authErr.Error()}})
		}
		_ = conn.WriteMessage(gorillaWs.CloseMessage, gorillaWs.FormatCloseMessage(closeCode, authErr.Error()))
		return "", authErr
	}

	if err := conn.WriteJSON(map[string]any{"op": "event", "type": "auth_success"}); err != nil {
		return "", err
	}
	return userID, nil
}

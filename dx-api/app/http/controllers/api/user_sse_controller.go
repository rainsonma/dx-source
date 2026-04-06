package api

import (
	"net/http"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/helpers"
)

type UserSSEController struct{}

func NewUserSSEController() *UserSSEController {
	return &UserSSEController{}
}

// Ping marks the user as online immediately (bridges gap before SSE connects).
func (c *UserSSEController) Ping(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, 0, "unauthorized")
	}
	_ = helpers.RedisSetAdd("online_users", userID)
	return helpers.Success(ctx, nil)
}

// Events establishes a persistent SSE connection for user-level events.
func (c *UserSSEController) Events(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, 0, "unauthorized")
	}

	w := ctx.Response().Writer()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	conn := helpers.UserHub.Register(userID, w)
	defer helpers.UserHub.Unregister(userID, conn)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	clientGone := ctx.Request().Origin().Context().Done()

	for {
		select {
		case <-clientGone:
			return nil
		case <-conn.Done():
			return nil
		case <-ticker.C:
			if err := conn.SendHeartbeat(); err != nil {
				return nil
			}
		}
	}
}

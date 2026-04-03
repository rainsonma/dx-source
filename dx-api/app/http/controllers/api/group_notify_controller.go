package api

import (
	nethttp "net/http"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/helpers"
	"dx-api/app/models"
)

type GroupNotifyController struct{}

func NewGroupNotifyController() *GroupNotifyController {
	return &GroupNotifyController{}
}

// Notify establishes a persistent SSE connection for group detail notifications.
func (c *GroupNotifyController) Notify(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, nethttp.StatusUnauthorized, 0, "unauthorized")
	}

	groupID := ctx.Request().Route("id")

	var member models.GameGroupMember
	if err := facades.Orm().Query().Where("game_group_id", groupID).
		Where("user_id", userID).First(&member); err != nil || member.ID == "" {
		return helpers.Error(ctx, nethttp.StatusForbidden, 0, "not a group member")
	}

	w := ctx.Response().Writer()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	if f, ok := w.(nethttp.Flusher); ok {
		f.Flush()
	}

	conn := helpers.NewSSEConnection(w)
	helpers.GroupNotifyHub.Register(groupID, userID, conn)
	defer helpers.GroupNotifyHub.Unregister(groupID, userID, conn)

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

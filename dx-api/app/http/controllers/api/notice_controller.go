package api

import (
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type NoticeController struct{}

func NewNoticeController() *NoticeController {
	return &NoticeController{}
}

// GetNotices returns active notices with cursor pagination.
func (c *NoticeController) GetNotices(ctx contractshttp.Context) contractshttp.Response {
	cursor, limit := helpers.ParseCursorParams(ctx, 20)

	items, nextCursor, hasMore, err := services.GetNotices(cursor, limit)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get notices")
	}

	return helpers.Paginated(ctx, items, nextCursor, hasMore)
}

// MarkNoticesRead updates the user's last_read_notice_at.
func (c *NoticeController) MarkNoticesRead(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	if err := services.MarkNoticesRead(userID); err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to mark notices read")
	}

	return helpers.Success(ctx, nil)
}

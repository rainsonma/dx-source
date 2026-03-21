package helpers

import (
	"strconv"

	"github.com/goravel/framework/contracts/http"
)

const (
	DefaultPageSize    = 20
	DefaultCursorLimit = 12
	MaxPageSize        = 100
)

// ParseCursorParams extracts cursor and limit from query string
func ParseCursorParams(ctx http.Context, defaultLimit int) (cursor string, limit int) {
	cursor = ctx.Request().Query("cursor", "")
	limitStr := ctx.Request().Query("limit", "")

	if defaultLimit <= 0 {
		defaultLimit = DefaultCursorLimit
	}

	limit = defaultLimit
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= MaxPageSize {
			limit = parsed
		}
	}

	return cursor, limit
}

// ParseOffsetParams extracts page, pageSize, and computed offset from query string
func ParseOffsetParams(ctx http.Context, defaultPageSize int) (page int, pageSize int, offset int) {
	if defaultPageSize <= 0 {
		defaultPageSize = DefaultPageSize
	}

	pageStr := ctx.Request().Query("page", "1")
	pageSizeStr := ctx.Request().Query("pageSize", "")

	page = 1
	if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
		page = parsed
	}

	pageSize = defaultPageSize
	if pageSizeStr != "" {
		if parsed, err := strconv.Atoi(pageSizeStr); err == nil && parsed > 0 && parsed <= MaxPageSize {
			pageSize = parsed
		}
	}

	offset = (page - 1) * pageSize
	return page, pageSize, offset
}

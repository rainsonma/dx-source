package api

import (
	"fmt"
	"time"

	"dx-api/app/facades"
	"dx-api/app/models"
)

// NoticeItem represents a notice in the list.
type NoticeItem struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Content   *string `json:"content"`
	Icon      *string `json:"icon"`
	CreatedAt any     `json:"createdAt"`
}

// GetNotices returns active notices with cursor pagination.
func GetNotices(cursor string, limit int) ([]NoticeItem, string, bool, error) {
	if limit <= 0 {
		limit = 20
	}

	query := facades.Orm().Query().
		Where("is_active", true).
		Order("created_at desc").
		Limit(limit + 1)

	if cursor != "" {
		// Keyset pagination: fetch the cursor notice's created_at, then filter
		var cursorNotice models.Notice
		if err := facades.Orm().Query().Where("id", cursor).First(&cursorNotice); err == nil && cursorNotice.ID != "" {
			query = query.Where("created_at < ?", cursorNotice.CreatedAt)
		}
	}

	var notices []models.Notice
	if err := query.Get(&notices); err != nil {
		return nil, "", false, fmt.Errorf("failed to query notices: %w", err)
	}

	hasMore := len(notices) > limit
	if hasMore {
		notices = notices[:limit]
	}

	items := make([]NoticeItem, 0, len(notices))
	var nextCursor string
	for _, n := range notices {
		items = append(items, NoticeItem{
			ID:        n.ID,
			Title:     n.Title,
			Content:   n.Content,
			Icon:      n.Icon,
			CreatedAt: n.CreatedAt,
		})
	}

	if hasMore && len(items) > 0 {
		nextCursor = items[len(items)-1].ID
	}

	return items, nextCursor, hasMore, nil
}

// MarkNoticesRead updates the user's last_read_notice_at to now.
func MarkNoticesRead(userID string) error {
	now := time.Now()
	if _, err := facades.Orm().Query().Model(&models.User{}).
		Where("id", userID).
		Update("last_read_notice_at", now); err != nil {
		return fmt.Errorf("failed to mark notices read: %w", err)
	}
	return nil
}

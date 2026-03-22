package adm

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/goravel/framework/facades"
	"dx-api/app/models"

	"github.com/oklog/ulid/v2"
)

// CreateNotice creates a new system notice.
func CreateNotice(title string, content, icon *string) (*models.Notice, error) {
	notice := models.Notice{
		ID:       ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String(),
		Title:    title,
		Content:  content,
		Icon:     icon,
		IsActive: true,
	}

	if err := facades.Orm().Query().Create(&notice); err != nil {
		return nil, fmt.Errorf("failed to create notice: %w", err)
	}

	return &notice, nil
}

// UpdateNotice updates an existing notice.
func UpdateNotice(id string, title string, content, icon *string) (*models.Notice, error) {
	var notice models.Notice
	if err := facades.Orm().Query().Where("id", id).First(&notice); err != nil || notice.ID == "" {
		return nil, ErrNoticeNotFound
	}

	notice.Title = title
	notice.Content = content
	notice.Icon = icon

	if err := facades.Orm().Query().Save(&notice); err != nil {
		return nil, fmt.Errorf("failed to update notice: %w", err)
	}

	return &notice, nil
}

// DeleteNotice soft-deletes a notice by setting is_active to false.
func DeleteNotice(id string) error {
	result, err := facades.Orm().Query().Model(&models.Notice{}).
		Where("id", id).
		Update("is_active", false)
	if err != nil {
		return fmt.Errorf("failed to delete notice: %w", err)
	}
	if result == nil {
		return ErrNoticeNotFound
	}
	return nil
}

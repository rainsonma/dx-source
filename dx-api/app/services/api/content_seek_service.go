package api

import (
	"fmt"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"
)

// ContentSeekItem represents a content seek record.
type ContentSeekItem struct {
	ID          string `json:"id"`
	CourseName  string `json:"courseName"`
	Description string `json:"description"`
	DiskUrl     string `json:"diskUrl"`
	Count       int    `json:"count"`
	CreatedAt   any    `json:"createdAt"`
}

// ContentSeekResult indicates whether the submission was a duplicate.
type ContentSeekResult struct {
	Duplicate bool `json:"duplicate"`
}

// GetContentSeeks returns the user's content seek records.
func GetContentSeeks(userID string) ([]ContentSeekItem, error) {
	var seeks []models.ContentSeek
	if err := facades.Orm().Query().
		Where("user_id", userID).
		Order("created_at desc").
		Get(&seeks); err != nil {
		return nil, fmt.Errorf("failed to query content seeks: %w", err)
	}

	items := make([]ContentSeekItem, 0, len(seeks))
	for _, s := range seeks {
		items = append(items, ContentSeekItem{
			ID:          s.ID,
			CourseName:  s.CourseName,
			Description: s.Description,
			DiskUrl:     s.DiskUrl,
			Count:       s.Count,
			CreatedAt:   s.CreatedAt,
		})
	}

	return items, nil
}

// SubmitContentSeek creates a content seek record or increments count on duplicate course name.
func SubmitContentSeek(userID, courseName, description, diskUrl string) (*ContentSeekResult, error) {
	query := facades.Orm().Query()

	// Check for existing seek with same course name by this user
	var existing models.ContentSeek
	if err := query.Where("user_id", userID).
		Where("course_name", courseName).
		First(&existing); err == nil && existing.ID != "" {
		// Duplicate: increment count
		if _, err := facades.Orm().Query().Model(&models.ContentSeek{}).
			Where("id", existing.ID).
			Update("count", existing.Count+1); err != nil {
			return nil, fmt.Errorf("failed to increment content seek count: %w", err)
		}
		return &ContentSeekResult{Duplicate: true}, nil
	}

	// Create new content seek
	seek := models.ContentSeek{
		ID:          newID(),
		UserID:      userID,
		CourseName:  courseName,
		Description: description,
		DiskUrl:     diskUrl,
		Count:       1,
	}
	if err := query.Create(&seek); err != nil {
		return nil, fmt.Errorf("failed to create content seek: %w", err)
	}

	return &ContentSeekResult{Duplicate: false}, nil
}

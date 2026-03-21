package api

import (
	"fmt"

	"dx-api/app/facades"
	"dx-api/app/helpers"
	"dx-api/app/models"
)

// FeedbackResult indicates whether the feedback was a duplicate or new.
type FeedbackResult struct {
	Duplicate bool `json:"duplicate"`
}

// SubmitFeedback creates a feedback record or increments count on duplicate.
func SubmitFeedback(userID, feedbackType, description string) (*FeedbackResult, error) {
	query := facades.Orm().Query()

	// Check for existing feedback with same type + description
	var existing models.Feedback
	if err := query.Where("user_id", userID).
		Where("type", feedbackType).
		Where("description", description).
		First(&existing); err == nil && existing.ID != "" {
		// Duplicate: increment count
		if _, err := facades.Orm().Query().Model(&models.Feedback{}).
			Where("id", existing.ID).
			Update("count", existing.Count+1); err != nil {
			return nil, fmt.Errorf("failed to increment feedback count: %w", err)
		}
		return &FeedbackResult{Duplicate: true}, nil
	}

	// Create new feedback
	feedback := models.Feedback{
		ID:          newID(),
		UserID:      userID,
		Type:        feedbackType,
		Description: description,
		Count:       1,
	}
	if err := query.Create(&feedback); err != nil {
		return nil, fmt.Errorf("failed to create feedback: %w", err)
	}

	return &FeedbackResult{Duplicate: false}, nil
}

// ReportResult indicates whether the report was a duplicate or new.
type ReportResult struct {
	ID    string `json:"id"`
	Count int    `json:"count"`
}

// SubmitReport creates a game report or increments count on duplicate.
// Rate limited to 10 requests per 60 seconds.
func SubmitReport(userID, gameID, gameLevelID, contentItemID, reason string, note *string) (*ReportResult, error) {
	// Rate limit check
	key := fmt.Sprintf("ratelimit:report:%s", userID)
	allowed, err := helpers.CheckRateLimit(key, 10, 60)
	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return nil, ErrRateLimited
	}

	query := facades.Orm().Query()

	// Check for existing report with same user + content item + reason
	var existing models.GameReport
	if err := query.Where("user_id", userID).
		Where("content_item_id", contentItemID).
		Where("reason", reason).
		First(&existing); err == nil && existing.ID != "" {
		// Duplicate: increment count, update note and level
		if note != nil {
			if _, err := facades.Orm().Query().Exec(
				"UPDATE game_reports SET count = count + 1, game_level_id = ?, note = ?, updated_at = NOW() WHERE id = ?",
				gameLevelID, *note, existing.ID,
			); err != nil {
				return nil, fmt.Errorf("failed to update report: %w", err)
			}
		} else {
			if _, err := facades.Orm().Query().Exec(
				"UPDATE game_reports SET count = count + 1, game_level_id = ?, updated_at = NOW() WHERE id = ?",
				gameLevelID, existing.ID,
			); err != nil {
				return nil, fmt.Errorf("failed to update report: %w", err)
			}
		}
		return &ReportResult{ID: existing.ID, Count: existing.Count + 1}, nil
	}

	// Create new report
	report := models.GameReport{
		ID:            newID(),
		UserID:        userID,
		GameID:        gameID,
		GameLevelID:   gameLevelID,
		ContentItemID: contentItemID,
		Reason:        reason,
		Note:          note,
		Count:         1,
	}
	if err := query.Create(&report); err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	return &ReportResult{ID: report.ID, Count: 1}, nil
}

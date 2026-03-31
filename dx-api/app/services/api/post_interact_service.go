package api

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"

	"dx-api/app/models"
)

type ToggleLikeResult struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"like_count"`
}

type ToggleBookmarkResult struct {
	Bookmarked bool `json:"bookmarked"`
}

func ToggleLike(userID, postID string) (*ToggleLikeResult, error) {
	var result *ToggleLikeResult

	err := facades.Orm().Transaction(func(tx orm.Query) error {
		var post models.Post
		if err := tx.Where("id", postID).Where("is_active", true).First(&post); err != nil || post.ID == "" {
			return ErrPostNotFound
		}

		var existing models.PostLike
		if err := tx.Where("post_id", postID).Where("user_id", userID).First(&existing); err == nil && existing.ID != "" {
			if _, err := tx.Exec("DELETE FROM post_likes WHERE id = ?", existing.ID); err != nil {
				return fmt.Errorf("failed to unlike: %w", err)
			}
			if _, err := tx.Exec("UPDATE posts SET like_count = like_count - 1 WHERE id = ? AND like_count > 0", postID); err != nil {
				return fmt.Errorf("failed to decrement like count: %w", err)
			}
			result = &ToggleLikeResult{Liked: false, LikeCount: post.LikeCount - 1}
			if result.LikeCount < 0 {
				result.LikeCount = 0
			}
			return nil
		}

		like := models.PostLike{
			ID:     uuid.Must(uuid.NewV7()).String(),
			PostID: postID,
			UserID: userID,
		}
		if err := tx.Create(&like); err != nil {
			return fmt.Errorf("failed to like: %w", err)
		}
		if _, err := tx.Exec("UPDATE posts SET like_count = like_count + 1 WHERE id = ?", postID); err != nil {
			return fmt.Errorf("failed to increment like count: %w", err)
		}
		result = &ToggleLikeResult{Liked: true, LikeCount: post.LikeCount + 1}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func ToggleBookmark(userID, postID string) (*ToggleBookmarkResult, error) {
	query := facades.Orm().Query()

	var post models.Post
	if err := query.Where("id", postID).Where("is_active", true).First(&post); err != nil || post.ID == "" {
		return nil, ErrPostNotFound
	}

	var existing models.PostBookmark
	if err := query.Where("post_id", postID).Where("user_id", userID).First(&existing); err == nil && existing.ID != "" {
		if _, err := query.Exec("DELETE FROM post_bookmarks WHERE id = ?", existing.ID); err != nil {
			return nil, fmt.Errorf("failed to unbookmark: %w", err)
		}
		return &ToggleBookmarkResult{Bookmarked: false}, nil
	}

	bookmark := models.PostBookmark{
		ID:     uuid.Must(uuid.NewV7()).String(),
		PostID: postID,
		UserID: userID,
	}
	if err := query.Create(&bookmark); err != nil {
		return nil, fmt.Errorf("failed to bookmark: %w", err)
	}

	return &ToggleBookmarkResult{Bookmarked: true}, nil
}

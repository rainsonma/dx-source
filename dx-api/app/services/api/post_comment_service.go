package api

import (
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"

	"dx-api/app/models"
)

// CommentItem is the response shape for a single comment.
type CommentItem struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	Author    PostAuthor `json:"author"`
	ParentID  *string    `json:"parent_id"`
	CreatedAt time.Time  `json:"created_at"`
}

// CommentWithReplies groups a top-level comment with its direct replies.
type CommentWithReplies struct {
	Comment CommentItem   `json:"comment"`
	Replies []CommentItem `json:"replies"`
}

// commentRow is used to scan raw SQL results for comment list queries.
type commentRow struct {
	models.PostComment
	AuthorNickname  *string `gorm:"column:author_nickname"`
	AuthorAvatarURL *string `gorm:"column:author_avatar_url"`
}

// CreateComment creates a comment (or reply) on a post.
// If parentID is set it is a reply; nested replies are rejected.
func CreateComment(userID, postID string, parentID *string, content string) (*CommentItem, error) {
	var result *CommentItem

	err := facades.Orm().Transaction(func(tx orm.Query) error {
		var post models.Post
		if err := tx.Where("id", postID).Where("is_active", true).First(&post); err != nil || post.ID == "" {
			return ErrPostNotFound
		}

		if parentID != nil && *parentID != "" {
			var parent models.PostComment
			if err := tx.Where("id", *parentID).Where("post_id", postID).First(&parent); err != nil || parent.ID == "" {
				return ErrCommentNotFound
			}
			if parent.ParentID != nil {
				return ErrNestedReply
			}
		}

		comment := models.PostComment{
			ID:       newID(),
			PostID:   postID,
			UserID:   userID,
			Content:  content,
			ParentID: parentID,
		}
		if err := tx.Create(&comment); err != nil {
			return fmt.Errorf("failed to create comment: %w", err)
		}

		if _, err := tx.Exec("UPDATE posts SET comment_count = comment_count + 1 WHERE id = ?", postID); err != nil {
			return fmt.Errorf("failed to increment comment count: %w", err)
		}

		var user models.User
		if err := tx.Where("id", userID).First(&user); err != nil {
			return fmt.Errorf("failed to load author: %w", err)
		}

		nickname := user.Username
		if user.Nickname != nil && *user.Nickname != "" {
			nickname = *user.Nickname
		}

		avatarURL := user.AvatarURL

		var createdAt time.Time
		if comment.CreatedAt != nil {
			createdAt = comment.CreatedAt.StdTime()
		}

		result = &CommentItem{
			ID:      comment.ID,
			Content: comment.Content,
			Author: PostAuthor{
				ID:        userID,
				Nickname:  nickname,
				AvatarURL: avatarURL,
			},
			ParentID:  parentID,
			CreatedAt: createdAt,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListComments returns paginated top-level comments with their replies for a post.
// Returns items, nextCursor, hasMore, error.
func ListComments(postID, cursor string, limit int) ([]CommentWithReplies, string, bool, error) {
	var post models.Post
	if err := facades.Orm().Query().Where("id", postID).Where("is_active", true).First(&post); err != nil || post.ID == "" {
		return nil, "", false, ErrPostNotFound
	}

	sql := `
		SELECT c.*, u.nickname AS author_nickname, u.avatar_url AS author_avatar_url
		FROM post_comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ? AND c.parent_id IS NULL`
	args := []any{postID}

	if cursor != "" {
		sql += " AND (c.created_at > (SELECT created_at FROM post_comments WHERE id = ?) OR (c.created_at = (SELECT created_at FROM post_comments WHERE id = ?) AND c.id != ?))"
		args = append(args, cursor, cursor, cursor)
	}
	sql += " ORDER BY c.created_at ASC LIMIT ?"
	args = append(args, limit+1)

	var topRows []commentRow
	if err := facades.Orm().Query().Raw(sql, args...).Scan(&topRows); err != nil {
		return nil, "", false, fmt.Errorf("failed to list comments: %w", err)
	}

	hasMore := len(topRows) > limit
	if hasMore {
		topRows = topRows[:limit]
	}

	if len(topRows) == 0 {
		return []CommentWithReplies{}, "", false, nil
	}

	// Collect parent IDs for batch reply loading.
	parentIDs := make([]string, 0, len(topRows))
	for _, r := range topRows {
		parentIDs = append(parentIDs, r.PostComment.ID)
	}

	replySql := `
		SELECT c.*, u.nickname AS author_nickname, u.avatar_url AS author_avatar_url
		FROM post_comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.parent_id IN ?
		ORDER BY c.created_at ASC`

	var replyRows []commentRow
	if err := facades.Orm().Query().Raw(replySql, parentIDs).Scan(&replyRows); err != nil {
		return nil, "", false, fmt.Errorf("failed to load replies: %w", err)
	}

	// Group replies by parent ID.
	replyMap := make(map[string][]CommentItem, len(parentIDs))
	for _, r := range replyRows {
		pid := ""
		if r.PostComment.ParentID != nil {
			pid = *r.PostComment.ParentID
		}
		replyMap[pid] = append(replyMap[pid], toCommentItem(r))
	}

	items := make([]CommentWithReplies, 0, len(topRows))
	for _, r := range topRows {
		replies := replyMap[r.PostComment.ID]
		if replies == nil {
			replies = []CommentItem{}
		}
		items = append(items, CommentWithReplies{
			Comment: toCommentItem(r),
			Replies: replies,
		})
	}

	var nextCursor string
	if hasMore {
		nextCursor = topRows[len(topRows)-1].PostComment.ID
	}

	return items, nextCursor, hasMore, nil
}

// UpdateComment updates the content of a comment the user owns.
func UpdateComment(userID, postID, commentID, content string) error {
	var comment models.PostComment
	if err := facades.Orm().Query().Where("id", commentID).Where("post_id", postID).First(&comment); err != nil || comment.ID == "" {
		return ErrCommentNotFound
	}
	if comment.UserID != userID {
		return ErrCommentNotOwner
	}

	if _, err := facades.Orm().Query().Model(&models.PostComment{}).Where("id", commentID).Update("content", content); err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}
	return nil
}

// DeleteComment deletes a comment and its replies (if top-level), decrementing the post's comment_count.
func DeleteComment(userID, postID, commentID string) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		var comment models.PostComment
		if err := tx.Where("id", commentID).Where("post_id", postID).First(&comment); err != nil || comment.ID == "" {
			return ErrCommentNotFound
		}
		if comment.UserID != userID {
			return ErrCommentNotOwner
		}

		decrement := 1

		if comment.ParentID == nil {
			// Count and delete replies first.
			type countRow struct {
				Count int `gorm:"column:count"`
			}
			var cr countRow
			if err := tx.Raw("SELECT COUNT(*) AS count FROM post_comments WHERE parent_id = ?", commentID).Scan(&cr); err != nil {
				return fmt.Errorf("failed to count replies: %w", err)
			}
			if cr.Count > 0 {
				if _, err := tx.Exec("DELETE FROM post_comments WHERE parent_id = ?", commentID); err != nil {
					return fmt.Errorf("failed to delete replies: %w", err)
				}
				decrement += cr.Count
			}
		}

		if _, err := tx.Exec("DELETE FROM post_comments WHERE id = ?", commentID); err != nil {
			return fmt.Errorf("failed to delete comment: %w", err)
		}

		if _, err := tx.Exec("UPDATE posts SET comment_count = GREATEST(comment_count - ?, 0) WHERE id = ?", decrement, postID); err != nil {
			return fmt.Errorf("failed to decrement comment count: %w", err)
		}

		return nil
	})
}

// toCommentItem converts a raw commentRow to a CommentItem.
func toCommentItem(r commentRow) CommentItem {
	nickname := ""
	if r.AuthorNickname != nil {
		nickname = *r.AuthorNickname
	}

	avatarURL := r.AuthorAvatarURL

	var createdAt time.Time
	if r.PostComment.CreatedAt != nil {
		createdAt = r.PostComment.CreatedAt.StdTime()
	}

	return CommentItem{
		ID:      r.PostComment.ID,
		Content: r.PostComment.Content,
		Author: PostAuthor{
			ID:        r.PostComment.UserID,
			Nickname:  nickname,
			AvatarURL: avatarURL,
		},
		ParentID:  r.PostComment.ParentID,
		CreatedAt: createdAt,
	}
}

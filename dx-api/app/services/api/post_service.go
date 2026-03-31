package api

import (
	"fmt"
	"strconv"
	"time"

	"dx-api/app/helpers"
	"dx-api/app/models"

	"github.com/goravel/framework/facades"
	"github.com/lib/pq"
)

// PostAuthor holds the author info embedded in a PostItem.
type PostAuthor struct {
	ID        string  `json:"id"`
	Nickname  string  `json:"nickname"`
	AvatarURL *string `json:"avatar_url"`
}

// PostItem is the response shape for a post.
type PostItem struct {
	ID           string     `json:"id"`
	Content      string     `json:"content"`
	ImageURL     *string    `json:"image_url"`
	Tags         []string   `json:"tags"`
	LikeCount    int        `json:"like_count"`
	CommentCount int        `json:"comment_count"`
	IsLiked      bool       `json:"is_liked"`
	IsBookmarked bool       `json:"is_bookmarked"`
	Author       PostAuthor `json:"author"`
	CreatedAt    time.Time  `json:"created_at"`
}

// postRow is used to scan raw SQL results for list queries.
type postRow struct {
	models.Post
	AuthorNickname *string `gorm:"column:author_nickname"`
	AuthorAvatarID *string `gorm:"column:author_avatar_id"`
	BookmarkID     string  `gorm:"column:bookmark_id"`
}

// CreatePost creates a new post and returns the PostItem.
func CreatePost(userID, content string, imageID *string, tags []string) (*PostItem, error) {
	if imageID != nil && *imageID != "" {
		var img models.Image
		if err := facades.Orm().Query().Where("id", *imageID).First(&img); err != nil || img.ID == "" {
			return nil, ErrImageNotFound
		}
		if img.UserID == nil || *img.UserID != userID {
			return nil, ErrImageNotOwned
		}
	}

	if tags == nil {
		tags = []string{}
	}

	post := models.Post{
		ID:       newID(),
		UserID:   userID,
		Content:  content,
		ImageID:  imageID,
		Tags:     pq.StringArray(tags),
		IsActive: true,
	}
	if err := facades.Orm().Query().Create(&post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return buildPostItem(&post, userID)
}

// GetPost returns a single post with like/bookmark state for the given user.
func GetPost(userID, postID string) (*PostItem, error) {
	var post models.Post
	if err := facades.Orm().Query().Where("id", postID).Where("is_active", true).First(&post); err != nil || post.ID == "" {
		return nil, ErrPostNotFound
	}
	return buildPostItem(&post, userID)
}

// ListPosts returns a paginated feed of posts for the given tab.
// Returns items, nextCursor, hasMore, error.
func ListPosts(userID, tab, cursor string, limit int) ([]PostItem, string, bool, error) {
	switch tab {
	case "hot":
		return listHotPosts(userID, cursor, limit)
	case "following":
		return listFollowingPosts(userID, cursor, limit)
	case "bookmarked":
		return listBookmarkedPosts(userID, cursor, limit)
	default:
		return listLatestPosts(userID, cursor, limit)
	}
}

// UpdatePost updates content, image, and tags of an owned post.
func UpdatePost(userID, postID, content string, imageID *string, tags []string) error {
	var post models.Post
	if err := facades.Orm().Query().Where("id", postID).Where("is_active", true).First(&post); err != nil || post.ID == "" {
		return ErrPostNotFound
	}
	if post.UserID != userID {
		return ErrPostNotOwner
	}

	if imageID != nil && *imageID != "" {
		var img models.Image
		if err := facades.Orm().Query().Where("id", *imageID).First(&img); err != nil || img.ID == "" {
			return ErrImageNotFound
		}
		if img.UserID == nil || *img.UserID != userID {
			return ErrImageNotOwned
		}
	}

	if tags == nil {
		tags = []string{}
	}

	updates := map[string]any{
		"content":  content,
		"image_id": imageID,
		"tags":     pq.StringArray(tags),
	}
	if _, err := facades.Orm().Query().Model(&models.Post{}).Where("id", postID).Update(updates); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}
	return nil
}

// DeletePost soft-deletes a post by setting is_active = false.
func DeletePost(userID, postID string) error {
	var post models.Post
	if err := facades.Orm().Query().Where("id", postID).Where("is_active", true).First(&post); err != nil || post.ID == "" {
		return ErrPostNotFound
	}
	if post.UserID != userID {
		return ErrPostNotOwner
	}

	if _, err := facades.Orm().Query().Model(&models.Post{}).Where("id", postID).Update("is_active", false); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	return nil
}

// --- Feed helpers ---

func listLatestPosts(userID, cursor string, limit int) ([]PostItem, string, bool, error) {
	sql := `
		SELECT p.*, u.nickname AS author_nickname, u.avatar_id AS author_avatar_id, '' AS bookmark_id
		FROM posts p
		JOIN users u ON u.id = p.user_id
		WHERE p.is_active = true`
	args := []any{}

	if cursor != "" {
		sql += " AND p.created_at < ?"
		args = append(args, cursor)
	}
	sql += " ORDER BY p.created_at DESC LIMIT ?"
	args = append(args, limit+1)

	var rows []postRow
	if err := facades.Orm().Query().Raw(sql, args...).Scan(&rows); err != nil {
		return nil, "", false, fmt.Errorf("failed to list latest posts: %w", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items, err := buildPostItems(rows, userID)
	if err != nil {
		return nil, "", false, err
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		ts := rows[len(rows)-1].Post.CreatedAt
		if ts != nil {
			nextCursor = ts.ToDateTimeString()
		}
	}

	return items, nextCursor, hasMore, nil
}

func listHotPosts(userID, cursor string, limit int) ([]PostItem, string, bool, error) {
	offset := 0
	if cursor != "" {
		fmt.Sscanf(cursor, "%d", &offset)
	}

	sql := `
		SELECT p.*, u.nickname AS author_nickname, u.avatar_id AS author_avatar_id, '' AS bookmark_id
		FROM posts p
		JOIN users u ON u.id = p.user_id
		WHERE p.is_active = true
		ORDER BY (p.like_count + p.comment_count * 2.0) / POWER(EXTRACT(EPOCH FROM (NOW() - p.created_at)) / 3600 + 2, 1.5) DESC
		LIMIT ? OFFSET ?`

	var rows []postRow
	if err := facades.Orm().Query().Raw(sql, limit+1, offset).Scan(&rows); err != nil {
		return nil, "", false, fmt.Errorf("failed to list hot posts: %w", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items, err := buildPostItems(rows, userID)
	if err != nil {
		return nil, "", false, err
	}

	var nextCursor string
	if hasMore {
		nextCursor = strconv.Itoa(offset + limit)
	}

	return items, nextCursor, hasMore, nil
}

func listFollowingPosts(userID, cursor string, limit int) ([]PostItem, string, bool, error) {
	sql := `
		SELECT p.*, u.nickname AS author_nickname, u.avatar_id AS author_avatar_id, '' AS bookmark_id
		FROM posts p
		JOIN users u ON u.id = p.user_id
		WHERE p.is_active = true
		  AND p.user_id IN (SELECT following_id FROM user_follows WHERE follower_id = ?)`
	args := []any{userID}

	if cursor != "" {
		sql += " AND p.created_at < ?"
		args = append(args, cursor)
	}
	sql += " ORDER BY p.created_at DESC LIMIT ?"
	args = append(args, limit+1)

	var rows []postRow
	if err := facades.Orm().Query().Raw(sql, args...).Scan(&rows); err != nil {
		return nil, "", false, fmt.Errorf("failed to list following posts: %w", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items, err := buildPostItems(rows, userID)
	if err != nil {
		return nil, "", false, err
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		ts := rows[len(rows)-1].Post.CreatedAt
		if ts != nil {
			nextCursor = ts.ToDateTimeString()
		}
	}

	return items, nextCursor, hasMore, nil
}

func listBookmarkedPosts(userID, cursor string, limit int) ([]PostItem, string, bool, error) {
	sql := `
		SELECT p.*, u.nickname AS author_nickname, u.avatar_id AS author_avatar_id, b.id AS bookmark_id
		FROM post_bookmarks b
		JOIN posts p ON p.id = b.post_id
		JOIN users u ON u.id = p.user_id
		WHERE b.user_id = ? AND p.is_active = true`
	args := []any{userID}

	if cursor != "" {
		sql += " AND b.created_at < ?"
		args = append(args, cursor)
	}
	sql += " ORDER BY b.created_at DESC LIMIT ?"
	args = append(args, limit+1)

	var rows []postRow
	if err := facades.Orm().Query().Raw(sql, args...).Scan(&rows); err != nil {
		return nil, "", false, fmt.Errorf("failed to list bookmarked posts: %w", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items, err := buildPostItems(rows, userID)
	if err != nil {
		return nil, "", false, err
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].BookmarkID
	}

	return items, nextCursor, hasMore, nil
}

// --- Builders ---

// buildPostItems batch-loads like/bookmark state and maps rows to PostItems.
func buildPostItems(rows []postRow, userID string) ([]PostItem, error) {
	if len(rows) == 0 {
		return []PostItem{}, nil
	}

	postIDs := make([]string, 0, len(rows))
	for _, r := range rows {
		postIDs = append(postIDs, r.Post.ID)
	}

	likedSet := loadUserPostSet("post_likes", userID, postIDs)
	bookmarkedSet := loadUserPostSet("post_bookmarks", userID, postIDs)

	items := make([]PostItem, 0, len(rows))
	for _, r := range rows {
		p := &r.Post

		var imageURL *string
		if p.ImageID != nil && *p.ImageID != "" {
			u := helpers.ImageServeURL(*p.ImageID)
			imageURL = &u
		}

		tags := []string{}
		if len(p.Tags) > 0 {
			tags = []string(p.Tags)
		}

		nickname := ""
		if r.AuthorNickname != nil {
			nickname = *r.AuthorNickname
		}

		var avatarURL *string
		if r.AuthorAvatarID != nil && *r.AuthorAvatarID != "" {
			u := helpers.ImageServeURL(*r.AuthorAvatarID)
			avatarURL = &u
		}

		var createdAt time.Time
		if p.CreatedAt != nil {
			createdAt = p.CreatedAt.StdTime()
		}

		items = append(items, PostItem{
			ID:           p.ID,
			Content:      p.Content,
			ImageURL:     imageURL,
			Tags:         tags,
			LikeCount:    p.LikeCount,
			CommentCount: p.CommentCount,
			IsLiked:      likedSet[p.ID],
			IsBookmarked: bookmarkedSet[p.ID],
			Author: PostAuthor{
				ID:        p.UserID,
				Nickname:  nickname,
				AvatarURL: avatarURL,
			},
			CreatedAt: createdAt,
		})
	}

	return items, nil
}

// buildPostItem builds a PostItem for a single post with individual queries.
func buildPostItem(post *models.Post, userID string) (*PostItem, error) {
	var imageURL *string
	if post.ImageID != nil && *post.ImageID != "" {
		u := helpers.ImageServeURL(*post.ImageID)
		imageURL = &u
	}

	tags := []string{}
	if len(post.Tags) > 0 {
		tags = []string(post.Tags)
	}

	var user models.User
	if err := facades.Orm().Query().Where("id", post.UserID).First(&user); err != nil {
		return nil, fmt.Errorf("failed to load post author: %w", err)
	}

	nickname := user.Username
	if user.Nickname != nil && *user.Nickname != "" {
		nickname = *user.Nickname
	}

	var avatarURL *string
	if user.AvatarID != nil && *user.AvatarID != "" {
		u := helpers.ImageServeURL(*user.AvatarID)
		avatarURL = &u
	}

	var like models.PostLike
	isLiked := false
	if err := facades.Orm().Query().Where("post_id", post.ID).Where("user_id", userID).First(&like); err == nil && like.ID != "" {
		isLiked = true
	}

	var bookmark models.PostBookmark
	isBookmarked := false
	if err := facades.Orm().Query().Where("post_id", post.ID).Where("user_id", userID).First(&bookmark); err == nil && bookmark.ID != "" {
		isBookmarked = true
	}

	var createdAt time.Time
	if post.CreatedAt != nil {
		createdAt = post.CreatedAt.StdTime()
	}

	return &PostItem{
		ID:           post.ID,
		Content:      post.Content,
		ImageURL:     imageURL,
		Tags:         tags,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		IsLiked:      isLiked,
		IsBookmarked: isBookmarked,
		Author: PostAuthor{
			ID:        post.UserID,
			Nickname:  nickname,
			AvatarURL: avatarURL,
		},
		CreatedAt: createdAt,
	}, nil
}

// loadUserPostSet returns a set of post IDs that the user has liked/bookmarked.
// table must be "post_likes" or "post_bookmarks" (hardcoded, not user input).
func loadUserPostSet(table, userID string, postIDs []string) map[string]bool {
	set := make(map[string]bool, len(postIDs))
	if len(postIDs) == 0 {
		return set
	}

	sql := fmt.Sprintf("SELECT post_id FROM %s WHERE user_id = ? AND post_id IN ?", table)

	type row struct {
		PostID string `gorm:"column:post_id"`
	}
	var rows []row
	facades.Orm().Query().Raw(sql, userID, postIDs).Scan(&rows)

	for _, r := range rows {
		set[r.PostID] = true
	}
	return set
}

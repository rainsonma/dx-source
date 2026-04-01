# Community Feature Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the core community feature (斗学社) — posts, comments, likes, bookmarks, follows, and feed tabs — across Go backend and Next.js frontend.

**Architecture:** Split-by-domain backend (post_service, post_comment_service, post_interact_service, follow_service) with matching controllers. Frontend refactors existing mock CommunityFeed into real data-driven components using SWR infinite scroll. All endpoints JWT-protected with cursor-based pagination.

**Tech Stack:** Go/Goravel, PostgreSQL, Next.js 16, React 19, SWR, Zod, Tailwind CSS, shadcn/ui

**Spec:** `docs/superpowers/specs/2026-03-31-community-feature-design.md`

---

### Task 1: Update PostComment Model and Migration

**Files:**
- Modify: `dx-api/app/models/post_comment.go`
- Modify: `dx-api/database/migrations/20260322000033_create_post_comments_table.go`

- [ ] **Step 1: Add ParentID field to PostComment model**

```go
// dx-api/app/models/post_comment.go
package models

import "github.com/goravel/framework/database/orm"

type PostComment struct {
	orm.Timestamps
	ID        string  `gorm:"column:id;primaryKey" json:"id"`
	PostID    string  `gorm:"column:post_id" json:"post_id"`
	UserID    string  `gorm:"column:user_id" json:"user_id"`
	Content   string  `gorm:"column:content" json:"content"`
	ParentID  *string `gorm:"column:parent_id" json:"parent_id"`
	LikeCount int     `gorm:"column:like_count" json:"like_count"`
}

func (p *PostComment) TableName() string {
	return "post_comments"
}
```

- [ ] **Step 2: Add parent_id column and index to migration**

```go
// dx-api/database/migrations/20260322000033_create_post_comments_table.go
// Inside the Create callback, add after existing columns:
			table.Uuid("parent_id").Nullable()
			// ... existing indexes ...
			table.Index("parent_id")
```

- [ ] **Step 3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/models/post_comment.go dx-api/database/migrations/20260322000033_create_post_comments_table.go
git commit -m "feat: add parent_id to post_comment model and migration"
```

---

### Task 2: Add Community Error Codes and Sentinels

**Files:**
- Modify: `dx-api/app/consts/error_code.go`
- Create: `dx-api/app/services/api/post_errors.go`

- [ ] **Step 1: Add error codes to consts**

Add these constants to `dx-api/app/consts/error_code.go` in the 404xx section:

```go
	CodePostNotFound    = 40409
	CodeCommentNotFound = 40410
```

- [ ] **Step 2: Create post error sentinels**

```go
// dx-api/app/services/api/post_errors.go
package api

import "errors"

var (
	ErrPostNotFound    = errors.New("帖子不存在")
	ErrPostNotOwner    = errors.New("无权操作此帖子")
	ErrCommentNotFound = errors.New("评论不存在")
	ErrCommentNotOwner = errors.New("无权操作此评论")
	ErrNestedReply     = errors.New("不能回复评论的回复")
	ErrSelfFollow      = errors.New("不能关注自己")
	ErrUserNotFound    = errors.New("用户不存在")
)
```

- [ ] **Step 3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/consts/error_code.go dx-api/app/services/api/post_errors.go
git commit -m "feat: add community error codes and sentinels"
```

---

### Task 3: Implement Follow Service

**Files:**
- Create: `dx-api/app/services/api/follow_service.go`

- [ ] **Step 1: Write follow service**

```go
// dx-api/app/services/api/follow_service.go
package api

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"

	"dx-api/app/models"
)

type ToggleFollowResult struct {
	Followed bool `json:"followed"`
}

func ToggleFollow(userID, targetUserID string) (*ToggleFollowResult, error) {
	if userID == targetUserID {
		return nil, ErrSelfFollow
	}

	query := facades.Orm().Query()

	var target models.User
	if err := query.Where("id", targetUserID).First(&target); err != nil || target.ID == "" {
		return nil, ErrUserNotFound
	}

	var existing models.UserFollow
	if err := query.Where("follower_id", userID).Where("following_id", targetUserID).First(&existing); err == nil && existing.ID != "" {
		if _, err := query.Exec("DELETE FROM user_follows WHERE id = ?", existing.ID); err != nil {
			return nil, fmt.Errorf("failed to unfollow: %w", err)
		}
		return &ToggleFollowResult{Followed: false}, nil
	}

	follow := models.UserFollow{
		ID:          uuid.Must(uuid.NewV7()).String(),
		FollowerID:  userID,
		FollowingID: targetUserID,
	}
	if err := query.Create(&follow); err != nil {
		return nil, fmt.Errorf("failed to follow: %w", err)
	}

	return &ToggleFollowResult{Followed: true}, nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/follow_service.go
git commit -m "feat: implement follow/unfollow toggle service"
```

---

### Task 4: Implement Post Interact Service

**Files:**
- Create: `dx-api/app/services/api/post_interact_service.go`

- [ ] **Step 1: Write post interact service**

```go
// dx-api/app/services/api/post_interact_service.go
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

	err := facades.Orm().Query().Transaction(func(tx orm.Query) error {
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
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/post_interact_service.go
git commit -m "feat: implement like/bookmark toggle services"
```

---

### Task 5: Implement Post Service

**Files:**
- Create: `dx-api/app/services/api/post_service.go`

- [ ] **Step 1: Write post service with response types and helpers**

```go
// dx-api/app/services/api/post_service.go
package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
	"github.com/lib/pq"

	"dx-api/app/helpers"
	"dx-api/app/models"
)

type PostAuthor struct {
	ID        string  `json:"id"`
	Nickname  string  `json:"nickname"`
	AvatarURL *string `json:"avatar_url"`
}

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

func CreatePost(userID, content string, imageID *string, tags []string) (*PostItem, error) {
	query := facades.Orm().Query()

	if imageID != nil && *imageID != "" {
		var image models.Image
		if err := query.Where("id", *imageID).First(&image); err != nil || image.ID == "" {
			return nil, fmt.Errorf("image not found: %s", *imageID)
		}
		if image.UserID == nil || *image.UserID != userID {
			return nil, ErrPostNotOwner
		}
	}

	post := models.Post{
		ID:       uuid.Must(uuid.NewV7()).String(),
		UserID:   userID,
		Content:  content,
		ImageID:  imageID,
		Tags:     pq.StringArray(tags),
		IsActive: true,
	}
	if err := query.Create(&post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return buildPostItem(query, &post, userID)
}

func GetPost(userID, postID string) (*PostItem, error) {
	query := facades.Orm().Query()

	var post models.Post
	if err := query.Where("id", postID).Where("is_active", true).First(&post); err != nil || post.ID == "" {
		return nil, ErrPostNotFound
	}

	return buildPostItem(query, &post, userID)
}

func ListPosts(userID, tab, cursor string, limit int) ([]PostItem, string, bool, error) {
	switch tab {
	case "hot":
		return listHotPosts(userID, cursor, limit)
	case "following":
		return listFollowingPosts(userID, cursor, limit)
	case "bookmarks":
		return listBookmarkedPosts(userID, cursor, limit)
	default:
		return listLatestPosts(userID, cursor, limit)
	}
}

func UpdatePost(userID, postID, content string, imageID *string, tags []string) error {
	query := facades.Orm().Query()

	var post models.Post
	if err := query.Where("id", postID).Where("is_active", true).First(&post); err != nil || post.ID == "" {
		return ErrPostNotFound
	}
	if post.UserID != userID {
		return ErrPostNotOwner
	}

	if imageID != nil && *imageID != "" {
		var image models.Image
		if err := query.Where("id", *imageID).First(&image); err != nil || image.ID == "" {
			return fmt.Errorf("image not found: %s", *imageID)
		}
		if image.UserID == nil || *image.UserID != userID {
			return ErrPostNotOwner
		}
	}

	post.Content = content
	post.ImageID = imageID
	post.Tags = pq.StringArray(tags)
	if err := query.Save(&post); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}
	return nil
}

func DeletePost(userID, postID string) error {
	query := facades.Orm().Query()

	var post models.Post
	if err := query.Where("id", postID).Where("is_active", true).First(&post); err != nil || post.ID == "" {
		return ErrPostNotFound
	}
	if post.UserID != userID {
		return ErrPostNotOwner
	}

	if _, err := query.Exec("UPDATE posts SET is_active = false WHERE id = ?", postID); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	return nil
}

// --- internal helpers ---

func listLatestPosts(userID, cursor string, limit int) ([]PostItem, string, bool, error) {
	q := facades.Orm().Query()

	var rows []postRow
	sql := `SELECT p.*, u.nickname AS author_nickname, u.avatar_id AS author_avatar_id
		FROM posts p JOIN users u ON p.user_id = u.id
		WHERE p.is_active = true`
	args := []any{}

	if cursor != "" {
		sql += ` AND p.created_at <= (SELECT created_at FROM posts WHERE id = ?) AND p.id != ?`
		args = append(args, cursor, cursor)
	}

	sql += ` ORDER BY p.created_at DESC LIMIT ?`
	args = append(args, limit+1)

	if err := q.Raw(sql, args...).Get(&rows); err != nil {
		return nil, "", false, fmt.Errorf("failed to list posts: %w", err)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items, err := buildPostItems(rows, userID)
	if err != nil {
		return nil, "", false, err
	}

	nextCursor := ""
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].ID
	}

	return items, nextCursor, hasMore, nil
}

func listHotPosts(userID, cursor string, limit int) ([]PostItem, string, bool, error) {
	q := facades.Orm().Query()

	offset := 0
	if cursor != "" {
		fmt.Sscanf(cursor, "%d", &offset)
	}

	var rows []postRow
	sql := `SELECT p.*, u.nickname AS author_nickname, u.avatar_id AS author_avatar_id,
		(p.like_count + p.comment_count * 2.0) / POWER(EXTRACT(EPOCH FROM (NOW() - p.created_at)) / 3600 + 2, 1.5) AS hot_score
		FROM posts p JOIN users u ON p.user_id = u.id
		WHERE p.is_active = true
		ORDER BY hot_score DESC
		LIMIT ? OFFSET ?`

	if err := q.Raw(sql, limit+1, offset).Get(&rows); err != nil {
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

	nextCursor := ""
	if hasMore {
		nextCursor = fmt.Sprintf("%d", offset+limit)
	}

	return items, nextCursor, hasMore, nil
}

func listFollowingPosts(userID, cursor string, limit int) ([]PostItem, string, bool, error) {
	q := facades.Orm().Query()

	var rows []postRow
	sql := `SELECT p.*, u.nickname AS author_nickname, u.avatar_id AS author_avatar_id
		FROM posts p JOIN users u ON p.user_id = u.id
		WHERE p.is_active = true
		AND p.user_id IN (SELECT following_id FROM user_follows WHERE follower_id = ?)`
	args := []any{userID}

	if cursor != "" {
		sql += ` AND p.created_at <= (SELECT created_at FROM posts WHERE id = ?) AND p.id != ?`
		args = append(args, cursor, cursor)
	}

	sql += ` ORDER BY p.created_at DESC LIMIT ?`
	args = append(args, limit+1)

	if err := q.Raw(sql, args...).Get(&rows); err != nil {
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

	nextCursor := ""
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].ID
	}

	return items, nextCursor, hasMore, nil
}

func listBookmarkedPosts(userID, cursor string, limit int) ([]PostItem, string, bool, error) {
	q := facades.Orm().Query()

	var rows []postRow
	sql := `SELECT p.*, u.nickname AS author_nickname, u.avatar_id AS author_avatar_id, pb.id AS bookmark_id
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN post_bookmarks pb ON p.id = pb.post_id
		WHERE p.is_active = true AND pb.user_id = ?`
	args := []any{userID}

	if cursor != "" {
		sql += ` AND pb.created_at <= (SELECT created_at FROM post_bookmarks WHERE id = ?) AND pb.id != ?`
		args = append(args, cursor, cursor)
	}

	sql += ` ORDER BY pb.created_at DESC LIMIT ?`
	args = append(args, limit+1)

	if err := q.Raw(sql, args...).Get(&rows); err != nil {
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

	nextCursor := ""
	if hasMore && len(rows) > 0 {
		nextCursor = rows[len(rows)-1].BookmarkID
	}

	return items, nextCursor, hasMore, nil
}

type postRow struct {
	models.Post
	AuthorNickname *string `gorm:"column:author_nickname"`
	AuthorAvatarID *string `gorm:"column:author_avatar_id"`
	BookmarkID     string  `gorm:"column:bookmark_id"`
}

func buildPostItems(rows []postRow, userID string) ([]PostItem, error) {
	if len(rows) == 0 {
		return []PostItem{}, nil
	}

	postIDs := make([]string, len(rows))
	for i, r := range rows {
		postIDs[i] = r.ID
	}

	likedSet := loadUserPostSet("post_likes", userID, postIDs)
	bookmarkedSet := loadUserPostSet("post_bookmarks", userID, postIDs)

	items := make([]PostItem, len(rows))
	for i, r := range rows {
		var imageURL *string
		if r.ImageID != nil && *r.ImageID != "" {
			url := helpers.ImageServeURL(*r.ImageID)
			imageURL = &url
		}

		nickname := r.UserID
		if r.AuthorNickname != nil {
			nickname = *r.AuthorNickname
		}

		var avatarURL *string
		if r.AuthorAvatarID != nil {
			url := helpers.ImageServeURL(*r.AuthorAvatarID)
			avatarURL = &url
		}

		tags := []string{}
		if r.Tags != nil {
			tags = r.Tags
		}

		items[i] = PostItem{
			ID:           r.ID,
			Content:      r.Content,
			ImageURL:     imageURL,
			Tags:         tags,
			LikeCount:    r.LikeCount,
			CommentCount: r.CommentCount,
			IsLiked:      likedSet[r.ID],
			IsBookmarked: bookmarkedSet[r.ID],
			Author: PostAuthor{
				ID:        r.UserID,
				Nickname:  nickname,
				AvatarURL: avatarURL,
			},
			CreatedAt: r.CreatedAt.StdTime(),
		}
	}
	return items, nil
}

func loadUserPostSet(table, userID string, postIDs []string) map[string]bool {
	set := make(map[string]bool)
	if len(postIDs) == 0 {
		return set
	}

	type row struct {
		PostID string `gorm:"column:post_id"`
	}
	var rows []row

	q := facades.Orm().Query()
	if err := q.Raw(
		fmt.Sprintf("SELECT post_id FROM %s WHERE user_id = ? AND post_id IN ?", table),
		userID, postIDs,
	).Get(&rows); err != nil {
		return set
	}

	for _, r := range rows {
		set[r.PostID] = true
	}
	return set
}

func buildPostItem(q interface{}, post *models.Post, userID string) (*PostItem, error) {
	query := facades.Orm().Query()

	var user models.User
	if err := query.Where("id", post.UserID).First(&user); err != nil || user.ID == "" {
		return nil, fmt.Errorf("author not found: %s", post.UserID)
	}

	var imageURL *string
	if post.ImageID != nil && *post.ImageID != "" {
		url := helpers.ImageServeURL(*post.ImageID)
		imageURL = &url
	}

	nickname := user.Username
	if user.Nickname != nil {
		nickname = *user.Nickname
	}

	var avatarURL *string
	if user.AvatarID != nil {
		url := helpers.ImageServeURL(*user.AvatarID)
		avatarURL = &url
	}

	var likeExists models.PostLike
	isLiked := false
	if err := query.Where("post_id", post.ID).Where("user_id", userID).First(&likeExists); err == nil && likeExists.ID != "" {
		isLiked = true
	}

	var bookmarkExists models.PostBookmark
	isBookmarked := false
	if err := query.Where("post_id", post.ID).Where("user_id", userID).First(&bookmarkExists); err == nil && bookmarkExists.ID != "" {
		isBookmarked = true
	}

	tags := []string{}
	if post.Tags != nil {
		tags = post.Tags
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
			ID:        user.ID,
			Nickname:  nickname,
			AvatarURL: avatarURL,
		},
		CreatedAt: post.CreatedAt.StdTime(),
	}, nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS (fix any type issues — the generic function may need to be replaced with a concrete type)

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/post_service.go
git commit -m "feat: implement post CRUD and feed listing service"
```

---

### Task 6: Implement Post Comment Service

**Files:**
- Create: `dx-api/app/services/api/post_comment_service.go`

- [ ] **Step 1: Write post comment service**

```go
// dx-api/app/services/api/post_comment_service.go
package api

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"

	"dx-api/app/helpers"
	"dx-api/app/models"
)

type CommentItem struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	Author    PostAuthor `json:"author"`
	ParentID  *string    `json:"parent_id"`
	CreatedAt time.Time  `json:"created_at"`
}

type CommentWithReplies struct {
	Comment CommentItem   `json:"comment"`
	Replies []CommentItem `json:"replies"`
}

func CreateComment(userID, postID string, parentID *string, content string) (*CommentItem, error) {
	var item *CommentItem

	err := facades.Orm().Query().Transaction(func(tx orm.Query) error {
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
			ID:       uuid.Must(uuid.NewV7()).String(),
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
		if err := tx.Where("id", userID).First(&user); err != nil || user.ID == "" {
			return fmt.Errorf("user not found: %s", userID)
		}

		nickname := user.Username
		if user.Nickname != nil {
			nickname = *user.Nickname
		}
		var avatarURL *string
		if user.AvatarID != nil {
			url := helpers.ImageServeURL(*user.AvatarID)
			avatarURL = &url
		}

		item = &CommentItem{
			ID:      comment.ID,
			Content: comment.Content,
			Author: PostAuthor{
				ID:        user.ID,
				Nickname:  nickname,
				AvatarURL: avatarURL,
			},
			ParentID:  parentID,
			CreatedAt: comment.CreatedAt.StdTime(),
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return item, nil
}

func ListComments(postID, cursor string, limit int) ([]CommentWithReplies, string, bool, error) {
	query := facades.Orm().Query()

	var post models.Post
	if err := query.Where("id", postID).Where("is_active", true).First(&post); err != nil || post.ID == "" {
		return nil, "", false, ErrPostNotFound
	}

	// Load top-level comments
	type commentRow struct {
		models.PostComment
		AuthorNickname *string `gorm:"column:author_nickname"`
		AuthorAvatarID *string `gorm:"column:author_avatar_id"`
	}

	var topRows []commentRow
	sql := `SELECT c.*, u.nickname AS author_nickname, u.avatar_id AS author_avatar_id
		FROM post_comments c JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ? AND c.parent_id IS NULL`
	args := []any{postID}

	if cursor != "" {
		sql += ` AND c.created_at >= (SELECT created_at FROM post_comments WHERE id = ?) AND c.id != ?`
		args = append(args, cursor, cursor)
	}

	sql += ` ORDER BY c.created_at ASC LIMIT ?`
	args = append(args, limit+1)

	if err := query.Raw(sql, args...).Get(&topRows); err != nil {
		return nil, "", false, fmt.Errorf("failed to list comments: %w", err)
	}

	hasMore := len(topRows) > limit
	if hasMore {
		topRows = topRows[:limit]
	}

	if len(topRows) == 0 {
		return []CommentWithReplies{}, "", false, nil
	}

	// Batch-load replies
	parentIDs := make([]string, len(topRows))
	for i, r := range topRows {
		parentIDs[i] = r.ID
	}

	var replyRows []commentRow
	replySql := `SELECT c.*, u.nickname AS author_nickname, u.avatar_id AS author_avatar_id
		FROM post_comments c JOIN users u ON c.user_id = u.id
		WHERE c.parent_id IN ?
		ORDER BY c.created_at ASC`
	if err := query.Raw(replySql, parentIDs).Get(&replyRows); err != nil {
		return nil, "", false, fmt.Errorf("failed to list replies: %w", err)
	}

	replyMap := make(map[string][]CommentItem)
	for _, r := range replyRows {
		replyMap[*r.ParentID] = append(replyMap[*r.ParentID], toCommentItem(r))
	}

	result := make([]CommentWithReplies, len(topRows))
	for i, r := range topRows {
		replies := replyMap[r.ID]
		if replies == nil {
			replies = []CommentItem{}
		}
		result[i] = CommentWithReplies{
			Comment: toCommentItem(r),
			Replies: replies,
		}
	}

	nextCursor := ""
	if hasMore {
		nextCursor = topRows[len(topRows)-1].ID
	}

	return result, nextCursor, hasMore, nil
}

func UpdateComment(userID, postID, commentID, content string) error {
	query := facades.Orm().Query()

	var comment models.PostComment
	if err := query.Where("id", commentID).Where("post_id", postID).First(&comment); err != nil || comment.ID == "" {
		return ErrCommentNotFound
	}
	if comment.UserID != userID {
		return ErrCommentNotOwner
	}

	comment.Content = content
	if err := query.Save(&comment); err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}
	return nil
}

func DeleteComment(userID, postID, commentID string) error {
	return facades.Orm().Query().Transaction(func(tx orm.Query) error {
		var comment models.PostComment
		if err := tx.Where("id", commentID).Where("post_id", postID).First(&comment); err != nil || comment.ID == "" {
			return ErrCommentNotFound
		}
		if comment.UserID != userID {
			return ErrCommentNotOwner
		}

		deleteCount := 1

		// If top-level comment, delete its replies too
		if comment.ParentID == nil {
			type countRow struct {
				Count int `gorm:"column:count"`
			}
			var cr countRow
			if err := tx.Raw("SELECT COUNT(*) AS count FROM post_comments WHERE parent_id = ?", commentID).First(&cr); err == nil {
				deleteCount += cr.Count
			}
			if _, err := tx.Exec("DELETE FROM post_comments WHERE parent_id = ?", commentID); err != nil {
				return fmt.Errorf("failed to delete replies: %w", err)
			}
		}

		if _, err := tx.Exec("DELETE FROM post_comments WHERE id = ?", commentID); err != nil {
			return fmt.Errorf("failed to delete comment: %w", err)
		}

		if _, err := tx.Exec("UPDATE posts SET comment_count = GREATEST(comment_count - ?, 0) WHERE id = ?", deleteCount, postID); err != nil {
			return fmt.Errorf("failed to decrement comment count: %w", err)
		}

		return nil
	})
}

func toCommentItem(r commentRow) CommentItem {
	nickname := r.PostComment.UserID
	if r.AuthorNickname != nil {
		nickname = *r.AuthorNickname
	}
	var avatarURL *string
	if r.AuthorAvatarID != nil {
		url := helpers.ImageServeURL(*r.AuthorAvatarID)
		avatarURL = &url
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
		CreatedAt: r.PostComment.CreatedAt.StdTime(),
	}
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 3: Commit**

```bash
git add dx-api/app/services/api/post_comment_service.go
git commit -m "feat: implement comment CRUD service with replies"
```

---

### Task 7: Implement Request Validators

**Files:**
- Create: `dx-api/app/http/requests/api/post_request.go`
- Create: `dx-api/app/http/requests/api/post_comment_request.go`

- [ ] **Step 1: Write post request validators**

```go
// dx-api/app/http/requests/api/post_request.go
package api

import "github.com/goravel/framework/contracts/http"

type CreatePostRequest struct {
	Content string   `form:"content" json:"content"`
	ImageID *string  `form:"image_id" json:"image_id"`
	Tags    []string `form:"tags" json:"tags"`
}

func (r *CreatePostRequest) Authorize(ctx http.Context) error { return nil }

func (r *CreatePostRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "required|min_len:1|max_len:2000",
		"tags":    "max_len:5",
	}
}

func (r *CreatePostRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "trim",
	}
}

func (r *CreatePostRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content.required": "请输入内容",
		"content.min_len":  "内容不能为空",
		"content.max_len":  "内容不能超过2000个字符",
		"tags.max_len":     "标签不能超过5个",
	}
}

type UpdatePostRequest struct {
	Content string   `form:"content" json:"content"`
	ImageID *string  `form:"image_id" json:"image_id"`
	Tags    []string `form:"tags" json:"tags"`
}

func (r *UpdatePostRequest) Authorize(ctx http.Context) error { return nil }

func (r *UpdatePostRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "required|min_len:1|max_len:2000",
		"tags":    "max_len:5",
	}
}

func (r *UpdatePostRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "trim",
	}
}

func (r *UpdatePostRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content.required": "请输入内容",
		"content.min_len":  "内容不能为空",
		"content.max_len":  "内容不能超过2000个字符",
		"tags.max_len":     "标签不能超过5个",
	}
}
```

- [ ] **Step 2: Write comment request validators**

```go
// dx-api/app/http/requests/api/post_comment_request.go
package api

import "github.com/goravel/framework/contracts/http"

type CreateCommentRequest struct {
	Content  string  `form:"content" json:"content"`
	ParentID *string `form:"parent_id" json:"parent_id"`
}

func (r *CreateCommentRequest) Authorize(ctx http.Context) error { return nil }

func (r *CreateCommentRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "required|min_len:1|max_len:500",
	}
}

func (r *CreateCommentRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "trim",
	}
}

func (r *CreateCommentRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content.required": "请输入评论内容",
		"content.min_len":  "评论内容不能为空",
		"content.max_len":  "评论内容不能超过500个字符",
	}
}

type UpdateCommentRequest struct {
	Content string `form:"content" json:"content"`
}

func (r *UpdateCommentRequest) Authorize(ctx http.Context) error { return nil }

func (r *UpdateCommentRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "required|min_len:1|max_len:500",
	}
}

func (r *UpdateCommentRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"content": "trim",
	}
}

func (r *UpdateCommentRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content.required": "请输入评论内容",
		"content.min_len":  "评论内容不能为空",
		"content.max_len":  "评论内容不能超过500个字符",
	}
}
```

- [ ] **Step 3: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 4: Commit**

```bash
git add dx-api/app/http/requests/api/post_request.go dx-api/app/http/requests/api/post_comment_request.go
git commit -m "feat: add post and comment request validators"
```

---

### Task 8: Implement Controllers

**Files:**
- Create: `dx-api/app/http/controllers/api/post_controller.go`
- Create: `dx-api/app/http/controllers/api/post_comment_controller.go`
- Create: `dx-api/app/http/controllers/api/post_interact_controller.go`
- Create: `dx-api/app/http/controllers/api/follow_controller.go`

- [ ] **Step 1: Write post controller**

```go
// dx-api/app/http/controllers/api/post_controller.go
package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type PostController struct{}

func NewPostController() *PostController {
	return &PostController{}
}

func (c *PostController) Create(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CreatePostRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	result, err := services.CreatePost(userID, req.Content, req.ImageID, req.Tags)
	if err != nil {
		return mapPostError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

func (c *PostController) List(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	tab := ctx.Request().Query("tab", "latest")
	cursor, limit := helpers.ParseCursorParams(ctx, helpers.DefaultCursorLimit)

	items, nextCursor, hasMore, err := services.ListPosts(userID, tab, cursor, limit)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to list posts")
	}

	return helpers.Paginated(ctx, items, nextCursor, hasMore)
}

func (c *PostController) Show(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	postID := ctx.Request().Route("id")
	result, err := services.GetPost(userID, postID)
	if err != nil {
		return mapPostError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

func (c *PostController) Update(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.UpdatePostRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	postID := ctx.Request().Route("id")
	if err := services.UpdatePost(userID, postID, req.Content, req.ImageID, req.Tags); err != nil {
		return mapPostError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

func (c *PostController) Delete(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	postID := ctx.Request().Route("id")
	if err := services.DeletePost(userID, postID); err != nil {
		return mapPostError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

func mapPostError(ctx contractshttp.Context, err error) contractshttp.Response {
	switch {
	case errors.Is(err, services.ErrPostNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodePostNotFound, "帖子不存在")
	case errors.Is(err, services.ErrPostNotOwner):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeForbidden, "无权操作此帖子")
	default:
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "操作失败")
	}
}
```

- [ ] **Step 2: Write post comment controller**

```go
// dx-api/app/http/controllers/api/post_comment_controller.go
package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type PostCommentController struct{}

func NewPostCommentController() *PostCommentController {
	return &PostCommentController{}
}

func (c *PostCommentController) Create(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CreateCommentRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	postID := ctx.Request().Route("id")
	result, err := services.CreateComment(userID, postID, req.ParentID, req.Content)
	if err != nil {
		return mapCommentError(ctx, err)
	}

	return helpers.Success(ctx, result)
}

func (c *PostCommentController) List(ctx contractshttp.Context) contractshttp.Response {
	postID := ctx.Request().Route("id")
	cursor, limit := helpers.ParseCursorParams(ctx, helpers.DefaultCursorLimit)

	items, nextCursor, hasMore, err := services.ListComments(postID, cursor, limit)
	if err != nil {
		return mapCommentError(ctx, err)
	}

	return helpers.Paginated(ctx, items, nextCursor, hasMore)
}

func (c *PostCommentController) Update(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.UpdateCommentRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	postID := ctx.Request().Route("id")
	commentID := ctx.Request().Route("commentId")
	if err := services.UpdateComment(userID, postID, commentID, req.Content); err != nil {
		return mapCommentError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

func (c *PostCommentController) Delete(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	postID := ctx.Request().Route("id")
	commentID := ctx.Request().Route("commentId")
	if err := services.DeleteComment(userID, postID, commentID); err != nil {
		return mapCommentError(ctx, err)
	}

	return helpers.Success(ctx, nil)
}

func mapCommentError(ctx contractshttp.Context, err error) contractshttp.Response {
	switch {
	case errors.Is(err, services.ErrPostNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodePostNotFound, "帖子不存在")
	case errors.Is(err, services.ErrCommentNotFound):
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeCommentNotFound, "评论不存在")
	case errors.Is(err, services.ErrCommentNotOwner):
		return helpers.Error(ctx, http.StatusForbidden, consts.CodeForbidden, "无权操作此评论")
	case errors.Is(err, services.ErrNestedReply):
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "不能回复评论的回复")
	default:
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "操作失败")
	}
}
```

- [ ] **Step 3: Write post interact controller**

```go
// dx-api/app/http/controllers/api/post_interact_controller.go
package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type PostInteractController struct{}

func NewPostInteractController() *PostInteractController {
	return &PostInteractController{}
}

func (c *PostInteractController) ToggleLike(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	postID := ctx.Request().Route("id")
	result, err := services.ToggleLike(userID, postID)
	if err != nil {
		if errors.Is(err, services.ErrPostNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodePostNotFound, "帖子不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "操作失败")
	}

	return helpers.Success(ctx, result)
}

func (c *PostInteractController) ToggleBookmark(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	postID := ctx.Request().Route("id")
	result, err := services.ToggleBookmark(userID, postID)
	if err != nil {
		if errors.Is(err, services.ErrPostNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodePostNotFound, "帖子不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "操作失败")
	}

	return helpers.Success(ctx, result)
}
```

- [ ] **Step 4: Write follow controller**

```go
// dx-api/app/http/controllers/api/follow_controller.go
package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type FollowController struct{}

func NewFollowController() *FollowController {
	return &FollowController{}
}

func (c *FollowController) ToggleFollow(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	targetUserID := ctx.Request().Route("id")
	result, err := services.ToggleFollow(userID, targetUserID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSelfFollow):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "不能关注自己")
		case errors.Is(err, services.ErrUserNotFound):
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeNotFound, "用户不存在")
		default:
			return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "操作失败")
		}
	}

	return helpers.Success(ctx, result)
}
```

- [ ] **Step 5: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 6: Commit**

```bash
git add dx-api/app/http/controllers/api/post_controller.go dx-api/app/http/controllers/api/post_comment_controller.go dx-api/app/http/controllers/api/post_interact_controller.go dx-api/app/http/controllers/api/follow_controller.go
git commit -m "feat: add post, comment, interact, and follow controllers"
```

---

### Task 9: Register Routes

**Files:**
- Modify: `dx-api/routes/api.go`

- [ ] **Step 1: Add controller instantiation and route registration**

Add inside the `Api()` function, within the `protected` middleware group (after existing route registrations):

```go
		// Community
		postController := apicontrollers.NewPostController()
		postCommentController := apicontrollers.NewPostCommentController()
		postInteractController := apicontrollers.NewPostInteractController()
		followController := apicontrollers.NewFollowController()

		protected.Post("/posts", postController.Create)
		protected.Get("/posts", postController.List)
		protected.Get("/posts/{id}", postController.Show)
		protected.Put("/posts/{id}", postController.Update)
		protected.Delete("/posts/{id}", postController.Delete)

		protected.Post("/posts/{id}/comments", postCommentController.Create)
		protected.Get("/posts/{id}/comments", postCommentController.List)
		protected.Put("/posts/{id}/comments/{commentId}", postCommentController.Update)
		protected.Delete("/posts/{id}/comments/{commentId}", postCommentController.Delete)

		protected.Post("/posts/{id}/like", postInteractController.ToggleLike)
		protected.Post("/posts/{id}/bookmark", postInteractController.ToggleBookmark)

		protected.Post("/users/{id}/follow", followController.ToggleFollow)
```

**Important:** Make sure the post routes (`/posts/{id}`) don't conflict with any existing routes. Place them AFTER specific routes like `/posts/something` if any exist (there shouldn't be any since this is new).

- [ ] **Step 2: Verify compilation**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 3: Run go vet**

Run: `cd dx-api && go vet ./...`
Expected: No issues

- [ ] **Step 4: Commit**

```bash
git add dx-api/routes/api.go
git commit -m "feat: register community API routes"
```

---

### Task 10: Backend Lint and Build Verification

**Files:** All Go files created/modified in Tasks 1-9

- [ ] **Step 1: Run gofmt**

Run: `cd dx-api && gofmt -l .`
Expected: No files listed (all formatted). If files are listed, run `gofmt -w .` to fix.

- [ ] **Step 2: Run go vet**

Run: `cd dx-api && go vet ./...`
Expected: No issues

- [ ] **Step 3: Run go build**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 4: Fix any issues found, then commit**

```bash
git add -A
git commit -m "fix: resolve lint and build issues in community backend"
```

---

### Task 11: Frontend Types and Schemas

**Files:**
- Create: `dx-web/src/features/web/community/types/post.ts`
- Create: `dx-web/src/features/web/community/schemas/post.schema.ts`

- [ ] **Step 1: Create types directory and post types**

```typescript
// dx-web/src/features/web/community/types/post.ts
export type PostAuthor = {
  id: string
  nickname: string
  avatar_url: string | null
}

export type Post = {
  id: string
  content: string
  image_url: string | null
  tags: string[]
  like_count: number
  comment_count: number
  is_liked: boolean
  is_bookmarked: boolean
  author: PostAuthor
  created_at: string
}

export type Comment = {
  id: string
  content: string
  author: PostAuthor
  parent_id: string | null
  created_at: string
}

export type CommentWithReplies = {
  comment: Comment
  replies: Comment[]
}

export type FeedTab = "latest" | "hot" | "following" | "bookmarks"
```

- [ ] **Step 2: Create schemas directory and post schemas**

```typescript
// dx-web/src/features/web/community/schemas/post.schema.ts
import { z } from "zod"

export const createPostSchema = z.object({
  content: z
    .string()
    .min(1, "请输入内容")
    .max(2000, "内容不能超过2000个字符"),
  image_id: z.string().optional(),
  tags: z
    .array(z.string().max(20, "标签不能超过20个字符"))
    .max(5, "标签不能超过5个")
    .optional(),
})

export type CreatePostInput = z.infer<typeof createPostSchema>

export const createCommentSchema = z.object({
  content: z
    .string()
    .min(1, "请输入评论内容")
    .max(500, "评论内容不能超过500个字符"),
  parent_id: z.string().optional(),
})

export type CreateCommentInput = z.infer<typeof createCommentSchema>
```

- [ ] **Step 3: Verify lint**

Run: `cd dx-web && npx eslint src/features/web/community/types/post.ts src/features/web/community/schemas/post.schema.ts`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/community/types/ dx-web/src/features/web/community/schemas/
git commit -m "feat: add community types and Zod schemas"
```

---

### Task 12: Frontend API Actions

**Files:**
- Create: `dx-web/src/features/web/community/actions/post.action.ts`

- [ ] **Step 1: Write API action wrappers**

```typescript
// dx-web/src/features/web/community/actions/post.action.ts
import { apiClient } from "@/lib/api-client"
import type { CursorPaginated } from "@/lib/api-client"
import type {
  Post,
  CommentWithReplies,
  Comment,
  FeedTab,
} from "../types/post"

export const postApi = {
  async list(tab: FeedTab, cursor?: string, limit?: number) {
    const params = new URLSearchParams()
    params.set("tab", tab)
    if (cursor) params.set("cursor", cursor)
    if (limit) params.set("limit", String(limit))
    return apiClient.get<CursorPaginated<Post>>(`/api/posts?${params}`)
  },

  async detail(id: string) {
    return apiClient.get<Post>(`/api/posts/${id}`)
  },

  async create(data: { content: string; image_id?: string; tags?: string[] }) {
    return apiClient.post<Post>("/api/posts", data)
  },

  async update(
    id: string,
    data: { content: string; image_id?: string; tags?: string[] }
  ) {
    return apiClient.put<null>(`/api/posts/${id}`, data)
  },

  async delete(id: string) {
    return apiClient.delete<null>(`/api/posts/${id}`)
  },

  async toggleLike(id: string) {
    return apiClient.post<{ liked: boolean; like_count: number }>(
      `/api/posts/${id}/like`
    )
  },

  async toggleBookmark(id: string) {
    return apiClient.post<{ bookmarked: boolean }>(
      `/api/posts/${id}/bookmark`
    )
  },

  async listComments(postId: string, cursor?: string, limit?: number) {
    const params = new URLSearchParams()
    if (cursor) params.set("cursor", cursor)
    if (limit) params.set("limit", String(limit))
    const qs = params.toString()
    return apiClient.get<CursorPaginated<CommentWithReplies>>(
      `/api/posts/${postId}/comments${qs ? `?${qs}` : ""}`
    )
  },

  async createComment(
    postId: string,
    data: { content: string; parent_id?: string }
  ) {
    return apiClient.post<Comment>(`/api/posts/${postId}/comments`, data)
  },

  async updateComment(postId: string, commentId: string, content: string) {
    return apiClient.put<null>(`/api/posts/${postId}/comments/${commentId}`, {
      content,
    })
  },

  async deleteComment(postId: string, commentId: string) {
    return apiClient.delete<null>(
      `/api/posts/${postId}/comments/${commentId}`
    )
  },

  async toggleFollow(userId: string) {
    return apiClient.post<{ followed: boolean }>(`/api/users/${userId}/follow`)
  },
}
```

- [ ] **Step 2: Verify lint**

Run: `cd dx-web && npx eslint src/features/web/community/actions/post.action.ts`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/community/actions/
git commit -m "feat: add community API action wrappers"
```

---

### Task 13: Frontend Hooks

**Files:**
- Create: `dx-web/src/features/web/community/hooks/use-post-feed.ts`
- Create: `dx-web/src/features/web/community/hooks/use-comments.ts`

- [ ] **Step 1: Write post feed hook with infinite scroll**

```typescript
// dx-web/src/features/web/community/hooks/use-post-feed.ts
"use client"

import { useRef, useCallback, useEffect } from "react"
import useSWRInfinite from "swr/infinite"

import type { CursorPaginated } from "@/lib/api-client"
import type { Post, FeedTab } from "../types/post"

export function usePostFeed(tab: FeedTab) {
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  const getKey = (
    pageIndex: number,
    previousPageData: CursorPaginated<Post> | null
  ) => {
    if (previousPageData && !previousPageData.hasMore) return null
    const params = new URLSearchParams()
    params.set("tab", tab)
    if (previousPageData?.nextCursor) {
      params.set("cursor", previousPageData.nextCursor)
    }
    return `/api/posts?${params}`
  }

  const { data, size, setSize, isLoading, isValidating, mutate } =
    useSWRInfinite<CursorPaginated<Post>>(getKey)

  const posts: Post[] =
    data?.flatMap((page) => page.items ?? []) ?? []
  const hasMore = data?.[data.length - 1]?.hasMore ?? false

  const loadMore = useCallback(() => {
    if (!isValidating && hasMore) setSize(size + 1)
  }, [isValidating, hasMore, size, setSize])

  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) loadMore()
      },
      { rootMargin: "200px" }
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [loadMore])

  return { posts, isLoading, isValidating, hasMore, sentinelRef, mutate }
}
```

- [ ] **Step 2: Write comments hook**

```typescript
// dx-web/src/features/web/community/hooks/use-comments.ts
"use client"

import { useRef, useCallback, useEffect } from "react"
import useSWRInfinite from "swr/infinite"

import type { CursorPaginated } from "@/lib/api-client"
import type { CommentWithReplies } from "../types/post"

export function useComments(postId: string) {
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  const getKey = (
    pageIndex: number,
    previousPageData: CursorPaginated<CommentWithReplies> | null
  ) => {
    if (previousPageData && !previousPageData.hasMore) return null
    const params = new URLSearchParams()
    if (previousPageData?.nextCursor) {
      params.set("cursor", previousPageData.nextCursor)
    }
    return `/api/posts/${postId}/comments?${params}`
  }

  const { data, size, setSize, isLoading, isValidating, mutate } =
    useSWRInfinite<CursorPaginated<CommentWithReplies>>(getKey)

  const comments: CommentWithReplies[] =
    data?.flatMap((page) => page.items ?? []) ?? []
  const hasMore = data?.[data.length - 1]?.hasMore ?? false

  const loadMore = useCallback(() => {
    if (!isValidating && hasMore) setSize(size + 1)
  }, [isValidating, hasMore, size, setSize])

  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) loadMore()
      },
      { rootMargin: "200px" }
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [loadMore])

  return { comments, isLoading, isValidating, hasMore, sentinelRef, mutate }
}
```

- [ ] **Step 3: Verify lint**

Run: `cd dx-web && npx eslint src/features/web/community/hooks/`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/community/hooks/
git commit -m "feat: add post feed and comments hooks with infinite scroll"
```

---

### Task 14: Frontend Components — Feed Tabs, Post Actions, Comment Input

**Files:**
- Create: `dx-web/src/features/web/community/components/feed-tabs.tsx`
- Create: `dx-web/src/features/web/community/components/post-actions.tsx`
- Create: `dx-web/src/features/web/community/components/comment-input.tsx`

- [ ] **Step 1: Write feed tabs component**

```tsx
// dx-web/src/features/web/community/components/feed-tabs.tsx
"use client"

import { Bookmark } from "lucide-react"
import { TabPill } from "@/components/in/tab-pill"
import type { FeedTab } from "../types/post"

const TABS: { key: FeedTab; label: string }[] = [
  { key: "latest", label: "最新" },
  { key: "hot", label: "热门" },
  { key: "following", label: "关注" },
  { key: "bookmarks", label: "书签" },
]

interface FeedTabsProps {
  active: FeedTab
  onChange: (tab: FeedTab) => void
}

export function FeedTabs({ active, onChange }: FeedTabsProps) {
  return (
    <div className="flex items-center gap-2">
      {TABS.map((t) => (
        <TabPill
          key={t.key}
          label={
            t.key === "bookmarks" ? undefined : t.label
          }
          active={active === t.key}
          onClick={() => onChange(t.key)}
        >
          {t.key === "bookmarks" && (
            <span className="flex items-center gap-1.5">
              <Bookmark className="h-3.5 w-3.5" />
              {t.label}
            </span>
          )}
        </TabPill>
      ))}
    </div>
  )
}
```

**Note:** If `TabPill` doesn't accept `children`, the implementing agent should use the `label` prop for all tabs and skip the Bookmark icon, or adjust the TabPill component accordingly. Check `src/components/in/tab-pill.tsx` interface.

- [ ] **Step 2: Write post actions component**

```tsx
// dx-web/src/features/web/community/components/post-actions.tsx
"use client"

import { useState } from "react"
import { Heart, MessageCircle, Bookmark } from "lucide-react"
import { toast } from "sonner"

import { postApi } from "../actions/post.action"

interface PostActionsProps {
  postId: string
  likeCount: number
  commentCount: number
  isLiked: boolean
  isBookmarked: boolean
  onCommentClick?: () => void
  onMutate?: () => void
}

export function PostActions({
  postId,
  likeCount: initialLikeCount,
  commentCount,
  isLiked: initialIsLiked,
  isBookmarked: initialIsBookmarked,
  onCommentClick,
  onMutate,
}: PostActionsProps) {
  const [isLiked, setIsLiked] = useState(initialIsLiked)
  const [likeCount, setLikeCount] = useState(initialLikeCount)
  const [isBookmarked, setIsBookmarked] = useState(initialIsBookmarked)
  const [pending, setPending] = useState(false)

  async function handleLike() {
    if (pending) return
    const prevLiked = isLiked
    const prevCount = likeCount
    setIsLiked(!prevLiked)
    setLikeCount(prevLiked ? prevCount - 1 : prevCount + 1)

    setPending(true)
    try {
      const res = await postApi.toggleLike(postId)
      if (res.code !== 0) {
        setIsLiked(prevLiked)
        setLikeCount(prevCount)
        toast.error(res.message)
        return
      }
      setIsLiked(res.data.liked)
      setLikeCount(res.data.like_count)
      onMutate?.()
    } catch {
      setIsLiked(prevLiked)
      setLikeCount(prevCount)
      toast.error("操作失败")
    } finally {
      setPending(false)
    }
  }

  async function handleBookmark() {
    if (pending) return
    const prev = isBookmarked
    setIsBookmarked(!prev)

    setPending(true)
    try {
      const res = await postApi.toggleBookmark(postId)
      if (res.code !== 0) {
        setIsBookmarked(prev)
        toast.error(res.message)
        return
      }
      setIsBookmarked(res.data.bookmarked)
      onMutate?.()
    } catch {
      setIsBookmarked(prev)
      toast.error("操作失败")
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="flex w-full items-center justify-around">
      <button
        type="button"
        onClick={handleLike}
        className="flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm text-muted-foreground hover:bg-muted"
      >
        <Heart
          className={`h-4 w-4 ${isLiked ? "fill-red-500 text-red-500" : ""}`}
        />
        <span>{likeCount || ""}</span>
      </button>
      <button
        type="button"
        onClick={onCommentClick}
        className="flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm text-muted-foreground hover:bg-muted"
      >
        <MessageCircle className="h-4 w-4" />
        <span>{commentCount || ""}</span>
      </button>
      <button
        type="button"
        onClick={handleBookmark}
        className="flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm text-muted-foreground hover:bg-muted"
      >
        <Bookmark
          className={`h-4 w-4 ${isBookmarked ? "fill-teal-600 text-teal-600" : ""}`}
        />
      </button>
    </div>
  )
}
```

- [ ] **Step 3: Write comment input component**

```tsx
// dx-web/src/features/web/community/components/comment-input.tsx
"use client"

import { useState } from "react"
import { Send, Loader2 } from "lucide-react"
import { toast } from "sonner"

import { postApi } from "../actions/post.action"
import { createCommentSchema } from "../schemas/post.schema"

interface CommentInputProps {
  postId: string
  parentId?: string
  placeholder?: string
  onCreated?: () => void
  onCancel?: () => void
}

export function CommentInput({
  postId,
  parentId,
  placeholder = "写下你的评论...",
  onCreated,
  onCancel,
}: CommentInputProps) {
  const [content, setContent] = useState("")
  const [pending, setPending] = useState(false)

  async function handleSubmit() {
    const result = createCommentSchema.safeParse({
      content,
      parent_id: parentId,
    })
    if (!result.success) {
      toast.error(result.error.issues[0].message)
      return
    }

    setPending(true)
    try {
      const res = await postApi.createComment(postId, {
        content,
        parent_id: parentId,
      })
      if (res.code !== 0) {
        toast.error(res.message)
        return
      }
      setContent("")
      onCreated?.()
    } catch {
      toast.error("评论失败")
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="flex items-center gap-2.5">
      <input
        type="text"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder={placeholder}
        maxLength={500}
        onKeyDown={(e) => {
          if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault()
            handleSubmit()
          }
        }}
        className="flex-1 rounded-lg border border-border bg-card px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-teal-600"
      />
      {onCancel && (
        <button
          type="button"
          onClick={onCancel}
          className="text-xs text-muted-foreground hover:text-foreground"
        >
          取消
        </button>
      )}
      <button
        type="button"
        onClick={handleSubmit}
        disabled={pending || !content.trim()}
        className="flex items-center rounded-lg bg-teal-600 px-3 py-2 text-white disabled:opacity-50"
      >
        {pending ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : (
          <Send className="h-4 w-4" />
        )}
      </button>
    </div>
  )
}
```

- [ ] **Step 4: Verify lint**

Run: `cd dx-web && npx eslint src/features/web/community/components/feed-tabs.tsx src/features/web/community/components/post-actions.tsx src/features/web/community/components/comment-input.tsx`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add dx-web/src/features/web/community/components/feed-tabs.tsx dx-web/src/features/web/community/components/post-actions.tsx dx-web/src/features/web/community/components/comment-input.tsx
git commit -m "feat: add feed tabs, post actions, and comment input components"
```

---

### Task 15: Frontend Components — Comment Item and Comment Section

**Files:**
- Create: `dx-web/src/features/web/community/components/comment-item.tsx`
- Create: `dx-web/src/features/web/community/components/comment-section.tsx`

- [ ] **Step 1: Write comment item component**

```tsx
// dx-web/src/features/web/community/components/comment-item.tsx
"use client"

import { useState } from "react"
import { Reply } from "lucide-react"

import { formatDistanceToNow } from "@/lib/format"
import { avatarColor } from "@/lib/avatar"
import type { Comment } from "../types/post"
import { CommentInput } from "./comment-input"

interface CommentItemProps {
  comment: Comment
  postId: string
  showReplyButton?: boolean
  onMutate?: () => void
}

export function CommentItem({
  comment,
  postId,
  showReplyButton = false,
  onMutate,
}: CommentItemProps) {
  const [replying, setReplying] = useState(false)

  return (
    <div className="flex gap-2.5">
      <div
        className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-bold text-white"
        style={{ backgroundColor: avatarColor(comment.author.id) }}
      >
        {comment.author.nickname.charAt(0).toUpperCase()}
      </div>
      <div className="flex flex-1 flex-col gap-1">
        <div className="flex items-center gap-2">
          <span className="text-xs font-semibold text-foreground">
            {comment.author.nickname}
          </span>
          <span className="text-xs text-muted-foreground">
            {formatDistanceToNow(comment.created_at)}
          </span>
        </div>
        <p className="text-sm leading-relaxed text-foreground">
          {comment.content}
        </p>
        {showReplyButton && (
          <button
            type="button"
            onClick={() => setReplying(!replying)}
            className="flex w-fit items-center gap-1 text-xs text-muted-foreground hover:text-teal-600"
          >
            <Reply className="h-3 w-3" />
            回复
          </button>
        )}
        {replying && (
          <div className="mt-1">
            <CommentInput
              postId={postId}
              parentId={comment.id}
              placeholder={`回复 ${comment.author.nickname}...`}
              onCreated={() => {
                setReplying(false)
                onMutate?.()
              }}
              onCancel={() => setReplying(false)}
            />
          </div>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Write comment section component**

```tsx
// dx-web/src/features/web/community/components/comment-section.tsx
"use client"

import { Spinner } from "@/components/ui/spinner"
import { useComments } from "../hooks/use-comments"
import { CommentItem } from "./comment-item"
import { CommentInput } from "./comment-input"

interface CommentSectionProps {
  postId: string
  commentCount: number
}

export function CommentSection({ postId, commentCount }: CommentSectionProps) {
  const { comments, isLoading, hasMore, sentinelRef, mutate } =
    useComments(postId)

  if (commentCount === 0 && !isLoading) {
    return (
      <div className="flex flex-col gap-3 rounded-[10px] bg-muted/50 p-4">
        <p className="text-center text-xs text-muted-foreground">
          暂无评论，来说两句吧
        </p>
        <CommentInput postId={postId} onCreated={() => mutate()} />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3.5 rounded-[10px] bg-muted/50 p-4">
      {isLoading && (
        <div className="flex justify-center py-2">
          <Spinner className="h-4 w-4" />
        </div>
      )}

      {comments.map((cwr) => (
        <div key={cwr.comment.id} className="flex flex-col gap-2.5">
          <CommentItem
            comment={cwr.comment}
            postId={postId}
            showReplyButton
            onMutate={() => mutate()}
          />
          {cwr.replies.length > 0 && (
            <div className="ml-9 flex flex-col gap-2.5 border-l-2 border-border pl-3">
              {cwr.replies.map((reply) => (
                <CommentItem
                  key={reply.id}
                  comment={reply}
                  postId={postId}
                  onMutate={() => mutate()}
                />
              ))}
            </div>
          )}
        </div>
      ))}

      {hasMore && <div ref={sentinelRef} className="h-1" />}

      <CommentInput postId={postId} onCreated={() => mutate()} />
    </div>
  )
}
```

- [ ] **Step 3: Verify lint**

Run: `cd dx-web && npx eslint src/features/web/community/components/comment-item.tsx src/features/web/community/components/comment-section.tsx`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/community/components/comment-item.tsx dx-web/src/features/web/community/components/comment-section.tsx
git commit -m "feat: add comment item and section components"
```

---

### Task 16: Frontend Components — Post Card

**Files:**
- Create: `dx-web/src/features/web/community/components/post-card.tsx`

- [ ] **Step 1: Write post card component**

```tsx
// dx-web/src/features/web/community/components/post-card.tsx
"use client"

import { useState } from "react"
import { UserPlus, UserCheck, MoreHorizontal } from "lucide-react"
import { toast } from "sonner"

import { formatDistanceToNow } from "@/lib/format"
import { avatarColor } from "@/lib/avatar"
import type { Post } from "../types/post"
import { postApi } from "../actions/post.action"
import { PostActions } from "./post-actions"
import { CommentSection } from "./comment-section"

interface PostCardProps {
  post: Post
  currentUserId?: string
  onMutate?: () => void
}

export function PostCard({ post, currentUserId, onMutate }: PostCardProps) {
  const [showComments, setShowComments] = useState(false)
  const [isFollowing, setIsFollowing] = useState(false)
  const isOwner = currentUserId === post.author.id

  async function handleFollow() {
    const prev = isFollowing
    setIsFollowing(!prev)
    try {
      const res = await postApi.toggleFollow(post.author.id)
      if (res.code !== 0) {
        setIsFollowing(prev)
        toast.error(res.message)
        return
      }
      setIsFollowing(res.data.followed)
    } catch {
      setIsFollowing(prev)
      toast.error("操作失败")
    }
  }

  return (
    <div className="flex flex-col gap-4 rounded-xl border border-border bg-card p-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          {post.author.avatar_url ? (
            <img
              src={post.author.avatar_url}
              alt={post.author.nickname}
              className="h-10 w-10 rounded-full object-cover"
            />
          ) : (
            <div
              className="flex h-10 w-10 items-center justify-center rounded-full text-sm font-bold text-white"
              style={{ backgroundColor: avatarColor(post.author.id) }}
            >
              {post.author.nickname.charAt(0).toUpperCase()}
            </div>
          )}
          <div className="flex flex-col">
            <span className="text-sm font-semibold text-foreground">
              {post.author.nickname}
            </span>
            <span className="text-xs text-muted-foreground">
              {formatDistanceToNow(post.created_at)}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {!isOwner && (
            <button
              type="button"
              onClick={handleFollow}
              className={`flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium ${
                isFollowing
                  ? "border border-border text-muted-foreground"
                  : "bg-teal-600 text-white"
              }`}
            >
              {isFollowing ? (
                <>
                  <UserCheck className="h-3.5 w-3.5" />
                  已关注
                </>
              ) : (
                <>
                  <UserPlus className="h-3.5 w-3.5" />
                  关注
                </>
              )}
            </button>
          )}
          {isOwner && (
            <button
              type="button"
              className="rounded-lg p-1.5 text-muted-foreground hover:bg-muted"
            >
              <MoreHorizontal className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>

      {/* Content */}
      <p className="whitespace-pre-wrap text-sm leading-relaxed text-foreground">
        {post.content}
      </p>

      {/* Image */}
      {post.image_url && (
        <img
          src={post.image_url}
          alt=""
          className="w-full rounded-[10px] object-cover"
        />
      )}

      {/* Tags */}
      {post.tags.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {post.tags.map((tag) => (
            <span
              key={tag}
              className="rounded-md bg-teal-50 px-2.5 py-1 text-xs font-medium text-teal-700"
            >
              #{tag}
            </span>
          ))}
        </div>
      )}

      {/* Divider */}
      <div className="h-px w-full bg-border" />

      {/* Actions */}
      <PostActions
        postId={post.id}
        likeCount={post.like_count}
        commentCount={post.comment_count}
        isLiked={post.is_liked}
        isBookmarked={post.is_bookmarked}
        onCommentClick={() => setShowComments(!showComments)}
        onMutate={onMutate}
      />

      {/* Comments */}
      {showComments && (
        <CommentSection postId={post.id} commentCount={post.comment_count} />
      )}
    </div>
  )
}
```

- [ ] **Step 2: Verify lint**

Run: `cd dx-web && npx eslint src/features/web/community/components/post-card.tsx`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/community/components/post-card.tsx
git commit -m "feat: add post card component"
```

---

### Task 17: Frontend Components — Create Post Dialog

**Files:**
- Create: `dx-web/src/features/web/community/components/create-post-dialog.tsx`

- [ ] **Step 1: Write create post dialog matching the .pen design**

```tsx
// dx-web/src/features/web/community/components/create-post-dialog.tsx
"use client"

import { useState } from "react"
import { X, Hash, Image, Loader2, Send } from "lucide-react"
import { toast } from "sonner"

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { postApi } from "../actions/post.action"
import { createPostSchema } from "../schemas/post.schema"

interface CreatePostDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreated: () => void
}

export function CreatePostDialog({
  open,
  onOpenChange,
  onCreated,
}: CreatePostDialogProps) {
  const [content, setContent] = useState("")
  const [tagInput, setTagInput] = useState("")
  const [tags, setTags] = useState<string[]>([])
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)

  function resetForm() {
    setContent("")
    setTagInput("")
    setTags([])
    setErrors({})
  }

  function addTag() {
    const tag = tagInput.trim()
    if (!tag) return
    if (tags.length >= 5) {
      toast.error("标签不能超过5个")
      return
    }
    if (tag.length > 20) {
      toast.error("标签不能超过20个字符")
      return
    }
    if (tags.includes(tag)) {
      setTagInput("")
      return
    }
    setTags([...tags, tag])
    setTagInput("")
  }

  function removeTag(tag: string) {
    setTags(tags.filter((t) => t !== tag))
  }

  async function handleSubmit() {
    setErrors({})
    const result = createPostSchema.safeParse({
      content,
      tags: tags.length > 0 ? tags : undefined,
    })
    if (!result.success) {
      const fieldErrors: Record<string, string> = {}
      for (const issue of result.error.issues) {
        const key = issue.path[0] as string
        if (!fieldErrors[key]) fieldErrors[key] = issue.message
      }
      setErrors(fieldErrors)
      return
    }

    setPending(true)
    try {
      const res = await postApi.create({
        content,
        tags: tags.length > 0 ? tags : undefined,
      })
      if (res.code !== 0) {
        toast.error(res.message)
        return
      }
      toast.success("发布成功")
      resetForm()
      onOpenChange(false)
      onCreated()
    } catch {
      toast.error("发布失败")
    } finally {
      setPending(false)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!v) resetForm()
        onOpenChange(v)
      }}
    >
      <DialogContent
        className="max-w-[640px] gap-0 p-0"
        aria-describedby={undefined}
      >
        <DialogHeader className="border-b border-border px-6 py-5">
          <DialogTitle className="text-lg font-bold">发布新帖子</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-5 px-6 py-5">
          {/* Textarea */}
          <div className="flex flex-col gap-1.5">
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="分享你的学习心得、提问或想法..."
              maxLength={2000}
              rows={6}
              className="w-full resize-none rounded-[10px] border border-border bg-muted/50 px-4 py-3 text-sm leading-relaxed placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-teal-600"
            />
            <div className="flex justify-end">
              <span className="text-xs text-muted-foreground">
                {content.length} / 2000
              </span>
            </div>
            {errors.content && (
              <p className="text-xs text-red-500">{errors.content}</p>
            )}
          </div>

          {/* Tags */}
          <div className="flex flex-wrap items-center gap-2.5">
            <div className="flex items-center gap-1.5">
              <Hash className="h-4 w-4 text-teal-600" />
              <span className="text-sm font-semibold text-teal-600">
                添加话题
              </span>
            </div>
            {tags.map((tag) => (
              <span
                key={tag}
                className="flex items-center gap-1.5 rounded-full bg-teal-50 px-3 py-1.5 text-xs font-medium text-teal-700"
              >
                #{tag}
                <button type="button" onClick={() => removeTag(tag)}>
                  <X className="h-3 w-3" />
                </button>
              </span>
            ))}
            {tags.length < 5 && (
              <input
                type="text"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault()
                    addTag()
                  }
                }}
                placeholder="输入标签后回车"
                maxLength={20}
                className="w-28 border-none bg-transparent text-xs text-foreground placeholder:text-muted-foreground focus:outline-none"
              />
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between border-t border-border px-6 py-4">
          <div className="flex items-center gap-1">
            <button
              type="button"
              className="flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground hover:bg-muted"
            >
              <Image className="h-5 w-5" />
            </button>
          </div>
          <Button
            onClick={handleSubmit}
            disabled={pending || !content.trim()}
            className="gap-2 bg-teal-600 hover:bg-teal-700"
          >
            {pending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Send className="h-4 w-4" />
            )}
            发布
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
```

- [ ] **Step 2: Verify lint**

Run: `cd dx-web && npx eslint src/features/web/community/components/create-post-dialog.tsx`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add dx-web/src/features/web/community/components/create-post-dialog.tsx
git commit -m "feat: add create post dialog component"
```

---

### Task 18: Refactor Community Feed — Replace Mock Data with Real Data

**Files:**
- Rewrite: `dx-web/src/features/web/community/components/community-feed.tsx`
- Modify: `dx-web/src/app/(web)/hall/(main)/community/page.tsx`

- [ ] **Step 1: Rewrite community-feed.tsx with real data integration**

```tsx
// dx-web/src/features/web/community/components/community-feed.tsx
"use client"

import { useState } from "react"
import { Plus } from "lucide-react"
import { Spinner } from "@/components/ui/spinner"
import { Empty } from "@/components/ui/empty"

import type { FeedTab } from "../types/post"
import { usePostFeed } from "../hooks/use-post-feed"
import { FeedTabs } from "./feed-tabs"
import { PostCard } from "./post-card"
import { CreatePostDialog } from "./create-post-dialog"

interface CommunityFeedProps {
  currentUserId?: string
}

export function CommunityFeed({ currentUserId }: CommunityFeedProps) {
  const [tab, setTab] = useState<FeedTab>("latest")
  const [createOpen, setCreateOpen] = useState(false)
  const { posts, isLoading, hasMore, sentinelRef, mutate } = usePostFeed(tab)

  return (
    <>
      {/* Tab row */}
      <div className="flex items-center justify-between">
        <FeedTabs active={tab} onChange={setTab} />
        <button
          type="button"
          onClick={() => setCreateOpen(true)}
          className="flex items-center gap-2 rounded-[10px] bg-teal-600 px-5 py-2.5 text-sm font-semibold text-white hover:bg-teal-700"
        >
          <Plus className="h-4 w-4" />
          发帖
        </button>
      </div>

      {/* Feed */}
      <div className="flex flex-col gap-4">
        {isLoading && posts.length === 0 && (
          <div className="flex justify-center py-12">
            <Spinner />
          </div>
        )}

        {!isLoading && posts.length === 0 && (
          <Empty description="暂无帖子" />
        )}

        {posts.map((post) => (
          <PostCard
            key={post.id}
            post={post}
            currentUserId={currentUserId}
            onMutate={() => mutate()}
          />
        ))}

        {hasMore && <div ref={sentinelRef} className="h-1" />}
      </div>

      <CreatePostDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onCreated={() => mutate()}
      />
    </>
  )
}
```

- [ ] **Step 2: Update community page to pass currentUserId**

Check how other pages get the current user. The `CommunityFeed` component is a client component, so it needs the user ID passed from the page or obtained client-side. Look at how `AuthGuard` or token-based approaches work.

If the page uses server-side auth via `auth()` from `@/lib/auth`, update `page.tsx`:

```tsx
// dx-web/src/app/(web)/hall/(main)/community/page.tsx
import { auth } from "@/lib/auth"
import { CommunityFeed } from "@/features/web/community/components/community-feed"

export default async function CommunityPage() {
  const session = await auth()

  return (
    <div className="flex h-full flex-col gap-6 px-4 py-7 md:px-8">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold text-foreground">斗学社</h1>
        <p className="text-sm text-muted-foreground">
          分享学习心得，与学友互动交流
        </p>
      </div>
      <CommunityFeed currentUserId={session?.user?.id} />
    </div>
  )
}
```

**Note:** The implementing agent should check the actual `auth()` return type from `src/lib/auth.ts` and adjust accordingly. If it returns `{ user: { id: string, name: string } }` or similar, use that shape. If the page was using `PageTopBar`, keep it if it fits, or replace with inline markup matching the design.

- [ ] **Step 3: Verify lint**

Run: `cd dx-web && npx eslint src/features/web/community/components/community-feed.tsx src/app/\(web\)/hall/\(main\)/community/page.tsx`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add dx-web/src/features/web/community/components/community-feed.tsx dx-web/src/app/\(web\)/hall/\(main\)/community/page.tsx
git commit -m "feat: refactor community feed from mock data to real API integration"
```

---

### Task 19: Full Build Verification

**Files:** All files created/modified in Tasks 1-18

- [ ] **Step 1: Backend — gofmt check**

Run: `cd dx-api && gofmt -l .`
Expected: No files listed. If any, run `gofmt -w .`

- [ ] **Step 2: Backend — go vet**

Run: `cd dx-api && go vet ./...`
Expected: No issues

- [ ] **Step 3: Backend — full build**

Run: `cd dx-api && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 4: Frontend — lint check**

Run: `cd dx-web && npm run lint`
Expected: No errors

- [ ] **Step 5: Frontend — build check**

Run: `cd dx-web && npm run build`
Expected: BUILD SUCCESS

- [ ] **Step 6: Fix any issues found, then commit**

```bash
git add -A
git commit -m "fix: resolve all lint and build issues for community feature"
```

---

### Task 20: Backend Tests

**Files:**
- Create: `dx-api/app/services/api/follow_service_test.go`
- Create: `dx-api/app/services/api/post_interact_service_test.go`

**Note:** These tests require a running PostgreSQL database. The implementing agent should check how existing tests are structured (look for `_test.go` files in the codebase) and follow the same setup pattern. If there's a test helper for database setup, use it. If tests use the real database, ensure the test environment is configured.

- [ ] **Step 1: Check for existing test patterns**

Run: `find dx-api -name "*_test.go" -type f | head -10`

Examine any existing test files to understand the test setup pattern (test database, fixtures, etc.).

- [ ] **Step 2: Write follow service tests**

Write table-driven tests for:
- Follow a user (returns `followed: true`)
- Unfollow a user (returns `followed: false`)
- Self-follow rejected (returns `ErrSelfFollow`)
- Follow non-existent user (returns `ErrUserNotFound`)

- [ ] **Step 3: Write post interact service tests**

Write table-driven tests for:
- Like a post (returns `liked: true`, count incremented)
- Unlike a post (returns `liked: false`, count decremented)
- Like non-existent post (returns `ErrPostNotFound`)
- Bookmark/unbookmark toggle
- Bookmark non-existent post (returns `ErrPostNotFound`)

- [ ] **Step 4: Run tests**

Run: `cd dx-api && go test -race ./app/services/api/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add dx-api/app/services/api/*_test.go
git commit -m "test: add follow and post interact service tests"
```

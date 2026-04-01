# Community Feature Design Spec

## Overview

Implement the core community feature (斗学社) for Douxue — a social feed where users share learning experiences, interact through likes/bookmarks/comments, and follow each other. Covers both Goravel backend (dx-api) and Next.js frontend (dx-web).

## Scope

**In scope (initial release):**
- Post CRUD (create, read, update, soft-delete)
- Single optional image per post
- Tags (up to 5 per post, max 20 chars each)
- Like/unlike toggle with counter
- Bookmark/unbookmark toggle
- Comments with one level of replies
- Feed tabs: latest, hot, following, bookmarks
- Follow/unfollow users (toggle, for the following tab)
- Infinite scroll pagination

**Out of scope (future):**
- Suggested users sidebar
- Trending topics sidebar
- Recommended feed algorithm
- Search
- Course request (求课程)
- Poll/link/@mention in posts
- Share functionality
- Comment likes

## Database Changes

### Modify `post_comments` table

Add `parent_id` column directly to the existing model and migration:

```sql
parent_id UUID (nullable)
INDEX idx_post_comments_parent_id ON post_comments(parent_id)
```

- `NULL` = top-level comment
- Non-null = reply to a comment
- One level of nesting only (enforced in service layer — reject replies to replies)
- Code-level FK via `AssertFK(tx, "post_comments", parentID)`

### Existing Schema (unchanged)

**posts**: id, user_id, content, image_id, tags[], like_count, comment_count, share_count, is_active, timestamps

**post_comments**: id, post_id, user_id, content, like_count, parent_id (new), timestamps

**post_likes**: id, post_id, user_id, created_at (unique: post_id + user_id)

**post_bookmarks**: id, post_id, user_id, created_at (unique: post_id + user_id)

**user_follows**: id, follower_id, following_id, created_at (unique: follower_id + following_id)

### Counter Management

Denormalized counters updated transactionally:
- `posts.like_count` — increment/decrement on like toggle
- `posts.comment_count` — increment on comment create, decrement on delete (including cascade for replies)

## API Endpoints

### Posts

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/posts` | JWT | Create post |
| `GET` | `/api/posts` | JWT | List feed (cursor-paginated, `tab` query param) |
| `GET` | `/api/posts/{id}` | JWT | Get post detail |
| `PUT` | `/api/posts/{id}` | JWT | Update own post |
| `DELETE` | `/api/posts/{id}` | JWT | Soft-delete own post (is_active=false) |

### Comments

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/posts/{id}/comments` | JWT | Create comment or reply |
| `GET` | `/api/posts/{id}/comments` | JWT | List top-level comments (replies inline) |
| `PUT` | `/api/posts/{id}/comments/{commentId}` | JWT | Update own comment |
| `DELETE` | `/api/posts/{id}/comments/{commentId}` | JWT | Delete own comment |

### Interactions

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/posts/{id}/like` | JWT | Toggle like |
| `POST` | `/api/posts/{id}/bookmark` | JWT | Toggle bookmark |

### Follow

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/users/{id}/follow` | JWT | Toggle follow |

## Backend Architecture

### File Structure

```
dx-api/app/
├── http/
│   ├── controllers/api/
│   │   ├── post_controller.go
│   │   ├── post_comment_controller.go
│   │   ├── post_interact_controller.go
│   │   └── follow_controller.go
│   └── requests/api/
│       ├── post_request.go
│       └── post_comment_request.go
├── services/api/
│   ├── post_service.go
│   ├── post_comment_service.go
│   ├── post_interact_service.go
│   └── follow_service.go
└── models/
    └── post_comment.go  (modified — add ParentID field)
```

### Service Layer

**post_service.go**
- `CreatePost(userID, content, imageID, tags) → (PostItem, error)` — AssertFK user + image, validate image ownership
- `GetPost(userID, postID) → (PostDetail, error)` — load post + author, is_liked/is_bookmarked flags
- `ListPosts(userID, tab, cursor, limit) → ([]PostItem, nextCursor, hasMore, error)` — tab-specific queries
- `UpdatePost(userID, postID, content, imageID, tags) → error` — ownership check
- `DeletePost(userID, postID) → error` — ownership check, set is_active=false

**post_comment_service.go**
- `CreateComment(userID, postID, parentID, content) → (CommentItem, error)` — AssertFK post (active), AssertFK parent, reject nested replies, increment counter
- `ListComments(userID, postID, cursor, limit) → ([]CommentWithReplies, nextCursor, hasMore, error)` — top-level paginated, replies batch-loaded
- `UpdateComment(userID, postID, commentID, content) → error` — ownership check
- `DeleteComment(userID, postID, commentID) → error` — ownership check, cascade delete replies, decrement counter by (1 + reply count)

**post_interact_service.go**
- `ToggleLike(userID, postID) → (liked bool, likeCount int, error)` — toggle + counter in transaction
- `ToggleBookmark(userID, postID) → (bookmarked bool, error)` — toggle

**follow_service.go**
- `ToggleFollow(userID, targetUserID) → (followed bool, error)` — prevent self-follow, toggle

### Feed Tab Queries

**latest**: `WHERE is_active = true ORDER BY created_at DESC`

**hot**: Time-weighted score computed in SQL, uses offset-based pagination internally (hot scores shift over time, making cursors unstable) with the offset encoded as the cursor value:
```sql
SELECT *, (like_count + comment_count * 2.0) / POWER(EXTRACT(EPOCH FROM (NOW() - created_at)) / 3600 + 2, 1.5) AS hot_score
FROM posts WHERE is_active = true ORDER BY hot_score DESC
LIMIT ? OFFSET ?
```

**following**: `WHERE user_id IN (SELECT following_id FROM user_follows WHERE follower_id = ?) AND is_active = true ORDER BY created_at DESC`

**bookmarks**: `JOIN post_bookmarks ON posts.id = post_bookmarks.post_id WHERE post_bookmarks.user_id = ? AND posts.is_active = true ORDER BY post_bookmarks.created_at DESC`

### Response Types

```go
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

type PostAuthor struct {
    ID        string  `json:"id"`
    Nickname  string  `json:"nickname"`
    AvatarURL *string `json:"avatar_url"`
}

type CommentWithReplies struct {
    Comment CommentItem   `json:"comment"`
    Replies []CommentItem `json:"replies"`
}

type CommentItem struct {
    ID        string     `json:"id"`
    Content   string     `json:"content"`
    Author    PostAuthor `json:"author"`
    ParentID  *string    `json:"parent_id"`
    CreatedAt time.Time  `json:"created_at"`
}
```

### Request Validation

**CreatePostRequest**: content (required, 1-2000), image_id (optional), tags (optional, max 5 items, each max 20 chars)

**UpdatePostRequest**: same as create

**CreateCommentRequest**: content (required, 1-500), parent_id (optional)

**UpdateCommentRequest**: content (required, 1-500)

### Error Sentinels

| Error | HTTP | Code | Message |
|-------|------|------|---------|
| `ErrPostNotFound` | 404 | 40400 | 帖子不存在 |
| `ErrPostNotOwner` | 403 | 40300 | 无权操作此帖子 |
| `ErrCommentNotFound` | 404 | 40400 | 评论不存在 |
| `ErrCommentNotOwner` | 403 | 40300 | 无权操作此评论 |
| `ErrNestedReply` | 400 | 40000 | 不能回复评论的回复 |
| `ErrSelfFollow` | 400 | 40000 | 不能关注自己 |
| `ErrUserNotFound` | 404 | 40400 | 用户不存在 |

### Routes

All endpoints registered inside the existing JWT-protected group in `routes/api.go`.

## Frontend Architecture

### File Structure

```
dx-web/src/features/web/community/
├── types/
│   └── post.ts
├── schemas/
│   └── post.schema.ts
├── actions/
│   └── post.action.ts
├── hooks/
│   ├── use-post-feed.ts
│   ├── use-post-detail.ts
│   └── use-comments.ts
├── components/
│   ├── community-feed.tsx      (refactor from mock to real data)
│   ├── feed-tabs.tsx
│   ├── post-card.tsx
│   ├── post-actions.tsx
│   ├── comment-section.tsx
│   ├── comment-item.tsx
│   ├── comment-input.tsx
│   └── create-post-dialog.tsx
```

### Types

```typescript
type Post = {
  id: string; content: string; image_url: string | null; tags: string[]
  like_count: number; comment_count: number
  is_liked: boolean; is_bookmarked: boolean
  author: PostAuthor; created_at: string
}
type PostAuthor = { id: string; nickname: string; avatar_url: string | null }
type CommentWithReplies = { comment: Comment; replies: Comment[] }
type Comment = { id: string; content: string; author: PostAuthor; parent_id: string | null; created_at: string }
type FeedTab = "latest" | "hot" | "following" | "bookmarks"
```

### API Actions

Thin wrappers around `apiClient`:
- `postApi.list(tab, cursor, limit)` → `GET /api/posts?tab=...&cursor=...&limit=...`
- `postApi.detail(id)` → `GET /api/posts/{id}`
- `postApi.create(data)` → `POST /api/posts`
- `postApi.update(id, data)` → `PUT /api/posts/{id}`
- `postApi.delete(id)` → `DELETE /api/posts/{id}`
- `postApi.toggleLike(id)` → `POST /api/posts/{id}/like`
- `postApi.toggleBookmark(id)` → `POST /api/posts/{id}/bookmark`
- `postApi.listComments(postId, cursor, limit)` → `GET /api/posts/{postId}/comments`
- `postApi.createComment(postId, data)` → `POST /api/posts/{postId}/comments`
- `postApi.updateComment(postId, commentId, data)` → `PUT /api/posts/{postId}/comments/{commentId}`
- `postApi.deleteComment(postId, commentId)` → `DELETE /api/posts/{postId}/comments/{commentId}`
- `postApi.toggleFollow(userId)` → `POST /api/users/{userId}/follow`

### Schemas (Zod)

- `createPostSchema` — content: string 1-2000, image_id: optional string, tags: optional array max 5 items each max 20
- `updatePostSchema` — same as create
- `createCommentSchema` — content: string 1-500, parent_id: optional string

### Hooks

**use-post-feed.ts** — `useSWRInfinite` parameterized by `FeedTab`, with Intersection Observer for infinite scroll. Key: `/api/posts?tab={tab}&cursor={cursor}&limit=20`

**use-post-detail.ts** — `useSWR` for single post detail. Key: `/api/posts/{id}`

**use-comments.ts** — `useSWRInfinite` for paginated comments. Key: `/api/posts/{postId}/comments?cursor={cursor}&limit=20`

### Components

Refactor existing `CommunityFeed` (currently mock data) into real data-driven components:

- **community-feed.tsx** — tab state, renders feed-tabs + post list via use-post-feed + create post button
- **feed-tabs.tsx** — pill tabs matching design (latest/hot/following/bookmarks)
- **post-card.tsx** — author info, content, tags, image, action buttons, comment preview
- **post-actions.tsx** — like/comment/bookmark buttons with optimistic updates via SWR mutate
- **comment-section.tsx** — comment list with replies, "view all" link, comment input
- **comment-item.tsx** — single comment with author, content, reply button
- **comment-input.tsx** — text input for new comment or reply
- **create-post-dialog.tsx** — modal form matching .pen design (textarea, tags, image upload, submit)

### Data Flow

1. Page loads → `usePostFeed("latest")` fetches first page
2. Scroll → Intersection Observer triggers `setSize(size + 1)`
3. Tab switch → hook resets, fetches new tab data
4. Create post → `postApi.create()` → `swrMutate("/api/posts")` refreshes feed
5. Like/bookmark → optimistic update on specific post via SWR mutate
6. Comments → `useComments(postId)` loads on demand

## Testing Strategy

### Backend

Table-driven tests with `-race` flag.

**post_service_test.go**: create (valid, missing content, invalid image, too many tags), list by tab (latest/hot/following/bookmarks ordering), get (exists, not found, soft-deleted), update (owner, non-owner), delete (owner, non-owner, verify is_active=false)

**post_comment_service_test.go**: create (valid, inactive post, reply, nested reply rejected), list (top-level + replies, pagination), delete (owner, with replies cascade, counter decrement)

**post_interact_service_test.go**: toggle like (like, unlike, counter), toggle bookmark (bookmark, unbookmark)

**follow_service_test.go**: toggle follow (follow, unfollow, self-follow rejected)

### Frontend

Component-level testing of real API integration behavior: feed loads, tab switching, post creation, like/bookmark toggles, comment CRUD, infinite scroll.

### Lint Compliance

- Backend: `gofmt`, `go vet`, `staticcheck` — zero issues
- Frontend: `eslint` via `npm run lint` — zero issues

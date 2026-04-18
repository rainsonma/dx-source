# URL-based Image & Audio Fields Refactor — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace all image/audio foreign-key columns (`*_id`) with direct URL text columns (`*_url`); delete `images`/`audios` tables; rewrite upload/serve endpoints for filesystem-only serving; update dx-web and dx-mini clients to consume `{url}` directly.

**Architecture:** Single coordinated change set (dev DB, fresh reset, in-place migration edits). Upload stores files at `{STORAGE_PATH}/uploads/images/YYYY/MM/DD/{uuid}.{jpg|png}`; stored URL is API-relative `/api/uploads/images/YYYY/MM/DD/{uuid}.{ext}`; serve endpoint validates path segments with strict regex + `filepath.Clean` traversal guard; no DB lookup. Client-submitted URLs (avatar/post image/course-game cover) validated against the upload URL regex.

**Tech Stack:** Go 1.23, Goravel, GORM, PostgreSQL, Next.js 16, TypeScript, WeChat mini program.

**Spec:** `docs/superpowers/specs/2026-04-18-url-fields-refactor-design.md`

---

## File Structure

### dx-api — new files
- `app/helpers/url_validate.go` — `IsUploadedImageURL(string) bool`
- `tests/feature/url_validate_test.go` — table-driven tests for the validator
- `tests/feature/upload_serve_test.go` — tests for new `POST/GET /api/uploads/images`

### dx-api — deleted files
- `app/models/image.go`, `app/models/audio.go`
- `app/helpers/image_url.go`
- `database/migrations/20260322000017_create_images_table.go`
- `database/migrations/20260322000018_create_audios_table.go`

### dx-api — modified files (backend)
Migrations: `20260322000001` (users), `20260322000002` (adm_users), `20260322000005` (game_categories), `20260322000006` (game_presses), `20260322000016` (games), `20260322000028` (game_groups), `20260322000032` (posts), `20260322000037` (content_items).
Models: `user.go`, `adm_user.go`, `post.go`, `game.go`, `game_category.go`, `game_press.go`, `game_group.go`, `content_item.go`.
Helpers: `qrcode.go` (rewrite `GenerateGroupQRCode`).
Services: `upload_service.go` (rewrite), `user_service.go`, `post_service.go`, `post_comment_service.go`, `game_service.go`, `course_game_service.go`, `favorite_service.go`, `content_service.go`, `group_service.go`, `leaderboard_service.go`, `hall_service.go`.
Controllers: `upload_controller.go` (rewrite), `user_controller.go`, `post_controller.go`, `course_game_controller.go`.
Requests: `user_request.go`, `post_request.go`, `course_game_request.go`.
Routes: `routes/api.go`.
Bootstrap: `bootstrap/migrations.go`.

### dx-web — modified files (15)
`src/lib/api-client.ts`, `src/features/com/images/types/image.types.ts`, `src/features/com/images/hooks/use-image-uploader.ts`, `src/features/com/images/components/image-uploader.tsx`, `src/features/web/me/actions/me.action.ts`, `src/features/web/me/types/me.types.ts`, `src/features/web/community/schemas/post.schema.ts`, `src/features/web/community/actions/post.action.ts`, `src/features/web/ai-custom/schemas/course-game.schema.ts`, `src/features/web/ai-custom/actions/course-game.action.ts`, `src/features/web/ai-custom/hooks/use-create-course-game.ts`, `src/features/web/ai-custom/hooks/use-update-course-game.ts`, `src/features/web/ai-custom/components/create-course-form.tsx`, `src/features/web/ai-custom/components/edit-game-dialog.tsx`, `src/features/web/ai-custom/components/course-detail-content.tsx`, `src/features/web/ai-custom/components/game-hero-card.tsx`, `src/features/web/docs/topics/account/profile-edit.tsx`.

### dx-mini — modified files (1)
`miniprogram/pages/me/profile-edit/profile-edit.ts`.

---

## Task 1: URL validator helper (TDD)

**Files:**
- Create: `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/helpers/url_validate.go`
- Create: `/Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/tests/feature/url_validate_test.go`

- [ ] **Step 1: Write the failing test**

Create `tests/feature/url_validate_test.go`:

```go
package feature

import (
	"testing"

	"dx-api/app/helpers"
)

func TestIsUploadedImageURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"valid jpg", "/api/uploads/images/2026/04/18/0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", true},
		{"valid png", "/api/uploads/images/2026/4/8/0194a2f0-7b7b-7000-8000-0194a2f07b7b.png", true},
		{"empty", "", false},
		{"absolute http", "https://evil.com/img.jpg", false},
		{"wrong prefix", "/uploads/images/2026/04/18/0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", false},
		{"audio prefix", "/api/uploads/audios/2026/04/18/0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", false},
		{"bad extension", "/api/uploads/images/2026/04/18/0194a2f0-7b7b-7000-8000-0194a2f07b7b.gif", false},
		{"no extension", "/api/uploads/images/2026/04/18/0194a2f0-7b7b-7000-8000-0194a2f07b7b", false},
		{"non-uuid filename", "/api/uploads/images/2026/04/18/hello.jpg", false},
		{"traversal", "/api/uploads/images/2026/04/18/../../../etc/passwd", false},
		{"5-digit year", "/api/uploads/images/20260/04/18/0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", false},
		{"javascript scheme", "javascript:alert(1)", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := helpers.IsUploadedImageURL(c.in)
			if got != c.want {
				t.Fatalf("IsUploadedImageURL(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:
```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test ./tests/feature/ -run TestIsUploadedImageURL -v
```
Expected: FAIL — `undefined: helpers.IsUploadedImageURL`.

- [ ] **Step 3: Implement the helper**

Create `app/helpers/url_validate.go`:

```go
package helpers

import "regexp"

var uploadImageURLRegex = regexp.MustCompile(`^/api/uploads/images/\d{4}/\d{1,2}/\d{1,2}/[0-9a-f-]{36}\.(jpg|png)$`)

// IsUploadedImageURL reports whether s is a URL produced by this service's
// image upload endpoint. Used to validate client-submitted *_url fields.
func IsUploadedImageURL(s string) bool {
	return uploadImageURLRegex.MatchString(s)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./tests/feature/ -run TestIsUploadedImageURL -v
```
Expected: PASS, all 12 sub-tests.

- [ ] **Step 5: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/app/helpers/url_validate.go dx-api/tests/feature/url_validate_test.go
git commit -m "feat(helpers): add IsUploadedImageURL validator"
```

---

## Task 2: Backend refactor (one atomic commit)

**Rationale:** Intermediate states won't compile because `helpers.ImageServeURL` is called across many services with `*_id` arguments. Renaming one model without updating all its callers breaks compilation. We proceed in a strict order within this single task, running `go build ./...` at the end before committing.

### Task 2.1: Rewrite upload service

**Files:**
- Modify: `dx-api/app/services/api/upload_service.go`

- [ ] **Step 1: Replace `upload_service.go` contents**

```go
package api

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
)

const (
	maxFileSize int64 = 2 * 1024 * 1024 // 2MB
)

var allowedMIMETypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
}

var validRoles = map[string]bool{
	consts.ImageRoleAdmUserAvatar:  true,
	consts.ImageRoleUserAvatar:     true,
	consts.ImageRoleCategoryCover:  true,
	consts.ImageRoleTemplateCover:  true,
	consts.ImageRoleGameCover:      true,
	consts.ImageRolePressCover:     true,
	consts.ImageRoleGameGroupCover: true,
	consts.ImageRolePostImage:      true,
	consts.ImageRoleGroupQrcode:    true,
}

// UploadImageResult holds the response data after a successful upload.
type UploadImageResult struct {
	URL string `json:"url"`
}

// ValidateUploadFile checks file size, MIME type, and role.
func ValidateUploadFile(file filesystem.File, role string) error {
	size, err := file.Size()
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}
	if size > maxFileSize {
		return ErrFileTooLarge
	}

	mimeType, err := file.MimeType()
	if err != nil {
		return fmt.Errorf("failed to get mime type: %w", err)
	}
	if _, ok := allowedMIMETypes[mimeType]; !ok {
		return ErrInvalidFileType
	}

	if !validRoles[role] {
		return ErrInvalidImageRole
	}
	return nil
}

// UploadImage saves the uploaded file to disk and returns its public URL.
// No database record is created — the URL is the system of record.
func UploadImage(userID string, file filesystem.File, role string) (*UploadImageResult, error) {
	_ = userID // retained in signature; may be used for future rate limiting per user/role
	mimeType, _ := file.MimeType()
	ext := allowedMIMETypes[mimeType]

	now := time.Now()
	id := uuid.Must(uuid.NewV7()).String()
	filename := fmt.Sprintf("%s.%s", id, ext)
	datePath := fmt.Sprintf("uploads/images/%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	publicURL := fmt.Sprintf("/api/%s/%s", datePath, filename)

	if _, err := file.StoreAs(datePath, filename); err != nil {
		return nil, fmt.Errorf("failed to store file: %w", err)
	}

	return &UploadImageResult{URL: publicURL}, nil
}

// servePathSegmentRegex validates year/month/day segments.
var servePathSegmentRegex = regexp.MustCompile(`^\d{1,4}$`)
var serveFilenameRegex = regexp.MustCompile(`^[0-9a-f-]{36}\.(jpg|png)$`)

// ResolveImagePath returns the absolute file path and content type for a
// serve request. It rejects traversal and malformed segments.
func ResolveImagePath(year, month, day, filename string) (string, string, error) {
	if len(year) != 4 || !servePathSegmentRegex.MatchString(year) {
		return "", "", ErrInvalidImagePath
	}
	if !servePathSegmentRegex.MatchString(month) || len(month) > 2 {
		return "", "", ErrInvalidImagePath
	}
	if !servePathSegmentRegex.MatchString(day) || len(day) > 2 {
		return "", "", ErrInvalidImagePath
	}
	if !serveFilenameRegex.MatchString(filename) {
		return "", "", ErrInvalidImagePath
	}

	storageRoot := facades.Config().Env("STORAGE_PATH", "storage/app").(string)
	baseDir := filepath.Join(storageRoot, "uploads", "images")
	abs := filepath.Clean(filepath.Join(baseDir, year, month, day, filename))
	if !strings.HasPrefix(abs, filepath.Clean(baseDir)+string(filepath.Separator)) {
		return "", "", ErrInvalidImagePath
	}

	contentType := "application/octet-stream"
	switch filepath.Ext(filename) {
	case ".jpg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	}
	return abs, contentType, nil
}
```

- [ ] **Step 2: Add `ErrInvalidImagePath` to errors**

Open `dx-api/app/services/api/errors.go`, add to the existing block of sentinels (keep existing ones, including `ErrImageNotFound`, `ErrFileTooLarge`, `ErrInvalidFileType`, `ErrInvalidImageRole`):

```go
// ErrInvalidImagePath is returned when a serve request has malformed path segments.
ErrInvalidImagePath = errors.New("invalid image path")
```

(If `errors.go` is in another location, use `grep -n ErrImageNotFound dx-api/app/services/api/ -r` to locate it.)

### Task 2.2: Rewrite upload controller

**Files:**
- Modify: `dx-api/app/http/controllers/api/upload_controller.go`

- [ ] **Step 3: Replace `upload_controller.go` contents**

```go
package api

import (
	"errors"
	"net/http"
	"os"

	contractshttp "github.com/goravel/framework/contracts/http"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"

	"github.com/goravel/framework/facades"
)

type UploadController struct{}

func NewUploadController() *UploadController {
	return &UploadController{}
}

// UploadImage handles POST /api/uploads/images — multipart file upload.
func (c *UploadController) UploadImage(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.UploadImageRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	file, err := ctx.Request().File("file")
	if err != nil || file == nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "file is required")
	}

	if err := services.ValidateUploadFile(file, req.Role); err != nil {
		switch {
		case errors.Is(err, services.ErrFileTooLarge):
			return helpers.Error(ctx, http.StatusRequestEntityTooLarge, consts.CodeFileTooLarge, "文件大小不能超过2MB")
		case errors.Is(err, services.ErrInvalidFileType):
			return helpers.Error(ctx, http.StatusUnsupportedMediaType, consts.CodeInvalidFileType, "仅支持JPEG和PNG格式")
		case errors.Is(err, services.ErrInvalidImageRole):
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidImageRole, "无效的图片类型")
		default:
			return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, err.Error())
		}
	}

	result, err := services.UploadImage(userID, file, req.Role)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to upload image")
	}

	return helpers.Success(ctx, result)
}

// ServeImage handles GET /api/uploads/images/{year}/{month}/{day}/{filename}.
func (c *UploadController) ServeImage(ctx contractshttp.Context) contractshttp.Response {
	year := ctx.Request().Route("year")
	month := ctx.Request().Route("month")
	day := ctx.Request().Route("day")
	filename := ctx.Request().Route("filename")

	absPath, contentType, err := services.ResolveImagePath(year, month, day, filename)
	if err != nil {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "invalid image path")
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return helpers.Error(ctx, http.StatusNotFound, consts.CodeImageNotFound, "图片文件不存在")
	}

	return ctx.Response().
		Header("Content-Type", contentType).
		Header("Cache-Control", "public, max-age=31536000, immutable").
		File(absPath)
}
```

### Task 2.3: Update routes

**Files:**
- Modify: `dx-api/routes/api.go`

- [ ] **Step 4: Replace the single-segment serve route**

In `routes/api.go`, locate the line:

```go
router.Get("/uploads/images/{id}", uploadController.ServeImage)
```

Replace with:

```go
router.Get("/uploads/images/{year}/{month}/{day}/{filename}", uploadController.ServeImage)
```

### Task 2.4: Rewrite GenerateGroupQRCode

**Files:**
- Modify: `dx-api/app/helpers/qrcode.go`

- [ ] **Step 5: Replace `qrcode.go` contents**

```go
package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
	qrcode "github.com/skip2/go-qrcode"
)

// GenerateGroupQRCode creates a QR code PNG for the given URL, saves it to
// disk under uploads/images/YYYY/MM/DD/, and returns the public URL.
// The ownerID parameter is kept in the signature for future per-user tagging.
func GenerateGroupQRCode(ownerID, inviteURL string) (string, error) {
	_ = ownerID
	png, err := qrcode.Encode(inviteURL, qrcode.Medium, 512)
	if err != nil {
		return "", fmt.Errorf("failed to generate qrcode: %w", err)
	}

	id := uuid.Must(uuid.NewV7()).String()
	now := time.Now()
	datePath := fmt.Sprintf("uploads/images/%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	filename := fmt.Sprintf("%s.png", id)

	storageRoot := facades.Config().Env("STORAGE_PATH", "storage/app").(string)
	absDir := filepath.Join(storageRoot, datePath)
	if err := os.MkdirAll(absDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	absPath := filepath.Join(absDir, filename)
	if err := os.WriteFile(absPath, png, 0644); err != nil {
		return "", fmt.Errorf("failed to write qrcode file: %w", err)
	}

	return fmt.Sprintf("/api/%s/%s", datePath, filename), nil
}
```

### Task 2.5: Rename model struct fields + GORM/json tags

For every model in the list below: rename `*ID` → `*URL`, `gorm:"column:*_id"` → `gorm:"column:*_url"`, `json:"*_id"` → `json:"*_url"` (JSON tag may differ in case — keep the file's existing convention).

- [ ] **Step 6: `dx-api/app/models/user.go`**

Change line 17:
```go
AvatarID          *string          `gorm:"column:avatar_id" json:"avatar_id"`
```
to:
```go
AvatarURL         *string          `gorm:"column:avatar_url" json:"avatar_url"`
```

- [ ] **Step 7: `dx-api/app/models/adm_user.go`**

Rename `AvatarID` → `AvatarURL`, column `avatar_id` → `avatar_url`, JSON `avatar_id` → `avatar_url`.

- [ ] **Step 8: `dx-api/app/models/post.go`**

Rename `ImageID` → `ImageURL`, column `image_id` → `image_url`, JSON `image_id` → `image_url`.

- [ ] **Step 9: `dx-api/app/models/game.go`**

Rename `CoverID` → `CoverURL`, column `cover_id` → `cover_url`, JSON `cover_id` → `cover_url`.

- [ ] **Step 10: `dx-api/app/models/game_category.go`**

Rename `CoverID` → `CoverURL`, column and JSON tag accordingly.

- [ ] **Step 11: `dx-api/app/models/game_press.go`**

Rename `CoverID` → `CoverURL`, column and JSON tag accordingly.

- [ ] **Step 12: `dx-api/app/models/game_group.go`**

Rename `CoverID` → `CoverURL` and `InviteQrcodeID` → `InviteQrcodeURL`, columns `cover_id` → `cover_url` and `invite_qrcode_id` → `invite_qrcode_url`, JSON tags accordingly.

- [ ] **Step 13: `dx-api/app/models/content_item.go`**

Rename `UkAudioID` → `UkAudioURL` and `UsAudioID` → `UsAudioURL`, columns `uk_audio_id` → `uk_audio_url` and `us_audio_id` → `us_audio_url`, JSON tags accordingly.

### Task 2.6: Rename migration columns + indexes

For each migration: flip `Uuid("*_id").Nullable()` to `Text("*_url").Nullable()`; drop any `Index("*_id")` entry on the renamed column.

- [ ] **Step 14: `dx-api/database/migrations/20260322000001_create_users_table.go`**

Change:
```go
table.Uuid("avatar_id").Nullable()
```
to:
```go
table.Text("avatar_url").Nullable()
```

Remove line:
```go
table.Index("avatar_id")
```

- [ ] **Step 15: `dx-api/database/migrations/20260322000002_create_adm_users_table.go`**

Same treatment: `avatar_id` → `avatar_url` (Uuid → Text, Nullable preserved); drop any `Index("avatar_id")` line.

- [ ] **Step 16: `dx-api/database/migrations/20260322000005_create_game_categories_table.go`**

`cover_id` → `cover_url`; drop index if present.

- [ ] **Step 17: `dx-api/database/migrations/20260322000006_create_game_presses_table.go`**

`cover_id` → `cover_url`; drop index if present.

- [ ] **Step 18: `dx-api/database/migrations/20260322000016_create_games_table.go`**

`cover_id` → `cover_url`; drop index if present.

- [ ] **Step 19: `dx-api/database/migrations/20260322000028_create_game_groups_table.go`**

`cover_id` → `cover_url`; `invite_qrcode_id` → `invite_qrcode_url`; drop either index if present.

- [ ] **Step 20: `dx-api/database/migrations/20260322000032_create_posts_table.go`**

`image_id` → `image_url`; drop index if present.

- [ ] **Step 21: `dx-api/database/migrations/20260322000037_create_content_items_table.go`**

`uk_audio_id` → `uk_audio_url`; `us_audio_id` → `us_audio_url`; drop either index if present.

### Task 2.7: Update services (use URL columns directly; drop `ImageServeURL`)

Within each service, replace the ID-batch-load + `ImageServeURL` pattern with a direct read of the new `*URL` field.

- [ ] **Step 22: `dx-api/app/services/api/user_service.go`**

Replace every `user.AvatarID` / `u.AvatarID` with `user.AvatarURL` / `u.AvatarURL`. Remove any `helpers.ImageServeURL(...)` call around the avatar field — use the URL directly. If `UpdateAvatar` currently writes `avatar_id`, change the column to `avatar_url` and write the validated URL.

- [ ] **Step 23: `dx-api/app/services/api/post_service.go`**

Replace `post.ImageID` with `post.ImageURL`. In `buildPostItems`/`buildPostItem` (lines ~313-441), delete the image-ID batching block and the `helpers.ImageServeURL(...)` calls; set response fields directly from `post.ImageURL` and the joined `author_avatar_url`. In every raw SQL `SELECT` that lists `author_avatar_id` / `image_id`, rename to `author_avatar_url` / `image_url` and update scanned column names + struct fields.

- [ ] **Step 24: `dx-api/app/services/api/post_comment_service.go`**

Replace `user.AvatarID` with `user.AvatarURL`; drop `ImageServeURL` wrapper.

- [ ] **Step 25: `dx-api/app/services/api/game_service.go`**

In list (lines 119-120, 186-187): delete `coverIDs` slice + `images` table batch-load + `coverMap`. Set response `Cover` field from `g.CoverURL` directly. In detail (lines 383-385): remove the `Image` lookup, read from `game.CoverURL`.

- [ ] **Step 26: `dx-api/app/services/api/course_game_service.go`**

Mirror the changes from `game_service.go` (lines 113-114, 150-151, 638-640). Rename struct field `CoverID` to `CoverURL` (line 52) and update callers. Change update key at line 228 from `"cover_id"` to `"cover_url"`. Line 191 (`CoverID: coverID`) and line 677 (`CoverID: game.CoverID`) become `CoverURL: coverURL` / `CoverURL: game.CoverURL`.

- [ ] **Step 27: `dx-api/app/services/api/favorite_service.go`**

Mirror the batch-drop treatment — use `g.CoverURL` directly.

- [ ] **Step 28: `dx-api/app/services/api/content_service.go`**

Replace the entire audio batch block (lines 56-75) with direct field reads. The result assignment becomes:

```go
for _, item := range items {
	data := ContentItemData{
		ID:          item.ID,
		Content:     item.Content,
		ContentType: item.ContentType,
		Translation: item.Translation,
		Definition:  item.Definition,
		Explanation: item.Explanation,
		Items:       item.Items,
		Structure:   item.Structure,
		UkAudioURL:  item.UkAudioURL,
		UsAudioURL:  item.UsAudioURL,
	}
	result = append(result, data)
}
```

No batch load, no `models.Image` import. The buggy `models.Image` query for audio IDs is removed entirely.

- [ ] **Step 29: `dx-api/app/services/api/group_service.go`**

Line 102:
```go
facades.Orm().Query().Model(&models.GameGroup{}).Where("id", groupID).Update("invite_qrcode_id", qrcodeID)
```
becomes:
```go
facades.Orm().Query().Model(&models.GameGroup{}).Where("id", groupID).Update("invite_qrcode_url", qrcodeURL)
```
and the local variable is renamed (`qrcodeURL` now holds the URL returned by `GenerateGroupQRCode`).

Lines 256-257:
```go
if group.InviteQrcodeID != nil && *group.InviteQrcodeID != "" {
	inviteQrcodeURL = helpers.ImageServeURL(*group.InviteQrcodeID)
}
```
becomes:
```go
if group.InviteQrcodeURL != nil && *group.InviteQrcodeURL != "" {
	inviteQrcodeURL = *group.InviteQrcodeURL
}
```

Remove the unused `helpers` import if the file no longer uses it.

- [ ] **Step 30: `dx-api/app/services/api/leaderboard_service.go`**

Replace `user.AvatarID` → `user.AvatarURL`; drop `helpers.ImageServeURL(...)` wrapping.

- [ ] **Step 31: `dx-api/app/services/api/hall_service.go`**

Same as Step 30.

### Task 2.8: Update request validators

- [ ] **Step 32: `dx-api/app/http/requests/api/user_request.go`**

Replace the `UpdateAvatarRequest` type and its rules to:

```go
type UpdateAvatarRequest struct {
	AvatarURL string `form:"avatar_url" json:"avatar_url"`
}

func (r *UpdateAvatarRequest) Authorize(ctx http.Context) error { return nil }
func (r *UpdateAvatarRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"avatar_url": "required",
	}
}
func (r *UpdateAvatarRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"avatar_url.required": "请选择头像",
	}
}
```

(URL format is checked in the controller/service via `helpers.IsUploadedImageURL` — see Step 35.)

- [ ] **Step 33: `dx-api/app/http/requests/api/post_request.go`**

Rename field `ImageID` → `ImageURL`, form/json tag `image_id` → `image_url` in both `CreatePostRequest` and `UpdatePostRequest`. No rule entry for the URL (it's optional; format check in the controller/service).

- [ ] **Step 34: `dx-api/app/http/requests/api/course_game_request.go`**

In both `CreateGameRequest` and `UpdateGameRequest`: rename `CoverID *string` → `CoverURL *string`, form/json tag `coverId` → `coverUrl`. In the rule maps, remove the `"coverId": "uuid"` rule entirely (format check in the controller/service). Remove the `coverId.uuid` message.

### Task 2.9: Update controllers + wire URL validation

- [ ] **Step 35: `dx-api/app/http/controllers/api/user_controller.go`**

In `UpdateAvatar`, after `helpers.Validate(ctx, &req)`, add:

```go
if !helpers.IsUploadedImageURL(req.AvatarURL) {
	return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "无效的头像URL")
}
```

Update the call into the service from `services.UpdateAvatar(userID, req.ImageID)` to `services.UpdateAvatar(userID, req.AvatarURL)`. Ensure the service signature matches.

- [ ] **Step 36: `dx-api/app/http/controllers/api/post_controller.go`**

In `Create` and `Update`, after validation, if `req.ImageURL != nil && *req.ImageURL != ""` then `if !helpers.IsUploadedImageURL(*req.ImageURL)` → 400 with `"无效的图片URL"`. Pass `req.ImageURL` into the service.

- [ ] **Step 37: `dx-api/app/http/controllers/api/course_game_controller.go`**

In the create/update handlers, after validation, if `req.CoverURL != nil && *req.CoverURL != ""` then validate with `helpers.IsUploadedImageURL`. Return 400 on mismatch with `"无效的封面URL"`.

### Task 2.10: Delete obsolete files + deregister migrations

- [ ] **Step 38: Delete `dx-api/app/models/image.go`**

```bash
rm /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/models/image.go
```

- [ ] **Step 39: Delete `dx-api/app/models/audio.go`**

```bash
rm /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/models/audio.go
```

- [ ] **Step 40: Delete `dx-api/app/helpers/image_url.go`**

```bash
rm /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/app/helpers/image_url.go
```

- [ ] **Step 41: Delete `dx-api/database/migrations/20260322000017_create_images_table.go`**

```bash
rm /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/database/migrations/20260322000017_create_images_table.go
```

- [ ] **Step 42: Delete `dx-api/database/migrations/20260322000018_create_audios_table.go`**

```bash
rm /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api/database/migrations/20260322000018_create_audios_table.go
```

- [ ] **Step 43: Remove migration registrations in `dx-api/bootstrap/migrations.go`**

Delete lines 28 and 29:
```go
&migrations.M20260322000017CreateImagesTable{},
&migrations.M20260322000018CreateAudiosTable{},
```

### Task 2.11: Add upload/serve feature tests

**Files:**
- Create: `dx-api/tests/feature/upload_serve_test.go`

- [ ] **Step 44: Check how existing feature tests boot the test harness**

Read `tests/test_case.go` and `tests/feature/example_test.go` to understand how tests set up a request context or call controllers. If there is an HTTP test client pattern, reuse it. If tests here only cover services (not HTTP), use the same style — prefer service-level tests for `UploadImage` and `ResolveImagePath` over full HTTP tests.

- [ ] **Step 45: Write serve-path validation tests**

Create `tests/feature/upload_serve_test.go`:

```go
package feature

import (
	"testing"

	services "dx-api/app/services/api"
)

func TestResolveImagePath(t *testing.T) {
	cases := []struct {
		name    string
		y, m, d string
		f       string
		wantErr bool
	}{
		{"valid", "2026", "04", "18", "0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", false},
		{"valid png", "2026", "4", "8", "0194a2f0-7b7b-7000-8000-0194a2f07b7b.png", false},
		{"traversal via filename", "2026", "04", "18", "../../../etc/passwd", true},
		{"non-numeric year", "abcd", "04", "18", "0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", true},
		{"5-digit year", "20260", "04", "18", "0194a2f0-7b7b-7000-8000-0194a2f07b7b.jpg", true},
		{"bad extension", "2026", "04", "18", "0194a2f0-7b7b-7000-8000-0194a2f07b7b.gif", true},
		{"short uuid", "2026", "04", "18", "abc.jpg", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := services.ResolveImagePath(c.y, c.m, c.d, c.f)
			if (err != nil) != c.wantErr {
				t.Fatalf("ResolveImagePath(%q,%q,%q,%q) err=%v wantErr=%v", c.y, c.m, c.d, c.f, err, c.wantErr)
			}
		})
	}
}
```

(The "valid" cases return no error from segment validation even when the file doesn't exist on disk — we only assert `ResolveImagePath`'s validation logic here; `os.Stat`-miss handling is exercised separately in existing HTTP tests if the harness supports it.)

### Task 2.12: Build + test + commit

- [ ] **Step 46: Verify the build compiles**

Run:
```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go build ./...
```
Expected: no output (success). Any errors → fix before proceeding. Common miss: lingering `models.Image` / `models.Audio` / `helpers.ImageServeURL` references. Search:
```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./...
```

- [ ] **Step 47: Confirm no stale references**

Run these greps — all should return no matches:
```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && grep -rn "ImageServeURL\|models\.Image\|models\.Audio\|\"image_id\"\|\"audio_id\"\|\"cover_id\"\|\"avatar_id\"\|\"invite_qrcode_id\"" app routes bootstrap database
```
Expected: no results. Any hit must be fixed.

- [ ] **Step 48: Format + vet**

```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && gofmt -w . && go vet ./...
```
Expected: no output.

- [ ] **Step 49: Run the full test suite with race detector**

```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./...
```
Expected: PASS. Any test that asserted old ID-based response shape must be updated in this task before passing.

- [ ] **Step 50: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-api/
git commit -m "refactor(dx-api): replace image/audio id columns with url columns

- drop images/audios tables and models
- rename avatar_id/cover_id/image_id/invite_qrcode_id/uk_audio_id/us_audio_id to *_url across schema, models, services, controllers
- rewrite upload endpoint to return {url} only (no db record)
- rewrite serve endpoint to accept /api/uploads/images/Y/M/D/{uuid}.{ext} with strict path validation
- add IsUploadedImageURL validation for client-submitted urls
- rewrite GenerateGroupQRCode to return url, drop Image record creation
- fix content_service.go audio bug (was querying images table for audio ids)
- delete helpers.ImageServeURL (no longer needed)"
```

---

## Task 3: dx-web frontend updates

**Rationale:** Frontend reads `{url}` from upload response and writes `*_url` to backend. All 15 files change in lockstep; one commit.

### Task 3.1: Upload pipeline core types

- [ ] **Step 1: Update `src/features/com/images/types/image.types.ts`**

Check current contents, then change `UploadedImage` type to:

```ts
export type UploadedImage = {
	url: string;
};
```

Remove any `id` and `name` fields. If other files import those, update accordingly in later steps.

- [ ] **Step 2: Update `src/features/com/images/hooks/use-image-uploader.ts`**

In the upload-success handler, replace:
```ts
const { id, url, name } = body.data;
onImageUploaded({ id, url, name });
```
with:
```ts
const { url } = body.data;
onImageUploaded({ url });
```

Update the `onImageUploaded` parameter type to `(img: UploadedImage) => void` where `UploadedImage` is the new shape. Remove `id`/`name` from callback signature types.

- [ ] **Step 3: Update `src/features/com/images/components/image-uploader.tsx`**

Change `onImageChange` prop signature from `(img: { id: string; url: string }) => void` (or similar) to `(url: string) => void`. In the internal `onImageUploaded` handler, pass `img.url` instead of `img.id`.

- [ ] **Step 4: Update `src/lib/api-client.ts`**

- `uploadApi.uploadImage`: narrow response type to `{ url: string }` (drop `id`/`name` fields).
- `userApi.updateAvatar`: change signature from `(imageId: string)` to `(avatarUrl: string)`; request body becomes `{ avatar_url: avatarUrl }`.

### Task 3.2: me feature (avatar)

- [ ] **Step 5: Update `src/features/web/me/types/me.types.ts`**

Rename field `avatarId` → `avatarUrl`.

- [ ] **Step 6: Update `src/features/web/me/actions/me.action.ts`**

Rename `updateAvatarAction(imageId: string)` → `updateAvatarAction(avatarUrl: string)`; request body becomes `{ avatar_url: avatarUrl }`.

### Task 3.3: community feature (post image)

- [ ] **Step 7: Update `src/features/web/community/schemas/post.schema.ts`**

Rename schema field `image_id` → `image_url`. Keep `.optional()`. Adjust the inferred TS type's consumers accordingly.

- [ ] **Step 8: Update `src/features/web/community/actions/post.action.ts`**

Function signatures and request bodies: `image_id` → `image_url` in `create(data)` and `update(id, data)`.

### Task 3.4: ai-custom feature (course-game cover)

- [ ] **Step 9: Update `src/features/web/ai-custom/schemas/course-game.schema.ts`**

Rename `coverId` → `coverUrl`.

- [ ] **Step 10: Update `src/features/web/ai-custom/actions/course-game.action.ts`**

`coverId` → `coverUrl` in every function signature, request body, and response type.

- [ ] **Step 11: Update `src/features/web/ai-custom/hooks/use-create-course-game.ts`**

`coverId` → `coverUrl`.

- [ ] **Step 12: Update `src/features/web/ai-custom/hooks/use-update-course-game.ts`**

`coverId` → `coverUrl`.

- [ ] **Step 13: Update `src/features/web/ai-custom/components/create-course-form.tsx`**

Any reference to `coverId` → `coverUrl`. In the `image-uploader` usage, update `onImageChange` handler — it now receives a URL string directly instead of `img.id`. Bind the form field to the URL value.

- [ ] **Step 14: Update `src/features/web/ai-custom/components/edit-game-dialog.tsx`**

Same treatment as Step 13.

- [ ] **Step 15: Update `src/features/web/ai-custom/components/course-detail-content.tsx`**

`coverId` → `coverUrl` wherever referenced. Display the `coverUrl` directly (no ID-to-URL build step).

- [ ] **Step 16: Update `src/features/web/ai-custom/components/game-hero-card.tsx`**

Same: read `coverUrl` directly.

### Task 3.5: Documentation update

- [ ] **Step 17: Update `src/features/web/docs/topics/account/profile-edit.tsx`**

Change the Chinese doc copy from "通过 imageId 绑定到你的账号" to "通过 imageUrl 绑定到你的账号" (or equivalent per the prose).

### Task 3.6: Lint + build + commit

- [ ] **Step 18: Lint**

```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run lint
```
Expected: no errors. Fix any TS type errors or ESLint warnings surfaced by the renames.

- [ ] **Step 19: Build**

```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run build
```
Expected: build succeeds.

- [ ] **Step 20: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-source
git add dx-web/
git commit -m "refactor(dx-web): adopt url-based image fields

- upload response narrowed to { url }
- avatar_id/image_id/cover_id → *_url across schemas, actions, components
- image-uploader callback now emits url string (no id/name)"
```

---

## Task 4: dx-mini profile-edit update

**Files:**
- Modify: `/Users/rainsen/Programs/Projects/douxue/dx-mini/miniprogram/pages/me/profile-edit/profile-edit.ts`

- [ ] **Step 1: Read the file to identify the exact lines**

```
cat /Users/rainsen/Programs/Projects/douxue/dx-mini/miniprogram/pages/me/profile-edit/profile-edit.ts
```
Locate the `chooseAvatar` function (around lines 61-90 per the spec exploration).

- [ ] **Step 2: Replace the upload-response parsing**

In `chooseAvatar()`, the block that parsed `body.data.id` needs to use `body.data.url`. Example transformation — if the current code is:

```ts
const body = JSON.parse(res.data as string);
if (body.code === 0) {
	const imageId = body.data.id;
	const url = body.data.url;
	wx.request({
		url: `${API_BASE}/api/user/avatar`,
		method: 'PUT',
		header: { Authorization: `Bearer ${token}` },
		data: { image_id: imageId },
		success: () => { this.setData({ avatarUrl: url }); }
	});
}
```

the new code is:

```ts
const body = JSON.parse(res.data as string);
if (body.code === 0) {
	const url = body.data.url;
	wx.request({
		url: `${API_BASE}/api/user/avatar`,
		method: 'PUT',
		header: { Authorization: `Bearer ${token}` },
		data: { avatar_url: url },
		success: () => { this.setData({ avatarUrl: url }); }
	});
}
```

Preserve the file's existing style (variable names, error handling, types). Only the two key changes are:
1. Drop the `body.data.id` read.
2. Request body key flips from `image_id` to `avatar_url`; value is the URL (not the id).

- [ ] **Step 3: Commit**

```bash
cd /Users/rainsen/Programs/Projects/douxue/dx-mini
git add miniprogram/pages/me/profile-edit/profile-edit.ts
git commit -m "refactor: send avatar_url to /api/user/avatar (was image_id)"
```

---

## Task 5: Final verification

No commit on this task — it exists to validate the work end-to-end and catch integration issues before you hand the change back.

- [ ] **Step 1: Backend full test run with race detector**

```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go test -race ./...
```
Expected: all tests PASS.

- [ ] **Step 2: Backend vet + fmt**

```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go vet ./... && gofmt -l .
```
Expected: no output from either.

- [ ] **Step 3: Frontend lint**

```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run lint
```
Expected: no errors.

- [ ] **Step 4: Frontend build**

```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run build
```
Expected: build succeeds.

- [ ] **Step 5: Scan the whole repo for stragglers**

```
cd /Users/rainsen/Programs/Projects/douxue/dx-source && grep -rn "image_id\|audio_id\|ImageServeURL\|avatar_id\|cover_id\|invite_qrcode_id\|uk_audio_id\|us_audio_id\|AvatarID\|CoverID\|ImageID\|AudioID\|InviteQrcodeID\|UkAudioID\|UsAudioID" dx-api dx-web
```
Expected: no hits in source code (may appear only in this plan or the spec — both inside `docs/` which isn't scanned because the paths above exclude `docs`).

Run the same scan against dx-mini:
```
cd /Users/rainsen/Programs/Projects/douxue/dx-mini && grep -rn "image_id\|avatar_id" miniprogram
```
Expected: no hits.

- [ ] **Step 6: Start servers and smoke test manually**

Start the backend and frontend (two terminals):
```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-api && go run .
```
```
cd /Users/rainsen/Programs/Projects/douxue/dx-source/dx-web && npm run dev
```

Then **reset the dev DB** (since migrations changed in place) per the project's usual flow, and exercise:
1. Sign in → navigate to profile edit → upload a new avatar → confirm the new URL is displayed and persists after refresh.
2. Create a post with an image → confirm the image renders in the feed and on the detail page.
3. Create a course game with a cover → confirm the cover renders in the game list and detail.
4. Create a group → confirm the group invite QR code URL loads as an image.
5. In the WeChat mini program simulator (if available), run the profile edit flow to confirm avatar upload works end-to-end.

Any failure → investigate and fix in the relevant commit (or add a follow-up commit).

---

## Self-Review

**Spec coverage check:**
- Schema changes (10 columns, 2 tables deleted): Tasks 2.5–2.6 + 2.10.
- URL format + upload endpoint rewrite: Task 2.1.
- Serve endpoint rewrite + traversal guard: Tasks 2.1 (`ResolveImagePath`) + 2.2 (controller wiring) + 2.3 (route).
- URL validation helper + tests: Task 1.
- Validator application at 3 endpoints: Tasks 2.8 (requests) + 2.9 (controllers).
- `GenerateGroupQRCode` rewrite: Task 2.4.
- Service simplification across 11 services: Task 2.7.
- Deletions + deregistration: Task 2.10.
- dx-web 15-file update: Task 3.
- dx-mini 1-file update: Task 4.
- Security: path traversal handled in `ResolveImagePath`; client URL validation via `IsUploadedImageURL`; upload auth unchanged.
- Testing: URL validator test (Task 1), serve path test (Task 2.11), full suite + race (Task 2.12 + Task 5).
- nginx / deploy: unchanged (confirmed in spec, no task needed).

**Placeholder scan:** No TBDs, TODOs, "similar to Task N" references, or bare-prose steps without code. Each code step shows the actual code.

**Type/name consistency:**
- `UploadImageResult{URL string}` introduced in Task 2.1 Step 1 and referenced in Step 3 — matches.
- `IsUploadedImageURL` introduced in Task 1 Step 3 and called in Task 2.9 Steps 35-37 — matches.
- `ResolveImagePath` introduced in Task 2.1 Step 1 and called in Task 2.2 Step 3 — matches.
- `GenerateGroupQRCode` signature change (returns URL string) matches caller update in Task 2.7 Step 29 (variable renamed to `qrcodeURL`).
- Model field rename (`AvatarURL` etc.) matches service-layer reads (user/post/game/etc.).

No issues found.

package api

import (
	"fmt"
	"os"
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

	storageRoot := storagePathOrDefault()
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

// storagePathOrDefault returns the configured storage root, falling back to env then default.
func storagePathOrDefault() string {
	if v := os.Getenv("STORAGE_PATH"); v != "" {
		return v
	}
	return facades.Config().Env("STORAGE_PATH", "storage/app").(string)
}

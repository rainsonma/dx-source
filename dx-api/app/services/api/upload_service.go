package api

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	"dx-api/app/models"
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
	ID   string `json:"id"`
	URL  string `json:"url"`
	Name string `json:"name"`
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

// UploadImage saves the uploaded file to disk and creates a DB record.
func UploadImage(userID string, file filesystem.File, role string) (*UploadImageResult, error) {
	mimeType, _ := file.MimeType()
	ext := allowedMIMETypes[mimeType]
	size, _ := file.Size()

	// Generate date-based path with UUID v7 filename
	now := time.Now()
	id := uuid.Must(uuid.NewV7()).String()
	filename := fmt.Sprintf("%s.%s", id, ext)
	datePath := fmt.Sprintf("uploads/images/%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	relativePath := fmt.Sprintf("/%s/%s", datePath, filename)

	// Store the file using Goravel's filesystem (saves to configured local disk)
	storedPath, err := file.StoreAs(datePath, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to store file: %w", err)
	}
	_ = storedPath

	// Create DB record
	image := models.Image{
		ID:     id,
		UserID: &userID,
		Url:    relativePath,
		Name:   file.GetClientOriginalName(),
		Mime:   ext,
		Size:   int(size),
		Role:   role,
	}

	if err := facades.Orm().Query().Create(&image); err != nil {
		return nil, fmt.Errorf("failed to create image record: %w", err)
	}

	return &UploadImageResult{
		ID:   image.ID,
		URL:  helpers.ImageServeURL(image.ID),
		Name: image.Name,
	}, nil
}

// GetImagePath looks up an image by ID and returns the absolute file path and content type.
func GetImagePath(imageID string) (string, string, error) {
	var image models.Image
	if err := facades.Orm().Query().Where("id", imageID).First(&image); err != nil {
		return "", "", fmt.Errorf("failed to find image: %w", err)
	}
	if image.ID == "" {
		return "", "", ErrImageNotFound
	}

	// Determine content type from stored mime
	contentType := "application/octet-stream"
	switch strings.ToLower(image.Mime) {
	case "jpg", "jpeg":
		contentType = "image/jpeg"
	case "png":
		contentType = "image/png"
	}

	// Resolve absolute path from storage root
	storageRoot := facades.Config().Env("STORAGE_PATH", "storage/app").(string)
	// relativePath is like /uploads/images/2026/03/21/xxx.jpg — strip leading /
	absPath := filepath.Join(storageRoot, strings.TrimPrefix(image.Url, "/"))

	return absPath, contentType, nil
}

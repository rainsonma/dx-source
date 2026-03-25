package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/goravel/framework/facades"
	qrcode "github.com/skip2/go-qrcode"

	"dx-api/app/consts"
	"dx-api/app/models"
)

// GenerateGroupQRCode creates a QR code PNG for the given URL, saves it to disk
// and creates an Image record. Returns the image ID.
func GenerateGroupQRCode(ownerID, inviteURL string) (string, error) {
	// Generate QR code PNG bytes
	png, err := qrcode.Encode(inviteURL, qrcode.Medium, 512)
	if err != nil {
		return "", fmt.Errorf("failed to generate qrcode: %w", err)
	}

	// Save to disk
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

	// Create DB record
	relativePath := fmt.Sprintf("/%s/%s", datePath, filename)
	image := models.Image{
		ID:     id,
		UserID: &ownerID,
		Url:    relativePath,
		Name:   "group-invite-qrcode.png",
		Mime:   "png",
		Size:   len(png),
		Role:   consts.ImageRoleGroupQrcode,
	}
	if err := facades.Orm().Query().Create(&image); err != nil {
		return "", fmt.Errorf("failed to create qrcode image record: %w", err)
	}

	return id, nil
}

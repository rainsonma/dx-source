package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
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

	storageRoot := StoragePath()
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

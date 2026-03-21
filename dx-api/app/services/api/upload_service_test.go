package api

import (
	"errors"
	"testing"
	"time"

	"github.com/goravel/framework/contracts/filesystem"
)

// mockFile implements filesystem.File for testing validation.
type mockFile struct {
	size     int64
	mimeType string
	name     string
}

func (f *mockFile) Disk(disk string) filesystem.File                { return f }
func (f *mockFile) Extension() (string, error)                     { return "", nil }
func (f *mockFile) File() string                                   { return "" }
func (f *mockFile) GetClientOriginalName() string                  { return f.name }
func (f *mockFile) GetClientOriginalExtension() string             { return "" }
func (f *mockFile) HashName(path ...string) string                 { return "" }
func (f *mockFile) LastModified() (time.Time, error)               { return time.Time{}, nil }
func (f *mockFile) MimeType() (string, error)                      { return f.mimeType, nil }
func (f *mockFile) Size() (int64, error)                           { return f.size, nil }
func (f *mockFile) Store(path string) (string, error)              { return "", nil }
func (f *mockFile) StoreAs(path, name string) (string, error)      { return "", nil }

func TestValidateUploadFile(t *testing.T) {
	tests := []struct {
		name    string
		size    int64
		mime    string
		role    string
		wantErr error
	}{
		{"valid jpeg avatar", 1024, "image/jpeg", "user-avatar", nil},
		{"valid png game cover", 500_000, "image/png", "game-cover", nil},
		{"exact 2MB limit", maxFileSize, "image/jpeg", "user-avatar", nil},
		{"exceeds 2MB", maxFileSize + 1, "image/jpeg", "user-avatar", ErrFileTooLarge},
		{"invalid mime gif", 1024, "image/gif", "user-avatar", ErrInvalidFileType},
		{"invalid mime webp", 1024, "image/webp", "user-avatar", ErrInvalidFileType},
		{"invalid mime svg", 1024, "image/svg+xml", "user-avatar", ErrInvalidFileType},
		{"empty mime", 1024, "", "user-avatar", ErrInvalidFileType},
		{"invalid role", 1024, "image/jpeg", "invalid-role", ErrInvalidImageRole},
		{"empty role", 1024, "image/jpeg", "", ErrInvalidImageRole},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &mockFile{size: tt.size, mimeType: tt.mime, name: "test.jpg"}
			err := ValidateUploadFile(file, tt.role)

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateUploadFile_AllRoles(t *testing.T) {
	for role := range validRoles {
		t.Run(role, func(t *testing.T) {
			file := &mockFile{size: 1024, mimeType: "image/jpeg", name: "test.jpg"}
			if err := ValidateUploadFile(file, role); err != nil {
				t.Errorf("role %q should be valid, got %v", role, err)
			}
		})
	}
}

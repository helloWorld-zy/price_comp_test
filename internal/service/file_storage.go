package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"crypto/sha256"
	"encoding/hex"
)

// FileStorageService handles file upload and storage
type FileStorageService struct {
	uploadDir string
}

// NewFileStorageService creates a new file storage service
func NewFileStorageService(uploadDir string) *FileStorageService {
	return &FileStorageService{uploadDir: uploadDir}
}

// UploadFile stores an uploaded file and returns the file path and hash
func (s *FileStorageService) UploadFile(ctx context.Context, filename string, content io.Reader) (string, string, int64, error) {
	// Ensure upload directory exists
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return "", "", 0, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename with timestamp
	timestamp := time.Now().Format("20060102150405")
	ext := filepath.Ext(filename)
	baseName := filename[:len(filename)-len(ext)]
	uniqueFilename := fmt.Sprintf("%s_%s%s", baseName, timestamp, ext)
	filePath := filepath.Join(s.uploadDir, uniqueFilename)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Calculate hash while copying
	hasher := sha256.New()
	multiWriter := io.MultiWriter(file, hasher)

	size, err := io.Copy(multiWriter, content)
	if err != nil {
		os.Remove(filePath)
		return "", "", 0, fmt.Errorf("failed to write file: %w", err)
	}

	hash := hex.EncodeToString(hasher.Sum(nil))

	return filePath, hash, size, nil
}

// GetFilePath returns the full path for a stored file
func (s *FileStorageService) GetFilePath(filename string) string {
	return filepath.Join(s.uploadDir, filename)
}

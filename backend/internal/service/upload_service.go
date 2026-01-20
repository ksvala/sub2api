package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxUploadSizeBytes = 5 << 20 // 5MB
)

type UploadService struct {
	baseDir string
}

func NewUploadService() *UploadService {
	baseDir := filepath.Join(resolveDataDir(), "uploads")
	return &UploadService{baseDir: baseDir}
}

func (s *UploadService) SaveImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	_ = ctx
	if file == nil {
		return "", fmt.Errorf("file is required")
	}
	if file.Size <= 0 || file.Size > maxUploadSizeBytes {
		return "", fmt.Errorf("file size exceeds limit")
	}

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(file.Filename))
	}
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("unsupported file type")
	}

	if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
		return "", fmt.Errorf("prepare upload dir: %w", err)
	}

	name, err := randomFileName(file.Filename)
	if err != nil {
		return "", err
	}
	path := filepath.Join(s.baseDir, name)

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = src.Close() }()

	buf := make([]byte, 512)
	if _, err := src.Read(buf); err != nil && err != io.EOF {
		return "", fmt.Errorf("read file: %w", err)
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("reset file: %w", err)
	}

	detected := http.DetectContentType(buf)
	if !strings.HasPrefix(detected, "image/") {
		return "", fmt.Errorf("unsupported file type")
	}

	fileHandle, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer func() { _ = fileHandle.Close() }()

	if _, err := io.Copy(fileHandle, src); err != nil {
		return "", fmt.Errorf("save file: %w", err)
	}

	return "/uploads/" + name, nil
}

func randomFileName(original string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate file name: %w", err)
	}
	ext := strings.ToLower(filepath.Ext(original))
	if ext == "" {
		ext = ".png"
	}
	return hex.EncodeToString(bytes) + ext, nil
}

func resolveDataDir() string {
	if value := strings.TrimSpace(os.Getenv("DATA_DIR")); value != "" {
		return value
	}

	dockerDataDir := "/app/data"
	if info, err := os.Stat(dockerDataDir); err == nil && info.IsDir() {
		testFile := filepath.Join(dockerDataDir, ".write_test")
		if err := os.WriteFile(testFile, []byte("test"), 0o644); err == nil {
			_ = os.Remove(testFile)
			return dockerDataDir
		}
	}

	return "."
}

package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/webp"
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
	if strings.Contains(strings.ToLower(contentType), "svg") {
		return "", fmt.Errorf("unsupported file type")
	}
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("unsupported file type")
	}

	if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
		return "", fmt.Errorf("prepare upload dir: %w", err)
	}

	name, err := randomFileNameWithExt(".png")
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
	if strings.Contains(strings.ToLower(detected), "svg") {
		return "", fmt.Errorf("unsupported file type")
	}
	if !strings.HasPrefix(detected, "image/") {
		return "", fmt.Errorf("unsupported file type")
	}
	if detected != "image/png" && detected != "image/jpeg" && detected != "image/webp" {
		return "", fmt.Errorf("unsupported file type")
	}

	var img image.Image
	if detected == "image/webp" {
		img, err = webp.Decode(src)
	} else {
		img, _, err = image.Decode(src)
	}
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}

	fileHandle, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer func() { _ = fileHandle.Close() }()

	if err := png.Encode(fileHandle, img); err != nil {
		return "", fmt.Errorf("encode image: %w", err)
	}

	return "/uploads/" + name, nil
}

func randomFileNameWithExt(ext string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate file name: %w", err)
	}
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		ext = ".png"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
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

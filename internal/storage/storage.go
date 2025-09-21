package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type FileStorage struct {
	basePath string
}

func NewFileStorage(basePath string) (*FileStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &FileStorage{basePath: basePath}, nil
}

func (fs *FileStorage) SaveFile(fileHeader *multipart.FileHeader, projectID, taskID uint) (string, error) {
	dir := filepath.Join(fs.basePath, fmt.Sprintf("project_%d", projectID), fmt.Sprintf("task_%d", taskID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	filename := fmt.Sprintf("%d_%s", taskID, fileHeader.Filename)
	fullPath := filepath.Join(dir, filename)

	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	relativePath := filepath.Join(fmt.Sprintf("project_%d", projectID), fmt.Sprintf("task_%d", taskID), filename)
	return relativePath, nil
}

func (fs *FileStorage) GetFile(path string) (*os.File, error) {
	fullPath := filepath.Join(fs.basePath, path)
	return os.Open(fullPath)
}

func (fs *FileStorage) DeleteFile(path string) error {
	fullPath := filepath.Join(fs.basePath, path)
	return os.Remove(fullPath)
}

func (fs *FileStorage) GetFilePath(path string) string {
	return filepath.Join(fs.basePath, path)
}
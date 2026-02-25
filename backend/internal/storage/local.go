package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

func (l *LocalStorage) fullPath(objectKey string) string {
	return filepath.Join(l.basePath, objectKey)
}

func (l *LocalStorage) Upload(
	ctx context.Context,
	objectKey string,
	r io.Reader,
	size int64,
	contentType string,
) error {

	path := l.fullPath(objectKey)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, r)
	return err
}

func (l *LocalStorage) Download(
	ctx context.Context,
	objectKey string,
) (io.ReadCloser, error) {

	return os.Open(l.fullPath(objectKey))
}

func (l *LocalStorage) Delete(
	ctx context.Context,
	objectKey string,
) error {
	return os.Remove(l.fullPath(objectKey))
}

func (l *LocalStorage) Exists(
	ctx context.Context,
	objectKey string,
) (bool, error) {

	_, err := os.Stat(l.fullPath(objectKey))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

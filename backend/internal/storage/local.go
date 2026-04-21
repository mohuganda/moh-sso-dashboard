package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	basePath      string
	publicBaseURL string
}

func NewLocalStorage(basePath, publicBaseURL string) *LocalStorage {
	return &LocalStorage{
		basePath:      basePath,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}
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

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
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

func (l *LocalStorage) GetObjectURL(ctx context.Context, objectKey string) (string, error) {
	if l.publicBaseURL == "" {
		return "", fmt.Errorf("public base url is required for local storage object url")
	}

	return fmt.Sprintf(
		"%s/api/v1/documents/files/%s",
		l.publicBaseURL,
		url.PathEscape(objectKey),
	), nil
}

func (l *LocalStorage) GetViewURL(
	ctx context.Context,
	objectKey string,
	filename string,
) (string, error) {
	if l.publicBaseURL == "" {
		return "", fmt.Errorf("public base url is required for local storage view url")
	}

	escapedKey := url.PathEscape(objectKey)
	q := url.Values{}
	if strings.TrimSpace(filename) != "" {
		q.Set("filename", filename)
	}

	u := fmt.Sprintf("%s/api/v1/documents/files/%s/view", l.publicBaseURL, escapedKey)
	if encoded := q.Encode(); encoded != "" {
		u += "?" + encoded
	}

	return u, nil
}

func (l *LocalStorage) GetDownloadURL(
	ctx context.Context,
	objectKey string,
	filename string,
) (string, error) {
	if l.publicBaseURL == "" {
		return "", fmt.Errorf("public base url is required for local storage download url")
	}

	escapedKey := url.PathEscape(objectKey)
	q := url.Values{}
	if strings.TrimSpace(filename) != "" {
		q.Set("filename", filename)
	}

	u := fmt.Sprintf("%s/api/v1/documents/files/%s/download", l.publicBaseURL, escapedKey)
	if encoded := q.Encode(); encoded != "" {
		u += "?" + encoded
	}

	return u, nil
}

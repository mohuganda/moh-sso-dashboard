package storage

import (
	"context"
	"io"
)

type Storage interface {
	Upload(ctx context.Context, objectKey string, r io.Reader, size int64, contentType string) error
	Download(ctx context.Context, objectKey string) (io.ReadCloser, error)
	Delete(ctx context.Context, objectKey string) error
	Exists(ctx context.Context, objectKey string) (bool, error)
}

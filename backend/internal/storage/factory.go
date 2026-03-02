package storage

import (
	"fmt"
	"strings"

	"github.com/moh-sso-dashboard/internal/config"
)

const (
	ProviderLocal = "local"
	ProviderNFS   = "nfs"
	ProviderS3    = "s3"
	ProviderMinIO = "minio"
)

func NewFileStorage(provider string, cfg *config.Config) (Storage, error) {
	p := strings.ToLower(strings.TrimSpace(provider))
	switch p {
	case ProviderLocal:
		if cfg.LocalBasePath == "" {
			return nil, fmt.Errorf("LOCAL_BASE_PATH is required for local storage")
		}
		return NewLocalStorage(cfg.LocalBasePath), nil

	case ProviderNFS:
		if cfg.NFSBasePath == "" {
			return nil, fmt.Errorf("NFS_BASE_PATH is required for nfs storage")
		}
		// NFS can be mounted to filesystem; use local implementation
		return NewLocalStorage(cfg.NFSBasePath), nil

	case ProviderS3:
		// AWS S3 (or any S3-compatible without custom endpoint)
		if cfg.S3Bucket == "" {
			return nil, fmt.Errorf("S3_BUCKET is required for s3 storage")
		}
		if cfg.S3Region == "" {
			return nil, fmt.Errorf("S3_REGION is required for s3 storage")
		}
		// Access/secret can be optional if using IAM role / workload identity.
		return NewS3Storage(S3Config{
			Region:          cfg.S3Region,
			Bucket:          cfg.S3Bucket,
			AccessKeyID:     cfg.S3AccessKeyID,
			SecretAccessKey: cfg.S3SecretAccessKey,
			Endpoint:        "", // empty => AWS default resolver
			UsePathStyle:    false,
		}), nil

	case ProviderMinIO:
		// MinIO uses S3 API but needs custom endpoint + path-style often
		if cfg.MinioEndpoint == "" {
			return nil, fmt.Errorf("MINIO_ENDPOINT is required for minio storage")
		}
		if cfg.MinioBucket == "" {
			return nil, fmt.Errorf("MINIO_BUCKET is required for minio storage")
		}
		// MinIO typically uses explicit creds
		if cfg.MinioAccessKeyID == "" || cfg.MinioSecretAccessKey == "" {
			return nil, fmt.Errorf("MINIO_ACCESS_KEY_ID and MINIO_SECRET_ACCESS_KEY are required for minio storage")
		}

		return NewS3Storage(S3Config{
			Region:          cfg.MinioRegion, // can be "us-east-1" if you don’t care
			Bucket:          cfg.MinioBucket,
			AccessKeyID:     cfg.MinioAccessKeyID,
			SecretAccessKey: cfg.MinioSecretAccessKey,
			Endpoint:        cfg.MinioEndpoint, // e.g. http://minio:9000
			UsePathStyle:    true,              // usually required for MinIO
		}), nil

	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", provider)
	}
}

// ------------------------------------
// ✅ Minimal addition: StorageFactory + Get
// ------------------------------------
type StorageFactory struct {
	cfg   *config.Config
	cache map[string]Storage
}

func NewStorageFactory(cfg *config.Config) *StorageFactory {
	return &StorageFactory{
		cfg:   cfg,
		cache: make(map[string]Storage),
	}
}

func (f *StorageFactory) Get(provider string) (Storage, error) {
	key := strings.ToLower(strings.TrimSpace(provider))
	if key == "" {
		return nil, fmt.Errorf("storage provider is required")
	}

	// cached?
	if s, ok := f.cache[key]; ok && s != nil {
		return s, nil
	}

	// build using your existing function
	s, err := NewFileStorage(key, f.cfg)
	if err != nil {
		return nil, err
	}

	f.cache[key] = s
	return s, nil
}

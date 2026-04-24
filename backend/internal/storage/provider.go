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

func normalizeProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

func NewFileStorage(provider string, cfg *config.Config) (Storage, error) {
	p := normalizeProvider(provider)

	switch p {
	case ProviderLocal:
		if strings.TrimSpace(cfg.LocalBasePath) == "" {
			return nil, fmt.Errorf("LOCAL_BASE_PATH is required for local storage")
		}
		if strings.TrimSpace(cfg.AppBaseURL) == "" {
			return nil, fmt.Errorf("APP_BASE_URL is required for local storage")
		}

		return NewLocalStorage(cfg.LocalBasePath, cfg.AppBaseURL), nil

	case ProviderNFS:
		if strings.TrimSpace(cfg.NFSBasePath) == "" {
			return nil, fmt.Errorf("NFS_BASE_PATH is required for nfs storage")
		}
		if strings.TrimSpace(cfg.AppBaseURL) == "" {
			return nil, fmt.Errorf("APP_BASE_URL is required for nfs storage")
		}

		// NFS is mounted on the filesystem, so reuse the local implementation.
		return NewLocalStorage(cfg.NFSBasePath, cfg.AppBaseURL), nil

	case ProviderS3:
		if strings.TrimSpace(cfg.S3Bucket) == "" {
			return nil, fmt.Errorf("S3_BUCKET is required for s3 storage")
		}
		if strings.TrimSpace(cfg.S3Region) == "" {
			return nil, fmt.Errorf("S3_REGION is required for s3 storage")
		}

		return NewS3Storage(S3Config{
			Region:          cfg.S3Region,
			Bucket:          cfg.S3Bucket,
			AccessKeyID:     cfg.S3AccessKeyID,
			SecretAccessKey: cfg.S3SecretAccessKey,
			Endpoint:        "",
			UsePathStyle:    false,
		})

	case ProviderMinIO:
		if strings.TrimSpace(cfg.MinioEndpoint) == "" {
			return nil, fmt.Errorf("MINIO_ENDPOINT is required for minio storage")
		}
		if strings.TrimSpace(cfg.MinioBucket) == "" {
			return nil, fmt.Errorf("MINIO_BUCKET is required for minio storage")
		}
		if strings.TrimSpace(cfg.MinioAccessKeyID) == "" || strings.TrimSpace(cfg.MinioSecretAccessKey) == "" {
			return nil, fmt.Errorf("MINIO_ACCESS_KEY_ID and MINIO_SECRET_ACCESS_KEY are required for minio storage")
		}

		region := strings.TrimSpace(cfg.MinioRegion)
		if region == "" {
			region = "us-east-1"
		}

		return NewS3Storage(S3Config{
			Region:          region,
			Bucket:          cfg.MinioBucket,
			AccessKeyID:     cfg.MinioAccessKeyID,
			SecretAccessKey: cfg.MinioSecretAccessKey,
			Endpoint:        cfg.MinioEndpoint,
			UsePathStyle:    true,
		})

	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", provider)
	}
}

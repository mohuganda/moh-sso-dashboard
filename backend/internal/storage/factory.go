package storage

import (
	"fmt"

	"github.com/moh-sso-dashboard/internal/config"
)

func NewFileStorage(provider string, cfg *config.Config) (Storage, error) {
	switch provider {
	case "local":
		if cfg.LocalBasePath == "" {
			return nil, fmt.Errorf("LOCAL_BASE_PATH is required for local storage")
		}
		return NewLocalStorage(cfg.LocalBasePath), nil

	case "nfs":
		if cfg.NFSBasePath == "" {
			return nil, fmt.Errorf("NFS_BASE_PATH is required for nfs storage")
		}
		return NewLocalStorage(cfg.NFSBasePath), nil

	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", provider)
	}
}

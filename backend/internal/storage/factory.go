package storage

import "github.com/moh-sso-dashboard/internal/config"

func New(provider string, cfg config.Config) Storage {

	switch provider {
	case "local":
		return NewLocalStorage(cfg.LocalBasePath)

	default:
		panic("unsupported storage provider")
	}
}

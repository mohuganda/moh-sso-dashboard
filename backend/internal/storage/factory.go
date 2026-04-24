package storage

import (
	"fmt"
	"sync"

	"github.com/moh-sso-dashboard/internal/config"
)

type StorageFactory struct {
	cfg   *config.Config
	mu    sync.RWMutex
	cache map[string]Storage
}

func NewStorageFactory(cfg *config.Config) *StorageFactory {
	return &StorageFactory{
		cfg:   cfg,
		cache: make(map[string]Storage),
	}
}

func (f *StorageFactory) Get(provider string) (Storage, error) {
	key := normalizeProvider(provider)
	if key == "" {
		return nil, fmt.Errorf("storage provider is required")
	}

	f.mu.RLock()
	if s, ok := f.cache[key]; ok && s != nil {
		f.mu.RUnlock()
		return s, nil
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	if s, ok := f.cache[key]; ok && s != nil {
		return s, nil
	}

	s, err := NewFileStorage(key, f.cfg)
	if err != nil {
		return nil, err
	}

	f.cache[key] = s
	return s, nil
}

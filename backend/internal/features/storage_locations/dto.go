package storage_locations

import "time"

type createStorageLocationRequest struct {
	Code     string `json:"code" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Provider string `json:"provider"`
	BaseUri  string `json:"base_uri"`
	IsActive bool   `json:"is_active" binding:"required"`
}

type StorageLocationResponse struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Provider  string     `json:"provider"`
	BaseUri   string     `json:"base_uri"`
	IsActive  bool       `json:"is_active"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type deleteStorageLocationResponse struct {
	Message string `json:"message"`
}

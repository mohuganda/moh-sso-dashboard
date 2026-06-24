package storage_locations

import db "github.com/moh-sso-dashboard/internal/db/sqlc"

func toStorageLocationResponse(location db.StorageLocation) StorageLocationResponse {
	out := StorageLocationResponse{
		ID:       location.ID.String(),
		Code:     location.Code,
		Name:     location.Name,
		Provider: location.Provider,
		BaseUri:  location.BaseUri,
	}
	if location.IsActive.Valid {
		out.IsActive = location.IsActive.Bool
	}
	if location.CreatedAt.Valid {
		createdAt := location.CreatedAt.Time
		out.CreatedAt = &createdAt
	}
	return out
}

func toStorageLocationResponses(locations []db.StorageLocation) []StorageLocationResponse {
	out := make([]StorageLocationResponse, 0, len(locations))
	for _, location := range locations {
		out = append(out, toStorageLocationResponse(location))
	}
	return out
}

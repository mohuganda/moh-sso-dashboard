package documents

import (
	"time"

	"github.com/google/uuid"
)

const (
	DocumentStatusPending   = "PENDING"
	DocumentStatusCompleted = "COMPLETED"
)

type UpdateDocumentRequest struct {
	OriginalFilename *string `json:"original_filename"`
	ContentType      *string `json:"content_type"`
}

type DocumentResponse struct {
	ID               uuid.UUID              `json:"id"`
	OriginalFilename string                 `json:"original_filename"`
	ContentType      string                 `json:"content_type"`
	SizeBytes        int64                  `json:"size_bytes"`
	ChecksumSHA256   *string                `json:"checksum_sha256,omitempty"`
	StorageLocation  uuid.UUID              `json:"storage_location"`
	ObjectKey        string                 `json:"object_key"`
	UploadedBy       uuid.UUID              `json:"uploaded_by"`
	Status           string                 `json:"status"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	ObjectURL        string                 `json:"object_url,omitempty"`
	ViewURL          string                 `json:"view_url,omitempty"`
	DownloadURL      string                 `json:"download_url,omitempty"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	HealthContextID  *uuid.UUID             `json:"health_context_id,omitempty"`
}

type ProcessResponse struct {
	ID          uuid.UUID  `json:"id"`
	DocumentID  uuid.UUID  `json:"document_id"`
	ProcessType string     `json:"process_type"`
	Status      string     `json:"status"`
	Progress    int32      `json:"progress"`
	Message     *string    `json:"message,omitempty"`
	Error       *string    `json:"error,omitempty"`
	Attempts    int32      `json:"attempts"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

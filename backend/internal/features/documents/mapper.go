package documents

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

func toDocumentResponse(doc db.Document, objectURL, viewURL, downloadURL string) DocumentResponse {
	return DocumentResponse{
		ID:               doc.ID,
		OriginalFilename: doc.OriginalFilename,
		ContentType:      nullStringValue(doc.ContentType),
		SizeBytes:        doc.SizeBytes,
		ChecksumSHA256:   nullStringPtr(doc.ChecksumSha256),
		StorageLocation:  doc.StorageLocationID,
		ObjectKey:        doc.ObjectKey,
		UploadedBy:       doc.UploadedBy,
		Status:           normalizeStatus(doc.Status),
		ObjectURL:        objectURL,
		ViewURL:          viewURL,
		DownloadURL:      downloadURL,
		CreatedAt:        doc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        doc.UpdatedAt.Format(time.RFC3339),
	}
}

func toProcessResponse(process db.Process) ProcessResponse {
	return ProcessResponse{
		ID:          process.ID,
		DocumentID:  process.DocumentID,
		ProcessType: process.ProcessType,
		Status:      normalizeStatus(process.Status),
		Progress:    process.Progress,
		Message:     nullStringPtr(process.Message),
		Error:       nullStringPtr(process.Error),
		Attempts:    process.Attempts,
		CreatedBy:   process.CreatedBy,
		StartedAt:   nullTimePtr(process.StartedAt),
		FinishedAt:  nullTimePtr(process.FinishedAt),
		CreatedAt:   nullTimePtr(process.CreatedAt),
		UpdatedAt:   nullTimePtr(process.UpdatedAt),
	}
}

func normalizeStatus(v any) string {
	switch s := v.(type) {
	case nil:
		return ""
	case string:
		return s
	case []byte:
		return string(s)
	case fmt.Stringer:
		return s.String()
	default:
		return fmt.Sprint(v)
	}
}

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func nullTimePtr(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

func nullStringValue(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func requiresProcessing(mimeType, fileName string) bool {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "text/csv",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return true
	case "application/pdf":
		return false
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".csv", ".xls", ".xlsx":
		return true
	case ".pdf":
		return false
	default:
		return false
	}
}

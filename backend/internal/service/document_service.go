package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	models "github.com/moh-sso-dashboard/internal/model"
	documentRepo "github.com/moh-sso-dashboard/internal/repository/document"
	processRepo "github.com/moh-sso-dashboard/internal/repository/processes"
	"github.com/moh-sso-dashboard/internal/storage"
	"github.com/moh-sso-dashboard/internal/utils"
)

type CreateDocumentInput struct {
	OriginalFilename string
	ContentType      string
	SizeBytes        int64
	ChecksumSHA256   *string
	StorageLocation  uuid.UUID
	ObjectKey        string
	UploadedBy       uuid.UUID
	ProcessType      models.ProcessType
	Status           db.DocumentStatus
	IsTemplate       bool
	Metadata         map[string]any
}

type EditDocumentInput struct {
	ID               uuid.UUID
	OriginalFilename string
	ContentType      string
}

type DocumentService struct {
	repo          documentRepo.DocumentRepository
	processRepo   processRepo.ProcessRepository
	notifications NotificationsService
	storage       storage.Storage
}

func NewDocumentService(
	repo documentRepo.DocumentRepository,
	processRepo processRepo.ProcessRepository,
	notifications NotificationsService,
	storage storage.Storage,
) *DocumentService {
	return &DocumentService{
		repo:          repo,
		processRepo:   processRepo,
		notifications: notifications,
		storage:       storage,
	}
}

func (s *DocumentService) CreateDocument(
	ctx context.Context,
	input CreateDocumentInput,
) (db.Document, error) {
	if s == nil {
		return db.Document{}, errors.New("document service is nil")
	}

	if s.repo == nil {
		return db.Document{}, errors.New("document repository is nil")
	}

	if s.processRepo == nil {
		return db.Document{}, errors.New("process repository is nil")
	}

	if strings.TrimSpace(input.OriginalFilename) == "" {
		return db.Document{}, errors.New("original filename is required")
	}

	if input.StorageLocation == uuid.Nil {
		return db.Document{}, errors.New("storage location is required")
	}

	if input.UploadedBy == uuid.Nil {
		return db.Document{}, errors.New("uploaded by is required")
	}

	if strings.TrimSpace(input.ObjectKey) == "" {
		return db.Document{}, errors.New("object key is required")
	}

	docID := uuid.New()

	var checksum sql.NullString
	if input.ChecksumSHA256 != nil && strings.TrimSpace(*input.ChecksumSHA256) != "" {
		checksum = sql.NullString{
			String: strings.TrimSpace(*input.ChecksumSHA256),
			Valid:  true,
		}
	}

	needsProcessing := requiresProcessing(input.ContentType, input.OriginalFilename)

	status := input.Status
	if status == "" {
		if needsProcessing && !input.IsTemplate {
			status = db.DocumentStatusPENDING
		} else {
			status = db.DocumentStatusCOMPLETED
		}
	}

	if needsProcessing && !input.IsTemplate && !input.ProcessType.IsValid() {
		return db.Document{}, errors.New("invalid process type for processable document")
	}

	metadata := json.RawMessage([]byte(`{}`))
	if len(input.Metadata) > 0 {
		raw, err := json.Marshal(input.Metadata)
		if err != nil {
			return db.Document{}, fmt.Errorf("marshal document metadata: %w", err)
		}

		metadata = raw
	}

	doc, err := s.repo.CreateDocument(
		ctx,
		db.CreateDocumentParams{
			ID:               docID,
			OriginalFilename: strings.TrimSpace(input.OriginalFilename),
			ContentType: sql.NullString{
				String: strings.TrimSpace(input.ContentType),
				Valid:  strings.TrimSpace(input.ContentType) != "",
			},
			SizeBytes:         input.SizeBytes,
			ChecksumSha256:    checksum,
			StorageLocationID: input.StorageLocation,
			ObjectKey:         strings.TrimSpace(input.ObjectKey),
			UploadedBy:        input.UploadedBy,
			Status:            status,
			Metadata:          metadata,
			IsTemplate:        input.IsTemplate,
		},
	)
	if err != nil {
		return db.Document{}, fmt.Errorf("create document: %w", err)
	}

	if needsProcessing && !input.IsTemplate {
		_, err = s.processRepo.CreateProcess(ctx, db.CreateProcessParams{
			ID:          uuid.New(),
			DocumentID:  doc.ID,
			ProcessType: string(input.ProcessType),
			CreatedBy:   input.UploadedBy,
		})
		if err != nil {
			return db.Document{}, fmt.Errorf("create document process: %w", err)
		}
	}

	nt := models.DocumentCreated
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    buildDocumentCreatedMessage(doc, needsProcessing, input.IsTemplate),
		TargetRole: "admin",
		UserID:     input.UploadedBy.String(),
		Metadata: utils.MustJSON(map[string]any{
			"document_id":       doc.ID.String(),
			"filename":          doc.OriginalFilename,
			"content_type":      nullStringValue(doc.ContentType),
			"size_bytes":        doc.SizeBytes,
			"status":            string(doc.Status),
			"object_key":        doc.ObjectKey,
			"uploaded_by":       input.UploadedBy.String(),
			"is_template":       doc.IsTemplate,
			"needs_processing":  needsProcessing,
			"process_type":      string(input.ProcessType),
			"storage_location":  input.StorageLocation.String(),
			"checksum_provided": checksum.Valid,
		}),
	})

	return doc, nil
}

func (s *DocumentService) GetDocument(
	ctx context.Context,
	id uuid.UUID,
) (db.Document, error) {
	if s == nil {
		return db.Document{}, errors.New("document service is nil")
	}

	if s.repo == nil {
		return db.Document{}, errors.New("document repository is nil")
	}

	doc, err := s.repo.GetDocument(ctx, id)
	if err != nil {
		return db.Document{}, fmt.Errorf("get document: %w", err)
	}

	return doc, nil
}

func (s *DocumentService) EditDocument(
	ctx context.Context,
	input EditDocumentInput,
) (db.Document, error) {
	if s == nil {
		return db.Document{}, errors.New("document service is nil")
	}

	if s.repo == nil {
		return db.Document{}, errors.New("document repository is nil")
	}

	if input.ID == uuid.Nil {
		return db.Document{}, errors.New("document id is required")
	}

	if strings.TrimSpace(input.OriginalFilename) == "" {
		return db.Document{}, errors.New("original filename is required")
	}

	doc, err := s.repo.EditDocument(ctx, db.UpdateDocumentParams{
		ID:               input.ID,
		OriginalFilename: strings.TrimSpace(input.OriginalFilename),
		ContentType: sql.NullString{
			String: strings.TrimSpace(input.ContentType),
			Valid:  strings.TrimSpace(input.ContentType) != "",
		},
	})
	if err != nil {
		return db.Document{}, fmt.Errorf("edit document: %w", err)
	}

	nt := models.DocumentEdited
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Document edited",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"document_id":  doc.ID.String(),
			"filename":     doc.OriginalFilename,
			"content_type": nullStringValue(doc.ContentType),
			"status":       string(doc.Status),
		}),
	})

	return doc, nil
}

func (s *DocumentService) DeleteDocument(
	ctx context.Context,
	id uuid.UUID,
) error {
	if s == nil {
		return errors.New("document service is nil")
	}

	if s.repo == nil {
		return errors.New("document repository is nil")
	}

	if id == uuid.Nil {
		return errors.New("document id is required")
	}

	doc, err := s.repo.GetDocument(ctx, id)
	if err != nil {
		return fmt.Errorf("get document before delete: %w", err)
	}

	if err := s.repo.DeleteDocument(ctx, id); err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	nt := models.DocumentDeleted
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      nt.Title(),
		Severity:   nt.Severity(),
		Message:    "Document deleted",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"document_id":  id.String(),
			"filename":     doc.OriginalFilename,
			"content_type": nullStringValue(doc.ContentType),
			"status":       string(doc.Status),
			"object_key":   doc.ObjectKey,
			"is_template":  doc.IsTemplate,
		}),
	})

	return nil
}

func (s *DocumentService) ListDocuments(
	ctx context.Context,
	page models.Pagination,
) ([]db.Document, error) {
	if s == nil {
		return nil, errors.New("document service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("document repository is nil")
	}

	return s.repo.ListDocuments(ctx, page)
}

func (s *DocumentService) ListProcessesByDocument(
	ctx context.Context,
	documentID string,
) ([]db.Process, error) {
	if s == nil {
		return nil, errors.New("document service is nil")
	}

	if s.repo == nil {
		return nil, errors.New("document repository is nil")
	}

	docUUID, err := uuid.Parse(documentID)
	if err != nil {
		return nil, fmt.Errorf("invalid document id: %w", err)
	}

	_, err = s.repo.GetDocument(ctx, docUUID)
	if err != nil {
		return nil, fmt.Errorf("get document before listing processes: %w", err)
	}

	return s.repo.ListProcessesByDocument(ctx, docUUID)
}

func (s *DocumentService) Reprocess(
	ctx context.Context,
	documentID uuid.UUID,
) error {
	if s == nil {
		return errors.New("document service is nil")
	}

	if s.repo == nil {
		return errors.New("document repository is nil")
	}

	if s.processRepo == nil {
		return errors.New("process repository is nil")
	}

	if documentID == uuid.Nil {
		return errors.New("document id is required")
	}

	doc, err := s.repo.GetDocument(ctx, documentID)
	if err != nil {
		return fmt.Errorf("get document before reprocess: %w", err)
	}

	if !requiresProcessing(
		nullStringValue(doc.ContentType),
		doc.OriginalFilename,
	) {
		return fmt.Errorf("document type does not support processing")
	}

	proc, err := s.repo.GetLatestByDocumentID(ctx, documentID)
	if err != nil {
		return fmt.Errorf("get latest process by document id: %w", err)
	}

	if proc.Status == models.ProcessStatusPROCESSING {
		return fmt.Errorf("document already processing")
	}

	_, err = s.repo.MarkDocumentPending(ctx, documentID)
	if err != nil {
		return fmt.Errorf("mark document pending: %w", err)
	}

	_, err = s.processRepo.CreateProcess(ctx, db.CreateProcessParams{
		ID:          uuid.New(),
		DocumentID:  documentID,
		ProcessType: proc.ProcessType,
		CreatedBy:   proc.CreatedBy,
	})
	if err != nil {
		return fmt.Errorf("create reprocess process: %w", err)
	}

	nt := models.DocumentEdited
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      "Document reprocessing queued",
		Severity:   nt.Severity(),
		Message:    "Document reprocessing queued",
		TargetRole: "admin",
		UserID:     proc.CreatedBy.String(),
		Metadata: utils.MustJSON(map[string]any{
			"document_id":  doc.ID.String(),
			"filename":     doc.OriginalFilename,
			"content_type": nullStringValue(doc.ContentType),
			"process_type": proc.ProcessType,
			"created_by":   proc.CreatedBy.String(),
		}),
	})

	return nil
}

func (s *DocumentService) MarkDocumentProcessing(
	ctx context.Context,
	documentID uuid.UUID,
) (db.Document, error) {
	if s == nil {
		return db.Document{}, errors.New("document service is nil")
	}

	if s.repo == nil {
		return db.Document{}, errors.New("document repository is nil")
	}

	doc, err := s.repo.MarkDocumentProcessing(ctx, documentID)
	if err != nil {
		return db.Document{}, fmt.Errorf("mark document processing: %w", err)
	}

	return doc, nil
}

func (s *DocumentService) MarkDocumentCompleted(
	ctx context.Context,
	documentID uuid.UUID,
) (db.Document, error) {
	if s == nil {
		return db.Document{}, errors.New("document service is nil")
	}

	if s.repo == nil {
		return db.Document{}, errors.New("document repository is nil")
	}

	doc, err := s.repo.MarkDocumentCompleted(ctx, documentID)
	if err != nil {
		return db.Document{}, fmt.Errorf("mark document completed: %w", err)
	}

	nt := models.DocumentEdited
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      "Document processing completed",
		Severity:   "info",
		Message:    "Document processing completed",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"document_id":  doc.ID.String(),
			"filename":     doc.OriginalFilename,
			"content_type": nullStringValue(doc.ContentType),
			"status":       string(doc.Status),
			"object_key":   doc.ObjectKey,
		}),
	})

	return doc, nil
}

func (s *DocumentService) MarkDocumentFailed(
	ctx context.Context,
	documentID uuid.UUID,
) (db.Document, error) {
	if s == nil {
		return db.Document{}, errors.New("document service is nil")
	}

	if s.repo == nil {
		return db.Document{}, errors.New("document repository is nil")
	}

	doc, err := s.repo.MarkDocumentFailed(ctx, documentID)
	if err != nil {
		return db.Document{}, fmt.Errorf("mark document failed: %w", err)
	}

	nt := models.DocumentDeleted
	s.notify(ctx, models.Notification{
		Type:       string(nt),
		Title:      "Document processing failed",
		Severity:   "critical",
		Message:    "Document processing failed",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"document_id":  doc.ID.String(),
			"filename":     doc.OriginalFilename,
			"content_type": nullStringValue(doc.ContentType),
			"status":       string(doc.Status),
			"object_key":   doc.ObjectKey,
		}),
	})

	return doc, nil
}

func (s *DocumentService) notify(
	ctx context.Context,
	notification models.Notification,
) {
	if s == nil || s.notifications == nil {
		return
	}

	if strings.TrimSpace(notification.TargetRole) == "" {
		notification.TargetRole = "admin"
	}

	if _, err := s.notifications.Notify(ctx, notification); err != nil {
		fmt.Printf("document notification failed type=%s error=%v\n", notification.Type, err)
	}
}

func buildDocumentCreatedMessage(
	doc db.Document,
	needsProcessing bool,
	isTemplate bool,
) string {
	if isTemplate {
		return "Document template created"
	}

	if needsProcessing {
		return "Document created and queued for processing"
	}

	return "Document created"
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

func nullStringValue(v sql.NullString) string {
	if !v.Valid {
		return ""
	}

	return v.String
}

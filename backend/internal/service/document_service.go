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
			return db.Document{}, err
		}

		metadata = raw
	}

	doc, err := s.repo.CreateDocument(
		ctx,
		db.CreateDocumentParams{
			ID:               docID,
			OriginalFilename: input.OriginalFilename,
			ContentType: sql.NullString{
				String: input.ContentType,
				Valid:  strings.TrimSpace(input.ContentType) != "",
			},
			SizeBytes:         input.SizeBytes,
			ChecksumSha256:    checksum,
			StorageLocationID: input.StorageLocation,
			ObjectKey:         input.ObjectKey,
			UploadedBy:        input.UploadedBy,
			Status:            status,
			Metadata:          metadata,
			IsTemplate:        input.IsTemplate,
		},
	)
	if err != nil {
		return db.Document{}, err
	}

	if needsProcessing && !input.IsTemplate {
		_, err = s.processRepo.CreateProcess(ctx, db.CreateProcessParams{
			ID:          uuid.New(),
			DocumentID:  doc.ID,
			ProcessType: string(input.ProcessType),
			CreatedBy:   input.UploadedBy,
		})
		if err != nil {
			return db.Document{}, err
		}
	}

	if s.notifications != nil {
		nt := models.DocumentCreated

		s.notifications.Notify(ctx, models.Notification{
			Type:     string(nt),
			Title:    nt.Title(),
			Severity: nt.Severity(),
			Message:  "Document created",
		})
	}

	return doc, nil
}

func (s *DocumentService) GetDocument(
	ctx context.Context,
	id uuid.UUID,
) (db.Document, error) {
	doc, err := s.repo.GetDocument(ctx, id)
	if err != nil {
		return db.Document{}, err
	}
	return doc, nil
}

func (s *DocumentService) EditDocument(
	ctx context.Context,
	input EditDocumentInput,
) (db.Document, error) {

	doc, err := s.repo.EditDocument(ctx, db.UpdateDocumentParams{
		ID:               input.ID,
		OriginalFilename: input.OriginalFilename,
		ContentType: sql.NullString{
			String: input.ContentType,
			Valid:  input.ContentType != "",
		},
	})
	if err != nil {
		return db.Document{}, err
	}

	nt := models.DocumentEdited
	if s.notifications != nil {
		s.notifications.Notify(ctx, models.Notification{
			Type:     string(nt),
			Title:    nt.Title(),
			Severity: nt.Severity(),
			Message:  "Document edited",
		})
	}

	return doc, nil
}

func (s *DocumentService) DeleteDocument(
	ctx context.Context,
	id uuid.UUID,
) error {

	err := s.repo.DeleteDocument(ctx, id)
	if err != nil {
		return err
	}

	nt := models.DocumentDeleted
	if s.notifications != nil {
		s.notifications.Notify(ctx, models.Notification{
			Type:       string(nt),
			Title:      nt.Title(),
			Severity:   nt.Severity(),
			Message:    "Document deleted",
			TargetRole: "admin",
		})
	}

	return nil
}

func (s *DocumentService) ListDocuments(
	ctx context.Context,
	page models.Pagination,
) ([]db.Document, error) {
	return s.repo.ListDocuments(ctx, page)
}

func (s *DocumentService) ListProcessesByDocument(
	ctx context.Context,
	documentID string,
) ([]db.Process, error) {

	docUUID, err := uuid.Parse(documentID)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.GetDocument(ctx, docUUID)
	if err != nil {
		return nil, err
	}

	return s.repo.ListProcessesByDocument(ctx, docUUID)
}

func (s *DocumentService) Reprocess(
	ctx context.Context,
	documentID uuid.UUID,
) error {

	doc, err := s.repo.GetDocument(ctx, documentID)
	if err != nil {
		return err
	}

	if !requiresProcessing(
		nullStringValue(doc.ContentType),
		doc.OriginalFilename,
	) {
		return fmt.Errorf("document type does not support processing")
	}

	proc, err := s.repo.GetLatestByDocumentID(ctx, documentID)
	if err != nil {
		return err
	}

	if proc.Status == models.ProcessStatusPROCESSING {
		return fmt.Errorf("document already processing")
	}

	_, err = s.repo.MarkDocumentPending(ctx, documentID)
	if err != nil {
		return err
	}

	_, err = s.processRepo.CreateProcess(ctx, db.CreateProcessParams{
		ID:          uuid.New(),
		DocumentID:  documentID,
		ProcessType: proc.ProcessType,
		CreatedBy:   proc.CreatedBy,
	})

	return err
}

func (s *DocumentService) MarkDocumentProcessing(
	ctx context.Context,
	documentID uuid.UUID,
) (db.Document, error) {
	return s.repo.MarkDocumentProcessing(ctx, documentID)
}

func (s *DocumentService) MarkDocumentCompleted(
	ctx context.Context,
	documentID uuid.UUID,
) (db.Document, error) {
	return s.repo.MarkDocumentCompleted(ctx, documentID)
}

func (s *DocumentService) MarkDocumentFailed(
	ctx context.Context,
	documentID uuid.UUID,
) (db.Document, error) {
	return s.repo.MarkDocumentFailed(ctx, documentID)
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

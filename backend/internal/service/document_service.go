package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
	models "github.com/moh-sso-dashboard/internal/model"
	repository "github.com/moh-sso-dashboard/internal/repository/document"
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
}

type EditDocumentInput struct {
	ID               uuid.UUID
	OriginalFilename string
	ContentType      string
}

type DocumentService struct {
	repo          repository.DocumentRepository
	notifications NotificationsService
	storage       storage.Storage
}

func NewDocumentService(repo repository.DocumentRepository,
	notifications NotificationsService, storage storage.Storage,
) *DocumentService {
	return &DocumentService{
		repo:          repo,
		notifications: notifications,
		storage:       storage,
	}
}

func (s *DocumentService) CreateDocument(
	ctx context.Context,
	input CreateDocumentInput,
) (db.Document, error) {

	docID := uuid.New()
	processID := uuid.New()

	var checksum sql.NullString
	if input.ChecksumSHA256 != nil {
		checksum = sql.NullString{
			String: *input.ChecksumSHA256,
			Valid:  true,
		}
	}

	doc, err := s.repo.CreateDocumentWithProcess(
		ctx,
		db.CreateDocumentParams{
			ID:               docID,
			OriginalFilename: input.OriginalFilename,
			ContentType: sql.NullString{
				String: input.ContentType,
				Valid:  true,
			},
			SizeBytes:         input.SizeBytes,
			ChecksumSha256:    checksum,
			StorageLocationID: input.StorageLocation,
			ObjectKey:         input.ObjectKey,
			UploadedBy:        input.UploadedBy,
		},
		db.CreateProcessParams{
			ID:          processID,
			DocumentID:  docID,
			ProcessType: string(models.ProcessTypeDocumentImport),
			CreatedBy:   input.UploadedBy,
		},
	)

	if err != nil {
		return db.Document{}, err
	}

	// notify AFTER success
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
			Message:    "Document  deleted",
			TargetRole: "admin",
		})
	}

	return nil
}

func (s *DocumentService) ListDocuments(
	ctx context.Context,
	page model.Pagination,
) ([]db.Document, error) {

	docs, err := s.repo.ListDocuments(ctx, page)
	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (s *DocumentService) ListProcessesByDocument(
	ctx context.Context,
	documentID string,
) ([]db.Process, error) {

	// 1️⃣ Validate UUID
	docUUID, err := uuid.Parse(documentID)
	if err != nil {
		return nil, err
	}

	// 2️⃣ (Optional but recommended) Ensure document exists
	_, err = s.repo.GetDocument(ctx, docUUID)
	if err != nil {
		return nil, err
	}

	// 3️⃣ Fetch processes
	processes, err := s.repo.ListProcessesByDocument(ctx, docUUID)
	if err != nil {
		return nil, err
	}

	return processes, nil
}

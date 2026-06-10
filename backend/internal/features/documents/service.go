package documents

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	models "github.com/moh-sso-dashboard/internal/model"
	processRepo "github.com/moh-sso-dashboard/internal/repository/processes"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
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

type Service struct {
	repo          Repository
	processRepo   processRepo.ProcessRepository
	notifications sharedservice.NotificationsService
	storage       storage.Storage
	cfg           *config.Config
}

func NewService(
	repo Repository,
	processRepo processRepo.ProcessRepository,
	notifications sharedservice.NotificationsService,
	storage storage.Storage,
	cfg ...*config.Config,
) *Service {
	var appConfig *config.Config
	if len(cfg) > 0 {
		appConfig = cfg[0]
	}

	return &Service{
		repo:          repo,
		processRepo:   processRepo,
		notifications: notifications,
		storage:       storage,
		cfg:           appConfig,
	}
}

func (s *Service) CreateDocument(
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
	notification := models.Notification{
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
	}

	// Email only when the document has been queued for processing.
	// Simple uploads and template uploads remain in-app only.
	if needsProcessing && !input.IsTemplate {
		s.attachAdminEmailDelivery(
			&notification,
			"document-created",
			"Document created and queued for processing",
			fmt.Sprintf("Document %s was created and queued for processing.", doc.OriginalFilename),
			map[string]any{
				"Name":         s.systemAdminName(),
				"Platform":     s.platformName(),
				"DocumentName": doc.OriginalFilename,
				"DocumentType": nullStringValue(doc.ContentType),
				"Status":       string(doc.Status),
				"ActionURL":    s.adminDocumentsURL(),
				"Details": fmt.Sprintf(
					"Document ID: %s\nFilename: %s\nContent Type: %s\nProcess Type: %s\nUploaded By: %s\nObject Key: %s",
					doc.ID.String(),
					doc.OriginalFilename,
					nullStringValue(doc.ContentType),
					string(input.ProcessType),
					input.UploadedBy.String(),
					doc.ObjectKey,
				),
			},
		)
	}

	s.notify(ctx, notification)

	return doc, nil
}

func (s *Service) GetDocument(
	ctx context.Context,
	id uuid.UUID,
) (db.Document, error) {
	if s == nil {
		return db.Document{}, errors.New("document service is nil")
	}

	if s.repo == nil {
		return db.Document{}, errors.New("document repository is nil")
	}

	if id == uuid.Nil {
		return db.Document{}, errors.New("document id is required")
	}

	doc, err := s.repo.GetDocument(ctx, id)
	if err != nil {
		return db.Document{}, fmt.Errorf("get document: %w", err)
	}

	return doc, nil
}

func (s *Service) EditDocument(
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

	// In-app only.
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

func (s *Service) DeleteDocument(
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

	// In-app only. Add email here later if document deletion must be escalated.
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

func (s *Service) ListDocuments(
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

func (s *Service) ListProcessesByDocument(
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

func (s *Service) Reprocess(
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
	notification := models.Notification{
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
	}

	s.attachAdminEmailDelivery(
		&notification,
		"document-created",
		"Document reprocessing queued",
		fmt.Sprintf("Document %s was queued for reprocessing.", doc.OriginalFilename),
		map[string]any{
			"Name":         s.systemAdminName(),
			"Platform":     s.platformName(),
			"DocumentName": doc.OriginalFilename,
			"DocumentType": nullStringValue(doc.ContentType),
			"Status":       "PENDING",
			"ActionURL":    s.adminDocumentsURL(),
			"Details": fmt.Sprintf(
				"Document ID: %s\nFilename: %s\nContent Type: %s\nProcess Type: %s\nQueued By: %s",
				doc.ID.String(),
				doc.OriginalFilename,
				nullStringValue(doc.ContentType),
				proc.ProcessType,
				proc.CreatedBy.String(),
			),
		},
	)

	s.notify(ctx, notification)

	return nil
}

func (s *Service) MarkDocumentProcessing(
	ctx context.Context,
	documentID uuid.UUID,
) (db.Document, error) {
	if s == nil {
		return db.Document{}, errors.New("document service is nil")
	}

	if s.repo == nil {
		return db.Document{}, errors.New("document repository is nil")
	}

	if documentID == uuid.Nil {
		return db.Document{}, errors.New("document id is required")
	}

	doc, err := s.repo.MarkDocumentProcessing(ctx, documentID)
	if err != nil {
		return db.Document{}, fmt.Errorf("mark document processing: %w", err)
	}

	return doc, nil
}

func (s *Service) MarkDocumentCompleted(
	ctx context.Context,
	documentID uuid.UUID,
) (db.Document, error) {
	if s == nil {
		return db.Document{}, errors.New("document service is nil")
	}

	if s.repo == nil {
		return db.Document{}, errors.New("document repository is nil")
	}

	if documentID == uuid.Nil {
		return db.Document{}, errors.New("document id is required")
	}

	doc, err := s.repo.MarkDocumentCompleted(ctx, documentID)
	if err != nil {
		return db.Document{}, fmt.Errorf("mark document completed: %w", err)
	}

	nt := models.DocumentEdited
	notification := models.Notification{
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
	}

	s.attachAdminEmailDelivery(
		&notification,
		"document-processed",
		"Document processing completed",
		fmt.Sprintf("Document %s was processed successfully.", doc.OriginalFilename),
		map[string]any{
			"Name":         s.systemAdminName(),
			"Platform":     s.platformName(),
			"DocumentName": doc.OriginalFilename,
			"DocumentType": nullStringValue(doc.ContentType),
			"ActionURL":    s.adminDocumentsURL(),
			"Details": fmt.Sprintf(
				"Document ID: %s\nFilename: %s\nContent Type: %s\nStatus: %s\nObject Key: %s",
				doc.ID.String(),
				doc.OriginalFilename,
				nullStringValue(doc.ContentType),
				string(doc.Status),
				doc.ObjectKey,
			),
		},
	)

	s.notify(ctx, notification)

	return doc, nil
}

func (s *Service) MarkDocumentFailed(
	ctx context.Context,
	documentID uuid.UUID,
) (db.Document, error) {
	if s == nil {
		return db.Document{}, errors.New("document service is nil")
	}

	if s.repo == nil {
		return db.Document{}, errors.New("document repository is nil")
	}

	if documentID == uuid.Nil {
		return db.Document{}, errors.New("document id is required")
	}

	doc, err := s.repo.MarkDocumentFailed(ctx, documentID)
	if err != nil {
		return db.Document{}, fmt.Errorf("mark document failed: %w", err)
	}

	nt := models.DocumentDeleted
	notification := models.Notification{
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
	}

	s.attachAdminEmailDelivery(
		&notification,
		"document-failed",
		"Document processing failed",
		fmt.Sprintf("Document %s failed during processing.", doc.OriginalFilename),
		map[string]any{
			"Name":         s.systemAdminName(),
			"Platform":     s.platformName(),
			"DocumentName": doc.OriginalFilename,
			"DocumentType": nullStringValue(doc.ContentType),
			"Reason":       "Processing failed. Please review process logs for details.",
			"ActionURL":    s.adminDocumentsURL(),
			"Details": fmt.Sprintf(
				"Document ID: %s\nFilename: %s\nContent Type: %s\nStatus: %s\nObject Key: %s",
				doc.ID.String(),
				doc.OriginalFilename,
				nullStringValue(doc.ContentType),
				string(doc.Status),
				doc.ObjectKey,
			),
		},
	)

	s.notify(ctx, notification)

	return doc, nil
}

func (s *Service) notify(
	ctx context.Context,
	notification models.Notification,
) {
	if s == nil || s.notifications == nil {
		return
	}

	if strings.TrimSpace(notification.TargetRole) == "" {
		notification.TargetRole = "admin"
	}

	_, _ = s.notifications.Notify(ctx, notification)
}

func (s *Service) attachAdminEmailDelivery(
	notification *models.Notification,
	templateName string,
	subject string,
	textBody string,
	templateData map[string]any,
) {
	if notification == nil {
		return
	}

	adminEmail := strings.TrimSpace(s.systemAdminEmail())
	if adminEmail == "" {
		return
	}

	if templateData == nil {
		templateData = map[string]any{}
	}

	if _, ok := templateData["Name"]; !ok {
		templateData["Name"] = s.systemAdminName()
	}

	if _, ok := templateData["Platform"]; !ok {
		templateData["Platform"] = s.platformName()
	}

	if _, ok := templateData["ActionURL"]; !ok {
		templateData["ActionURL"] = s.adminDocumentsURL()
	}

	notification.Deliveries = []models.NotificationDeliveryRequest{
		{
			Channel: models.NotificationChannelInApp,
			Recipient: map[string]any{
				"target_role": notification.TargetRole,
			},
			Payload: map[string]any{
				"title":    notification.Title,
				"message":  notification.Message,
				"type":     notification.Type,
				"severity": notification.Severity,
			},
			MaxAttempts: 1,
		},
		{
			Channel: models.NotificationChannelEmail,
			Recipient: map[string]any{
				"name":  s.systemAdminName(),
				"email": adminEmail,
			},
			TemplateName: templateName,
			TemplateData: templateData,
			Payload: map[string]any{
				"subject":   subject,
				"text_body": textBody,
			},
			MaxAttempts: 5,
		},
	}
}

func (s *Service) platformName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.PlatformName) != "" {
		return strings.TrimSpace(s.cfg.Notification.PlatformName)
	}

	return "MOH Integrated Health Portal"
}

func (s *Service) systemAdminName() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminName) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminName)
	}

	return "System Administrator"
}

func (s *Service) systemAdminEmail() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.SystemAdminEmail) != "" {
		return strings.TrimSpace(s.cfg.Notification.SystemAdminEmail)
	}

	return ""
}

func (s *Service) adminDashboardURL() string {
	if s != nil && s.cfg != nil && strings.TrimSpace(s.cfg.Notification.AdminDashboardURL) != "" {
		return strings.TrimSpace(s.cfg.Notification.AdminDashboardURL)
	}

	return "http://localhost:3000/admin/home"
}

func (s *Service) adminDocumentsURL() string {
	base := strings.TrimRight(s.adminDashboardURL(), "/")

	if strings.HasSuffix(base, "/admin/home") {
		return strings.TrimSuffix(base, "/admin/home") + "/admin/documents"
	}

	return base + "/documents"
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

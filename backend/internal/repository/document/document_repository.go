package document

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
)

type DocumentRepository interface {
	CreateDocumentWithProcess(
		ctx context.Context,
		docArg db.CreateDocumentParams,
		processArg db.CreateProcessParams,
	) (db.Document, error)

	CreateDocument(
		ctx context.Context,
		arg db.CreateDocumentParams,
	) (db.Document, error)

	GetDocument(
		ctx context.Context,
		id uuid.UUID,
	) (db.Document, error)

	ListDocuments(
		ctx context.Context,
		page model.Pagination,
	) ([]db.Document, error)

	ListDocumentsByUser(
		ctx context.Context,
		userID uuid.UUID,
		page model.Pagination,
	) ([]db.Document, error)

	ListProcessesByDocument(
		ctx context.Context,
		documentID uuid.UUID,
	) ([]db.Process, error)

	EditDocument(
		ctx context.Context,
		arg db.UpdateDocumentParams,
	) (db.Document, error)

	UpdateDocumentStatus(
		ctx context.Context,
		arg db.UpdateDocumentStatusParams,
	) (db.Document, error)

	MarkDocumentPending(
		ctx context.Context,
		id uuid.UUID,
	) (db.Document, error)

	MarkDocumentProcessing(
		ctx context.Context,
		id uuid.UUID,
	) (db.Document, error)

	MarkDocumentCompleted(
		ctx context.Context,
		id uuid.UUID,
	) (db.Document, error)

	MarkDocumentFailed(
		ctx context.Context,
		id uuid.UUID,
	) (db.Document, error)

	DeleteDocument(
		ctx context.Context,
		id uuid.UUID,
	) error

	GetLatestByDocumentID(
		ctx context.Context,
		documentID uuid.UUID,
	) (db.Process, error)
}

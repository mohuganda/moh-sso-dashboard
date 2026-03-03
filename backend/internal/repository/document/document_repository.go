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
	CreateDocument(ctx context.Context,
		arg db.CreateDocumentParams) (db.Document, error)
	GetDocument(context.Context, uuid.UUID) (db.Document, error)
	ListDocuments(ctx context.Context, page model.Pagination) ([]db.Document, error)
	ListProcessesByDocument(
		ctx context.Context,
		documentID uuid.UUID,
	) ([]db.Process, error)
	ListDocumentsByUser(ctx context.Context, userID uuid.UUID, page model.Pagination) ([]db.Document, error)
	EditDocument(context.Context, db.UpdateDocumentParams) (db.Document, error)
	DeleteDocument(context.Context, uuid.UUID) error

	GetLatestByDocumentID(
		ctx context.Context,
		documentID uuid.UUID,
	) (db.Process, error)
}

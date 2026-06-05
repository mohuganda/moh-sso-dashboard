package document_templates

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type DocumentTemplateColumnRepository interface {
	Create(ctx context.Context, arg db.CreateDocumentTemplateColumnParams) (db.DocumentTemplateColumn, error)

	GetByID(ctx context.Context, id uuid.UUID) (db.DocumentTemplateColumn, error)

	GetByKey(ctx context.Context, sheetID uuid.UUID, key string) (db.DocumentTemplateColumn, error)

	List(ctx context.Context, sheetID uuid.UUID) ([]db.DocumentTemplateColumn, error)

	ListRequired(ctx context.Context, sheetID uuid.UUID) ([]db.DocumentTemplateColumn, error)

	ListUnique(ctx context.Context, sheetID uuid.UUID) ([]db.DocumentTemplateColumn, error)

	Update(ctx context.Context, arg db.UpdateDocumentTemplateColumnParams) (db.DocumentTemplateColumn, error)

	Archive(ctx context.Context, id uuid.UUID) error

	Delete(ctx context.Context, id uuid.UUID) error

	ExistsKey(ctx context.Context, sheetID uuid.UUID, key string) (bool, error)

	ExistsName(ctx context.Context, sheetID uuid.UUID, name string) (bool, error)
}

type ColumnRepository = DocumentTemplateColumnRepository

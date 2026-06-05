package document_templates

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type DocumentTemplateSheetRepository interface {
	Create(ctx context.Context, arg db.CreateDocumentTemplateSheetParams) (db.DocumentTemplateSheet, error)

	GetByID(ctx context.Context, id uuid.UUID) (db.DocumentTemplateSheet, error)

	GetByCode(ctx context.Context, templateID uuid.UUID, code string) (db.DocumentTemplateSheet, error)

	List(ctx context.Context, templateID uuid.UUID) ([]db.DocumentTemplateSheet, error)

	ListRequired(ctx context.Context, templateID uuid.UUID) ([]db.DocumentTemplateSheet, error)

	Update(ctx context.Context, arg db.UpdateDocumentTemplateSheetParams) (db.DocumentTemplateSheet, error)

	Archive(ctx context.Context, id uuid.UUID) error

	Delete(ctx context.Context, id uuid.UUID) error

	ExistsCode(ctx context.Context, templateID uuid.UUID, code string) (bool, error)

	ExistsName(ctx context.Context, templateID uuid.UUID, name string) (bool, error)
}

type SheetRepository = DocumentTemplateSheetRepository

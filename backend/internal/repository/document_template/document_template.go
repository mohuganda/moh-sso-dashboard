package document_template

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type DocumentTemplateRepository interface {
	Create(ctx context.Context, arg db.CreateDocumentTemplateParams) (db.DocumentTemplate, error)

	GetByID(ctx context.Context, id uuid.UUID) (db.DocumentTemplate, error)

	GetByCode(ctx context.Context, code string) (db.DocumentTemplate, error)

	List(ctx context.Context) ([]db.DocumentTemplate, error)

	ListActive(ctx context.Context) ([]db.DocumentTemplate, error)

	ListVersions(ctx context.Context, code string) ([]db.DocumentTemplate, error)

	Update(ctx context.Context, arg db.UpdateDocumentTemplateParams) (db.DocumentTemplate, error)

	Archive(ctx context.Context, id uuid.UUID) error

	Delete(ctx context.Context, id uuid.UUID) error

	ExistsCode(ctx context.Context, code string) (bool, error)
}

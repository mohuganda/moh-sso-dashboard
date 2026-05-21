package document_template

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

var (
	ErrNotFound = errors.New("resource not found")
)

type DocumentTemplateRepositoryImpl struct {
	q db.Store
}

func NewDocumentTemplateRepository(q db.Store) DocumentTemplateRepository {
	return &DocumentTemplateRepositoryImpl{
		q: q,
	}
}

func (r *DocumentTemplateRepositoryImpl) Create(
	ctx context.Context,
	arg db.CreateDocumentTemplateParams,
) (db.DocumentTemplate, error) {

	return r.q.CreateDocumentTemplate(ctx, arg)
}

func (r *DocumentTemplateRepositoryImpl) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (db.DocumentTemplate, error) {

	template, err := r.q.GetDocumentTemplateByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.DocumentTemplate{}, ErrNotFound
		}
		return db.DocumentTemplate{}, err
	}

	return template, nil
}

func (r *DocumentTemplateRepositoryImpl) GetByCode(
	ctx context.Context,
	code string,
) (db.DocumentTemplate, error) {

	template, err := r.q.GetDocumentTemplateByCode(ctx, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.DocumentTemplate{}, ErrNotFound
		}
		return db.DocumentTemplate{}, err
	}

	return template, nil
}

func (r *DocumentTemplateRepositoryImpl) List(
	ctx context.Context,
) ([]db.DocumentTemplate, error) {

	return r.q.ListDocumentTemplates(ctx)
}

func (r *DocumentTemplateRepositoryImpl) ListActive(
	ctx context.Context,
) ([]db.DocumentTemplate, error) {

	return r.q.ListActiveDocumentTemplates(ctx)
}

func (r *DocumentTemplateRepositoryImpl) ListVersions(
	ctx context.Context,
	code string,
) ([]db.DocumentTemplate, error) {

	return r.q.ListDocumentTemplateVersions(ctx, code)
}

func (r *DocumentTemplateRepositoryImpl) Update(
	ctx context.Context,
	arg db.UpdateDocumentTemplateParams,
) (db.DocumentTemplate, error) {

	return r.q.UpdateDocumentTemplate(ctx, arg)
}

func (r *DocumentTemplateRepositoryImpl) Archive(
	ctx context.Context,
	id uuid.UUID,
) error {

	return r.q.ArchiveDocumentTemplate(ctx, id)
}

func (r *DocumentTemplateRepositoryImpl) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	return r.q.DeleteDocumentTemplate(ctx, id)
}

func (r *DocumentTemplateRepositoryImpl) ExistsCode(
	ctx context.Context,
	code string,
) (bool, error) {

	exists, err := r.q.ExistsDocumentTemplateCode(ctx, code)
	if err != nil {
		return false, err
	}

	return exists, nil
}

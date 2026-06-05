package document_templates

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type DocumentTemplateSheetRepositoryImpl struct {
	q db.Store
}

func NewDocumentTemplateSheetRepository(q db.Store) DocumentTemplateSheetRepository {
	return &DocumentTemplateSheetRepositoryImpl{
		q: q,
	}
}

func (r *DocumentTemplateSheetRepositoryImpl) Create(
	ctx context.Context,
	arg db.CreateDocumentTemplateSheetParams,
) (db.DocumentTemplateSheet, error) {

	return r.q.CreateDocumentTemplateSheet(ctx, arg)
}

func (r *DocumentTemplateSheetRepositoryImpl) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (db.DocumentTemplateSheet, error) {

	sheet, err := r.q.GetDocumentTemplateSheetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.DocumentTemplateSheet{}, ErrNotFound
		}
		return db.DocumentTemplateSheet{}, err
	}

	return sheet, nil
}

func (r *DocumentTemplateSheetRepositoryImpl) GetByCode(
	ctx context.Context,
	templateID uuid.UUID,
	code string,
) (db.DocumentTemplateSheet, error) {

	sheet, err := r.q.GetDocumentTemplateSheetByCode(ctx, db.GetDocumentTemplateSheetByCodeParams{
		TemplateID: templateID,
		Code:       code,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return db.DocumentTemplateSheet{}, ErrNotFound
		}
		return db.DocumentTemplateSheet{}, err
	}

	return sheet, nil
}

func (r *DocumentTemplateSheetRepositoryImpl) List(
	ctx context.Context,
	templateID uuid.UUID,
) ([]db.DocumentTemplateSheet, error) {

	return r.q.ListDocumentTemplateSheets(ctx, templateID)
}

func (r *DocumentTemplateSheetRepositoryImpl) ListRequired(
	ctx context.Context,
	templateID uuid.UUID,
) ([]db.DocumentTemplateSheet, error) {

	return r.q.ListRequiredDocumentTemplateSheets(ctx, templateID)
}

func (r *DocumentTemplateSheetRepositoryImpl) Update(
	ctx context.Context,
	arg db.UpdateDocumentTemplateSheetParams,
) (db.DocumentTemplateSheet, error) {

	return r.q.UpdateDocumentTemplateSheet(ctx, arg)
}

func (r *DocumentTemplateSheetRepositoryImpl) Archive(
	ctx context.Context,
	id uuid.UUID,
) error {

	return r.q.ArchiveDocumentTemplateSheet(ctx, id)
}

func (r *DocumentTemplateSheetRepositoryImpl) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	return r.q.DeleteDocumentTemplateSheet(ctx, id)
}

func (r *DocumentTemplateSheetRepositoryImpl) ExistsCode(
	ctx context.Context,
	templateID uuid.UUID,
	code string,
) (bool, error) {

	exists, err := r.q.ExistsDocumentTemplateSheetCode(ctx, db.ExistsDocumentTemplateSheetCodeParams{
		TemplateID: templateID,
		Code:       code,
	})
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *DocumentTemplateSheetRepositoryImpl) ExistsName(
	ctx context.Context,
	templateID uuid.UUID,
	name string,
) (bool, error) {

	exists, err := r.q.ExistsDocumentTemplateSheetName(ctx, db.ExistsDocumentTemplateSheetNameParams{
		TemplateID: templateID,
		Name:       name,
	})
	if err != nil {
		return false, err
	}

	return exists, nil
}

package document_templates

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type DocumentTemplateColumnRepositoryImpl struct {
	q db.Store
}

func NewDocumentTemplateColumnRepository(q db.Store) DocumentTemplateColumnRepository {
	return &DocumentTemplateColumnRepositoryImpl{
		q: q,
	}
}

func (r *DocumentTemplateColumnRepositoryImpl) Create(
	ctx context.Context,
	arg db.CreateDocumentTemplateColumnParams,
) (db.DocumentTemplateColumn, error) {

	return r.q.CreateDocumentTemplateColumn(ctx, arg)
}

func (r *DocumentTemplateColumnRepositoryImpl) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (db.DocumentTemplateColumn, error) {

	col, err := r.q.GetDocumentTemplateColumnByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.DocumentTemplateColumn{}, ErrNotFound
		}
		return db.DocumentTemplateColumn{}, err
	}

	return col, nil
}

func (r *DocumentTemplateColumnRepositoryImpl) GetByKey(
	ctx context.Context,
	sheetID uuid.UUID,
	key string,
) (db.DocumentTemplateColumn, error) {

	col, err := r.q.GetDocumentTemplateColumnByKey(ctx, db.GetDocumentTemplateColumnByKeyParams{
		SheetID:   sheetID,
		ColumnKey: key,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return db.DocumentTemplateColumn{}, ErrNotFound
		}
		return db.DocumentTemplateColumn{}, err
	}

	return col, nil
}

func (r *DocumentTemplateColumnRepositoryImpl) List(
	ctx context.Context,
	sheetID uuid.UUID,
) ([]db.DocumentTemplateColumn, error) {

	return r.q.ListDocumentTemplateColumns(ctx, sheetID)
}

func (r *DocumentTemplateColumnRepositoryImpl) ListRequired(
	ctx context.Context,
	sheetID uuid.UUID,
) ([]db.DocumentTemplateColumn, error) {

	return r.q.ListRequiredDocumentTemplateColumns(ctx, sheetID)
}

func (r *DocumentTemplateColumnRepositoryImpl) ListUnique(
	ctx context.Context,
	sheetID uuid.UUID,
) ([]db.DocumentTemplateColumn, error) {

	return r.q.ListUniqueDocumentTemplateColumns(ctx, sheetID)
}

func (r *DocumentTemplateColumnRepositoryImpl) Update(
	ctx context.Context,
	arg db.UpdateDocumentTemplateColumnParams,
) (db.DocumentTemplateColumn, error) {

	return r.q.UpdateDocumentTemplateColumn(ctx, arg)
}

func (r *DocumentTemplateColumnRepositoryImpl) Archive(
	ctx context.Context,
	id uuid.UUID,
) error {

	return r.q.ArchiveDocumentTemplateColumn(ctx, id)
}

func (r *DocumentTemplateColumnRepositoryImpl) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	return r.q.DeleteDocumentTemplateColumn(ctx, id)
}

func (r *DocumentTemplateColumnRepositoryImpl) ExistsKey(
	ctx context.Context,
	sheetID uuid.UUID,
	key string,
) (bool, error) {

	exists, err := r.q.ExistsDocumentTemplateColumnKey(ctx, db.ExistsDocumentTemplateColumnKeyParams{
		SheetID:   sheetID,
		ColumnKey: key,
	})
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *DocumentTemplateColumnRepositoryImpl) ExistsName(
	ctx context.Context,
	sheetID uuid.UUID,
	name string,
) (bool, error) {

	exists, err := r.q.ExistsDocumentTemplateColumnName(ctx, db.ExistsDocumentTemplateColumnNameParams{
		SheetID:    sheetID,
		ColumnName: name,
	})
	if err != nil {
		return false, err
	}

	return exists, nil
}

package document

import (
	"context"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
)

type documentRepository struct {
	db     db.Store
	config *config.Config
	logger *logger.Logger
}

func NewDocumentRepository(
	cfg *config.Config,
	store db.Store,
	log logger.Logger) DocumentRepository {
	return &documentRepository{
		db:     store,
		config: cfg,
		logger: &log,
	}
}

func (r *documentRepository) CreateDocumentWithProcess(
	ctx context.Context,
	docArg db.CreateDocumentParams,
	processArg db.CreateProcessParams,
) (db.Document, error) {

	var createdDoc db.Document

	err := r.db.ExecTx(ctx, func(q db.Querier) error {

		doc, err := q.CreateDocument(ctx, docArg)
		if err != nil {
			return err
		}

		createdDoc = doc

		processArg.DocumentID = doc.ID

		_, err = q.CreateProcess(ctx, processArg)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return db.Document{}, err
	}

	return createdDoc, nil
}

func (r *documentRepository) CreateDocument(
	ctx context.Context,
	arg db.CreateDocumentParams,
) (db.Document, error) {

	doc, err := r.db.CreateDocument(ctx, arg)
	if err != nil {
		return db.Document{}, err
	}

	return doc, nil
}

func (r *documentRepository) ListDocuments(
	ctx context.Context,
	page model.Pagination,
) ([]db.Document, error) {

	limit := page.Limit
	if limit <= 0 {
		limit = 25
	}

	docs, err := r.db.ListDocuments(ctx, db.ListDocumentsParams{
		Limit:  limit,
		Offset: page.Offset,
	})
	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (r *documentRepository) ListDocumentsByUser(
	ctx context.Context,
	userID uuid.UUID,
	page model.Pagination,
) ([]db.Document, error) {

	limit := page.Limit
	if limit <= 0 {
		limit = 25
	}

	docs, err := r.db.ListDocumentsByUser(ctx, db.ListDocumentsByUserParams{
		UploadedBy: userID,
		Limit:      limit,
		Offset:     page.Offset,
	})
	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (r *documentRepository) GetDocument(
	ctx context.Context,
	id uuid.UUID,
) (db.Document, error) {

	doc, err := r.db.GetDocumentByID(ctx, id)
	if err != nil {
		return db.Document{}, err
	}

	return doc, nil
}

func (r *documentRepository) EditDocument(
	ctx context.Context,
	arg db.UpdateDocumentParams,
) (db.Document, error) {

	doc, err := r.db.UpdateDocument(ctx, arg)
	if err != nil {
		return db.Document{}, err
	}

	return doc, nil
}

func (r *documentRepository) DeleteDocument(
	ctx context.Context,
	id uuid.UUID,
) error {

	err := r.db.DeleteDocument(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

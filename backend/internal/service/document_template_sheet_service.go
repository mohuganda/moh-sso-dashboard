package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
	documentTemplateSheetRepo "github.com/moh-sso-dashboard/internal/repository/document_template_sheet"
)

var (
	ErrSheetCodeExists = errors.New("sheet code already exists in template")
	ErrSheetNameExists = errors.New("sheet name already exists in template")

	ErrInvalidHeaderRow = errors.New("header row must be >= 1")
	ErrInvalidStartRow  = errors.New("start row must be >= header row")
)

type DocumentTemplateSheetService interface {
	CreateSheet(ctx context.Context, req model.CreateSheetRequest) (*model.DocumentTemplateSheet, error)
	GetSheet(ctx context.Context, id uuid.UUID) (*model.DocumentTemplateSheet, error)
	GetSheetByCode(ctx context.Context, templateID uuid.UUID, code string) (*model.DocumentTemplateSheet, error)
	ListSheets(ctx context.Context, templateID uuid.UUID) ([]model.DocumentTemplateSheet, error)
	ListRequiredSheets(ctx context.Context, templateID uuid.UUID) ([]model.DocumentTemplateSheet, error)
	UpdateSheet(ctx context.Context, req model.UpdateSheetRequest) (*model.DocumentTemplateSheet, error)
	ArchiveSheet(ctx context.Context, id uuid.UUID) error
	DeleteSheet(ctx context.Context, id uuid.UUID) error
}

type documentTemplateSheetService struct {
	repo documentTemplateSheetRepo.DocumentTemplateSheetRepository
}

func NewDocumentTemplateSheetService(
	repo documentTemplateSheetRepo.DocumentTemplateSheetRepository,
) DocumentTemplateSheetService {
	return &documentTemplateSheetService{
		repo: repo,
	}
}

func (s *documentTemplateSheetService) CreateSheet(
	ctx context.Context,
	req model.CreateSheetRequest,
) (*model.DocumentTemplateSheet, error) {

	exists, err := s.repo.ExistsCode(ctx, req.TemplateID, req.Code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrSheetCodeExists
	}

	nameExists, err := s.repo.ExistsName(ctx, req.TemplateID, req.Name)
	if err != nil {
		return nil, err
	}
	if nameExists {
		return nil, ErrSheetNameExists
	}

	if req.HeaderRow < 1 {
		return nil, ErrInvalidHeaderRow
	}
	if req.StartRow < req.HeaderRow {
		return nil, ErrInvalidStartRow
	}

	// -----------------------
	// SheetOrder conversion
	// -----------------------
	var sheetOrder sql.NullInt32
	if req.SheetOrder != nil {
		sheetOrder = sql.NullInt32{
			Int32: int32(*req.SheetOrder),
			Valid: true,
		}
	}

	// -----------------------
	// Configuration conversion
	// -----------------------
	var configJSON json.RawMessage
	if req.Configuration != nil {
		b, err := json.Marshal(req.Configuration)
		if err != nil {
			return nil, err
		}
		configJSON = b
	}

	sheet, err := s.repo.Create(ctx, db.CreateDocumentTemplateSheetParams{
		ID:         uuid.New(),
		TemplateID: req.TemplateID,
		Code:       req.Code,
		Name:       req.Name,

		DisplayName: sql.NullString{
			String: req.DisplayName,
			Valid:  req.DisplayName != "",
		},

		Required: req.Required,

		SheetOrder: sheetOrder,

		HeaderRow: int32(req.HeaderRow),
		StartRow:  int32(req.StartRow),

		AllowExtraColumns:     req.AllowExtraColumns,
		AllowDuplicateHeaders: req.AllowDuplicateHeaders, // FIXED

		Configuration: configJSON,
	})
	if err != nil {
		return nil, err
	}

	return mapSheet(sheet), nil
}

func (s *documentTemplateSheetService) GetSheet(
	ctx context.Context,
	id uuid.UUID,
) (*model.DocumentTemplateSheet, error) {

	sheet, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapSheet(sheet), nil
}

func (s *documentTemplateSheetService) GetSheetByCode(
	ctx context.Context,
	templateID uuid.UUID,
	code string,
) (*model.DocumentTemplateSheet, error) {

	sheet, err := s.repo.GetByCode(ctx, templateID, code)
	if err != nil {
		return nil, err
	}

	return mapSheet(sheet), nil
}

func (s *documentTemplateSheetService) ListSheets(
	ctx context.Context,
	templateID uuid.UUID,
) ([]model.DocumentTemplateSheet, error) {

	items, err := s.repo.List(ctx, templateID)
	if err != nil {
		return nil, err
	}

	var result []model.DocumentTemplateSheet
	for _, it := range items {
		result = append(result, *mapSheet(it))
	}

	return result, nil
}

func (s *documentTemplateSheetService) ListRequiredSheets(
	ctx context.Context,
	templateID uuid.UUID,
) ([]model.DocumentTemplateSheet, error) {

	items, err := s.repo.ListRequired(ctx, templateID)
	if err != nil {
		return nil, err
	}

	var result []model.DocumentTemplateSheet
	for _, it := range items {
		result = append(result, *mapSheet(it))
	}

	return result, nil
}

func (s *documentTemplateSheetService) UpdateSheet(
	ctx context.Context,
	req model.UpdateSheetRequest,
) (*model.DocumentTemplateSheet, error) {

	if req.HeaderRow < 1 {
		return nil, ErrInvalidHeaderRow
	}
	if req.StartRow < req.HeaderRow {
		return nil, ErrInvalidStartRow
	}

	var sheetOrder sql.NullInt32
	if req.SheetOrder != nil {
		sheetOrder = sql.NullInt32{
			Int32: int32(*req.SheetOrder),
			Valid: true,
		}
	} else {
		sheetOrder = sql.NullInt32{
			Valid: false,
		}
	}

	var configJSON json.RawMessage

	if req.Configuration != nil {
		b, err := json.Marshal(req.Configuration)
		if err != nil {
			return nil, err
		}
		configJSON = b
	} else {
		configJSON = nil
	}

	sheet, err := s.repo.Update(ctx, db.UpdateDocumentTemplateSheetParams{
		ID:   req.ID,
		Name: req.Name,

		DisplayName: sql.NullString{
			String: req.DisplayName,
			Valid:  req.DisplayName != "",
		},

		Required: req.Required,

		SheetOrder: sheetOrder,

		HeaderRow: int32(req.HeaderRow),
		StartRow:  int32(req.StartRow),

		AllowExtraColumns:     req.AllowExtraColumns,
		AllowDuplicateHeaders: req.AllowDuplicateHeaders,

		Configuration: configJSON,
	})
	if err != nil {
		return nil, err
	}

	return mapSheet(sheet), nil
}

func (s *documentTemplateSheetService) ArchiveSheet(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.repo.Archive(ctx, id)
}

func (s *documentTemplateSheetService) DeleteSheet(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.repo.Delete(ctx, id)
}

func mapSheet(s db.DocumentTemplateSheet) *model.DocumentTemplateSheet {

	var display string
	if s.DisplayName.Valid {
		display = s.DisplayName.String
	}

	var order *int
	if s.SheetOrder.Valid {
		v := int(s.SheetOrder.Int32)
		order = &v
	}

	return &model.DocumentTemplateSheet{
		ID:          s.ID,
		TemplateID:  s.TemplateID,
		Code:        s.Code,
		Name:        s.Name,
		DisplayName: display,

		Required: s.Required,

		SheetOrder: order,

		HeaderRow: int(s.HeaderRow),
		StartRow:  int(s.StartRow),

		AllowExtraColumns:     s.AllowExtraColumns,
		AllowDuplicateHeaders: s.AllowDuplicateHeaders,

		Configuration: func() map[string]any {
			var m map[string]any
			if len(s.Configuration) > 0 {
				_ = json.Unmarshal(s.Configuration, &m)
			}
			return m
		}(),
	}
}

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
	ErrSheetNotFound    = errors.New("sheet not found in template")
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

	// runtime helpers
	GetSheetRuntime(
		ctx context.Context,
		tpl *model.TemplateRuntime,
		sheetCode string,
	) (*model.TemplateSheetRuntime, error)

	ValidateSheetExists(
		ctx context.Context,
		tpl *model.TemplateRuntime,
		sheetCode string,
	) error
}

type documentTemplateSheetService struct {
	repo documentTemplateSheetRepo.DocumentTemplateSheetRepository
}

func NewDocumentTemplateSheetService(
	repo documentTemplateSheetRepo.DocumentTemplateSheetRepository,
) DocumentTemplateSheetService {
	return &documentTemplateSheetService{repo: repo}
}

/* =========================================================
 * CREATE
 * ========================================================= */
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

	var sheetOrder sql.NullInt32
	if req.SheetOrder != nil {
		sheetOrder = sql.NullInt32{
			Int32: int32(*req.SheetOrder),
			Valid: true,
		}
	}

	configJSON, err := json.Marshal(req.Configuration)
	if err != nil {
		return nil, err
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
		AllowDuplicateHeaders: req.AllowDuplicateHeaders,

		Configuration: configJSON,
	})
	if err != nil {
		return nil, err
	}

	return mapSheet(sheet), nil
}

/* =========================================================
 * READ
 * ========================================================= */
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

	result := make([]model.DocumentTemplateSheet, 0, len(items))
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

	result := make([]model.DocumentTemplateSheet, 0, len(items))
	for _, it := range items {
		result = append(result, *mapSheet(it))
	}

	return result, nil
}

/* =========================================================
 * UPDATE
 * ========================================================= */
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
	}

	configJSON, err := json.Marshal(req.Configuration)
	if err != nil {
		return nil, err
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

/* =========================================================
 * DELETE / ARCHIVE
 * ========================================================= */
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

/* =========================================================
 * RUNTIME HELPERS
 * ========================================================= */
func (s *documentTemplateSheetService) GetSheetRuntime(
	ctx context.Context,
	tpl *model.TemplateRuntime,
	sheetCode string,
) (*model.TemplateSheetRuntime, error) {

	if tpl == nil || tpl.Sheets == nil {
		return nil, ErrSheetNotFound
	}

	sheet, exists := tpl.Sheets[sheetCode]
	if !exists {
		return nil, ErrSheetNotFound
	}

	return sheet, nil
}

func (s *documentTemplateSheetService) ValidateSheetExists(
	ctx context.Context,
	tpl *model.TemplateRuntime,
	sheetCode string,
) error {

	if tpl == nil || tpl.Sheets == nil {
		return ErrSheetNotFound
	}

	if _, exists := tpl.Sheets[sheetCode]; !exists {
		return ErrSheetNotFound
	}

	return nil
}

/* =========================================================
 * MAPPER
 * ========================================================= */
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

	var config map[string]any
	if len(s.Configuration) > 0 {
		_ = json.Unmarshal(s.Configuration, &config)
	}

	return &model.DocumentTemplateSheet{
		ID:                    s.ID,
		TemplateID:            s.TemplateID,
		Code:                  s.Code,
		Name:                  s.Name,
		DisplayName:           display,
		Required:              s.Required,
		SheetOrder:            order,
		HeaderRow:             int(s.HeaderRow),
		StartRow:              int(s.StartRow),
		AllowExtraColumns:     s.AllowExtraColumns,
		AllowDuplicateHeaders: s.AllowDuplicateHeaders,
		Configuration:         config,
	}
}

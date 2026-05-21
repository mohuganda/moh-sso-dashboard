package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"

	documentTemplateRepo "github.com/moh-sso-dashboard/internal/repository/document_template"
	documentTemplateColumnRepo "github.com/moh-sso-dashboard/internal/repository/document_template_column"
	documentTemplateSheetRepo "github.com/moh-sso-dashboard/internal/repository/document_template_sheet"
)

var (
	ErrTemplateCodeExists = errors.New("template code already exists")
)

type DocumentTemplateService interface {
	CreateTemplate(ctx context.Context, req model.CreateTemplateRequest) (*model.DocumentTemplate, error)

	GetTemplate(ctx context.Context, id uuid.UUID) (*model.DocumentTemplate, error)

	GetTemplateByCode(ctx context.Context, code string) (*model.DocumentTemplate, error)

	ListTemplates(ctx context.Context) ([]model.DocumentTemplate, error)

	ListActiveTemplates(ctx context.Context) ([]model.DocumentTemplate, error)

	UpdateTemplate(ctx context.Context, req model.UpdateTemplateRequest) (*model.DocumentTemplate, error)

	ArchiveTemplate(ctx context.Context, id uuid.UUID) error

	PublishTemplate(ctx context.Context, id uuid.UUID) error

	GetTemplateStructure(ctx context.Context, code string) (*model.TemplateStructure, error)
}

type documentTemplateService struct {
	repo       documentTemplateRepo.DocumentTemplateRepository
	sheetRepo  documentTemplateSheetRepo.DocumentTemplateSheetRepository
	columnRepo documentTemplateColumnRepo.DocumentTemplateColumnRepository
}

func NewDocumentTemplateService(
	repo documentTemplateRepo.DocumentTemplateRepository,
	sheetRepo documentTemplateSheetRepo.DocumentTemplateSheetRepository,
	columnRepo documentTemplateColumnRepo.DocumentTemplateColumnRepository,
) DocumentTemplateService {

	return &documentTemplateService{
		repo:       repo,
		sheetRepo:  sheetRepo,
		columnRepo: columnRepo,
	}
}

func (s *documentTemplateService) GetTemplate(
	ctx context.Context,
	id uuid.UUID,
) (*model.DocumentTemplate, error) {

	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var config map[string]any
	if len(t.Configuration) > 0 {
		_ = json.Unmarshal(t.Configuration, &config)
	}

	return &model.DocumentTemplate{
		ID:            t.ID,
		Code:          t.Code,
		Name:          t.Name,
		Description:   t.Description.String,
		FileType:      t.FileType,
		Version:       int(t.Version),
		IsActive:      t.IsActive,
		Configuration: config,
	}, nil
}

func (s *documentTemplateService) CreateTemplate(
	ctx context.Context,
	req model.CreateTemplateRequest,
) (*model.DocumentTemplate, error) {

	// 1. Validate uniqueness
	exists, err := s.repo.ExistsCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTemplateCodeExists
	}

	// 2. Create template
	template, err := s.repo.Create(ctx, db.CreateDocumentTemplateParams{
		ID:   uuid.New(),
		Code: req.Code,
		Name: req.Name,
		Description: sql.NullString{
			String: req.Description,
			Valid:  req.Description != "",
		},
		FileType:      req.FileType,
		Version:       1,
		IsActive:      false,
		Configuration: toJSON(req.Configuration),
		CreatedBy:     req.CreatedBy,
	})
	if err != nil {
		return nil, err
	}

	return &model.DocumentTemplate{
		ID:            template.ID,
		Code:          template.Code,
		Name:          template.Name,
		Description:   template.Description.String,
		FileType:      template.FileType,
		Version:       int(template.Version),
		IsActive:      template.IsActive,
		Configuration: mustMap(template.Configuration),
	}, nil
}

func (s *documentTemplateService) GetTemplateByCode(
	ctx context.Context,
	code string,
) (*model.DocumentTemplate, error) {

	t, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	var config map[string]any
	if len(t.Configuration) > 0 {
		_ = json.Unmarshal(t.Configuration, &config)
	}

	return &model.DocumentTemplate{
		ID:            t.ID,
		Code:          t.Code,
		Name:          t.Name,
		Description:   t.Description.String,
		FileType:      t.FileType,
		Version:       int(t.Version),
		IsActive:      t.IsActive,
		Configuration: config,
	}, nil
}

func (s *documentTemplateService) ListTemplates(ctx context.Context) ([]model.DocumentTemplate, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var result []model.DocumentTemplate
	for _, t := range items {
		var config map[string]any
		if len(t.Configuration) > 0 {
			_ = json.Unmarshal(t.Configuration, &config)
		}
		result = append(result, model.DocumentTemplate{
			ID:            t.ID,
			Code:          t.Code,
			Name:          t.Name,
			Description:   t.Description.String,
			FileType:      t.FileType,
			Version:       int(t.Version),
			IsActive:      t.IsActive,
			Configuration: config,
		})
	}

	return result, nil
}

func (s *documentTemplateService) ListActiveTemplates(ctx context.Context) ([]model.DocumentTemplate, error) {
	items, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	var result []model.DocumentTemplate
	for _, t := range items {
		var config map[string]any
		if len(t.Configuration) > 0 {
			_ = json.Unmarshal(t.Configuration, &config)
		}
		result = append(result, model.DocumentTemplate{
			ID:            t.ID,
			Code:          t.Code,
			Name:          t.Name,
			Description:   t.Description.String,
			FileType:      t.FileType,
			Version:       int(t.Version),
			IsActive:      t.IsActive,
			Configuration: config,
		})
	}

	return result, nil
}

func (s *documentTemplateService) PublishTemplate(
	ctx context.Context,
	id uuid.UUID,
) error {

	// 1. Get template
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Deactivate all other versions
	versions, err := s.repo.ListVersions(ctx, t.Code)
	if err != nil {
		return err
	}

	for _, v := range versions {
		v.IsActive = false

		_, err := s.repo.Update(ctx, db.UpdateDocumentTemplateParams{
			ID:            v.ID,
			Name:          v.Name,
			Description:   v.Description,
			FileType:      v.FileType,
			Configuration: v.Configuration,
			IsActive:      false,
		})
		if err != nil {
			return err
		}
	}

	// 3. Activate selected version
	t.IsActive = true

	_, err = s.repo.Update(ctx, db.UpdateDocumentTemplateParams{
		ID:            t.ID,
		Name:          t.Name,
		Description:   t.Description,
		FileType:      t.FileType,
		Configuration: t.Configuration,
		IsActive:      true,
	})

	return err
}

func (s *documentTemplateService) ArchiveTemplate(
	ctx context.Context,
	id uuid.UUID,
) error {

	return s.repo.Archive(ctx, id)
}

func (s *documentTemplateService) GetTemplateStructure(
	ctx context.Context,
	code string,
) (*model.TemplateStructure, error) {

	// =====================================================
	// Load Template
	// =====================================================

	t, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// =====================================================
	// Load Sheets
	// =====================================================

	sheets, err := s.sheetRepo.List(ctx, t.ID)
	if err != nil {
		return nil, err
	}

	// =====================================================
	// Build Structure
	// =====================================================

	structure := &model.TemplateStructure{
		ID:       t.ID,
		Code:     t.Code,
		Name:     t.Name,
		Version:  int(t.Version),
		IsActive: t.IsActive,

		Sheets: make([]model.TemplateSheetStructure, 0),
	}

	// =====================================================
	// Process Sheets
	// =====================================================

	for _, sh := range sheets {

		cols, err := s.columnRepo.List(ctx, sh.ID)
		if err != nil {
			return nil, err
		}

		sheet := model.TemplateSheetStructure{
			ID: sh.ID,

			Code: sh.Code,
			Name: sh.Name,

			Required: sh.Required,

			HeaderRow: int(sh.HeaderRow),
			StartRow:  int(sh.StartRow),

			Columns: make([]model.TemplateColumnStructure, 0),
		}

		// =================================================
		// Process Columns
		// =================================================

		for _, c := range cols {

			sheet.Columns = append(sheet.Columns, model.TemplateColumnStructure{
				ID: c.ID,

				Key:         c.ColumnKey,
				Name:        c.ColumnName,
				DisplayName: nullStringValue(c.DisplayName),

				DataType: string(c.DataType),

				Required: c.Required,
				IsUnique: c.IsUnique,

				DefaultValue: nullStringToPtr(c.DefaultValue),

				AllowedValues: mustSliceString(c.AllowedValues),
				Aliases:       mustSliceString(c.Aliases),

				Configuration: mustMap(c.Configuration),
			})
		}

		structure.Sheets = append(structure.Sheets, sheet)
	}

	return structure, nil
}

func (s *documentTemplateService) UpdateTemplate(
	ctx context.Context,
	req model.UpdateTemplateRequest,
) (*model.DocumentTemplate, error) {

	// 1. Get existing template
	t, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 2. Marshal configuration
	configJSON, err := json.Marshal(req.Configuration)
	if err != nil {
		return nil, err
	}

	// 3. Update template
	updated, err := s.repo.Update(ctx, db.UpdateDocumentTemplateParams{
		ID:   t.ID,
		Name: req.Name,

		Description: sql.NullString{
			String: req.Description,
			Valid:  req.Description != "",
		},

		FileType: req.FileType,

		Configuration: configJSON,

		IsActive: req.IsActive,
	})
	if err != nil {
		return nil, err
	}

	// 4. Convert configuration back to map
	var config map[string]any

	if len(updated.Configuration) > 0 {
		_ = json.Unmarshal(updated.Configuration, &config)
	}

	return &model.DocumentTemplate{
		ID:            updated.ID,
		Code:          updated.Code,
		Name:          updated.Name,
		Description:   updated.Description.String,
		FileType:      updated.FileType,
		Version:       int(updated.Version),
		IsActive:      updated.IsActive,
		Configuration: config,
	}, nil
}

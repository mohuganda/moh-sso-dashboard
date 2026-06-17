package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"

	documentRepo "github.com/moh-sso-dashboard/internal/repository/document"
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
	CreateTemplateStructure(

		ctx context.Context,
		req model.CreateTemplateStructureRequest,
	) (*model.TemplateStructure, error)

	GetTemplateRuntime(ctx context.Context, code string) (*model.TemplateRuntime, error)

	GetTemplateByCode(ctx context.Context, code string) (*model.DocumentTemplate, error)

	ListTemplates(ctx context.Context) ([]model.DocumentTemplate, error)

	ListActiveTemplates(ctx context.Context) ([]model.DocumentTemplate, error)

	UpdateTemplate(ctx context.Context, req model.UpdateTemplateRequest) (*model.DocumentTemplate, error)

	DeleteTemplate(ctx context.Context, id uuid.UUID) error

	ArchiveTemplate(ctx context.Context, id uuid.UUID) error

	PublishTemplate(ctx context.Context, id uuid.UUID) error

	GetTemplateStructure(ctx context.Context, code string) (*model.TemplateStructure, error)
	GetTemplateStructureByDocumentID(
		ctx context.Context,
		documentID uuid.UUID,
	) (*model.TemplateStructure, error)

	ReplaceStructure(
		ctx context.Context,
		templateID uuid.UUID,
		sheets []model.CreateTemplateSheetWithColumnsRequest,
	) (*model.TemplateStructure, error)
}

type documentTemplateService struct {
	documentRepo         documentRepo.DocumentRepository
	documentTemplateRepo documentTemplateRepo.DocumentTemplateRepository
	sheetRepo            documentTemplateSheetRepo.DocumentTemplateSheetRepository
	columnRepo           documentTemplateColumnRepo.DocumentTemplateColumnRepository
}

func NewDocumentTemplateService(
	documentTemplateRepo documentTemplateRepo.DocumentTemplateRepository,
	sheetRepo documentTemplateSheetRepo.DocumentTemplateSheetRepository,
	columnRepo documentTemplateColumnRepo.DocumentTemplateColumnRepository,
) DocumentTemplateService {

	return &documentTemplateService{
		documentTemplateRepo: documentTemplateRepo,
		sheetRepo:            sheetRepo,
		columnRepo:           columnRepo,
	}
}

func (s *documentTemplateService) GetTemplate(
	ctx context.Context,
	id uuid.UUID,
) (*model.DocumentTemplate, error) {

	t, err := s.documentTemplateRepo.GetByID(ctx, id)
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
	exists, err := s.documentTemplateRepo.ExistsCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTemplateCodeExists
	}

	// 2. Create template
	template, err := s.documentTemplateRepo.Create(ctx, db.CreateDocumentTemplateParams{
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

func (s *documentTemplateService) GetTemplateRuntime(
	ctx context.Context,
	code string,
) (*model.TemplateRuntime, error) {

	t, err := s.documentTemplateRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return &model.TemplateRuntime{
		ID:      t.ID,
		Code:    t.Code,
		Version: int(t.Version),
	}, nil

}

func (s *documentTemplateService) GetTemplateByCode(
	ctx context.Context,
	code string,
) (*model.DocumentTemplate, error) {

	t, err := s.documentTemplateRepo.GetByCode(ctx, code)
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
	items, err := s.documentTemplateRepo.List(ctx)
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
	items, err := s.documentTemplateRepo.ListActive(ctx)
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
	t, err := s.documentTemplateRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Deactivate all other versions
	versions, err := s.documentTemplateRepo.ListVersions(ctx, t.Code)
	if err != nil {
		return err
	}

	for _, v := range versions {
		v.IsActive = false

		_, err := s.documentTemplateRepo.Update(ctx, db.UpdateDocumentTemplateParams{
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

	_, err = s.documentTemplateRepo.Update(ctx, db.UpdateDocumentTemplateParams{
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

	return s.documentTemplateRepo.Archive(ctx, id)
}

func (s *documentTemplateService) GetTemplateStructure(
	ctx context.Context,
	code string,
) (*model.TemplateStructure, error) {
	t, err := s.documentTemplateRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	sheets, err := s.sheetRepo.List(ctx, t.ID)
	if err != nil {
		return nil, err
	}

	structure := &model.TemplateStructure{
		Template: model.DocumentTemplate{
			ID:            t.ID,
			DocumentID:    nullUUIDPtr(t.DocumentID),
			Code:          t.Code,
			Name:          t.Name,
			Description:   nullStringValue(t.Description),
			FileType:      t.FileType,
			Version:       int(t.Version),
			IsActive:      t.IsActive,
			Configuration: fromJSON(t.Configuration),
			CreatedBy:     t.CreatedBy,
			CreatedAt:     t.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     t.UpdatedAt.Format(time.RFC3339),
			ArchivedAt:    nullTimeStringPtr(t.ArchivedAt),
		},
		Sheets: make([]model.TemplateSheetStructure, 0, len(sheets)),
	}

	for _, sh := range sheets {
		cols, err := s.columnRepo.List(ctx, sh.ID)
		if err != nil {
			return nil, err
		}

		sheet := model.TemplateSheetStructure{
			ID:                    sh.ID,
			TemplateID:            sh.TemplateID,
			Code:                  sh.Code,
			Name:                  sh.Name,
			DisplayName:           nullStringValue(sh.DisplayName),
			Required:              sh.Required,
			SheetOrder:            nullInt32Ptr(sh.SheetOrder),
			HeaderRow:             int(sh.HeaderRow),
			StartRow:              int(sh.StartRow),
			AllowExtraColumns:     sh.AllowExtraColumns,
			AllowDuplicateHeaders: sh.AllowDuplicateHeaders,
			Configuration:         fromJSON(sh.Configuration),
			Columns:               make([]model.TemplateColumnStructure, 0, len(cols)),
		}

		for _, c := range cols {
			sheet.Columns = append(sheet.Columns, model.TemplateColumnStructure{
				ID:            c.ID,
				SheetID:       c.SheetID,
				ColumnKey:     c.ColumnKey,
				ColumnName:    c.ColumnName,
				DisplayName:   nullStringValue(c.DisplayName),
				DataType:      string(c.DataType),
				Required:      c.Required,
				IsUnique:      c.IsUnique,
				ColumnOrder:   nullInt32Ptr(c.ColumnOrder),
				AllowedValues: fromJSONStringArray(c.AllowedValues),
				Aliases:       fromJSONStringArray(c.Aliases),
				Configuration: fromJSON(c.Configuration),
			})
		}

		structure.Sheets = append(structure.Sheets, sheet)
	}

	return structure, nil
}

func (s *documentTemplateService) GetTemplateStructureByDocumentID(
	ctx context.Context,
	documentID uuid.UUID,
) (*model.TemplateStructure, error) {
	t, err := s.documentTemplateRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	return s.GetTemplateStructure(ctx, t.Code)
}

func (s *documentTemplateService) UpdateTemplate(
	ctx context.Context,
	req model.UpdateTemplateRequest,
) (*model.DocumentTemplate, error) {

	// 1. Get existing template
	t, err := s.documentTemplateRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// 2. Marshal configuration
	configJSON, err := json.Marshal(req.Configuration)
	if err != nil {
		return nil, err
	}

	// 3. Update template
	updated, err := s.documentTemplateRepo.Update(ctx, db.UpdateDocumentTemplateParams{
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

func (s *documentTemplateService) DeleteTemplate(
	ctx context.Context,
	id uuid.UUID,
) error {

	// 1. Get template
	t, err := s.documentTemplateRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Delete all versions of the template
	versions, err := s.documentTemplateRepo.ListVersions(ctx, t.Code)
	if err != nil {
		return err
	}

	for _, v := range versions {
		err := s.documentTemplateRepo.Delete(ctx, v.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *documentTemplateService) CreateTemplateStructure(
	ctx context.Context,
	req model.CreateTemplateStructureRequest,
) (*model.TemplateStructure, error) {
	exists, err := s.documentTemplateRepo.ExistsCode(ctx, req.Template.Code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTemplateCodeExists
	}

	template, err := s.documentTemplateRepo.Create(ctx, db.CreateDocumentTemplateParams{
		ID: uuid.New(),
		DocumentID: uuid.NullUUID{
			UUID: func() uuid.UUID {
				if req.Template.DocumentID == nil {
					return uuid.Nil
				}

				return *req.Template.DocumentID
			}(),
			Valid: req.Template.DocumentID != nil,
		},
		Code: req.Template.Code,
		Name: req.Template.Name,
		Description: sql.NullString{
			String: req.Template.Description,
			Valid:  req.Template.Description != "",
		},
		FileType: req.Template.FileType,
		Version:  1,
		IsActive: false,
		Configuration: toJSON(
			req.Template.Configuration,
		),
		CreatedBy: req.Template.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	sheets := make([]model.TemplateSheetStructure, 0, len(req.Sheets))

	for _, sheetReq := range req.Sheets {
		sheetInput := sheetReq.Sheet

		sheetInput.TemplateID = template.ID

		sheet, err := s.sheetRepo.Create(ctx, db.CreateDocumentTemplateSheetParams{
			ID:         uuid.New(),
			TemplateID: sheetInput.TemplateID,
			Code:       sheetInput.Code,
			Name:       sheetInput.Name,
			DisplayName: sql.NullString{
				String: sheetInput.DisplayName,
				Valid:  sheetInput.DisplayName != "",
			},
			Required: sheetInput.Required,
			SheetOrder: sql.NullInt32{
				Int32: int32(*sheetInput.SheetOrder),
				Valid: sheetInput.SheetOrder != nil,
			},
			HeaderRow:             int32(sheetInput.HeaderRow),
			StartRow:              int32(sheetInput.StartRow),
			AllowExtraColumns:     sheetInput.AllowExtraColumns,
			AllowDuplicateHeaders: sheetInput.AllowDuplicateHeaders,
			Configuration:         toJSON(sheetInput.Configuration),
		})
		if err != nil {
			return nil, err
		}

		columnStructures := make([]model.TemplateColumnStructure, 0, len(sheetReq.Columns))

		for _, columnReq := range sheetReq.Columns {
			columnReq.SheetID = sheet.ID

			column, err := s.columnRepo.Create(ctx, db.CreateDocumentTemplateColumnParams{
				ID:         uuid.New(),
				SheetID:    columnReq.SheetID,
				ColumnKey:  columnReq.ColumnKey,
				ColumnName: columnReq.ColumnName,
				DisplayName: sql.NullString{
					String: columnReq.DisplayName,
					Valid:  columnReq.DisplayName != "",
				},
				DataType: normalizeColumnDataType(columnReq.DataType),
				Required: columnReq.Required,
				IsUnique: columnReq.IsUnique,
				ColumnOrder: sql.NullInt32{
					Int32: int32(*columnReq.ColumnOrder),
					Valid: columnReq.ColumnOrder != nil,
				},
				DefaultValue: sql.NullString{
					String: func() string {
						if columnReq.DefaultValue == nil {
							return ""
						}

						return *columnReq.DefaultValue
					}(),
					Valid: columnReq.DefaultValue != nil,
				},
				Configuration: toJSON(columnReq.Configuration),
				AllowedValues: toJSON(columnReq.AllowedValues),
				Aliases:       toJSON(columnReq.Aliases),
			})
			if err != nil {
				return nil, err
			}

			columnStructures = append(columnStructures, model.TemplateColumnStructure{
				ID:            column.ID,
				SheetID:       column.SheetID,
				ColumnKey:     column.ColumnKey,
				ColumnName:    column.ColumnName,
				DisplayName:   nullStringValue(column.DisplayName),
				DataType:      string(column.DataType),
				Required:      column.Required,
				IsUnique:      column.IsUnique,
				ColumnOrder:   nullInt32Ptr(column.ColumnOrder),
				AllowedValues: fromJSONStringArray(column.AllowedValues),
				Aliases:       fromJSONStringArray(column.Aliases),
				Configuration: fromJSON(column.Configuration),
			})
		}

		sheets = append(sheets, model.TemplateSheetStructure{
			ID:                    sheet.ID,
			TemplateID:            sheet.TemplateID,
			Code:                  sheet.Code,
			Name:                  sheet.Name,
			DisplayName:           nullStringValue(sheet.DisplayName),
			Required:              sheet.Required,
			SheetOrder:            nullInt32Ptr(sheet.SheetOrder),
			HeaderRow:             int(sheet.HeaderRow),
			StartRow:              int(sheet.StartRow),
			AllowExtraColumns:     sheet.AllowExtraColumns,
			AllowDuplicateHeaders: sheet.AllowDuplicateHeaders,
			Configuration:         fromJSON(sheet.Configuration),
			Columns:               columnStructures,
		})
	}

	return &model.TemplateStructure{
		Template: model.DocumentTemplate{
			ID:            template.ID,
			DocumentID:    nullUUIDPtr(template.DocumentID),
			Code:          template.Code,
			Name:          template.Name,
			Description:   nullStringValue(template.Description),
			FileType:      template.FileType,
			Version:       int(template.Version),
			IsActive:      template.IsActive,
			Configuration: fromJSON(template.Configuration),
			CreatedBy:     template.CreatedBy,
			CreatedAt:     template.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     template.UpdatedAt.Format(time.RFC3339),
			ArchivedAt:    nullTimeStringPtr(template.ArchivedAt),
		},
		Sheets: sheets,
	}, nil
}

func normalizeColumnDataType(value string) db.DocumentTemplateColumnType {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "TEXT", "STRING", "VARCHAR":
		return db.DocumentTemplateColumnTypeSTRING

	case "NUMBER", "NUMERIC", "DECIMAL", "FLOAT", "DOUBLE":
		return db.DocumentTemplateColumnTypeDECIMAL

	case "INTEGER", "INT":
		return db.DocumentTemplateColumnTypeINTEGER

	case "DATE":
		return db.DocumentTemplateColumnTypeDATE

	case "DATETIME", "TIMESTAMP":
		return db.DocumentTemplateColumnTypeDATETIME

	case "BOOLEAN", "BOOL":
		return db.DocumentTemplateColumnTypeBOOLEAN

	default:
		return db.DocumentTemplateColumnTypeSTRING
	}
}

func fromJSON(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}

	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return map[string]any{}
	}

	return result
}

func fromJSONStringArray(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return []string{}
	}

	var result []string
	if err := json.Unmarshal(raw, &result); err != nil {
		return []string{}
	}

	return result
}

func nullInt32Ptr(value sql.NullInt32) *int {
	if !value.Valid {
		return nil
	}

	v := int(value.Int32)
	return &v
}

func nullUUIDPtr(value uuid.NullUUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}

	return &value.UUID
}

func nullTimeStringPtr(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}

	v := value.Time.Format(time.RFC3339)
	return &v
}

func (s *documentTemplateService) ReplaceStructure(
	ctx context.Context,
	templateID uuid.UUID,
	sheets []model.CreateTemplateSheetWithColumnsRequest,
) (*model.TemplateStructure, error) {
	t, err := s.documentTemplateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	// Drop all existing sheets (columns cascade-delete via FK)
	existing, err := s.sheetRepo.List(ctx, templateID)
	if err != nil {
		return nil, err
	}
	for _, sh := range existing {
		if err := s.sheetRepo.Delete(ctx, sh.ID); err != nil {
			return nil, err
		}
	}

	// Recreate sheets and columns
	structure := &model.TemplateStructure{
		Template: model.DocumentTemplate{
			ID:            t.ID,
			DocumentID:    nullUUIDPtr(t.DocumentID),
			Code:          t.Code,
			Name:          t.Name,
			Description:   nullStringValue(t.Description),
			FileType:      t.FileType,
			Version:       int(t.Version),
			IsActive:      t.IsActive,
			Configuration: fromJSON(t.Configuration),
			CreatedBy:     t.CreatedBy,
			CreatedAt:     t.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     t.UpdatedAt.Format(time.RFC3339),
			ArchivedAt:    nullTimeStringPtr(t.ArchivedAt),
		},
		Sheets: make([]model.TemplateSheetStructure, 0, len(sheets)),
	}

	for _, sheetReq := range sheets {
		si := sheetReq.Sheet
		sheet, err := s.sheetRepo.Create(ctx, db.CreateDocumentTemplateSheetParams{
			ID:         uuid.New(),
			TemplateID: t.ID,
			Code:       si.Code,
			Name:       si.Name,
			DisplayName: sql.NullString{
				String: si.DisplayName,
				Valid:  si.DisplayName != "",
			},
			Required: si.Required,
			SheetOrder: sql.NullInt32{
				Int32: func() int32 {
					if si.SheetOrder != nil {
						return int32(*si.SheetOrder)
					}
					return 0
				}(),
				Valid: si.SheetOrder != nil,
			},
			HeaderRow:             int32(si.HeaderRow),
			StartRow:              int32(si.StartRow),
			AllowExtraColumns:     si.AllowExtraColumns,
			AllowDuplicateHeaders: si.AllowDuplicateHeaders,
			Configuration:         toJSON(si.Configuration),
		})
		if err != nil {
			return nil, err
		}

		sheetStructure := model.TemplateSheetStructure{
			ID:                    sheet.ID,
			TemplateID:            sheet.TemplateID,
			Code:                  sheet.Code,
			Name:                  sheet.Name,
			DisplayName:           nullStringValue(sheet.DisplayName),
			Required:              sheet.Required,
			SheetOrder:            nullInt32Ptr(sheet.SheetOrder),
			HeaderRow:             int(sheet.HeaderRow),
			StartRow:              int(sheet.StartRow),
			AllowExtraColumns:     sheet.AllowExtraColumns,
			AllowDuplicateHeaders: sheet.AllowDuplicateHeaders,
			Configuration:         fromJSON(sheet.Configuration),
			Columns:               make([]model.TemplateColumnStructure, 0, len(sheetReq.Columns)),
		}

		for _, colReq := range sheetReq.Columns {
			col, err := s.columnRepo.Create(ctx, db.CreateDocumentTemplateColumnParams{
				ID:         uuid.New(),
				SheetID:    sheet.ID,
				ColumnKey:  colReq.ColumnKey,
				ColumnName: colReq.ColumnName,
				DisplayName: sql.NullString{
					String: colReq.DisplayName,
					Valid:  colReq.DisplayName != "",
				},
				DataType: normalizeColumnDataType(colReq.DataType),
				Required: colReq.Required,
				IsUnique: colReq.IsUnique,
				ColumnOrder: sql.NullInt32{
					Int32: func() int32 {
						if colReq.ColumnOrder != nil {
							return int32(*colReq.ColumnOrder)
						}
						return 0
					}(),
					Valid: colReq.ColumnOrder != nil,
				},
				DefaultValue: sql.NullString{
					String: func() string {
						if colReq.DefaultValue != nil {
							return *colReq.DefaultValue
						}
						return ""
					}(),
					Valid: colReq.DefaultValue != nil,
				},
				AllowedValues: toJSON(colReq.AllowedValues),
				Aliases:       toJSON(colReq.Aliases),
				Configuration: toJSON(colReq.Configuration),
			})
			if err != nil {
				return nil, err
			}
			sheetStructure.Columns = append(sheetStructure.Columns, model.TemplateColumnStructure{
				ID:            col.ID,
				SheetID:       col.SheetID,
				ColumnKey:     col.ColumnKey,
				ColumnName:    col.ColumnName,
				DisplayName:   nullStringValue(col.DisplayName),
				DataType:      string(col.DataType),
				Required:      col.Required,
				IsUnique:      col.IsUnique,
				ColumnOrder:   nullInt32Ptr(col.ColumnOrder),
				AllowedValues: fromJSONStringArray(col.AllowedValues),
				Aliases:       fromJSONStringArray(col.Aliases),
				Configuration: fromJSON(col.Configuration),
			})
		}

		structure.Sheets = append(structure.Sheets, sheetStructure)
	}

	return structure, nil
}

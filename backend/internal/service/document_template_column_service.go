package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"

	documentTemplateColumnRepo "github.com/moh-sso-dashboard/internal/repository/document_template_column"
)

var (
	ErrColumnKeyExists    = errors.New("column key already exists")
	ErrColumnNameExists   = errors.New("column name already exists")
	ErrEnumRequiresValues = errors.New("ENUM columns must have allowed values")
	ErrInvalidDataType    = errors.New("invalid column data type")
)

type DocumentTemplateColumnService interface {
	CreateColumn(ctx context.Context, req model.CreateColumnRequest) (*model.DocumentTemplateColumn, error)

	GetColumn(ctx context.Context, id uuid.UUID) (*model.DocumentTemplateColumn, error)

	ListColumns(ctx context.Context, sheetID uuid.UUID) ([]model.DocumentTemplateColumn, error)

	ListRequiredColumns(ctx context.Context, sheetID uuid.UUID) ([]model.DocumentTemplateColumn, error)

	ListUniqueColumns(ctx context.Context, sheetID uuid.UUID) ([]model.DocumentTemplateColumn, error)

	UpdateColumn(ctx context.Context, req model.UpdateColumnRequest) (*model.DocumentTemplateColumn, error)

	ArchiveColumn(ctx context.Context, id uuid.UUID) error

	DeleteColumn(ctx context.Context, id uuid.UUID) error
}

type documentTemplateColumnService struct {
	repo documentTemplateColumnRepo.DocumentTemplateColumnRepository
}

func NewDocumentTemplateColumnService(
	repo documentTemplateColumnRepo.DocumentTemplateColumnRepository,
) DocumentTemplateColumnService {

	return &documentTemplateColumnService{
		repo: repo,
	}
}

func (s *documentTemplateColumnService) ListUniqueColumns(
	ctx context.Context,
	sheetID uuid.UUID,
) ([]model.DocumentTemplateColumn, error) {

	items, err := s.repo.ListUnique(ctx, sheetID)
	if err != nil {
		return nil, err
	}

	var result []model.DocumentTemplateColumn
	for _, c := range items {
		result = append(result, *mapColumn(c))
	}

	return result, nil
}

func (s *documentTemplateColumnService) CreateColumn(
	ctx context.Context,
	req model.CreateColumnRequest,
) (*model.DocumentTemplateColumn, error) {

	// 1. Validate key uniqueness
	exists, err := s.repo.ExistsKey(ctx, req.SheetID, req.Key)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrColumnKeyExists
	}

	// 2. Validate name uniqueness
	nameExists, err := s.repo.ExistsName(ctx, req.SheetID, req.Name)
	if err != nil {
		return nil, err
	}
	if nameExists {
		return nil, ErrColumnNameExists
	}

	// 3. Validate data type
	if err := validateDataType(req.DataType); err != nil {
		return nil, err
	}

	// 4. Validate enum consistency
	if req.DataType == "ENUM" && len(req.AllowedValues) == 0 {
		return nil, ErrEnumRequiresValues
	}

	// 5. Create column
	col, err := s.repo.Create(ctx, db.CreateDocumentTemplateColumnParams{
		ID:         uuid.New(),
		SheetID:    req.SheetID,
		ColumnKey:  req.Key,
		ColumnName: req.Name,
		DisplayName: sql.NullString{
			String: req.DisplayName,
			Valid:  req.DisplayName != "",
		},
		DataType: db.DocumentTemplateColumnType(req.DataType),
		Required: req.Required,
		IsUnique: req.IsUnique,

		DefaultValue: sql.NullString{
			String: func() string {
				if req.DefaultValue != nil {
					return *req.DefaultValue
				}
				return ""
			}(),
			Valid: req.DefaultValue != nil,
		},

		AllowedValues: toJSON(req.AllowedValues),
		Aliases:       toJSON(req.Aliases),
		Configuration: toJSON(req.Configuration),
	})
	if err != nil {
		return nil, err
	}

	return mapColumn(col), nil
}

func (s *documentTemplateColumnService) GetColumn(
	ctx context.Context,
	id uuid.UUID,
) (*model.DocumentTemplateColumn, error) {

	col, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapColumn(col), nil
}

func (s *documentTemplateColumnService) ListColumns(
	ctx context.Context,
	sheetID uuid.UUID,
) ([]model.DocumentTemplateColumn, error) {

	items, err := s.repo.List(ctx, sheetID)
	if err != nil {
		return nil, err
	}

	var result []model.DocumentTemplateColumn
	for _, c := range items {
		result = append(result, *mapColumn(c))
	}

	return result, nil
}

func (s *documentTemplateColumnService) ListRequiredColumns(
	ctx context.Context,
	sheetID uuid.UUID,
) ([]model.DocumentTemplateColumn, error) {

	items, err := s.repo.ListRequired(ctx, sheetID)
	if err != nil {
		return nil, err
	}

	var result []model.DocumentTemplateColumn
	for _, c := range items {
		result = append(result, *mapColumn(c))
	}

	return result, nil
}

func (s *documentTemplateColumnService) UpdateColumn(
	ctx context.Context,
	req model.UpdateColumnRequest,
) (*model.DocumentTemplateColumn, error) {

	col, err := s.repo.Update(ctx, db.UpdateDocumentTemplateColumnParams{
		ID:         req.ID,
		ColumnName: req.Name,
		DisplayName: sql.NullString{
			String: req.DisplayName,
			Valid:  req.DisplayName != "",
		},
		DataType: db.DocumentTemplateColumnType(req.DataType),
		Required: req.Required,
		IsUnique: req.IsUnique,
		DefaultValue: sql.NullString{
			String: func() string {
				if req.DefaultValue != nil {
					return *req.DefaultValue
				}
				return ""
			}(),
			Valid: req.DefaultValue != nil,
		},
		AllowedValues: toJSON(req.AllowedValues),
		Aliases:       toJSON(req.Aliases),
		Configuration: toJSON(req.Configuration),
	})
	if err != nil {
		return nil, err
	}

	return mapColumn(col), nil
}

func (s *documentTemplateColumnService) ArchiveColumn(
	ctx context.Context,
	id uuid.UUID,
) error {

	return s.repo.Archive(ctx, id)
}

func (s *documentTemplateColumnService) DeleteColumn(
	ctx context.Context,
	id uuid.UUID,
) error {

	return s.repo.Delete(ctx, id)
}

func validateDataType(t string) error {
	switch t {
	case "STRING", "TEXT", "INTEGER", "DECIMAL",
		"BOOLEAN", "DATE", "DATETIME", "TIME",
		"ENUM", "UUID", "EMAIL", "PHONE", "JSON":
		return nil
	default:
		return ErrInvalidDataType
	}
}

func mapColumn(c db.DocumentTemplateColumn) *model.DocumentTemplateColumn {
	return &model.DocumentTemplateColumn{
		ID:          c.ID,
		SheetID:     c.SheetID,
		Key:         c.ColumnKey,
		Name:        c.ColumnName,
		DisplayName: toString(c.DisplayName),
		DataType:    string(c.DataType),
		Required:    c.Required,
		IsUnique:    c.IsUnique,

		DefaultValue: nullStringToPtr(c.DefaultValue),

		AllowedValues: mustSliceString(c.AllowedValues),
		Aliases:       mustSliceString(c.Aliases),

		Configuration: mustMap(c.Configuration),
	}
}

func toString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func mustSliceString(v json.RawMessage) []string {
	if len(v) == 0 {
		return nil
	}
	var arr []string
	_ = json.Unmarshal(v, &arr)
	return arr
}

func mustMap(v json.RawMessage) map[string]any {
	if len(v) == 0 {
		return map[string]any{}
	}
	var m map[string]any
	_ = json.Unmarshal(v, &m)
	return m
}

func toJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

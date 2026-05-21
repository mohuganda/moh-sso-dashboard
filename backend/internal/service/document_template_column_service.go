package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
	documentTemplateColumnRepo "github.com/moh-sso-dashboard/internal/repository/document_template_column"
)

var (
	ErrColumnKeyExists  = errors.New("column key already exists")
	ErrColumnNameExists = errors.New("column name already exists")
	ErrInvalidDataType  = errors.New("invalid column data type")
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

	// runtime processing
	Handle(ctx context.Context, col model.DocumentTemplateColumn, value any) (*model.ProcessedColumnValue, error)
}

type documentTemplateColumnService struct {
	repo documentTemplateColumnRepo.DocumentTemplateColumnRepository
}

func NewDocumentTemplateColumnService(
	repo documentTemplateColumnRepo.DocumentTemplateColumnRepository,
) DocumentTemplateColumnService {
	return &documentTemplateColumnService{repo: repo}
}

/* =========================================================
 * CRUD
 * ========================================================= */

func (s *documentTemplateColumnService) CreateColumn(
	ctx context.Context,
	req model.CreateColumnRequest,
) (*model.DocumentTemplateColumn, error) {

	exists, err := s.repo.ExistsKey(ctx, req.SheetID, req.Key)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrColumnKeyExists
	}

	nameExists, err := s.repo.ExistsName(ctx, req.SheetID, req.Name)
	if err != nil {
		return nil, err
	}
	if nameExists {
		return nil, ErrColumnNameExists
	}

	if err := validateDataType(req.DataType); err != nil {
		return nil, err
	}

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

	out := make([]model.DocumentTemplateColumn, 0, len(items))
	for _, c := range items {
		out = append(out, *mapColumn(c))
	}
	return out, nil
}

func (s *documentTemplateColumnService) ListRequiredColumns(
	ctx context.Context,
	sheetID uuid.UUID,
) ([]model.DocumentTemplateColumn, error) {

	items, err := s.repo.ListRequired(ctx, sheetID)
	if err != nil {
		return nil, err
	}

	out := make([]model.DocumentTemplateColumn, 0, len(items))
	for _, c := range items {
		out = append(out, *mapColumn(c))
	}
	return out, nil
}

func (s *documentTemplateColumnService) ListUniqueColumns(
	ctx context.Context,
	sheetID uuid.UUID,
) ([]model.DocumentTemplateColumn, error) {

	items, err := s.repo.ListUnique(ctx, sheetID)
	if err != nil {
		return nil, err
	}

	out := make([]model.DocumentTemplateColumn, 0, len(items))
	for _, c := range items {
		out = append(out, *mapColumn(c))
	}
	return out, nil
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

func (s *documentTemplateColumnService) ArchiveColumn(ctx context.Context, id uuid.UUID) error {
	return s.repo.Archive(ctx, id)
}

func (s *documentTemplateColumnService) DeleteColumn(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

/* =========================================================
 * RUNTIME HANDLER
 * ========================================================= */

func (s *documentTemplateColumnService) Handle(
	ctx context.Context,
	col model.DocumentTemplateColumn,
	value any,
) (*model.ProcessedColumnValue, error) {

	result := &model.ProcessedColumnValue{
		ColumnKey: col.Key,
		RawValue:  value,
		IsValid:   true,
		Value:     value,
	}

	if value == nil || value == "" {
		if col.Required {
			result.IsValid = false
			result.Error = &model.ColumnError{
				Code:    "REQUIRED_FIELD",
				Message: col.Name + " is required",
			}
		}

		if col.DefaultValue != nil {
			result.Value = *col.DefaultValue
		}
		return result, nil
	}

	switch col.DataType {

	case "STRING", "TEXT":
		result.Value = fmt.Sprintf("%v", value)

	case "INTEGER":
		v, err := strconv.ParseInt(fmt.Sprintf("%v", value), 10, 64)
		if err != nil {
			result.IsValid = false
		}
		result.Value = v

	case "DECIMAL":
		v, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
		if err != nil {
			result.IsValid = false
		}
		result.Value = v

	case "BOOLEAN":
		s := strings.ToLower(fmt.Sprintf("%v", value))
		result.Value = (s == "true" || s == "1" || s == "yes")

	case "ENUM":
		str := fmt.Sprintf("%v", value)
		valid := false
		for _, v := range col.AllowedValues {
			if v == str {
				valid = true
				break
			}
		}
		if !valid {
			result.IsValid = false
		}
		result.Value = str

	default:
		result.IsValid = false
		result.Error = &model.ColumnError{
			Code:    "INVALID_TYPE",
			Message: "unsupported type",
		}
	}

	return result, nil
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

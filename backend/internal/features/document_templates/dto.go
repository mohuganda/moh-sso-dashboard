package document_templates

import (
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/model"
)

type processValueRequest struct {
	Column model.DocumentTemplateColumn `json:"column"`
	Value  any                          `json:"value"`
}

type ExistsResponse struct {
	Exists bool `json:"exists"`
}

type TemplateResponse struct {
	ID            uuid.UUID      `json:"id"`
	DocumentID    *uuid.UUID     `json:"document_id,omitempty"`
	Code          string         `json:"code"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	FileType      string         `json:"file_type"`
	Version       int            `json:"version"`
	IsActive      bool           `json:"is_active"`
	Configuration map[string]any `json:"configuration"`
	CreatedBy     uuid.UUID      `json:"created_by"`
	CreatedAt     string         `json:"created_at,omitempty"`
	UpdatedAt     string         `json:"updated_at,omitempty"`
	ArchivedAt    *string        `json:"archived_at,omitempty"`
}

type SheetResponse struct {
	ID                    uuid.UUID      `json:"id"`
	TemplateID            uuid.UUID      `json:"template_id"`
	Code                  string         `json:"code"`
	Name                  string         `json:"name"`
	DisplayName           string         `json:"display_name"`
	Required              bool           `json:"required"`
	SheetOrder            *int           `json:"sheet_order"`
	HeaderRow             int            `json:"header_row"`
	StartRow              int            `json:"start_row"`
	AllowExtraColumns     bool           `json:"allow_extra_columns"`
	AllowDuplicateHeaders bool           `json:"allow_duplicate_headers"`
	Configuration         map[string]any `json:"configuration"`
	CreatedAt             string         `json:"created_at,omitempty"`
	UpdatedAt             string         `json:"updated_at,omitempty"`
	ArchivedAt            *string        `json:"archived_at,omitempty"`
}

type ColumnResponse struct {
	ID            uuid.UUID      `json:"id"`
	SheetID       uuid.UUID      `json:"sheet_id"`
	ColumnKey     string         `json:"column_key"`
	ColumnName    string         `json:"column_name"`
	DisplayName   string         `json:"display_name"`
	DataType      string         `json:"data_type"`
	Required      bool           `json:"required"`
	IsUnique      bool           `json:"is_unique"`
	ColumnOrder   *int           `json:"column_order"`
	DefaultValue  *string        `json:"default_value"`
	AllowedValues []string       `json:"allowed_values"`
	Aliases       []string       `json:"aliases"`
	Configuration map[string]any `json:"configuration"`
	CreatedAt     string         `json:"created_at,omitempty"`
	UpdatedAt     string         `json:"updated_at,omitempty"`
	ArchivedAt    *string        `json:"archived_at,omitempty"`
}

type TemplateRuntimeResponse struct {
	ID       uuid.UUID                       `json:"id"`
	Code     string                          `json:"code"`
	Version  int                             `json:"version"`
	IsActive bool                            `json:"is_active"`
	Sheets   map[string]SheetRuntimeResponse `json:"sheets"`
}

type SheetRuntimeResponse struct {
	ID                    uuid.UUID                        `json:"id"`
	Code                  string                           `json:"code"`
	Name                  string                           `json:"name"`
	DisplayName           string                           `json:"display_name"`
	Required              bool                             `json:"required"`
	SheetOrder            *int                             `json:"sheet_order"`
	HeaderRow             int                              `json:"header_row"`
	StartRow              int                              `json:"start_row"`
	AllowExtraColumns     bool                             `json:"allow_extra_columns"`
	AllowDuplicateHeaders bool                             `json:"allow_duplicate_headers"`
	Configuration         map[string]any                   `json:"configuration"`
	Columns               map[string]ColumnRuntimeResponse `json:"columns"`
}

type ColumnRuntimeResponse struct {
	ID            uuid.UUID      `json:"id"`
	ColumnKey     string         `json:"column_key"`
	ColumnName    string         `json:"column_name"`
	DisplayName   string         `json:"display_name"`
	DataType      string         `json:"data_type"`
	Required      bool           `json:"required"`
	IsUnique      bool           `json:"is_unique"`
	ColumnOrder   *int           `json:"column_order"`
	DefaultValue  *string        `json:"default_value"`
	AllowedValues []string       `json:"allowed_values"`
	Aliases       []string       `json:"aliases"`
	Configuration map[string]any `json:"configuration"`
}

type TemplateStructureResponse struct {
	Template TemplateResponse         `json:"template"`
	Sheets   []SheetStructureResponse `json:"sheets"`
}

type SheetStructureResponse struct {
	ID                    uuid.UUID                 `json:"id"`
	TemplateID            uuid.UUID                 `json:"template_id"`
	Code                  string                    `json:"code"`
	Name                  string                    `json:"name"`
	DisplayName           string                    `json:"display_name"`
	Required              bool                      `json:"required"`
	SheetOrder            *int                      `json:"sheet_order"`
	HeaderRow             int                       `json:"header_row"`
	StartRow              int                       `json:"start_row"`
	AllowExtraColumns     bool                      `json:"allow_extra_columns"`
	AllowDuplicateHeaders bool                      `json:"allow_duplicate_headers"`
	Configuration         map[string]any            `json:"configuration"`
	Columns               []ColumnStructureResponse `json:"columns"`
}

type ColumnStructureResponse struct {
	ID            uuid.UUID      `json:"id"`
	SheetID       uuid.UUID      `json:"sheet_id"`
	ColumnKey     string         `json:"column_key"`
	ColumnName    string         `json:"column_name"`
	DisplayName   string         `json:"display_name"`
	DataType      string         `json:"data_type"`
	Required      bool           `json:"required"`
	IsUnique      bool           `json:"is_unique"`
	ColumnOrder   *int           `json:"column_order"`
	DefaultValue  *string        `json:"default_value"`
	AllowedValues []string       `json:"allowed_values"`
	Aliases       []string       `json:"aliases"`
	Configuration map[string]any `json:"configuration"`
}

type ProcessedColumnValueResponse struct {
	ColumnKey string               `json:"column_key"`
	RawValue  any                  `json:"raw_value"`
	Value     any                  `json:"value"`
	IsValid   bool                 `json:"is_valid"`
	Error     *ColumnErrorResponse `json:"error,omitempty"`
}

type ColumnErrorResponse struct {
	Code         string `json:"code"`
	Message      string `json:"message"`
	InvalidValue any    `json:"invalid_value"`
}

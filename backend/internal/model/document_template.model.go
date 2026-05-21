package model

import "github.com/google/uuid"

//
// =====================================================
// TEMPLATE DOMAIN MODELS
// =====================================================
//

type DocumentTemplate struct {
	ID            uuid.UUID
	Code          string
	Name          string
	Description   string
	FileType      string
	Version       int
	IsActive      bool
	Configuration map[string]any
}

type DocumentTemplateSheet struct {
	ID                    uuid.UUID
	TemplateID            uuid.UUID
	Code                  string
	Name                  string
	DisplayName           string
	Required              bool
	SheetOrder            *int
	HeaderRow             int
	StartRow              int
	AllowExtraColumns     bool
	AllowDuplicateHeaders bool
	Configuration         map[string]any
}

type DocumentTemplateColumn struct {
	ID            uuid.UUID
	SheetID       uuid.UUID
	Key           string
	Name          string
	DisplayName   string
	DataType      string
	Required      bool
	IsUnique      bool
	DefaultValue  *string
	AllowedValues []string
	Aliases       []string
	Configuration map[string]any
}

//
// =====================================================
// TEMPLATE RUNTIME MODELS
// =====================================================
// Used during ingestion/validation
// =====================================================
//

type TemplateRuntime struct {
	ID       uuid.UUID
	Code     string
	Version  int
	IsActive bool
	Sheets   map[string]*TemplateSheetRuntime
}

type TemplateSheetRuntime struct {
	ID        uuid.UUID
	Code      string
	Name      string
	Required  bool
	HeaderRow int
	StartRow  int
	Columns   map[string]*TemplateColumnRuntime
}

type TemplateColumnRuntime struct {
	ID            uuid.UUID
	Key           string
	Name          string
	DataType      string
	Required      bool
	IsUnique      bool
	DefaultValue  *string
	AllowedValues []string
	Aliases       []string
	Configuration map[string]any
}

//
// =====================================================
// TEMPLATE STRUCTURE DTOs
// =====================================================
// Used for frontend/UI/API responses
// =====================================================
//

type TemplateStructure struct {
	ID       uuid.UUID
	Code     string
	Name     string
	Version  int
	IsActive bool
	Sheets   []TemplateSheetStructure
}

type TemplateSheetStructure struct {
	ID        uuid.UUID
	Code      string
	Name      string
	Required  bool
	HeaderRow int
	StartRow  int
	Columns   []TemplateColumnStructure
}

type TemplateColumnStructure struct {
	ID            uuid.UUID
	Key           string
	Name          string
	DisplayName   string
	DataType      string
	Required      bool
	IsUnique      bool
	DefaultValue  *string
	AllowedValues []string
	Aliases       []string
	Configuration map[string]any
}

//
// =====================================================
// INGESTION / VALIDATION
// =====================================================
//

type ProcessedColumnValue struct {
	ColumnKey string
	RawValue  any
	Value     any
	IsValid   bool
	Error     *ColumnError
}

type ColumnError struct {
	Code         string
	Message      string
	InvalidValue any
}

//
// =====================================================
// REQUEST MODELS
// =====================================================
//

type CreateTemplateRequest struct {
	Code          string
	Name          string
	Description   string
	FileType      string
	CreatedBy     uuid.UUID
	Configuration map[string]any
}

type UpdateTemplateRequest struct {
	ID            uuid.UUID
	Name          string
	Description   string
	FileType      string
	IsActive      bool
	Configuration map[string]any
}

type CreateSheetRequest struct {
	TemplateID            uuid.UUID
	Code                  string
	Name                  string
	DisplayName           string
	Required              bool
	SheetOrder            *int
	HeaderRow             int
	StartRow              int
	AllowExtraColumns     bool
	AllowDuplicateHeaders bool
	Configuration         map[string]any
}

type UpdateSheetRequest struct {
	ID                    uuid.UUID
	Name                  string
	DisplayName           string
	Required              bool
	SheetOrder            *int
	HeaderRow             int
	StartRow              int
	AllowExtraColumns     bool
	AllowDuplicateHeaders bool
	Configuration         map[string]any
}

type CreateColumnRequest struct {
	SheetID       uuid.UUID
	Key           string
	Name          string
	DisplayName   string
	DataType      string
	Required      bool
	IsUnique      bool
	DefaultValue  *string
	AllowedValues []string
	Aliases       []string
	Configuration map[string]any
}

type UpdateColumnRequest struct {
	ID            uuid.UUID
	Name          string
	DisplayName   string
	DataType      string
	Required      bool
	IsUnique      bool
	DefaultValue  *string
	AllowedValues []string
	Aliases       []string
	Configuration map[string]any
}

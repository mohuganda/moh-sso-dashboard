package model

import "github.com/google/uuid"

type TemplateRuntime struct {
	ID       uuid.UUID
	Code     string
	Version  int
	IsActive bool

	Sheets map[string]*TemplateSheetRuntime
}

type TemplateSheetRuntime struct {
	ID uuid.UUID

	Code string
	Name string

	Required bool

	HeaderRow int
	StartRow  int

	Columns map[string]*DocumentTemplateColumn
}

type DocumentTemplateColumn struct {
	ID uuid.UUID

	SheetID uuid.UUID

	Key         string
	Name        string
	DisplayName string

	DataType string

	Required bool
	IsUnique bool

	DefaultValue *string

	AllowedValues []string
	Aliases       []string

	Configuration map[string]any
}

type TemplateStructure struct {
	ID       uuid.UUID
	Code     string
	Version  int
	IsActive bool

	Sheets []TemplateSheetStructure
}

type TemplateSheetStructure struct {
	ID       uuid.UUID
	Code     string
	Name     string
	Required bool

	HeaderRow int
	StartRow  int

	Columns []DocumentTemplateColumn
}

type ProcessedColumnValue struct {
	ColumnKey string

	RawValue any

	// Final converted value ready for DB
	Value any

	IsValid bool

	Error *ColumnError
}
type ColumnError struct {
	Code    string
	Message string

	InvalidValue any
}

type CreateColumnRequest struct {
	SheetID uuid.UUID

	Key         string
	Name        string
	DisplayName string

	DataType string

	Required bool
	IsUnique bool

	DefaultValue *string

	AllowedValues []string
	Aliases       []string

	Configuration map[string]any
}

type UpdateColumnRequest struct {
	ID uuid.UUID

	Name        string
	DisplayName string
	DataType    string

	Required bool
	IsUnique bool

	DefaultValue *string

	AllowedValues []string
	Aliases       []string

	Configuration map[string]any
}

type CreateTemplateRequest struct {
	Code        string
	Name        string
	Description string
	FileType    string

	CreatedBy uuid.UUID

	Configuration map[string]any
}

type UpdateTemplateRequest struct {
	ID uuid.UUID

	Name        string
	Description string
	FileType    string

	Configuration map[string]any

	IsActive bool
}

type DocumentTemplate struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Description string
	FileType    string

	Version  int
	IsActive bool

	Configuration map[string]any
}

type TemplateSheet struct {
	ID        uuid.UUID
	Code      string
	Name      string
	Required  bool
	HeaderRow int
	StartRow  int

	Columns []TemplateColumn
}

type TemplateColumn struct {
	ID       uuid.UUID
	Key      string
	Name     string
	DataType string
	Required bool
	Unique   bool

	AllowedValues []string
	Aliases       []string

	Config map[string]any
}

type CreateSheetRequest struct {
	TemplateID uuid.UUID

	Code        string
	Name        string
	DisplayName string

	Required bool

	SheetOrder *int

	HeaderRow int
	StartRow  int

	AllowExtraColumns    bool
	AllowDuplicateHeader bool

	Configuration map[string]any
}

type UpdateSheetRequest struct {
	ID uuid.UUID

	Name        string
	DisplayName string

	Required bool

	SheetOrder *int

	HeaderRow int
	StartRow  int

	AllowExtraColumns    bool
	AllowDuplicateHeader bool

	Configuration map[string]any
}

type DocumentTemplateSheet struct {
	ID uuid.UUID

	TemplateID uuid.UUID

	Code        string
	Name        string
	DisplayName string

	Required bool

	SheetOrder *int

	HeaderRow int
	StartRow  int

	AllowExtraColumns    bool
	AllowDuplicateHeader bool

	Configuration map[string]any
}

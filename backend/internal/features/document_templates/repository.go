package document_templates

import (
	templaterepo "github.com/moh-sso-dashboard/internal/repository/document_template"
	columnrepo "github.com/moh-sso-dashboard/internal/repository/document_template_column"
	sheetrepo "github.com/moh-sso-dashboard/internal/repository/document_template_sheet"
)

type Repository = templaterepo.DocumentTemplateRepository
type SheetRepository = sheetrepo.DocumentTemplateSheetRepository
type ColumnRepository = columnrepo.DocumentTemplateColumnRepository

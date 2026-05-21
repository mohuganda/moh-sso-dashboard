package handler

import (
	"context"
	"errors"

	"github.com/moh-sso-dashboard/internal/model"
)

var (
	ErrSheetNotFound = errors.New("sheet not found in template")
)

type DocumentTemplateSheetHandler interface {
	GetSheet(ctx context.Context, tpl *model.TemplateRuntime, sheetCode string) (*model.TemplateSheetRuntime, error)

	ValidateSheetExists(ctx context.Context, tpl *model.TemplateRuntime, sheetCode string) error
}

type documentTemplateSheetHandler struct{}

func (h *documentTemplateSheetHandler) GetSheet(
	ctx context.Context,
	tpl *model.TemplateRuntime,
	sheetCode string,
) (*model.TemplateSheetRuntime, error) {

	sheet, exists := tpl.Sheets[sheetCode]
	if !exists {
		return nil, ErrSheetNotFound
	}

	return sheet, nil
}

func (h *documentTemplateSheetHandler) ValidateSheetExists(
	ctx context.Context,
	tpl *model.TemplateRuntime,
	sheetCode string,
) error {

	if _, exists := tpl.Sheets[sheetCode]; !exists {
		return ErrSheetNotFound
	}

	return nil
}

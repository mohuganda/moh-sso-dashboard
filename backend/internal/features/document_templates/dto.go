package document_templates

import "github.com/moh-sso-dashboard/internal/model"

type processValueRequest struct {
	Column model.DocumentTemplateColumn `json:"column"`
	Value  any                          `json:"value"`
}

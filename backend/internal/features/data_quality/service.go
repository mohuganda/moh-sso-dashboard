package data_quality

import "context"

type Service interface {
	CreateIssue(ctx context.Context, input createIssueInput, scope healthContextScope) (issueResponse, error)
	ListIssues(ctx context.Context, limit int, offset int, program string, scope healthContextScope) ([]issueResponse, error)
	ListIssueSummaryByProgram(ctx context.Context, limit int, offset int, scope healthContextScope) ([]issueProgramSummaryResponse, error)
	UpdateIssue(ctx context.Context, input updateIssueInput, scope healthContextScope) (issueResponse, error)
	ResolveIssue(ctx context.Context, input resolveIssueInput, scope healthContextScope) (issueStageResponse, error)
	ListIssueResolutionTransactions(ctx context.Context, issueCode string, limit int, offset int, scope healthContextScope) ([]issueStageResponse, error)
	ImportValidationRules(ctx context.Context, inputs []validationRuleInput) (validationRuleImportResult, error)
	ListValidationRules(ctx context.Context, limit int, offset int) ([]validationRuleResponse, error)
	CountIssues(ctx context.Context, program string, scope healthContextScope) (int64, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) CreateIssue(ctx context.Context, input createIssueInput, scope healthContextScope) (issueResponse, error) {
	return s.repository.CreateIssue(ctx, input, scope)
}

func (s *service) ListIssues(ctx context.Context, limit int, offset int, program string, scope healthContextScope) ([]issueResponse, error) {
	return s.repository.ListIssues(ctx, limit, offset, program, scope)
}

func (s *service) ListIssueSummaryByProgram(ctx context.Context, limit int, offset int, scope healthContextScope) ([]issueProgramSummaryResponse, error) {
	return s.repository.ListIssueSummaryByProgram(ctx, limit, offset, scope)
}

func (s *service) UpdateIssue(ctx context.Context, input updateIssueInput, scope healthContextScope) (issueResponse, error) {
	return s.repository.UpdateIssue(ctx, input, scope)
}

func (s *service) ResolveIssue(ctx context.Context, input resolveIssueInput, scope healthContextScope) (issueStageResponse, error) {
	return s.repository.ResolveIssue(ctx, input, scope)
}

func (s *service) ListIssueResolutionTransactions(ctx context.Context, issueCode string, limit int, offset int, scope healthContextScope) ([]issueStageResponse, error) {
	return s.repository.ListIssueResolutionTransactions(ctx, issueCode, limit, offset, scope)
}

func (s *service) ImportValidationRules(ctx context.Context, inputs []validationRuleInput) (validationRuleImportResult, error) {
	return s.repository.ImportValidationRules(ctx, inputs)
}

func (s *service) ListValidationRules(ctx context.Context, limit int, offset int) ([]validationRuleResponse, error) {
	return s.repository.ListValidationRules(ctx, limit, offset)
}

func (s *service) CountIssues(ctx context.Context, program string, scope healthContextScope) (int64, error) {
	return s.repository.CountIssues(ctx, program, scope)
}

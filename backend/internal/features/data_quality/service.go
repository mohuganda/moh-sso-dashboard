package data_quality

import "context"

type Service interface {
	CreateIssue(ctx context.Context, input createIssueInput) (issueResponse, error)
	ListIssues(ctx context.Context, limit int, offset int, program string) ([]issueResponse, error)
	ListIssueSummaryByProgram(ctx context.Context, limit int, offset int) ([]issueProgramSummaryResponse, error)
	UpdateIssue(ctx context.Context, input updateIssueInput) (issueResponse, error)
	ResolveIssue(ctx context.Context, input resolveIssueInput) (issueStageResponse, error)
	ListIssueResolutionTransactions(ctx context.Context, issueCode string, limit int, offset int) ([]issueStageResponse, error)
	ImportValidationRules(ctx context.Context, inputs []validationRuleInput) (validationRuleImportResult, error)
	ListValidationRules(ctx context.Context, limit int, offset int) ([]validationRuleResponse, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) CreateIssue(ctx context.Context, input createIssueInput) (issueResponse, error) {
	return s.repository.CreateIssue(ctx, input)
}

func (s *service) ListIssues(ctx context.Context, limit int, offset int, program string) ([]issueResponse, error) {
	return s.repository.ListIssues(ctx, limit, offset, program)
}

func (s *service) ListIssueSummaryByProgram(ctx context.Context, limit int, offset int) ([]issueProgramSummaryResponse, error) {
	return s.repository.ListIssueSummaryByProgram(ctx, limit, offset)
}

func (s *service) UpdateIssue(ctx context.Context, input updateIssueInput) (issueResponse, error) {
	return s.repository.UpdateIssue(ctx, input)
}

func (s *service) ResolveIssue(ctx context.Context, input resolveIssueInput) (issueStageResponse, error) {
	return s.repository.ResolveIssue(ctx, input)
}

func (s *service) ListIssueResolutionTransactions(ctx context.Context, issueCode string, limit int, offset int) ([]issueStageResponse, error) {
	return s.repository.ListIssueResolutionTransactions(ctx, issueCode, limit, offset)
}

func (s *service) ImportValidationRules(ctx context.Context, inputs []validationRuleInput) (validationRuleImportResult, error) {
	return s.repository.ImportValidationRules(ctx, inputs)
}

func (s *service) ListValidationRules(ctx context.Context, limit int, offset int) ([]validationRuleResponse, error) {
	return s.repository.ListValidationRules(ctx, limit, offset)
}

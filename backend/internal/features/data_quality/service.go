package data_quality

import (
	"context"
	"fmt"
	"strings"

	"github.com/moh-sso-dashboard/internal/model"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
)

type Service interface {
	CreateIssue(ctx context.Context, input createIssueInput) (issueResponse, error)
	ListIssues(ctx context.Context, limit int, offset int, program string) ([]issueResponse, error)
	ListIssueSummaryByProgram(ctx context.Context, limit int, offset int) ([]issueProgramSummaryResponse, error)
	UpdateIssue(ctx context.Context, input updateIssueInput) (issueResponse, error)
	ResolveIssue(ctx context.Context, input resolveIssueInput) (issueStageResponse, error)
	AssignIssues(ctx context.Context, input assignIssuesInput) ([]issueStageResponse, error)
	ListIssueResolutionTransactions(ctx context.Context, issueCode string, limit int, offset int) ([]issueStageResponse, error)
	ImportValidationRules(ctx context.Context, inputs []validationRuleInput) (validationRuleImportResult, error)
	ListValidationRules(ctx context.Context, limit int, offset int) ([]validationRuleResponse, error)
	CountIssues(ctx context.Context, program string) (int64, error)
}

type service struct {
	repository   Repository
	emailService sharedservice.EmailService
}

func NewService(repository Repository, emailService ...sharedservice.EmailService) Service {
	var emailSvc sharedservice.EmailService
	if len(emailService) > 0 {
		emailSvc = emailService[0]
	}
	return &service{
		repository:   repository,
		emailService: emailSvc,
	}
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

func (s *service) AssignIssues(ctx context.Context, input assignIssuesInput) ([]issueStageResponse, error) {
	results, err := s.repository.AssignIssues(ctx, input)
	if err != nil {
		return nil, err
	}

	if s.emailService != nil && strings.TrimSpace(input.AssignedTo) != "" {
		go s.sendAssignmentEmail(context.Background(), input)
	}

	return results, nil
}

func (s *service) sendAssignmentEmail(ctx context.Context, input assignIssuesInput) {
	assignedToEmail := strings.TrimSpace(input.AssignedTo)
	if assignedToEmail == "" {
		return
	}

	issueCodes := input.IssueCodes
	if len(issueCodes) == 0 {
		return
	}

	subject := fmt.Sprintf("[MOH Issue Tracker] %d Issue(s) Assigned to You", len(issueCodes))
	if len(issueCodes) == 1 {
		subject = fmt.Sprintf("[MOH Issue Tracker] Issue %s Assigned to You", issueCodes[0])
	}

	var bodyBuilder strings.Builder
	bodyBuilder.WriteString("<h2>MOH Issue Tracker Notification</h2>")
	bodyBuilder.WriteString(fmt.Sprintf("<p>Hello,</p><p>The following issue(s) have been assigned to you by <strong>%s</strong>:</p>", input.AssignedBy))
	bodyBuilder.WriteString("<ul>")
	for _, code := range issueCodes {
		bodyBuilder.WriteString(fmt.Sprintf("<li><strong>%s</strong></li>", code))
	}
	bodyBuilder.WriteString("</ul>")

	if strings.TrimSpace(input.Comment) != "" {
		bodyBuilder.WriteString(fmt.Sprintf("<p><strong>Notes/Comment:</strong> %s</p>", input.Comment))
	}

	bodyBuilder.WriteString("<p>Please log in to the MOH SSO Dashboard Issue Tracker to review and take action.</p>")
	bodyBuilder.WriteString("<hr/><p style='font-size:12px;color:#777;'>This is an automated notification from the MOH Integrated Health Portal Issue Tracker.</p>")

	textBody := fmt.Sprintf("Issue Assignment Notification\n\nThe following issue(s) have been assigned to you by %s: %s\n\nNotes: %s\n\nPlease log in to the MOH SSO Dashboard to view.",
		input.AssignedBy, strings.Join(issueCodes, ", "), input.Comment)

	msg := model.Message{
		To: []model.Address{
			{Email: assignedToEmail},
		},
		Subject:  subject,
		HTMLBody: bodyBuilder.String(),
		TextBody: textBody,
	}

	_ = s.emailService.Send(ctx, msg)
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

func (s *service) CountIssues(ctx context.Context, program string) (int64, error) {
	return s.repository.CountIssues(ctx, program)
}

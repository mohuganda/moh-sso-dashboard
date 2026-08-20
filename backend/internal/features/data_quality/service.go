package data_quality

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/email"
	"github.com/moh-sso-dashboard/internal/model"
	sharedservice "github.com/moh-sso-dashboard/internal/service"
)

func isNilInterface(i interface{}) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

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
	cfg          *config.Config
}

func NewService(repository Repository, emailService sharedservice.EmailService, cfg ...*config.Config) Service {
	var appCfg *config.Config
	if len(cfg) > 0 {
		appCfg = cfg[0]
	}
	return &service{
		repository:   repository,
		emailService: emailService,
		cfg:          appCfg,
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

	if strings.TrimSpace(input.AssignedTo) != "" {
		go s.sendAssignmentEmail(context.Background(), input)
	}

	return results, nil
}

func (s *service) sendAssignmentEmail(ctx context.Context, input assignIssuesInput) {
	defer func() {
		if r := recover(); r != nil {
			// Prevent panic from crashing background goroutine
		}
	}()

	assignedToEmail := strings.TrimSpace(input.AssignedTo)
	if assignedToEmail == "" {
		return
	}

	issueCodes := input.IssueCodes
	if len(issueCodes) == 0 {
		return
	}

	// Fetch detailed issue records including description
	fetchedIssues, _ := s.repository.GetIssuesByCodes(ctx, issueCodes)
	issueMap := make(map[string]issueResponse, len(fetchedIssues))
	for _, issue := range fetchedIssues {
		if issue.IssueCode != nil {
			issueMap[*issue.IssueCode] = issue
		}
	}

	// Determine portal link
	portalURL := "https://dashboards.health.go.ug/portal"
	if s.cfg != nil && strings.TrimSpace(s.cfg.FrontendBaseURL) != "" {
		portalURL = strings.TrimRight(strings.TrimSpace(s.cfg.FrontendBaseURL), "/")
	}

	subject := fmt.Sprintf("[MOH Issue Tracker] %d Issue(s) Assigned to You", len(issueCodes))
	if len(issueCodes) == 1 {
		subject = fmt.Sprintf("[MOH Issue Tracker] Issue %s Assigned to You", issueCodes[0])
	}

	var bodyBuilder strings.Builder
	var textBuilder strings.Builder

	bodyBuilder.WriteString("<p>Hello,</p><p>The following issue(s) have been assigned to you:</p>")
	bodyBuilder.WriteString("<div style='margin: 16px 0;'>")

	textBuilder.WriteString("Hello,\n\nThe following issue(s) have been assigned to you:\n\n")

	for _, code := range issueCodes {
		issueDetail := issueMap[code]
		desc := "No description provided"
		if issueDetail.Issue != nil && strings.TrimSpace(*issueDetail.Issue) != "" {
			desc = strings.TrimSpace(*issueDetail.Issue)
		}

		datasetStr := ""
		if issueDetail.Dataset != nil {
			datasetStr = *issueDetail.Dataset
		}

		orgUnitStr := ""
		if issueDetail.OrgUnit != nil {
			orgUnitStr = *issueDetail.OrgUnit
		}

		dataElementStr := ""
		if issueDetail.DataElement != nil {
			dataElementStr = *issueDetail.DataElement
		}

		// HTML Card for each issue
		bodyBuilder.WriteString("<div style='border: 1px solid #e0e0e0; border-left: 4px solid #0f62fe; padding: 12px 16px; margin-bottom: 12px; background-color: #f8f9fa; border-radius: 4px;'>")
		bodyBuilder.WriteString(fmt.Sprintf("<div style='font-size: 15px; font-weight: bold; color: #0f62fe;'>Issue Code: %s</div>", code))
		bodyBuilder.WriteString(fmt.Sprintf("<div style='margin-top: 6px; font-size: 14px; color: #161616;'><strong>Description:</strong> %s</div>", desc))

		metaParts := make([]string, 0, 3)
		if datasetStr != "" {
			metaParts = append(metaParts, fmt.Sprintf("<strong>Dataset:</strong> %s", datasetStr))
		}
		if orgUnitStr != "" {
			metaParts = append(metaParts, fmt.Sprintf("<strong>Organization Unit:</strong> %s", orgUnitStr))
		}
		if dataElementStr != "" {
			metaParts = append(metaParts, fmt.Sprintf("<strong>Data Element:</strong> %s", dataElementStr))
		}
		if len(metaParts) > 0 {
			bodyBuilder.WriteString(fmt.Sprintf("<div style='margin-top: 6px; font-size: 12px; color: #525252;'>%s</div>", strings.Join(metaParts, " &nbsp;|&nbsp; ")))
		}
		bodyBuilder.WriteString("</div>")

		// Text format for each issue
		textBuilder.WriteString(fmt.Sprintf("• Issue Code: %s\n  Description: %s\n", code, desc))
		if datasetStr != "" || orgUnitStr != "" {
			textBuilder.WriteString(fmt.Sprintf("  Dataset: %s | Data Element: %s | Organization Unit: %s\n", datasetStr, dataElementStr, orgUnitStr))
		}
		textBuilder.WriteString("\n")
	}
	bodyBuilder.WriteString("</div>")

	if strings.TrimSpace(input.Comment) != "" {
		bodyBuilder.WriteString(fmt.Sprintf("<div style='background-color: #edf5ff; border: 1px solid #c6e2ff; padding: 10px 14px; border-radius: 4px; margin-bottom: 16px;'><strong>Comment:</strong> %s</div>", strings.TrimSpace(input.Comment)))
		textBuilder.WriteString(fmt.Sprintf("Comment: %s\n\n", strings.TrimSpace(input.Comment)))
	}

	bodyBuilder.WriteString(fmt.Sprintf("<p style='margin-top: 20px;'><a href='%s' style='background-color: #0f62fe; color: #ffffff; padding: 10px 20px; text-decoration: none; border-radius: 4px; font-weight: bold; display: inline-block;'>Open MOH Dashboard Portal &rarr;</a></p>", portalURL))
	bodyBuilder.WriteString(fmt.Sprintf("<p style='font-size: 13px; color: #6f6f6f; margin-top: 12px;'>Or visit the portal directly: <a href='%s' style='color: #0f62fe;'>%s</a></p>", portalURL, portalURL))
	bodyBuilder.WriteString("<hr style='border: none; border-top: 1px solid #e0e0e0; margin-top: 24px;'/>")
	bodyBuilder.WriteString("<p style='font-size: 12px; color: #777;'>This is an automated notification from the MOH Integrated Health Portal Issue Tracker.</p>")

	textBuilder.WriteString(fmt.Sprintf("Please log in to the MOH Dashboard Portal to review and take action:\n%s\n", portalURL))

	msg := model.Message{
		To: []model.Address{
			{Email: assignedToEmail},
		},
		Subject:  subject,
		HTMLBody: bodyBuilder.String(),
		TextBody: textBuilder.String(),
	}

	// Dedicated email function using SMTP_HOST_GM configurations (Gmail SMTP)
	if err := email.SendEmailViaGM(ctx, msg, s.cfg); err != nil {
		if !isNilInterface(s.emailService) {
			_ = s.emailService.Send(ctx, msg)
		}
	}
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

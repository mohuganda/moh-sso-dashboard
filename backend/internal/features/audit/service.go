package audit

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (AuditLogListResponse, error)
	GetAuditLog(ctx context.Context, id uuid.UUID) (AuditLogResponse, error)
	ListAuditActions(ctx context.Context) (AuditActionsResponse, error)
	AuditMetricsOverview(ctx context.Context, input AuditWindowInput) (AuditMetricsOverviewResponse, error)
	FailedLoginsByDay(ctx context.Context, input AuditWindowInput) (AuditFailedLoginsByDayResponse, error)
	TopFailureIPs(ctx context.Context, input TopFailureIPsInput) (AuditTopFailureIPsResponse, error)
	ExportAuditLogs(ctx context.Context, input ExportAuditLogsInput) ([]AuditLogResponse, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (AuditLogListResponse, error) {
	return s.repository.ListAuditLogs(ctx, input)
}

func (s *service) GetAuditLog(ctx context.Context, id uuid.UUID) (AuditLogResponse, error) {
	return s.repository.GetAuditLog(ctx, id)
}

func (s *service) ListAuditActions(ctx context.Context) (AuditActionsResponse, error) {
	return s.repository.ListAuditActions(ctx)
}

func (s *service) AuditMetricsOverview(ctx context.Context, input AuditWindowInput) (AuditMetricsOverviewResponse, error) {
	return s.repository.AuditMetricsOverview(ctx, input)
}

func (s *service) FailedLoginsByDay(ctx context.Context, input AuditWindowInput) (AuditFailedLoginsByDayResponse, error) {
	return s.repository.FailedLoginsByDay(ctx, input)
}

func (s *service) TopFailureIPs(ctx context.Context, input TopFailureIPsInput) (AuditTopFailureIPsResponse, error) {
	return s.repository.TopFailureIPs(ctx, input)
}

func (s *service) ExportAuditLogs(ctx context.Context, input ExportAuditLogsInput) ([]AuditLogResponse, error) {
	return s.repository.ExportAuditLogs(ctx, input)
}

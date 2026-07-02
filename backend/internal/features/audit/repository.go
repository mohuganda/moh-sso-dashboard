package audit

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type Repository interface {
	ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (AuditLogListResponse, error)
	GetAuditLog(ctx context.Context, id uuid.UUID) (AuditLogResponse, error)
	ListAuditActions(ctx context.Context) (AuditActionsResponse, error)
	AuditMetricsOverview(ctx context.Context, input AuditWindowInput) (AuditMetricsOverviewResponse, error)
	FailedLoginsByDay(ctx context.Context, input AuditWindowInput) (AuditFailedLoginsByDayResponse, error)
	TopFailureIPs(ctx context.Context, input TopFailureIPsInput) (AuditTopFailureIPsResponse, error)
	ExportAuditLogs(ctx context.Context, input ExportAuditLogsInput) ([]AuditLogResponse, error)
}

type postgresRepository struct {
	store db.Store
}

func NewRepository(store db.Store) Repository {
	return &postgresRepository{store: store}
}

func (r *postgresRepository) ListAuditLogs(ctx context.Context, input ListAuditLogsInput) (AuditLogListResponse, error) {
	rows, err := r.store.ListAuditLogs(
		ctx,
		db.ListAuditLogsParams{
			StartTime: input.StartTime,
			EndTime:   input.EndTime,

			Action:   toNullString(input.Action),
			UserID:   toNullUUID(input.UserID),
			ClientID: toNullString(input.ClientID),
			Ip:       toNullString(input.IP),
			Success:  toNullString(input.Success),

			CursorCreatedAt: toNullTime(input.CursorCreatedAt),
			CursorID:        toNullUUID(input.CursorID),

			RowLimit: input.Limit + 1,
		},
	)
	if err != nil {
		return AuditLogListResponse{}, err
	}

	hasMore := len(rows) > int(input.Limit)
	if hasMore {
		rows = rows[:input.Limit]
	}

	var nextCreatedAt *time.Time
	var nextID *uuid.UUID
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		if last.CreatedAt.Valid {
			t := last.CreatedAt.Time
			nextCreatedAt = &t
		}
		id := last.ID
		nextID = &id
	}

	return toAuditLogListResponse(
		toAuditLogResponses(rows),
		toAuditCursorResponse(nextCreatedAt, nextID),
		hasMore,
	), nil
}

func (r *postgresRepository) GetAuditLog(ctx context.Context, id uuid.UUID) (AuditLogResponse, error) {
	row, err := r.store.GetAuditLog(ctx, id)
	if err != nil {
		return AuditLogResponse{}, err
	}
	return toAuditLogResponseFromGet(row), nil
}

func (r *postgresRepository) ListAuditActions(ctx context.Context) (AuditActionsResponse, error) {
	rows, err := r.store.ListAuditActions(ctx)
	if err != nil {
		return AuditActionsResponse{}, err
	}
	return toAuditActionsResponse(rows), nil
}

func (r *postgresRepository) AuditMetricsOverview(ctx context.Context, input AuditWindowInput) (AuditMetricsOverviewResponse, error) {
	row, err := r.store.AuditMetricsOverview(
		ctx,
		db.AuditMetricsOverviewParams{
			StartTime: toNullTime(&input.StartTime),
			EndTime:   toNullTime(&input.EndTime),
		},
	)
	if err != nil {
		return AuditMetricsOverviewResponse{}, err
	}
	return toAuditMetricsOverviewResponse(row), nil
}

func (r *postgresRepository) FailedLoginsByDay(ctx context.Context, input AuditWindowInput) (AuditFailedLoginsByDayResponse, error) {
	rows, err := r.store.FailedLoginsByDay(
		ctx,
		db.FailedLoginsByDayParams{
			StartTime: toNullTime(&input.StartTime),
			EndTime:   toNullTime(&input.EndTime),
		},
	)
	if err != nil {
		return AuditFailedLoginsByDayResponse{}, err
	}
	return toAuditFailedLoginsByDayResponse(rows), nil
}

func (r *postgresRepository) TopFailureIPs(ctx context.Context, input TopFailureIPsInput) (AuditTopFailureIPsResponse, error) {
	rows, err := r.store.TopFailureIPs(
		ctx,
		db.TopFailureIPsParams{
			StartTime: toNullTime(&input.StartTime),
			EndTime:   toNullTime(&input.EndTime),
			RowLimit:  input.Limit,
		},
	)
	if err != nil {
		return AuditTopFailureIPsResponse{}, err
	}
	return toAuditTopFailureIPsResponse(rows), nil
}

func (r *postgresRepository) ExportAuditLogs(ctx context.Context, input ExportAuditLogsInput) ([]AuditLogResponse, error) {
	rows, err := r.store.ExportAuditLogs(
		ctx,
		db.ExportAuditLogsParams{
			StartTime: toNullTime(&input.StartTime),
			EndTime:   toNullTime(&input.EndTime),
			Action:    toNullString(input.Action),
			UserID:    toNullUUID(input.UserID),
			ClientID:  toNullString(input.ClientID),
			Ip:        toNullString(input.IP),
			Success:   toNullString(input.Success),
		},
	)
	if err != nil {
		return nil, err
	}
	return toExportAuditLogResponses(rows), nil
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func toNullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

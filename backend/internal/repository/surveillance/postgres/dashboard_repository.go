package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type DashboardRepository struct {
	db db.Store
}

func NewDashboardRepository(db db.Store) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) GetNationalWeeklyStatusSummary(ctx context.Context, epiWeekID uuid.UUID) ([]db.GetNationalWeeklyStatusSummaryRow, error) {
	return r.db.GetNationalWeeklyStatusSummary(ctx, epiWeekID)
}

func (r *DashboardRepository) ListNationalWeeklySubjectsByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListNationalWeeklySubjectsByWeekRow, error) {
	return r.db.ListNationalWeeklySubjectsByWeek(ctx, epiWeekID)
}

func (r *DashboardRepository) ListDistrictWeeklySubjectsByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListDistrictWeeklySubjectsByWeekRow, error) {
	return r.db.ListDistrictWeeklySubjectsByWeek(ctx, epiWeekID)
}

func (r *DashboardRepository) ListRegionWeeklySubjectsByWeek(ctx context.Context, epiWeekID uuid.UUID) ([]db.ListRegionWeeklySubjectsByWeekRow, error) {
	return r.db.ListRegionWeeklySubjectsByWeek(ctx, epiWeekID)
}

func (r *DashboardRepository) GetWeeklySubjectTotals(ctx context.Context, epiWeekID uuid.UUID) ([]db.GetWeeklySubjectTotalsRow, error) {
	return r.db.GetWeeklySubjectTotals(ctx, epiWeekID)
}

func (r *DashboardRepository) GetDiseaseDashboardSummaryByWeek(ctx context.Context, arg db.GetDiseaseDashboardSummaryByWeekParams) (db.GetDiseaseDashboardSummaryByWeekRow, error) {
	return r.db.GetDiseaseDashboardSummaryByWeek(ctx, arg)
}

func (r *DashboardRepository) GetIndicatorDashboardSummaryByWeek(ctx context.Context, arg db.GetIndicatorDashboardSummaryByWeekParams) (db.GetIndicatorDashboardSummaryByWeekRow, error) {
	return r.db.GetIndicatorDashboardSummaryByWeek(ctx, arg)
}

func (r *DashboardRepository) GetTopFacilitiesByDiseaseAndWeek(ctx context.Context, arg db.GetTopFacilitiesByDiseaseAndWeekParams) ([]db.GetTopFacilitiesByDiseaseAndWeekRow, error) {
	return r.db.GetTopFacilitiesByDiseaseAndWeek(ctx, arg)
}

func (r *DashboardRepository) GetTopFacilitiesByIndicatorAndWeek(ctx context.Context, arg db.GetTopFacilitiesByIndicatorAndWeekParams) ([]db.GetTopFacilitiesByIndicatorAndWeekRow, error) {
	return r.db.GetTopFacilitiesByIndicatorAndWeek(ctx, arg)
}

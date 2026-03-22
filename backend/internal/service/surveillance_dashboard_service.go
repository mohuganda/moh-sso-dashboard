package service

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type SurveillanceDashboardService struct {
	log           *logger.Logger
	dashboardRepo interfaces.DashboardRepository
}

func NewSurveillanceDashboardService(
	log *logger.Logger,
	dashboardRepo interfaces.DashboardRepository,
) *SurveillanceDashboardService {
	return &SurveillanceDashboardService{
		log:           log,
		dashboardRepo: dashboardRepo,
	}
}

func (s *SurveillanceDashboardService) GetNationalWeeklyStatusSummary(ctx context.Context, epiWeekID uuid.UUID) ([]db.GetNationalWeeklyStatusSummaryRow, error) {
	s.log.Debug(ctx, "getting national weekly status summary")

	items, err := s.dashboardRepo.GetNationalWeeklyStatusSummary(ctx, epiWeekID)
	if err != nil {
		s.log.Error(ctx, "failed to get national weekly status summary", "error", err)
		return []db.GetNationalWeeklyStatusSummaryRow{}, err
	}

	return items, nil
}

func (s *SurveillanceDashboardService) GetDiseaseDashboardSummaryByWeek(ctx context.Context, arg db.GetDiseaseDashboardSummaryByWeekParams) (db.GetDiseaseDashboardSummaryByWeekRow, error) {
	s.log.Debug(ctx, "getting disease dashboard summary by week")

	item, err := s.dashboardRepo.GetDiseaseDashboardSummaryByWeek(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to get disease dashboard summary by week", "error", err)
		return db.GetDiseaseDashboardSummaryByWeekRow{}, err
	}

	return item, nil
}

func (s *SurveillanceDashboardService) GetIndicatorDashboardSummaryByWeek(ctx context.Context, arg db.GetIndicatorDashboardSummaryByWeekParams) (db.GetIndicatorDashboardSummaryByWeekRow, error) {
	s.log.Debug(ctx, "getting indicator dashboard summary by week")

	item, err := s.dashboardRepo.GetIndicatorDashboardSummaryByWeek(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to get indicator dashboard summary by week", "error", err)
		return db.GetIndicatorDashboardSummaryByWeekRow{}, err
	}

	return item, nil
}

func (s *SurveillanceDashboardService) GetTopFacilitiesByDiseaseAndWeek(ctx context.Context, arg db.GetTopFacilitiesByDiseaseAndWeekParams) ([]db.GetTopFacilitiesByDiseaseAndWeekRow, error) {
	s.log.Debug(ctx, "getting top facilities by disease and week")

	items, err := s.dashboardRepo.GetTopFacilitiesByDiseaseAndWeek(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to get top facilities by disease and week", "error", err)
		return []db.GetTopFacilitiesByDiseaseAndWeekRow{}, err
	}

	return items, nil
}

func (s *SurveillanceDashboardService) GetTopFacilitiesByIndicatorAndWeek(ctx context.Context, arg db.GetTopFacilitiesByIndicatorAndWeekParams) ([]db.GetTopFacilitiesByIndicatorAndWeekRow, error) {
	s.log.Debug(ctx, "getting top facilities by indicator and week")

	items, err := s.dashboardRepo.GetTopFacilitiesByIndicatorAndWeek(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to get top facilities by indicator and week", "error", err)
		return []db.GetTopFacilitiesByIndicatorAndWeekRow{}, err
	}

	return items, nil
}

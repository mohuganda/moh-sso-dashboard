package service

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type SurveillanceEpiWeekService struct {
	log         *logger.Logger
	epiWeekRepo interfaces.EpiWeekRepository
}

func NewSurveillanceEpiWeekService(
	log *logger.Logger,
	epiWeekRepo interfaces.EpiWeekRepository,
) *SurveillanceEpiWeekService {
	return &SurveillanceEpiWeekService{
		log:         log,
		epiWeekRepo: epiWeekRepo,
	}
}

func (s *SurveillanceEpiWeekService) GetEpiWeekByID(ctx context.Context, id uuid.UUID) (db.EpiWeek, error) {
	s.log.Debug(ctx, "getting epi week by id", "week_id", id)

	item, err := s.epiWeekRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get epi week by id", "week_id", id, "error", err)
		return db.EpiWeek{}, err
	}

	return item, nil
}

func (s *SurveillanceEpiWeekService) GetEpiWeekByYearWeek(ctx context.Context, arg db.GetEpiWeekByYearWeekParams) (db.EpiWeek, error) {
	s.log.Debug(ctx, "getting epi week by year/week", "year", arg.EpiYear, "week", arg.EpiWeek)

	item, err := s.epiWeekRepo.GetByYearWeek(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to get epi week by year/week", "year", arg.EpiYear, "week", arg.EpiWeek, "error", err)
		return db.EpiWeek{}, err
	}

	return item, nil
}

func (s *SurveillanceEpiWeekService) ListEpiWeeksByYear(ctx context.Context, year int32) ([]db.EpiWeek, error) {
	s.log.Debug(ctx, "listing epi weeks by year", "year", year)

	items, err := s.epiWeekRepo.ListByYear(ctx, year)
	if err != nil {
		s.log.Error(ctx, "failed to list epi weeks by year", "year", year, "error", err)
		return []db.EpiWeek{}, err
	}

	return items, nil
}

func (s *SurveillanceEpiWeekService) UpsertEpiWeek(ctx context.Context, arg db.UpsertEpiWeekParams) (db.EpiWeek, error) {
	s.log.Info(ctx, "upserting epi week", "year", arg.EpiYear, "week", arg.EpiWeek)

	item, err := s.epiWeekRepo.Upsert(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert epi week", "year", arg.EpiYear, "week", arg.EpiWeek, "error", err)
		return db.EpiWeek{}, err
	}

	return item, nil
}

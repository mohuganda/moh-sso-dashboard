package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
	"github.com/moh-sso-dashboard/internal/utils"
)

type SurveillanceEpiWeekService struct {
	log           *logger.Logger
	epiWeekRepo   interfaces.EpiWeekRepository
	notifications NotificationsService
}

func NewSurveillanceEpiWeekService(
	log *logger.Logger,
	epiWeekRepo interfaces.EpiWeekRepository,
	notifications ...NotificationsService,
) *SurveillanceEpiWeekService {
	var notificationSvc NotificationsService
	if len(notifications) > 0 {
		notificationSvc = notifications[0]
	}

	return &SurveillanceEpiWeekService{
		log:           log,
		epiWeekRepo:   epiWeekRepo,
		notifications: notificationSvc,
	}
}

func (s *SurveillanceEpiWeekService) GetEpiWeekByID(
	ctx context.Context,
	id uuid.UUID,
) (db.EpiWeek, error) {
	if s == nil {
		return db.EpiWeek{}, errors.New("surveillance epi week service is nil")
	}

	if s.epiWeekRepo == nil {
		return db.EpiWeek{}, errors.New("epi week repository is nil")
	}

	if id == uuid.Nil {
		return db.EpiWeek{}, errors.New("epi week id is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting epi week by id", "week_id", id)
	}

	item, err := s.epiWeekRepo.GetByID(ctx, id)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get epi week by id", "week_id", id, "error", err)
		}

		return db.EpiWeek{}, fmt.Errorf("get epi week by id: %w", err)
	}

	return item, nil
}

func (s *SurveillanceEpiWeekService) GetEpiWeekByYearWeek(
	ctx context.Context,
	arg db.GetEpiWeekByYearWeekParams,
) (db.EpiWeek, error) {
	if s == nil {
		return db.EpiWeek{}, errors.New("surveillance epi week service is nil")
	}

	if s.epiWeekRepo == nil {
		return db.EpiWeek{}, errors.New("epi week repository is nil")
	}

	if arg.EpiYear <= 0 {
		return db.EpiWeek{}, errors.New("epi year is required")
	}

	if arg.EpiWeek <= 0 {
		return db.EpiWeek{}, errors.New("epi week must be greater than 0")
	}

	if s.log != nil {
		s.log.Debug(ctx, "getting epi week by year/week", "year", arg.EpiYear, "week", arg.EpiWeek)
	}

	item, err := s.epiWeekRepo.GetByYearWeek(ctx, arg)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to get epi week by year/week", "year", arg.EpiYear, "week", arg.EpiWeek, "error", err)
		}

		return db.EpiWeek{}, fmt.Errorf("get epi week by year/week: %w", err)
	}

	return item, nil
}

func (s *SurveillanceEpiWeekService) ListEpiWeeksByYear(
	ctx context.Context,
	year int32,
) ([]db.EpiWeek, error) {
	if s == nil {
		return nil, errors.New("surveillance epi week service is nil")
	}

	if s.epiWeekRepo == nil {
		return nil, errors.New("epi week repository is nil")
	}

	if year <= 0 {
		return nil, errors.New("epi year is required")
	}

	if s.log != nil {
		s.log.Debug(ctx, "listing epi weeks by year", "year", year)
	}

	items, err := s.epiWeekRepo.ListByYear(ctx, year)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to list epi weeks by year", "year", year, "error", err)
		}

		return []db.EpiWeek{}, fmt.Errorf("list epi weeks by year: %w", err)
	}

	return items, nil
}

func (s *SurveillanceEpiWeekService) UpsertEpiWeek(
	ctx context.Context,
	arg db.UpsertEpiWeekParams,
) (db.EpiWeek, error) {
	if s == nil {
		return db.EpiWeek{}, errors.New("surveillance epi week service is nil")
	}

	if s.epiWeekRepo == nil {
		return db.EpiWeek{}, errors.New("epi week repository is nil")
	}

	if arg.EpiYear <= 0 {
		return db.EpiWeek{}, errors.New("epi year is required")
	}

	if arg.EpiWeek <= 0 {
		return db.EpiWeek{}, errors.New("epi week must be greater than 0")
	}

	if arg.EpiWeek > 53 {
		return db.EpiWeek{}, errors.New("epi week must not be greater than 53")
	}

	if s.log != nil {
		s.log.Info(ctx, "upserting epi week", "year", arg.EpiYear, "week", arg.EpiWeek)
	}

	item, err := s.epiWeekRepo.Upsert(ctx, arg)
	if err != nil {
		if s.log != nil {
			s.log.Error(ctx, "failed to upsert epi week", "year", arg.EpiYear, "week", arg.EpiWeek, "error", err)
		}

		return db.EpiWeek{}, fmt.Errorf("upsert epi week: %w", err)
	}

	s.notify(ctx, model.Notification{
		Type:       "SURVEILLANCE_EPI_WEEK_UPSERTED",
		Title:      "EPI week updated",
		Severity:   "info",
		Message:    "Surveillance EPI week created or updated",
		TargetRole: "admin",
		Metadata: utils.MustJSON(map[string]any{
			"epi_week_id": item.ID.String(),
			"epi_year":    item.EpiYear,
			"epi_week":    item.EpiWeek,
			"start_date":  item.WeekStartDate,
			"end_date":    item.WeekEndDate,
		}),
	})

	return item, nil
}

func (s *SurveillanceEpiWeekService) notify(
	ctx context.Context,
	notification model.Notification,
) {
	if s == nil || s.notifications == nil {
		return
	}

	if notification.TargetRole == "" {
		notification.TargetRole = "admin"
	}

	if _, err := s.notifications.Notify(ctx, notification); err != nil {
		if s.log != nil {
			s.log.Error(
				ctx,
				"surveillance epi week notification failed",
				"type", notification.Type,
				"error", err,
			)
		}
	}
}

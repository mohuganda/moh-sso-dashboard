package service

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	logger "github.com/moh-sso-dashboard/internal/log"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
)

type SurveillanceDiseaseService struct {
	log         *logger.Logger
	diseaseRepo interfaces.DiseaseRepository
}

func NewSurveillanceDiseaseService(
	log *logger.Logger,
	diseaseRepo interfaces.DiseaseRepository,
) *SurveillanceDiseaseService {
	return &SurveillanceDiseaseService{
		log:         log,
		diseaseRepo: diseaseRepo,
	}
}

func (s *SurveillanceDiseaseService) ListDiseases(ctx context.Context) ([]db.Disease, error) {
	s.log.Debug(ctx, "listing diseases")

	items, err := s.diseaseRepo.List(ctx)
	if err != nil {
		s.log.Error(ctx, "failed to list diseases", "error", err)
		return []db.Disease{}, err
	}

	return items, nil
}

func (s *SurveillanceDiseaseService) ListActiveDiseases(ctx context.Context) ([]db.Disease, error) {
	s.log.Debug(ctx, "listing active diseases")

	items, err := s.diseaseRepo.ListActive(ctx)
	if err != nil {
		s.log.Error(ctx, "failed to list active diseases", "error", err)
		return []db.Disease{}, err
	}

	return items, nil
}

func (s *SurveillanceDiseaseService) GetDiseaseByID(ctx context.Context, id uuid.UUID) (db.Disease, error) {
	s.log.Debug(ctx, "getting disease by id", "disease_id", id)

	item, err := s.diseaseRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "failed to get disease by id", "disease_id", id, "error", err)
		return db.Disease{}, err
	}

	return item, nil
}

func (s *SurveillanceDiseaseService) GetDiseaseByName(ctx context.Context, name string) (db.Disease, error) {
	s.log.Debug(ctx, "getting disease by name", "name", name)

	item, err := s.diseaseRepo.GetByName(ctx, name)
	if err != nil {
		s.log.Error(ctx, "failed to get disease by name", "name", name, "error", err)
		return db.Disease{}, err
	}

	return item, nil
}

func (s *SurveillanceDiseaseService) CreateDisease(ctx context.Context, arg db.CreateDiseaseParams) (db.Disease, error) {
	s.log.Info(ctx, "creating disease", "name", arg.Name)

	item, err := s.diseaseRepo.Create(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to create disease", "name", arg.Name, "error", err)
		return db.Disease{}, err
	}

	return item, nil
}

func (s *SurveillanceDiseaseService) UpsertDisease(ctx context.Context, arg db.UpsertDiseaseParams) (db.Disease, error) {
	s.log.Info(ctx, "upserting disease", "name", arg.Name)

	item, err := s.diseaseRepo.Upsert(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to upsert disease", "name", arg.Name, "error", err)
		return db.Disease{}, err
	}

	return item, nil
}

func (s *SurveillanceDiseaseService) SetDiseaseActiveState(ctx context.Context, arg db.SetDiseaseActiveStateParams) (db.Disease, error) {
	s.log.Info(ctx, "setting disease active state", "disease_id", arg.ID)

	item, err := s.diseaseRepo.SetActiveState(ctx, arg)
	if err != nil {
		s.log.Error(ctx, "failed to set disease active state", "disease_id", arg.ID, "error", err)
		return db.Disease{}, err
	}

	return item, nil
}

func (s *SurveillanceDiseaseService) DeleteDisease(ctx context.Context, id uuid.UUID) error {
	s.log.Info(ctx, "deleting disease", "disease_id", id)

	if err := s.diseaseRepo.Delete(ctx, id); err != nil {
		s.log.Error(ctx, "failed to delete disease", "disease_id", id, "error", err)
		return err
	}

	return nil
}

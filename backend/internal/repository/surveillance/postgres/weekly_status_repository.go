package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type WeeklyStatusRepository struct {
	db db.Store
}

func NewWeeklyStatusRepository(store db.Store) *WeeklyStatusRepository {
	return &WeeklyStatusRepository{
		db: store,
	}
}

func (r *WeeklyStatusRepository) Create(
	ctx context.Context,
	arg db.CreateWeeklyStatusParams,
) (db.WeeklyStatus, error) {
	return r.db.CreateWeeklyStatus(ctx, arg)
}

func (r *WeeklyStatusRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (db.WeeklyStatus, error) {
	return r.db.GetWeeklyStatusByID(ctx, id)
}

func (r *WeeklyStatusRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	return r.db.DeleteWeeklyStatus(ctx, id)
}

func (r *WeeklyStatusRepository) List(
	ctx context.Context,
	arg db.ListWeeklyStatusesParams,
) ([]db.WeeklyStatus, error) {
	return r.db.ListWeeklyStatuses(ctx, arg)
}

func (r *WeeklyStatusRepository) ListDetailed(
	ctx context.Context,
	arg db.ListWeeklyStatusesDetailedParams,
) ([]db.ListWeeklyStatusesDetailedRow, error) {
	return r.db.ListWeeklyStatusesDetailed(ctx, arg)
}

func (r *WeeklyStatusRepository) ListByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.WeeklyStatus, error) {
	return r.db.ListWeeklyStatusesByWeek(ctx, epiWeekID)
}

func (r *WeeklyStatusRepository) ListByRegion(
	ctx context.Context,
	regionID uuid.UUID,
) ([]db.WeeklyStatus, error) {
	return r.db.ListWeeklyStatusesByRegion(ctx, uuid.NullUUID{
		UUID:  regionID,
		Valid: regionID != uuid.Nil,
	})
}

func (r *WeeklyStatusRepository) ListByDistrict(
	ctx context.Context,
	districtID uuid.UUID,
) ([]db.WeeklyStatus, error) {
	return r.db.ListWeeklyStatusesByDistrict(ctx, uuid.NullUUID{
		UUID:  districtID,
		Valid: districtID != uuid.Nil,
	})
}

func (r *WeeklyStatusRepository) ListBySubCounty(
	ctx context.Context,
	subCountyID uuid.UUID,
) ([]db.WeeklyStatus, error) {
	return r.db.ListWeeklyStatusesBySubCounty(ctx, uuid.NullUUID{
		UUID:  subCountyID,
		Valid: subCountyID != uuid.Nil,
	})
}

func (r *WeeklyStatusRepository) ListNationalByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.WeeklyStatus, error) {
	return r.db.ListNationalWeeklyStatusesByWeek(ctx, epiWeekID)
}

func (r *WeeklyStatusRepository) ListRegionByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.WeeklyStatus, error) {
	return r.db.ListRegionWeeklyStatusesByWeek(ctx, epiWeekID)
}

func (r *WeeklyStatusRepository) ListDistrictByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.WeeklyStatus, error) {
	return r.db.ListDistrictWeeklyStatusesByWeek(ctx, epiWeekID)
}

func (r *WeeklyStatusRepository) ListSubCountyByWeek(
	ctx context.Context,
	epiWeekID uuid.UUID,
) ([]db.WeeklyStatus, error) {
	return r.db.ListSubCountyWeeklyStatusesByWeek(ctx, epiWeekID)
}

func (r *WeeklyStatusRepository) UpsertNationalDisease(
	ctx context.Context,
	arg db.UpsertNationalDiseaseWeeklyStatusParams,
) (db.WeeklyStatus, error) {
	return r.db.UpsertNationalDiseaseWeeklyStatus(ctx, arg)
}

func (r *WeeklyStatusRepository) UpsertNationalIndicator(
	ctx context.Context,
	arg db.UpsertNationalIndicatorWeeklyStatusParams,
) (db.WeeklyStatus, error) {
	return r.db.UpsertNationalIndicatorWeeklyStatus(ctx, arg)
}

func (r *WeeklyStatusRepository) UpsertRegionDisease(
	ctx context.Context,
	arg db.UpsertRegionDiseaseWeeklyStatusParams,
) (db.WeeklyStatus, error) {
	return r.db.UpsertRegionDiseaseWeeklyStatus(ctx, arg)
}

func (r *WeeklyStatusRepository) UpsertRegionIndicator(
	ctx context.Context,
	arg db.UpsertRegionIndicatorWeeklyStatusParams,
) (db.WeeklyStatus, error) {
	return r.db.UpsertRegionIndicatorWeeklyStatus(ctx, arg)
}

func (r *WeeklyStatusRepository) UpsertDistrictDisease(
	ctx context.Context,
	arg db.UpsertDistrictDiseaseWeeklyStatusParams,
) (db.WeeklyStatus, error) {
	return r.db.UpsertDistrictDiseaseWeeklyStatus(ctx, arg)
}

func (r *WeeklyStatusRepository) UpsertDistrictIndicator(
	ctx context.Context,
	arg db.UpsertDistrictIndicatorWeeklyStatusParams,
) (db.WeeklyStatus, error) {
	return r.db.UpsertDistrictIndicatorWeeklyStatus(ctx, arg)
}

func (r *WeeklyStatusRepository) UpsertSubCountyDisease(
	ctx context.Context,
	arg db.UpsertSubCountyDiseaseWeeklyStatusParams,
) (db.WeeklyStatus, error) {
	return r.db.UpsertSubCountyDiseaseWeeklyStatus(ctx, arg)
}

func (r *WeeklyStatusRepository) UpsertSubCountyIndicator(
	ctx context.Context,
	arg db.UpsertSubCountyIndicatorWeeklyStatusParams,
) (db.WeeklyStatus, error) {
	return r.db.UpsertSubCountyIndicatorWeeklyStatus(ctx, arg)
}

func (r *WeeklyStatusRepository) WithTx(
	ctx context.Context, fn func(q db.Querier) error,
) error {
	return r.db.ExecTx(ctx, fn)
}

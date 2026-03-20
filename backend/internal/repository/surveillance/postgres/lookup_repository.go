package postgres

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

type LookupRepository struct {
	db db.Store
}

func NewLookupRepository(db db.Store) *LookupRepository {
	return &LookupRepository{db: db}
}

func (r *LookupRepository) ResolveDiseaseByName(ctx context.Context, name string) (db.Disease, error) {
	return r.db.ResolveDiseaseByName(ctx, name)
}

func (r *LookupRepository) ResolveIndicatorByName(ctx context.Context, name string) (db.Indicator, error) {
	return r.db.ResolveIndicatorByName(ctx, name)
}

func (r *LookupRepository) ResolveWeekByYearWeek(ctx context.Context, arg db.ResolveWeekByYearWeekParams) (db.EpiWeek, error) {
	return r.db.ResolveWeekByYearWeek(ctx, arg)
}

func (r *LookupRepository) ResolveDistrictByName(ctx context.Context, name string) (db.District, error) {
	return r.db.ResolveDistrictByName(ctx, name)
}

func (r *LookupRepository) ResolveRegionByName(ctx context.Context, name string) (db.Region, error) {
	return r.db.ResolveRegionByName(ctx, name)
}

func (r *LookupRepository) ResolveSubCountyByNameAndDistrict(ctx context.Context, arg db.ResolveSubCountyByNameAndDistrictParams) (db.SubCounty, error) {
	return r.db.ResolveSubCountyByNameAndDistrict(ctx, arg)
}

func (r *LookupRepository) ResolveFacilityByExternalID(ctx context.Context, externalID string) (db.Facility, error) {

	if externalID == "" {
		return db.Facility{}, errors.New("external id is required")
	}

	param := sql.NullString{
		String: externalID,
		Valid:  true,
	}

	facility, err := r.db.ResolveFacilityByExternalID(ctx, param)
	if err != nil {
		return db.Facility{}, err
	}

	return facility, nil
}

func (r *LookupRepository) ResolveFacilityByNameAndDistrict(ctx context.Context, arg db.ResolveFacilityByNameAndDistrictParams) (db.Facility, error) {
	return r.db.ResolveFacilityByNameAndDistrict(ctx, arg)
}

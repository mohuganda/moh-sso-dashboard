package geography

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	db "github.com/moh-sso-dashboard/internal/db/sqlc"
)

var ErrDeletionRestricted = errors.New("shared geography cannot be deleted; disable its health context instead")

type registryQueries struct {
	db.Querier
	primary   *sql.DB
	replicate func(context.Context) error
}
type registryStore struct {
	*registryQueries
	base db.Store
}

// NewStore keeps all surveillance facts/reads in Observatory while routing every
// reference-data mutation, including transactional import queries, through primary.
func NewStore(remote db.Store, primary *sql.DB, replicate func(context.Context) error) db.Store {
	return &registryStore{&registryQueries{remote, primary, replicate}, remote}
}
func (s *registryStore) ExecTx(ctx context.Context, fn func(db.Querier) error) error {
	return s.base.ExecTx(ctx, func(q db.Querier) error { return fn(&registryQueries{q, s.primary, s.replicate}) })
}
func resolve(ctx context.Context, tx *sql.Tx, kind string, local uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRowContext(ctx, `SELECT canonical.external_id::uuid FROM public.health_context_aliases local
 JOIN public.health_context_aliases canonical ON canonical.context_node_id=local.context_node_id AND canonical.namespace=$1
 WHERE local.namespace=$2 AND local.external_id=$3`, "geography-"+kind, "surveillance-"+kind, local.String()).Scan(&id)
	return id, err
}
func resolveNullable(ctx context.Context, tx *sql.Tx, kind string, local *uuid.NullUUID) error {
	if !local.Valid {
		return nil
	}
	id, err := resolve(ctx, tx, kind, local.UUID)
	if err != nil {
		return err
	}
	local.UUID = id
	return nil
}
func (q *registryQueries) mutate(ctx context.Context, kind string, fn func(*sql.Tx, *db.Queries) (uuid.UUID, error)) (uuid.UUID, error) {
	tx, err := q.primary.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SET LOCAL search_path TO geography, public`); err != nil {
		return uuid.Nil, err
	}
	source, err := fn(tx, db.New(tx))
	if err != nil {
		return uuid.Nil, err
	}
	if err = tx.Commit(); err != nil {
		return uuid.Nil, err
	}
	if err = q.replicate(ctx); err != nil {
		return uuid.Nil, err
	}
	var local uuid.UUID
	err = q.primary.QueryRowContext(ctx, `SELECT replica.external_id::uuid FROM public.health_context_aliases canonical
 JOIN public.health_context_aliases replica ON replica.context_node_id=canonical.context_node_id AND replica.namespace=$1
 WHERE canonical.namespace=$2 AND canonical.external_id=$3`, "surveillance-"+kind, "geography-"+kind, source.String()).Scan(&local)
	return local, err
}
func (q *registryQueries) CreateRegion(ctx context.Context, arg db.CreateRegionParams) (db.Region, error) {
	id, err := q.mutate(ctx, "region", func(tx *sql.Tx, primary *db.Queries) (uuid.UUID, error) {

		row, err := primary.UpsertRegion(ctx, db.UpsertRegionParams(arg))
		return row.ID, err
	})
	if err != nil {
		return db.Region{}, err
	}
	return q.GetRegionByID(ctx, id)
}
func (q *registryQueries) UpsertRegion(ctx context.Context, arg db.UpsertRegionParams) (db.Region, error) {
	id, err := q.mutate(ctx, "region", func(tx *sql.Tx, primary *db.Queries) (uuid.UUID, error) {

		row, err := primary.UpsertRegion(ctx, arg)
		return row.ID, err
	})
	if err != nil {
		return db.Region{}, err
	}
	return q.GetRegionByID(ctx, id)
}
func (q *registryQueries) DeleteRegion(ctx context.Context, id uuid.UUID) error {
	return ErrDeletionRestricted
}
func (q *registryQueries) CreateDistrict(ctx context.Context, arg db.CreateDistrictParams) (db.District, error) {
	id, err := q.mutate(ctx, "district", func(tx *sql.Tx, primary *db.Queries) (uuid.UUID, error) {
		if err := resolveNullable(ctx, tx, "region", &arg.RegionID); err != nil {
			return uuid.Nil, err
		}

		row, err := primary.UpsertDistrict(ctx, db.UpsertDistrictParams(arg))
		return row.ID, err
	})
	if err != nil {
		return db.District{}, err
	}
	return q.GetDistrictByID(ctx, id)
}
func (q *registryQueries) UpsertDistrict(ctx context.Context, arg db.UpsertDistrictParams) (db.District, error) {
	id, err := q.mutate(ctx, "district", func(tx *sql.Tx, primary *db.Queries) (uuid.UUID, error) {
		if err := resolveNullable(ctx, tx, "region", &arg.RegionID); err != nil {
			return uuid.Nil, err
		}

		row, err := primary.UpsertDistrict(ctx, arg)
		return row.ID, err
	})
	if err != nil {
		return db.District{}, err
	}
	return q.GetDistrictByID(ctx, id)
}
func (q *registryQueries) DeleteDistrict(ctx context.Context, id uuid.UUID) error {
	return ErrDeletionRestricted
}
func (q *registryQueries) CreateSubCounty(ctx context.Context, arg db.CreateSubCountyParams) (db.SubCounty, error) {
	id, err := q.mutate(ctx, "sub-county", func(tx *sql.Tx, primary *db.Queries) (uuid.UUID, error) {
		parent, err := resolve(ctx, tx, "district", arg.DistrictID)
		if err != nil {
			return uuid.Nil, err
		}
		arg.DistrictID = parent

		row, err := primary.UpsertSubCounty(ctx, db.UpsertSubCountyParams(arg))
		return row.ID, err
	})
	if err != nil {
		return db.SubCounty{}, err
	}
	return q.GetSubCountyByID(ctx, id)
}
func (q *registryQueries) UpsertSubCounty(ctx context.Context, arg db.UpsertSubCountyParams) (db.SubCounty, error) {
	id, err := q.mutate(ctx, "sub-county", func(tx *sql.Tx, primary *db.Queries) (uuid.UUID, error) {
		parent, err := resolve(ctx, tx, "district", arg.DistrictID)
		if err != nil {
			return uuid.Nil, err
		}
		arg.DistrictID = parent

		row, err := primary.UpsertSubCounty(ctx, arg)
		return row.ID, err
	})
	if err != nil {
		return db.SubCounty{}, err
	}
	return q.GetSubCountyByID(ctx, id)
}
func (q *registryQueries) DeleteSubCounty(ctx context.Context, id uuid.UUID) error {
	return ErrDeletionRestricted
}
func (q *registryQueries) CreateFacility(ctx context.Context, arg db.CreateFacilityParams) (db.Facility, error) {
	id, err := q.mutate(ctx, "facility", func(tx *sql.Tx, primary *db.Queries) (uuid.UUID, error) {
		if err := resolveNullable(ctx, tx, "region", &arg.RegionID); err != nil {
			return uuid.Nil, err
		}
		if err := resolveNullable(ctx, tx, "district", &arg.DistrictID); err != nil {
			return uuid.Nil, err
		}
		if err := resolveNullable(ctx, tx, "sub-county", &arg.SubCountyID); err != nil {
			return uuid.Nil, err
		}

		var row db.Facility
		var err error
		if arg.ExternalID.Valid && arg.ExternalID.String != "" {
			row, err = primary.UpsertFacilityByExternalID(ctx, db.UpsertFacilityByExternalIDParams(arg))
		} else {
			row, err = primary.UpsertFacilityByNameDistrict(ctx, db.UpsertFacilityByNameDistrictParams(arg))
		}
		return row.ID, err
	})
	if err != nil {
		return db.Facility{}, err
	}
	return q.GetFacilityByID(ctx, id)
}
func (q *registryQueries) UpsertFacilityByExternalID(ctx context.Context, arg db.UpsertFacilityByExternalIDParams) (db.Facility, error) {
	id, err := q.mutate(ctx, "facility", func(tx *sql.Tx, primary *db.Queries) (uuid.UUID, error) {
		if err := resolveNullable(ctx, tx, "region", &arg.RegionID); err != nil {
			return uuid.Nil, err
		}
		if err := resolveNullable(ctx, tx, "district", &arg.DistrictID); err != nil {
			return uuid.Nil, err
		}
		if err := resolveNullable(ctx, tx, "sub-county", &arg.SubCountyID); err != nil {
			return uuid.Nil, err
		}

		row, err := primary.UpsertFacilityByExternalID(ctx, arg)
		return row.ID, err
	})
	if err != nil {
		return db.Facility{}, err
	}
	return q.GetFacilityByID(ctx, id)
}
func (q *registryQueries) UpsertFacilityByNameDistrict(ctx context.Context, arg db.UpsertFacilityByNameDistrictParams) (db.Facility, error) {
	id, err := q.mutate(ctx, "facility", func(tx *sql.Tx, primary *db.Queries) (uuid.UUID, error) {
		if err := resolveNullable(ctx, tx, "region", &arg.RegionID); err != nil {
			return uuid.Nil, err
		}
		if err := resolveNullable(ctx, tx, "district", &arg.DistrictID); err != nil {
			return uuid.Nil, err
		}
		if err := resolveNullable(ctx, tx, "sub-county", &arg.SubCountyID); err != nil {
			return uuid.Nil, err
		}

		row, err := primary.UpsertFacilityByNameDistrict(ctx, arg)
		return row.ID, err
	})
	if err != nil {
		return db.Facility{}, err
	}
	return q.GetFacilityByID(ctx, id)
}
func (q *registryQueries) DeleteFacility(ctx context.Context, id uuid.UUID) error {
	return ErrDeletionRestricted
}

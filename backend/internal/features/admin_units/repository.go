package admin_units

import (
	"context"
	"database/sql"
)

type Repository interface {
	ListOrgUnits(ctx context.Context) ([]OrgUnitFull, error)
	ListFacilities(ctx context.Context) ([]Facility, error)
	ListDistricts(ctx context.Context) ([]OrgUnitSimple, error)
	ListSubCounties(ctx context.Context, subCounty *string) ([]OrgUnitSimple, error)
	ListLocalGovt(ctx context.Context, district *string) ([]OrgUnitSimple, error)
	ListDistrictsByRegion(ctx context.Context, region *string) ([]OrgUnitSimple, error)
	ListRegions(ctx context.Context) ([]OrgUnitSimple, error)
	ListNational(ctx context.Context) ([]OrgUnitSimple, error)
	ListHierarchy(ctx context.Context) ([]OrgUnit, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) ListOrgUnits(ctx context.Context) ([]OrgUnitFull, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			dim_org_hierarchy_key, org_unit_id, org_unit_name, "level",
			country_uid, region_uid, region, district_uid, district,
			sub_county_uid, sub_county, facility_uid, facility_name,
			division_uid, division, row_version, is_current
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		  AND is_current = true
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]OrgUnitFull, 0)
	for rows.Next() {
		var item OrgUnitFull
		if err := rows.Scan(
			&item.DimOrgHierarchyKey,
			&item.OrgUnitID,
			&item.OrgUnitName,
			&item.Level,
			&item.CountryUID,
			&item.RegionUID,
			&item.Region,
			&item.DistrictUID,
			&item.District,
			&item.SubCountyUID,
			&item.SubCounty,
			&item.FacilityUID,
			&item.FacilityName,
			&item.DivisionUID,
			&item.Division,
			&item.RowVersion,
			&item.IsCurrent,
		); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

func (r *postgresRepository) ListFacilities(ctx context.Context) ([]Facility, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			dim_org_hierarchy_key, org_unit_id, org_unit_name, "level",
			country_uid, region_uid, region, district_uid, district,
			sub_county_uid, sub_county, division_uid, division,
			facility_uid, facility_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		  AND is_current = true
		  AND "level" = '6'
		ORDER BY region, district, sub_county, facility_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]Facility, 0)
	for rows.Next() {
		var item Facility
		if err := rows.Scan(
			&item.DimOrgHierarchyKey,
			&item.OrgUnitID,
			&item.OrgUnitName,
			&item.Level,
			&item.CountryUID,
			&item.RegionUID,
			&item.Region,
			&item.DistrictUID,
			&item.District,
			&item.SubCountyUID,
			&item.SubCounty,
			&item.DivisionUID,
			&item.Division,
			&item.FacilityUID,
			&item.FacilityName,
		); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

func (r *postgresRepository) ListDistricts(ctx context.Context) ([]OrgUnitSimple, error) {
	return r.listSimple(ctx, `
		SELECT dim_org_hierarchy_key, org_unit_id, org_unit_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		  AND is_current = true
		  AND "level" = '3'
		ORDER BY region, district, sub_county, facility_name
	`)
}

func (r *postgresRepository) ListSubCounties(ctx context.Context, subCounty *string) ([]OrgUnitSimple, error) {
	query := `
		SELECT dim_org_hierarchy_key, org_unit_id, org_unit_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		  AND is_current = true
		  AND "level" = '5'
	`
	if subCounty != nil && *subCounty != "" {
		return r.listSimple(ctx, query+` AND sub_county = $1 ORDER BY region, district, sub_county, facility_name`, *subCounty)
	}
	return r.listSimple(ctx, query+` ORDER BY region, district, sub_county, facility_name`)
}

func (r *postgresRepository) ListLocalGovt(ctx context.Context, district *string) ([]OrgUnitSimple, error) {
	query := `
		SELECT dim_org_hierarchy_key, org_unit_id, org_unit_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		  AND is_current = true
		  AND "level" = '4'
	`
	if district != nil && *district != "" {
		return r.listSimple(ctx, query+` AND district = $1 ORDER BY region, district, sub_county`, *district)
	}
	return r.listSimple(ctx, query+` ORDER BY region, district, sub_county`)
}

func (r *postgresRepository) ListDistrictsByRegion(ctx context.Context, region *string) ([]OrgUnitSimple, error) {
	query := `
		SELECT dim_org_hierarchy_key, org_unit_id, org_unit_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		  AND is_current = true
		  AND "level" = '3'
	`
	if region != nil && *region != "" {
		return r.listSimple(ctx, query+` AND region = $1 ORDER BY region, district`, *region)
	}
	return r.listSimple(ctx, query+` ORDER BY region, district`)
}

func (r *postgresRepository) ListRegions(ctx context.Context) ([]OrgUnitSimple, error) {
	return r.listSimple(ctx, `
		SELECT dim_org_hierarchy_key, org_unit_id, org_unit_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		  AND is_current = true
		  AND "level" = '2'
		ORDER BY region, district, sub_county, facility_name
	`)
}

func (r *postgresRepository) ListNational(ctx context.Context) ([]OrgUnitSimple, error) {
	return r.listSimple(ctx, `
		SELECT dim_org_hierarchy_key, org_unit_id, org_unit_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		  AND is_current = true
		  AND "level" = '1'
		ORDER BY region, district, sub_county, facility_name
	`)
}

func (r *postgresRepository) ListHierarchy(ctx context.Context) ([]OrgUnit, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			dim_org_hierarchy_key, org_unit_id, org_unit_name, "level",
			country_uid, region_uid, region, district_uid, district,
			sub_county_uid, sub_county, division_uid, division,
			facility_uid, facility_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		  AND is_current = true
		ORDER BY "level", region, district, sub_county, division, facility_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]OrgUnit, 0)
	for rows.Next() {
		var item OrgUnit
		if err := rows.Scan(
			&item.DimOrgHierarchyKey,
			&item.OrgUnitID,
			&item.OrgUnitName,
			&item.Level,
			&item.CountryUID,
			&item.RegionUID,
			&item.Region,
			&item.DistrictUID,
			&item.District,
			&item.SubCountyUID,
			&item.SubCounty,
			&item.DivisionUID,
			&item.Division,
			&item.FacilityUID,
			&item.FacilityName,
		); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

func (r *postgresRepository) listSimple(ctx context.Context, query string, args ...any) ([]OrgUnitSimple, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]OrgUnitSimple, 0)
	for rows.Next() {
		var item OrgUnitSimple
		if err := rows.Scan(
			&item.DimOrgHierarchyKey,
			&item.OrgUnitID,
			&item.OrgUnitName,
		); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

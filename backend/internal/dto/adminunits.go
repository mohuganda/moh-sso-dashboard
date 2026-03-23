package dto

// OrgUnit represents an organizational unit from the database
type OrgUnit struct {
	DimOrgHierarchyKey int64   `json:"dim_org_hierarchy_key"`
	OrgUnitID          *string `json:"org_unit_id"`
	OrgUnitName        *string `json:"org_unit_name"`
	Level              string  `json:"level"`
	CountryUID         *string `json:"country_uid"`
	RegionUID          *string `json:"region_uid"`
	Region             *string `json:"region"`
	DistrictUID        *string `json:"district_uid"`
	District           *string `json:"district"`
	SubCountyUID       *string `json:"sub_county_uid"`
	SubCounty          *string `json:"sub_county"`
	DivisionUID        *string `json:"division_uid"`
	Division           *string `json:"division"`
	FacilityUID        *string `json:"facility_uid"`
	FacilityName       *string `json:"facility_name"`
}

// OrgUnitFull represents a full organizational unit with row_version and is_current
type OrgUnitFull struct {
	DimOrgHierarchyKey int64   `json:"dim_org_hierarchy_key"`
	OrgUnitID          *string `json:"org_unit_id"`
	OrgUnitName        *string `json:"org_unit_name"`
	Level              string  `json:"level"`
	CountryUID         *string `json:"country_uid"`
	RegionUID          *string `json:"region_uid"`
	Region             *string `json:"region"`
	DistrictUID        *string `json:"district_uid"`
	District           *string `json:"district"`
	SubCountyUID       *string `json:"sub_county_uid"`
	SubCounty          *string `json:"sub_county"`
	FacilityUID        *string `json:"facility_uid"`
	FacilityName       *string `json:"facility_name"`
	DivisionUID        *string `json:"division_uid"`
	Division           *string `json:"division"`
	RowVersion         int64   `json:"row_version"`
	IsCurrent          bool    `json:"is_current"`
}

// Facility represents a facility (level 6) organizational unit
type Facility struct {
	DimOrgHierarchyKey int64   `json:"dim_org_hierarchy_key"`
	OrgUnitID          *string `json:"org_unit_id"`
	OrgUnitName        *string `json:"org_unit_name"`
	Level              string  `json:"level"`
	CountryUID         *string `json:"country_uid"`
	RegionUID          *string `json:"region_uid"`
	Region             *string `json:"region"`
	DistrictUID        *string `json:"district_uid"`
	District           *string `json:"district"`
	SubCountyUID       *string `json:"sub_county_uid"`
	SubCounty          *string `json:"sub_county"`
	DivisionUID        *string `json:"division_uid"`
	Division           *string `json:"division"`
	FacilityUID        *string `json:"facility_uid"`
	FacilityName       *string `json:"facility_name"`
}

// OrgUnitSimple represents a simple organizational unit with only basic fields
type OrgUnitSimple struct {
	DimOrgHierarchyKey int64   `json:"dim_org_hierarchy_key"`
	OrgUnitID          *string `json:"org_unit_id"`
	OrgUnitName        *string `json:"org_unit_name"`
}

// TreeNode represents a node in the organizational hierarchy tree
type TreeNode struct {
	ID       int64      `json:"id"`
	UID      *string    `json:"uid"`
	Name     *string    `json:"name"`
	Level    string     `json:"level"`
	Type     string     `json:"type"`
	Division *string    `json:"division,omitempty"`
	Children []TreeNode `json:"children"`
}

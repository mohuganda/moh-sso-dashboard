package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/dto"
)

// GetOrgUnits gets all organizational units
func GetOrgUnits(c *gin.Context) {
	query := `
		SELECT 
			dim_org_hierarchy_key, 
			org_unit_id, org_unit_name, 
			"level", 
			country_uid, 
			region_uid, 
			region, 
			district_uid, 
			district, 
			sub_county_uid, 
			sub_county, 
			facility_uid, 
			facility_name,
			division_uid, 
			division,
			row_version, 
			is_current
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		AND is_current = true
	`

	rows, err := config.DWH.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.OrgUnitFull
	for rows.Next() {
		var ou dto.OrgUnitFull
		err := rows.Scan(&ou.DimOrgHierarchyKey, &ou.OrgUnitID, &ou.OrgUnitName, &ou.Level, &ou.CountryUID, &ou.RegionUID, &ou.Region, &ou.DistrictUID, &ou.District, &ou.SubCountyUID, &ou.SubCounty, &ou.FacilityUID, &ou.FacilityName, &ou.DivisionUID, &ou.Division, &ou.RowVersion, &ou.IsCurrent)
		if err != nil {
			continue
		}
		results = append(results, ou)
	}

	c.JSON(http.StatusOK, results)
}

// GetFacilities gets all facilities (level 6)
func GetFacilities(c *gin.Context) {
	query := `
		SELECT 
			dim_org_hierarchy_key, 
			org_unit_id, 
			org_unit_name, 
			"level", 
			country_uid, 
			region_uid, 
			region, 
			district_uid, 
			district, 
			sub_county_uid, 
			sub_county, 
			division_uid, 
			division,
			facility_uid, 
			facility_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		AND is_current = true
		AND "level" = '6'
		ORDER BY region, district, sub_county, facility_name
	`

	rows, err := configs.DWH.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.Facility
	for rows.Next() {
		var f dto.Facility
		err := rows.Scan(&f.DimOrgHierarchyKey, &f.OrgUnitID, &f.OrgUnitName, &f.Level, &f.CountryUID, &f.RegionUID, &f.Region, &f.DistrictUID, &f.District, &f.SubCountyUID, &f.SubCounty, &f.DivisionUID, &f.Division, &f.FacilityUID, &f.FacilityName)
		if err != nil {
			continue
		}
		results = append(results, f)
	}

	c.JSON(http.StatusOK, results)
}

// GetDistricts gets all districts (level 3)
func GetDistricts(c *gin.Context) {
	query := `
		SELECT 
			dim_org_hierarchy_key, 
			org_unit_id, 
			org_unit_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		AND is_current = true
		AND "level" = '3'
		ORDER BY region, district, sub_county, facility_name
	`

	rows, err := configs.DWH.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.OrgUnitSimple
	for rows.Next() {
		var ou dto.OrgUnitSimple
		err := rows.Scan(&ou.DimOrgHierarchyKey, &ou.OrgUnitID, &ou.OrgUnitName)
		if err != nil {
			continue
		}
		results = append(results, ou)
	}

	c.JSON(http.StatusOK, results)
}

// GetSubCounties gets subcounties (level 5), optionally filtered by sub_county
func GetSubCounties(c *gin.Context) {
	type Request struct {
		SubCounty *string `json:"sub_county"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		// If no JSON body, proceed without filter
		req = Request{}
	}

	query := `
		SELECT 
			dim_org_hierarchy_key, 
			org_unit_id, 
			org_unit_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		AND is_current = true
		AND "level" = '5'
	`

	var rows *sql.Rows
	var err error

	if req.SubCounty != nil && *req.SubCounty != "" {
		query += ` AND sub_county = $1 ORDER BY region, district, sub_county, facility_name`
		rows, err = configs.DWH.Query(query, *req.SubCounty)
	} else {
		query += ` ORDER BY region, district, sub_county, facility_name`
		rows, err = configs.DWH.Query(query)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.OrgUnitSimple
	for rows.Next() {
		var ou dto.OrgUnitSimple
		err := rows.Scan(&ou.DimOrgHierarchyKey, &ou.OrgUnitID, &ou.OrgUnitName)
		if err != nil {
			continue
		}
		results = append(results, ou)
	}

	c.JSON(http.StatusOK, results)
}

// GetLocalGovt gets local government units (level 4), optionally filtered by district
func GetLocalGovt(c *gin.Context) {
	type Request struct {
		District *string `json:"district"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		// If no JSON body, proceed without filter
		req = Request{}
	}

	query := `
		SELECT 
			dim_org_hierarchy_key, 
			org_unit_id, 
			org_unit_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		AND is_current = true
		AND "level" = '4'
	`

	var rows *sql.Rows
	var err error

	if req.District != nil && *req.District != "" {
		query += ` AND district = $1 ORDER BY region, district, sub_county, facility_name`
		rows, err = configs.DWH.Query(query, *req.District)
	} else {
		query += ` ORDER BY region, district, sub_county, facility_name`
		rows, err = configs.DWH.Query(query)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.OrgUnitSimple
	for rows.Next() {
		var ou dto.OrgUnitSimple
		err := rows.Scan(&ou.DimOrgHierarchyKey, &ou.OrgUnitID, &ou.OrgUnitName)
		if err != nil {
			continue
		}
		results = append(results, ou)
	}

	c.JSON(http.StatusOK, results)
}

// GetDistrictsByRegion gets districts (level 3) filtered by region
func GetDistrictsByRegion(c *gin.Context) {
	type Request struct {
		Region *string `json:"region"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		// If no JSON body, proceed without filter
		req = Request{}
	}

	query := `
		SELECT 
			dim_org_hierarchy_key, 
			org_unit_id, 
			org_unit_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		AND is_current = true
		AND "level" = '3'
	`

	var rows *sql.Rows
	var err error

	if req.Region != nil && *req.Region != "" {
		query += ` AND region = $1 ORDER BY region, district, sub_county, facility_name`
		rows, err = configs.DWH.Query(query, *req.Region)
	} else {
		query += ` ORDER BY region, district, sub_county, facility_name`
		rows, err = configs.DWH.Query(query)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.OrgUnitSimple
	for rows.Next() {
		var ou dto.OrgUnitSimple
		err := rows.Scan(&ou.DimOrgHierarchyKey, &ou.OrgUnitID, &ou.OrgUnitName)
		if err != nil {
			continue
		}
		results = append(results, ou)
	}

	c.JSON(http.StatusOK, results)
}

// GetRegions gets all regions (level 2)
func GetRegions(c *gin.Context) {
	query := `
		SELECT 
			dim_org_hierarchy_key, 
			org_unit_id, 
			org_unit_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		AND is_current = true
		AND "level" = '2'
		ORDER BY region, district, sub_county, facility_name
	`

	rows, err := configs.DWH.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.OrgUnitSimple
	for rows.Next() {
		var ou dto.OrgUnitSimple
		err := rows.Scan(&ou.DimOrgHierarchyKey, &ou.OrgUnitID, &ou.OrgUnitName)
		if err != nil {
			continue
		}
		results = append(results, ou)
	}

	c.JSON(http.StatusOK, results)
}

// GetNational gets national level (level 1)
func GetNational(c *gin.Context) {
	query := `
		SELECT 
			dim_org_hierarchy_key, 
			org_unit_id, 
			org_unit_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		AND is_current = true
		AND "level" = '1'
		ORDER BY region, district, sub_county, facility_name
	`

	rows, err := configs.DWH.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.OrgUnitSimple
	for rows.Next() {
		var ou dto.OrgUnitSimple
		err := rows.Scan(&ou.DimOrgHierarchyKey, &ou.OrgUnitID, &ou.OrgUnitName)
		if err != nil {
			continue
		}
		results = append(results, ou)
	}

	c.JSON(http.StatusOK, results)
}

// GetHierarchy gets the full organizational hierarchy as a tree structure
func GetHierarchy(c *gin.Context) {
	query := `
		SELECT 
			dim_org_hierarchy_key,
			org_unit_id,
			org_unit_name,
			"level",
			country_uid,
			region_uid,
			region,
			district_uid,
			district,
			sub_county_uid,
			sub_county,
			division_uid,
			division,
			facility_uid,
			facility_name
		FROM dwh.dim_org_hierarchy
		WHERE dim_org_hierarchy_key <> -1
		AND is_current = true
		ORDER BY "level", region, district, sub_county, division, facility_name
	`

	rows, err := configs.DWH.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Read all rows into a slice
	var data []dto.OrgUnit
	for rows.Next() {
		var ou dto.OrgUnit
		err := rows.Scan(
			&ou.DimOrgHierarchyKey, &ou.OrgUnitID, &ou.OrgUnitName, &ou.Level,
			&ou.CountryUID, &ou.RegionUID, &ou.Region,
			&ou.DistrictUID, &ou.District, &ou.SubCountyUID, &ou.SubCounty,
			&ou.DivisionUID, &ou.Division, &ou.FacilityUID, &ou.FacilityName,
		)
		if err != nil {
			continue
		}
		data = append(data, ou)
	}

	// Build hierarchical tree structure
	buildTree := func(data []dto.OrgUnit) []dto.TreeNode {
		tree := []dto.TreeNode{}

		// Group by level
		grouped := make(map[string][]dto.OrgUnit)
		for _, item := range data {
			level := item.Level
			grouped[level] = append(grouped[level], item)
		}

		// Build tree starting from level 1 (National)
		if level1, ok := grouped["1"]; ok {
			for _, national := range level1 {
				nationalName := national.OrgUnitName
				if nationalName == nil {
					name := "National"
					nationalName = &name
				}
				nationalNode := dto.TreeNode{
					ID:       national.DimOrgHierarchyKey,
					UID:      national.CountryUID,
					Name:     nationalName,
					Level:    "1",
					Type:     "national",
					Children: []dto.TreeNode{},
				}

				// Add regions (level 2)
				if level2, ok := grouped["2"]; ok {
					for _, reg := range level2 {
						regionName := reg.Region
						if regionName == nil {
							continue
						}
						regionNode := dto.TreeNode{
							ID:       reg.DimOrgHierarchyKey,
							UID:      reg.RegionUID,
							Name:     regionName,
							Level:    "2",
							Type:     "region",
							Children: []dto.TreeNode{},
						}

						// Add districts (level 3)
						if level3, ok := grouped["3"]; ok {
							for _, dist := range level3 {
								if dist.Region == nil || *dist.Region != *reg.Region {
									continue
								}
								districtName := dist.District
								if districtName == nil {
									continue
								}
								districtNode := dto.TreeNode{
									ID:       dist.DimOrgHierarchyKey,
									UID:      dist.DistrictUID,
									Name:     districtName,
									Level:    "3",
									Type:     "district",
									Children: []dto.TreeNode{},
								}

								// Add local govt (level 4)
								if level4, ok := grouped["4"]; ok {
									for _, lg := range level4 {
										if lg.District == nil || *lg.District != *dist.District {
											continue
										}
										lgName := lg.OrgUnitName
										if lgName == nil {
											continue
										}
										districtNode.Children = append(districtNode.Children, dto.TreeNode{
											ID:    lg.DimOrgHierarchyKey,
											UID:   lg.OrgUnitID,
											Name:  lgName,
											Level: "4",
											Type:  "localgovt",
										})
									}
								}

								// Add subcounties (level 5)
								if level5, ok := grouped["5"]; ok {
									for _, sc := range level5 {
										if sc.District == nil || *sc.District != *dist.District {
											continue
										}
										scName := sc.SubCounty
										if scName == nil {
											continue
										}
										subcountyNode := dto.TreeNode{
											ID:       sc.DimOrgHierarchyKey,
											UID:      sc.SubCountyUID,
											Name:     scName,
											Level:    "5",
											Type:     "subcounty",
											Children: []dto.TreeNode{},
										}

										// Add facilities (level 6)
										if level6, ok := grouped["6"]; ok {
											for _, fac := range level6 {
												if fac.SubCounty == nil || *fac.SubCounty != *sc.SubCounty {
													continue
												}
												if fac.District == nil || *fac.District != *dist.District {
													continue
												}
												facName := fac.FacilityName
												if facName == nil {
													continue
												}
												subcountyNode.Children = append(subcountyNode.Children, dto.TreeNode{
													ID:       fac.DimOrgHierarchyKey,
													UID:      fac.FacilityUID,
													Name:     facName,
													Level:    "6",
													Type:     "facility",
													Division: fac.Division,
												})
											}
										}

										districtNode.Children = append(districtNode.Children, subcountyNode)
									}
								}

								regionNode.Children = append(regionNode.Children, districtNode)
							}
						}

						nationalNode.Children = append(nationalNode.Children, regionNode)
					}
				}

				tree = append(tree, nationalNode)
			}
		}

		return tree
	}

	tree := buildTree(data)
	c.JSON(http.StatusOK, tree)
}

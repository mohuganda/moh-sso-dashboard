package admin_units

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/http/response"
)

type Handler struct {
	config *config.Config
	db     *sql.DB
}

func NewHandler(
	config *config.Config,
	db *sql.DB,
) *Handler {
	return &Handler{
		config: config,
		db:     db,
	}
}

func (h *Handler) GetOrgUnits(c *gin.Context) {
	ctx := c.Request.Context()

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

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_ORG_UNITS_FAILED", "failed to list organizational units")
		return
	}
	defer rows.Close()

	results := make([]OrgUnitFull, 0)
	for rows.Next() {
		var ou OrgUnitFull

		err := rows.Scan(
			&ou.DimOrgHierarchyKey,
			&ou.OrgUnitID,
			&ou.OrgUnitName,
			&ou.Level,
			&ou.CountryUID,
			&ou.RegionUID,
			&ou.Region,
			&ou.DistrictUID,
			&ou.District,
			&ou.SubCountyUID,
			&ou.SubCounty,
			&ou.FacilityUID,
			&ou.FacilityName,
			&ou.DivisionUID,
			&ou.Division,
			&ou.RowVersion,
			&ou.IsCurrent,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "LIST_ORG_UNITS_FAILED", "failed to list organizational units")
			return
		}

		results = append(results, ou)
	}

	if err := rows.Err(); err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_ORG_UNITS_FAILED", "failed to list organizational units")
		return
	}

	response.OK(c, http.StatusOK, results)
}

// GetFacilities gets all facilities (level 6)
func (h *Handler) GetFacilities(c *gin.Context) {
	ctx := c.Request.Context()

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

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_FACILITIES_FAILED", "failed to list facilities")
		return
	}
	defer rows.Close()

	results := make([]Facility, 0)
	for rows.Next() {
		var f Facility

		err := rows.Scan(
			&f.DimOrgHierarchyKey,
			&f.OrgUnitID,
			&f.OrgUnitName,
			&f.Level,
			&f.CountryUID,
			&f.RegionUID,
			&f.Region,
			&f.DistrictUID,
			&f.District,
			&f.SubCountyUID,
			&f.SubCounty,
			&f.DivisionUID,
			&f.Division,
			&f.FacilityUID,
			&f.FacilityName,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "LIST_FACILITIES_FAILED", "failed to list facilities")
			return
		}

		results = append(results, f)
	}

	if err := rows.Err(); err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_FACILITIES_FAILED", "failed to list facilities")
		return
	}

	response.OK(c, http.StatusOK, results)
}

// GetDistricts gets all districts (level 3)
func (h *Handler) GetDistricts(c *gin.Context) {
	ctx := c.Request.Context()

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

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_DISTRICTS_FAILED", "failed to list districts")
		return
	}
	defer rows.Close()

	results := make([]OrgUnitSimple, 0)
	for rows.Next() {
		var ou OrgUnitSimple

		if err := rows.Scan(
			&ou.DimOrgHierarchyKey,
			&ou.OrgUnitID,
			&ou.OrgUnitName,
		); err != nil {
			response.Fail(c, http.StatusInternalServerError, "LIST_DISTRICTS_FAILED", "failed to list districts")
			return
		}

		results = append(results, ou)
	}

	if err := rows.Err(); err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_DISTRICTS_FAILED", "failed to list districts")
		return
	}

	response.OK(c, http.StatusOK, results)
}

// GetSubCounties gets subcounties (level 5), optionally filtered by sub_county
func (h *Handler) GetSubCounties(c *gin.Context) {
	ctx := c.Request.Context()

	type Request struct {
		SubCounty *string `json:"sub_county"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		// No JSON body or invalid JSON: continue without filter
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

	var (
		rows *sql.Rows
		err  error
	)

	if req.SubCounty != nil && *req.SubCounty != "" {
		query += ` AND sub_county = $1 ORDER BY region, district, sub_county, facility_name`
		rows, err = h.db.QueryContext(ctx, query, *req.SubCounty)
	} else {
		query += ` ORDER BY region, district, sub_county, facility_name`
		rows, err = h.db.QueryContext(ctx, query)
	}

	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_SUBCOUNTIES_FAILED", "failed to list subcounties")
		return
	}
	defer rows.Close()

	results := make([]OrgUnitSimple, 0)
	for rows.Next() {
		var ou OrgUnitSimple

		if err := rows.Scan(
			&ou.DimOrgHierarchyKey,
			&ou.OrgUnitID,
			&ou.OrgUnitName,
		); err != nil {
			response.Fail(c, http.StatusInternalServerError, "LIST_SUBCOUNTIES_FAILED", "failed to list subcounties")
			return
		}

		results = append(results, ou)
	}

	if err := rows.Err(); err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_SUBCOUNTIES_FAILED", "failed to list subcounties")
		return
	}

	response.OK(c, http.StatusOK, results)
}

// GetLocalGovt gets local government units (level 4), optionally filtered by district
func (h *Handler) GetLocalGovt(c *gin.Context) {
	ctx := c.Request.Context()

	type Request struct {
		District *string `json:"district"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
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

	var (
		rows *sql.Rows
		err  error
	)

	if req.District != nil && *req.District != "" {
		query += ` AND district = $1 ORDER BY region, district, sub_county`
		rows, err = h.db.QueryContext(ctx, query, *req.District)
	} else {
		query += ` ORDER BY region, district, sub_county`
		rows, err = h.db.QueryContext(ctx, query)
	}

	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_LOCAL_GOVT_FAILED", "failed to list local governments")
		return
	}
	defer rows.Close()

	results := make([]OrgUnitSimple, 0)
	for rows.Next() {
		var ou OrgUnitSimple

		if err := rows.Scan(
			&ou.DimOrgHierarchyKey,
			&ou.OrgUnitID,
			&ou.OrgUnitName,
		); err != nil {
			response.Fail(c, http.StatusInternalServerError, "LIST_LOCAL_GOVT_FAILED", "failed to list local governments")
			return
		}

		results = append(results, ou)
	}

	if err := rows.Err(); err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_LOCAL_GOVT_FAILED", "failed to list local governments")
		return
	}

	response.OK(c, http.StatusOK, results)
}

// GetDistrictsByRegion gets districts (level 3) filtered by region
func (h *Handler) GetDistrictsByRegion(c *gin.Context) {
	ctx := c.Request.Context()

	type Request struct {
		Region *string `json:"region"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
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

	var (
		rows *sql.Rows
		err  error
	)

	if req.Region != nil && *req.Region != "" {
		query += ` AND region = $1 ORDER BY region, district`
		rows, err = h.db.QueryContext(ctx, query, *req.Region)
	} else {
		query += ` ORDER BY region, district`
		rows, err = h.db.QueryContext(ctx, query)
	}

	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_REGIONS_FAILED", "failed to list regions")
		return
	}
	defer rows.Close()

	results := make([]OrgUnitSimple, 0)
	for rows.Next() {
		var ou OrgUnitSimple

		if err := rows.Scan(
			&ou.DimOrgHierarchyKey,
			&ou.OrgUnitID,
			&ou.OrgUnitName,
		); err != nil {
			response.Fail(c, http.StatusInternalServerError, "LIST_REGIONS_FAILED", "failed to list regions")
			return
		}

		results = append(results, ou)
	}

	if err := rows.Err(); err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_REGIONS_FAILED", "failed to list regions")
		return
	}

	response.OK(c, http.StatusOK, results)
}

// GetRegions gets all regions (level 2)
func (h *Handler) GetRegions(c *gin.Context) {
	ctx := c.Request.Context()

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

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_NATIONAL_FAILED", "failed to list national hierarchy")
		return
	}
	defer rows.Close()

	results := make([]OrgUnitSimple, 0)
	for rows.Next() {
		var ou OrgUnitSimple

		if err := rows.Scan(
			&ou.DimOrgHierarchyKey,
			&ou.OrgUnitID,
			&ou.OrgUnitName,
		); err != nil {
			response.Fail(c, http.StatusInternalServerError, "LIST_NATIONAL_FAILED", "failed to list national hierarchy")
			return
		}

		results = append(results, ou)
	}

	if err := rows.Err(); err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_NATIONAL_FAILED", "failed to list national hierarchy")
		return
	}

	response.OK(c, http.StatusOK, results)
}

// GetNational gets national level (level 1)
func (h *Handler) GetNational(c *gin.Context) {
	ctx := c.Request.Context()

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

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_HIERARCHY_FAILED", "failed to list hierarchy")
		return
	}
	defer rows.Close()

	results := make([]OrgUnitSimple, 0)
	for rows.Next() {
		var ou OrgUnitSimple

		if err := rows.Scan(
			&ou.DimOrgHierarchyKey,
			&ou.OrgUnitID,
			&ou.OrgUnitName,
		); err != nil {
			response.Fail(c, http.StatusInternalServerError, "LIST_HIERARCHY_FAILED", "failed to list hierarchy")
			return
		}

		results = append(results, ou)
	}

	if err := rows.Err(); err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_HIERARCHY_FAILED", "failed to list hierarchy")
		return
	}

	response.OK(c, http.StatusOK, results)
}

// GetHierarchy gets the full organizational hierarchy as a tree structure
func (h *Handler) GetHierarchy(c *gin.Context) {
	ctx := c.Request.Context()

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

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_HIERARCHY_FAILED", "failed to list hierarchy")
		return
	}
	defer rows.Close()

	var data []OrgUnit
	for rows.Next() {
		var ou OrgUnit

		if err := rows.Scan(
			&ou.DimOrgHierarchyKey,
			&ou.OrgUnitID,
			&ou.OrgUnitName,
			&ou.Level,
			&ou.CountryUID,
			&ou.RegionUID,
			&ou.Region,
			&ou.DistrictUID,
			&ou.District,
			&ou.SubCountyUID,
			&ou.SubCounty,
			&ou.DivisionUID,
			&ou.Division,
			&ou.FacilityUID,
			&ou.FacilityName,
		); err != nil {
			response.Fail(c, http.StatusInternalServerError, "LIST_HIERARCHY_FAILED", "failed to list hierarchy")
			return
		}

		data = append(data, ou)
	}

	if err := rows.Err(); err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_HIERARCHY_FAILED", "failed to list hierarchy")
		return
	}

	tree := buildHierarchyTree(data)
	response.OK(c, http.StatusOK, tree)
}

func buildHierarchyTree(data []OrgUnit) []TreeNode {
	grouped := make(map[string][]OrgUnit, 6)
	for _, item := range data {
		grouped[item.Level] = append(grouped[item.Level], item)
	}

	level1 := grouped["1"]
	if len(level1) == 0 {
		return []TreeNode{}
	}

	nationals := make([]*hierarchyNode, 0, len(level1))
	nationalByUID := make(map[string]*hierarchyNode)
	seenNational := make(map[string]struct{})
	for _, national := range level1 {
		nationalName := national.OrgUnitName
		if nationalName == nil {
			name := "National"
			nationalName = &name
		}

		key := dedupeKey(national.CountryUID, nationalName, national.DimOrgHierarchyKey)
		if _, exists := seenNational[key]; exists {
			continue
		}
		seenNational[key] = struct{}{}

		node := &hierarchyNode{
			ID:    national.DimOrgHierarchyKey,
			UID:   national.CountryUID,
			Name:  nationalName,
			Level: "1",
			Type:  "national",
		}
		nationals = append(nationals, node)
		if national.CountryUID != nil && *national.CountryUID != "" {
			nationalByUID[*national.CountryUID] = node
		}
	}

	if len(nationals) == 0 {
		return []TreeNode{}
	}
	defaultNational := nationals[0]

	regionByKey := make(map[string]*hierarchyNode)
	regionKeyByUID := make(map[string]string)
	regionKeyByName := make(map[string]string)
	for _, reg := range grouped["2"] {
		if reg.Region == nil {
			continue
		}

		regionKey := dedupeKey(reg.RegionUID, reg.Region, reg.DimOrgHierarchyKey)
		if _, exists := regionByKey[regionKey]; exists {
			continue
		}

		node := &hierarchyNode{
			ID:    reg.DimOrgHierarchyKey,
			UID:   reg.RegionUID,
			Name:  reg.Region,
			Level: "2",
			Type:  "region",
		}

		parent := defaultNational
		if reg.CountryUID != nil && *reg.CountryUID != "" {
			if n, ok := nationalByUID[*reg.CountryUID]; ok {
				parent = n
			}
		}
		parent.Children = append(parent.Children, node)

		regionByKey[regionKey] = node
		if reg.RegionUID != nil && *reg.RegionUID != "" {
			regionKeyByUID[*reg.RegionUID] = regionKey
		}
		regionKeyByName[normalizeName(*reg.Region)] = regionKey
	}

	districtByUID := make(map[string]districtRef)
	districtByRegionAndName := make(map[string]districtRef)
	districtByName := make(map[string]districtRef)
	seenDistrictByRegion := make(map[string]map[string]struct{})

	for _, dist := range grouped["3"] {
		if dist.District == nil {
			continue
		}

		regionKey, ok := resolveRegionKey(dist.RegionUID, dist.Region, regionKeyByUID, regionKeyByName)
		if !ok {
			continue
		}
		regionNode, ok := regionByKey[regionKey]
		if !ok {
			continue
		}

		districtKey := dedupeKey(dist.DistrictUID, dist.District, dist.DimOrgHierarchyKey)
		if _, ok := seenDistrictByRegion[regionKey]; !ok {
			seenDistrictByRegion[regionKey] = make(map[string]struct{})
		}
		if _, exists := seenDistrictByRegion[regionKey][districtKey]; exists {
			continue
		}
		seenDistrictByRegion[regionKey][districtKey] = struct{}{}

		node := &hierarchyNode{
			ID:    dist.DimOrgHierarchyKey,
			UID:   dist.DistrictUID,
			Name:  dist.District,
			Level: "3",
			Type:  "district",
		}
		regionNode.Children = append(regionNode.Children, node)

		ref := districtRef{Key: districtKey, Node: node}
		if dist.DistrictUID != nil && *dist.DistrictUID != "" {
			districtByUID[*dist.DistrictUID] = ref
		}

		lookupName := normalizeName(*dist.District)
		districtByRegionAndName[regionKey+"|"+lookupName] = ref
		if _, exists := districtByName[lookupName]; !exists {
			districtByName[lookupName] = ref
		}
	}

	subCountyUIDsByDistrict := make(map[string]map[string]struct{})
	subCountyNamesByDistrict := make(map[string]map[string]struct{})
	for _, sc := range grouped["5"] {
		ref, ok := resolveDistrictRef(sc, districtByUID, districtByRegionAndName, districtByName, regionKeyByUID, regionKeyByName)
		if !ok {
			continue
		}
		if sc.SubCountyUID != nil && *sc.SubCountyUID != "" {
			setAdd(subCountyUIDsByDistrict, ref.Key, *sc.SubCountyUID)
		}
		if sc.SubCounty != nil && *sc.SubCounty != "" {
			setAdd(subCountyNamesByDistrict, ref.Key, normalizeName(*sc.SubCounty))
		}
	}

	seenLocalGovtsByDistrict := make(map[string]map[string]struct{})
	for _, lg := range grouped["4"] {
		if lg.OrgUnitName == nil {
			continue
		}

		ref, ok := resolveDistrictRef(lg, districtByUID, districtByRegionAndName, districtByName, regionKeyByUID, regionKeyByName)
		if !ok {
			continue
		}

		localGovtKey := dedupeKey(lg.OrgUnitID, lg.OrgUnitName, lg.DimOrgHierarchyKey)
		if !markSeen(seenLocalGovtsByDistrict, ref.Key, localGovtKey) {
			continue
		}

		if lg.OrgUnitID != nil && *lg.OrgUnitID != "" && inSet(subCountyUIDsByDistrict, ref.Key, *lg.OrgUnitID) {
			continue
		}
		if inSet(subCountyNamesByDistrict, ref.Key, normalizeName(*lg.OrgUnitName)) {
			continue
		}

		ref.Node.Children = append(ref.Node.Children, &hierarchyNode{
			ID:    lg.DimOrgHierarchyKey,
			UID:   lg.OrgUnitID,
			Name:  lg.OrgUnitName,
			Level: "4",
			Type:  "localgovt",
		})
	}

	facByDistrictAndSubUID := make(map[string]map[string][]OrgUnit)
	facByDistrictAndSubName := make(map[string]map[string][]OrgUnit)
	for _, fac := range grouped["6"] {
		if fac.FacilityName == nil {
			continue
		}

		ref, ok := resolveDistrictRef(fac, districtByUID, districtByRegionAndName, districtByName, regionKeyByUID, regionKeyByName)
		if !ok {
			continue
		}

		if fac.SubCountyUID != nil && *fac.SubCountyUID != "" {
			if _, ok := facByDistrictAndSubUID[ref.Key]; !ok {
				facByDistrictAndSubUID[ref.Key] = make(map[string][]OrgUnit)
			}
			facByDistrictAndSubUID[ref.Key][*fac.SubCountyUID] = append(facByDistrictAndSubUID[ref.Key][*fac.SubCountyUID], fac)
		}
		if fac.SubCounty != nil && *fac.SubCounty != "" {
			if _, ok := facByDistrictAndSubName[ref.Key]; !ok {
				facByDistrictAndSubName[ref.Key] = make(map[string][]OrgUnit)
			}
			nameKey := normalizeName(*fac.SubCounty)
			facByDistrictAndSubName[ref.Key][nameKey] = append(facByDistrictAndSubName[ref.Key][nameKey], fac)
		}
	}

	seenSubCountiesByDistrict := make(map[string]map[string]struct{})
	for _, sc := range grouped["5"] {
		if sc.SubCounty == nil {
			continue
		}

		ref, ok := resolveDistrictRef(sc, districtByUID, districtByRegionAndName, districtByName, regionKeyByUID, regionKeyByName)
		if !ok {
			continue
		}

		subCountyKey := dedupeKey(sc.SubCountyUID, sc.SubCounty, sc.DimOrgHierarchyKey)
		if !markSeen(seenSubCountiesByDistrict, ref.Key, subCountyKey) {
			continue
		}

		facilities := collectFacilitiesForSubCounty(ref.Key, sc, facByDistrictAndSubUID, facByDistrictAndSubName)
		if len(facilities) == 0 {
			continue
		}

		subcountyNode := &hierarchyNode{
			ID:    sc.DimOrgHierarchyKey,
			UID:   sc.SubCountyUID,
			Name:  sc.SubCounty,
			Level: "5",
			Type:  "subcounty",
		}

		for _, fac := range facilities {
			subcountyNode.Children = append(subcountyNode.Children, &hierarchyNode{
				ID:       fac.DimOrgHierarchyKey,
				UID:      fac.FacilityUID,
				Name:     fac.FacilityName,
				Level:    "6",
				Type:     "facility",
				Division: fac.Division,
			})
		}

		ref.Node.Children = append(ref.Node.Children, subcountyNode)
	}

	tree := make([]TreeNode, 0, len(nationals))
	for _, national := range nationals {
		tree = append(tree, national.toDTO())
	}
	return tree
}

type hierarchyNode struct {
	ID       int64
	UID      *string
	Name     *string
	Level    string
	Type     string
	Division *string
	Children []*hierarchyNode
}

func (n *hierarchyNode) toDTO() TreeNode {
	node := TreeNode{
		ID:       n.ID,
		UID:      n.UID,
		Name:     n.Name,
		Level:    n.Level,
		Type:     n.Type,
		Division: n.Division,
	}
	if len(n.Children) > 0 {
		node.Children = make([]TreeNode, 0, len(n.Children))
		for _, child := range n.Children {
			node.Children = append(node.Children, child.toDTO())
		}
	}
	return node
}

type districtRef struct {
	Key  string
	Node *hierarchyNode
}

func resolveRegionKey(uid *string, name *string, byUID map[string]string, byName map[string]string) (string, bool) {
	if uid != nil && *uid != "" {
		if key, ok := byUID[*uid]; ok {
			return key, true
		}
	}
	if name != nil && *name != "" {
		if key, ok := byName[normalizeName(*name)]; ok {
			return key, true
		}
	}
	return "", false
}

func resolveDistrictRef(
	item OrgUnit,
	byUID map[string]districtRef,
	byRegionAndName map[string]districtRef,
	byName map[string]districtRef,
	regionKeyByUID map[string]string,
	regionKeyByName map[string]string,
) (districtRef, bool) {
	if item.DistrictUID != nil && *item.DistrictUID != "" {
		if ref, ok := byUID[*item.DistrictUID]; ok {
			return ref, true
		}
	}

	if item.District == nil || *item.District == "" {
		return districtRef{}, false
	}
	districtName := normalizeName(*item.District)

	if regionKey, ok := resolveRegionKey(item.RegionUID, item.Region, regionKeyByUID, regionKeyByName); ok {
		if ref, ok := byRegionAndName[regionKey+"|"+districtName]; ok {
			return ref, true
		}
	}
	if ref, ok := byName[districtName]; ok {
		return ref, true
	}

	return districtRef{}, false
}

func setAdd(sets map[string]map[string]struct{}, bucket string, value string) {
	if _, ok := sets[bucket]; !ok {
		sets[bucket] = make(map[string]struct{})
	}
	sets[bucket][value] = struct{}{}
}

func inSet(sets map[string]map[string]struct{}, bucket string, value string) bool {
	values, ok := sets[bucket]
	if !ok {
		return false
	}
	_, exists := values[value]
	return exists
}

func markSeen(seen map[string]map[string]struct{}, bucket string, key string) bool {
	if _, ok := seen[bucket]; !ok {
		seen[bucket] = make(map[string]struct{})
	}
	if _, exists := seen[bucket][key]; exists {
		return false
	}
	seen[bucket][key] = struct{}{}
	return true
}

func collectFacilitiesForSubCounty(
	districtKey string,
	subCounty OrgUnit,
	bySubUID map[string]map[string][]OrgUnit,
	bySubName map[string]map[string][]OrgUnit,
) []OrgUnit {
	out := []OrgUnit{}
	seen := make(map[string]struct{})
	appendUnique := func(list []OrgUnit) {
		for _, fac := range list {
			key := dedupeKey(fac.FacilityUID, fac.FacilityName, fac.DimOrgHierarchyKey)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, fac)
		}
	}

	if subCounty.SubCountyUID != nil && *subCounty.SubCountyUID != "" {
		if subMap, ok := bySubUID[districtKey]; ok {
			appendUnique(subMap[*subCounty.SubCountyUID])
		}
	}
	if subCounty.SubCounty != nil && *subCounty.SubCounty != "" {
		if subMap, ok := bySubName[districtKey]; ok {
			appendUnique(subMap[normalizeName(*subCounty.SubCounty)])
		}
	}
	return out
}

func dedupeKey(uid *string, name *string, id int64) string {
	if uid != nil && *uid != "" {
		return "uid:" + *uid
	}
	if name != nil && *name != "" {
		return "name:" + *name
	}
	return "id:" + strconv.FormatInt(id, 10)
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

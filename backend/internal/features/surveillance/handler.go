package surveillance

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	healthcontext "github.com/moh-sso-dashboard/internal/features/health_context"
	"github.com/moh-sso-dashboard/internal/http/response"
)

type Handler struct {
	epiWeekService               *EpiWeekService
	diseaseService               *DiseaseService
	locationService              *LocationService
	facilityWeeklyMetricsService *FacilityWeeklyMetricsService
	weeklyStatusService          *WeeklyStatusService
	alertService                 *AlertService
	importService                *ImportService
	healthContextService         *healthcontext.Service
}

func NewHandler(
	epiWeekService *EpiWeekService,
	diseaseService *DiseaseService,
	locationService *LocationService,
	facilityWeeklyMetricsService *FacilityWeeklyMetricsService,
	weeklyStatusService *WeeklyStatusService,
	alertService *AlertService,
	importService *ImportService,
	healthContextService *healthcontext.Service,
) *Handler {
	return &Handler{
		epiWeekService:               epiWeekService,
		diseaseService:               diseaseService,
		locationService:              locationService,
		facilityWeeklyMetricsService: facilityWeeklyMetricsService,
		weeklyStatusService:          weeklyStatusService,
		alertService:                 alertService,
		importService:                importService,
		healthContextService:         healthContextService,
	}
}

func healthContextScopeFromRequest(c *gin.Context) (HealthContextScope, bool) {
	value, ok := c.Get("health_context_id")
	if !ok {
		return HealthContextScope{}, false
	}
	id, ok := value.(uuid.UUID)
	if !ok || id == uuid.Nil {
		return HealthContextScope{}, false
	}
	scopeMode, _ := c.Get("health_context_scope_mode")
	return HealthContextScope{
		ID:                 id,
		IncludeDescendants: scopeMode == healthcontext.ScopeNodeAndDescendants,
	}, true
}

func (h *Handler) requireAliasInHealthContext(
	c *gin.Context,
	namespace string,
	externalID uuid.UUID,
) bool {
	scope, scoped := healthContextScopeFromRequest(c)
	if !scoped {
		return true
	}
	allowed, err := scope.containsAlias(c.Request.Context(), h.healthContextService, namespace, externalID)
	if err != nil || !allowed {
		response.Fail(c, http.StatusForbidden, "resource is outside the selected health context", "access denied")
		return false
	}
	return true
}

func parseWeeklyStatusListInput(c *gin.Context, validateStatus bool) (WeeklyStatusListInput, bool) {
	var input WeeklyStatusListInput

	if epiWeekID := c.Query("epiWeekID"); epiWeekID != "" {
		id, err := uuid.Parse(epiWeekID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", "request failed")
			return input, false
		}
		input.EpiWeekID = id
	}

	if regionID := c.Query("regionID"); regionID != "" {
		id, err := uuid.Parse(regionID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid regionID", "request failed")
			return input, false
		}
		input.RegionID = id
	}

	if districtID := c.Query("districtID"); districtID != "" {
		id, err := uuid.Parse(districtID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid districtID", "request failed")
			return input, false
		}
		input.DistrictID = id
	}

	if subCountyID := c.Query("subCountyID"); subCountyID != "" {
		id, err := uuid.Parse(subCountyID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid subCountyID", "request failed")
			return input, false
		}
		input.SubCountyID = id
	}

	if diseaseID := c.Query("diseaseID"); diseaseID != "" {
		id, err := uuid.Parse(diseaseID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid diseaseID", "request failed")
			return input, false
		}
		input.DiseaseID = id
	}

	if indicatorID := c.Query("indicatorID"); indicatorID != "" {
		id, err := uuid.Parse(indicatorID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid indicatorID", "request failed")
			return input, false
		}
		input.IndicatorID = id
	}

	if status := c.Query("status"); status != "" {
		if validateStatus && !isValidRiskLevel(status) {
			response.Fail(c, http.StatusBadRequest, "invalid status", status)
			return input, false
		}
		input.Status = status
	}

	return input, true
}

func isValidRiskLevel(status string) bool {
	switch status {
	case "MAROON", "RED", "YELLOW", "GREEN":
		return true
	default:
		return false
	}
}

// weeks
func (h *Handler) ListEpiWeeksByYearWeek(c *gin.Context) {
	ctx := c.Request.Context()

	yearStr := c.Query("year")
	if yearStr == "" {
		response.Fail(c, http.StatusBadRequest, "year is required", "nil")
		return
	}

	weekStr := c.Query("week")
	if weekStr == "" {
		response.Fail(c, http.StatusBadRequest, "week is required", "nil")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid year", "request failed")
		return
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid week", "request failed")
		return
	}

	data, err := h.epiWeekService.GetEpiWeekByYearWeekValue(ctx, int32(year), int32(week))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to get epi week by year and week", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toEpiWeekResponse(data))
}

func (h *Handler) ListEpiWeeksByYear(c *gin.Context) {
	ctx := c.Request.Context()

	yearStr := c.Query("year")
	if yearStr == "" {
		response.Fail(c, http.StatusBadRequest, "year is required", "nil")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid year", "request failed")
		return
	}

	data, err := h.epiWeekService.ListEpiWeeksByYear(ctx, int32(year))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list epi weeks by year", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toEpiWeekResponses(data))
}

// diseases
func (h *Handler) ListDiseases(c *gin.Context) {
	ctx := c.Request.Context()

	data, err := h.diseaseService.ListDiseases(ctx)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list diseases", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toDiseaseResponses(data))
}

// locations
func (h *Handler) ListRegions(c *gin.Context) {
	ctx := c.Request.Context()

	scope, scoped := healthContextScopeFromRequest(c)
	var (
		data []db.Region
		err  error
	)
	if scoped {
		data, err = h.locationService.ListRegionsInHealthContext(ctx, scope)
	} else {
		data, err = h.locationService.ListRegions(ctx)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list regions", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toRegionResponses(data))
}

func (h *Handler) ListDistricts(c *gin.Context) {
	ctx := c.Request.Context()

	scope, scoped := healthContextScopeFromRequest(c)
	var (
		data []db.ListDistrictsRow
		err  error
	)
	if scoped {
		data, err = h.locationService.ListDistrictsInHealthContext(ctx, scope)
	} else {
		data, err = h.locationService.ListDistricts(ctx)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list districts", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toDistrictListResponses(data))
}

func (h *Handler) ListDistrictsByRegion(c *gin.Context) {
	ctx := c.Request.Context()

	regionIDParam := c.Param("regionID")
	if regionIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "region id is required", "nil")
		return
	}

	regionID, err := uuid.Parse(regionIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid region id", "request failed")
		return
	}
	if !h.requireAliasInHealthContext(c, healthContextNamespaceRegion, regionID) {
		return
	}

	scope, scoped := healthContextScopeFromRequest(c)
	var data []db.District
	if scoped {
		data, err = h.locationService.ListDistrictsByRegionInHealthContext(ctx, regionID, scope)
	} else {
		data, err = h.locationService.ListDistrictsByRegion(ctx, regionID)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list districts by region", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toDistrictResponses(data))
}
func (h *Handler) ListSubcountiesByDistrict(c *gin.Context) {
	ctx := c.Request.Context()

	districtIDParam := c.Param("districtID")
	if districtIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "districtID is required", "nil")
		return
	}

	districtID, err := uuid.Parse(districtIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid districtID", "request failed")
		return
	}
	if !h.requireAliasInHealthContext(c, healthContextNamespaceDistrict, districtID) {
		return
	}

	scope, scoped := healthContextScopeFromRequest(c)
	var data []db.SubCounty
	if scoped {
		data, err = h.locationService.ListSubcountiesByDistrictInHealthContext(ctx, districtID, scope)
	} else {
		data, err = h.locationService.ListSubcountiesByDistrict(ctx, districtID)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list subcounties", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toSubCountyResponses(data))
}

func (h *Handler) GetSubcountyByID(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	if idParam == "" {
		response.Fail(c, http.StatusBadRequest, "id is required", "")
		return
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id", "request failed")
		return
	}
	if !h.requireAliasInHealthContext(c, healthContextNamespaceSubCounty, id) {
		return
	}

	data, err := h.locationService.GetSubcountyByID(ctx, id)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to get subcounty", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toSubCountyResponse(data))
}

func (h *Handler) UpsertDistrict(c *gin.Context) {
	ctx := c.Request.Context()

	var req UpsertDistrictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid request body", "request failed")
		return
	}

	data, err := h.locationService.UpsertDistrictFromRequest(ctx, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to upsert district", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toDistrictResponse(data))
}

func (h *Handler) UpsertSubcounty(c *gin.Context) {
	ctx := c.Request.Context()

	var req UpsertSubCountyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid request body", "request failed")
		return
	}
	if req.DistrictID == nil || *req.DistrictID == uuid.Nil {
		response.Fail(c, http.StatusBadRequest, "invalid request body", "district_id is required")
		return
	}

	data, err := h.locationService.UpsertSubcountyFromRequest(ctx, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to upsert subcounty", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toSubCountyResponse(data))
}

func (h *Handler) DeleteSubcounty(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	if idParam == "" {
		response.Fail(c, http.StatusBadRequest, "id is required", "nil")
		return
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id", "request failed")
		return
	}

	if err := h.locationService.DeleteSubcounty(ctx, id); err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to delete subcounty", "request failed")
		return
	}

	response.OK(c, http.StatusOK, DeleteResponse{Deleted: true})
}

// facility weekly metrics
func (h *Handler) ListFacilityWeeklyMetricsByWeek(c *gin.Context) {
	ctx := c.Request.Context()

	epiWeekIDParam := c.Param("epiWeekID")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiWeekID is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", "request failed")
		return
	}

	scope, scoped := healthContextScopeFromRequest(c)
	var diseaseMetrics []db.ListFacilityWeeklyDiseaseMetricsByWeekRow
	if scoped {
		diseaseMetrics, err = h.facilityWeeklyMetricsService.
			ListFacilityWeeklyDiseaseMetricsByWeekInHealthContext(ctx, epiWeekID, scope)
	} else {
		diseaseMetrics, err = h.facilityWeeklyMetricsService.ListFacilityWeeklyDiseaseMetricsByWeek(ctx, epiWeekID)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list facility weekly disease metrics by week", "request failed")
		return
	}

	data := make([]FacilityWeeklyMetricResponse, 0, len(diseaseMetrics))
	data = append(data, toFacilityWeeklyDiseaseMetricResponses(diseaseMetrics)...)

	response.OK(c, http.StatusOK, data)
}

func (h *Handler) ListFacilityWeeklyMetricsByFacility(c *gin.Context) {
	ctx := c.Request.Context()

	facilityIDParam := c.Param("facilityID")
	if facilityIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "facilityID is required", "nil")
		return
	}

	facilityID, err := uuid.Parse(facilityIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid facilityID", "request failed")
		return
	}
	if !h.requireAliasInHealthContext(c, healthContextNamespaceFacility, facilityID) {
		return
	}

	data, err := h.facilityWeeklyMetricsService.ListFacilityWeeklyMetricsByFacility(ctx, facilityID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list facility weekly metrics by facility", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toFacilityMetricByFacilityResponses(data))
}

func (h *Handler) ListFacilityDiseaseMetricsTrend(c *gin.Context) {
	ctx := c.Request.Context()

	facilityIDParam := c.Param("facilityID")
	if facilityIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "facilityID is required", "nil")
		return
	}

	diseaseIDParam := c.Param("diseaseID")
	if diseaseIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "diseaseID is required", "nil")
		return
	}

	facilityID, err := uuid.Parse(facilityIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid facilityID", "request failed")
		return
	}
	if !h.requireAliasInHealthContext(c, healthContextNamespaceFacility, facilityID) {
		return
	}

	diseaseID, err := uuid.Parse(diseaseIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid diseaseID", "request failed")
		return
	}

	data, err := h.facilityWeeklyMetricsService.ListFacilityDiseaseMetricsTrend(ctx, facilityID, diseaseID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list facility disease trend", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toFacilityDiseaseMetricTrendResponses(data))
}

func (h *Handler) ListFacilityIndicatorMetricsTrend(c *gin.Context) {
	ctx := c.Request.Context()

	facilityIDParam := c.Param("facilityID")
	if facilityIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "facilityID is required", "nil")
		return
	}

	indicatorIDParam := c.Param("indicatorID")
	if indicatorIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "indicatorID is required", "nil")
		return
	}

	facilityID, err := uuid.Parse(facilityIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid facilityID", "request failed")
		return
	}
	if !h.requireAliasInHealthContext(c, healthContextNamespaceFacility, facilityID) {
		return
	}

	indicatorID, err := uuid.Parse(indicatorIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid indicatorID", "request failed")
		return
	}

	data, err := h.facilityWeeklyMetricsService.ListFacilityIndicatorMetricsTrend(ctx, facilityID, indicatorID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list facility indicator trend", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toFacilityIndicatorMetricTrendResponses(data))
}

func (h *Handler) ListFacilityDiseaseMetricsByWeekAndDisease(c *gin.Context) {
	ctx := c.Request.Context()

	epiWeekIDParam := c.Param("epiWeekID")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiWeekID is required", "nil")
		return
	}

	diseaseIDParam := c.Param("diseaseID")
	if diseaseIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "diseaseID is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", "request failed")
		return
	}

	diseaseID, err := uuid.Parse(diseaseIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid diseaseID", "request failed")
		return
	}

	scope, scoped := healthContextScopeFromRequest(c)
	var rows []db.ListFacilityWeeklyDiseaseMetricsByWeekAndDiseaseRow
	if scoped {
		rows, err = h.facilityWeeklyMetricsService.
			ListFacilityDiseaseMetricsByWeekAndDiseaseInHealthContext(ctx, epiWeekID, diseaseID, scope)
	} else {
		rows, err = h.facilityWeeklyMetricsService.
			ListFacilityDiseaseMetricsByWeekAndDisease(ctx, epiWeekID, diseaseID)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list facility disease metrics by week and disease", "request failed")
		return
	}

	data := make([]FacilityWeeklyMetricResponse, 0, len(rows))
	data = append(data, toFacilityWeeklyDiseaseMetricByWeekAndDiseaseResponses(rows)...)

	response.OK(c, http.StatusOK, data)
}

func (h *Handler) ListDiseaseWeeklyTrendAggregated(c *gin.Context) {
	ctx := c.Request.Context()

	epiYearParam := c.Query("epiYear")
	if epiYearParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiYear is required", "nil")
		return
	}

	diseaseIDParam := c.Query("diseaseID")
	if diseaseIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "diseaseID is required", "nil")
		return
	}

	epiYear64, err := strconv.ParseInt(epiYearParam, 10, 32)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiYear", "request failed")
		return
	}

	diseaseID, err := uuid.Parse(diseaseIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid diseaseID", "request failed")
		return
	}

	var regionID *uuid.UUID
	regionIDParam := c.Query("regionID")
	if regionIDParam != "" {
		parsedRegionID, err := uuid.Parse(regionIDParam)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid regionID", "request failed")
			return
		}
		regionID = &parsedRegionID
	}

	var districtID *uuid.UUID
	districtIDParam := c.Query("districtID")
	if districtIDParam != "" {
		parsedDistrictID, err := uuid.Parse(districtIDParam)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid districtID", "request failed")
			return
		}
		districtID = &parsedDistrictID
	}

	scope, scoped := healthContextScopeFromRequest(c)
	var data []db.ListDiseaseWeeklyTrendAggregatedRow
	if scoped {
		data, err = h.facilityWeeklyMetricsService.ListDiseaseWeeklyTrendAggregatedInHealthContext(
			ctx, int32(epiYear64), diseaseID, regionID, districtID, scope,
		)
	} else {
		data, err = h.facilityWeeklyMetricsService.ListDiseaseWeeklyTrendAggregated(
			ctx, int32(epiYear64), diseaseID, regionID, districtID,
		)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list aggregated disease weekly trend", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toDiseaseWeeklyTrendAggregateResponses(data))
}

// district weekly status

func (h *Handler) ListDistrictWeeklyStatusesByWeek(c *gin.Context) {
	ctx := c.Request.Context()

	epiWeekIDParam := c.Param("epiWeekID")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiWeekID is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", "request failed")
		return
	}

	scope, scoped := healthContextScopeFromRequest(c)
	var data []db.WeeklyStatus
	if scoped {
		data, err = h.weeklyStatusService.ListLevelByWeekInHealthContext(ctx, epiWeekID, "district", scope)
	} else {
		data, err = h.weeklyStatusService.ListDistrictByWeek(ctx, epiWeekID)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list district weekly statuses by week", "request failed")
		return
	}

	response.OK(c, http.StatusOK, data)
}

// weekly status

func (h *Handler) ListWeeklyStatuses(c *gin.Context) {
	ctx := c.Request.Context()

	params, ok := parseWeeklyStatusListInput(c, false)
	if !ok {
		return
	}

	if params.RegionID != uuid.Nil &&
		!h.requireAliasInHealthContext(c, healthContextNamespaceRegion, params.RegionID) {
		return
	}
	if params.DistrictID != uuid.Nil &&
		!h.requireAliasInHealthContext(c, healthContextNamespaceDistrict, params.DistrictID) {
		return
	}
	if params.SubCountyID != uuid.Nil &&
		!h.requireAliasInHealthContext(c, healthContextNamespaceSubCounty, params.SubCountyID) {
		return
	}
	scope, scoped := healthContextScopeFromRequest(c)
	var (
		data []db.WeeklyStatus
		err  error
	)
	if scoped {
		data, err = h.weeklyStatusService.ListFromInputInHealthContext(ctx, params, scope)
	} else {
		data, err = h.weeklyStatusService.ListFromInput(ctx, params)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list weekly statuses", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toWeeklyStatusResponses(data))
}

func (h *Handler) ListWeeklyStatusesDetailed(c *gin.Context) {
	ctx := c.Request.Context()

	params, ok := parseWeeklyStatusListInput(c, true)
	if !ok {
		return
	}

	if params.RegionID != uuid.Nil &&
		!h.requireAliasInHealthContext(c, healthContextNamespaceRegion, params.RegionID) {
		return
	}
	if params.DistrictID != uuid.Nil &&
		!h.requireAliasInHealthContext(c, healthContextNamespaceDistrict, params.DistrictID) {
		return
	}
	if params.SubCountyID != uuid.Nil &&
		!h.requireAliasInHealthContext(c, healthContextNamespaceSubCounty, params.SubCountyID) {
		return
	}
	scope, scoped := healthContextScopeFromRequest(c)
	var (
		rows []db.ListWeeklyStatusesDetailedRow
		err  error
	)
	if scoped {
		rows, err = h.weeklyStatusService.ListDetailedFromInputInHealthContext(ctx, params, scope)
	} else {
		rows, err = h.weeklyStatusService.ListDetailedFromInput(ctx, params)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list weekly statuses detailed", "request failed")
		return
	}

	data := make([]WeeklyStatusDetailedResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toWeeklyStatusDetailedResponse(row))
	}

	response.OK(c, http.StatusOK, data)
}

// region weekly status

func (h *Handler) ListRegionWeeklyStatusesByWeek(c *gin.Context) {
	ctx := c.Request.Context()

	epiWeekIDParam := c.Param("epiWeekID")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiWeekID is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", "request failed")
		return
	}

	scope, scoped := healthContextScopeFromRequest(c)
	var data []db.WeeklyStatus
	if scoped {
		data, err = h.weeklyStatusService.ListLevelByWeekInHealthContext(ctx, epiWeekID, "region", scope)
	} else {
		data, err = h.weeklyStatusService.ListRegionByWeek(ctx, epiWeekID)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list region weekly statuses by week", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toWeeklyStatusResponses(data))
}

// national weekly status

func (h *Handler) ListNationalWeeklyStatusesByWeek(c *gin.Context) {
	ctx := c.Request.Context()

	epiWeekIDParam := c.Param("epiWeekID")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiWeekID is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", "request failed")
		return
	}

	scope, scoped := healthContextScopeFromRequest(c)
	var data []db.WeeklyStatus
	if scoped {
		data, err = h.weeklyStatusService.ListLevelByWeekInHealthContext(ctx, epiWeekID, "national", scope)
	} else {
		data, err = h.weeklyStatusService.ListNationalByWeek(ctx, epiWeekID)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list national weekly statuses by week", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toWeeklyStatusResponses(data))
}

// alerts
func (h *Handler) ListAlerts(c *gin.Context) {
	ctx := c.Request.Context()

	params := AlertListParams{}
	var err error

	if value := c.Query("epiWeekID"); value != "" {
		params.EpiWeekID, err = uuid.Parse(value)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", "request failed")
			return
		}
	}

	if value := c.Query("diseaseID"); value != "" {
		params.DiseaseID, err = uuid.Parse(value)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid diseaseID", "request failed")
			return
		}
	}

	if value := c.Query("districtID"); value != "" {
		params.DistrictID, err = uuid.Parse(value)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid districtID", "request failed")
			return
		}
		if !h.requireAliasInHealthContext(c, healthContextNamespaceDistrict, params.DistrictID) {
			return
		}
	}

	if value := c.Query("regionID"); value != "" {
		params.RegionID, err = uuid.Parse(value)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid regionID", "request failed")
			return
		}
		if !h.requireAliasInHealthContext(c, healthContextNamespaceRegion, params.RegionID) {
			return
		}
	}

	scope, scoped := healthContextScopeFromRequest(c)
	var data []db.ListAlertsRow
	if scoped {
		data, err = h.alertService.ListAlertsFromParamsInHealthContext(ctx, params, scope)
	} else {
		data, err = h.alertService.ListAlertsFromParams(ctx, params)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list alerts", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toAlertResponses(data))
}

// imports

func (h *Handler) ListImportBatches(c *gin.Context) {
	ctx := c.Request.Context()

	data, err := h.importService.ListImportBatches(ctx)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list import batches", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toImportBatchResponses(data))
}

func (h *Handler) GetImportBatchByID(c *gin.Context) {
	ctx := c.Request.Context()

	batchIDParam := c.Param("batchID")
	if batchIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "batchID is required", "nil")
		return
	}

	batchID, err := uuid.Parse(batchIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid batchID", "request failed")
		return
	}

	data, err := h.importService.GetImportBatchByID(ctx, batchID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to get import batch", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toImportBatchResponse(data))
}

func (h *Handler) ListImportRawRowsByBatch(c *gin.Context) {
	ctx := c.Request.Context()

	batchIDParam := c.Param("batchID")
	if batchIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "batchID is required", "nil")
		return
	}

	batchID, err := uuid.Parse(batchIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid batchID", "request failed")
		return
	}

	data, err := h.importService.ListImportRawRowsByBatch(ctx, batchID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list import raw rows", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toImportRawRowResponses(data))
}

func (h *Handler) CreateImportBatch(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateImportBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid request body", "request failed")
		return
	}

	status := req.Status
	if status == "" {
		status = "pending"
	}

	data, err := h.importService.CreateImportBatchFromInput(ctx, CreateImportBatchInput{
		SourceName:  req.SourceName,
		FileName:    req.FileName,
		DatasetType: req.DatasetType,
		ImportedBy:  req.ImportedBy,
		Status:      status,
		Notes:       req.Notes,
		DocumentID:  req.DocumentID,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to create import batch", "request failed")
		return
	}

	response.OK(c, http.StatusCreated, toImportBatchResponse(data))
}

func (h *Handler) UpdateImportBatchStatus(c *gin.Context) {
	ctx := c.Request.Context()

	batchIDParam := c.Param("batchID")
	if batchIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "batchID is required", "nil")
		return
	}

	batchID, err := uuid.Parse(batchIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid batchID", "request failed")
		return
	}

	var req UpdateImportBatchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid request body", "request failed")
		return
	}

	data, err := h.importService.UpdateImportBatchStatusFromInput(ctx, UpdateImportBatchStatusInput{
		ID:     batchID,
		Status: req.Status,
		Notes:  req.Notes,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to update import batch status", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toImportBatchResponse(data))
}

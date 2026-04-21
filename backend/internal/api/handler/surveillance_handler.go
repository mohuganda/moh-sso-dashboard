package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/service"
)

type SurveillanceHandler struct {
	epiWeekService               *service.SurveillanceEpiWeekService
	diseaseService               *service.SurveillanceDiseaseService
	locationService              *service.SurveillanceLocationService
	facilityWeeklyMetricsService *service.SurveillanceFacilityWeeklyMetricsService
	weeklyStatusService          *service.SurveillanceWeeklyStatusService
	alertService                 *service.SurveillanceAlertService
	importService                *service.SurveillanceImportService
}

type AlertListParams struct {
	EpiWeekID  uuid.UUID
	DiseaseID  uuid.UUID
	DistrictID uuid.UUID
	RegionID   uuid.UUID
}

type AlertResponse struct {
	ID           uuid.UUID  `json:"id"`
	ExternalID   *string    `json:"external_id,omitempty"`
	DiseaseID    uuid.UUID  `json:"disease_id"`
	DiseaseName  *string    `json:"disease_name,omitempty"`
	DistrictID   *uuid.UUID `json:"district_id,omitempty"`
	DistrictName *string    `json:"district_name,omitempty"`
	RegionID     *uuid.UUID `json:"region_id,omitempty"`
	EpiWeekID    *uuid.UUID `json:"epi_week_id,omitempty"`
	EpiYear      *int32     `json:"epi_year,omitempty"`
	EpiWeek      *int32     `json:"epi_week,omitempty"`
	OccurredOn   *time.Time `json:"occurred_on,omitempty"`
	CreatedOn    *time.Time `json:"created_on,omitempty"`
	Narrative    string     `json:"narrative"`
	SubmittedBy  *string    `json:"submitted_by,omitempty"`
	Status       string     `json:"status"`
	SourceName   *string    `json:"source_name,omitempty"`
	ImportedAt   time.Time  `json:"imported_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func toAlertResponse(row db.ListAlertsRow) AlertResponse {
	var externalID *string
	if row.ExternalID.Valid {
		externalID = &row.ExternalID.String
	}

	var diseaseName *string
	if row.DiseaseName.Valid {
		diseaseName = &row.DiseaseName.String
	}

	var districtID *uuid.UUID
	if row.DistrictID.Valid {
		districtID = &row.DistrictID.UUID
	}

	var districtName *string
	if row.DistrictName.Valid {
		districtName = &row.DistrictName.String
	}

	var regionID *uuid.UUID
	if row.RegionID.Valid {
		regionID = &row.RegionID.UUID
	}

	var epiWeekID *uuid.UUID
	if row.EpiWeekID.Valid {
		epiWeekID = &row.EpiWeekID.UUID
	}

	var occurredOn *time.Time
	if row.OccurredOn.Valid {
		occurredOn = &row.OccurredOn.Time
	}

	var createdOn *time.Time
	if row.CreatedOn.Valid {
		createdOn = &row.CreatedOn.Time
	}

	var submittedBy *string
	if row.SubmittedBy.Valid {
		submittedBy = &row.SubmittedBy.String
	}

	var sourceName *string
	if row.SourceName.Valid {
		sourceName = &row.SourceName.String
	}

	return AlertResponse{
		ID:           row.ID,
		ExternalID:   externalID,
		DiseaseID:    row.DiseaseID,
		DiseaseName:  diseaseName,
		DistrictID:   districtID,
		DistrictName: districtName,
		RegionID:     regionID,
		EpiWeekID:    epiWeekID,
		OccurredOn:   occurredOn,
		CreatedOn:    createdOn,
		Narrative:    row.Narrative,
		SubmittedBy:  submittedBy,
		Status:       string(row.Status),
		SourceName:   sourceName,
		ImportedAt:   row.ImportedAt,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func toAlertResponses(items []db.ListAlertsRow) []AlertResponse {
	out := make([]AlertResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toAlertResponse(item))
	}
	return out
}

type WeeklyStatusDetailedResponse struct {
	ID            uuid.UUID  `json:"id"`
	RegionID      *uuid.UUID `json:"region_id,omitempty"`
	DistrictID    *uuid.UUID `json:"district_id,omitempty"`
	SubCountyID   *uuid.UUID `json:"sub_county_id,omitempty"`
	DiseaseID     *uuid.UUID `json:"disease_id,omitempty"`
	IndicatorID   *uuid.UUID `json:"indicator_id,omitempty"`
	EpiWeekID     uuid.UUID  `json:"epi_week_id"`
	Status        string     `json:"status"`
	SourceName    string     `json:"source_name"`
	ImportedAt    time.Time  `json:"imported_at"`
	CreatedAt     time.Time  `json:"created_at"`
	RegionName    string     `json:"region_name"`
	DistrictName  string     `json:"district_name"`
	SubCountyName string     `json:"sub_county_name"`
	DiseaseName   string     `json:"disease_name"`
	IndicatorName string     `json:"indicator_name"`
}

func uuidPtr(v uuid.UUID) *uuid.UUID {
	id := v
	return &id
}

func sqlNullStringValue(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return strings.TrimSpace(ns.String)
}

func toWeeklyStatusDetailedResponse(row db.ListWeeklyStatusesDetailedRow) WeeklyStatusDetailedResponse {
	var regionID *uuid.UUID
	if row.RegionID.Valid {
		regionID = uuidPtr(row.RegionID.UUID)
	}

	var districtID *uuid.UUID
	if row.DistrictID.Valid {
		districtID = uuidPtr(row.DistrictID.UUID)
	}

	var subCountyID *uuid.UUID
	if row.SubCountyID.Valid {
		subCountyID = uuidPtr(row.SubCountyID.UUID)
	}

	var diseaseID *uuid.UUID
	if row.DiseaseID.Valid {
		diseaseID = uuidPtr(row.DiseaseID.UUID)
	}

	var indicatorID *uuid.UUID
	if row.IndicatorID.Valid {
		indicatorID = uuidPtr(row.IndicatorID.UUID)
	}

	return WeeklyStatusDetailedResponse{
		ID:            row.ID,
		RegionID:      regionID,
		DistrictID:    districtID,
		SubCountyID:   subCountyID,
		DiseaseID:     diseaseID,
		IndicatorID:   indicatorID,
		EpiWeekID:     row.EpiWeekID,
		Status:        string(row.Status),
		SourceName:    sqlNullStringValue(row.SourceName),
		ImportedAt:    row.ImportedAt,
		CreatedAt:     row.CreatedAt,
		RegionName:    sqlNullStringValue(row.RegionName),
		DistrictName:  sqlNullStringValue(row.DistrictName),
		SubCountyName: sqlNullStringValue(row.SubCountyName),
		DiseaseName:   sqlNullStringValue(row.DiseaseName),
		IndicatorName: sqlNullStringValue(row.IndicatorName),
	}
}
func NewSurveillanceHandler(
	epiWeekService *service.SurveillanceEpiWeekService,
	diseaseService *service.SurveillanceDiseaseService,
	locationService *service.SurveillanceLocationService,
	facilityWeeklyMetricsService *service.SurveillanceFacilityWeeklyMetricsService,
	weeklyStatusService *service.SurveillanceWeeklyStatusService,
	alertService *service.SurveillanceAlertService,
	importService *service.SurveillanceImportService,

) *SurveillanceHandler {
	return &SurveillanceHandler{
		epiWeekService:               epiWeekService,
		diseaseService:               diseaseService,
		locationService:              locationService,
		facilityWeeklyMetricsService: facilityWeeklyMetricsService,
		weeklyStatusService:          weeklyStatusService,
		alertService:                 alertService,
		importService:                importService,
	}
}

// weeks
func (h *SurveillanceHandler) ListEpiWeeksByYearWeek(c *gin.Context) {
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
		response.Fail(c, http.StatusBadRequest, "invalid year", err.Error())
		return
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid week", err.Error())
		return
	}

	params := db.GetEpiWeekByYearWeekParams{
		EpiYear: int32(year),
		EpiWeek: int32(week),
	}

	data, err := h.epiWeekService.GetEpiWeekByYearWeek(ctx, params)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to get epi week by year and week", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) ListEpiWeeksByYear(c *gin.Context) {
	ctx := c.Request.Context()

	yearStr := c.Query("year")
	if yearStr == "" {
		response.Fail(c, http.StatusBadRequest, "year is required", "nil")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid year", err.Error())
		return
	}

	data, err := h.epiWeekService.ListEpiWeeksByYear(ctx, int32(year))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list epi weeks by year", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

// diseases
func (h *SurveillanceHandler) ListDiseases(c *gin.Context) {
	ctx := c.Request.Context()

	data, err := h.diseaseService.ListDiseases(ctx)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list diseases", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

// locations
func (h *SurveillanceHandler) ListRegions(c *gin.Context) {
	ctx := c.Request.Context()

	data, err := h.locationService.ListRegions(ctx)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list regions", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) ListDistricts(c *gin.Context) {
	ctx := c.Request.Context()

	data, err := h.locationService.ListDistricts(ctx)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list districts", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) ListDistrictsByRegion(c *gin.Context) {
	ctx := c.Request.Context()

	regionIDParam := c.Param("regionID")
	if regionIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "region id is required", "nil")
		return
	}

	regionID, err := uuid.Parse(regionIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid region id", err.Error())
		return
	}

	data, err := h.locationService.ListDistrictsByRegion(ctx, regionID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list districts by region", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}
func (h *SurveillanceHandler) ListSubcountiesByDistrict(c *gin.Context) {
	ctx := c.Request.Context()

	districtIDParam := c.Param("districtID")
	if districtIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "districtID is required", "nil")
		return
	}

	districtID, err := uuid.Parse(districtIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid districtID", err.Error())
		return
	}

	data, err := h.locationService.ListSubcountiesByDistrict(ctx, districtID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list subcounties", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) GetSubcountyByID(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	if idParam == "" {
		response.Fail(c, http.StatusBadRequest, "id is required", "")
		return
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id", err.Error())
		return
	}

	data, err := h.locationService.GetSubcountyByID(ctx, id)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to get subcounty", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) UpsertDistrict(c *gin.Context) {
	ctx := c.Request.Context()

	var req db.UpsertDistrictParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	data, err := h.locationService.UpsertDistrict(ctx, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to upsert district", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) UpsertSubcounty(c *gin.Context) {
	ctx := c.Request.Context()

	var req db.UpsertSubCountyParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	data, err := h.locationService.UpsertSubcounty(ctx, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to upsert subcounty", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) DeleteSubcounty(c *gin.Context) {
	ctx := c.Request.Context()

	idParam := c.Param("id")
	if idParam == "" {
		response.Fail(c, http.StatusBadRequest, "id is required", "nil")
		return
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id", err.Error())
		return
	}

	if err := h.locationService.DeleteSubcounty(ctx, id); err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to delete subcounty", err.Error())
		return
	}

	response.OK(c, http.StatusOK, gin.H{"deleted": true})
}

// facility weekly metrics
func (h *SurveillanceHandler) ListFacilityWeeklyMetricsByWeek(c *gin.Context) {
	ctx := c.Request.Context()

	epiWeekIDParam := c.Param("epiWeekID")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiWeekID is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", err.Error())
		return
	}

	data, err := h.facilityWeeklyMetricsService.ListFacilityWeeklyMetricsByWeek(ctx, epiWeekID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list facility weekly metrics by week", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) ListFacilityWeeklyMetricsByFacility(c *gin.Context) {
	ctx := c.Request.Context()

	facilityIDParam := c.Param("facilityID")
	if facilityIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "facilityID is required", "nil")
		return
	}

	facilityID, err := uuid.Parse(facilityIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid facilityID", err.Error())
		return
	}

	data, err := h.facilityWeeklyMetricsService.ListFacilityWeeklyMetricsByFacility(ctx, facilityID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list facility weekly metrics by facility", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) ListFacilityDiseaseMetricsTrend(c *gin.Context) {
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
		response.Fail(c, http.StatusBadRequest, "invalid facilityID", err.Error())
		return
	}

	diseaseID, err := uuid.Parse(diseaseIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid diseaseID", err.Error())
		return
	}

	data, err := h.facilityWeeklyMetricsService.ListFacilityDiseaseMetricsTrend(ctx, facilityID, diseaseID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list facility disease trend", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) ListFacilityIndicatorMetricsTrend(c *gin.Context) {
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
		response.Fail(c, http.StatusBadRequest, "invalid facilityID", err.Error())
		return
	}

	indicatorID, err := uuid.Parse(indicatorIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid indicatorID", err.Error())
		return
	}

	data, err := h.facilityWeeklyMetricsService.ListFacilityIndicatorMetricsTrend(ctx, facilityID, indicatorID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list facility indicator trend", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

// district weekly status

func (h *SurveillanceHandler) ListDistrictWeeklyStatusesByWeek(c *gin.Context) {
	ctx := c.Request.Context()

	epiWeekIDParam := c.Param("epiWeekID")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiWeekID is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", err.Error())
		return
	}

	data, err := h.weeklyStatusService.ListDistrictByWeek(ctx, epiWeekID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list district weekly statuses by week", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

// weekly status

func (h *SurveillanceHandler) ListWeeklyStatuses(c *gin.Context) {
	ctx := c.Request.Context()

	params := db.ListWeeklyStatusesParams{}

	if epiWeekID := c.Query("epiWeekID"); epiWeekID != "" {
		id, err := uuid.Parse(epiWeekID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", err.Error())
			return
		}
		params.EpiWeekID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if regionID := c.Query("regionID"); regionID != "" {
		id, err := uuid.Parse(regionID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid regionID", err.Error())
			return
		}
		params.RegionID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if districtID := c.Query("districtID"); districtID != "" {
		id, err := uuid.Parse(districtID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid districtID", err.Error())
			return
		}
		params.DistrictID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if subCountyID := c.Query("subCountyID"); subCountyID != "" {
		id, err := uuid.Parse(subCountyID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid subCountyID", err.Error())
			return
		}
		params.SubCountyID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if diseaseID := c.Query("diseaseID"); diseaseID != "" {
		id, err := uuid.Parse(diseaseID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid diseaseID", err.Error())
			return
		}
		params.DiseaseID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if indicatorID := c.Query("indicatorID"); indicatorID != "" {
		id, err := uuid.Parse(indicatorID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid indicatorID", err.Error())
			return
		}
		params.IndicatorID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if status := c.Query("status"); status != "" {
		params.Status = db.NullRiskLevel{
			RiskLevel: db.RiskLevel(status),
			Valid:     true,
		}
	}

	data, err := h.weeklyStatusService.List(ctx, params)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list weekly statuses", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) ListWeeklyStatusesDetailed(c *gin.Context) {
	ctx := c.Request.Context()

	params := db.ListWeeklyStatusesDetailedParams{}

	if epiWeekID := c.Query("epiWeekID"); epiWeekID != "" {
		id, err := uuid.Parse(epiWeekID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", err.Error())
			return
		}
		params.EpiWeekID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if regionID := c.Query("regionID"); regionID != "" {
		id, err := uuid.Parse(regionID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid regionID", err.Error())
			return
		}
		params.RegionID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if districtID := c.Query("districtID"); districtID != "" {
		id, err := uuid.Parse(districtID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid districtID", err.Error())
			return
		}
		params.DistrictID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if subCountyID := c.Query("subCountyID"); subCountyID != "" {
		id, err := uuid.Parse(subCountyID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid subCountyID", err.Error())
			return
		}
		params.SubCountyID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if diseaseID := c.Query("diseaseID"); diseaseID != "" {
		id, err := uuid.Parse(diseaseID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid diseaseID", err.Error())
			return
		}
		params.DiseaseID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if indicatorID := c.Query("indicatorID"); indicatorID != "" {
		id, err := uuid.Parse(indicatorID)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid indicatorID", err.Error())
			return
		}
		params.IndicatorID = uuid.NullUUID{UUID: id, Valid: true}
	}

	if status := c.Query("status"); status != "" {
		switch db.RiskLevel(status) {
		case db.RiskLevelMAROON, db.RiskLevelRED, db.RiskLevelYELLOW, db.RiskLevelGREEN:
			params.Status = db.NullRiskLevel{
				RiskLevel: db.RiskLevel(status),
				Valid:     true,
			}
		default:
			response.Fail(c, http.StatusBadRequest, "invalid status", status)
			return
		}
	}

	rows, err := h.weeklyStatusService.ListDetailed(ctx, params)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list weekly statuses detailed", err.Error())
		return
	}

	data := make([]WeeklyStatusDetailedResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toWeeklyStatusDetailedResponse(row))
	}

	response.OK(c, http.StatusOK, data)
}

// region weekly status

func (h *SurveillanceHandler) ListRegionWeeklyStatusesByWeek(c *gin.Context) {
	ctx := c.Request.Context()

	epiWeekIDParam := c.Param("epiWeekID")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiWeekID is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", err.Error())
		return
	}

	data, err := h.weeklyStatusService.ListRegionByWeek(ctx, epiWeekID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list region weekly statuses by week", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

// national weekly status

func (h *SurveillanceHandler) ListNationalWeeklyStatusesByWeek(c *gin.Context) {
	ctx := c.Request.Context()

	epiWeekIDParam := c.Param("epiWeekID")
	if epiWeekIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "epiWeekID is required", "nil")
		return
	}

	epiWeekID, err := uuid.Parse(epiWeekIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", err.Error())
		return
	}

	data, err := h.weeklyStatusService.ListNationalByWeek(ctx, epiWeekID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list national weekly statuses by week", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

// alerts
func (h *SurveillanceHandler) ListAlerts(c *gin.Context) {
	ctx := c.Request.Context()

	params := AlertListParams{}
	var err error

	if value := c.Query("epiWeekID"); value != "" {
		params.EpiWeekID, err = uuid.Parse(value)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid epiWeekID", err.Error())
			return
		}
	}

	if value := c.Query("diseaseID"); value != "" {
		params.DiseaseID, err = uuid.Parse(value)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid diseaseID", err.Error())
			return
		}
	}

	if value := c.Query("districtID"); value != "" {
		params.DistrictID, err = uuid.Parse(value)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid districtID", err.Error())
			return
		}
	}

	if value := c.Query("regionID"); value != "" {
		params.RegionID, err = uuid.Parse(value)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "invalid regionID", err.Error())
			return
		}
	}

	filters := db.ListAlertsParams{
		EpiWeekID: uuid.NullUUID{
			UUID:  params.EpiWeekID,
			Valid: params.EpiWeekID != uuid.Nil,
		},
		DiseaseID: uuid.NullUUID{
			UUID:  params.DiseaseID,
			Valid: params.DiseaseID != uuid.Nil,
		},
		DistrictID: uuid.NullUUID{
			UUID:  params.DistrictID,
			Valid: params.DistrictID != uuid.Nil,
		},
		RegionID: uuid.NullUUID{
			UUID:  params.RegionID,
			Valid: params.RegionID != uuid.Nil,
		},
	}

	data, err := h.alertService.ListAlerts(ctx, filters)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list alerts", err.Error())
		return
	}

	response.OK(c, http.StatusOK, toAlertResponses(data))
}

// imports

func (h *SurveillanceHandler) ListImportBatches(c *gin.Context) {
	ctx := c.Request.Context()

	data, err := h.importService.ListImportBatches(ctx)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list import batches", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) GetImportBatchByID(c *gin.Context) {
	ctx := c.Request.Context()

	batchIDParam := c.Param("batchID")
	if batchIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "batchID is required", "nil")
		return
	}

	batchID, err := uuid.Parse(batchIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid batchID", err.Error())
		return
	}

	data, err := h.importService.GetImportBatchByID(ctx, batchID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to get import batch", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) ListImportRawRowsByBatch(c *gin.Context) {
	ctx := c.Request.Context()

	batchIDParam := c.Param("batchID")
	if batchIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "batchID is required", "nil")
		return
	}

	batchID, err := uuid.Parse(batchIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid batchID", err.Error())
		return
	}

	data, err := h.importService.ListImportRawRowsByBatch(ctx, batchID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list import raw rows", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}

func (h *SurveillanceHandler) CreateImportBatch(c *gin.Context) {
	ctx := c.Request.Context()

	var req db.CreateImportBatchParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	data, err := h.importService.CreateImportBatch(ctx, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to create import batch", err.Error())
		return
	}

	response.OK(c, http.StatusCreated, data)
}

func (h *SurveillanceHandler) UpdateImportBatchStatus(c *gin.Context) {
	ctx := c.Request.Context()

	batchIDParam := c.Param("batchID")
	if batchIDParam == "" {
		response.Fail(c, http.StatusBadRequest, "batchID is required", "nil")
		return
	}

	batchID, err := uuid.Parse(batchIDParam)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid batchID", err.Error())
		return
	}

	var req struct {
		Status string  `json:"status" binding:"required"`
		Notes  *string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	data, err := h.importService.UpdateImportBatchStatus(ctx, db.UpdateImportBatchStatusParams{
		ID:     batchID,
		Status: req.Status,
		Notes: sql.NullString{
			String: *req.Notes,
			Valid:  true,
		},
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to update import batch status", err.Error())
		return
	}

	response.OK(c, http.StatusOK, data)
}
